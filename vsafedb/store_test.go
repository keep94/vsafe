package vsafedb_test

import (
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/etag"
  "github.com/keep94/appcommon/kdf"
  "github.com/keep94/gofunctional3/functional"
  "github.com/keep94/vsafe"
  "github.com/keep94/vsafe/vsafedb"
  "net/url"
  "testing"
)

var (
  kAnEntry = &vsafe.Entry{
      UName: "uname",
      Password: "password",
      Special: "special",
  }
  kOrigEntry = &vsafe.Entry{
      UName: "somename",
      Password: "somethingelse",
      Special: "xxx",
  }
  kKey = &vsafe.Key{Id: 7, Value: kdf.Random(32)}
  kTransaction db.Transaction = 0
)  

func TestAddEntry(t *testing.T) {
  var store FakeStore
  entry := *kAnEntry
  var id int64
  var err error
  if id, err = vsafedb.AddEntry(&store, kKey, &entry); err != nil {
    t.Fatalf("Error adding tostore: %v", err)
  }
  if id != 1 {
    t.Errorf("expected 1, got %d", id)
  }
  // entry should not change as side effect
  if entry != *kAnEntry {
    t.Errorf("Expected %v, got %v", *kAnEntry, entry)
  }
  // Key Id should be stored with entry.
  if out := store[0].Owner; out != kKey.Id {
    t.Errorf("Expected %d, got %d", kKey.Id, out)
  }
  if store[0].UName == entry.UName || store[0].Password == entry.Password || store[0].Special == entry.Special {
    t.Error("Expected database to be encrypted.")
  }
  var readEntry vsafe.Entry
  if err = vsafedb.EntryById(store, nil, 1, kKey, &readEntry); err != nil {
    t.Fatalf("Error reading store: %v", err)
  }
  origEntry := *kAnEntry
  origEntry.Id = 1
  origEntry.Owner = kKey.Id
  if readEntry != origEntry {
    t.Errorf("Expected %v, got %v", origEntry, readEntry)
  }
  if err = vsafedb.EntryById(
      store, nil, 9999, kKey, &readEntry); err != vsafedb.ErrNoSuchId {
    t.Errorf("Expected ErrNoSuchId, got %v", err)
  }
}

func TestUpdateEntry(t *testing.T) {
  store := make(FakeStore, 1)
  origEntry := *kAnEntry
  origEntry.Id = 1
  entry := origEntry
  if err := vsafedb.UpdateEntry(store, nil, kKey, &entry); err != nil {
    t.Fatalf("Error updating store: %v", err)
  }
  // entry should not change as side effect
  if entry != origEntry {
    t.Errorf("Expected %v, got %v", *kAnEntry, entry)
  }
  // Key Id stored with entry
  if out := store[0].Owner; out != kKey.Id {
    t.Errorf("Expected %d, got %d", kKey.Id, out)
  }
  if store[0].UName == entry.UName || store[0].Password == entry.Password || store[0].Special == entry.Special {
    t.Error("Expected database to be encrypted.")
  }
  var readEntry vsafe.Entry
  if err := vsafedb.EntryById(store, nil, 1, kKey, &readEntry); err != nil {
    t.Fatalf("Error reading store: %v", err)
  }
  origEntry.Owner = kKey.Id
  if readEntry != origEntry {
    t.Errorf("Expected %v, got %v", origEntry, readEntry)
  }
}

func TestUpdateEntryWithEtag(t *testing.T) {
  origEntry := *kOrigEntry
  var store FakeStore
  newId, err := vsafedb.AddEntry(&store, kKey, &origEntry)
  if err != nil {
    t.Fatalf("Error saving original entry %v", err)
  }
  if err := vsafedb.EntryById(
      store, nil, newId, kKey, &origEntry); err != nil {
    t.Fatalf("Error readingoriginal entry %v", err)
  }
  tag, err := etag.Etag32(&origEntry)
  if err != nil {
    t.Fatalf("Error computing tag %v", err)
  }
  entry := *kAnEntry
  entry.Id = origEntry.Id
  if err := vsafedb.UpdateEntryWithEtag(
      store, kTransaction, tag, kKey, &entry); err != nil {
    t.Fatalf("Error updating store: %v", err)
  }
  var readEntry vsafe.Entry
  if err := vsafedb.EntryById(store, nil, newId, kKey, &readEntry); err != nil {
    t.Fatalf("Error reading store: %v", err)
  }
  entry.Owner = readEntry.Owner
  if readEntry != entry {
    t.Errorf("Expected %v, got %v", entry, readEntry)
  }
}

