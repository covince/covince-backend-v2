package covince

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThreads(t *testing.T) {
	so := SearchOpts{
		Lineage:       "B",
		Skip:          0,
		Limit:         20,
		SortProperty:  "count",
		SortDirection: "desc",
	}

	q := Query{
		Lineages: []QueryLineage{
			{Key: "B", PangoClade: "B."},
		},
	}

	chunkSize := 1
	foreach := func(agg func(r *Record), i int) {
		if i == -1 {
			for _, r := range testRecords {
				agg(&r)
			}
		} else {
			min := i * chunkSize
			max := min + chunkSize
			for _, r := range testRecords[min:max] {
				agg(&r)
			}
		}
	}

	sr := SearchMutations(foreach, &q, &so)

	so.Threads = len(testRecords)
	sr2 := SearchMutations(foreach, &q, &so)

	assert.Equal(t, sr, sr2)
}
