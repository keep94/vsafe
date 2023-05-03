// Package for_sqlite provides a sqlite implementation of interfaces in
// vsafedb package.
package for_sqlite

import (
	"database/sql"
	"net/url"

	"github.com/keep94/consume2"
	"github.com/keep94/toolbox/db"
	"github.com/keep94/toolbox/db/sqlite3_db"
	"github.com/keep94/toolbox/db/sqlite3_rw"
	"github.com/keep94/toolbox/idset"
	"github.com/keep94/vsafe"
	"github.com/keep94/vsafe/vsafedb"
)

const (
	kSQLUserById        = "select id, owner, name, key, checksum from user where id = ?"
	kSQLUserByName      = "select id, owner, name, key, checksum from user where name = ?"
	kSQLUsers           = "select id, owner, name, key, checksum from user order by name"
	kSQLAddUser         = "insert into user (owner, name, key, checksum) values (?, ?, ?, ?)"
	kSQLUpdateUser      = "update user set owner = ?, name = ?, key = ?, checksum = ? where id = ?"
	kSQLRemoveUser      = "delete from user where name = ?"
	kSQLAddCategory     = "insert into category (owner, name) values (?, ?)"
	kSQLCategoryByOwner = "select id, owner, name from category where owner = ? order by name"
	kSQLCategoryById    = "select id, owner, name from category where id = ?"
	kSQLUpdateCategory  = "update category set owner = ?, name = ? where id = ?"
	kSQLRemoveCategory  = "delete from category where id = ?"
	kSQLEntryById       = "select id, owner, url, title, desc, uname, password, special, categories from entry where id = ?"
	kSQLEntryByOwner    = "select id, owner, url, title, desc, uname, password, special, categories from entry where owner = ? order by id"
	kSQLAddEntry        = "insert into entry (owner, url, title, desc, uname, password, special, categories) values (?, ?, ?, ?, ?, ?, ?, ?)"
	kSQLUpdateEntry     = "update entry set owner = ?, url = ?, title = ?, desc = ?, uname = ?, password = ?, special = ?, categories = ? where id = ?"
	kSQLRemoveEntry     = "delete from entry where id = ? and owner = ?"
)

type Store struct {
	db sqlite3_db.Doer
}

// New creates a sqlite implementation of the vsafe app datastore.
func New(db *sqlite3_db.Db) Store {
	return Store{db}
}

// ConnNew creates a sqlite implementation of the vsafe app datastore from
// a sqlite connection instance.
func ConnNew(tx *sql.Tx) Store {
	return Store{sqlite3_db.NewSqlite3Doer(tx)}
}

func (s Store) AddUser(
	t db.Transaction, user *vsafe.User) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.AddRow(
			tx, (&rawUser{}).init(user), &user.Id, kSQLAddUser)
	})
}

func (s Store) UserById(
	t db.Transaction, id int64, user *vsafe.User) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadSingle(
			tx,
			(&rawUser{}).init(user),
			vsafedb.ErrNoSuchId,
			kSQLUserById,
			id)
	})
}

func (s Store) UserByName(
	t db.Transaction, name string, user *vsafe.User) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadSingle(
			tx,
			(&rawUser{}).init(user),
			vsafedb.ErrNoSuchId,
			kSQLUserByName,
			name)
	})
}

func (s Store) Users(
	t db.Transaction, consumer consume2.Consumer[vsafe.User]) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadMultiple[vsafe.User](
			tx,
			(&rawUser{}).init(&vsafe.User{}),
			consumer,
			kSQLUsers)
	})
}

func (s Store) UpdateUser(
	t db.Transaction, user *vsafe.User) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.UpdateRow(
			tx, (&rawUser{}).init(user), kSQLUpdateUser)
	})
}

func (s Store) RemoveUser(
	t db.Transaction, name string) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		_, err := tx.Exec(kSQLRemoveUser, name)
		return err
	})
}

func (s Store) AddCategory(
	t db.Transaction, category *vsafe.Category) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.AddRow(
			tx, (&rawCategory{}).init(category), &category.Id, kSQLAddCategory)
	})
}

