package frontend

import (
	"context"
	"net/http"

	"github.com/justinas/nosurf"
)

type key int

const (
	ctxToken key = iota
)

func csrf_token(ctx context.Context) string {
	if token, ok := ctx.Value(ctxToken).(string); ok {
		return token
	}
	return ""
}

func contextWithCSRFToken(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, ctxToken, nosurf.Token(r))
}
