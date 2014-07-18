// Package aes provides encryption for the vsafe app. These encryption routines
// do not support embedded null characters in strings.
package aes

import (
  "crypto/aes"
  "crypto/cipher"
  "encoding/base64"
  "errors"
  "github.com/keep94/appcommon/kdf"
)

var (
  errNoPKCS7 = errors.New("no PKCS7 padding")
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
  return EncryptB(PKCS7([]byte(plain)), key)
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
  plainb, err = UnPKCS7(plainb)
  if err != nil {
    return "", err
  }
  return string(plainb), nil
}

// PKCS7 returns a new slice that contains data along with PKCS7 padding.
// Returned slice is suitable for aes encryption.
func PKCS7(data []byte) []byte {
  blockSize := aes.BlockSize
  datalen := len(data)
  padSize := blockSize - datalen % blockSize
  result := make([]byte, datalen + padSize)
  idx := copy(result, data)
  padding := result[idx:]
  for i := range padding {
    padding[i] = byte(padSize)
  }
  return result
}

// UnPKCS7 returns a view of data without the PKCS7 padding. UnPKCS7
// returns an error if data is not PKCS7 padded.
func UnPKCS7(data []byte) ([]byte, error) {
  blockSize := aes.BlockSize
  datalen := len(data)
  if datalen % blockSize != 0 {
    return nil, errNoPKCS7
  }
  padbyte := data[datalen - 1]
  padSize := int(padbyte)
  if padSize > blockSize {
    return nil, errNoPKCS7
  }
  padding := data[datalen - padSize:]
  for i := range padding {
    if padding[i] != padbyte {
       return nil, errNoPKCS7
    }
  }
  return data[:datalen - padSize], nil
}

