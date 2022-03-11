package covince

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexMutations(t *testing.T) {
	db := CreateDatabase()
	db.IndexMutations([]string{"A:A"})
	db.IndexMutations([]string{"A:A", "B:B"})
	db.IndexMutations([]string{"A:A", "B:B", "C:C"})
	assert.Equal(t,
		[]Mutation{
			{Key: "A:A", Prefix: "A", Suffix: "A"},
			{Key: "B:B", Prefix: "B", Suffix: "B"},
			{Key: "C:C", Prefix: "C", Suffix: "C"},
		},
		db.Mutations,
	)
}
