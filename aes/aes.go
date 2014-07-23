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
  b := ([]byte)(plain)
  PKCS7(&b)
  return EncryptB(b, key)
}

// DecryptB decrypts encoded using key. encoded is a base64 encoded string
// from EncryptB. Decrypt returns a new slice.
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
  err = UnPKCS7(&plainb)
  if err != nil {
    return "", err
  }
  return string(plainb), nil
}

// PKCS7 adds PKCS7 padding to data in place for AES encryption.
func PKCS7(data *[]byte) {
  blockSize := aes.BlockSize
  padSize := blockSize - len(*data) % blockSize
  padding := make([]byte, padSize)
  for i := range padding {
    padding[i] = byte(padSize)
  }
  *data = append(*data, padding...)
}

// UnPKCS7 removes PKCS7 padding for AES encryption from data in place and
// returns an error if data is not PKCS7 padded.
func UnPKCS7(data *[]byte) error {
  blockSize := aes.BlockSize
  datalen := len(*data)
  if datalen % blockSize != 0 {
    return errNoPKCS7
  }
  padbyte := (*data)[datalen - 1]
  padSize := int(padbyte)
  if padSize > blockSize {
    return errNoPKCS7
  }
  padding := (*data)[datalen - padSize:]
  for i := range padding {
    if padding[i] != padbyte {
       return errNoPKCS7
    }
  }
  *data = (*data)[:datalen - padSize]
  return nil
}

