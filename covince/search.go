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

type SortByCount []*SearchResult

func (s SortByCount) Len() int           { return len(s) }
func (s SortByCount) Less(i, j int) bool { return s[i].Count < s[j].Count }
func (s SortByCount) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func SearchMutations(foreach func(func(r *Record)), q *Query, skip int, limit int, sortOrder string) map[string]int {
	m := make(map[string]*SearchResult)
	foreach(func(r *Record) {
		Mutations(m, q, r)
	})
	startSort := time.Now()
	results := make([]*SearchResult, 0)
	for _, sr := range m {
		results = append(results, sr)
	}
	if sortOrder == "asc" {
		sort.Sort(SortByCount(results))
	} else {
		sort.Sort(sort.Reverse(SortByCount(results)))
	}
	m2 := make(map[string]int)
	i := skip
	end := min(skip+limit, len(results))
	for i < end {
		sr := results[i]
		m2[sr.Key] = sr.Count
		i++
	}
	perf.LogDuration("sorting", startSort)
	return m2
}
