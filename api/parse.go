package api

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/covince/covince-backend-v2/covince"
)

var isPangoLineage = regexp.MustCompile(`^[A-Z]{1,3}(\.[0-9]+)*$`)
var isDateString = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

func parseMutation(s string, genes *map[string]bool) (covince.Mutation, error) {
	var m covince.Mutation
	split := strings.Split(s, ":")
	if len(split) == 1 || len(split[1]) == 0 {
		return m, fmt.Errorf("invalid mutation: %v", s)
	}
	for gene := range *genes {
		if gene == split[0] {
			m.Prefix = gene
			break
		}
	}
	if m.Prefix == "" {
		return m, fmt.Errorf("invalid gene: %v", split[0])
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
	if search, ok := qs["search"]; ok {
		if len(search[0]) > 24 {
			return q, fmt.Errorf("search string too long")
		}
		m, err := parseMutation(search[0], genes)
		if err != nil {
			return q, err
		}
		q.Mutation = m
	}
	return q, nil
}
