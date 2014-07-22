// Package vsafedb handles storing instances in the vsafe app to persistent
// storage.
package vsafedb

import (
  "errors"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/etag"
  "github.com/keep94/appcommon/str_util"
  "github.com/keep94/gofunctional3/consume"
  "github.com/keep94/gofunctional3/functional"
  "github.com/keep94/vsafe"
  "sort"
  "strings"
)

var (
  // Indicates that the id does not exist in the database.
  ErrNoSuchId = errors.New("vsafedb: No such Id.")
  // Indicates concurrent modification
  ErrConcurrentModification = errors.New("vsafedb: Concurrent Modification")
)

type AddUserRunner interface {
  // AddUser adds a new user to persistent storage.
  AddUser(t db.Transaction, user *vsafe.User) error
}

type UserByIdRunner interface {
  // UserById retrieves a user by id from persistent storage.
  UserById(t db.Transaction, id int64, user *vsafe.User) error
}

type UserByNameRunner interface {
  // UserByName retrieves a user by name from persistent storage.
  UserByName(t db.Transaction, name string, user *vsafe.User) error
}

type UsersRunner interface {
  // Users retrieves all users from persistent storage ordered by name.
  Users(t db.Transaction, consumer functional.Consumer) error
}

type UpdateUserRunner interface {
  // UpdateUser modifies a user in persistent storage.
  UpdateUser(t db.Transaction, user *vsafe.User) error
}

type SafeUpdateUserRunner interface {
  UpdateUserRunner
  UserByIdRunner
}

type RemoveUserRunner interface {
  // RemoveUser removes a user by name from persistent storage.
  RemoveUser(t db.Transaction, name string) error
}

type AddEntryRunner interface {
  // AddEntryRunner adds a new entry to persistent storage.
  AddEntry(t db.Transaction, entry *vsafe.Entry) error
}

type EntryByIdRunner interface {
  // EntryById retrieves an entry by id from persistent storage.
  EntryById(t db.Transaction, id int64, entry *vsafe.Entry) error
}

type EntriesByOwnerRunner interface {
  // EntriesByOwner retrieves all entries with a particular owner from
  // persistent storage ordered by Id.
  EntriesByOwner(t db.Transaction, owner int64, consumer functional.Consumer) error
}

type UpdateEntryRunner interface {
  // UpdateEntry updates an entry in persistent storage.
  UpdateEntry(t db.Transaction, entry *vsafe.Entry) error
}

type SafeUpdateEntryRunner interface {
  UpdateEntryRunner
  EntryByIdRunner
}

type RemoveEntryRunner interface {
  // RemoveEntry removes an entry with given id and owner from persistent
  // storage.
  RemoveEntry(t db.Transaction, id, owner int64) error
}

// AddEntry adds a new entry to persistent storage so that sensitive fields
// are encrypted in persistent storage.
func AddEntry(
    store AddEntryRunner,
    t db.Transaction,
    key *vsafe.Key,
    entry *vsafe.Entry) (newId int64, err error) {
  encrypted := *entry
  if err = encrypted.Encrypt(key); err != nil {
    return
  }
  if err = store.AddEntry(t, &encrypted); err != nil {
    return
  }
  return encrypted.Id, nil
}

// UpdateEntryWithEtag updates an entry in persistent storage in a way that
// detects concurrent modification. It also prevents users from modifying
// entries they do not own by returning ErrNoSuchId. t, the transaction,
// must be non nil.
func UpdateEntryWithEtag(
    store SafeUpdateEntryRunner,
    t db.Transaction,
    tag uint32,
    key *vsafe.Key,
    entry *vsafe.Entry) error {
  if t == nil {
    panic("Transaction must be non-nil")
  }
  var origEntry vsafe.Entry
  err := EntryById(store, t, entry.Id, key, &origEntry)
  if err != nil {
   return err
  }
  origTag, err := etag.Etag32(&origEntry)
  if tag != origTag {
    return ErrConcurrentModification
  }
  return UpdateEntry(store, t, key, entry)
}

// UpdateEntry updates an entry in persistent storage so that sensitive fields
// remain encrypted in persistent storage. Beware: UpdateEntry does not
// protect against concurrent modification nor does it protect against 
// users clobbering an entry they do not own.
func UpdateEntry(
    store UpdateEntryRunner,
    t db.Transaction,
    key *vsafe.Key,
    entry *vsafe.Entry) (err error) {
  encrypted := *entry
  if err = encrypted.Encrypt(key); err != nil {
    return
  }
  if err = store.UpdateEntry(t, &encrypted); err != nil {
    return
  }
  return
}

