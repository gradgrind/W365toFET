package ttengine

import (
	"W365toFET/ttbase"
)

// TODO: Probably need to remember where the real activities end. Though these
// dummies would have some special feature (CourseInfo == nil?, Lesson == nil?)
// that would enable them to be identified easily ...
func DummyActivity(ttinfo *ttbase.TtInfo) *ttbase.Activity {
	// Index of new Activity:
	ttlix := len(ttinfo.Activities)
	ttl := &ttbase.Activity{
		Index:    ttlix,
		Duration: 1,
		//Lesson:     nil,
		//CourseInfo: nil,
		//Resources: to be set,
		//Fixed: true, // for blocking lesson
		//Placement: to be set
	}
	ttinfo.Activities = append(ttinfo.Activities, ttl)
	return ttl
}

// Handle gaps and lunch breaks

//TODO: This doesn't work at all. My guess is that individual atomic groups are
// getting too many blockers. Probably only whole classes should get blockers ...

// Testing a blanket approach initially – try to minimize gaps in students'
// timetables and ensure that all get a lunch break.
func findGapProblems(ttinfo *ttbase.TtInfo, pmon *placementMonitor,
) []ttbase.ActivityIndex {
	ndays := ttinfo.NDays
	nhours := ttinfo.NHours
	nslots := ttinfo.SlotsPerWeek
	ttslots := ttinfo.TtSlots
	var unplaced []ttbase.ActivityIndex
	for _, cl := range ttinfo.Db.Classes {
		// Skip invalid classes
		//TODO: How and where are invalid classes detected?
		if cl.Tag == "" {
			continue
		}

		//TODO: This handles the lower classes first, which may be a good
		// idea, but later classes may never be reached!

		aglist := ttinfo.AtomicGroups[cl.ClassGroup]
		aggaps := make([]int, len(aglist))
		for agn, ag := range aglist {
			aggaps[agn] = 0
			agix := ag.Index // Resource index
			for d := 0; d < ndays; d++ {
				//aixlast := 0
				ngaps := 0
				npending := 0
				//hlast := -1
				slot0 := agix*nslots + d*nhours
				for h := 0; h < nhours; h++ {
					aix := ttslots[slot0+h]
					if aix == 0 {
						npending++
					} else if aix < 0 {

						//TODO??

					} else if pmon.added[aix] < 0 {

						//TODO: handle blockers??

					} else {
						//aixlast = aix
						//hlast = h
						//if len(pending) != 0 {
						if npending != 0 {
							// Gaps are only gaps if an activity follows
							//gaps = append(gaps, pending...)
							ngaps += npending
							npending = 0
							//pending = pending[:0]
						}
					}
				} // end of hour loop
				aggaps[agn] += ngaps

				//TODO: Do I need the last filled cells, the last aix, ...

			} // end of day loop

		} // end of ag loop
		/*
			// Check whether a lunch break would be necessary and one
			// is not present.
			if hlast-ngaps >= ttinfo.PMStart {
				// Add a lunch-break, if there isn't one already
				for _, h := range ttinfo.LunchTimes {
					aixl := ttslots[slot0+h]
					if aixl < 0 {
						goto nolb
					}
					if aixl == 0 {
						continue
					}
					if ttinfo.Activities[aixl].Lesson == nil {
						// There is already a lunch-break
						goto nolb
					}
				}
				// Add a lunch break
				{
					a := DummyActivity(ttinfo)
					pmon.added = append(pmon.added, 0)
					a.Resources = []ttbase.ResourceIndex{agix}
					a.Placement = -1
					lbslots := make([]int, len(ttinfo.LunchTimes))
					for i, h := range ttinfo.LunchTimes {
						lbslots[i] = d*nhours + h
					}
					a.PossibleSlots = lbslots
					unplaced = append(unplaced, a.Index)
					goto nextclass
				}
			nolb:

				if ngaps == 0 {
					continue // Go to next day
				}

				if a.Fixed || pmon.check(aixlast) {
					// fixed or placed only recently
					continue // Go to next day
				}

				ttinfo.UnplaceActivity(aixlast)
				unplaced = append(unplaced, aixlast)

				////// TODO: Not like this ...
				// Add blockers, but don't encroach on minHours limit.
				for hlast >= cl.MinLessonsPerDay {
					b := DummyActivity(ttinfo)
					b.Fixed = true
					b.Resources = []ttbase.ResourceIndex{agix}
					ttinfo.PlaceActivity(b.Index, d*nhours+hlast)
					pmon.added = append(pmon.added, -1)
					ngaps--
					if ngaps == 0 {
						break
					}
					hlast--
					if ttslots[slot0+hlast] != 0 {
						// continue only if there is a free lesson
						break
					}
				}
				//////
				goto nextclass
			}
		*/
		//nextclass:
	} // end of class loop

	//TODO: When a class completes, check its gap constraints. If not fulfilled,
	// increment count and try again.

	// An empty "unplaced" doesn't necessarily mean that there are no gaps or
	// that no further improvement is possible ...

	return unplaced
}
