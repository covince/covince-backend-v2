package covince

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testRecords = []Record{
	{Lineage: "B", PangoClade: "B.", Date: "2020-09-01", Area: "A", Count: 1, Mutations: "|A|"},
	{Lineage: "B.1", PangoClade: "B.1.", Date: "2020-10-01", Area: "B", Count: 2, Mutations: "|A|B|"},
	{Lineage: "B.1.2", PangoClade: "B.1.2.", Date: "2020-11-01", Area: "C", Count: 3, Mutations: "|A|B|C|"},
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
				{Key: "B+B", PangoClade: "B.", Mutations: []string{"B"}},
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

}

func TestSpatiotemporal(t *testing.T) {

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
			"B":     1,
			"B.1":   2,
			"B.1.2": 3,
		}, m)
	})
}
