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

func TestMultipleMuts(t *testing.T) {
	qs := url.Values{"lineages": {"B+S:V36F+S:V36H"}}
	opts := Opts{
		Genes:       map[string]bool{"S": true},
		MaxLineages: 16,
	}

	t.Run("error if multiple muts disabled", func(t *testing.T) {
		opts.MultipleMuts = false
		_, err := parseQuery(qs, &opts)
		if err == nil {
			t.Error(err)
		}
	})

	t.Run("not error if multiple muts enabled", func(t *testing.T) {
		opts.MultipleMuts = true
		_, err := parseQuery(qs, &opts)
		if err != nil {
			t.Error(err)
		}
	})
}
