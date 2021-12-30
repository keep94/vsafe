package main

import (
	"flag"
	"fmt"
	"github.com/keep94/consume2"
	"github.com/keep94/gosqlite/sqlite"
	"github.com/keep94/toolbox/db/sqlite_db"
	"github.com/keep94/vsafe"
	"github.com/keep94/vsafe/vsafedb"
	"github.com/keep94/vsafe/vsafedb/for_sqlite"
	"github.com/keep94/vsafe/vsafedb/sqlite_setup"
	"os"
)

const (
	kDbFlag   = "db"
	kNameFlag = "name"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: vsafeuser <command> [<args>]")
		fmt.Println("  list   list the users")
		fmt.Println("  add    add a user")
		fmt.Println("  remove remove user")
		return
	}
	switch os.Args[1] {
	case "list":
		if !doList(os.Args[2:]) {
			os.Exit(1)
		}
	case "add":
		if !doAdd(os.Args[2:]) {
			os.Exit(1)
		}
	case "remove":
		if !doRemove(os.Args[2:]) {
			os.Exit(1)
		}
	default:
		fmt.Printf("%q is not a valid command.\n", os.Args[1])
		os.Exit(2)
	}
}

func checkStrFlag(f *flag.FlagSet, flagName, flagValue string) {
	if flagValue == "" {
		fmt.Fprintf(f.Output(), "Need to specify -%s flag.\n", flagName)
		os.Exit(2)
	}
}

func checkDbAndName(f *flag.FlagSet, db, name string) {
	checkStrFlag(f, kDbFlag, db)
	checkStrFlag(f, kNameFlag, name)
}

func addDbFlag(f *flag.FlagSet) *string {
	return f.String(kDbFlag, "", "Path to vsafe file")
}

func addNameFlag(f *flag.FlagSet) *string {
	return f.String(kNameFlag, "", "User name")
}

func doList(args []string) bool {
	flags := flag.NewFlagSet("list", flag.ExitOnError)
	dbPath := addDbFlag(flags)
	flags.Parse(args)
	checkStrFlag(flags, kDbFlag, *dbPath)
	dbase := openDb(*dbPath)
	defer dbase.Close()
	store, ok := initDb(dbase)
	if !ok {
		return false
	}
	return listUsers(store)
}

func doAdd(args []string) bool {
	flags := flag.NewFlagSet("add", flag.ExitOnError)
	dbPath := addDbFlag(flags)
	name := addNameFlag(flags)
	password := flags.String("password", "password", "User password")
	masterName := flags.String("master", "", "Master user name")
	masterPassword := flags.String("mp", "password", "Master password")
	flags.Parse(args)
	checkDbAndName(flags, *dbPath, *name)
	dbase := openDb(*dbPath)
	defer dbase.Close()
	store, ok := initDb(dbase)
	if !ok {
		return false
	}
	var user vsafe.User
	var err error
	if *masterName == "" {
		err = user.Init(*name, *password)
	} else {
		var master vsafe.User
		if err = store.UserByName(nil, *masterName, &master); err != nil {
			fmt.Printf("Error retrieving master user - %v\n", err)
			return false
		}
		var key *vsafe.Key
		if key, err = master.VerifyPassword(*masterPassword); err != nil {
			fmt.Printf("Error verifying master password - %v\n", err)
			return false
		}
		err = user.InitWithKey(*name, *password, key)
	}
	if err != nil {
		fmt.Printf("Error initializing user -%v\n", err)
		return false
	}
	if err = store.AddUser(nil, &user); err != nil {
		fmt.Printf("Error storing user in database - %v\n", err)
		return false
	}
	return true
}

func doRemove(args []string) bool {
	flags := flag.NewFlagSet("remove", flag.ExitOnError)
	dbPath := addDbFlag(flags)
	name := addNameFlag(flags)
	flags.Parse(args)
	checkDbAndName(flags, *dbPath, *name)
	dbase := openDb(*dbPath)
	defer dbase.Close()
	store, ok := initDb(dbase)
	if !ok {
		return false
	}
	if err := store.RemoveUser(nil, *name); err != nil {
		fmt.Printf("Error removing user - %v\n", err)
		return false
	}
	return true
}

func openDb(dbPath string) *sqlite_db.Db {
	conn, err := sqlite.Open(dbPath)
	if err != nil {
		fmt.Printf("Unable to open database - %s\n", dbPath)
		os.Exit(1)
	}
	return sqlite_db.New(conn)
}

func initDb(dbase *sqlite_db.Db) (store for_sqlite.Store, ok bool) {
	err := dbase.Do(func(conn *sqlite.Conn) error {
		return sqlite_setup.SetUpTables(conn)
	})
	if err != nil {

		fmt.Printf("Unable to create tables - %v\n", err)
		return
	}
	return for_sqlite.New(dbase), true
}

func listUsers(store vsafedb.UsersRunner) bool {
	var users []*vsafe.User
	if err := store.Users(nil, consume2.AppendPtrsTo(&users)); err != nil {
		fmt.Printf("Error fetching users - %v\n", err)
		return false
	}
	usersById := make(map[int64]*vsafe.User)
	for _, user := range users {
		usersById[user.Id] = user
	}
	for _, user := range users {
		fmt.Printf("%-20s %-20s\n", user.Name, ownerStr(usersById, user.GetOwner()))
	}
	return true
}

func ownerStr(userMap map[int64]*vsafe.User, ownerId int64) string {
	if userMap[ownerId] == nil {
		return fmt.Sprintf("(%d)", ownerId)
	}
	return userMap[ownerId].Name
}
