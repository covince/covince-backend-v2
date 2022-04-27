package covince

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testMutations = []Mutation{
	{Key: "A:A", Prefix: "A", Suffix: "A"},
	{Key: "B:B", Prefix: "B", Suffix: "B"},
	{Key: "C:C", Prefix: "C", Suffix: "C"},
}

func value(s string) *Value {
	return &Value{Value: s}
}

var testRecords = []Record{
	{PangoClade: value("B."), Date: value("2020-09-01"), Area: value("A"), Count: 1, Mutations: []*Mutation{&testMutations[0]}},
	{PangoClade: value("B.1."), Date: value("2020-10-01"), Area: value("B"), Count: 2, Mutations: []*Mutation{&testMutations[0], &testMutations[1]}},
	{PangoClade: value("B.1.2."), Date: value("2020-11-01"), Area: value("C"), Count: 3, Mutations: []*Mutation{&testMutations[0], &testMutations[1], &testMutations[2]}},
}

func TestFrequency(t *testing.T) {
	var i Index
	var q Query

	t.Run("Roll up to B", func(t *testing.T) {
		i = Index{}
		q = Query{
			Lineages: []QueryLineage{
				{Key: "B", PangoClade: "B."},
			},
		}
		for _, r := range testRecords {
			Frequency(i, &q, &r)
		}
		assert.Equal(t, Index{
			"2020-09-01": {"B": 1},
			"2020-10-01": {"B": 2},
			"2020-11-01": {"B": 3},
		}, i)
	})

	t.Run("Roll up to B, filter by area", func(t *testing.T) {
		i = Index{}
		q = Query{
			Lineages: []QueryLineage{
				{Key: "B", PangoClade: "B."},
			},
			Area: "A",
		}
		for _, r := range testRecords {
			Frequency(i, &q, &r)
		}
		assert.Equal(t, Index{
			"2020-09-01": {"B": 1},
		}, i)
	})

	t.Run("Roll up to B with mutation", func(t *testing.T) {
		i = Index{}
		q = Query{
			Lineages: []QueryLineage{
				{Key: "B+B:B", PangoClade: "B.", Mutations: []Mutation{
					{Prefix: "B", Suffix: "B"},
				}},
				{Key: "B", PangoClade: "B."},
			},
		}
		for _, r := range testRecords {
			Frequency(i, &q, &r)
		}
		assert.Equal(t, Index{
			"2020-09-01": {"B": 1},
			"2020-10-01": {"B+B:B": 2},
			"2020-11-01": {"B+B:B": 3},
		}, i)
	})
}

func TestTotals(t *testing.T) {
	var i Index
	var q Query

	t.Run("Roll up to B", func(t *testing.T) {
		i = Index{}
		q = Query{
			Lineages: []QueryLineage{
				{Key: "B", PangoClade: "B."},
			},
		}
		foreach := func(agg func(r *Record), sliceNum int) {
			for _, r := range testRecords {
				agg(&r)
			}
		}
		Totals(foreach, &q, 0)
		assert.Equal(t, Index{
			"2020-09-01": {"A": 1},
			"2020-10-01": {"B": 2},
			"2020-11-01": {"C": 3},
		}, i)
	})
}

func TestSpatiotemporal(t *testing.T) {
	var i Index
	var q Query

	t.Run("Roll up to B", func(t *testing.T) {
		i = Index{}
		q = Query{
			Lineages: []QueryLineage{
				{Key: "B", PangoClade: "B."},
			},
		}
		for _, r := range testRecords {
			Spatiotemporal(i, &q, &r)
		}
		assert.Equal(t, Index{
			"2020-09-01": {"A": 1},
			"2020-10-01": {"B": 2},
			"2020-11-01": {"C": 3},
		}, i)
	})

	t.Run("Exclude sublineage", func(t *testing.T) {
		i = Index{}
		q = Query{
			Lineages: []QueryLineage{
				{Key: "B", PangoClade: "B."},
			},
			Excluding: []QueryLineage{
				{Key: "B.1.2", PangoClade: "B.1.2."},
			},
		}
		for _, r := range testRecords {
			Spatiotemporal(i, &q, &r)
		}
		assert.Equal(t, Index{
			"2020-09-01": {"A": 1},
			"2020-10-01": {"B": 2},
		}, i)
	})

	t.Run("Filter by mutation", func(t *testing.T) {
		i = Index{}
		q = Query{
			Lineages: []QueryLineage{
				{Key: "B.1", PangoClade: "B.1.", Mutations: []Mutation{
					{Prefix: "C", Suffix: "C"},
				}},
			},
		}
		for _, r := range testRecords {
			Spatiotemporal(i, &q, &r)
		}
		assert.Equal(t, Index{
			"2020-11-01": {"C": 3},
		}, i)
	})
}

