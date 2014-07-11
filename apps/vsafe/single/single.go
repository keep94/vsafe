package single

import (
  "fmt"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/etag"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/vsafe"
  "github.com/keep94/vsafe/apps/vsafe/common"
  "github.com/keep94/vsafe/vsafedb"
  "html/template"
  "net/http"
  "net/url"
  "strconv"
  "strings"
)

var (
kTemplateSpec = `
<html>
<head>
  <link rel="stylesheet" type="text/css" href="/static/theme.css">
  <script type="text/javascript" src="/static/vsafe.js"></script>
</head>
<body>
{{if .Error}}
  <span class="error">{{.Error}}</span>
{{end}}
<form method="post">
  <input type="hidden" name="etag" value="{{.Get "etag"}}">
  <table>
    <tr>
      <td align="right">URL: </td>
      <td><input type="text" id="url" name="url" value="{{.Get "url"}}" size="50" />&nbsp;&nbsp;<a href="#" onclick="window.open(document.getElementById('url').value, '_blank'); return false">Open page</a></td>
    </tr>
    <tr>
      <td align="right">Title: </td>
      <td><input type="text" name="title" value="{{.Get "title"}}" size="50" /></td>
    </tr>
    <tr>
      <td align="right">Description: </td>
      <td><input type="text" name="desc" value="{{.Get "desc"}}" size="75" /></td>
    </tr>
  </table>
  <hr/>
  <b>Everything below is encrypted</b>
  <table>
    <tr>
      <td align="right">User Name: </td>
      <td><input type="text" name="uname" value="{{.Get "uname"}}" size="20" /></td>
    </tr>
    <tr>
      <td align="right">Password: </td>
      <td><input type="text" name="password" value="{{.Get "password"}}" size="20" /></td>
    </tr>
   <tr>
      <td align="right">Special: </td>
      <td><textarea name="special" rows="6" cols="75">{{.Get "special"}}</textarea></td>
    </tr>
  </table>
  <table>
    <tr>
      <td><input type="submit" name="save" value="Save" /></td>
      <td><input type="submit" name="cancel" value="Cancel" /></td>
{{if .ExistingEntry}}
      <td><input type="submit" name="delete" value="Delete" onclick="return confirm('Are you sure you want to delete this entry?')"/></td>
{{end}}
   </tr>
 </table>
</form>
<script type="text/javascript">
  var autoLogout = new AutoLogout("/auth/poll?kid={{.KeyId}}", "/auth/login", 60000);
  autoLogout.start();
</script>
</body>
</html>`
)

var (
  kTemplate *template.Template
)

type Store interface {
  vsafedb.AddEntryRunner
  vsafedb.UpdateEntryRunner
  vsafedb.RemoveEntryRunner
  vsafedb.EntryByIdRunner
}

type Handler struct {
  Doer db.Doer
  Store Store
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  id, _ := strconv.ParseInt(r.Form.Get("id"), 10, 64)
  if r.Method == "GET" {
    h.doGet(w, r, id)
  } else {
    h.doPost(w, r, id)
  }
}

func (h *Handler) doPost(w http.ResponseWriter, r *http.Request, id int64) {
  var err error
  session := common.GetUserSession(r)
  if http_util.HasParam(r.Form, "delete") {
    if isIdValid(id) {
      err = h.Store.RemoveEntry(nil, id, session.User.GetOwner())
    }
  } else if http_util.HasParam(r.Form, "cancel") {
    // Do nothing
  } else {
    var entry vsafe.Entry
    err = toEntry(r.Form, &entry)
    if err == nil {
      if isIdValid(id) {
        tag, _ := strconv.ParseUint(r.Form.Get("etag"), 10, 32)
        entry.Id = id
        err = h.Doer.Do(func(t db.Transaction) error {
          return vsafedb.UpdateEntryWithEtag(
              h.Store, t, uint32(tag), session.Key(), &entry)
        })
      } else {
        _, err = vsafedb.AddEntry(h.Store, session.Key(), &entry)
      }
    }
  }
  if err != nil {
    http_util.WriteTemplate(
        w,
        kTemplate,
        newView(
            r.Form,
            isIdValid(id),
            session.Key().Id,
            err))
  } else {
    http_util.Redirect(w, r, r.Form.Get("prev"))
  }
}

func (h *Handler) doGet(w http.ResponseWriter, r *http.Request, id int64) {
  session := common.GetUserSession(r)
  if isIdValid(id) {
    var entry vsafe.Entry
    err := vsafedb.EntryById(h.Store, nil, id, session.Key(), &entry)
    if err == vsafedb.ErrNoSuchId {
      fmt.Fprintln(w, "No entry found.")
      return
    }
    if err != nil {
      http_util.ReportError(w, "Error reading database.", err)
      return
    }
    http_util.WriteTemplate(
        w,
        kTemplate,
        newView(
            fromEntry(&entry),
            true,
            session.Key().Id,
            nil))
  } else {
    initValues := make(url.Values)
    initValues.Set("url", "http://")
    http_util.WriteTemplate(
        w,
        kTemplate,
        newView(
            initValues,
            false,
            session.Key().Id,
            nil))
  }
}

func toEntry(values url.Values, entry *vsafe.Entry) error {
  var err error
  if entry.Url, err = safeUrlParse(values.Get("url")); err != nil {
    return err
  }
  entry.Title = values.Get("title")
  entry.Desc = values.Get("desc")
  entry.UName = values.Get("uname")
  entry.Password = values.Get("password")
  entry.Special = values.Get("special")
  return nil
}

func fromEntry(entry *vsafe.Entry) url.Values {
  result := make(url.Values)
  result.Set("url", safeUrlString(entry.Url))
  result.Set("title", entry.Title)
  result.Set("desc", entry.Desc)
  result.Set("uname", entry.UName)
  result.Set("password", entry.Password)
  result.Set("special", entry.Special)
  tag, err := etag.Etag32(entry)
  if err != nil {
    panic(err)
  }
  result.Set("etag", strconv.FormatInt(int64(tag), 10))
  return result
}

func safeUrlParse(str string) (* url.URL, error) {
  str = strings.TrimSpace(str)
  if str == "" {
    return nil, nil
  }
  return url.Parse(str)
}

func safeUrlString(u *url.URL) string {
  if u == nil {
    return ""
  }
  return u.String()
}

type view struct {
  http_util.Values
  Error error
  ExistingEntry bool
  KeyId int64
}

func newView(
    values url.Values,
    existingEntry bool,
    keyId int64,
    err error) *view {
  return &view{
      Values: http_util.Values{values},
      ExistingEntry: existingEntry,
      KeyId: keyId,
      Error: err}
}

func isIdValid(id int64) bool {
  return id > 0
}

func init() {
  kTemplate = common.NewTemplate("single", kTemplateSpec)
}
