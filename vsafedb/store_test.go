package vsafedb_test

import (
	"github.com/keep94/consume"
	"github.com/keep94/toolbox/db"
	"github.com/keep94/toolbox/kdf"
	"github.com/keep94/vsafe"
	"github.com/keep94/vsafe/vsafedb"
	"net/url"
	"testing"
)

var (
	kAnEntry = &vsafe.Entry{
		UName:    "uname",
		Password: "password",
		Special:  "special",
	}
	kOrigEntry = &vsafe.Entry{
		UName:    "somename",
		Password: "somethingelse",
		Special:  "xxx",
	}
	kKey                        = &vsafe.Key{Id: 7, Value: kdf.Random(32)}
	kTransaction db.Transaction = 0
)

func TestUpdateCategory(t *testing.T) {
	store := &FakeCategoryStore{
		Category: &vsafe.Category{Id: 5, Owner: 3, Name: "five"}}
	// wrong Id throws ErrNoSuchId
	_, err := vsafedb.UpdateCategory(store, kTransaction, 2, 3, "updated")
	if err != vsafedb.ErrNoSuchId {
		t.Error("Expected ErrNoSuchId")
	}

	// Wrong owner throws ErrNoSuchId
	_, err = vsafedb.UpdateCategory(store, kTransaction, 5, 2, "updated")
	if err != vsafedb.ErrNoSuchId {
		t.Error("Expected ErrNoSuchId")
	}

	if store.Category.Name != "five" {
		t.Error("Expected category to remain unchanged")
	}

	oldName, err := vsafedb.UpdateCategory(
		store, kTransaction, 5, 3, "updated")
	if err != nil {
		t.Fatal("Got error updating category")
	}
	if oldName != "five" {
		t.Error("Expected old name to be five")
	}

	if store.Category.Name != "updated" {
		t.Error("Expected category to be updated")
	}
}

func TestRemoveCategory(t *testing.T) {
	store := &FakeCategoryStore{
		Category: &vsafe.Category{Id: 5, Owner: 3, Name: "five"}}
	// wrong Id throws ErrNoSuchId
	_, err := vsafedb.RemoveCategory(store, kTransaction, 2, 3)
	if err != vsafedb.ErrNoSuchId {
		t.Error("Expected ErrNoSuchId")
	}

	// Wrong owner throws ErrNoSuchId
	_, err = vsafedb.RemoveCategory(store, kTransaction, 5, 2)
	if err != vsafedb.ErrNoSuchId {
		t.Error("Expected ErrNoSuchId")
	}

	if store.Category.Name != "five" {
		t.Error("Expected category to remain unchanged")
	}

	oldName, err := vsafedb.RemoveCategory(store, kTransaction, 5, 3)
	if err != nil {
		t.Fatal("Got error removing category")
	}
	if oldName != "five" {
		t.Error("Expected old name to be five")
	}

	if store.Category != nil {
		t.Error("Expected category to be removed")
	}
}

