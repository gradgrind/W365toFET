package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
)

// Handle gaps (and TODO: lunch breaks)

//TODO: Under construction

func (pmon *placementMonitor) getGroupGaps() (ttbase.ResourceIndex, []int) {
	// Return an atomic-group-resource-index and list of gaps (timeslots),
	// one of which should be filled.
	ttinfo := pmon.ttinfo
	ndays := ttinfo.NDays
	nhours := ttinfo.NHours
	nslots := ttinfo.SlotsPerWeek
	ttslots := ttinfo.TtSlots

	// All gaps are collected for a single atomic group
	dgaps := make([]int, nhours) // collect gaps in one day
	wgaps := make([]int, nhours) // accumulate dgaps over a week
	// This one accumulates gaps in days which are not over the limit:
	xgaps := make([]int, nhours)
	for _, cl := range ttinfo.Db.Classes {

		//TODO: Take lunch-breaks into account

		//TODO: Take max afternoons into account

		maxdaygaps := cl.MaxGapsPerDay
		maxweekgaps := cl.MaxGapsPerWeek

		for _, ag := range ttinfo.AtomicGroups[cl.ClassGroup] {
			agix := ag.Index // Resource index
			agbase := agix * nslots

			wgaps := wgaps[:0]
			xgaps := xgaps[:0]
			for d := 0; d < ndays; d++ {
				dgaps := dgaps[:0]
				ngaps := 0
				//ipending := 0
				for h := 0; h < nhours; h++ {
					p := d*nhours + h
					aix := ttslots[agbase+p]
					if aix == 0 {
						dgaps = append(dgaps, p)
						//npending++
					} else if aix < 0 {

						//TODO??

					} else {
						ngaps = len(dgaps)
					}
				} // end hour loop
				if maxdaygaps >= 0 && ngaps > maxdaygaps {
					wgaps = append(wgaps, dgaps...)
					//fmt.Printf("???1 %d, %d: %+v\n", ngaps, maxdaygaps, dgaps)
				} else {
					xgaps = append(xgaps, dgaps...)
				}
			} // end day loop
			if len(wgaps) != 0 {
				//fmt.Printf("???2 (%d) %+v\n", maxweekgaps, wgaps)
				return agix, wgaps
			}
			if maxweekgaps >= 0 && len(xgaps) > maxweekgaps {
				return agix, xgaps
			}
		} // end ag loop
	} // end class loop
	return -1, nil
}

func (pmon *placementMonitor) findFiller(
	agix ttbase.ResourceIndex,
	gap int,
) ttbase.ActivityIndex {
	// Find an activity for the given atomic group to put in the gap
	ttinfo := pmon.ttinfo
	// Get possible activities for this ag & slot
	alist := pmon.resourceSlotActivityMap[agix][gap]
	plist := []int{}
	aplist := []ttbase.ActivityIndex{}
	for _, aix := range alist {
		p := ttinfo.Activities[aix].Placement
		if p < 0 {
			return aix
		} else {
			if pmon.check(aix) {
				continue
			}
			plist = append(plist, p)
			aplist = append(aplist, aix)
		}
	}
	//fmt.Println("????????????? +++++")
	if len(plist) == 0 {
		// currently no possible activities for this ag and slot
		return 0
	}
	// All activities are placed
	aix := 0
	if len(alist) == 1 {
		aix = alist[0]
	} else {
		aix = aplist[chooseWeightedSlot(pmon.preferLater, plist)]
	}
	//fmt.Printf("????????????? ::::: %d\n", aixp)
	ttinfo.UnplaceActivity(aix)
	return aix
}

/*
func (pmon *placementMonitor) minimizeGaps() {
	//TODO: Possibly use a stack of states, up to a particular limit?

	// For now, I could use the agix as a measure of the quality? Coupled
	// with the number of gaps – but there can be a jump when all days are
	// satisfied, but not the week, or when a subsequent day has more gaps.

	//TODO: Some break-out mechanism may be required?

	states := []ttState{}

	for {
		states = append(states, pmon.saveState())
	repeat:
		agix, gaps := pmon.getGroupGaps()
		if agix < 0 {
			return
		}
		fmt.Printf("§GAPS: %d :: %+v\n", agix, gaps)

		//TODO:
		if improved {

			// Save state and repeat
			continue

		} else {

			// update "best", if necessary

			// try another replacement until no other possibilities, then FAIL?

			// FAIL:
			n := len(states) - 1

			state := state[n]
			pmon.restoreState(state0)

		}

		//TODO: choose activity, place it, resolve unplaced activities
	}
}
*/

