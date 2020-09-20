package home

import (
	"github.com/keep94/toolbox/http_util"
	"github.com/keep94/vsafe"
	"github.com/keep94/vsafe/apps/vsafe/common"
	"github.com/keep94/vsafe/vsafedb"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
)

const (
	kRowsAtTop = 1
)

var (
	kTemplateSpec = `
<html>
<head>
  <title>Vsafe using Go</title>
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
  <link rel="shortcut icon" href="/images/favicon.ico" type="image/x-icon" />
</head>
<body>
<h2>Vsafe using Go for {{.Name}}</h2>
<form action="/vsafe/home">
  <input type="text" name="q" value="{{.Get "q"}}" />
  <select name="cat" size=1>
{{with .GetSelection .CatSelections "cat"}}
    <option value="{{.Value}}">{{.Name}}</option>
{{end}}
    <option value="">--All--</option>
{{range .CatSelections}}
    <option value="{{.Value}}">{{.Name}}</option>
{{end}}
  <input type="submit" value="Search" />
</form>
<form method="post" action="{{.EntryLink 0}}">
   <input type="submit" accesskey="n" value="New Entry (Ctrl+Alt+N)">
</form>
&nbsp;
&nbsp;
<a href="/vsafe/chpasswd">Change password</a>
&nbsp;
&nbsp;
<a href="/vsafe/catedit">Edit categories</a>
&nbsp;
&nbsp;
<a href="/vsafe/logout">Sign out</a>
&nbsp;
&nbsp;
<br/>
<br/>
<table>
  <tr>
    <td>
      {{if .Equals "sort" ""}}
        Title
      {{else}}
        <a href="{{.SortBy ""}}">Title</a>
      {{end}}
    </td>
    <td>
      {{if .Equals "sort" "newest"}}
        Newest First
      {{else}}
        <a href="{{.SortBy "newest"}}">Newest First</a>
      {{end}}
    </td>
  </tr>
 {{with $top := .}}
 {{range $idx, $element := .Entries}}
  {{if $top.IsCurrent .Id}}
    <tr bgcolor="#CCCC00">
  {{else}}
    <tr class="lineitem">
  {{end}}
    <td>
      {{if $top.HasAnchor $idx}}<a name="{{$top.Anchor $idx}}" />{{end}}
      <a href="{{if .Url}}{{.Url}}{{end}}" target="_blank">{{.Title}}</a>
    </td>
    <td rowspan="2" bgcolor="#FFFFFF">
      <form method="post" action="{{$top.EntryLink .Id}}">
        <input type="submit" value="View">
      </form>
   </td>
  </tr>
  <tr>
    {{if .Desc}}
      <td>{{.Desc}}</td>
    {{else}}
      <td>&nbsp;</td>
    {{end}}
  </tr>
 {{end}}
 {{end}}
</table>
</body>
</html>`
)

var (
	kTemplate *template.Template
)

type Store interface {
	vsafedb.EntriesByOwnerRunner
	vsafedb.CategoriesByOwnerRunner
}

type Handler struct {
	Store Store
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	session := common.GetUserSession(r)
	sortBy := r.Form.Get("sort")
	id, _ := strconv.ParseInt(r.Form.Get("id"), 10, 64)
	catId, _ := strconv.ParseInt(r.Form.Get("cat"), 10, 64)
	categories, err := h.Store.CategoriesByOwner(nil, session.Key().Id)
	if err != nil {
		http_util.ReportError(w, "Error reading database", err)
		return
	}
	entries, err := vsafedb.Entries(h.Store, session.Key().Id, r.Form.Get("q"), catId)
	if err != nil {
		http_util.ReportError(w, "Error reading database", err)
		return
	}
	switch sortBy {
	case "newest":
		vsafedb.Reverse(entries)
	default:
		vsafedb.SortByTitle(entries)
	}
	http_util.WriteTemplate(
		w,
		kTemplate,
		&view{
			Values:        http_util.Values{r.Form},
			Name:          session.User.Name,
			Entries:       entries,
			Url:           r.URL,
			Id:            id,
			CatSelections: common.CatSelections(categories)})
}

type view struct {
	http_util.Values
	Name          string
	Entries       []*vsafe.Entry
	Url           *url.URL
	Id            int64
	CatSelections http_util.Selections
}

func (v *view) HasAnchor(idx int) bool {
	return idx+kRowsAtTop < len(v.Entries)
}

func (v *view) Anchor(idx int) int64 {
	return v.Entries[idx+kRowsAtTop].Id
}

func (v *view) EntryLink(id int64) *url.URL {
	return http_util.NewUrl(
		"/vsafe/single",
		"id", strconv.FormatInt(id, 10),
		"prev", v.Url.String())
}

func (v *view) SortBy(sortBy string) *url.URL {
	return http_util.WithParams(
		v.Url,
		"sort", sortBy,
		"id", "0")
}

func (v *view) IsCurrent(id int64) bool {
	return id == v.Id
}

func init() {
	kTemplate = common.NewTemplate("home", kTemplateSpec)
}
