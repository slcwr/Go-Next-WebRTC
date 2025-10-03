package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// データベースから取得したハッシュ
	hash := "$2a$10$p.3ls0qg8jLoRLABAnDx3eELDb2W2bHJ9cOUe6ek39Pn.ibeZUcqi"

	// 試すパスワード
	passwords := []string{
		"P@ssw0rd",
		"p@ssw0rd",
		"P@ssword",
		"P@ssw0rd123",
	}

	for _, password := range passwords {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if err == nil {
			fmt.Printf("✓ Password '%s' matches!\n", password)
		} else {
			fmt.Printf("✗ Password '%s' does not match\n", password)
		}
	}
}
