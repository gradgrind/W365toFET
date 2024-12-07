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
	added := map[ttbase.ActivityIndex]int{}
	count := 0
	for {
		l := len(failed) - 1
		if l < 0 {
			if len(pending) == 0 {
				return
			} else {

				/*-- TODO: Under certain circumstances pick a particular
				// activity to "unplace"? This seems ineffective!
				if rand.IntN(100) < 10 {
					aixmap := map[ttbase.ActivityIndex]int{}
					for _, aix := range pending {
						for _, slot := range possibleSlots(ttinfo, aix) {
							for _, aixx := range ttinfo.FindClashes(aix, slot) {
								aixmap[aixx]++
							}
						}
					}
					aix0 := 0
					nc := 0
					for aix, n := range aixmap {
						if n > nc {
							nc = n
							aix0 = aix
						}
					}
					ttinfo.UnplaceActivity(aix0)
					failed = []ttbase.ActivityIndex{aix0}
					failed = append(failed, pending...)
				} else {
					failed = pending
				}
				*/

				failed = pending

				//

				l = len(failed) - 1
				fmt.Printf("Remaining: %d\n", l+1)
				if l < l0 {
					fmt.Printf(" *!!* Remaining: %d\n", l0)
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
			if len(clashes) < ncmin {
				ncmin = len(clashes)
				sclist = []slotClashes{{slot, clashes}}
			} else if len(clashes) == ncmin {
				sclist = append(sclist, slotClashes{slot, clashes})
			}
		}
		scn := len(sclist)
		//fmt.Printf("   *** Clashes for %d: %d (%d)\n", aix, ncmin, scn)
		if scn == 0 {
			fmt.Printf("!!!!! Couldn't place %d\n", aix)
			return
		}

		i := 0
		if scn > 1 {
			i = rand.IntN(scn)
		}
		sc := sclist[i]

		/*
			//fmt.Printf("%s\n", ttinfo.View(a.CourseInfo))

			slotix0 := rand.IntN(len(poss))

			//fmt.Printf(" ? Slot: %d\n", slot0)

			slotix := slotix0
			var slot int
			for {
				slot = poss[slotix]
				for _, aixx := range ttinfo.FindClashes(aix, slot) {
					if added[aixx] > 0 {
						goto next
					}
				}
				break
			next:
				slotix++
				if slotix == len(poss) {
					slotix = 0
				}
				if slotix == slotix0 {
					slotix = -1
					break
				}
			}
			if slotix < 0 {
				//TODO
				fmt.Printf(" ??? Trouble placing %d: %d\n", aix, added[aix])
				//fmt.Printf("   --- Added %+v\n", added)
				//fmt.Printf("   --- Pending %+v\n", pending)
				//return
			}
		*/

		//for _, aixx := range ttinfo.FindClashes(aix, slot) {
		for _, aixx := range sc.clashes {
			//fmt.Printf("   --- Remove %d\n", aixx)
			ttinfo.UnplaceActivity(aixx)
			pending = append(pending, aixx)
			if len(pending) == 30 {
				//fmt.Printf("   --- Added %+v\n", added)
				//fmt.Printf("   --- Pending %+v\n", pending)
				//				return
			}
		}

		//TODO--- Just testing
		if !ttinfo.TestPlacement(aix, sc.slot) {
			base.Bug.Fatalf("Clashes removed but still failed: %d\n", aix)
		}
		ttinfo.PlaceActivity(aix, sc.slot)
		added[aix]++
		count++
		if count == 1000 {
			// Show unplaced lessons
			for _, aix := range append(failed, pending...) {
				a := ttinfo.Activities[aix]
				//cinfo := a.CourseInfo
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
						r, ok := r0.(*base.Group)
						if ok {
							reslist = append(reslist, r.Tag)
							continue
						}
					}
					reslist = append(reslist, fmt.Sprintf("??%+v", r0))
				}
				//fmt.Printf("$$$ %s (%s) %+v\n",
				//	ttinfo.View(cinfo), strings.Join(reslist, ","), a)
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
