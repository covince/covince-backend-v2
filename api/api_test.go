package api

import (
	"fmt"
	"net/url"
	"testing"
)

func TestParseQuery(t *testing.T) {
	g := map[string]bool{"S": true}

	t.Run("can parse mutations", func(t *testing.T) {
		qs := url.Values{"lineages": {"B+S:V36F"}}
		q, err := parseQuery(qs, &g, 16)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(q)
	})
	t.Run("can parse empty lineages", func(t *testing.T) {
		qs := url.Values{"lineages": {""}}
		_, err := parseQuery(qs, &g, 16)
		if err != nil {
			t.Error(err)
		}
	})
}
