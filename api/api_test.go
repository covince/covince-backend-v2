package api

import (
	"fmt"
	"net/url"
	"testing"
)

func TestParseQuery(t *testing.T) {
	t.Run("can parse mutations", func(t *testing.T) {
		qs := url.Values{"lineages": {"B+S:V36F"}}
		q, err := parseQuery(qs, 16)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(q)
	})
	t.Run("can parse empty lineages", func(t *testing.T) {
		qs := url.Values{"lineages": {""}}
		_, err := parseQuery(qs, 16)
		if err != nil {
			t.Error(err)
		}
	})
}
