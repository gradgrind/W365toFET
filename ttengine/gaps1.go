package ttengine

import (
	"W365toFET/ttbase"
	"fmt"
	"math/rand/v2"
	"slices"
)

// Handle gaps and lunch breaks

// This one didn't really help much at all – no real progress.

// Testing a blanket approach initially – try to minimize gaps in students'
// timetables and ensure that all get a lunch break.
func findGapProblems1(ttinfo *ttbase.TtInfo, pmon *placementMonitor,
) []ttbase.ActivityIndex {
	ndays := ttinfo.NDays
	nhours := ttinfo.NHours
	nslots := ttinfo.SlotsPerWeek
	ttslots := ttinfo.TtSlots
	//lunchbreaks := ttinfo.Db.Info.MiddayBreak
	var unplaced []ttbase.ActivityIndex
	for _, cl := range ttinfo.Db.Classes {
		//TODO: This handles the lower classes first, which may be a good
		// idea, but later classes may never be reached!

		for _, ag := range ttinfo.AtomicGroups[cl.ClassGroup] {
			agix := ag.Index // Resource index
		sameag:
			gaps := []int{}
			aixlasts := []ttbase.ActivityIndex{}
			for d := 0; d < ndays; d++ {
				pending := []int{}
				aixlast := 0
				for h := 0; h < nhours; h++ {
					p := d*nhours + h
					aix := ttslots[agix*nslots+p]
					if aix == 0 {
						pending = append(pending, p)
					} else if aix < 0 {

						//TODO??

					} else {
						aixlast = aix
						if len(pending) != 0 {
							// Gaps are only gaps if an activity follows
							gaps = append(gaps, pending...)
							pending = []int{}
						}
					}
				}
				if aixlast != 0 {
					aixlasts = append(aixlasts, aixlast)
				}
			}
			ng := len(gaps)
			//fmt.Printf("??? %s: %d\n", ag.Tag, ng)
			if ng != 0 {
				n := 0
				if ng != 1 {
					n = rand.IntN(ng)
				}
				n0 := n
				for {
					slot := gaps[n]

					//TODO: Try to place one of the aixlasts here. Then return to
					// the main placement loop.
					for _, aix := range aixlasts {

						if pmon.check(aix) {
							// placed only recently
							continue
						}

						//TODO: Test for last-lesson-of-day constraint

						a := ttinfo.Activities[aix]
						if slices.Contains(a.PossibleSlots, slot) {
							ttinfo.UnplaceActivity(aix)
							unplaced = ttinfo.FindClashes(aix, slot)
							for _, aixx := range unplaced {
								ttinfo.UnplaceActivity(aixx)
							}
							//TODO: add removed stuff to pending ... how to do this?
							// As return value?
							ttinfo.PlaceActivity(aix, slot)
							pmon.added[aix] = pmon.count
							pmon.count++
							fmt.Printf("Gap Fill: %s aix=%d slot=%d gaps=%d\n",
								ag.Tag, aix, slot, ng-1)
							if len(unplaced) == 0 {
								if ng == 1 {
									// no more gaps here
									goto nextag
								}
								goto sameag
							}
							return unplaced
						}
					}
					//TODO: Special case if last-lesson-of-day activity?

					//TODO: try next slot before skipping this ag?
					n++
					if n == ng {
						n = 0
					}
					if n == n0 {
						fmt.Printf("Group %s -- Replacements too new: %d\n",
							ag.Tag, ng)
						break
					}
				}
			}
		nextag:
		}
	}
	return unplaced
}
