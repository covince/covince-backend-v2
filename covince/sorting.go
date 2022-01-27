package covince

import "strings"

type SortLineagesForQuery []QueryLineage

func (s SortLineagesForQuery) Len() int {
	return len(s)
}

func (s SortLineagesForQuery) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

const PANGO_SEPARATOR = "."

func (s SortLineagesForQuery) Less(i, j int) bool {
	a := s[i]
	b := s[j]
	depthA := strings.Count(a.PangoClade, PANGO_SEPARATOR)
	depthB := strings.Count(b.PangoClade, PANGO_SEPARATOR)
	if depthA == depthB {
		return len(a.Mutations) > len(b.Mutations)
	}
	return depthA > depthB
}
