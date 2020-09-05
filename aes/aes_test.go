package aes_test

import (
	"github.com/keep94/vsafe/aes"
	"testing"
)

var (
	someKey = []byte("12345678901234567890123456789012")
)

func TestEncryptDecrypt(t *testing.T) {
	verifyEncryptDecrypt(t, "aardvark")
	verifyEncryptDecrypt(t, "1234567890123456")
	verifyEncryptDecrypt(t, "1234567890123456 ")
	verifyEncryptDecrypt(t, "123456789012345 ")
	verifyEncryptDecrypt(t, " now is the time for all good men to come to the aid of their party ")
}

func TestEncryptsSameTextDifferently(t *testing.T) {
	encoded, err := aes.Encrypt("aardvark", someKey)
	if err != nil {
		t.Fatal(err)
	}
	encodedAgain, err := aes.Encrypt("aardvark", someKey)
	if err != nil {
		t.Fatal(err)
	}
	if encoded == encodedAgain {
		t.Error("Expected same text to be encrypted differently every time.")
	}
}

func TestEncryptSecurity(t *testing.T) {
	anotherKey := []byte("12345678901234567890123456789013")
	encoded, err := aes.Encrypt("aardvark", someKey)
	if err != nil {
		t.Fatal(err)
	}
	decoded, _ := aes.Decrypt(encoded, anotherKey)
	if decoded == "aardvark" {
		t.Error("Expected different decryption with different key")
	}
}

func verifyEncryptDecrypt(t *testing.T, plain string) {
	encoded, err := aes.Encrypt(plain, someKey)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := aes.Decrypt(encoded, someKey)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != plain {
		t.Errorf("Expected to get same thing back: '%s', got '%s' %d %d", plain, decoded, len(plain), len(decoded))
	}
}
