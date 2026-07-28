package user

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zy84338719/fileCodeBox/backend/internal/pkg/auth"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
	"gorm.io/gorm"
)

// =====================================================================
// user service 单测（Shape B：DAO 通过全局 db.GetDB() 访问）
// 复用 share service 的 SetDatabaseInstance 注入模式。
// =====================================================================

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&model.User{}, &model.UserAPIKey{}))
	db.SetDatabaseInstance(gormDB)
	t.Cleanup(func() { db.SetDatabaseInstance(nil) })
	return gormDB
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	newTestDB(t)
	// 登录会生成 JWT，需要设置 secret（避免使用硬编码默认值）
	auth.SetJWTSecret("test-secret-for-user-service")
	return NewService()
}

// 测试：创建用户成功
func TestCreate_Success(t *testing.T) {
	svc := newTestService(t)
	resp, err := svc.Create(context.Background(), &CreateUserReq{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "passw0rd!",
		Nickname: "Alice",
	})
	require.NoError(t, err)
	assert.Equal(t, "alice", resp.Username)
	assert.Equal(t, "alice@example.com", resp.Email)
	assert.Equal(t, "Alice", resp.Nickname)
}

// 测试：重复用户名拒绝
func TestCreate_DuplicateUsername(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.Create(context.Background(), &CreateUserReq{
		Username: "bob", Email: "bob1@example.com", Password: "x",
	})
	require.NoError(t, err)

	_, err = svc.Create(context.Background(), &CreateUserReq{
		Username: "bob", Email: "bob2@example.com", Password: "x",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// 测试：重复邮箱拒绝
func TestCreate_DuplicateEmail(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.Create(context.Background(), &CreateUserReq{
		Username: "carol", Email: "carol@example.com", Password: "x",
	})
	require.NoError(t, err)

	_, err = svc.Create(context.Background(), &CreateUserReq{
		Username: "carol2", Email: "carol@example.com", Password: "x",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// 测试：登录成功返回 JWT token
func TestLogin_Success(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.Create(context.Background(), &CreateUserReq{
		Username: "dave", Email: "dave@example.com", Password: "secret123",
	})
	require.NoError(t, err)

	resp, token, err := svc.Login(context.Background(), "dave", "secret123")
	require.NoError(t, err)
	assert.Equal(t, "dave", resp.Username)
	assert.NotEmpty(t, token, "应返回 JWT token")

	// token 应可解析
	claims, err := auth.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, "dave", claims.Username)
}

// 测试：密码错误
func TestLogin_WrongPassword(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.Create(context.Background(), &CreateUserReq{
		Username: "eve", Email: "eve@example.com", Password: "correct",
	})
	require.NoError(t, err)

	_, _, err = svc.Login(context.Background(), "eve", "wrong")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "密码错误")
}

// 测试：登录不存在的用户
func TestLogin_UserNotFound(t *testing.T) {
	svc := newTestService(t)
	_, _, err := svc.Login(context.Background(), "ghost", "whatever")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "密码错误") // 不泄露用户是否存在
}

// 测试：GetByID
func TestGetByID(t *testing.T) {
	svc := newTestService(t)
	created, err := svc.Create(context.Background(), &CreateUserReq{
		Username: "frank", Email: "frank@example.com", Password: "x",
	})
	require.NoError(t, err)

	got, err := svc.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "frank", got.Username)
}

// 测试：GetByID 不存在
func TestGetByID_NotFound(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.GetByID(context.Background(), 9999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// 测试：删除用户
func TestDelete(t *testing.T) {
	svc := newTestService(t)
	created, err := svc.Create(context.Background(), &CreateUserReq{
		Username: "grace", Email: "grace@example.com", Password: "x",
	})
	require.NoError(t, err)

	require.NoError(t, svc.Delete(context.Background(), created.ID))
	_, err = svc.GetByID(context.Background(), created.ID)
	assert.Error(t, err)
}
