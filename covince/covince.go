package covince

import (
	"strings"
)

type Index map[string]map[string]int

type QueryLineage struct {
	Key        string
	PangoClade string
	Mutations  []Mutation
}

type Query struct {
	Lineages     []QueryLineage
	Excluding    []QueryLineage
	Area         string
	DateFrom     string
	DateTo       string
	Prefix       string
	SuffixFilter string
}

type MutationSearch struct {
	Key         string  `json:"key"`
	Count       int     `json:"count"`
	Growth      float32 `json:"growth"`
	growthStart int
	growthEnd   int
}

func matchLineages(r *Record, lineages []QueryLineage) (bool, string) {
	for _, ql := range lineages {
		if strings.HasPrefix(r.PangoClade.Value, ql.PangoClade) {
			if len(ql.Mutations) > 0 {
				hasMuts := true
				for _, qm := range ql.Mutations {
					match := false
					for _, m := range r.Mutations {
						if qm.Prefix == m.Prefix && qm.Suffix == m.Suffix {
							match = true
							break
						}
					}
					if !match {
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

func matchMetadata(r *Record, q *Query) bool {
	if q.Area != "" && q.Area != "overview" && r.Area.Value != q.Area {
		return false
	}
	if q.DateFrom != "" && r.Date.Value < q.DateFrom {
		return false
	}
	if q.DateTo != "" && r.Date.Value > q.DateTo {
		return false
	}
	return true
}

func Frequency(i Index, q *Query, r *Record) {
	if matchMetadata(r, q) {
		if ok, key := matchLineages(r, q.Lineages); ok {
			dateCounts, ok := i[r.Date.Value]
			if !ok {
				dateCounts = make(map[string]int)
				i[r.Date.Value] = dateCounts
			}
			dateCounts[key] += r.Count
		}
	}
}

func Totals(foreach func(func(r *Record)), q *Query, mutSuppressionMin int) Index {
	perLineage := make(map[string]Index)
	for _, ql := range q.Lineages {
		perLineage[ql.Key] = Index{}
	}

	foreach(func(r *Record) {
		if ok, l := matchLineages(r, q.Lineages); ok {
			i := perLineage[l]
			dateCounts, ok := i[r.Date.Value]
			if !ok {
				dateCounts = make(map[string]int)
				i[r.Date.Value] = dateCounts
			}
			dateCounts[r.Area.Value] += r.Count
		}
	})

	if mutSuppressionMin > 0 {
		for _, ql := range q.Lineages {
			if len(ql.Mutations) > 0 {
				Suppress(perLineage[ql.Key], mutSuppressionMin)
			}
		}
	}

	totals := make(Index)
	for _, i := range perLineage {
		for date, areaCounts := range i {
			dateCounts, ok := totals[date]
			if !ok {
				totals[date] = areaCounts
			} else {
				for area, count := range areaCounts {
					dateCounts[area] += count
				}
			}
		}
	}
	return totals
}

func Spatiotemporal(i Index, q *Query, r *Record) {
	if ok, _ := matchLineages(r, q.Excluding); ok {
		return
	}
	if ok, _ := matchLineages(r, q.Lineages); ok {
		dateCounts, ok := i[r.Date.Value]
		if !ok {
			dateCounts = make(map[string]int)
			i[r.Date.Value] = dateCounts
		}
		dateCounts[r.Area.Value] += r.Count
	}
}

func Lineages(m map[string]int, q *Query, r *Record) {
	if matchMetadata(r, q) {
		m[r.PangoClade.Value] += r.Count
	}
}

func Mutations(m map[string]*MutationSearch, total *MutationSearch, so *SearchOpts, q *Query, r *Record) {
	if matchMetadata(r, q) {
		if ok, l := matchLineages(r, q.Lineages); ok {
			if r.Date.Value == so.Growth.Start {
				total.growthStart += r.Count
			} else if r.Date.Value == so.Growth.End {
				total.growthEnd += r.Count
			}
			if l == so.Lineage {
				total.Count += r.Count
				for _, rm := range r.Mutations {
					if (q.Prefix == "" || q.Prefix == rm.Prefix) && (q.SuffixFilter == "" || strings.Contains(rm.Suffix, q.SuffixFilter)) {
						var sr *MutationSearch
						var ok bool
						if sr, ok = m[rm.Key]; ok {
							sr.Count += r.Count
						} else {
							sr = &MutationSearch{Key: rm.Key, Count: r.Count}
							m[rm.Key] = sr
						}
						if r.Date.Value == so.Growth.Start {
							sr.growthStart += r.Count
						} else if r.Date.Value == so.Growth.End {
							sr.growthEnd += r.Count
						}
					}
				}
			}
		}
	}
}

func Info(foreach func(func(r *Record))) ([]string, []string) {
	dates := make(map[string]bool)
	areas := make(map[string]bool)

	foreach(func(r *Record) {
		dates[r.Date.Value] = true
		areas[r.Area.Value] = true
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
