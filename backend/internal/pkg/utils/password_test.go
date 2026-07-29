package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword_NonEmpty(t *testing.T) {
	hash, err := HashPassword("secret123")
	require.NoError(t, err)
	assert.NotEqual(t, "secret123", hash)
	assert.Len(t, hash, 60) // bcrypt cost=10 → 60 字符
}

func TestHashPassword_UniqueSalts(t *testing.T) {
	h1, _ := HashPassword("same")
	h2, _ := HashPassword("same")
	assert.NotEqual(t, h1, h2) // 盐不同
}

func TestHashPassword_EmptyReturnsEmpty(t *testing.T) {
	hash, err := HashPassword("")
	require.NoError(t, err)
	assert.Equal(t, "", hash)
}

func TestCheckPassword_Correct(t *testing.T) {
	hash, _ := HashPassword("mypass")
	assert.True(t, CheckPassword(hash, "mypass"))
}

func TestCheckPassword_Wrong(t *testing.T) {
	hash, _ := HashPassword("mypass")
	assert.False(t, CheckPassword(hash, "wrong"))
}

func TestCheckPassword_EmptyHashNoPassword(t *testing.T) {
	// 空 hash 表示"无密码"：空密码校验通过，非空密码拒绝
	assert.True(t, CheckPassword("", ""))
	assert.False(t, CheckPassword("", "anything"))
}
