package main

import (
  "flag"
  "fmt"
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/gosqlite/sqlite"
  "github.com/keep94/vsafe/vsafedb/for_sqlite"
)

var (
  fDb string
  fName string
)

func main() {
  flag.Parse()
  if fDb == "" || fName == "" {
    fmt.Println("Need to specify -db and -name flag.")
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
  store := for_sqlite.New(dbase)
  if err = store.RemoveUser(nil, fName); err != nil {
    fmt.Printf("Error removing user - %v\n", err)
    return
  }
}

func init() {
  flag.StringVar(&fDb, "db", "", "Path to vsafe file")
  flag.StringVar(&fName, "name", "", "User name")
}
