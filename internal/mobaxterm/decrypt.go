// Package mobaxterm 提供解析和解密 MobaXterm 会话文件的功能
package mobaxterm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
)

// decryptModern 使用 AES-CFB-8 解密 MobaXterm Professional 的加密密码
// 流程: Base64 解码 → SHA512(masterPassword)[0:32] 作为密钥 → ECB 加密 null bytes 生成 IV → AES-CFB-8 解密
func decryptModern(encryptedBase64, masterPassword string) (string, error) {
	// Base64 解码
	cipherData, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", fmt.Errorf("Base64 解码失败: %w", err)
	}

	if len(cipherData) == 0 {
		return "", fmt.Errorf("加密数据为空")
	}

	// 密钥 = SHA512(masterPassword)[0:32]
	hash := sha512.Sum512([]byte(masterPassword))
	key := hash[:32]

	// 创建 AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建 AES 密钥失败: %w", err)
	}

	// IV = AES-ECB 加密 16 字节全零
	iv := make([]byte, aes.BlockSize)
	block.Encrypt(iv, iv)

	// AES-CFB-8 解密（逐字节模式）
	plaintext := cfb8Decrypt(block, iv, cipherData)

	return string(plaintext), nil
}

// cfb8Decrypt 实现 AES-CFB-8 解密
// CFB-8 模式逐字节处理，与 Go 标准库的 CFB-128 (cipher.NewCFBDecrypter) 不同
// 算法:
//
//	对于每个密文字节:
//	  1. 加密当前 IV (AES-ECB) 得到 output
//	  2. plaintext_byte = ciphertext_byte XOR output[0]
//	  3. 左移 IV 一个字节，最后一个字节填入 ciphertext_byte
func cfb8Decrypt(block cipher.Block, iv, ciphertext []byte) []byte {
	plaintext := make([]byte, len(ciphertext))
	shiftReg := make([]byte, aes.BlockSize)
	copy(shiftReg, iv)

	output := make([]byte, aes.BlockSize)

	for i := 0; i < len(ciphertext); i++ {
		// 加密 shift register
		block.Encrypt(output, shiftReg)

		// 解密当前字节
		plaintext[i] = ciphertext[i] ^ output[0]

		// 左移 shift register，填入密文字节
		copy(shiftReg, shiftReg[1:])
		shiftReg[aes.BlockSize-1] = ciphertext[i]
	}

	return plaintext
}