func (s Store) CategoriesByOwner(
	t db.Transaction, owner int64) ([]vsafe.Category, error) {
	var result []vsafe.Category
	consumer := consume2.AppendTo(&result)
	err := sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadMultiple[vsafe.Category](
			tx,
			(&rawCategory{}).init(&vsafe.Category{}),
			consumer,
			kSQLCategoryByOwner,
			owner)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s Store) CategoryById(
	t db.Transaction, id int64, category *vsafe.Category) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadSingle(
			tx,
			(&rawCategory{}).init(category),
			vsafedb.ErrNoSuchId,
			kSQLCategoryById,
			id)
	})
}

func (s Store) UpdateCategory(t db.Transaction, category *vsafe.Category) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.UpdateRow(
			tx, (&rawCategory{}).init(category), kSQLUpdateCategory)
	})
}

func (s Store) RemoveCategory(t db.Transaction, id int64) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		_, err := tx.Exec(kSQLRemoveCategory, id)
		return err
	})
}

func (s Store) AddEntry(
	t db.Transaction, entry *vsafe.Entry) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.AddRow(
			tx, (&rawEntry{}).init(entry), &entry.Id, kSQLAddEntry)
	})
}

func (s Store) EntryById(
	t db.Transaction, id int64, entry *vsafe.Entry) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadSingle(
			tx,
			(&rawEntry{}).init(entry),
			vsafedb.ErrNoSuchId,
			kSQLEntryById,
			id)
	})
}

func (s Store) EntriesByOwner(
	t db.Transaction,
	owner int64,
	consumer consume2.Consumer[vsafe.Entry]) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.ReadMultiple[vsafe.Entry](
			tx,
			(&rawEntry{}).init(&vsafe.Entry{}),
			consumer,
			kSQLEntryByOwner,
			owner)
	})
}

func (s Store) UpdateEntry(t db.Transaction, entry *vsafe.Entry) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		return sqlite3_rw.UpdateRow(
			tx, (&rawEntry{}).init(entry), kSQLUpdateEntry)
	})
}

func (s Store) RemoveEntry(t db.Transaction, id, owner int64) error {
	return sqlite3_db.ToDoer(s.db, t).Do(func(tx *sql.Tx) error {
		_, err := tx.Exec(kSQLRemoveEntry, id, owner)
		return err
	})
}

type rawUser struct {
	*vsafe.User
	sqlite3_rw.SimpleRow
}

func (r *rawUser) init(bo *vsafe.User) *rawUser {
	r.User = bo
	return r
}

func (r *rawUser) Ptrs() []interface{} {
	return []interface{}{&r.Id, &r.Owner, &r.Name, &r.Key, &r.Checksum}
}

func (r *rawUser) Values() []interface{} {
	return []interface{}{r.Owner, r.Name, r.Key, r.Checksum, r.Id}
}

func (r *rawUser) ValueRead() vsafe.User {
	return *r.User
}

type rawCategory struct {
	*vsafe.Category
	sqlite3_rw.SimpleRow
}

func (r *rawCategory) init(bo *vsafe.Category) *rawCategory {
	r.Category = bo
	return r
}

func (r *rawCategory) Ptrs() []interface{} {
	return []interface{}{&r.Id, &r.Owner, &r.Name}
}

func (r *rawCategory) Values() []interface{} {
	return []interface{}{r.Owner, r.Name, r.Id}
}

func (r *rawCategory) ValueRead() vsafe.Category {
	return *r.Category
}

type rawEntry struct {
	*vsafe.Entry
	rawUrl        string
	rawCategories string
}

func (r *rawEntry) init(bo *vsafe.Entry) *rawEntry {
	r.Entry = bo
	return r
}

func (r *rawEntry) Ptrs() []interface{} {
	return []interface{}{&r.Id, &r.Owner, &r.rawUrl, &r.Title, &r.Desc, &r.UName, &r.Password, &r.Special, &r.rawCategories}
}

func (r *rawEntry) Values() []interface{} {
	return []interface{}{r.Owner, r.rawUrl, r.Title, r.Desc, r.UName, r.Password, r.Special, r.rawCategories, r.Id}
}

func (r *rawEntry) ValueRead() vsafe.Entry {
	return *r.Entry
}

func (r *rawEntry) SetEtag(etag uint64) {
	r.Etag = etag
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
	var err error
	r.Categories = idset.IdSet(r.rawCategories)
	if r.rawUrl == "" {
		r.Url = nil
	} else {
		r.Url, err = url.Parse(r.rawUrl)
	}
	return err
}
