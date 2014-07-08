package home

import (
  "github.com/keep94/appcommon/http_util"
  "github.com/keep94/vsafe"
  "github.com/keep94/vsafe/apps/common"
  "github.com/keep94/vsafe/vsafedb"
  "html/template"
  "net/http"
  "net/url"
  "strconv"
)

var (
kTemplateSpec = `
<html>
<head>
  <title>Vsafe using Go for {{.Name}}</title>
  <link rel="stylesheet" type="text/css" href="/static/theme.css" />
</head>
<body>
<h2>Vsafe using Go for {{.Name}}</h2>
<form>
  <input type="text" name="q" value="{{.Get "q"}}" />
  <input type="submit" value="Search" />
</form>
<a href="{{.EntryLink 0}}">New Entry</a>
&nbsp;
&nbsp;
<a href="/vsafe/chpasswd">Change password</a>
&nbsp;
&nbsp;
<a href="/vsafe/logout">Sign out</a>
&nbsp;
&nbsp;
<br/>
<br/>
<table>
  <tr>
    <td>Title</td>
    <td>Password</td>
  </tr>
 {{with $top := .}}
 {{range .Entries}}
  <tr class="lineitem">
    <td><a href="{{if .Url}}{{.Url}}{{end}}" target="_blank">{{.Title}}</a></td>
    <td><a href="{{$top.EntryLink .Id}}">View</a></td>
  </tr>
  <tr>
    {{if .Desc}}
      <td colspan=2>{{.Desc}}</td>
    {{else}}
      <td colspan=2>&nbsp;</td>
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

type Handler struct {
  Store vsafedb.EntriesByOwnerRunner
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  r.ParseForm()
  session := common.GetUserSession(r)
  entries, err := vsafedb.Entries(h.Store, session.Key(), r.Form.Get("q"))
  if err != nil {
    http_util.ReportError(w, "Error reading database", err)
    return
  }
  http_util.WriteTemplate(
      w,
      kTemplate,
      &view{
          Values: http_util.Values{r.Form},
          Name: session.User.Name,
          Entries: entries,
          Url: r.URL})
}

type view struct {
  http_util.Values
  Name string
  Entries []*vsafe.Entry
  Url *url.URL
}

func (v *view) EntryLink(id int64) *url.URL {
  return http_util.NewUrl(
      "/vsafe/single",
      "id", strconv.FormatInt(id, 10),
      "prev", v.Url.String())
}

func init() {
  kTemplate = common.NewTemplate("home", kTemplateSpec)
}
