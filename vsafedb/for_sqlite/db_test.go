package for_sqlite_test

import (
	"database/sql"
	"testing"

	"github.com/keep94/toolbox/db/sqlite3_db"
	"github.com/keep94/vsafe/vsafedb/fixture"
	"github.com/keep94/vsafe/vsafedb/for_sqlite"
	"github.com/keep94/vsafe/vsafedb/sqlite_setup"
	_ "github.com/mattn/go-sqlite3"
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

func TestCategoriesByOwner(t *testing.T) {
	db := openDb(t)
	defer closeDb(t, db)
	fixture.CategoriesByOwner(t, for_sqlite.New(db))
}

func TestCategoryById(t *testing.T) {
	db := openDb(t)
	defer closeDb(t, db)
	fixture.CategoryById(t, for_sqlite.New(db))
}

func TestUpdateCategory(t *testing.T) {
	db := openDb(t)
	defer closeDb(t, db)
	fixture.UpdateCategory(t, for_sqlite.New(db))
}

func TestRemoveCategory(t *testing.T) {
	db := openDb(t)
	defer closeDb(t, db)
	fixture.RemoveCategory(t, for_sqlite.New(db))
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

func closeDb(t *testing.T, db *sqlite3_db.Db) {
	if err := db.Close(); err != nil {
		t.Errorf("Error closing database: %v", err)
	}
}

func openDb(t *testing.T) *sqlite3_db.Db {
	rawdb, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}
	db := sqlite3_db.New(rawdb)
	err = db.Do(sqlite_setup.SetUpTables)
	if err != nil {
		t.Fatalf("Error creating tables: %v", err)
	}
	return db
}
