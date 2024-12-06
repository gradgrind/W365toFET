package ttengine

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"cmp"
	"fmt"
	"math/rand/v2"
	"slices"
)

func CollectCourseLessons(ttinfo *ttbase.TtInfo) []ttbase.ActivityIndex {
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
	return alist
}

//TODO: IMPORTANT! Chack that I am handling (hard) parallel lessons correctly.
// If possible, only one of a parallel set should be in the list of activities
// still to be placed. But that might be a bit tricky. This probably needs
// some thought ...

func PlaceLessons(ttinfo *ttbase.TtInfo, alist []ttbase.ActivityIndex) {
	failed := []ttbase.ActivityIndex{}
	for _, aix := range alist {
		if !tryToPlace(ttinfo, aix) {
			failed = append(failed, aix)
		}
	}
	slices.Reverse(failed)
	l0 := len(failed)
	var pending []ttbase.ActivityIndex
	for {
		l := len(failed) - 1
		if l < 0 {
			if len(pending) == 0 {
				break
			} else {
				failed = pending
				l = len(failed) - 1
				if l < l0 {
					fmt.Printf("Remaining: %d\n", l0)
					l0 = l
				}
				pending = nil
			}
		}

		/*
			if l < l0 {
				fmt.Printf("Remaining: %d\n", l)
				l0 = l
			}
		*/

		//fmt.Printf("??? %d: %+v\n", l+1, failed[l-5:])

		aix := failed[l]
		failed = failed[:l]
		//fmt.Printf("!!! Failed at %d: %d\n", l, aix)

		// Get possible slots
		a := ttinfo.Activities[aix]
		poss := []int{}
		for _, slot := range a.PossibleSlots {
			day := slot / ttinfo.NHours
			for _, addix := range a.DifferentDays {
				add := ttinfo.Activities[addix]
				if add.Placement >= 0 && add.Placement/ttinfo.NHours == day {
					//TODO: Maybe it's OK if the course is different?
					// Could try accepting these later, if there are
					// otherwise no possible slots?
					goto fail
				}
			}
			poss = append(poss, slot)
		fail:
		}

		//fmt.Printf("   *** Slots for %d: \n  -- %+v\n", aix, poss)

		//fmt.Printf("%s\n", ttinfo.View(a.CourseInfo))

		slot := poss[rand.IntN(len(poss))]

		//fmt.Printf(" ? Slot: %d\n", slot)

		for _, aixx := range ttinfo.FindClashes(aix, slot) {
			//fmt.Printf("   --- Remove %d\n", aixx)
			ttinfo.UnplaceActivity(aixx)
			pending = append(pending, aixx)
		}

		//TODO--- Just testing
		if !ttinfo.TestPlacement(aix, slot) {
			base.Bug.Fatalf("Clashes removed but still failed: %d\n", aix)
		}
		ttinfo.PlaceActivity(aix, slot)
	}
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
