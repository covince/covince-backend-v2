package api

import (
	"fmt"
	"net/url"
	"testing"
)

func TestParseQuery(t *testing.T) {
	opts := Opts{
		Genes:       map[string]bool{"S": true},
		MaxLineages: 16,
	}

	t.Run("can parse mutations", func(t *testing.T) {
		qs := url.Values{"lineages": {"B+S:V36F"}}
		q, err := parseQuery(qs, &opts)
		if err != nil {
			t.Error(err)
		}
		fmt.Println(q)
	})
	t.Run("can parse empty lineages", func(t *testing.T) {
		qs := url.Values{"lineages": {""}}
		_, err := parseQuery(qs, &opts)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestSingleMuts(t *testing.T) {
	opts := Opts{
		Genes:       map[string]bool{"S": true},
		MaxLineages: 16,
		SingleMuts:  true,
	}

	qs := url.Values{"lineages": {"B+S:V36F+S:V36H"}}
	_, err := parseQuery(qs, &opts)
	if err == nil {
		t.Error(err)
	}
}
