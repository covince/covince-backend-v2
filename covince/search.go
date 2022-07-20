package covince

import (
	"fmt"
	"sort"
	"sync"
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

func (s SortByCount) Less(i, j int) bool {
	if s[i].Count == s[j].Count {
		return SortByName(s).Less(i, j)
	}
	return s[i].Count < s[j].Count
}
func (s SortByCount) Len() int      { return len(s) }
func (s SortByCount) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type SortByGrowth []*MutationSearch

func (s SortByGrowth) Less(i, j int) bool {
	if s[i].Growth == s[j].Growth {
		return SortByName(s).Less(i, j)
	}
	return s[i].Growth < s[j].Growth
}
func (s SortByGrowth) Len() int      { return len(s) }
func (s SortByGrowth) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

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
	Skip           int
	Limit          int
	SortProperty   string
	SortDirection  string
	Growth         GrowthOpts
	Lineage        string
	SuppressionMin int
	Threads        int
}

func SearchMutations(foreach IteratorFunc, q *Query, opts *SearchOpts) SearchResult {
	m := make(map[string]*MutationSearch)
	totalRecords := MutationSearch{}

	if opts.Threads > 1 {
		var wg sync.WaitGroup
		wg.Add(opts.Threads)
		results := make([]map[string]*MutationSearch, opts.Threads)
		totals := make([]MutationSearch, opts.Threads)
		for i := 0; i < opts.Threads; i++ {
			go func(slice int) {
				m := make(map[string]*MutationSearch)
				results[slice] = m
				foreach(func(r *Record) {
					Mutations(m, &totals[slice], opts, q, r)
				}, slice)
				wg.Done()
			}(i)
		}
		wg.Wait()
		startSum := time.Now()
		for i := 0; i < opts.Threads; i++ {
			for k, v := range results[i] {
				if sr, ok := m[k]; ok {
					sr.Count += v.Count
				} else {
					m[k] = v
				}
			}
			t := totals[i]
			totalRecords.Count += t.Count
			totalRecords.growthStart += t.growthStart
			totalRecords.growthEnd += t.growthEnd
		}
		perf.LogDuration("summing", startSum)
	} else {
		foreach(func(r *Record) {
			Mutations(m, &totalRecords, opts, q, r)
		}, -1)
	}

	fmt.Println("num muts:", len(m))
	startSort := time.Now()
	ms := make([]*MutationSearch, len(m))
	i := 0
	suppressed := 0
	for _, sr := range m {
		if sr.Count < opts.SuppressionMin {
			suppressed++
			i++
			continue
		}
		if totalRecords.growthStart > 0 && totalRecords.growthEnd > 0 {
			growthStart := float32(sr.growthStart) / float32(totalRecords.growthStart)
			growthEnd := float32(sr.growthEnd) / float32(totalRecords.growthEnd)
			sr.Growth = growthEnd - growthStart
		}
		ms[i] = sr
		i++
	}
	if suppressed > 0 {
		_ms := ms
		ms = make([]*MutationSearch, len(ms)-suppressed)
		skipped := 0
		for i, sr := range _ms {
			if sr == nil {
				skipped++
			} else {
				ms[i-skipped] = sr
			}
		}
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
