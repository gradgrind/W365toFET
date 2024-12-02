package ttengine

import (
	"W365toFET/ttbase"
	"cmp"
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
	}
	return alist
}
