// Package for_sqlite provides a sqlite implementation of interfaces in
// vsafedb package.
package for_sqlite

import (
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/appcommon/idset"
  "github.com/keep94/gofunctional3/consume"
  "github.com/keep94/gofunctional3/functional"
  "github.com/keep94/gosqlite/sqlite"
  "github.com/keep94/vsafe"
  "github.com/keep94/vsafe/vsafedb"
  "net/url"
)

const (
  kSQLUserById = "select id, owner, name, key, checksum from user where id = ?"
  kSQLUserByName = "select id, owner, name, key, checksum from user where name = ?"
  kSQLUsers = "select id, owner, name, key, checksum from user order by name"
  kSQLAddUser = "insert into user (owner, name, key, checksum) values (?, ?, ?, ?)"
  kSQLUpdateUser = "update user set owner = ?, name = ?, key = ?, checksum = ? where id = ?"
  kSQLRemoveUser = "delete from user where name = ?"
  kSQLAddCategory = "insert into category (owner, name) values (?, ?)"
  kSQLCategoryByOwner = "select id, owner, name from category where owner = ? order by name"
  kSQLUpdateCategoryName = "update category set name = ? where id = ? and owner = ?"
  kSQLRemoveCategory = "delete from category where id = ? and owner = ?"
  kSQLEntryById = "select id, owner, url, title, desc, uname, password, special, categories from entry where id = ?"
  kSQLEntryByOwner = "select id, owner, url, title, desc, uname, password, special, categories from entry where owner = ? order by id"
  kSQLAddEntry = "insert into entry (owner, url, title, desc, uname, password, special, categories) values (?, ?, ?, ?, ?, ?, ?, ?)"
  kSQLUpdateEntry = "update entry set owner = ?, url = ?, title = ?, desc = ?, uname = ?, password = ?, special = ?, categories = ? where id = ?"
  kSQLRemoveEntry = "delete from entry where id = ? and owner = ?"
)

type Store struct {
  db sqlite_db.Doer
}

// New creates a sqlite implementation of the vsafe app datastore.
func New(db *sqlite_db.Db) Store {
  return Store{db}
}

// ConnNew creates a sqlite implementation of the vsafe app datastore from
// a sqlite connection instance.
func ConnNew(conn *sqlite.Conn) Store {
  return Store{sqlite_db.NewSqliteDoer(conn)}
}

func (s Store) AddUser(
    t db.Transaction, user *vsafe.User) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.AddRow(conn, &rawUser{}, user, &user.Id, kSQLAddUser)
  })
}

func (s Store) UserById(
    t db.Transaction, id int64, user *vsafe.User) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadSingle(
        conn,
        &rawUser{},
        vsafedb.ErrNoSuchId,
        user,
        kSQLUserById,
        id)
  })
}

func (s Store) UserByName(
    t db.Transaction, name string, user *vsafe.User) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadSingle(
        conn,
        &rawUser{},
        vsafedb.ErrNoSuchId,
        user,
        kSQLUserByName,
        name)
  })
}

func (s Store) Users(
    t db.Transaction, consumer functional.Consumer) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadMultiple(
        conn,
        &rawUser{},
        consumer,
        kSQLUsers)
  })
}

func (s Store) UpdateUser(
    t db.Transaction, user *vsafe.User) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.UpdateRow(conn, &rawUser{}, user, kSQLUpdateUser)
  })
}

func (s Store) RemoveUser(
    t db.Transaction, name string) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return conn.Exec(kSQLRemoveUser, name)
  })
}

func (s Store) AddCategory(
    t db.Transaction, category *vsafe.Category) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.AddRow(
        conn, &rawCategory{}, category, &category.Id, kSQLAddCategory)
  })
}

func (s Store) CategoriesByOwner(
    t db.Transaction, owner int64) ([]vsafe.Category, error) {
  var result []vsafe.Category
  consumer := consume.AppendTo(&result)
  err := sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadMultiple(
        conn,
        &rawCategory{},
        consumer,
        kSQLCategoryByOwner,
        owner)
  })
  if err != nil {
    return nil, err
  }
  return result, nil
}

