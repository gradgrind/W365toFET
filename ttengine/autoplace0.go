package ttengine

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"cmp"
	"fmt"
	"math/rand/v2"
	"slices"
)

const MAX_STEPS = 1000000

func CollectCourseLessons(ttinfo *ttbase.TtInfo) []ttbase.ActivityIndex {
	// Collect the lessons that need placing and sort them according to a
	// measure of their "placeability", based on the number of slots into
	// which they could be placed.
	// The idea is that those lessons with fewer available slots should
	// probably be placed first. These should end up at the beginning of
	// the list.
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

func PlaceLessons0(ttinfo *ttbase.TtInfo, alist []ttbase.ActivityIndex) {
	preferEarlier := buildEarlyHourWeights(ttinfo.NDays, ttinfo.NHours, 4)
	preferLater := buildLateHourWeights(ttinfo.NDays, ttinfo.NHours, 5)
	resourceSlotActivityMap := makeResourceSlotActivityMap(ttinfo)

	var pmon placementMonitor
	{
		var delta int64 = 7 // This might be a reasonable value?
		pmon = placementMonitor{
			count:                   delta,
			delta:                   delta,
			added:                   make([]int64, len(ttinfo.Activities)),
			ttinfo:                  ttinfo,
			preferEarlier:           preferEarlier,
			preferLater:             preferLater,
			resourceSlotActivityMap: resourceSlotActivityMap,
		}
	}
	failed := []ttbase.ActivityIndex{}
	for _, aix := range alist {
		//
		//if !tryToPlace(ttinfo, aix) {
		//	failed = append(failed, aix)
		//}
		//
		if !placeFree(ttinfo, preferEarlier, aix) {
			failed = append(failed, aix)
		}
	}
	slices.Reverse(failed)
	l0 := len(failed)
	fmt.Printf("Remaining: %d\n", l0)

	for {
		//TODO: Add constraint handling ... esp. gaps and afternoons ...

		ok := pmon.placeActivities(failed)
		if !ok {
			fmt.Println("$$$$$$ FAILED $$$$$$")
			return
		}
		fmt.Printf("$$$$$$ OK (%d) $$$$$$\n", pmon.count)
		//

		agix, gaps := pmon.getGroupGaps()
		if agix < 0 {
			break
		}
		fmt.Printf(" --> %d: %+v\n", agix, gaps)

		// Compare result with best so far? How? I can't get rid of that
		// single gap!

		// Choose a gap
		ngaps := len(gaps)
		i0 := 0
		i := 0
		if ngaps > 1 {
			i = rand.IntN(ngaps)
		}
		for {
			gap := gaps[i]
			aix := pmon.findFiller(agix, gap)
			fmt.Printf("  *** Place %d -> %d\n", aix, gap)
			if aix == 0 {
				i++
				if i == ngaps {
					i = 0
				}
				if i == i0 {
					// No possible replacement activity
					//TODO ???
					base.Error.Fatalf("Can't fill gaps for %d\n", agix)
				}
				continue
			}
			clashes := ttinfo.FindClashes(aix, gap)
			for _, aixx := range clashes {
				ttinfo.UnplaceActivity(aixx)
			}
			ttinfo.PlaceActivity(aix, gap)
			pmon.added[aix] = pmon.count + 10
			pmon.count++
			failed = clashes
			break
		}

		// One replacement made

	}
}

func (pmon *placementMonitor) placeActivities(
	alist []ttbase.ActivityIndex) bool {
	// Try to place all the given activities.

	// Make a local copy of the activity list to leave the source untouched.
	pending := make([]ttbase.ActivityIndex, len(alist))
	copy(pending, alist)

	ttinfo := pmon.ttinfo
	// Collect activities displaced when placing others
	unplaced := []ttbase.ActivityIndex{}
	l0 := len(pending)
	for {
		l := len(pending)
		if l == 0 {
			if len(unplaced) == 0 {
				return true
			}
			pending = append(pending, unplaced...)
			unplaced = unplaced[:0]
			l = len(pending)
			if l < l0 {
				fmt.Printf(" *!!* Remaining: %d (%d)\n",
					l0, pmon.count-pmon.delta)
				l0 = l
			}
		}

		// Pop activity from end of list
		l--
		aix := pending[l]
		pending = pending[:l]
		//fmt.Printf("!!! Pending at %d: %d\n", l, aix)

		// Get possible slots
		poss := possibleSlots(ttinfo, aix)

		//fmt.Printf("   *** Slots for %d: \n  -- %+v\n", aix, poss)

		// Seek least destructive placement
		ncmin := 1000
		type slotClashes struct {
			slot    int
			clashes []ttbase.ActivityIndex
		}
		sclist := []slotClashes{}
	repeat:
		for _, slot := range poss {
			clashes := ttinfo.FindClashes(aix, slot)
			for _, clash := range clashes {

				if pmon.check(clash) {
					// fixed or count-added[clash] < delta
					// Don't consider this slot because a clashing activity
					// cannot or should not be removed.
					goto skip
				}
			}
			// Add to list for this number of clashing activities
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
				fmt.Printf("!!!!! Couldn't place %d (%d)\n  -- %+v\n",
					aix, pmon.count-pmon.delta, ttinfo.Activities[aix])
				return false
			}

			//TODO???
			pmon.count++
			goto repeat
			//

		} else {
			i := 0
			if scn > 1 {
				i = rand.IntN(scn)
			}
			sc := sclist[i]
			slot = sc.slot
			clashes = sc.clashes
		}

		integrityCheck(ttinfo)

		if len(clashes) != 0 {
			//fmt.Println("*********** REMOVE ***********")
		}
		for _, aixx := range clashes {
			//fmt.Printf("   --- Remove %d\n", aixx)
			ttinfo.UnplaceActivity(aixx)
			unplaced = append(unplaced, aixx)
		}

		//TODO--- Just for testing, but useful there!
		if !ttinfo.TestPlacement(aix, slot) {
			a := ttinfo.Activities[aix]
			fmt.Printf("!!! %d: %+v\n", slot, a)
			fmt.Printf("  ++ %s\n", ttinfo.View(a.CourseInfo))

			//--

			day := slot / ttinfo.NHours
			for _, addix := range a.DifferentDays {
				add := ttinfo.Activities[addix]
				if add.Placement >= 0 && add.Placement/ttinfo.NHours == day {
					fmt.Printf("  --dd %+v\n", add)
				}
			}
			for _, rix := range a.Resources {
				i := rix*ttinfo.SlotsPerWeek + slot
				for ix := 0; ix < a.Duration; ix++ {
					if ttinfo.TtSlots[i+ix] != 0 {
						fmt.Printf("  --res %d %d\n", rix, ttinfo.TtSlots[i+ix])
					}
				}
			}

			//

			base.Bug.Fatalf("Clashes removed but still failed: %d\n", aix)
		}
		ttinfo.PlaceActivity(aix, slot)
		pmon.added[aix] = pmon.count
		pmon.count++
		//if pmon.count >= 400 {
		if pmon.count >= MAX_STEPS {
			// Show unplaced lessons
			for _, aix := range append(unplaced, pending...) {
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
	return false
}

func (pmon *placementMonitor) placeActivitiesWeighted(
	alist []ttbase.ActivityIndex) bool {
	// Try to place all the given activities.

	// This version should favour earlier slots, but I have no evidence that
	// it really does ...

	// Make a local copy of the activity list to leave the source untouched.
	pending := make([]ttbase.ActivityIndex, len(alist))
	copy(pending, alist)

	ttinfo := pmon.ttinfo
	// Collect activities displaced when placing others
	unplaced := []ttbase.ActivityIndex{}
	l0 := len(pending)
	for {
		l := len(pending)
		if l == 0 {
			if len(unplaced) == 0 {
				return true
			}
			pending = append(pending, unplaced...)
			unplaced = unplaced[:0]
			l = len(pending)
			if l < l0 {
				fmt.Printf(" *!!* Remaining: %d (%d)\n",
					l0, pmon.count-pmon.delta)
				l0 = l
			}
		}

		// Pop activity from end of list
		l--
		aix := pending[l]
		pending = pending[:l]
		//fmt.Printf("!!! Pending at %d: %d\n", l, aix)

		// Get possible slots
		poss := possibleSlots(ttinfo, aix)

		//fmt.Printf("   *** Slots for %d: \n  -- %+v\n", aix, poss)

		// Seek least destructive placement
		ncmin := 1000
		type slotClashes struct {
			slot    int
			clashes []ttbase.ActivityIndex
		}
		sclist := []slotClashes{}
	repeat:
		for _, slot := range poss {
			clashes := ttinfo.FindClashes(aix, slot)
			for _, clash := range clashes {

				if pmon.check(clash) {
					// fixed or count-added[clash] < delta
					// Don't consider this slot because a clashing activity
					// cannot or should not be removed.
					goto skip
				}
			}
			// Add to list for this number of clashing activities
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
				fmt.Printf("!!!!! Couldn't place %d (%d)\n  -- %+v\n",
					aix, pmon.count-pmon.delta, ttinfo.Activities[aix])
				return false
			}

			//TODO???
			pmon.count++
			goto repeat
			//

		} else {
			//
			//i := 0
			//
			if scn > 1 {
				//
				//i = rand.IntN(scn)
				//
				slots := []int{}
				for _, sc := range sclist {
					slots = append(slots, sc.slot)
				}
				i := chooseWeightedSlot(pmon.preferEarlier, slots)
				slot = slots[i]
				clashes = sclist[i].clashes
			} else {
				sc := sclist[0]
				slot = sc.slot
				clashes = sc.clashes
			}
			//
			//sc := sclist[i]
			//slot = sc.slot
			//clashes = sc.clashes
			//
		}

		integrityCheck(ttinfo)

		//for _, aixx := range ttinfo.FindClashes(aix, slot) {
		if len(clashes) != 0 {
			//fmt.Println("*********** REMOVE ***********")
		}
		for _, aixx := range clashes {
			//fmt.Printf("   --- Remove %d\n", aixx)
			ttinfo.UnplaceActivity(aixx)
			unplaced = append(unplaced, aixx)
		}

		//TODO--- Just for testing, but useful there!
		if !ttinfo.TestPlacement(aix, slot) {
			a := ttinfo.Activities[aix]
			fmt.Printf("!!! %d: %+v\n", slot, a)
			fmt.Printf("  ++ %s\n", ttinfo.View(a.CourseInfo))

			//--

			day := slot / ttinfo.NHours
			for _, addix := range a.DifferentDays {
				add := ttinfo.Activities[addix]
				if add.Placement >= 0 && add.Placement/ttinfo.NHours == day {
					fmt.Printf("  --dd %+v\n", add)
				}
			}
			for _, rix := range a.Resources {
				i := rix*ttinfo.SlotsPerWeek + slot
				for ix := 0; ix < a.Duration; ix++ {
					if ttinfo.TtSlots[i+ix] != 0 {
						fmt.Printf("  --res %d %d\n", rix, ttinfo.TtSlots[i+ix])
					}
				}
			}

			//

			base.Bug.Fatalf("Clashes removed but still failed: %d\n", aix)
		}
		ttinfo.PlaceActivity(aix, slot)
		pmon.added[aix] = pmon.count
		pmon.count++
		//if pmon.count >= 400 {
		if pmon.count >= MAX_STEPS {
			// Show unplaced lessons
			for _, aix := range append(unplaced, pending...) {
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
	return false
}
