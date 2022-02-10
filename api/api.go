package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/covince/covince-backend-v2/covince"
	"github.com/covince/covince-backend-v2/perf"
)

type Opts struct {
	PathPrefix       string
	MaxLineages      int
	GetLastModified  func() int64
	NumSearchResults int
}

func CovinceAPI(opts Opts, foreach func(func(r *covince.Record)), genes map[string]bool) http.HandlerFunc {
	uniqueGenes := make([]string, len(genes))
	i := 0
	for k := range genes {
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
			skip := 0
			if _skip, ok := qs["skip"]; ok {
				i, err := strconv.Atoi(_skip[0])
				if err == nil {
					skip = i
				}
			}
			limit := 10
			if opts.NumSearchResults > 0 {
				limit = opts.NumSearchResults
			}
			sort := "desc"
			if _sort, ok := qs["sort"]; ok {
				if _sort[0] == "asc" {
					sort = _sort[0]
				}
			}
			response = covince.SearchMutations(foreach, &q, skip, limit, sort)
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		perf.LogMemory()
		perf.LogDuration(r.URL.Path, start)
	}
}
