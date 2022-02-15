package covince

import (
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

type SortByName []*MutationSearch

func (s SortByName) Less(i, j int) bool { return s[i].Key < s[j].Key }
func (s SortByName) Len() int           { return len(s) }
func (s SortByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type SearchResult struct {
	Total int               `json:"total"`
	Page  []*MutationSearch `json:"page"`
}

func SearchMutations(foreach func(func(r *Record)), q *Query, skip int, limit int, sortProperty string, sortDirection string) SearchResult {
	m := make(map[string]*MutationSearch)
	foreach(func(r *Record) {
		Mutations(m, q, r)
	})
	startSort := time.Now()
	ms := make([]*MutationSearch, len(m))
	i := 0
	for _, sr := range m {
		ms[i] = sr
		i++
	}
	var sorter sort.Interface
	if sortProperty == "name" {
		sorter = SortByName(ms)
	} else {
		sorter = SortByCount(ms)
	}
	if sortDirection == "asc" {
		sort.Sort(sorter)
	} else {
		sort.Sort(sort.Reverse(sorter))
	}
	result := SearchResult{
		Total: len(ms),
		Page:  make([]*MutationSearch, limit),
	}
	i = skip
	end := min(skip+limit, len(ms))
	for i < end {
		sr := ms[i]
		result.Page[i-skip] = sr
		i++
	}
	perf.LogDuration("sorting", startSort)
	return result
}
