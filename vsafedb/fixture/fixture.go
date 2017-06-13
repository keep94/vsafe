// Package fixture provides test suites to test implementations of the
// datastore interfaces in the vsafedb package.
package fixture

import (
  "github.com/keep94/gofunctional3/consume"
  "github.com/keep94/vsafe"
  "github.com/keep94/vsafe/vsafedb"
  "net/url"
  "reflect"
  "testing"
)

const (
  kBadId = 9998
  kBadName = "Bad Name"
  kOwner = 5
)

var (
  url1 = mustParse("http://www.yahoo.com")
  url2 = mustParse("http://www.google.com")
)

var (
  kFirstUser = &vsafe.User{
      Name: "foo",
      Key: "bar",
      Checksum: "baz",
  }
  kSecondUser = &vsafe.User{
      Name: "blow",
      Key: "slow",
      Checksum: "mow",
  }
  kFirstEntry = &vsafe.Entry{
    Owner: kOwner,
    Url: url1,
    Title: "zbar",
    Desc: "baz",
    UName: "keep94",
    Password: "password",
    Special: "special",
    Categories: "3,7",
  }

  kSecondEntry = &vsafe.Entry{
    Owner: kOwner,
    Url: url2,
    Title: "title",
    Desc: "desc",
    UName: "keep95",
    Password: "loco",
    Special: "never",
    Categories: "2",
  }
)

type UserByIdStore interface {
  vsafedb.AddUserRunner
  vsafedb.UserByIdRunner
}

type UserByNameStore interface {
  vsafedb.AddUserRunner
  vsafedb.UserByNameRunner
}

type UsersStore interface {
  vsafedb.AddUserRunner
  vsafedb.UsersRunner
}

type UpdateUserStore interface {
  UserByIdStore
  vsafedb.UpdateUserRunner
}

type RemoveUserStore interface {
  UserByIdStore
  vsafedb.RemoveUserRunner
}

type CategoriesByOwnerStore interface {
  vsafedb.AddCategoryRunner
  vsafedb.CategoriesByOwnerRunner
}

type UpdateCategoryNameStore interface {
  CategoriesByOwnerStore
  vsafedb.UpdateCategoryNameRunner
}

type RemoveCategoryStore interface {
  CategoriesByOwnerStore
  vsafedb.RemoveCategoryRunner
}

type EntryByIdStore interface {
  vsafedb.AddEntryRunner
  vsafedb.EntryByIdRunner
}

type EntriesByOwnerStore interface {
  vsafedb.AddEntryRunner
  vsafedb.EntriesByOwnerRunner
}

type UpdateEntryStore interface {
  EntryByIdStore
  vsafedb.UpdateEntryRunner
}

type RemoveEntryStore interface {
  EntryByIdStore
  vsafedb.RemoveEntryRunner
}

func UserById(t *testing.T, store UserByIdStore) {
  var first, second vsafe.User
  var firstResult, secondResult vsafe.User
  createUsers(t, store, &first, &second)
  if err := store.UserById(nil, first.Id, &firstResult); err != nil {
    t.Fatalf("Got error reading database by id: %v", err)
  }
  if err := store.UserById(nil, second.Id, &secondResult); err != nil {
    t.Fatalf("Got error reading database by id: %v", err)
  }
  assertUserEqual(t, &first, &firstResult)
  assertUserEqual(t, &second, &secondResult)
  if err := store.UserById(nil, kBadId, &firstResult); err != vsafedb.ErrNoSuchId {
    t.Errorf("Expected ErrNoSuchId, got %v", err)
  }
}

func UserByName(t *testing.T, store UserByNameStore) {
  var first, second vsafe.User
  var firstResult, secondResult vsafe.User
  createUsers(t, store, &first, &second)
  if err := store.UserByName(nil, first.Name, &firstResult); err != nil {
    t.Fatalf("Got error reading database by id: %v", err)
  }
  if err := store.UserByName(nil, second.Name, &secondResult); err != nil {
    t.Fatalf("Got error reading database by id: %v", err)
  }
  assertUserEqual(t, &first, &firstResult)
  assertUserEqual(t, &second, &secondResult)
  if err := store.UserByName(nil, kBadName, &firstResult); err != vsafedb.ErrNoSuchId {
    t.Errorf("Expected ErrNoSuchId, got %v", err)
  }
}

func Users(t *testing.T, store UsersStore) {
  var first, second vsafe.User
  var users []*vsafe.User
  createUsers(t, store, &first, &second)
  if err := store.Users(nil, consume.AppendPtrsTo(&users, nil)); err != nil {
    t.Fatalf("Got error reading database: %v", err)
  }
  assertUserEqual(t, &second, users[0])
  assertUserEqual(t, &first, users[1])
  if out := len(users); out != 2 {
    t.Errorf("Expected 2, got %d", out)
  }
}

