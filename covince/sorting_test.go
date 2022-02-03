package covince

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSorting(t *testing.T) {
	q := []QueryLineage{
		{Key: "B+B:B", PangoClade: "B.", Mutations: []QueryMutation{{Gene: "B", Description: "B"}}},
		{Key: "B", PangoClade: "B."},
		{Key: "B.1", PangoClade: "B.1."},
	}

	sort.Sort(SortLineagesForQuery(q))

	assert.Equal(t, []QueryLineage{
		{Key: "B.1", PangoClade: "B.1."},
		{Key: "B+B:B", PangoClade: "B.", Mutations: []QueryMutation{{Gene: "B", Description: "B"}}},
		{Key: "B", PangoClade: "B."},
	}, q)
}
