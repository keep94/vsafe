package logout

import (
	"github.com/keep94/toolbox/http_util"
	"github.com/keep94/vsafe/apps/vsafe/common"
	"net/http"
)

type Handler struct {
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session := common.GetUserSession(r)
	session.ClearAll()
	session.Save(r, w)
	http_util.Redirect(w, r, "/vsafe/home")
}
