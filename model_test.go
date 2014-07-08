package vsafe_test

import (
  "github.com/keep94/appcommon/etag"
  "github.com/keep94/appcommon/kdf"
  "github.com/keep94/vsafe"
  "net/url"
  "testing"
)

func TestKeyEqual(t *testing.T) {
  first := &vsafe.Key{Id: 1}
  second := &vsafe.Key{Id: 1}
  third := &vsafe.Key{Id: 2}
  fourth := &vsafe.Key{Id: 1, Value: []byte{0x3}}
  var fifth *vsafe.Key
  var sixth *vsafe.Key
  if !first.Equal(first) {
    t.Error("Expected first and first to be equal.")
  }
  if !first.Equal(second) {
    t.Error("Expected first and second to be equal.")
  }
  if first.Equal(third) {
    t.Error("Expected first and third not to be equal.")
  }
  if first.Equal(fourth) {
    t.Error("Expected first and fouth not to be equal.")
  }
  if first.Equal(fifth) {
    t.Error("Expected first and fifth not to be equal.")
  }
  if fifth.Equal(first) {
    t.Error("Expected first and fifth not to be equal.")
  }
  if !fifth.Equal(sixth) {
    t.Error("Expected first and sixt to be equal.")
  }
}

func TestUser(t *testing.T) {
  user := vsafe.User{Id: 1}
  var err error
  var key *vsafe.Key
  var key2 *vsafe.Key
  if err = user.Init("keep94", "somepassword"); err != nil {
    t.Fatalf("Error creating user: %v", err)
  }
  if key, err = user.VerifyPassword("somepassword"); err != nil {
    t.Fatalf("Verify password failed: %v", err)
  }
  if _, err = user.VerifyPassword("aardvark"); err != vsafe.ErrWrongPassword {
    t.Errorf("Expected wrong password, got %v", err)
  }
  user2 := vsafe.User{Id: 2}
  if err = user2.InitWithKey("keep94+admin", "foo", key); err != nil {
    t.Fatalf("Error creating user: %v", err)
  }
  if key2, err = user2.VerifyPassword("foo"); err != nil {
    t.Fatalf("Verify password failed: %v", err)
  }
  if !key2.Equal(key) {
    t.Error("key and key2 should be equal.")
  }
  if key.Id != 1 {
    t.Error("Expected key Id to be 1")
  }
}

func TestChangePassword(t *testing.T) {
  user := vsafe.User{Id: 1}
  var err error
  if err = user.Init("keep94", "somepassword"); err != nil {
    t.Fatalf("Error creating user: %v", err)
  }
  if err = user.ChangePassword("somepassword", "another"); err != nil {
    t.Fatalf("Verify password failed: %v", err)
  }
  if _, err = user.VerifyPassword("another"); err != nil {
    t.Error("Expected password to change")
  }
  if err = user.ChangePassword("somepassword", "be"); err != vsafe.ErrWrongPassword {
    t.Errorf("Expected ErrWrongPassword, got %v", err)
  }
}

func TestKey(t *testing.T) {
  user := vsafe.User{Id: 1}
  var err error
  if err = user.Init("keep94", "somepassword"); err != nil {
    t.Fatalf("Error creating user: %v", err)
  }
  user2 := vsafe.User{Id: 2}
  if err = user2.Init("keep94", "somepassword"); err != nil {
    t.Fatalf("Error creating user: %v", err)
  }
  var key *vsafe.Key
  if key, err = user.VerifyPassword("somepassword"); err != nil {
    t.Fatalf("Verify password failed: %v", err)
  }
  var keyAgain *vsafe.Key
  if keyAgain, err = user.VerifyPassword("somepassword"); err != nil {
    t.Fatalf("Verify password failed: %v", err)
  }
  var key2 *vsafe.Key
  if key2, err = user2.VerifyPassword("somepassword"); err != nil {
    t.Fatalf("Verify password failed: %v", err)
  }
  if !key.Equal(keyAgain) {
    t.Errorf("Expected equal keys")
  }
  if key.Equal(key2) {
    t.Errorf("Expected unequal keys")
  }
}

func TestEncryptDecrypt(t *testing.T) {
  e := vsafe.Entry{
      Title: "title",
      Desc: "desc",
      UName: "foo",
      Password: "bar",
      Special: "baz"}
  key := &vsafe.Key{Id: 1, Value: kdf.Random(32)}
  encrypted := e
  if err := encrypted.Encrypt(key); err != nil {
    t.Fatalf("Got error encrypting: %v", err)
  }
  if encrypted.UName == e.UName || encrypted.Password == e.Password || encrypted.Special == e.Special {
    t.Error("Encrypted is the same as plain text")
  }
  if encrypted.Title != e.Title || encrypted.Desc != e.Desc {
    t.Error("Encrypted fields should be the same as plain text")
  }
  decrypted := encrypted
  if err := decrypted.Decrypt(key); err != nil {
    t.Fatalf("Got error encrypting: %v", err)
  }
  // Owner of entity changes to key Id
  e.Owner = key.Id
  if decrypted != e {
    t.Errorf("Expected %v, got %v", e, decrypted)
  }
  if err := decrypted.Decrypt(&vsafe.Key{Id: 2, Value: kdf.Random(32)}); err != vsafe.ErrKeyMismatch {
    t.Errorf("Expected ErrKeyMismatch, got %v", err)
  }
}

func TestEntryEtag(t *testing.T) {
  url1, err := url.Parse("http://www.google.com")
  if err != nil {
    t.Fatal("Error parsing url1")
  }
  url2, err := url.Parse("http://www.yahoo.com")
  if err != nil {
    t.Fatal("Error parsing url1")
  }
  entry1 := vsafe.Entry{Url: url1}
  entry2 := vsafe.Entry{Url: url2}
  entry3 := vsafe.Entry{}
  entry4 := entry2
  tag1, err := etag.Etag32(&entry1)
  if err != nil {
    t.Fatal("Error computing etag")
  }
  tag2, err := etag.Etag32(&entry2)
  if err != nil {
    t.Fatal("Error computing etag")
  }
  tag3, err := etag.Etag32(&entry3)
  if err != nil {
    t.Fatal("Error computing etag")
  }
  tag4, err := etag.Etag32(&entry4)
  if err != nil {
    t.Fatal("Error computing etag")
  }
  if tag1 == tag2 || tag1 == tag3 || tag2 == tag3 {
    t.Error("Etags should be unique")
  }
  if tag4 != tag2 {
    t.Error("Expected etags to be equal")
  }
}

