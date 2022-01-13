package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/covince/covince-backend-v2/covince"
)

type Opts struct {
	InfoFile    string
	PathPrefix  string
	MaxLineages int
}

var isPangoLineage = regexp.MustCompile(`^[A-Z]{1,3}(\.[0-9]+)*$`)
var isDateString = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

func parseLineages(lineages []string) ([]covince.QueryLineage, error) {
	index := make(map[string]covince.QueryLineage)
	for _, v := range lineages {
		split := strings.Split(v, "+")
		lineage := split[0]
		mutations := split[1:] // TODO: first two mutations only
		if !isPangoLineage.MatchString(split[0]) {
			return nil, fmt.Errorf("invalid lineages")
		}
		if _, ok := index[v]; !ok {
			index[v] = covince.QueryLineage{
				Key:        v,
				PangoClade: lineage + ".",
				Mutations:  mutations,
			}
		}
		if _, ok := index[lineage]; len(mutations) > 0 && !ok {
			index[lineage] = covince.QueryLineage{
				Key:        lineage,
				PangoClade: lineage + ".",
			}
		}
	}
	keys := make([]string, len(index))
	i := 0
	for k := range index {
		keys[i] = k
		i++
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys))) // TODO: implement real sort
	parsedLineages := make([]covince.QueryLineage, len(keys))
	for i, k := range keys {
		parsedLineages[i] = index[k]
	}
	return parsedLineages, nil
}

func parseQuery(qs map[string][]string, maxLineages int) (covince.Query, error) {
	var q covince.Query
	if lineages, ok := qs["lineages"]; ok {
		lineages = strings.Split(lineages[0], ",")
		if len(lineages) > maxLineages {
			return q, fmt.Errorf("too many lineages")
		}
		p, err := parseLineages(lineages)
		if err != nil {
			return q, err
		}
		q.Lineages = p
	}
	if a, ok := qs["area"]; ok {
		q.Area = a[0]
	}
	if from, ok := qs["from"]; ok {
		if !isDateString.MatchString(from[0]) {
			return q, fmt.Errorf("invalid date")
		}
		q.DateFrom = from[0]
	}
	if to, ok := qs["to"]; ok {
		if !isDateString.MatchString(to[0]) {
			return q, fmt.Errorf("invalid date")
		}
		q.DateTo = to[0]
	}
	if excluding, ok := qs["excluding"]; ok {
		excluding = strings.Split(excluding[0], ",")
		for _, lineage := range excluding {
			if !isPangoLineage.MatchString(lineage) {
				return q, fmt.Errorf("invalid lineages")
			}
		}
	}
	fmt.Println(q, qs)
	return q, nil
}

func CovinceAPI(opts Opts, foreach func(func(r covince.Record))) {
	http.HandleFunc(opts.PathPrefix+"/", func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if r.Method != "GET" {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		qs := r.URL.Query()
		q, err := parseQuery(qs, opts.MaxLineages)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		var response interface{}

		if r.URL.Path == opts.PathPrefix+"/frequency" {
			i := make(covince.Index)
			foreach(func(r covince.Record) {
				covince.Frequency(i, q, r)
			})
			response = i
		}
		if r.URL.Path == opts.PathPrefix+"/spatiotemporal/total" {
			i := make(covince.Index)
			foreach(func(r covince.Record) {
				covince.Totals(i, q, r)
			})
			response = i
		}
		if r.URL.Path == opts.PathPrefix+"/spatiotemporal/lineage" {
			i := make(covince.Index)
			foreach(func(r covince.Record) {
				covince.Spatiotemporal(i, q, r)
			})
			response = i
		}
		if r.URL.Path == opts.PathPrefix+"/lineages" {
			m := make(map[string]int)
			foreach(func(r covince.Record) {
				covince.Lineages(m, q, r)
			})
			response = m
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		duration := time.Since(start)
		fmt.Println(duration.Milliseconds(), "ms")
	})
}
