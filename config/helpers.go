package config

import "strings"

func (c *config) escapeUri(uri string) string {
	uri = strings.ToLower(uri)
	uri = strings.TrimSpace(uri)
	uri = strings.Replace(uri, ".", "_", -1)

	return uri
}

func (c *config) unescapeUri(uri string) string {
	uri = strings.Replace(uri, "_", ".", -1)
	return uri
}
