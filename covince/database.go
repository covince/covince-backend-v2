package covince

import (
	"strings"
)

type Value struct {
	Value string
}

type Record struct {
	// Metadata *Metadata
	Date *Value
	// Lineage    string
	PangoClade *Value
	Area       *Value
	Mutations  []*Mutation
	Count      int
}

type Mutation struct {
	Key    string
	Prefix string
	Suffix string
}

type Database struct {
	Count          int
	Genes          map[string]bool
	Mutations      []Mutation
	MutationLookup map[string]int
	Records        []Record
	Values         []Value
	ValueLookup    map[string]int
}

func CreateDatabase() *Database {
	return &Database{
		Genes:          make(map[string]bool),
		MutationLookup: make(map[string]int),
		ValueLookup:    make(map[string]int),
	}
}

func (db *Database) IndexMutations(muts []string, separator string) []*Mutation {
	ptrs := make([]*Mutation, len(muts))
	for i, m := range muts {
		var j int
		var ok bool
		if j, ok = db.MutationLookup[m]; !ok {
			j = len(db.Mutations)
			db.MutationLookup[m] = j

			split := strings.Split(m, separator)
			prefix := split[0]

			if _, ok = db.Genes[prefix]; !ok {
				db.Genes[prefix] = true
			}

			db.Mutations = append(
				db.Mutations,
				Mutation{
					Key:    m,
					Prefix: prefix,
					Suffix: split[1],
				},
			)
		}
		ptrs[i] = &db.Mutations[j]
	}
	return ptrs
}

func (db *Database) IndexValue(s string) *Value {
	var i int
	var ok bool
	if i, ok = db.ValueLookup[s]; !ok {
		i = len(db.Values)
		db.ValueLookup[s] = i
		db.Values = append(db.Values, Value{Value: s})
	}
	return &db.Values[i]
}
