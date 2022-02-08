package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/covince/covince-backend-v2/covince"
	"github.com/covince/covince-backend-v2/perf"
)

type Opts struct {
	PathPrefix      string
	MaxLineages     int
	GetLastModified func() int64
}

var isPangoLineage = regexp.MustCompile(`^[A-Z]{1,3}(\.[0-9]+)*$`)
var isDateString = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

func parseMutation(s string, genes *map[string]bool) (covince.Mutation, error) {
	var m covince.Mutation
	split := strings.Split(s, ":")
	if len(split) == 1 || len(split[1]) == 0 {
		return m, fmt.Errorf("invalid mutation: %v", s)
	}
	for gene, _ := range *genes {
		if gene == split[0] {
			m.Prefix = gene
			break
		}
	}
	if m.Prefix == "" {
		return m, fmt.Errorf("invalid gene: %v", split[0])
	}
	m.Suffix = split[1]
	return m, nil
}

func parseLineages(lineages []string, genes *map[string]bool) ([]covince.QueryLineage, error) {
	index := make(map[string]covince.QueryLineage)
	for _, v := range lineages {
		if len(v) == 0 {
			continue
		}
		split := strings.Split(v, "+")
		lineage := split[0]
		mutations := make([]covince.Mutation, 0)
		for i, m := range split[1:] {
			if i > 1 {
				break
			}
			parsed, err := parseMutation(m, genes)
			if err != nil {
				return nil, err
			}
			mutations = append(mutations, parsed)
		}
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
	}
	parsedLineages := make([]covince.QueryLineage, len(index))
	i := 0
	for _, v := range index {
		parsedLineages[i] = v
		i++
	}
	sort.Sort(covince.SortLineagesForQuery(parsedLineages))
	return parsedLineages, nil
}

func parseQuery(qs url.Values, genes *map[string]bool, maxLineages int) (covince.Query, error) {
	var q covince.Query
	if lineage, ok := qs["lineage"]; ok {
		p, err := parseLineages(lineage, genes)
		if err != nil {
			return q, err
		}
		q.Lineages = p
	} else if lineages, ok := qs["lineages"]; ok {
		lineages = strings.Split(lineages[0], ",")
		if len(lineages) > maxLineages {
			return q, fmt.Errorf("too many lineages, maximum is %v", maxLineages)
		}
		p, err := parseLineages(lineages, genes)
		if err != nil {
			return q, err
		}
		q.Lineages = p
	}
	if a, ok := qs["area"]; ok && a[0] != "overview" {
		q.Area = a[0]
	}
	if from, ok := qs["from"]; ok && len(from[0]) > 0 {
		if !isDateString.MatchString(from[0]) {
			return q, fmt.Errorf("invalid date")
		}
		q.DateFrom = from[0]
	}
	if to, ok := qs["to"]; ok && len(to[0]) > 0 {
		if !isDateString.MatchString(to[0]) {
			return q, fmt.Errorf("invalid date")
		}
		q.DateTo = to[0]
	}
	if excluding, ok := qs["excluding"]; ok {
		excluding = strings.Split(excluding[0], ",")
		excluding, err := parseLineages(excluding, genes)
		if err != nil {
			return q, err
		}
		q.Excluding = excluding
	}
	if search, ok := qs["search"]; ok {
		if len(search[0]) > 24 {
			return q, fmt.Errorf("search string too long")
		}
		m, err := parseMutation(search[0], genes)
		if err != nil {
			return q, err
		}
		q.Mutation = m
	}
	return q, nil
}

func CovinceAPI(opts Opts, foreach func(func(r *covince.Record)), genes map[string]bool) http.HandlerFunc {
	uniqueGenes := make([]string, len(genes))
	i := 0
	for k, _ := range genes {
		uniqueGenes[i] = k
		i++
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Println("Handle request")

		if r.Method != "GET" {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		qs := r.URL.Query()
		q, err := parseQuery(qs, &genes, opts.MaxLineages)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		var response interface{}

		if r.URL.Path == opts.PathPrefix+"/info" {
			dates, areas := covince.Info(foreach)

			m := make(map[string]interface{})
			m["dates"] = dates
			m["areas"] = areas
			m["lastModified"] = opts.GetLastModified()
			m["maxLineages"] = opts.MaxLineages
			m["genes"] = uniqueGenes

			response = m
		}
		if r.URL.Path == opts.PathPrefix+"/frequency" {
			i := make(covince.Index)
			foreach(func(r *covince.Record) {
				covince.Frequency(i, &q, r)
			})
			response = i
		}
		if r.URL.Path == opts.PathPrefix+"/spatiotemporal/total" {
			i := make(covince.Index)
			foreach(func(r *covince.Record) {
				covince.Totals(i, &q, r)
			})
			response = i
		}
		if r.URL.Path == opts.PathPrefix+"/spatiotemporal/lineage" {
			i := make(covince.Index)
			foreach(func(r *covince.Record) {
				covince.Spatiotemporal(i, &q, r)
			})
			response = i
		}
		if r.URL.Path == opts.PathPrefix+"/lineages" {
			m := make(map[string]int)
			foreach(func(r *covince.Record) {
				covince.Lineages(m, &q, r)
			})
			response = m
		}
		if r.URL.Path == opts.PathPrefix+"/mutations" {
			m := make(map[string]int)
			foreach(func(r *covince.Record) {
				covince.Mutations(m, &q, r)
			})
			response = m
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		perf.LogMemory()
		perf.LogDuration(r.URL.Path, start)
	}
}
