package covince

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testRecords = []Record{
	{Lineage: "B", PangoClade: "B.", Date: "2020-09-01", Area: "A", Count: 1, Mutations: "|A:A|"},
	{Lineage: "B.1", PangoClade: "B.1.", Date: "2020-10-01", Area: "B", Count: 2, Mutations: "|A:A|B:B|"},
	{Lineage: "B.1.2", PangoClade: "B.1.2.", Date: "2020-11-01", Area: "C", Count: 3, Mutations: "|A:A|B:B|C:C|"},
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
			Frequency(i, q, r)
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
			Frequency(i, q, r)
		}
		assert.Equal(t, Index{
			"2020-09-01": {"B": 1},
		}, i)
	})

	t.Run("Roll up to B with mutation", func(t *testing.T) {
		i = Index{}
		q = Query{
			Lineages: []QueryLineage{
				{Key: "B+B", PangoClade: "B.", Mutations: []string{"|B:B|"}},
				{Key: "B", PangoClade: "B."},
			},
		}
		for _, r := range testRecords {
			Frequency(i, q, r)
		}
		assert.Equal(t, Index{
			"2020-09-01": {"B": 1},
			"2020-10-01": {"B+B": 2},
			"2020-11-01": {"B+B": 3},
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
		for _, r := range testRecords {
			Totals(i, q, r)
		}
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
			Spatiotemporal(i, q, r)
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
			Spatiotemporal(i, q, r)
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
				{Key: "B.1", PangoClade: "B.1.", Mutations: []string{"|C:C|"}},
			},
		}
		for _, r := range testRecords {
			Spatiotemporal(i, q, r)
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
			Lineages(m, q, r)
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
			Lineages(m, q, r)
		}
		assert.Equal(t, map[string]int{
			"B.": 1,
		}, m)
	})

	t.Run("Filter by dateFrom", func(t *testing.T) {
		m = map[string]int{}
		q = Query{DateFrom: "2020-10-01"}
		for _, r := range testRecords {
			Lineages(m, q, r)
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
			Lineages(m, q, r)
		}
		assert.Equal(t, map[string]int{
			"B.":   1,
			"B.1.": 2,
		}, m)
	})
}

func TestInfo(t *testing.T) {
	foreach := func(agg func(r Record)) {
		for _, r := range testRecords {
			agg(r)
		}
	}
	dates, areas := Info(foreach)
	assert.EqualValues(t, []string{"2020-09-01", "2020-10-01", "2020-11-01"}, dates)
	assert.EqualValues(t, []string{"A", "B", "C"}, areas)
}

func TestMutations(t *testing.T) {
	var m map[string]int
	var q Query

	t.Run("A", func(t *testing.T) {
		m = map[string]int{}
		q = Query{Search: "|A:"}
		for _, r := range testRecords {
			Mutations(m, q, r)
		}
		assert.Equal(t, map[string]int{
			"A:A": 6,
		}, m)
	})

	t.Run("B", func(t *testing.T) {
		m = map[string]int{}
		q = Query{Search: "|B:"}
		for _, r := range testRecords {
			Mutations(m, q, r)
		}
		assert.Equal(t, map[string]int{
			"B:B": 5,
		}, m)
	})

	t.Run("C", func(t *testing.T) {
		m = map[string]int{}
		q = Query{Search: "|C:"}
		for _, r := range testRecords {
			Mutations(m, q, r)
		}
		assert.Equal(t, map[string]int{
			"C:C": 3,
		}, m)
	})
}
