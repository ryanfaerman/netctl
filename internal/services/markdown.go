package services

import (
	"bytes"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type markdown struct {
	md        goldmark.Markdown
	sanitizer *bluemonday.Policy
}

var Markdown = markdown{
	md: goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
		goldmark.WithExtensions(emoji.Emoji),
	),
	sanitizer: bluemonday.UGCPolicy(),
}

func (s markdown) Render(input []byte) ([]byte, error) {
	var b bytes.Buffer
	if err := s.md.Convert(input, &b); err != nil {
		return []byte{}, err
	}
	return s.sanitizer.SanitizeBytes(b.Bytes()), nil
}

func (s markdown) MustRender(input []byte) []byte {
	output, err := s.Render(input)
	if err != nil {
		global.log.Error("unable to render markdown", "error", err)
	}
	return output
}

func (s markdown) RenderString(input string) (string, error) {
	output, err := s.Render([]byte(input))
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (s markdown) MustRenderString(input string) string {
	output, err := s.RenderString(input)
	if err != nil {
		global.log.Error("unable to render markdown", "error", err)
	}
	return output
}
