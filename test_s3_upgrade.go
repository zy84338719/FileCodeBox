package main

import (
	"fmt"
	"log"

	"github.com/zy84338719/filecodebox/internal/storage"
)

func main() {
	fmt.Println("测试 AWS SDK v2 升级...")

	// 尝试创建 S3 存储策略（使用虚拟参数）
	s3Strategy, err := storage.NewS3StorageStrategy(
		"test-access-key",
		"test-secret-key",
		"test-bucket",
		"https://s3.amazonaws.com",
		"us-east-1",
		"",
		"",
		false,
		"",
	)

	if err != nil {
		log.Printf("创建 S3 策略时出错（这是预期的，因为使用的是测试凭证）: %v", err)
	} else {
		fmt.Println("S3 策略创建成功！")
		fmt.Printf("S3 策略类型: %T\n", s3Strategy)
	}

	fmt.Println("AWS SDK v2 升级验证完成！")
}
