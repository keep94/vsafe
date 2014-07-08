// Package aes provides encryption for the vsafe app. These encryption routines
// do not support embedded null characters in strings.
package aes

import (
  "crypto/aes"
  "crypto/cipher"
  "encoding/base64"
  "github.com/keep94/appcommon/kdf"
  "strings"
)

// EncryptB encrypts plain with key and returns a base64 encoded string.
func EncryptB(plain, key []byte) (string, error) {
  block, err := aes.NewCipher(key)
  if err != nil {
    return "", err
  }
  result := make([]byte, len(plain) + aes.BlockSize)
  iv := kdf.Random(aes.BlockSize)
  idx := copy(result, iv)
  mode := cipher.NewCBCEncrypter(block, iv)
  mode.CryptBlocks(result[idx:], []byte(plain))
  return base64.StdEncoding.EncodeToString(result), nil
}

// Encrypt encrypts plain with key and returns a base64 encoded string.
func Encrypt(plain string, key []byte) (string, error) {
  remainder := len(plain) % aes.BlockSize
  if remainder > 0 {
    plain += strings.Repeat("\u0000", aes.BlockSize - remainder)
  }
  return EncryptB([]byte(plain), key)
}

// DecryptB decrypts encoded using key. encoded is a base64 encoded string
// from EncryptB.
func DecryptB(encoded string, key []byte) ([]byte, error) {
  encodedb, err := base64.StdEncoding.DecodeString(encoded)
  if err != nil {
    return nil, err
  }
  block, err := aes.NewCipher(key)
  if err != nil {
    return nil, err
  }
  iv := encodedb[:aes.BlockSize]
  mode := cipher.NewCBCDecrypter(block, iv)
  plainb := make([]byte, len(encodedb) - aes.BlockSize)
  mode.CryptBlocks(plainb, encodedb[aes.BlockSize:])
  return plainb, nil
}

// Decrypt decrypts encoded using key. encoded is a base64 encoded string
// from Encrypt.
func Decrypt(encoded string, key []byte) (string, error) {
  plainb, err := DecryptB(encoded, key)
  if err != nil {
    return "", err
  }
  result := string(plainb)
  index := strings.Index(result, "\u0000")
  if index == -1 {
    return result, nil
  }
  return result[:index], nil
}
