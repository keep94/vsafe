package main

import (
  "code.google.com/p/gosqlite/sqlite"
  "flag"
  "fmt"
  "github.com/gorilla/context"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/appcommon/logging"
  "github.com/keep94/ramstore"
  "github.com/keep94/vsafe/apps/vsafe/common"
  "github.com/keep94/vsafe/apps/vsafe/chpasswd"
  "github.com/keep94/vsafe/apps/vsafe/home"
  "github.com/keep94/vsafe/apps/vsafe/login"
  "github.com/keep94/vsafe/apps/vsafe/logout"
  "github.com/keep94/vsafe/apps/vsafe/single"
  "github.com/keep94/vsafe/apps/vsafe/static"
  "github.com/keep94/vsafe/vsafedb/for_sqlite"
  "github.com/keep94/weblogs"
  "net/http"
)

const (
  kSessionTimeout = 900
)

var (
  fPort string
  fDb string
  fIcon string
)

var (
  kDoer db.Doer
  kStore for_sqlite.Store
  kSessionStore = ramstore.NewRAMStore(kSessionTimeout)
)

func main() {
  flag.Parse()
  if fDb == "" {
    fmt.Println("Need to specify at least -db flag.")
    flag.Usage()
    return
  }
  setupDb(fDb)
  mux := http.NewServeMux()
  http.HandleFunc("/", rootRedirect)
  http.Handle("/static/", http.StripPrefix("/static", static.New()))
  if fIcon != "" {
    err := http_util.AddStaticFromFile(
        http.DefaultServeMux, "/images/favicon.ico", fIcon)
    if err != nil {
      fmt.Printf("Icon file not found - %s\n", fIcon)
    }
  }
  http.Handle(
      "/auth/login",
      &login.Handler{SessionStore: kSessionStore, Store: kStore})
  http.Handle(
      "/vsafe/", &authHandler{mux})
  mux.Handle("/vsafe/chpasswd", &chpasswd.Handler{Store: kStore, Doer: kDoer})
  mux.Handle("/vsafe/home", &home.Handler{Store: kStore})
  mux.Handle("/vsafe/logout", &logout.Handler{})
  mux.Handle("/vsafe/single", &single.Handler{Store: kStore, Doer: kDoer})
  defaultHandler := context.ClearHandler(
      weblogs.HandlerWithOptions(
          http.DefaultServeMux,
          &weblogs.Options{Logger: logging.ApacheCommonLoggerWithLatency()}))
  if err := http.ListenAndServe(fPort, defaultHandler); err != nil {
    fmt.Println(err)
  }
}  

type authHandler struct {
  *http.ServeMux
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  session, err := common.NewUserSession(kStore, kSessionStore, r)
  if err != nil {
    http_util.ReportError(w, "Error reading database.", err)
    return
  }
  key := session.Key()
  if session.User == nil || key == nil || key.Id != session.User.GetOwner() {
    http_util.Redirect(
        w,
        r,
        http_util.NewUrl("/auth/login", "prev", r.URL.String()).String())
    return
  }
  logging.SetUserName(r, session.User.Name)
  h.ServeMux.ServeHTTP(w, r)
}

// TODO
func rootRedirect(w http.ResponseWriter, r *http.Request) {
  if r.URL.Path == "/" {
    http_util.Redirect(w, r, "/vsafe/home")
  } else {
    http_util.Error(w, http.StatusNotFound)
  }
}

func init() {
  flag.StringVar(&fPort, "http", ":8080", "Port to bind")
  flag.StringVar(&fDb, "db", "", "Path to database file")
  flag.StringVar(&fIcon, "icon", "", "Path to icon file")
}

func setupDb(filepath string) {
  conn, err := sqlite.Open(filepath)
  if err != nil {
    panic(err)
  }
  dbase := sqlite_db.New(conn)
  kDoer = sqlite_db.NewDoer(dbase)
  kStore = for_sqlite.New(dbase)
}

