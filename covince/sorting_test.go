package covince

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mut = Mutation{Prefix: "B", Suffix: "B"}

func TestSorting(t *testing.T) {
	q := []QueryLineage{
		{Key: "B+B:B", PangoClade: "B.", Mutations: []Mutation{mut}},
		{Key: "B", PangoClade: "B."},
		{Key: "B.1", PangoClade: "B.1."},
	}

	sort.Sort(SortLineagesForQuery(q))

	assert.Equal(t, []QueryLineage{
		{Key: "B+B:B", PangoClade: "B.", Mutations: []Mutation{mut}},
		{Key: "B.1", PangoClade: "B.1."},
		{Key: "B", PangoClade: "B."},
	}, q)
}
