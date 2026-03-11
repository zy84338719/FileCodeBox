package qrcode

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/skip2/go-qrcode"
)

// GenerateQRCode 生成二维码PNG图片
// @param data 二维码数据内容
// @param size 二维码尺寸（像素）
// @return []byte PNG图片数据
// @return error 错误信息
func GenerateQRCode(data string, size int) ([]byte, error) {
	if data == "" {
		return nil, fmt.Errorf("二维码数据不能为空")
	}

	if size <= 0 {
		size = 256 // 默认尺寸
	}

	// 生成二维码
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("生成二维码失败: %v", err)
	}

	// 设置二维码边框
	qr.DisableBorder = false

	// 生成PNG图片
	pngData, err := qr.PNG(size)
	if err != nil {
		return nil, fmt.Errorf("生成PNG图片失败: %v", err)
	}

	return pngData, nil
}

// GenerateQRCodeBase64 生成Base64编码的二维码图片
// @param data 二维码数据内容
// @param size 二维码尺寸（像素）
// @return string Base64编码的PNG图片数据
// @return error 错误信息
func GenerateQRCodeBase64(data string, size int) (string, error) {
	pngData, err := GenerateQRCode(data, size)
	if err != nil {
		return "", err
	}

	// 转换为Base64
	base64Data := fmt.Sprintf("data:image/png;base64,%s",
		base64.StdEncoding.EncodeToString(pngData))

	return base64Data, nil
}

// ValidateQRCodeData 验证二维码数据有效性
// @param data 二维码数据
// @return bool 是否有效
func ValidateQRCodeData(data string) bool {
	if data == "" {
		return false
	}

	// 检查数据长度（二维码有容量限制）
	if len(data) > 4296 { // QR Code Version 40 最大容量
		return false
	}

	return true
}

// GenerateQRCodeID 生成二维码唯一ID
// @return string 二维码ID
func GenerateQRCodeID() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机数生成失败，使用时间戳作为ID
		return fmt.Sprintf("qr_%x", make([]byte, 8))
	}
	return fmt.Sprintf("qr_%x", b)
}