func TestAddEntry(t *testing.T) {
	var store FakeStore
	entry := *kAnEntry
	var id int64
	var err error
	if id, err = vsafedb.AddEntry(&store, nil, kKey, &entry); err != nil {
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
	origEntry.Etag = readEntry.Etag
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
	origEntry.Etag = readEntry.Etag
	if readEntry != origEntry {
		t.Errorf("Expected %v, got %v", origEntry, readEntry)
	}
}

func TestUpdateEntryWithEtag(t *testing.T) {
	origEntry := *kOrigEntry
	var store FakeStore
	newId, err := vsafedb.AddEntry(&store, nil, kKey, &origEntry)
	if err != nil {
		t.Fatalf("Error saving original entry %v", err)
	}
	var origEntryWithEtag vsafe.Entry
	if err := vsafedb.EntryById(
		store, nil, newId, kKey, &origEntryWithEtag); err != nil {
		t.Fatalf("Error reading original entry %v", err)
	}
	if err := vsafedb.UpdateEntryWithEtag(
		store,
		kTransaction,
		newId,
		origEntryWithEtag.Etag,
		kKey,
		changeToAnEntry); err != nil {
		t.Fatalf("Error updating store: %v", err)
	}
	var readEntry vsafe.Entry
	if err := vsafedb.EntryById(store, nil, newId, kKey, &readEntry); err != nil {
		t.Fatalf("Error reading store: %v", err)
	}
	entry := *kAnEntry
	entry.Owner = readEntry.Owner
	entry.Id = readEntry.Id
	entry.Etag = readEntry.Etag
	if readEntry != entry {
		t.Errorf("Expected %v, got %v", entry, readEntry)
	}
}

func TestUpdateEntryConcurrent(t *testing.T) {
	origEntry := *kOrigEntry
	var store FakeStore
	newId, err := vsafedb.AddEntry(&store, nil, kKey, &origEntry)
	if err != nil {
		t.Fatalf("Error saving original entry %v", err)
	}
	var origEntryWithEtag vsafe.Entry
	if err := vsafedb.EntryById(
		store, nil, newId, kKey, &origEntryWithEtag); err != nil {
		t.Fatalf("Error reading original entry %v", err)
	}
	update := changeToAnEntry
	updateSkipped := func(entryPtr *vsafe.Entry) bool {
		*entryPtr = *kAnEntry
		return false
	}
	// An update that skips shouldn't throw an error even if etag is wrong
	if err := vsafedb.UpdateEntryWithEtag(
		store,
		kTransaction,
		newId,
		origEntryWithEtag.Etag+1,
		kKey,
		updateSkipped); err != nil {
		t.Fatalf("Error updating store: %v", err)
	}
	if err := vsafedb.UpdateEntryWithEtag(
		store,
		kTransaction,
		newId,
		origEntryWithEtag.Etag+1,
		kKey,
		update); err != vsafedb.ErrConcurrentModification {
		t.Errorf("Expected ErrConcurrentModfication, got %v", err)
	}
	var readEntry vsafe.Entry
	if err := vsafedb.EntryById(store, nil, newId, kKey, &readEntry); err != nil {
		t.Fatalf("Error reading store: %v", err)
	}
	if readEntry != origEntryWithEtag {
		t.Errorf("Entry should not have been updated")
	}
}

func TestUpdateEntryWithEtagBadKey(t *testing.T) {
	origEntry := *kOrigEntry
	var store FakeStore
	newId, err := vsafedb.AddEntry(&store, nil, kKey, &origEntry)
	if err != nil {
		t.Fatalf("Error saving original entry %v", err)
	}
	var origEntryWithEtag vsafe.Entry
	if err := vsafedb.EntryById(
		store, nil, newId, kKey, &origEntryWithEtag); err != nil {
		t.Fatalf("Error readingoriginal entry %v", err)
	}
	badKey := *kKey
	badKey.Id++
	if err := vsafedb.UpdateEntryWithEtag(
		store,
		kTransaction,
		newId,
		origEntryWithEtag.Etag,
		&badKey,
		changeToAnEntry); err != vsafedb.ErrNoSuchId {
		t.Errorf("Expected ErrNoSuchId, got %v", err)
	}
	var readEntry vsafe.Entry
	if err := vsafedb.EntryById(store, nil, newId, kKey, &readEntry); err != nil {
		t.Fatalf("Error reading store: %v", err)
	}
	if readEntry != origEntryWithEtag {
		t.Errorf("Entry should not have been updated")
	}
}

func TestEntryByIdKeyMismatch(t *testing.T) {
	var store FakeStore
	entry := *kAnEntry
	vsafedb.AddEntry(&store, nil, kKey, &entry)
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
	entry3 := vsafe.Entry{Title: "third again", Desc: "foo bar", Categories: "17"}
	vsafedb.AddEntry(&store, nil, kKey, &entry1)
	vsafedb.AddEntry(&store, nil, kKey, &entry2)
	vsafedb.AddEntry(&store, nil, kKey, &entry3)
	entries, err := vsafedb.Entries(store, kKey.Id, "", 0)
	if err != nil {
		t.Fatalf("Got error fetching entries: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %v", len(entries))
	}
	if entries[0].Title != entry1.Title || entries[1].Title != entry2.Title || entries[2].Title != entry3.Title {
		t.Error("Returned 3 entries in wrong order")
	}
	entries, err = vsafedb.Entries(store, kKey.Id, "  first", 0)
	if err != nil {
		t.Fatalf("Got error fetching entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entries, got %v", len(entries))
	}
	entries, err = vsafedb.Entries(store, kKey.Id, "second  ", 0)
	if err != nil {
		t.Fatalf("Got error fetching entries: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %v", len(entries))
	}
	entries, err = vsafedb.Entries(store, kKey.Id, "google", 0)
	if err != nil {
		t.Fatalf("Got error fetching entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entries, got %v", len(entries))
	}
	entries, err = vsafedb.Entries(store, kKey.Id, "biz", 0)
	if err != nil {
		t.Fatalf("Got error fetching entries: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries, got %v", len(entries))
	}
	entries, err = vsafedb.Entries(store, kKey.Id, " eCond  one ", 0)
	if err != nil {
		t.Fatalf("Got error fetching entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entries, got %v", len(entries))
	}
	entries, err = vsafedb.Entries(store, kKey.Id, " Gain   SEco ", 0)
	if err != nil {
		t.Fatalf("Got error fetching entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entries, got %v", len(entries))
	}
	entries, err = vsafedb.Entries(store, kKey.Id, " hain   SEco ", 0)
	if err != nil {
		t.Fatalf("Got error fetching entries: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries, got %v", len(entries))
	}
	entries, err = vsafedb.Entries(store, kKey.Id, "", 17)
	if err != nil {
		t.Fatalf("Got error fetching entries: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %v", len(entries))
	}
	entries, err = vsafedb.Entries(store, kKey.Id, "", 16)
	if err != nil {
		t.Fatalf("Got error fetching entries: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 enties, got %v", len(entries))
	}
}

func TestSortByTitle(t *testing.T) {
	entry1 := vsafe.Entry{Title: " First"}
	entry2 := vsafe.Entry{Title: "aGAiN  sEcond"}
	entry3 := vsafe.Entry{Title: "   Third", Desc: "foo bar"}
	entry4 := vsafe.Entry{Title: "fourth again", Desc: "foo bar"}
	s := []*vsafe.Entry{&entry1, &entry2, &entry3, &entry4}
	vsafedb.SortByTitle(s)
	if s[0] != &entry2 || s[1] != &entry1 || s[2] != &entry4 || s[3] != &entry3 {
		t.Error("Sort in wrong order.")
	}
}

func TestSortByTitleEmpty(t *testing.T) {
	var s []*vsafe.Entry
	vsafedb.SortByTitle(s)
}

func TestReverse(t *testing.T) {
	var entry1, entry2, entry3, entry4 vsafe.Entry
	var s []*vsafe.Entry
	vsafedb.Reverse(s)
	s = []*vsafe.Entry{&entry1}
	vsafedb.Reverse(s)
	if s[0] != &entry1 {
		t.Error("reverse with 1 entry wrong.")
	}
	s = []*vsafe.Entry{&entry1, &entry2}
	vsafedb.Reverse(s)
	if s[0] != &entry2 || s[1] != &entry1 {
		t.Error("reverse with 2 entries wrong.")
	}
	s = []*vsafe.Entry{&entry1, &entry2, &entry3}
	vsafedb.Reverse(s)
	if s[0] != &entry3 || s[1] != &entry2 || s[2] != &entry1 {
		t.Error("reverse with 3 entries wrong.")
	}
	s = []*vsafe.Entry{&entry1, &entry2, &entry3, &entry4}
	vsafedb.Reverse(s)
	if s[0] != &entry4 || s[1] != &entry3 || s[2] != &entry2 || s[3] != &entry1 {
		t.Error("reverse with 4 entries wrong.")
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
	if _, err := vsafedb.ChangePassword(
		store, kTransaction, user.Id+1, "password", "board"); err != vsafedb.ErrNoSuchId {
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
	f[stored.Id-1] = &stored
	return nil
}

func (f FakeUserStore) UserById(
	t db.Transaction, id int64, u *vsafe.User) error {
	if int(id) > len(f) {
		return vsafedb.ErrNoSuchId
	}
	*u = *f[id-1]
	return nil
}

type FakeCategoryStore struct {
	Category *vsafe.Category
}

func (f *FakeCategoryStore) CategoryById(
	t db.Transaction, id int64, c *vsafe.Category) error {
	if f.Category == nil || f.Category.Id != id {
		return vsafedb.ErrNoSuchId
	}
	*c = *f.Category
	return nil
}

func (f *FakeCategoryStore) UpdateCategory(
	t db.Transaction, c *vsafe.Category) error {
	if f.Category == nil || f.Category.Id != c.Id {
		return nil
	}
	category := *c
	f.Category = &category
	return nil
}

func (f *FakeCategoryStore) RemoveCategory(
	t db.Transaction, id int64) error {
	if f.Category == nil || f.Category.Id != id {
		return nil
	}
	f.Category = nil
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
	f[stored.Id-1] = &stored
	return nil
}

func (f FakeStore) EntryById(t db.Transaction, id int64, e *vsafe.Entry) error {
	if int(id) > len(f) {
		return vsafedb.ErrNoSuchId
	}
	*e = *f[id-1]
	e.Etag = 57
	return nil
}

func (f FakeStore) EntriesByOwner(
	t db.Transaction, owner int64, consumer consume.Consumer) error {
	for _, entry := range f {
		if !consumer.CanConsume() {
			break
		}
		if entry.Owner != owner {
			continue
		}
		entryCopy := *entry
		consumer.Consume(&entryCopy)
	}
	return nil
}

func changeToAnEntry(entryPtr *vsafe.Entry) bool {
	*entryPtr = *kAnEntry
	return true
}
