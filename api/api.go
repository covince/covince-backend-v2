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
	Genes             map[string]bool
	LastModified      int64
	MaxLineages       int
	MaxSearchResults  int
	MultipleMuts      bool
	MutSuppressionMin int
	PathPrefix        string
	Threads           int
}

func getInfo(opts *Opts, foreach covince.IteratorFunc) map[string]interface{} {
	m := make(map[string]interface{})

	m["lastModified"] = opts.LastModified
	m["maxLineages"] = opts.MaxLineages

	dates, areas := covince.Info(foreach)
	m["dates"] = dates
	m["areas"] = areas

	uniqueGenes := make([]string, len(opts.Genes))
	i := 0
	for k := range opts.Genes {
		uniqueGenes[i] = k
		i++
	}
	m["genes"] = uniqueGenes

	return m
}

func CovinceAPI(opts Opts, foreach covince.IteratorFunc) http.HandlerFunc {
	cachedInfo := getInfo(&opts, foreach)

	return func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Println("Handle request")

		if r.Method != "GET" {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		qs := r.URL.Query()
		q, err := parseQuery(qs, &opts)
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
				covince.Frequency(i, q, r)
			}, -1)
			if opts.MutSuppressionMin > 0 {
				covince.SuppressMutations(i, opts.MutSuppressionMin)
			}
			response = i
		}
		if r.URL.Path == opts.PathPrefix+"/spatiotemporal/total" {
			i := covince.Totals(foreach, q, opts.MutSuppressionMin)
			response = i
		}
		if r.URL.Path == opts.PathPrefix+"/spatiotemporal/lineage" {
			if len(q.Lineages) != 1 {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			i := make(covince.Index)
			foreach(func(r *covince.Record) {
				covince.Spatiotemporal(i, q, r)
			}, -1)
			if opts.MutSuppressionMin > 0 && len(q.Lineages[0].Mutations) > 0 {
				covince.Suppress(i, opts.MutSuppressionMin)
			}
			response = i
		}
		if r.URL.Path == opts.PathPrefix+"/lineages" {
			m := make(map[string]int)
			foreach(func(r *covince.Record) {
				covince.Lineages(m, q, r)
			}, -1)
			response = m
		}

		if r.URL.Path == opts.PathPrefix+"/mutations" {
			searchOpts := parseSearchOptions(qs, opts.MaxSearchResults)
			searchOpts.SuppressionMin = opts.MutSuppressionMin
			searchOpts.Threads = opts.Threads
			response = covince.SearchMutations(foreach, q, searchOpts)
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		perf.LogMemory()
		perf.LogDuration(r.URL.Path, start)
	}
}