// Testing a blanket approach initially – try to minimize gaps in students'
// timetables and ensure that all get a lunch break.
func findGapProblems(ttinfo *ttbase.TtInfo, pmon *placementMonitor,
) []ttbase.ActivityIndex {
	ndays := ttinfo.NDays
	nhours := ttinfo.NHours
	nslots := ttinfo.SlotsPerWeek
	ttslots := ttinfo.TtSlots
	//lunchbreaks := ttinfo.Db.Info.MiddayBreak
	var unplaced []ttbase.ActivityIndex
	gapmap := map[string]int{}
	for _, cl := range ttinfo.Db.Classes {
		//
		//TODO: Go through all classes and – if the gap constraints are not
		// met – place some activity in one of the gaps.
		//
		ng := 0
		for _, ag := range ttinfo.AtomicGroups[cl.ClassGroup] {
			agix := ag.Index // Resource index
			//sameag:
			gaps := []int{}
			//aixlasts := []ttbase.ActivityIndex{}
			for d := 0; d < ndays; d++ {
				pending := []int{}
				//aixlast := 0
				for h := 0; h < nhours; h++ {
					p := d*nhours + h
					aix := ttslots[agix*nslots+p]
					if aix == 0 {
						pending = append(pending, p)
					} else if aix < 0 {

						//TODO??

					} else {
						//aixlast = aix
						if len(pending) != 0 {
							// Gaps are only gaps if an activity follows
							gaps = append(gaps, pending...)
							pending = []int{}
						}
					}
				} // end hour loop
				//if aixlast != 0 {
				//	aixlasts = append(aixlasts, aixlast)
				//}
			} // end day loop
			ng = len(gaps)
			//fmt.Printf("??? %s: %d\n", ag.Tag, ng)
			if ng == 0 {
				continue // -> next ag (or class)
			}
			var gap int
			if ng == 1 {
				gap = gaps[0]
			} else {
				gap = gaps[chooseWeightedSlot(pmon.preferEarlier, gaps)]
			}

			// Get possible activities for this ag & slot
			alist := pmon.resourceSlotActivityMap[agix][gap]
			if len(alist) == 0 {
				// no possible activities for this ag and slot
				continue
			}

			plist := []int{}
			aixp := 0
			for _, aix := range alist {
				p := ttinfo.Activities[aix].Placement
				if p < 0 {

					//TODO: Break and use this one immediately?
					aixp = aix
					break

				} else {
					plist = append(plist, p)
				}

			}

			//fmt.Println("????????????? +++++")
			if aixp == 0 {
				// All activities are placed
				if len(alist) == 1 {
					aixp = alist[0]
				} else {
					aixp = alist[chooseWeightedSlot(pmon.preferLater, plist)]
				}
				//fmt.Printf("????????????? ::::: %d\n", aixp)
				ttinfo.UnplaceActivity(aixp)
			}
			toremove := ttinfo.FindClashes(aixp, gap)
			for _, aixx := range toremove {
				ttinfo.UnplaceActivity(aixx)
			}
			unplaced = append(unplaced, toremove...)
			ttinfo.PlaceActivity(aixp, gap)
			tmp := []ttbase.ActivityIndex{}
			for _, aixt := range unplaced {
				if aixt != aixp {
					tmp = append(tmp, aixt)
				}
			}
			unplaced = tmp
			pmon.added[aixp] = pmon.count
			pmon.count++
			//fmt.Println("????????????? -----")
			break // -> nextclass
		} // end ag loop
		gapmap[cl.Tag] = ng
	} // end class loop
	fmt.Printf("&&&&& GAPS: %+v\n", gapmap)
	//fmt.Println("????????????? /////")
	return unplaced
}
