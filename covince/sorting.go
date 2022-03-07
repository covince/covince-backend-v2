package covince

import "strings"

type SortLineagesForQuery []QueryLineage

func (s SortLineagesForQuery) Len() int      { return len(s) }
func (s SortLineagesForQuery) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

const PANGO_SEPARATOR = "."

func (s SortLineagesForQuery) Less(i, j int) bool {
	a := s[i]
	b := s[j]

	aMuts := len(a.Mutations)
	bMuts := len(b.Mutations)
	if aMuts > 0 && bMuts == 0 {
		return true
	}
	if bMuts > 0 && aMuts == 0 {
		return false
	}

	depthA := strings.Count(a.PangoClade, PANGO_SEPARATOR)
	depthB := strings.Count(b.PangoClade, PANGO_SEPARATOR)
	if depthA == depthB {
		return aMuts > bMuts
	}
	return depthA > depthB
}
