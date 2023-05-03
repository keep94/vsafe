package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/keep94/toolbox/db"
	"github.com/keep94/toolbox/db/sqlite3_db"
	"github.com/keep94/vsafe"
	"github.com/keep94/vsafe/vsafedb"
	"github.com/keep94/vsafe/vsafedb/for_sqlite"
	_ "github.com/mattn/go-sqlite3"
)

var (
	fDb       string
	fName     string
	fPassword string
)

type jsonEntry struct {
	Url      string
	Title    string
	Desc     string
	UName    string
	Password string
	Special  string
}

func main() {
	flag.Parse()
	if fDb == "" || fName == "" || fPassword == "" {
		fmt.Println("Need to specify -db, -name, and -password flag.")
		flag.Usage()
		return
	}
	rawdb, err := sql.Open("sqlite3", fDb)
	if err != nil {
		fmt.Printf("Unable to open database - %s\n", fDb)
		return
	}
	dbase := sqlite3_db.New(rawdb)
	defer dbase.Close()
	store := for_sqlite.New(dbase)
	doer := sqlite3_db.NewDoer(dbase)
	var user vsafe.User
	if err = store.UserByName(nil, fName, &user); err != nil {
		fmt.Printf("Error retrieving user - %v\n", err)
		return
	}
	var key *vsafe.Key
	if key, err = user.VerifyPassword(fPassword); err != nil {
		fmt.Printf("Error verifying user password - %v\n", err)
		return
	}
	decoder := json.NewDecoder(os.Stdin)
	var entryList []*jsonEntry
	if err = decoder.Decode(&entryList); err != nil {
		fmt.Printf("Error decoding json - %v\n", err)
		return
	}
	err = doer.Do(func(t db.Transaction) error {
		return doImport(t, store, entryList, key)
	})
	if err != nil {
		fmt.Printf("Import failed - %v\n", err)
	}
}

func doImport(
	t db.Transaction,
	store vsafedb.AddEntryRunner,
	entryList []*jsonEntry,
	key *vsafe.Key) error {
	for i := range entryList {
		if err := doSingleImport(t, store, entryList[i], key); err != nil {
			return err
		}
	}
	return nil
}

func doSingleImport(
	t db.Transaction,
	store vsafedb.AddEntryRunner,
	jentry *jsonEntry,
	key *vsafe.Key) error {
	var entry vsafe.Entry
	var err error
	entry.Url, err = url.Parse(jentry.Url)
	if err != nil {
		return err
	}
	entry.Title = jentry.Title
	entry.Desc = jentry.Desc
	entry.UName = jentry.UName
	entry.Password = jentry.Password
	entry.Special = jentry.Special
	_, err = vsafedb.AddEntry(store, t, key, &entry)
	return err
}

func init() {
	flag.StringVar(&fDb, "db", "", "Path to vsafe file")
	flag.StringVar(&fName, "name", "", "User name")
	flag.StringVar(&fPassword, "password", "password", "User password")
}
