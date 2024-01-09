package services

import (
	"context"
	"net/http"

	"github.com/justinas/nosurf"
)

type csrf struct{}

var CSRF csrf

func (csrf) GetContext(parent context.Context, r *http.Request) context.Context {
	return context.WithValue(parent, ctxKeyCSRF, nosurf.Token(r))
}

func (csrf) GetToken(ctx context.Context) string {
	if token, ok := ctx.Value(ctxKeyCSRF).(string); ok {
		return token
	}
	return ""
}