func TestUpdateEntryConcurrent(t *testing.T) {
  origEntry := *kOrigEntry
  var store FakeStore
  newId, err := vsafedb.AddEntry(&store, kKey, &origEntry)
  if err != nil {
    t.Fatalf("Error saving original entry %v", err)
  }
  if err := vsafedb.EntryById(
      store, nil, newId, kKey, &origEntry); err != nil {
    t.Fatalf("Error readingoriginal entry %v", err)
  }
  tag, err := etag.Etag32(&origEntry)
  if err != nil {
    t.Fatalf("Error computing tag %v", err)
  }
  entry := *kAnEntry
  entry.Id = origEntry.Id
  if err := vsafedb.UpdateEntryWithEtag(
      store,
      kTransaction,
      tag + 1,
      kKey,
      &entry); err != vsafedb.ErrConcurrentModification {
    t.Errorf("Expected ErrConcurrentModfication, got %v", err)
  }
  var readEntry vsafe.Entry
  if err := vsafedb.EntryById(store, nil, newId, kKey, &readEntry); err != nil {
    t.Fatalf("Error reading store: %v", err)
  }
  if readEntry != origEntry {
    t.Errorf("Entry should not have been updated")
  }
}

func TestUpdateEntryWithEtagBadKey(t *testing.T) {
  origEntry := *kOrigEntry
  var store FakeStore
  newId, err := vsafedb.AddEntry(&store, kKey, &origEntry)
  if err != nil {
    t.Fatalf("Error saving original entry %v", err)
  }
  if err := vsafedb.EntryById(
      store, nil, newId, kKey, &origEntry); err != nil {
    t.Fatalf("Error readingoriginal entry %v", err)
  }
  tag, err := etag.Etag32(&origEntry)
  if err != nil {
    t.Fatalf("Error computing tag %v", err)
  }
  entry := *kAnEntry
  entry.Id = origEntry.Id
  badKey := *kKey
  badKey.Id++
  if err := vsafedb.UpdateEntryWithEtag(
      store,
      kTransaction,
      tag,
      &badKey,
      &entry); err != vsafedb.ErrNoSuchId {
    t.Errorf("Expected ErrNoSuchId, got %v", err)
  }
  var readEntry vsafe.Entry
  if err := vsafedb.EntryById(store, nil, newId, kKey, &readEntry); err != nil {
    t.Fatalf("Error reading store: %v", err)
  }
  if readEntry != origEntry {
    t.Errorf("Entry should not have been updated")
  }
}

func TestEntryByIdKeyMismatch(t *testing.T) {
  var store FakeStore
  entry := *kAnEntry
  vsafedb.AddEntry(&store, kKey, &entry)
  var readEntry vsafe.Entry
  badKey := *kKey
  badKey.Id++
  if err := vsafedb.EntryById(
      store, nil, 1, &badKey, &readEntry); err != vsafedb.ErrNoSuchId {
    t.Errorf("Expected ErrNoSuchId, got %v", err)
  }
}

func TestEntries(t *testing.T) {
  yahoo, _ := url.Parse("http://www.yahoo.com")
  google, _ := url.Parse("http://www.google.com")
  var store FakeStore
  entry1 := vsafe.Entry{Title: " First", Url: yahoo, Desc: "the SeconD   oNe"}
  entry2 := vsafe.Entry{Title: "aGAiN  sEcond", Url: google, Desc: "a desc"}
  entry3 := vsafe.Entry{Title: "third again", Desc: "foo bar"}
  vsafedb.AddEntry(&store, kKey, &entry1)
  vsafedb.AddEntry(&store, kKey, &entry2)
  vsafedb.AddEntry(&store, kKey, &entry3)
  entries, err := vsafedb.Entries(store, kKey.Id, "")
  if err != nil {
    t.Fatalf("Got error fetching entries: %v", err)
  }
  if len(entries) != 3 {
    t.Errorf("Expected 3 entries, got %v", len(entries))
  }
  if entries[0].Title != entry2.Title || entries[1].Title != entry1.Title || entries[2].Title != entry3.Title {
    t.Error("Returned 3 entries in wrong order")
  }
  entries, err = vsafedb.Entries(store, kKey.Id, "  first")
  if err != nil {
    t.Fatalf("Got error fetching entries: %v", err)
  }
  if len(entries) != 1 {
    t.Errorf("Expected 1 entries, got %v", len(entries))
  }
  entries, err = vsafedb.Entries(store, kKey.Id, "second  ")
  if err != nil {
    t.Fatalf("Got error fetching entries: %v", err)
  }
  if len(entries) != 2 {
    t.Errorf("Expected 2 entries, got %v", len(entries))
  }
  entries, err = vsafedb.Entries(store, kKey.Id, "google")
  if err != nil {
    t.Fatalf("Got error fetching entries: %v", err)
  }
  if len(entries) != 1 {
    t.Errorf("Expected 1 entries, got %v", len(entries))
  }
  entries, err = vsafedb.Entries(store, kKey.Id, "biz")
  if err != nil {
    t.Fatalf("Got error fetching entries: %v", err)
  }
  if len(entries) != 0 {
    t.Errorf("Expected 0 entries, got %v", len(entries))
  }
  entries, err = vsafedb.Entries(store, kKey.Id, " eCond  one ")
  if err != nil {
    t.Fatalf("Got error fetching entries: %v", err)
  }
  if len(entries) != 1 {
    t.Errorf("Expected 1 entries, got %v", len(entries))
  }
  entries, err = vsafedb.Entries(store, kKey.Id, " Gain   SEco ")
  if err != nil {
    t.Fatalf("Got error fetching entries: %v", err)
  }
  if len(entries) != 1 {
    t.Errorf("Expected 1 entries, got %v", len(entries))
  }
  entries, err = vsafedb.Entries(store, kKey.Id, " hain   SEco ")
  if err != nil {
    t.Fatalf("Got error fetching entries: %v", err)
  }
  if len(entries) != 0 {
    t.Errorf("Expected 0 entries, got %v", len(entries))
  }
}

