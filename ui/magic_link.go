package ui

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

type MagicLinkInput struct {
	Email string `validate:"required,email"`
}

type MagicLinkErrors struct {
	Email string
}

func MagicLinkNewHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(session.GetString(r.Context(), "flash"))
	fmt.Println("authenticed?", session.GetBool(r.Context(), "authenticated"))
	render(Index())(w, r)
}

type MagicLinkPayload struct {
	SessionToken string
	Email        string
}

func init() {
	gob.Register(MagicLinkPayload{})
}

func MagicLinkCreateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	inputErrs := MagicLinkErrors{}

	input := MagicLinkInput{
		Email: r.Form.Get("email"),
	}

	if err := validate.Struct(input); err != nil {
		errs := err.(validator.ValidationErrors)
		for field, e := range errs.Translate(trans) {
			switch field {
			case "MagicLinkInput.Email":
				inputErrs.Email = e
			}
		}

		session.Put(r.Context(), "flash", "hello")

		if r.Header.Get("HX-Request") != "" {
			web.LogWith(r.Context(), "hx", "true")
			ctx := context.WithValue(r.Context(), ctxToken, nosurf.Token(r))
			MagicLinkFormWithErrors(input, inputErrs).Render(ctx, w)
			return
		}

		http.Redirect(w, r, named.URLFor("index"), http.StatusSeeOther)

	}
	// success

	payload := MagicLinkPayload{
		SessionToken: session.Token(r.Context()),
		Email:        input.Email,
	}

	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(payload); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	token, err := brc.EncodeToString(b.Bytes())
	if err != nil {
		http.Error(w, "Unable to generate magic link", http.StatusInternalServerError)
		return
	}

	web.LogWith(r.Context(), "magic-token", token)

	if r.Header.Get("HX-Request") != "" {
		MagicLinkSent().Render(r.Context(), w)
	} else {
		http.Redirect(w, r, named.URLFor("magic-link-sent"), http.StatusSeeOther)
	}

}

func MagicLinkSentHandler(w http.ResponseWriter, r *http.Request) {
	render(MagicLinkSent())(w, r)
}

func MagicLinkVerifyHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if len(token) == 0 {
		session.Put(r.Context(), "flash", "Invalid Magic Link")
		http.Redirect(w, r, named.URLFor("index"), http.StatusSeeOther)
		return
	}

	decoded, err := brc.DecodeString(token)
	if err != nil {
		http.Error(w, "Invalid Magic link", http.StatusForbidden)
		return
	}

	fmt.Println(decoded.Timestamp())
	if decoded.IsExpired(60) {
		http.Error(w, "Magic link expired", http.StatusForbidden)
		return
	}

	var payload MagicLinkPayload
	if err := gob.NewDecoder(bytes.NewReader(decoded.Payload())).Decode(&payload); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session.Put(r.Context(), "authenticated", true)

	spew.Dump(payload)

}
