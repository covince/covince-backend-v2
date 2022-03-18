package api

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/covince/covince-backend-v2/covince"
)

var isPangoLineage = regexp.MustCompile(`^[A-Z]{1,3}(\.[0-9]+)*$`)
var isDateString = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

func parseMutation(s string, genes *map[string]bool) (covince.Mutation, error) {
	var m covince.Mutation
	split := strings.Split(s, ":")
	for gene := range *genes {
		if gene == split[0] {
			m.Prefix = gene
			break
		}
	}
	if m.Prefix == "" {
		return m, fmt.Errorf("invalid gene for input: %v", s)
	}
	m.Suffix = split[1]
	return m, nil
}

func parseLineages(lineages []string, genes *map[string]bool) ([]covince.QueryLineage, error) {
	index := make(map[string]covince.QueryLineage)
	for _, v := range lineages {
		if len(v) == 0 {
			continue
		}
		split := strings.Split(v, "+")
		lineage := split[0]
		mutations := make([]covince.Mutation, 0)
		for i, m := range split[1:] {
			if i > 1 {
				break
			}
			parsed, err := parseMutation(m, genes)
			if err != nil {
				return nil, err
			}
			mutations = append(mutations, parsed)
		}
		if !isPangoLineage.MatchString(split[0]) {
			return nil, fmt.Errorf("invalid lineages")
		}
		if _, ok := index[v]; !ok {
			index[v] = covince.QueryLineage{
				Key:        v,
				PangoClade: lineage + ".",
				Mutations:  mutations,
			}
		}
	}
	parsedLineages := make([]covince.QueryLineage, len(index))
	i := 0
	for _, v := range index {
		parsedLineages[i] = v
		i++
	}
	sort.Sort(covince.SortLineagesForQuery(parsedLineages))
	return parsedLineages, nil
}

func parseQuery(qs url.Values, genes *map[string]bool, maxLineages int) (covince.Query, error) {
	var q covince.Query
	if lineage, ok := qs["lineage"]; ok {
		p, err := parseLineages(lineage, genes)
		if err != nil {
			return q, err
		}
		q.Lineages = p
	} else if lineages, ok := qs["lineages"]; ok {
		lineages = strings.Split(lineages[0], ",")
		if len(lineages) > maxLineages {
			return q, fmt.Errorf("too many lineages, maximum is %v", maxLineages)
		}
		p, err := parseLineages(lineages, genes)
		if err != nil {
			return q, err
		}
		q.Lineages = p
	}
	if a, ok := qs["area"]; ok && a[0] != "overview" {
		q.Area = a[0]
	}
	if from, ok := qs["from"]; ok && len(from[0]) > 0 {
		if !isDateString.MatchString(from[0]) {
			return q, fmt.Errorf("invalid date")
		}
		q.DateFrom = from[0]
	}
	if to, ok := qs["to"]; ok && len(to[0]) > 0 {
		if !isDateString.MatchString(to[0]) {
			return q, fmt.Errorf("invalid date")
		}
		q.DateTo = to[0]
	}
	if excluding, ok := qs["excluding"]; ok {
		excluding = strings.Split(excluding[0], ",")
		excluding, err := parseLineages(excluding, genes)
		if err != nil {
			return q, err
		}
		q.Excluding = excluding
	}
	if gene, ok := qs["gene"]; ok && len(gene[0]) > 0 {
		for g := range *genes {
			if g == gene[0] {
				q.Prefix = g
				break
			}
		}
		if q.Prefix == "" {
			return q, fmt.Errorf("gene not recognised")
		}
	}
	if filter, ok := qs["filter"]; ok && len(filter[0]) > 0 {
		if len(filter[0]) > 24 {
			return q, fmt.Errorf("filter string too long")
		}
		q.SuffixFilter = filter[0]
	}
	return q, nil
}

func parseSearchOptions(qs url.Values, defaultLimit int, suppressionMin int) covince.SearchOpts {
	so := covince.SearchOpts{
		Skip:           0,
		Limit:          defaultLimit,
		SortProperty:   "count",
		SortDirection:  "desc",
		SuppressionMin: suppressionMin,
	}

	if parent, ok := qs["parent"]; ok {
		so.Lineage = parent[0]
	}
	if skip, ok := qs["skip"]; ok {
		i, err := strconv.Atoi(skip[0])
		if err == nil {
			so.Skip = i
		}
	}
	if limit, ok := qs["limit"]; ok {
		i, err := strconv.Atoi(limit[0])
		if err == nil {
			so.Limit = i
		}
	}
	if sort, ok := qs["sort"]; ok {
		so.SortProperty = sort[0]
	}
	if dir, ok := qs["direction"]; ok {
		if dir[0] == "asc" {
			so.SortDirection = dir[0]
		}
	}
	if start, ok := qs["growthStart"]; ok {
		if end, ok := qs["growthEnd"]; ok {
			so.Growth.Start = start[0]
			so.Growth.End = end[0]
		}
	}

	return so
}
