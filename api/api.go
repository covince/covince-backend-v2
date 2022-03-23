package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/covince/covince-backend-v2/covince"
	"github.com/covince/covince-backend-v2/perf"
)

type Opts struct {
	PathPrefix        string
	MaxLineages       int
	GetLastModified   func() int64
	MaxSearchResults  int
	MutSuppressionMin int
}

func getInfo(opts Opts, foreach func(func(r *covince.Record)), genes map[string]bool) map[string]interface{} {
	m := make(map[string]interface{})

	m["lastModified"] = opts.GetLastModified()
	m["maxLineages"] = opts.MaxLineages

	dates, areas := covince.Info(foreach)
	m["dates"] = dates
	m["areas"] = areas

	uniqueGenes := make([]string, len(genes))
	i := 0
	for k := range genes {
		uniqueGenes[i] = k
		i++
	}
	m["genes"] = uniqueGenes

	return m
}

func CovinceAPI(opts Opts, foreach func(func(r *covince.Record)), genes map[string]bool) http.HandlerFunc {
	cachedInfo := getInfo(opts, foreach, genes)

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
			response = cachedInfo
		}
		if r.URL.Path == opts.PathPrefix+"/frequency" {
			i := make(covince.Index)
			foreach(func(r *covince.Record) {
				covince.Frequency(i, &q, r)
			})
			if opts.MutSuppressionMin > 0 {
				suppressMutations(i, opts.MutSuppressionMin)
			}
			response = i
		}
		if r.URL.Path == opts.PathPrefix+"/spatiotemporal/total" {
			i := make(covince.Index)
			foreach(func(r *covince.Record) {
				covince.Totals(i, &q, r)
			})
			if opts.MutSuppressionMin > 0 {
				suppress(i, opts.MutSuppressionMin, q.Lineages)
			}
			response = i
		}
		if r.URL.Path == opts.PathPrefix+"/spatiotemporal/lineage" {
			if len(q.Lineages) != 1 {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			i := make(covince.Index)
			foreach(func(r *covince.Record) {
				covince.Spatiotemporal(i, &q, r)
			})
			if opts.MutSuppressionMin > 0 {
				suppress(i, opts.MutSuppressionMin, q.Lineages)
			}
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
			searchOpts := parseSearchOptions(qs, opts.MaxSearchResults)
			searchOpts.SuppressionMin = opts.MutSuppressionMin
			response = covince.SearchMutations(foreach, &q, searchOpts)
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		perf.LogMemory()
		perf.LogDuration(r.URL.Path, start)
	}
}