func TestLineages(t *testing.T) {
	var m map[string]int
	var q Query

	t.Run("Unfiltered", func(t *testing.T) {
		m = map[string]int{}
		for _, r := range testRecords {
			Lineages(m, &q, &r)
		}
		assert.Equal(t, map[string]int{
			"B.":     1,
			"B.1.":   2,
			"B.1.2.": 3,
		}, m)
	})

	t.Run("Filter by area", func(t *testing.T) {
		m = map[string]int{}
		q = Query{Area: "A"}
		for _, r := range testRecords {
			Lineages(m, &q, &r)
		}
		assert.Equal(t, map[string]int{
			"B.": 1,
		}, m)
	})

	t.Run("Filter by dateFrom", func(t *testing.T) {
		m = map[string]int{}
		q = Query{DateFrom: "2020-10-01"}
		for _, r := range testRecords {
			Lineages(m, &q, &r)
		}
		assert.Equal(t, map[string]int{
			"B.1.":   2,
			"B.1.2.": 3,
		}, m)
	})

	t.Run("Filter by dateTo", func(t *testing.T) {
		m = map[string]int{}
		q = Query{DateTo: "2020-10-01"}
		for _, r := range testRecords {
			Lineages(m, &q, &r)
		}
		assert.Equal(t, map[string]int{
			"B.":   1,
			"B.1.": 2,
		}, m)
	})
}

func TestInfo(t *testing.T) {
	foreach := func(agg func(r *Record), sliceNum int) {
		for _, r := range testRecords {
			agg(&r)
		}
	}
	dates, areas := Info(foreach)
	assert.EqualValues(t, []string{"2020-09-01", "2020-10-01", "2020-11-01"}, dates)
	assert.EqualValues(t, []string{"A", "B", "C"}, areas)
}

func TestMutations(t *testing.T) {
	var m map[string]*MutationSearch
	var total MutationSearch
	var q Query
	so := SearchOpts{
		Lineage: "B",
		Growth:  GrowthOpts{Start: "2020-10-01", End: "2020-11-01"},
	}

	t.Run("A", func(t *testing.T) {
		m = map[string]*MutationSearch{}
		total = MutationSearch{}
		q = Query{
			Lineages:     []QueryLineage{{Key: "B", PangoClade: "B."}},
			Prefix:       "A",
			SuffixFilter: "A",
		}
		for _, r := range testRecords {
			Mutations(m, &total, &so, &q, &r)
		}
		assert.Equal(t, 6, m["A:A"].Count)
		assert.Equal(t, 6, total.Count)
	})

	t.Run("B", func(t *testing.T) {
		m = map[string]*MutationSearch{}
		total = MutationSearch{}
		q = Query{
			Lineages:     []QueryLineage{{Key: "B", PangoClade: "B."}},
			Prefix:       "B",
			SuffixFilter: "B"}
		for _, r := range testRecords {
			Mutations(m, &total, &so, &q, &r)
		}
		assert.Equal(t, 5, m["B:B"].Count)
		assert.Equal(t, 6, total.Count)
		assert.Equal(t, 2, m["B:B"].growthStart)
		assert.Equal(t, 3, m["B:B"].growthEnd)
	})

	t.Run("C", func(t *testing.T) {
		m = map[string]*MutationSearch{}
		total = MutationSearch{}
		q = Query{
			Lineages:     []QueryLineage{{Key: "B", PangoClade: "B."}},
			Prefix:       "C",
			SuffixFilter: "C"}
		for _, r := range testRecords {
			Mutations(m, &total, &so, &q, &r)
		}
		assert.Equal(t, 3, m["C:C"].Count)
		assert.Equal(t, 6, total.Count)
	})
}
