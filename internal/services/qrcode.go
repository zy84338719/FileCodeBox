package services

import (
	"encoding/base64"
	"fmt"

	"github.com/skip2/go-qrcode"
)

// QRCodeService 二维码服务
type QRCodeService struct{}

// NewQRCodeService 创建二维码服务实例
func NewQRCodeService() *QRCodeService {
	return &QRCodeService{}
}

// GenerateQRCode 生成二维码PNG图片
// @param data 二维码数据内容
// @param size 二维码尺寸（像素）
// @return []byte PNG图片数据
// @return error 错误信息
func (s *QRCodeService) GenerateQRCode(data string, size int) ([]byte, error) {
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

	// 设置二维码尺寸
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
func (s *QRCodeService) GenerateQRCodeBase64(data string, size int) (string, error) {
	pngData, err := s.GenerateQRCode(data, size)
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
func (s *QRCodeService) ValidateQRCodeData(data string) bool {
	if data == "" {
		return false
	}

	// 检查数据长度（二维码有容量限制）
	if len(data) > 4296 { // QR Code Version 40 最大容量
		return false
	}

	return true
}
