package covince

import (
	"fmt"
	"sort"
	"time"

	"github.com/covince/covince-backend-v2/perf"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type SortByCount []*MutationSearch

func (s SortByCount) Less(i, j int) bool { return s[i].Count < s[j].Count }
func (s SortByCount) Len() int           { return len(s) }
func (s SortByCount) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type SortByGrowth []*MutationSearch

func (s SortByGrowth) Less(i, j int) bool { return s[i].Growth < s[j].Growth }
func (s SortByGrowth) Len() int           { return len(s) }
func (s SortByGrowth) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type SortByName []*MutationSearch

func (s SortByName) Less(i, j int) bool { return s[i].Key < s[j].Key }
func (s SortByName) Len() int           { return len(s) }
func (s SortByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type SearchResult struct {
	TotalRows    int               `json:"total_rows"`
	TotalRecords int               `json:"total_records"`
	Page         []*MutationSearch `json:"page"`
}

type GrowthOpts struct {
	Start string
	End   string
	N     int
}

type SearchOpts struct {
	Skip          int
	Limit         int
	SortProperty  string
	SortDirection string
	Growth        GrowthOpts
	Lineage       string
}

func SearchMutations(foreach func(func(r *Record)), q *Query, opts SearchOpts) SearchResult {
	m := make(map[string]*MutationSearch)
	totalRecords := MutationSearch{}
	foreach(func(r *Record) {
		Mutations(m, &totalRecords, &opts, q, r)
	})
	fmt.Println("num muts:", len(m))
	startSort := time.Now()
	ms := make([]*MutationSearch, len(m))
	i := 0
	for _, sr := range m {
		if totalRecords.growthStart > 0 && totalRecords.growthEnd > 0 {
			growthStart := float32(sr.growthStart) / float32(totalRecords.growthStart)
			growthEnd := float32(sr.growthEnd) / float32(totalRecords.growthEnd)
			sr.Growth = growthEnd - growthStart
			// if growthStart > 0 {
			// 	sr.Growth = (growthEnd - growthStart) / growthStart
			// }
		}
		ms[i] = sr
		i++
	}
	var sorter sort.Interface
	if opts.SortProperty == "name" {
		sorter = SortByName(ms)
	} else if opts.SortProperty == "change" {
		sorter = SortByGrowth(ms)
	} else {
		sorter = SortByCount(ms)
	}
	if opts.SortDirection == "asc" {
		sort.Sort(sorter)
	} else {
		sort.Sort(sort.Reverse(sorter))
	}
	result := SearchResult{
		TotalRows:    len(ms),
		TotalRecords: totalRecords.Count,
		Page:         make([]*MutationSearch, opts.Limit),
	}
	i = opts.Skip
	end := min(opts.Skip+opts.Limit, len(ms))
	for i < end {
		sr := ms[i]
		result.Page[i-opts.Skip] = sr
		i++
	}
	perf.LogDuration("sorting", startSort)
	return result
}
