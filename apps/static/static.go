// Package static provides static content for the vsafe app.
package static

import (
  "github.com/keep94/appcommon/http_util"
  "net/http"
)

var (
  kThemeCss = `
.positive {color:#006600;}
.negative {color:#660000;}
.lineitem {background-color:#CCCCCC}
.error {color:#FF0000;font-weight:bold;}`
)

func New() http.Handler {
  result := http.NewServeMux()
  http_util.AddStatic(result, "/theme.css", kThemeCss)
  return result
}
