package services

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/forPelevin/gomoji"
	"github.com/mozillazg/go-unidecode"
)

type slugger struct{}

var Slugger slugger

var (
	regexpNonAuthorizedChars = regexp.MustCompile("[^a-zA-Z0-9-_]")
	regexpMultipleDashes     = regexp.MustCompile("-+")
)

func (s slugger) Generate(ctx context.Context, name string) string {
	name = unidecode.Unidecode(name)
	name = gomoji.RemoveEmojis(name)
	name = strings.ToLower(name)

	name = strings.ReplaceAll(name, ".", "")
	name = regexpNonAuthorizedChars.ReplaceAllString(name, "-")
	name = regexpMultipleDashes.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-_")

	return name
}

func (s slugger) ValidateUniqueForAccount(ctx context.Context, slug string) (string, error) {
	slug = s.Generate(ctx, slug)
	count, err := global.dao.CheckSlugAvailability(ctx, slug)
	if err != nil {
		return slug, err
	}
	if count != 0 {
		return slug, errors.New("slug already exists")
	}

	return slug, nil
}
