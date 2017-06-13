// Package model provides the data structures for the vsafe app
package vsafe

import (
  "crypto/hmac"
  "encoding/base64"
  "errors"
  "github.com/keep94/appcommon/idset"
  "github.com/keep94/appcommon/kdf"
  "github.com/keep94/vsafe/aes"
  "net/url"
)

var (
  // Wrong password provided for user.
  ErrWrongPassword = errors.New("vsafe: Wrong Password.")
  // Invalid key provided to decrypt an entry.
  ErrKeyMismatch = errors.New("vsafe: Key Mismatch.")
)

// Key instances are used to encrypt / decrypt user name, password, and
// special fields in an entry.
// These instances must be treated as immutable.
type Key struct {
  Id int64
  Value []byte
}

// Equal returns true if this instance represents the same key as other.
// Otherwise Equal returns false.
func (k *Key) Equal(other *Key) bool {
  if k == other {
    return true
  }
  if k == nil || other == nil {
    return false
  }
  return k.Id == other.Id && hmac.Equal(k.Value, other.Value)
}

// User instances represent a user of the vsafe app.
type User struct {
  // User ID
  Id int64
  // Master user lD. 0 means this user is also the master user. Users having
  // the same master are able to see the same entries in the database as the
  // ID of the key used to decrypt those entries corresponds to the master
  // user ID.
  Owner int64
  // The Name of this user.
  Name string
  // This user's encryption key encrypted by the user's own password.
  Key string
  // The checksum of the user's key. Used to verify that the password for
  // a user is correct.
  Checksum string
}

// Init initializes this user instance with a user name and password so that
// this user is a master user and has its own random key.
func (u *User) Init(name, password string) error {
  return u.InitWithKey(name, password, &Key{Id: 0, Value: kdf.Random(32)})
}

// InitWithKey initializes this user instance with a user name and password
// so that the user uses key as its key.
func (u *User) InitWithKey(name, password string, key *Key) (err error) {
  u.Owner = key.Id
  u.Name = name
  if u.Key, err = aes.EncryptB(
      key.Value,
      kdf.KDF(
          []byte(password),
          kdf.DefaultSalt,
          kdf.DefaultReps)); err != nil {
    return
  }
  u.Checksum = base64.StdEncoding.EncodeToString(
      kdf.NewHMAC(key.Value, kdf.DefaultReps))
  return
}

// VerifyPassword verifies that password is the password for this user.
// On success, it returns the key of this user.
func (u *User) VerifyPassword(password string) (*Key, error) {
  var key []byte
  var err error
  if key, err = u.verifyPassword(password); err != nil {
    return nil, err
  }
  return &Key{Id: u.GetOwner(), Value: key}, nil
}

// ChangePassword changes the password of this user.
func (u *User) ChangePassword(oldPass, newPass string) error {
  var key []byte
  var err error
  if key, err = u.verifyPassword(oldPass); err != nil {
    return err
  }
  u.Key, err = aes.EncryptB(key, kdf.KDF([]byte(newPass), kdf.DefaultSalt, kdf.DefaultReps))
  return err
}

// GetOwner returns the ID of the master user of this user. In the case
// that this user is a master user, GetOwner only works correctly after
// this user has been saved in persistent storage and has an ID.
func (u *User) GetOwner() int64 {
  if u.Owner != 0 {
    return u.Owner
  }
  return u.Id
}

func (u *User) verifyPassword(password string) ([]byte, error) {
  var key []byte
  var err error
  key, err = aes.DecryptB(u.Key, kdf.KDF([]byte(password), kdf.DefaultSalt, kdf.DefaultReps))
  if err != nil {
    return nil, err
  }
  var checksum []byte
  checksum, err = base64.StdEncoding.DecodeString(u.Checksum)
  if err != nil {
    return nil, err
  }
  if !kdf.VerifyHMAC(key, checksum, kdf.DefaultReps) {
    return nil, ErrWrongPassword
  }
  return key, nil
}

// Category represents a group of entries
type Category struct {
  // Category id
  Id int64
  // The owner which corresponds to the master user ID.
  Owner int64
  // Category name
  Name string
}

// Entry represents an entry stored in the vsafe database. UName
// Password and Special fields are encrypted in persistent storage.
type Entry struct {
  // The ID of the entry
  Id int64
  // The owner of the entry which corresponds to the master user ID.
  Owner int64
  // The URL of the website. May be nil.
  Url *url.URL
  // The title
  Title string
  // The description
  Desc string
  // The user name
  UName string
  // The password
  Password string
  // Special instructions
  Special string
  // Categories to which this entry belongs
  Categories idset.IdSet
}

// Read an EntryWithEtag instead of an Entry to collect the entry's etag
type EntryWithEtag struct {
  Entry
  Etag uint64
}

func (e *EntryWithEtag) GetPtr() interface{} {
  return &e.Entry
}

func (e *EntryWithEtag) SetEtag(etag uint64) {
  e.Etag = etag
}

// Encrypt encrypts sensitive fields in this instance using key namely
// UName, Password, and Special
func (e *Entry) Encrypt(key *Key) (err error) {
  if e.UName, err = aes.Encrypt(e.UName, key.Value); err != nil {
    return err
  }
  if e.Password, err = aes.Encrypt(e.Password, key.Value); err != nil {
    return err
  }
  if e.Special, err = aes.Encrypt(e.Special, key.Value); err != nil {
    return err
  }
  e.Owner = key.Id
  return
}

// Decrypt decrypts sensitive fields in this instance using key namely
// UName, Password, and Special. Decrypt returns ErrKeyMismatch if the
// provided key does not have the same Id as the key used to encrypt.
func (e *Entry) Decrypt(key *Key) (err error) {
  if e.Owner != key.Id {
    return ErrKeyMismatch
  }
  if e.UName, err = aes.Decrypt(e.UName, key.Value); err != nil {
    return err
  }
  if e.Password, err = aes.Decrypt(e.Password, key.Value); err != nil {
    return err
  }
  if e.Special, err = aes.Decrypt(e.Special, key.Value); err != nil {
    return err
  }
  return
}
