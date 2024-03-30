package httpproxy

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"guardian/internal/guardian/app/proxy/downstream"
)

type ErrUnauthorized struct {
	Reason string
}

func (e ErrUnauthorized) Error() string {
	return fmt.Sprintf("unauthorized: %s", e.Reason)
}

func (p *proxy) handleErr(err error, w http.ResponseWriter, log proxyLog) {
	p.logProxyErr(err, log)

	switch errors.Cause(err) {
	case ErrRequestNotMatched,
		downstream.ErrAuthDataNotFound,
		downstream.ErrAuthDataInvalid:
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	//nolint:gocritic
	switch errors.Cause(err).(type) {
	case *ErrUnauthorized:
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}
