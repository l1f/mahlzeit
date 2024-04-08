package app

import (
	"net/http"

	"github.com/l1f/mahlzeit/internal/zaphelper"
	"github.com/carlmjohnson/resperr"
	"go.uber.org/zap"
)

// HandleError handles all errors. They are logged. If a user message
// is provided with [resperr.UserMessenger], that message is written to w.
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	code := resperr.StatusCode(err)

	w.WriteHeader(code)
	if code == http.StatusInternalServerError {
		zaphelper.FromRequest(r).Error("unexpected error", zap.Error(err))
		return
	}

	zaphelper.FromRequest(r).Info("error during request", zap.Error(err))
	_, _ = w.Write([]byte(resperr.UserMessage(err)))
}
