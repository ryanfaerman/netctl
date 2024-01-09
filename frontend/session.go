package frontend

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/davecgh/go-spew/spew"

	validator "github.com/go-playground/validator/v10"
	"github.com/justinas/nosurf"
	"github.com/ryanfaerman/netctl/web"
	"github.com/ryanfaerman/netctl/web/named"
)

const (
	SessionKeyAuthenticated = "authenticated"
	SessionKeyFlash         = "flash"
)

func (f *Frontend) IsAuthenticated(ctx context.Context) bool {
	val := f.session.GetBool(ctx, SessionKeyAuthenticated)
	spew.Dump(val)
	return f.session.GetBool(ctx, SessionKeyAuthenticated)
}

func (f *Frontend) Flash(ctx context.Context) string {
	return f.session.PopString(ctx, SessionKeyFlash)
}

func (f *Frontend) SetFlash(ctx context.Context, msg string) {
	f.session.Put(ctx, SessionKeyFlash, msg)
}

type SessionCreateInput struct {
	Email string `validate:"required,email"`
}

type SessionCreateErrors struct {
	Email string
}

type SessionCreatePayload struct {
	SessionToken string
	Email        string
}

func init() {
	gob.Register(SessionCreatePayload{})
}

func (f *Frontend) SessionCreateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	inputErrs := SessionCreateErrors{}
	input := SessionCreateInput{
		Email: r.Form.Get("email"),
	}

	if err := validate.Struct(input); err != nil {
		errs := err.(validator.ValidationErrors)
		for field, e := range errs.Translate(trans) {
			switch field {
			case "SessionCreateInput.Email":
				inputErrs.Email = e
			}
		}

		web.LogWith(r.Context(), "hx", "true")
		ctx := context.WithValue(r.Context(), ctxToken, nosurf.Token(r))
		f.html.SessionLoginWithErrors(input, inputErrs).Render(ctx, w)
		return

	}

	f.session.Put(r.Context(), SessionKeyAuthenticated, true)
	f.session.Put(r.Context(), SessionKeyFlash, FlashSessionCreated)

	payload := SessionCreatePayload{
		SessionToken: f.session.Token(r.Context()),
		Email:        input.Email,
	}

	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(payload); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	token, err := f.brc.EncodeToString(b.Bytes())
	if err != nil {
		http.Error(w, "Unable to generate magic link", http.StatusInternalServerError)
		return
	}

	web.LogWith(r.Context(), "magic-token", token)

	f.html.SessionCreated().Render(r.Context(), w)

}

func (f *Frontend) SessionVerifyHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if len(token) == 0 {
		http.Error(w, "Invalid Magic link", http.StatusForbidden)
		http.Redirect(w, r, named.URLFor("index"), http.StatusSeeOther)
		return
	}

	decoded, err := f.brc.DecodeString(token)
	if err != nil {
		http.Error(w, "Invalid Magic link", http.StatusForbidden)
		return
	}

	fmt.Println(decoded.Timestamp())
	if decoded.IsExpired(60) {
		http.Error(w, "Magic link expired", http.StatusForbidden)
		return
	}

	var payload SessionCreatePayload
	if err := gob.NewDecoder(bytes.NewReader(decoded.Payload())).Decode(&payload); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	f.session.Put(r.Context(), "authenticated", true)
	http.Redirect(w, r, named.URLFor("root"), http.StatusSeeOther)

	spew.Dump(payload)
}

func (f *Frontend) SessionDestroyHandler(w http.ResponseWriter, r *http.Request) {
	f.session.Put(r.Context(), SessionKeyAuthenticated, false)
	f.session.Put(r.Context(), SessionKeyFlash, FlashSessionDestroyed)
	http.Redirect(w, r, "/", http.StatusFound)
}
