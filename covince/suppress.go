package covince

import (
	"strings"
)

const MUT_SEPARATOR = "+"

func SuppressMutations(i Index, min int) Index {
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

func Suppress(i Index, min int) Index {
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
	return i
}
