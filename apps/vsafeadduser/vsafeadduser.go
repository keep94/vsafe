package main

import (
  "code.google.com/p/gosqlite/sqlite"
  "flag"
  "fmt"
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/vsafe"
  "github.com/keep94/vsafe/vsafedb/for_sqlite"
  "github.com/keep94/vsafe/vsafedb/sqlite_setup"
)

var (
  fDb string
  fName string
  fPassword string
  fMasterName string
  fMasterPassword string
)

func main() {
  flag.Parse()
  if fDb == "" || fName == "" || fPassword == "" {
    fmt.Println("Need to specify -db, -name, and -password flag.")
    flag.Usage()
    return
  }
  conn, err := sqlite.Open(fDb)
  if err != nil {
    fmt.Printf("Unable to open database - %s\n", fDb)
    return
  }
  dbase := sqlite_db.New(conn)
  defer dbase.Close()
  err = dbase.Do(func(conn *sqlite.Conn) error {
    return sqlite_setup.SetUpTables(conn)
  })
  if err != nil {
    fmt.Printf("Unable to create tables - %v\n", err)
    return
  }
  store := for_sqlite.New(dbase)
  var user vsafe.User
  if fMasterName == "" {
    err = user.Init(fName, fPassword)
  } else {
    var master vsafe.User
    if err = store.UserByName(nil, fMasterName, &master); err != nil {
      fmt.Printf("Error retrieving master user - %v\n", err)
      return
    }
    var key *vsafe.Key
    if key, err = master.VerifyPassword(fMasterPassword); err != nil {
      fmt.Printf("Error verifying master password - %v\n", err)
      return
    }
    err = user.InitWithKey(fName, fPassword, key)
  }
  if err != nil {
    fmt.Printf("Error initializing user - %v\n", err)
    return
  }
  if err = store.AddUser(nil, &user); err != nil {
    fmt.Printf("Error storing user in database - %v\n", err)
    return
  }
}

func init() {
  flag.StringVar(&fDb, "db", "", "Path to vsafe file")
  flag.StringVar(&fName, "name", "", "User name")
  flag.StringVar(&fPassword, "password", "password", "User password")
  flag.StringVar(&fMasterName, "master", "", "Master user name")
  flag.StringVar(&fMasterPassword, "mp", "password", "Master password")
}
