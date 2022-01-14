package covince

import (
	"strings"
)

type Record struct {
	Date       string
	Lineage    string
	PangoClade string
	Area       string
	Mutations  string
	Count      int
}

type Index map[string]map[string]int

type QueryLineage struct {
	Key        string
	PangoClade string
	Mutations  []string
}

type Query struct {
	Lineages  []QueryLineage
	Excluding []string
	Area      string
	DateFrom  string
	DateTo    string
}

func matchLineages(r Record, q Query) (bool, string) {
	for _, ql := range q.Lineages {
		if strings.HasPrefix(r.PangoClade, ql.PangoClade) {
			if len(ql.Mutations) > 0 {
				hasMuts := true
				for _, m := range ql.Mutations {
					if !strings.Contains(r.Mutations, m) {
						hasMuts = false
						break
					}
				}
				if !hasMuts {
					continue
				}
			}
			return true, ql.Key
		}
	}
	return false, ""
}

func matchMetadata(r Record, q Query) bool {
	if q.Area != "" && r.Area != q.Area {
		return false
	}
	if q.DateFrom != "" && r.Date < q.DateFrom {
		return false
	}
	if q.DateTo != "" && r.Date > q.DateTo {
		return false
	}
	return true
}

func Frequency(i Index, q Query, r Record) {
	if matchMetadata(r, q) {
		if ok, key := matchLineages(r, q); ok {
			dateCounts, ok := i[r.Date]
			if !ok {
				dateCounts = make(map[string]int)
				i[r.Date] = dateCounts
			}
			dateCounts[key] += r.Count
		}
	}
}

func Totals(i Index, q Query, r Record) {
	if ok, _ := matchLineages(r, q); ok {
		dateCounts, ok := i[r.Date]
		if !ok {
			dateCounts = make(map[string]int)
			i[r.Date] = dateCounts
		}
		dateCounts[r.Area] += r.Count
	}
}

func Spatiotemporal(i Index, q Query, r Record) {
	for _, p := range q.Excluding {
		if strings.HasPrefix(r.PangoClade, p) {
			return
		}
	}
	if ok, _ := matchLineages(r, q); ok {
		dateCounts, ok := i[r.Date]
		if !ok {
			dateCounts = make(map[string]int)
			i[r.Date] = dateCounts
		}
		dateCounts[r.Area] += r.Count
	}
}

func Lineages(m map[string]int, q Query, r Record) {
	if matchMetadata(r, q) {
		m[r.Lineage] += r.Count
	}
}
