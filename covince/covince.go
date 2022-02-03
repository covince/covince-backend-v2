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
	Excluding []QueryLineage
	Area      string
	DateFrom  string
	DateTo    string
	Search    string
}

func matchLineages(r Record, lineages []QueryLineage) (bool, string) {
	for _, ql := range lineages {
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
	if q.Area != "" && q.Area != "overview" && r.Area != q.Area {
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
		if ok, key := matchLineages(r, q.Lineages); ok {
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
	if ok, _ := matchLineages(r, q.Lineages); ok {
		dateCounts, ok := i[r.Date]
		if !ok {
			dateCounts = make(map[string]int)
			i[r.Date] = dateCounts
		}
		dateCounts[r.Area] += r.Count
	}
}

func Spatiotemporal(i Index, q Query, r Record) {
	if ok, _ := matchLineages(r, q.Excluding); ok {
		return
	}
	if ok, _ := matchLineages(r, q.Lineages); ok {
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
		m[r.PangoClade] += r.Count
	}
}

func Mutations(m map[string]int, q Query, r Record) {
	if matchMetadata(r, q) {
		if ok, _ := matchLineages(r, q.Lineages); ok {
			j := 0
			for {
				i := strings.Index(r.Mutations[j:], q.Search) + 1
				if i == 0 {
					break
				}
				j = i + strings.Index(r.Mutations[i:], "|")
				mut := r.Mutations[i:j]
				m[mut] += r.Count
			}
		}
	}
}

func Info(foreach func(func(r Record))) ([]string, []string) {
	dates := make(map[string]bool)
	areas := make(map[string]bool)

	foreach(func(r Record) {
		dates[r.Date] = true
		areas[r.Area] = true
	})

	dateArray := make([]string, len(dates))
	i := 0
	for k := range dates {
		dateArray[i] = k
		i++
	}
	areaArray := make([]string, len(areas))
	i = 0
	for k := range areas {
		areaArray[i] = k
		i++
	}
	return dateArray, areaArray
}
