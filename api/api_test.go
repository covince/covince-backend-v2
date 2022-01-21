package api

import (
	"net/url"
	"testing"
)

func TestParseQuery(t *testing.T) {
	t.Run("can parse empty lineages", func(t *testing.T) {
		qs := url.Values{"lineages": {""}}
		_, err := parseQuery(qs, 16)
		if err != nil {
			t.Error(err)
		}
	})
}
