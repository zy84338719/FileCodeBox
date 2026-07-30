// Package db 提供数据库连接与迁移。
//
// 本文件实现轻量级版本化数据库迁移器（无需外部 CLI 依赖）：
//   - migrations/ 目录下的 SQL 文件用 //go:embed 嵌入二进制
//   - schema_migrations 表记录已执行的迁移版本
//   - 启动时按版本号顺序执行未应用的 up 迁移
//
// 命名约定：<version>_<name>.up.sql / <version>_<name>.down.sql
// version 为零填充整数（000001、000002…），确保字典序=数值序。
//
// 用法（bootstrap）：
//
//	migrator, err := db.NewMigrator(database, db.DriverSQLite)
//	if err != nil { return err }
//	if err := migrator.Up(); err != nil { return err }
package db

import (
	"embed"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const migrationsDir = "migrations"

// schemaMigrationsTable 记录已执行迁移的版本号。
type schemaMigrations struct {
	Version string `gorm:"primaryKey;size:255"`
}

// TableName 固定迁移记录表名。
func (schemaMigrations) TableName() string { return "schema_migrations" }

// Migrator 版本化迁移器。
type Migrator struct {
	db       *gorm.DB
	dialect  string // sqlite / mysql / postgres
}

// NewMigrator 创建迁移器并确保 schema_migrations 表存在。
func NewMigrator(gormDB *gorm.DB, dialect string) (*Migrator, error) {
	m := &Migrator{db: gormDB, dialect: dialect}
	if err := gormDB.AutoMigrate(&schemaMigrations{}); err != nil {
		return nil, fmt.Errorf("create schema_migrations table: %w", err)
	}
	return m, nil
}

// migrationFile 描述一个迁移文件。
type migrationFile struct {
	version string
	name    string
	path    string
	isDown  bool
}

// discoverUpMigrations 扫描嵌入的 up 迁移文件，按 version 排序。
func discoverUpMigrations() ([]migrationFile, error) {
	entries, err := migrationsFS.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}
	var files []migrationFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		// 解析 version: "000001_init.up.sql" → "000001"
		base := strings.TrimSuffix(name, ".up.sql")
		parts := strings.SplitN(base, "_", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid migration filename (expect <version>_<name>.up.sql): %s", name)
		}
		files = append(files, migrationFile{
			version: parts[0],
			name:    parts[1],
			path:    filepath.Join(migrationsDir, name),
			isDown:  false,
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })
	return files, nil
}

// appliedVersions 查询已执行的迁移版本。
func (m *Migrator) appliedVersions() ([]string, error) {
	var rows []schemaMigrations
	if err := m.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.Version
	}
	return out, nil
}

// Up 执行所有未应用的 up 迁移，返回应用的版本列表。
//
// 每个 up 文件在一个独立事务中执行；若某个迁移失败则停止并返回错误，
// 已成功提交的迁移保持有效（事务粒度=单文件）。
func (m *Migrator) Up() ([]string, error) {
	files, err := discoverUpMigrations()
	if err != nil {
		return nil, err
	}
	applied, err := m.appliedVersions()
	if err != nil {
		return nil, err
	}
	var newlyApplied []string
	for _, f := range files {
		if slices.Contains(applied, f.version) {
			continue
		}
		if err := m.applyOne(f); err != nil {
			return newlyApplied, fmt.Errorf("migration %s (%s) failed: %w", f.version, f.name, err)
		}
		newlyApplied = append(newlyApplied, f.version)
	}
	return newlyApplied, nil
}

// applyOne 读取并执行单个迁移文件，记录版本。
func (m *Migrator) applyOne(f migrationFile) error {
	content, err := migrationsFS.ReadFile(f.path)
	if err != nil {
		return fmt.Errorf("read migration file %s: %w", f.path, err)
	}
	statements := splitSQLStatements(string(content))
	return m.db.Transaction(func(tx *gorm.DB) error {
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" || strings.HasPrefix(stmt, "--") {
				continue
			}
			if err := tx.Exec(stmt).Error; err != nil {
				return fmt.Errorf("exec statement failed: %w\nstatement: %s", err, truncate(stmt, 200))
			}
		}
		// 记录版本
		return tx.Create(&schemaMigrations{Version: f.version}).Error
	})
}

// splitSQLStatements 按 ';' 懒分割 SQL 语句（忽略引号内的分号）。
// 处理 '...' 与 "..." 字符串字面量，避免误切。
func splitSQLStatements(sql string) []string {
	var stmts []string
	var cur strings.Builder
	inSingle, inDouble := false, false
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		cur.WriteByte(c)
		switch c {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case ';':
			if !inSingle && !inDouble {
				stmts = append(stmts, cur.String())
				cur.Reset()
			}
		}
	}
	if strings.TrimSpace(cur.String()) != "" {
		stmts = append(stmts, cur.String())
	}
	return stmts
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
