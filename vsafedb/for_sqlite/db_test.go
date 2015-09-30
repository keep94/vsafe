package for_sqlite_test

import (
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/gosqlite/sqlite"
  "github.com/keep94/vsafe/vsafedb/fixture"
  "github.com/keep94/vsafe/vsafedb/for_sqlite"
  "github.com/keep94/vsafe/vsafedb/sqlite_setup"
  "testing"
)

func TestUserById(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  fixture.UserById(t, for_sqlite.New(db))
}

func TestUserByName(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  fixture.UserByName(t, for_sqlite.New(db))
}

func TestUsers(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  fixture.Users(t, for_sqlite.New(db))
}

func TestUpdateUser(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  fixture.UpdateUser(t, for_sqlite.New(db))
}

func TestRemoveUser(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  fixture.RemoveUser(t, for_sqlite.New(db))
}

func TestUserDupName(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  fixture.UserDupName(t, for_sqlite.New(db))
}

func TestEntryById(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  fixture.EntryById(t, for_sqlite.New(db))
}

func TestEntriesByOwner(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  fixture.EntriesByOwner(t, for_sqlite.New(db))
}

func TestUpdateEntry(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  fixture.UpdateEntry(t, for_sqlite.New(db))
}

func TestRemoveEntry(t *testing.T) {
  db := openDb(t)
  defer closeDb(t, db)
  fixture.RemoveEntry(t, for_sqlite.New(db))
}

func closeDb(t *testing.T, db *sqlite_db.Db) {
  if err := db.Close(); err != nil {
    t.Errorf("Error closing database: %v", err)
  }
}

func openDb(t *testing.T) *sqlite_db.Db {
  conn, err := sqlite.Open(":memory:")
  if err != nil {
    t.Fatalf("Error opening database: %v", err)
  }
  db := sqlite_db.New(conn)
  err = db.Do(func(conn *sqlite.Conn) error {
    return sqlite_setup.SetUpTables(conn)
  })
  if err != nil {
    t.Fatalf("Error creating tables: %v", err)
  }
  return db
}

