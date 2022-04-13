package covince

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexMutations(t *testing.T) {
	db := CreateDatabase()
	db.IndexMutations([]string{"A:A"}, ":")
	db.IndexMutations([]string{"A:A", "B:B"}, ":")
	db.IndexMutations([]string{"A:A", "B:B", "C:C"}, ":")
	assert.Equal(t, testMutations, db.Mutations)
}
