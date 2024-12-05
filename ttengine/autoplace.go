package ttengine

import (
	"W365toFET/ttbase"
	"cmp"
	"fmt"
	"math/rand/v2"
	"slices"
)

func collectCourseLessons(ttinfo *ttbase.TtInfo) []ttbase.ActivityIndex {
	type wlix struct {
		lix ttbase.ActivityIndex
		w   float32
	}
	toplace := []wlix{}
	for _, cinfo := range ttinfo.CourseInfo {
		var w float32 = -1.0
		for _, lix := range cinfo.Lessons {
			a := ttinfo.Activities[lix]
			if a.Placement < 0 {
				if w < 0.0 {
					w = float32(len(a.PossibleSlots)) / float32(len(cinfo.Lessons))
				}
				toplace = append(toplace, wlix{lix, w})
			}
		}
	}
	slices.SortStableFunc(toplace, func(a, b wlix) int {
		return cmp.Compare(a.w, b.w)
	})
	alist := make([]ttbase.ActivityIndex, len(toplace))
	for i, wl := range toplace {
		alist[i] = wl.lix
		//fmt.Printf("??? %+v\n", wl)
	}

	for i, aix := range alist {
		if !tryToPlace(ttinfo, aix) {
			fmt.Printf("!!! Failed at %d: %d\n", i, aix)
		}

	}

	return alist
}

func tryToPlace(ttinfo *ttbase.TtInfo, aix ttbase.ActivityIndex) bool {
	a := ttinfo.Activities[aix]
	n := len(a.PossibleSlots)
	// Pick one at random
	i0 := rand.IntN(n)
	// Test all possible slots, starting at this index, until a free one
	// is found.
	i := i0
	for {
		if ttinfo.TestPlacement(aix, a.PossibleSlots[i]) {
			ttinfo.PlaceActivity(aix, a.PossibleSlots[i])
			return true
		}
		i++
		if i == n {
			i = 0
		}
		if i == i0 {
			break
		}
	}
	return false
}
