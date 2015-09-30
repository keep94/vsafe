package main

import (
  "errors"
  "flag"
  "fmt"
  "github.com/gorilla/context"
  "github.com/gorilla/sessions"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/db/sqlite_db"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/appcommon/logging"
  "github.com/keep94/gosqlite/sqlite"
  "github.com/keep94/ramstore"
  "github.com/keep94/vsafe"
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
  "strconv"
)

const (
  kSessionTimeout = 900
)

var (
  fSSLCrt string
  fSSLKey string
  fPort string
  fDb string
  fIcon string
)

var (
  errNotLoggedIn = errors.New("vsafe app: not logged in.")
)

var (
  kDoer db.Doer
  kStore for_sqlite.Store
  kSessionStore = ramstore.NewRAMStore(kSessionTimeout)
  kPollingStore = asPollingStore(kSessionStore)
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
      "/auth/poll", pollHandler{})
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
  if fSSLCrt != "" && fSSLKey != "" {
    if err := http.ListenAndServeTLS(
        fPort, fSSLCrt, fSSLKey, defaultHandler); err != nil {
      fmt.Println(err)
    }
    return
  }
  if err := http.ListenAndServe(fPort, defaultHandler); err != nil {
    fmt.Println(err)
  }
}  

type authHandler struct {
  *http.ServeMux
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  user, _, err := authorizeSession(r, kSessionStore)
  if err == errNotLoggedIn {
    http_util.Redirect(
        w,
        r,
        http_util.NewUrl("/auth/login", "prev", r.URL.String()).String())
    return
  }
  if err != nil {
    http_util.ReportError(w, "Error reading database.", err)
    return
  }
  logging.SetUserName(r, user.Name)
  h.ServeMux.ServeHTTP(w, r)
}

type pollHandler struct {
}

func (h pollHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  keyId, _ := strconv.ParseInt(r.Form.Get("kid"), 10, 64)
  user, key, err := authorizeSession(r, kPollingStore)
  if err == errNotLoggedIn {
    http_util.Error(w, 401)
    return
  }
  if err != nil {
    http_util.ReportError(w, "Error reading database.", err)
    return
  }
  logging.SetUserName(r, user.Name)
  if keyId != key.Id {
    http_util.Error(w, 401)
    return
  }
  http_util.Error(w, 200)
}

func authorizeSession(
    r *http.Request,
    sessionStore sessions.Store) (*vsafe.User, *vsafe.Key, error) {
  session, err := common.NewUserSession(kStore, sessionStore, r)
  if err != nil {
    return nil, nil, err
  }
  key := session.Key()
  if session.User == nil || key == nil || key.Id != session.User.GetOwner() {
    return nil, nil, errNotLoggedIn
  }
  return session.User, key, nil
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
  flag.StringVar(&fSSLCrt, "ssl_crt", "", "SSL Certificate file")
  flag.StringVar(&fSSLKey, "ssl_key", "", "SSL Key file")
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

func asPollingStore(store *ramstore.RAMStore) *ramstore.RAMStore {
  result := *store
  result.SData = result.Data.AsPoller()
  result.Data = nil
  return &result
}

