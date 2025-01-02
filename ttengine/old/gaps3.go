package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
)

// Handle gaps (and TODO: lunch breaks)

//TODO: Under construction

// Testing a blanket approach initially – try to minimize gaps in students'
// timetables and ensure that all get a lunch break.
func findGapProblems3(ttinfo *ttbase.TtInfo, pmon *placementMonitor,
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
