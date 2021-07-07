// Package common provides routines common to all handlers in the vsafe
// webapp.
package common

import (
	"errors"
	"github.com/gorilla/sessions"
	"github.com/keep94/toolbox/http_util"
	"github.com/keep94/toolbox/session_util"
	"github.com/keep94/vsafe"
	"github.com/keep94/vsafe/vsafedb"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

const (
	kCookieName = "session-cookie"
)

const (
	// Set to the same thing as kSessionTimeout in vsafe.go
	kXsrfTimeout = 15 * time.Minute
)

var (
	ErrXsrf = errors.New("Page had grown stale. Please resubmit.")
)

// NewGorillaSession creates a gorilla session for the vsafe app.
func NewGorillaSession(
	sessionStore sessions.Store,
	r *http.Request) (*sessions.Session, error) {
	return sessionStore.Get(r, kCookieName)
}

// UserSession represents a session in the vsafe web app.
type UserSession struct {
	session_util.UserIdSession
	*sessions.Session
	// The logged in user or nil if no user is logged in.
	User *vsafe.User
}

// CreateUserSession creates a UserSession instance from a gorilla session
// but does not populate the user field of the returned session.
func CreateUserSession(s *sessions.Session) *UserSession {
	return &UserSession{
		UserIdSession: session_util.UserIdSession{S: s},
		Session:       s,
	}
}

// NewUserSession creates a UserSession and associates it with the request
// instance. If user not logged in, the User field in returned UserSession
// is nil. Caller must call  must call context.Clear(r) from the
// github.com/gorilla/context package.
// It is up to the caller to ensure that the Id of the key in the returned
// session matches the result of GetOwner() on the user instance in the same
// session.
func NewUserSession(
	userStore vsafedb.UserByIdRunner,
	sessionStore sessions.Store,
	r *http.Request) (*UserSession, error) {
	us, err := session_util.NewUserSession(
		sessionStore,
		r,
		kCookieName,
		func(s *sessions.Session) session_util.UserSession {
			return CreateUserSession(s)
		},
		userGetter{userStore},
		vsafedb.ErrNoSuchId)
	if err != nil {
		return nil, err
	}
	return us.(*UserSession), nil
}

// GetUserSession returns the session associated with the request, r. It
// can only be called after successful completion of NewUserSession.
func GetUserSession(r *http.Request) *UserSession {
	return session_util.GetUserSession(r).(*UserSession)
}

func (s *UserSession) SetUser(userPtr interface{}) {
	s.User = userPtr.(*vsafe.User)
}

// Key returns the encryption key of the current logged in user.
func (s *UserSession) Key() *vsafe.Key {
	result := s.Values[kKeyKey]
	if result == nil {
		return nil
	}
	return result.(*vsafe.Key)
}

// SetKey sets the encryption key of the current logged in user.
func (s *UserSession) SetKey(key *vsafe.Key) {
	if key == nil {
		delete(s.Values, kKeyKey)
	} else {
		s.Values[kKeyKey] = key
	}
}

// NewTemplate returns a new template instance. name is the name
// of the template; templateStr is the template string.
func NewTemplate(name, templateStr string) *template.Template {
	return template.Must(template.New(name).Parse(templateStr))
}

// NewXsrfToken creates a new xsrf token for given action.
func NewXsrfToken(r *http.Request, action string) string {
	userSession := GetUserSession(r)
	return userSession.NewXsrfToken(action, time.Now().Add(kXsrfTimeout))
}

// VerifyXsrfToken verifies the xsrf token for given action.
// VerifyXsrfToken looks for the token under "xsrf" in request.
func VerifyXsrfToken(r *http.Request, action string) bool {
	userSession := GetUserSession(r)
	return userSession.VerifyXsrfToken(r.Form.Get("xsrf"), action, time.Now())
}

type userGetter struct {
	vsafedb.UserByIdRunner
}

func (g userGetter) GetUser(id int64) (interface{}, error) {
	var user vsafe.User
	if err := g.UserById(nil, id, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

type sessionKeyType int

const (
	kKeyKey sessionKeyType = iota
)

// CatSelections converts a list of categories to selections for a combo box
func CatSelections(cats []vsafe.Category) http_util.Selections {
	result := make(http_util.Selections, len(cats))
	for i := range result {
		result[i] = http_util.Selection{
			Value: strconv.FormatInt(cats[i].Id, 10),
			Name:  cats[i].Name}
	}
	return result
}
