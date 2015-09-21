package main

import (
  "code.google.com/p/gosqlite/sqlite"
  "flag"
  "fmt"
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/gofunctional3/consume"
  "github.com/keep94/vsafe"
  "github.com/keep94/vsafe/vsafedb/for_sqlite"
)

var (
  fDb string
)

func main() {
  flag.Parse()
  if fDb == "" {
    fmt.Println("Need to specify -db flag.")
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
  var users []*vsafe.User
  if err = store.Users(nil, consume.AppendPtrsTo(&users, nil)); err != nil {
    fmt.Printf("Error fetching users - %v\n", err)
    return
  }
  usersById := make(map[int64]*vsafe.User)
  for _, user := range users {
    usersById[user.Id] = user
  }
  for _, user := range users {
    fmt.Printf("%s\t%s\n", user.Name, ownerStr(usersById, user.GetOwner()))
  }
}

func ownerStr(userMap map[int64]*vsafe.User, ownerId int64) string {
  if userMap[ownerId] == nil {
    return fmt.Sprintf("(%d)", ownerId)
  }
  return userMap[ownerId].Name
}

func init() {
  flag.StringVar(&fDb, "db", "", "Path to vsafe file")
}