// EntryById retrieves an entry by its id from persistent storage while
// handling decryption of sensitive fields. If the Id of the provided key
// does not match the Owner field of fetched entry, that is the current user
// does not own the entry being fetched, EntryById returns ErrNoSuchId.
func EntryById(
    store EntryByIdRunner,
    t db.Transaction,
    id int64,
    key *vsafe.Key,
    entry *vsafe.Entry) (err error) {
  if err = store.EntryById(t, id, entry); err != nil {
    return
  }
  if err = entry.Decrypt(key); err != nil {
    if err == vsafe.ErrKeyMismatch {
      err = ErrNoSuchId
    }
    return
  }
  return nil
}

// Entries returns a new slice containing entries encrypted with keyId and
// matching query and orders them by Id. It does not decrypt the sensitive
// fields within the fetched entries. query is searched for within url,
// title, and description of each entry ignoring case to determine whether or
// not there is a match. Whitespace within query and entry fields are
// normalized to a single space before matching happens. The empty string
// matches all entries.
func Entries(
    store EntriesByOwnerRunner,
    keyId int64,
    query string) ([]*vsafe.Entry, error) {
  var results []*vsafe.Entry
  if err := store.EntriesByOwner(
      nil,
      keyId,
      functional.FilterConsumer(
          consume.AppendPtrsTo(&results, nil),
          newEntryFilter(query))); err != nil {
    return nil, err
  }
  return results, nil
}

// SortByTitle sorts entries by title in place ignoring case.
func SortByTitle(entries []*vsafe.Entry) {
  sort.Sort(newSortByTitle(entries))
}

// Reverse reverses entries in place.
func Reverse(entries []*vsafe.Entry) {
  start := 0
  end := len(entries) - 1
  for start < end {
    entries[start], entries[end] = entries[end], entries[start]
    start++
    end--
  }
}

// ChangePassword changes the password of a user in persistent storage.
// t, the transaction, must be non nil.
func ChangePassword(
    store SafeUpdateUserRunner,
    t db.Transaction,
    id int64,
    oldPass, newPass string) (*vsafe.User, error) {
  if t == nil {
    panic("Transaction must be non-nil")
  }
  var user vsafe.User
  err := store.UserById(t, id, &user)
  if err != nil {
    return nil, err 
  }   
  if err = user.ChangePassword(oldPass, newPass); err != nil {
    return nil, err 
  }   
  if err = store.UpdateUser(t, &user); err != nil {
    return nil, err 
  }   
  return &user, nil
}

type entryFilter string

func newEntryFilter(s string) functional.Filterer {
  s = str_util.Normalize(s)
  if s == "" {
    return functional.All()
  }
  return entryFilter(s)
}

func (f entryFilter) Filter(ptr interface{}) error {
  p := ptr.(*vsafe.Entry)
  pattern := string(f)
  if p.Url != nil {
    str := str_util.Normalize(p.Url.String())
    if strings.Index(str, pattern) != -1 {
      return nil
    }
  }
  if strings.Index(str_util.Normalize(p.Title), pattern) != -1 {
    return nil
  }
  if strings.Index(str_util.Normalize(p.Desc), pattern) != -1 {
    return nil
  }
  return functional.Skipped
}

type sortByTitle struct {
  entries []*vsafe.Entry
  trimmedLowerTitles []string
}

func newSortByTitle(entries []*vsafe.Entry) sort.Interface {
  titles := make([]string, len(entries))
  for i := range entries {
    titles[i] = strings.TrimSpace(strings.ToLower(entries[i].Title))
  }
  return &sortByTitle{entries: entries, trimmedLowerTitles: titles}
}

func (s *sortByTitle) Len() int {
  return len(s.entries)
}

func (s *sortByTitle) Less(i, j int) bool {
  return s.trimmedLowerTitles[i] < s.trimmedLowerTitles[j]
}

func (s *sortByTitle) Swap(i, j int) {
  s.entries[i], s.entries[j] = s.entries[j], s.entries[i]
  s.trimmedLowerTitles[i], s.trimmedLowerTitles[j] =
      s.trimmedLowerTitles[j], s.trimmedLowerTitles[i]
}

