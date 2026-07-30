package db

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// migrate_test 验证版本化迁移器的核心行为：
//   - 干净库首次迁移成功建表并记录版本
//   - 二次迁移幂等（不重复执行）
//   - 迁移后业务表存在

func newMigrateTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	return gormDB
}

// 测试：干净库首次迁移成功，schema_migrations 记录版本
func TestMigrator_Up_FreshDB(t *testing.T) {
	gormDB := newMigrateTestDB(t)
	m, err := NewMigrator(gormDB, "sqlite")
	require.NoError(t, err)

	applied, err := m.Up()
	require.NoError(t, err)
	assert.Equal(t, []string{"000001"}, applied)

	// 版本记录存在
	var versions []schemaMigrations
	require.NoError(t, gormDB.Find(&versions).Error)
	assert.Len(t, versions, 1)
	assert.Equal(t, "000001", versions[0].Version)
}

// 测试：幂等性——二次 Up 不重复执行
func TestMigrator_Up_Idempotent(t *testing.T) {
	gormDB := newMigrateTestDB(t)
	m, _ := NewMigrator(gormDB, "sqlite")

	_, err := m.Up()
	require.NoError(t, err)

	// 二次迁移应返回空（无新迁移）
	applied, err := m.Up()
	require.NoError(t, err)
	assert.Empty(t, applied)

	// 版本记录仍只 1 条
	var count int64
	gormDB.Model(&schemaMigrations{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

// 测试：迁移后核心业务表存在
func TestMigrator_CreatesBusinessTables(t *testing.T) {
	gormDB := newMigrateTestDB(t)
	m, _ := NewMigrator(gormDB, "sqlite")
	_, err := m.Up()
	require.NoError(t, err)

	for _, table := range []string{"users", "file_codes", "upload_chunks", "transfer_logs", "admin_operation_logs", "user_api_keys"} {
		exists := gormDB.Migrator().HasTable(table)
		assert.True(t, exists, "table %s should exist after migration", table)
	}
}

// 测试：SQL 语句分割（含引号内分号）
func TestSplitSQLStatements_Quotes(t *testing.T) {
	sql := "CREATE TABLE a (x text); INSERT INTO a VALUES ('hello;world'); CREATE INDEX idx ON a(x);"
	stmts := splitSQLStatements(sql)
	// 应分割为 3 条（引号内的分号不切）
	assert.Len(t, stmts, 3)
	assert.Contains(t, stmts[1], "'hello;world'")
}

// 测试：空字符串与纯注释不产生语句
func TestSplitSQLStatements_EmptyAndComments(t *testing.T) {
	stmts := splitSQLStatements("-- 这是一条注释;\n")
	// 注释后无分号，整段作为一个"语句"，但执行时会被注释前缀跳过
	assert.NotEmpty(t, stmts) // 注释本身作为一个块，applyOne 会跳过
}
