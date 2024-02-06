package services

import (
	"context"
	"testing"
)

func TestSlugGeneration(t *testing.T) {
	examples := map[string]struct {
		input  string
		output string
	}{
		"simple":   {"Hello There", "hello-there"},
		"unicode":  {"Hello, 世界", "hello-shi-jie"},
		"initials": {"H.E.L.L.O.", "hello"},
		"emoji":    {"👋 Hello There!", "hello-there"},
		"rtl":      {"שָׁלוֹם", "shalvom"},
		"leet":     {"h3ll0 7h3r3", "h3ll0-7h3r3"},
	}

	for name, example := range examples {
		name := name
		example := example
		t.Run(name, func(t *testing.T) {
			if got := Slugger.Generate(context.Background(), example.input); got != example.output {
				t.Errorf("expected %q, got %q", example.output, got)
			}
		})
	}
}