func UpdateUser(t *testing.T, store UpdateUserStore) {
  var first, second vsafe.User
  var firstResult, secondResult vsafe.User
  createUsers(t, store, &first, &second)
  first.Name = "John Doe"
  first.Key = "John Doe Key"
  if err := store.UpdateUser(nil, &first); err != nil {
    t.Fatalf("Got error updating user: %v", err)
  }
  if err := store.UserById(nil, first.Id, &firstResult); err != nil {
    t.Fatalf("Got error reading database by id: %v", err)
  }
  if err := store.UserById(nil, second.Id, &secondResult); err != nil {
    t.Fatalf("Got error reading database by id: %v", err)
  }
  assertUserEqual(t, &first, &firstResult)
  assertUserEqual(t, &second, &secondResult)
  if err := store.UserById(nil, kBadId, &firstResult); err != vsafedb.ErrNoSuchId {
    t.Errorf("Expected ErrNoSuchId, got %v", err)
  }
}

func RemoveUser(t *testing.T, store RemoveUserStore) {
  var first, second vsafe.User
  var firstResult vsafe.User
  createUsers(t, store, &first, &second)
  if err := store.RemoveUser(nil, first.Name); err != nil {
    t.Fatalf("Got error removing by id: %v", err)
  }
  if err := store.UserById(nil, first.Id, &firstResult); err != vsafedb.ErrNoSuchId {
    t.Errorf("Expected ErrNoSuchId, got %v", err)
  }
  if err := store.UserById(nil, second.Id, &firstResult); err != nil {
    t.Errorf("Got error reading database: %v", err)
  }
}

func UserDupName(t *testing.T, store vsafedb.AddUserRunner) {
  u := vsafe.User{Name: "foo"}
  if err := store.AddUser(nil, &u); err != nil {
    t.Fatal("Error adding user.")
  }
  if err := store.AddUser(nil, &u); err == nil {
    t.Error("Expected error adding duplicate user name.")
  }
}

func CategoriesByOwner(t *testing.T, store CategoriesByOwnerStore) {
  createCategories(t, store)
  categories, err := store.CategoriesByOwner(nil, 1)
  if err != nil {
    t.Fatalf("Got error reading categories: %v", err)
  }
  // sorted by name
  assertCategoryNames(t, categories, "one", "three", "two")
}

func UpdateCategoryName(t *testing.T, store UpdateCategoryNameStore) {
  createCategories(t, store)
  // 3 corresponds to category "three"
  if err := store.UpdateCategoryName(nil, 3, 1, "updated"); err != nil {
    t.Fatalf("Got error updating category: %v", err)
  }
  // wrong owner doesn't update
  if err := store.UpdateCategoryName(nil, 3, 2, "wrong"); err != nil {
    t.Fatalf("Got error updating category: %v", err)
  }
  categories, err := store.CategoriesByOwner(nil, 1)
  if err != nil {
    t.Fatalf("Got error reading categories: %v", err)
  }
  assertCategoryNames(t, categories, "one", "two", "updated")
}

func RemoveCategory(t *testing.T, store RemoveCategoryStore) {
  createCategories(t, store)
  if err := store.RemoveCategory(nil, 3, 1); err != nil {
    t.Fatalf("Got error removing category: %v", err)
  }
  // wrong owner doesn't remove
  if err := store.RemoveCategory(nil, 2, 2); err != nil {
    t.Fatalf("Got error removing category: %v", err)
  }
  categories, err := store.CategoriesByOwner(nil, 1)
  if err != nil {
    t.Fatalf("Got error reading categories: %v", err)
  }
  assertCategoryNames(t, categories, "one", "two")
}

func EntryById(t *testing.T, store EntryByIdStore) {
  var first, second vsafe.Entry
  var firstResult, secondResult vsafe.Entry
  createEntries(t, store, &first, &second)
  if err := store.EntryById(nil, first.Id, &firstResult); err != nil {
    t.Fatalf("Got error reading database by id: %v", err)
  }
  if err := store.EntryById(nil, second.Id, &secondResult); err != nil {
    t.Fatalf("Got error reading database by id: %v", err)
  }
  assertEntryEqual(t, &first, &firstResult)
  assertEntryEqual(t, &second, &secondResult)
  if err := store.EntryById(nil, kBadId, &firstResult); err != vsafedb.ErrNoSuchId {
    t.Errorf("Expected ErrNoSuchId, got %v", err)
  }
}

func EntriesByOwner(t *testing.T, store EntriesByOwnerStore) {
  var first, second vsafe.Entry
  createEntries(t, store, &first, &second)
  e := vsafe.Entry{Owner: kOwner + 1}
  if err := store.AddEntry(nil, &e); err != nil {
    t.Fatalf("Got error adding entry %v", err)
  }
  var entries []*vsafe.Entry
  if err := store.EntriesByOwner(nil, kOwner, consume.AppendPtrsTo(&entries, nil)); err != nil {
    t.Fatalf("Got error reading database by id: %v", err)
  }
  if len(entries) != 2 {
    t.Fatalf("Expected 2 entries back, got %d", len(entries))
  }
  assertEntryEqual(t, &first, entries[0])
  assertEntryEqual(t, &second, entries[1])
}

