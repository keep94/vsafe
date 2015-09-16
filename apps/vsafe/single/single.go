package single

import (
  "errors"
  "fmt"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/gofunctional3/functional"
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
  <title>Vsafe using Go</title>
  <link rel="stylesheet" type="text/css" href="/static/theme.css">
  <link rel="shortcut icon" href="/images/favicon.ico" type="image/x-icon" />
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
  vsafedb.EntryByIdWithEtagRunner
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
    var mutation functional.Filterer
    mutation, err = toEntry(r.Form)
    if err == nil {
      if isIdValid(id) {
        tag, _ := strconv.ParseUint(r.Form.Get("etag"), 10, 64)
        err = h.Doer.Do(func(t db.Transaction) error {
          return vsafedb.UpdateEntryWithEtag(
              h.Store, t, id, tag, session.Key(), mutation)
        })
      } else {
        var newId int64
        var entry vsafe.Entry
        mutation.Filter(&entry)
        newId, err = vsafedb.AddEntry(h.Store, nil, session.Key(), &entry)
        if err == nil {
          id = newId
        }
      }
    }
  }
  if err == vsafedb.ErrConcurrentModification {
    err = errors.New("Someone else updated this entry after you started. Click cancel and try again.")
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
    var u *url.URL
    u, err = url.Parse(r.Form.Get("prev"))
    if err != nil {
      http_util.ReportError(w, "Error parsing prev url", err)
      return
    }
    http_util.Redirect(w, r, withId(u, id).String())
  }
}

func (h *Handler) doGet(w http.ResponseWriter, r *http.Request, id int64) {
  session := common.GetUserSession(r)
  if isIdValid(id) {
    var entryWithEtag vsafe.EntryWithEtag
    err := vsafedb.EntryByIdWithEtag(
        h.Store, nil, id, session.Key(), &entryWithEtag)
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
            fromEntry(&entryWithEtag.Entry, entryWithEtag.Etag),
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

func withId(url *url.URL, id int64) *url.URL {
  idStr := strconv.FormatInt(id, 10)
  result := *http_util.WithParams(url, "id", idStr)
  result.Fragment = idStr
  return &result
}

func toEntry(values url.Values) (mutation functional.Filterer, err error) {
  url, err := safeUrlParse(values.Get("url"))
  if err != nil {
    return
  }
  title := values.Get("title")
  desc := values.Get("desc")
  uName := values.Get("uname")
  password := values.Get("password")
  special := values.Get("special")
  mutation = functional.NewFilterer(func(ptr interface{}) error {

    // We have to skip if nothing changed. Otherwise the etag will change
    // when we update even if we don't change anything. This is because
    // of the random seed added to the encryption.
    entryPtr := ptr.(*vsafe.Entry)
    changed := false
    if safeUrlString(entryPtr.Url) != safeUrlString(url) {
      entryPtr.Url = url
      changed = true
    }
    if entryPtr.Title != title {
      entryPtr.Title = title
      changed = true
    }
    if entryPtr.Desc != desc {
      entryPtr.Desc = desc
      changed = true
    }
    if entryPtr.UName != uName {
      entryPtr.UName = uName
      changed = true
    }
    if entryPtr.Password != password {
      entryPtr.Password = password
      changed = true
    }
    if entryPtr.Special != special {
      entryPtr.Special = special
      changed = true
    }
    if changed {
      return nil
    }
    return functional.Skipped
  })
  return
}

func fromEntry(entry *vsafe.Entry, tag uint64) url.Values {
  result := make(url.Values)
  result.Set("url", safeUrlString(entry.Url))
  result.Set("title", entry.Title)
  result.Set("desc", entry.Desc)
  result.Set("uname", entry.UName)
  result.Set("password", entry.Password)
  result.Set("special", entry.Special)
  result.Set("etag", strconv.FormatUint(tag, 10))
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
