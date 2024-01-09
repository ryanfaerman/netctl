package views

import (
	"strings"

	scs "github.com/alexedwards/scs/v2"
	"github.com/go-loremipsum/loremipsum"
)

type HTML struct {
	title       string
	description string
	author      string

	session *scs.SessionManager
}

// func (h HTML) hasFlashMessage(ctx context.Context) bool {
// 	if h.session == nil {
// 		return false
// 	}
// 	return h.session.Exists(ctx, SessionKeyFlash)
// }
//
// func (h HTML) getFlashMessage(ctx context.Context) FlashMessage {
// 	if h.session == nil {
// 		return FlashNone
// 	}
//
// 	f, ok := h.session.Pop(ctx, SessionKeyFlash).(FlashMessage)
// 	if !ok {
// 		return FlashNone
// 	}
// 	return f
// }

func (h HTML) Paragraphs(n int) string {
	var b strings.Builder

	ipsum := loremipsum.New()
	for i := 0; i < n; i++ {
		b.WriteString("<p>")
		b.WriteString(ipsum.Paragraph())
		b.WriteString("</p>")
	}

	return b.String()
}

func (h HTML) Words(n int) string {
	var b strings.Builder
	ipsum := loremipsum.New()
	for i := 0; i < n; i++ {
		b.WriteString(ipsum.Word())
		b.WriteString(" ")
	}
	return b.String()
}

func (h HTML) Spanify(s string) string {
	var b strings.Builder
	for _, w := range strings.Split(s, " ") {
		b.WriteString("<span>")
		b.WriteString(w)
		b.WriteString("</span>")
	}
	return b.String()
}
