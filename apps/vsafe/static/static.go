// Package static provides static content for the vsafe app.
package static

import (
	"github.com/keep94/toolbox/http_util"
	"net/http"
)

var (
	kThemeCss = `
.positive {color:#006600;}
.negative {color:#660000;}
.lineitem {background-color:#CCCCCC}
.error {color:#FF0000;font-weight:bold;}
`
)

var (
	kVsafeJs = `
function AutoLogout(ping_url, login_url, freq) {
  this._ping_url = ping_url;
  this._login_url = login_url;
  this._freq = freq;
}

AutoLogout.prototype.start = function() {
  this._schedulePing();
}

AutoLogout.prototype._initRequest = function() {
  if (window.isIE) {
    return new ActiveXObject("Microsoft.XMLHTTP");
  }
  return new XMLHttpRequest();
};

AutoLogout.prototype._handleLogout = function() {
  window.location = this._login_url + '?prev=' + escape(document.URL);
};

AutoLogout.prototype._delayTime = function() {
  var min = Math.floor(this._freq * 0.5);
  return min + Math.floor(Math.random() * min);
};

AutoLogout.prototype._schedulePing = function() {
  var that = this;
  var f = function() {
    that._ping();
  };
  setTimeout(f, this._delayTime());
};

AutoLogout.prototype._ping = function() {
  var req = this._initRequest();
  var that = this;
  req.onreadystatechange = function() {
    if (req.readyState == 4) {
      if (req.status == 401) {
        that._handleLogout();
      } else if (req.status == 200) {
        that._schedulePing();
      }
    }
  };
  req.open("GET", this._ping_url, true);
  req.send(null);
};
`
)

func New() http.Handler {
	result := http.NewServeMux()
	http_util.AddStatic(result, "/theme.css", kThemeCss)
	http_util.AddStatic(result, "/vsafe.js", kVsafeJs)
	return result
}