func (s Store) UpdateCategoryName(
    t db.Transaction, id, owner int64, newName string) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return conn.Exec(kSQLUpdateCategoryName, newName, id, owner)
  })
}

func (s Store) RemoveCategory(
    t db.Transaction, id, owner int64) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return conn.Exec(kSQLRemoveCategory, id, owner)
  })
}

func (s Store) AddEntry(
    t db.Transaction, entry *vsafe.Entry) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.AddRow(
        conn, &rawEntry{}, entry, &entry.Id, kSQLAddEntry)
  })
}

func (s Store) EntryById(
    t db.Transaction, id int64, entry *vsafe.Entry) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadSingle(
        conn,
        &rawEntry{},
        vsafedb.ErrNoSuchId,
        entry,
        kSQLEntryById,
        id)
  })
}

func (s Store) EntryByIdWithEtag(
    t db.Transaction, id int64, entry *vsafe.EntryWithEtag) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadSingle(
        conn,
        &rawEntry{},
        vsafedb.ErrNoSuchId,
        entry,
        kSQLEntryById,
        id)
  })
}

func (s Store) EntriesByOwner(
    t db.Transaction, owner int64, consumer functional.Consumer) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.ReadMultiple(
        conn,
        &rawEntry{},
        consumer,
        kSQLEntryByOwner,
        owner)
  })
}

func (s Store) UpdateEntry(t db.Transaction, entry *vsafe.Entry) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return sqlite_db.UpdateRow(conn, &rawEntry{}, entry, kSQLUpdateEntry)
  })
}

func (s Store) RemoveEntry(t db.Transaction, id, owner int64) error {
  return sqlite_db.ToDoer(s.db, t).Do(func(conn *sqlite.Conn) error {
    return conn.Exec(kSQLRemoveEntry, id, owner)
  })
}

type rawUser struct {
  *vsafe.User
  sqlite_db.SimpleRow
}

func (r *rawUser) Ptrs() []interface{} {
  return []interface{} {&r.Id, &r.Owner, &r.Name, &r.Key, &r.Checksum}
}

func (r *rawUser) Values() []interface{} {
  return []interface{} {r.Owner, r.Name, r.Key, r.Checksum, r.Id}
}

func (r *rawUser) Pair(ptr interface{}) {
  r.User = ptr.(*vsafe.User)
}

type rawCategory struct {
  *vsafe.Category
  sqlite_db.SimpleRow
}

func (r *rawCategory) Ptrs() []interface{} {
  return []interface{} {&r.Id, &r.Owner, &r.Name}
}

func (r *rawCategory) Values() []interface{} {
  return []interface{} {r.Owner, r.Name, r.Id}
}

func (r *rawCategory) Pair(ptr interface{}) {
  r.Category = ptr.(*vsafe.Category)
}

type rawEntry struct {
  *vsafe.Entry
  rawUrl string
  rawCategories string
}

func (r *rawEntry) Ptrs() []interface{} {
  return []interface{} {&r.Id, &r.Owner, &r.rawUrl, &r.Title, &r.Desc, &r.UName, &r.Password, &r.Special, &r.rawCategories}
}

func (r *rawEntry) Values() []interface{} {
  return []interface{} {r.Owner, r.rawUrl, r.Title, r.Desc, r.UName, r.Password, r.Special, r.rawCategories, r.Id}
}

func (r *rawEntry) Pair(ptr interface{}) {
  r.Entry = ptr.(*vsafe.Entry)
}

func (r *rawEntry) Marshall() error {
  r.rawCategories = string(r.Categories)
  if r.Url == nil {
    r.rawUrl = ""
  } else {
    r.rawUrl = r.Url.String()
  }
  return nil
}

func (r *rawEntry) Unmarshall() error {
  r.Categories = idset.IdSet(r.rawCategories)
  var err error
  if r.rawUrl == "" {
    r.Url = nil
  } else {
    r.Url, err = url.Parse(r.rawUrl)
  }
  return err
}
