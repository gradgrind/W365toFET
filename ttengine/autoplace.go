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
	fmt.Printf("Remaining: %d\n", l0)
	var pending []ttbase.ActivityIndex
	added := make(map[ttbase.ActivityIndex]int64, len(ttinfo.Activities))
	var delta int64 = 8 // This might be a reasonable value?
	var count int64 = delta
	for {
		l := len(failed) - 1
		if l < 0 {
			if len(pending) == 0 {
				fmt.Printf("========= DONE (%d)\n", count-delta)
				return
			} else {
				failed = pending
				l = len(failed) - 1
				//fmt.Printf("Remaining: %d\n", l+1)
				if l < l0 {
					fmt.Printf(" *!!* Remaining: %d (%d)\n", l0, count-delta)
					l0 = l
				}
				pending = nil
			}
		}

		aix := failed[l]
		failed = failed[:l]
		//fmt.Printf("!!! Failed at %d: %d\n", l, aix)

		// Get possible slots
		poss := possibleSlots(ttinfo, aix)

		//fmt.Printf("   *** Slots for %d: \n  -- %+v\n", aix, poss)

		// Seek least destructive placement
		ncmin := 1000
		type slotClashes struct {
			slot    int
			clashes []ttbase.ActivityIndex
		}
		var sclist []slotClashes
		for _, slot := range poss {
			clashes := ttinfo.FindClashes(aix, slot)
			for _, clash := range clashes {
				if count-added[clash] < delta {
					goto skip
				}
			}
			if len(clashes) < ncmin {
				ncmin = len(clashes)
				sclist = []slotClashes{{slot, clashes}}
			} else if len(clashes) == ncmin {
				sclist = append(sclist, slotClashes{slot, clashes})
			}
		skip:
		}
		scn := len(sclist)
		var slot int
		var clashes []ttbase.ActivityIndex
		//fmt.Printf("   *** Clashes for %d: %d (%d)\n", aix, ncmin, scn)
		if scn == 0 {
			n := len(poss)
			if n == 0 {
				fmt.Printf("!!!!! Couldn't place %d (%d)\n", aix, count-delta)
				return
			}
			if n == 1 {
				slot = poss[0]
			} else {
				slot = poss[rand.IntN(n)]
			}
			clashes = ttinfo.FindClashes(aix, slot)
		} else {
			i := 0
			if scn > 1 {
				i = rand.IntN(scn)
			}
			sc := sclist[i]
			slot = sc.slot
			clashes = sc.clashes
		}

		//for _, aixx := range ttinfo.FindClashes(aix, slot) {
		for _, aixx := range clashes {
			//fmt.Printf("   --- Remove %d\n", aixx)
			ttinfo.UnplaceActivity(aixx)
			pending = append(pending, aixx)
		}

		//TODO--- Just testing
		if !ttinfo.TestPlacement(aix, slot) {
			base.Bug.Fatalf("Clashes removed but still failed: %d\n", aix)
		}
		ttinfo.PlaceActivity(aix, slot)
		added[aix] = count
		count++
		if count == 1000000 {
			// Show unplaced lessons
			for _, aix := range append(failed, pending...) {
				a := ttinfo.Activities[aix]
				cinfo := a.CourseInfo
				reslist := []string{}
				for _, res := range a.Resources {
					r0 := ttinfo.Resources[res]
					{
						r, ok := r0.(*base.Teacher)
						if ok {
							reslist = append(reslist, r.Tag)
							continue
						}
					}
					{
						r, ok := r0.(*base.Room)
						if ok {
							reslist = append(reslist, r.Tag)
							continue
						}
					}
					{
						r, ok := r0.(*ttbase.AtomicGroup)
						if ok {
							reslist = append(reslist, r.Tag)
							continue
						}
					}
					reslist = append(reslist, fmt.Sprintf("??%+v", r0))
				}
				fmt.Printf("\n$$$ %s", ttinfo.View(cinfo))
				for _, res := range reslist {
					fmt.Printf("  -- %s\n", res)
				}
				fmt.Printf("  ++ %+v\n", a)
			}
			break
		}
	}
}

func possibleSlots(
	ttinfo *ttbase.TtInfo,
	aix ttbase.ActivityIndex,
) []int {
	// Get possible slots for the given activity
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
	return poss
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
