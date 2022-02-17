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
	MaxSearchResults int
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
			limit := opts.MaxSearchResults
			if _limit, ok := qs["limit"]; ok {
				i, err := strconv.Atoi(_limit[0])
				if err == nil {
					limit = i
				}
			}
			sort := "count"
			if _sort, ok := qs["sort"]; ok {
				sort = _sort[0]
			}
			dir := "desc"
			if _dir, ok := qs["direction"]; ok {
				if _dir[0] == "asc" {
					dir = _dir[0]
				}
			}
			response = covince.SearchMutations(foreach, &q, skip, limit, sort, dir)
		}

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		perf.LogMemory()
		perf.LogDuration(r.URL.Path, start)
	}
}
