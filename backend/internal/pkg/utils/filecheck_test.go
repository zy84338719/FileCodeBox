package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsBlockedExtension_Blocked(t *testing.T) {
	bl := DefaultBlockedExtensions()
	assert.True(t, IsBlockedExtension("malware.exe", bl))
	assert.True(t, IsBlockedExtension("script.SH", bl))   // 大小写不敏感
	assert.True(t, IsBlockedExtension("a.b.bat", bl))
	assert.True(t, IsBlockedExtension("scr.scr", bl))
}

func TestIsBlockedExtension_Allowed(t *testing.T) {
	bl := DefaultBlockedExtensions()
	assert.False(t, IsBlockedExtension("photo.jpg", bl))
	assert.False(t, IsBlockedExtension("doc.pdf", bl))
	assert.False(t, IsBlockedExtension("noext", bl))
	assert.False(t, IsBlockedExtension("archive.tar.gz", bl))
}

func TestCheckUploadSize_OK(t *testing.T) {
	require.NoError(t, CheckUploadSize(1024, 10*1024*1024))
	require.NoError(t, CheckUploadSize(0, 10*1024*1024)) // 空文件允许
}

func TestCheckUploadSize_TooLarge(t *testing.T) {
	err := CheckUploadSize(11*1024*1024, 10*1024*1024)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrFileTooLarge))
}

func TestCheckUploadSize_NoLimit(t *testing.T) {
	// maxSize=0 表示不限
	require.NoError(t, CheckUploadSize(999999999, 0))
}
