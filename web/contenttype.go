package web

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
)

//go:generate stringer -type=ContentType -linecomment
type ContentType int

const (
	ContentTypeText ContentType = iota // plaintext
	ContentTypeJSON                    // json
	ContentTypeXML                     // xml
	ContentTypeYAML                    // yaml
	ContentTypeHTML                    // html
)

func ParseContentType(t string) ContentType {
	switch t {
	case "json", "application/json", "text/json":
		return ContentTypeJSON
	case "xml", "application/xml", "text/xml":
		return ContentTypeXML
	case "html", "text/html":
		return ContentTypeHTML
	case "yml", "yaml", "text/yaml", "application/yaml":
		return ContentTypeYAML
	default:
		return ContentTypeText
	}
}

func contentType(r *http.Request) ContentType {
	format, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
	if format == "" {
		format = r.Header.Get("Content-type")
	}

	return ParseContentType(format)
}

func GetContentType(r *http.Request) ContentType {
	return contentType(r)
}
