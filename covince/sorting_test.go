package covince

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSorting(t *testing.T) {
	q := []QueryLineage{
		{Key: "B+B", PangoClade: "B.", Mutations: []string{"B"}},
		{Key: "B", PangoClade: "B."},
		{Key: "B.1", PangoClade: "B.1."},
	}

	sort.Sort(SortLineagesForQuery(q))

	assert.Equal(t, []QueryLineage{
		{Key: "B.1", PangoClade: "B.1."},
		{Key: "B+B", PangoClade: "B.", Mutations: []string{"B"}},
		{Key: "B", PangoClade: "B."},
	}, q)
}
