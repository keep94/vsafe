package catedit

import (
  "errors"
  "fmt"
  "github.com/keep94/appcommon/db"
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/vsafe"
  "github.com/keep94/vsafe/apps/vsafe/common"
  "github.com/keep94/vsafe/vsafedb"
  "html/template"
  "net/http"
  "strconv"
  "strings"
)

const (
  kCatEdit = "catedit"
)

var (
  kErrIdFieldRequired = errors.New("Category field required")
  kErrNameFieldRequired = errors.New("Name field required")
)

var (
  kTemplateSpec = `
<html>
  <head>
    <link rel="stylesheet" type="text/css" href="/static/theme.css" />
  </head>
<body>
<h2>Edit categories</h2>
<a href="/vsafe/home">Back</a>
<br><br>
{{if .Error}}
  <span class="error">{{.Error.Error}}</span>
{{end}}
{{if .Message}}
  <font color="#006600"><b>{{.Message}}</b></font>
{{end}}
<form method="post">
<input type="hidden" name="xsrf" value="{{.Xsrf}}">
<table>
  <tr>
    <td>Existing category:</td>
    <td>
      <select name="cat" size=1>
{{with .GetSelection .CatSelections "cat"}}
        <option value="{{.Value}}">{{.Name}}</option>
{{else}}
        <option value="">--Select One--</option>
{{end}}
{{range .CatSelections}}
        <option value="{{.Value}}">{{.Name}}</option>
{{end}}
      </select>
    </td>
  </tr>
  <tr>
    <td>New category:</td>
    <td><input type="text" name="name" value="{{.Get "name"}}" size="40"></td>
  </tr>
</table>
<br>
<input type="submit" name="add" value="Add">
<input type="submit" name="rename" value="Rename">
<input type="submit" name="remove" value="Remove" onclick="return confirm('Are you sure you want to remove this category?');">
</form>
</body>
</html>`
)

var (
  kTemplate *template.Template
)

type CategoryStore interface {
  vsafedb.AddCategoryRunner
  vsafedb.CategoryByIdRunner
  vsafedb.CategoriesByOwnerRunner
  vsafedb.UpdateCategoryRunner
  vsafedb.RemoveCategoryRunner
}

type Handler struct {
  Doer db.Doer
  Store CategoryStore
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  session := common.GetUserSession(r)
  owner := session.User.GetOwner()
  message := ""
  var err error
  var values http_util.Values
  if r.Method == "POST" {
    if !common.VerifyXsrfToken(r, kCatEdit) {
      err = common.ErrXsrf
    } else if http_util.HasParam(r.Form, "add") {
      name := r.Form.Get("name")
      if strings.TrimSpace(name) == "" {
        err = kErrNameFieldRequired
      } else {
        err = h.addCategory(owner, name)
        message = fmt.Sprintf("Category %s added.", name)
      }
    } else if http_util.HasParam(r.Form, "rename") {
      id, _ := strconv.ParseInt(r.Form.Get("cat"), 10, 64)
      name := r.Form.Get("name")
      if id == 0 {
        err = kErrIdFieldRequired
      } else if strings.TrimSpace(name) == "" {
        err = kErrNameFieldRequired
      } else {
        var oldName string
        oldName, err = h.renameCategory(id, owner, name)
        message = fmt.Sprintf(
            "Category %s renamed to %s.", oldName, name)
      }
    } else if http_util.HasParam(r.Form, "remove") {
      id, _ := strconv.ParseInt(r.Form.Get("cat"), 10, 64)
      if id == 0 {
        err = kErrIdFieldRequired
      } else {
        var oldName string
        oldName, err = h.removeCategory(id, owner)
        message = fmt.Sprintf(
            "Category %s removed.", oldName)
      }
    }
    if err != nil {
      values = http_util.Values{r.Form}
      message = ""
    }
  }
  categories, readErr := h.Store.CategoriesByOwner(nil, owner)
  if readErr != nil {
    http_util.ReportError(w, "Error reading database.", readErr)
    return
  }
  http_util.WriteTemplate(
      w,
      kTemplate,
      &view{
          CatSelections: common.CatSelections(categories),
          Values: values,
          Error: err,
          Message: message,
          Xsrf: common.NewXsrfToken(r, kCatEdit)})
}

func (h *Handler) addCategory(owner int64, name string) error {
  category := vsafe.Category{Name: name, Owner: owner}
  return h.Store.AddCategory(nil, &category)
}

func (h *Handler) renameCategory(id, owner int64, newName string) (
    oldName string, err error) {
  err = h.Doer.Do(func(t db.Transaction) error {
    var err error
    oldName, err = vsafedb.UpdateCategory(h.Store, t, id, owner, newName)
    return err
  })
  return
}

func (h *Handler) removeCategory(id, owner int64) (
    oldName string, err error) {
  err = h.Doer.Do(func(t db.Transaction) error {
    var err error
    oldName, err = vsafedb.RemoveCategory(h.Store, t, id, owner)
    return err
  })
  return
}

type view struct {
  http_util.Values
  Error error
  Message string
  Xsrf string
  CatSelections http_util.Selections
}

func init() {
  kTemplate = common.NewTemplate("catedit", kTemplateSpec)
}