func TestChangePassword(t *testing.T) {
  var user vsafe.User
  if err := user.Init("foo", "password"); err != nil {
    t.Fatalf("Error initializing user %v", err)
  }
  var store FakeUserStore
  if err := store.AddUser(nil, &user); err != nil {
    t.Fatalf("Error adding user %v", err)
  }
  if  _, err := vsafedb.ChangePassword(
      store, kTransaction, user.Id + 1, "password", "board"); err != vsafedb.ErrNoSuchId {
    t.Errorf("Expected ErrNoSuchId, got %v", err)
  }
  if _, err := vsafedb.ChangePassword(
      store, kTransaction, user.Id, "wrong", "board"); err != vsafe.ErrWrongPassword {
    t.Errorf("Expected ErrWrongPassword, got %v", err)
  }
  newUser, err := vsafedb.ChangePassword(
      store, kTransaction, user.Id, "password", "board")
  if err != nil {
    t.Errorf("Expected no error, got %v", err)
  }
  if _, err := newUser.VerifyPassword("board"); err != nil {
    t.Errorf("Got error verifying password, %v", err)
  }
  var readUser vsafe.User
  if err := store.UserById(nil, user.Id, &readUser); err != nil {
    t.Fatalf("Got error reading database, %v", err)
  }
  if _, err := readUser.VerifyPassword("board"); err != nil {
    t.Errorf("Got error verifying password, %v", err)
  }
}

type FakeUserStore []*vsafe.User

func (f *FakeUserStore) AddUser(t db.Transaction, u *vsafe.User) error {
  u.Id = int64(len(*f) + 1)
  stored := *u
  *f = append(*f, &stored)
  return nil
}
  
func (f FakeUserStore) UpdateUser(
    t db.Transaction, u *vsafe.User) error {
  stored := *u
  f[stored.Id - 1] = &stored
  return nil
}

func (f FakeUserStore) UserById(
    t db.Transaction, id int64, u *vsafe.User) error {
  if int(id) > len(f) {
    return vsafedb.ErrNoSuchId
  }
  *u = *f[id - 1]
  return nil
}

type FakeStore []*vsafe.Entry

func (f *FakeStore) AddEntry(t db.Transaction, e *vsafe.Entry) error {
  e.Id = int64(len(*f) + 1)
  stored := *e
  *f = append(*f, &stored)
  return nil
}
  
func (f FakeStore) UpdateEntry(t db.Transaction, e *vsafe.Entry) error {
  stored := *e
  f[stored.Id - 1] = &stored
  return nil
}

func (f FakeStore) EntryById(t db.Transaction, id int64, e *vsafe.Entry) error {
  if int(id) > len(f) {
    return vsafedb.ErrNoSuchId
  }
  *e = *f[id - 1]
  return nil
}

func (f FakeStore) EntriesByOwner(
    t db.Transaction, owner int64, consumer functional.Consumer) error {
  s := functional.NewStreamFromPtrs(f, nil)
  s = functional.Filter(functional.NewFilterer(func(ptr interface{}) error {
    p := ptr.(*vsafe.Entry)
    if p.Owner == owner {
      return nil
    }
    return functional.Skipped
  }),
  s)
  return consumer.Consume(s)
}
