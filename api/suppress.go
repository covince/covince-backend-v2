package api

import (
	"strings"

	"github.com/covince/covince-backend-v2/covince"
)

const MUT_SEPARATOR = "+"

func suppressMutations(i covince.Index, min int) covince.Index {
	for k, m := range i {
		for lineage, count := range m {
			if strings.Contains(lineage, MUT_SEPARATOR) && count < min {
				delete(m, lineage)
			}
		}
		if len(m) == 0 {
			delete(i, k)
		}
	}
	return i
}

func suppress(i covince.Index, min int, lineages []covince.QueryLineage) covince.Index {
	for _, l := range lineages {
		if len(l.Mutations) > 0 {
			for k, m := range i {
				for kk, count := range m {
					if count < min {
						delete(m, kk)
					}
				}
				if len(m) == 0 {
					delete(i, k)
				}
			}
			break
		}
	}
	return i
}