func UpdateEntry(t *testing.T, store UpdateEntryStore) {
  var first, second vsafe.Entry
  var firstResult, secondResult vsafe.Entry
  createEntries(t, store, &first, &second)
  first.Owner = 23
  first.Url = url2
  first.Title = "abar"
  first.Desc = "aaa"
  first.UName = "back"
  first.Password = "aardvark"
  first.Special = "new again"
  if err := store.UpdateEntry(nil, &first); err != nil {
    t.Fatalf("Got error updating database: %v", err)
  }
  if err := store.EntryById(nil, first.Id, &firstResult); err != nil {
    t.Fatalf("Got error reading database by id: %v", err)
  }
  if err := store.EntryById(nil, second.Id, &secondResult); err != nil {
    t.Fatalf("Got error reading database by id: %v", err)
  }
  assertEntryEqual(t, &first, &firstResult)
  assertEntryEqual(t, &second, &secondResult)
}

func RemoveEntry(t *testing.T, store RemoveEntryStore) {
  var first, second vsafe.Entry
  var firstResult vsafe.Entry
  createEntries(t, store, &first, &second)
  if err := store.RemoveEntry(nil, first.Id, first.Owner); err != nil {
    t.Fatalf("Got error removing by id: %v", err)
  }
  if err := store.RemoveEntry(nil, second.Id, second.Owner + 1); err != nil {
    t.Fatalf("Got error removing by id: %v", err)
  }
  if err := store.EntryById(nil, first.Id, &firstResult); err != vsafedb.ErrNoSuchId {
    t.Errorf("Expected ErrNoSuchId, got %v", err)
  }
  if err := store.EntryById(nil, second.Id, &firstResult); err != nil {
    t.Errorf("Got error reading database: %v", err)
  }
}

func createUsers(
    t *testing.T,
    store vsafedb.AddUserRunner,
    first *vsafe.User,
    second *vsafe.User) {
  createUser(t, store, kFirstUser, first)
  createUser(t, store, kSecondUser, second)
}

func createUser(
    t *testing.T,
    store vsafedb.AddUserRunner,
    toBeAdded *vsafe.User,
    result *vsafe.User) {
  *result = *toBeAdded
  if err := store.AddUser(nil, result); err != nil {
    t.Fatalf("Got %v adding to store", err)
  }
  if result.Id == 0 {
    t.Error("Expected Id to be set.")
  }
}

func createCategories(t *testing.T, store vsafedb.AddCategoryRunner) {
  cat := vsafe.Category{Name: "one", Owner: 1}
  createCategory(t, store, &cat)
  cat = vsafe.Category{Name: "two", Owner: 1}
  createCategory(t, store, &cat)
  cat = vsafe.Category{Name: "three", Owner: 1}
  createCategory(t, store, &cat)
  cat = vsafe.Category{Name: "wrong", Owner: 2}
  createCategory(t, store, &cat)
}

func createCategory(
     t *testing.T,
     store vsafedb.AddCategoryRunner,
     toBeAdded *vsafe.Category) {
  if err := store.AddCategory(nil, toBeAdded); err != nil {
    t.Fatalf("Got %v adding to category", err)
  }
  if toBeAdded.Id == 0 {
    t.Error("Expected Id to be set.")
  }
}

func createEntries(
    t *testing.T,
    store vsafedb.AddEntryRunner,
    first *vsafe.Entry,
    second *vsafe.Entry) {
  createEntry(t, store, kFirstEntry, first)
  createEntry(t, store, kSecondEntry, second)
}

func createEntry(
    t *testing.T,
    store vsafedb.AddEntryRunner,
    toBeAdded *vsafe.Entry,
    result *vsafe.Entry) {
  *result = *toBeAdded
  if err := store.AddEntry(nil, result); err != nil {
    t.Fatalf("Got %v adding to store", err)
  }
  if result.Id == 0 {
    t.Error("Expected Id to be set.")
  }
}

func assertCategoryNames(
    t *testing.T, categories []vsafe.Category, expected ...string) {
  catlen, explen := len(categories), len(expected)
  if catlen != explen {
    t.Errorf("Expected %d categories, got %d", explen, catlen)
    return
  }
  for i := range categories {
    if categories[i].Name != expected[i] {
      t.Errorf("Expected %s, got %s", expected[i], categories[i].Name)
    }
  }
}

func assertUserEqual(t *testing.T, expected, actual *vsafe.User) {
  if !reflect.DeepEqual(expected, actual) {
    t.Errorf("Expected %v, got %v", expected, actual)
  }
}

func assertEntryEqual(t *testing.T, expected, actual *vsafe.Entry) {
  if !reflect.DeepEqual(expected, actual) {
    t.Errorf("Expected %v, got %v", expected, actual)
  }
}

func mustParse(s string) *url.URL {
  url, err := url.Parse(s)
  if err != nil {
    panic(err)
  }
  return url
}
