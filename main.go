package main

import (
	"dario.cat/mergo"
	"github.com/davecgh/go-spew/spew"
	"github.com/ryanfaerman/netctl/internal/models"
	"github.com/ryanfaerman/netctl/internal/services"
)

func show(s models.Settings) {
	spew.Dump(s)
}

func main() {
	s := &models.PrivacySettings{
		Location: "PROTECTED",
	}
	if err := mergo.Merge(s, s.Defaults()); err != nil {
		panic(err.Error())
	}

	if err := services.Validation.Apply(s); err != nil {
		spew.Dump(err)
	}

	show(models.Settings(s))
}
