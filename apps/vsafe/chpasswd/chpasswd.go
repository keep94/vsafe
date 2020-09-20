package chpasswd

import (
	"fmt"
	"github.com/keep94/toolbox/db"
	"github.com/keep94/toolbox/http_util"
	"github.com/keep94/vsafe"
	"github.com/keep94/vsafe/apps/vsafe/common"
	"github.com/keep94/vsafe/vsafedb"
	"html/template"
	"net/http"
)

const (
	kMinPasswordLength = 6
)

const (
	kChPasswd = "chpasswd"
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
<h2>Changing password for {{.Name}}</h2>
<a href="/vsafe/home">Back</a>
<br><br>
{{if .Message}}
  {{if .Success}}
    <font color="#006600"><b>{{.Message}}</b></font>
  {{else}}
    <span class="error">{{.Message}}</span>
  {{end}}
  <br>
{{end}}
<form method="post">
<input type="hidden" name="xsrf" value="{{.Xsrf}}">
  <table>
    <tr>
      <td>Old Password: </td>
      <td><input type="password" name="old"></td>
    </tr>
    <tr>
      <td>New Password: </td>
      <td><input type="password" name="new"></td>
    </tr>
    <tr>
      <td>Verify: </td>
      <td><input type="password" name="verify"></td>
    </tr>
  </table>
  <br>
  <input type="submit" value="Change password">
</form>
</body>
</html>`
)

var (
	kTemplate *template.Template
)

type UserStore interface {
	vsafedb.UserByIdRunner
	vsafedb.UpdateUserRunner
}

type Handler struct {
	Store UserStore
	Doer  db.Doer
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session := common.GetUserSession(r)
	if r.Method == "GET" {
		http_util.WriteTemplate(
			w,
			kTemplate,
			&view{
				Name: session.User.Name,
				Xsrf: common.NewXsrfToken(r, kChPasswd)})
	} else {
		r.ParseForm()
		if !common.VerifyXsrfToken(r, kChPasswd) {
			http_util.WriteTemplate(
				w,
				kTemplate,
				&view{
					Name:    session.User.Name,
					Xsrf:    common.NewXsrfToken(r, kChPasswd),
					Message: common.ErrXsrf.Error()})
			return
		}
		old := r.Form.Get("old")
		new := r.Form.Get("new")
		verify := r.Form.Get("verify")
		if new != verify {
			http_util.WriteTemplate(
				w,
				kTemplate,
				&view{
					Name:    session.User.Name,
					Xsrf:    common.NewXsrfToken(r, kChPasswd),
					Message: "Password re-typed incorrectly."})
			return
		}
		if len(new) < kMinPasswordLength {
			http_util.WriteTemplate(
				w,
				kTemplate,
				&view{
					Name: session.User.Name,
					Xsrf: common.NewXsrfToken(r, kChPasswd),
					Message: fmt.Sprintf(
						"Password must be at least %d characters.",
						kMinPasswordLength)})
			return
		}
		err := h.Doer.Do(func(t db.Transaction) error {
			user, err := vsafedb.ChangePassword(
				h.Store, t, session.User.Id, old, new)
			if err != nil {
				return err
			}
			session.User = user
			return nil
		})
		if err == vsafe.ErrWrongPassword {
			http_util.WriteTemplate(
				w,
				kTemplate,
				&view{
					Name:    session.User.Name,
					Xsrf:    common.NewXsrfToken(r, kChPasswd),
					Message: "Old password wrong."})
			return
		}
		if err != nil {
			http_util.ReportError(w, "Error updating database", err)
			return
		}
		http_util.WriteTemplate(
			w,
			kTemplate,
			&view{
				Name:    session.User.Name,
				Message: "Password changed successfully.",
				Xsrf:    common.NewXsrfToken(r, kChPasswd),
				Success: true})
	}
}

type view struct {
	Name    string
	Message string
	Xsrf    string
	Success bool
}

func init() {
	kTemplate = common.NewTemplate("chpasswd", kTemplateSpec)
}
