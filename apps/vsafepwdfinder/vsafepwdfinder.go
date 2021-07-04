package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"

	"github.com/keep94/consume"
	"github.com/keep94/gosqlite/sqlite"
	"github.com/keep94/toolbox/db/sqlite_db"
	"github.com/keep94/vsafe"
	"github.com/keep94/vsafe/vsafedb"
	"github.com/keep94/vsafe/vsafedb/for_sqlite"
	"golang.org/x/term"
)

var (
	fDb   string
	fUser string
)

type store interface {
	vsafedb.EntriesByOwnerRunner
	vsafedb.UserByNameRunner
}

func main() {
	flag.Parse()
	if fDb == "" || fUser == "" {
		fmt.Println("Need to specify at least -db and -user flag.")
		flag.Usage()
		return
	}
	dbase := openDb(fDb)
	defer dbase.Close()
	store := for_sqlite.New(dbase)
	userPassword := getPassword("Enter user password: ")
	searchPassword := getPassword("Enter password to search for: ")
	fmt.Println()
	showResults(store, userPassword, searchPassword)
}

func getPassword(prompt string) string {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		log.Fatal(err)
	}
	return string(password)
}

func showResults(store store, userPassword, searchPassword string) {
	key := getKey(store, userPassword)
	consumer := consume.MapFilter(
		consume.ConsumerFunc(showOneResult),
		func(src, dest *vsafe.Entry) bool {
			*dest = *src
			return dest.Decrypt(key) == nil
		},
		func(entry *vsafe.Entry) bool {
			return entry.Password == searchPassword
		})
	if err := store.EntriesByOwner(nil, key.Id, consumer); err != nil {
		log.Fatal(err)
	}
}

func showOneResult(ptr interface{}) {
	entry := ptr.(*vsafe.Entry)
	fmt.Printf("Id: %d\n", entry.Id)
	fmt.Printf("Url: %v\n", entry.Url)
	fmt.Printf("Title: %v\n", entry.Title)
	fmt.Printf("Desc: %v\n", entry.Desc)
	fmt.Println()
}

func getKey(store vsafedb.UserByNameRunner, password string) *vsafe.Key {
	var user vsafe.User
	if err := store.UserByName(nil, fUser, &user); err != nil {
		log.Fatal(err)
	}
	key, err := user.VerifyPassword(password)
	if err != nil {
		log.Fatal(err)
	}
	return key
}

func openDb(filepath string) *sqlite_db.Db {
	conn, err := sqlite.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	return sqlite_db.New(conn)
}

func init() {
	flag.StringVar(&fDb, "db", "", "Path to database file")
	flag.StringVar(&fUser, "user", "", "Name of user")
}
