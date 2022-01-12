package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/covince/covince-backend-v2/covince"
)

type Opts struct {
	DataFile   string
	InfoFile   string
	PathPrefix string
}

func parseQuery(qs map[string][]string) covince.Query {
	var q covince.Query
	if _l, ok := qs["lineages"]; ok {
		lineages := strings.Split(_l[0], ",")
		sort.Sort(sort.Reverse(sort.StringSlice(lineages))) // TODO: implement real sort
		parsedLineages := make([]covince.QueryLineage, len(lineages))
		for i, v := range lineages {
			split := strings.Split(v, "+")
			parsedLineages[i] = covince.QueryLineage{
				Key:        v,
				PangoClade: split[0],
				Mutations:  split[1:],
			}
		}
		q.Lineages = parsedLineages
	}
	if _a, ok := qs["area"]; ok {
		q.Area = _a[0]
	}
	fmt.Println(q, qs)
	return q
}

func CovinceAPI(opts Opts, s []covince.Record) {
	http.HandleFunc(opts.PathPrefix+"/frequency", func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		qs := r.URL.Query()
		q := parseQuery(qs)

		i := make(covince.Index)
		for _, r := range s {
			covince.Frequency(i, q, r)
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(i)
	})

	http.HandleFunc(opts.PathPrefix+"/spatiotemporal/total", func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if r.Method != "GET" {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		qs := r.URL.Query()
		q := parseQuery(qs)

		i := make(covince.Index)
		for _, r := range s {
			covince.Totals(i, q, r)
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(i)

		duration := time.Since(start)
		fmt.Println(duration.Milliseconds(), "ms")
	})

	http.HandleFunc(opts.PathPrefix+"/spatiotemporal/lineage", func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if r.Method != "GET" {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		qs := r.URL.Query()
		q := parseQuery(qs)

		i := make(covince.Index)
		for _, r := range s {
			covince.Spatiotemporal(i, q, r)
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(i)

		duration := time.Since(start)
		fmt.Println(duration.Milliseconds(), "ms")
	})

	http.HandleFunc(opts.PathPrefix+"/lineages", func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if r.Method != "GET" {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		qs := r.URL.Query()
		q := parseQuery(qs)

		m := make(map[string]int)
		for _, r := range s {
			covince.Lineages(m, q, r)
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(m)

		duration := time.Since(start)
		fmt.Println(duration.Milliseconds(), "ms")
	})
}
