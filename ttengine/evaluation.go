package ttengine

import (
	"W365toFET/ttbase"
)

func (pmon *placementMonitor) resourcePenalty1(r ttbase.ResourceIndex) int {
	rc := pmon.constraintData[r]
	if rc != nil {
		rcag, ok := rc.(*AGConstraintData)
		if ok {
			return pmon.countAGHours(r, rcag)
		}
		//TODO: other resources
	}
	return 0
}

func (pmon *placementMonitor) evaluate1(aix ttbase.ActivityIndex) int {
	// A full evaluation could be quite expensive. In may cases only a
	// small number of things will have changed. The question is, which
	// ones?

	// Would a resource-based score be workable? If so, I could just
	// reevaluate the affected resources, based on the changed activities.

	ttinfo := pmon.ttinfo
	a := ttinfo.Activities[aix]
	penalty := 0
	pmon.pendingPenalties = pmon.pendingPenalties[:0]
	for _, r := range a.Resources {
		rp := pmon.resourcePenalty1(r)
		pmon.pendingPenalties = append(pmon.pendingPenalties,
			resourcePenalty{r, rp})
		penalty += rp - pmon.resourcePenalties[r]
	}

	// Count student gaps
	// + MaxGapsPerDay (int)
	// + MaxGapsPerWeek (int)

	// Student lunch breaks
	// + LunchBreak (bool)

	// Student afternoons, etc.
	// + MaxAfternoons (int)
	// + MaxLessonsPerDay (int)
	// - MinLessonsPerDay (int) ... probably not at stage 1
	// + ForceFirstHour (bool)

	// Count teacher gaps?
	// - MaxGapsPerDay (int)
	// - MaxGapsPerWeek (int)

	// Teacher lunch breaks?
	// - LunchBreak (bool)

	// Teacher afternoons, etc.?
	// - MaxDays (int)
	// - MaxAfternoons (int)
	// - MaxLessonsPerDay (int)
	// - MinLessonsPerDay (int) ... probably not at stage 1

	// Activities DaysBetween?

	// Activities LastLesson?

	return penalty
}

type AGConstraintData struct {
	lunchbreak     bool
	maxdaylessons  int
	maxdaygaps     int
	maxweekgaps    int
	maxpm          int
	forcefirsthour bool
}

func (pmon *placementMonitor) countAGHours(
	agix ttbase.ResourceIndex,
	cdata *AGConstraintData,
) int {
	ttinfo := pmon.ttinfo
	maxdaylessons := cdata.maxdaylessons
	maxdaygaps := cdata.maxdaygaps
	maxweekgaps := cdata.maxweekgaps
	maxpm := cdata.maxpm
	withlunch := cdata.lunchbreak
	firsthour := cdata.forcefirsthour
	WEIGHT_AG_MaxLessonsPerDay := 0
	WEIGHT_AG_MaxGapsPerDay := 0
	WEIGHT_AG_MaxGapsPerWeek := 0
	WEIGHT_AG_MaxAfternoons := 0
	WEIGHT_AG_NoLunchBreak := 1
	WEIGHT_AG_ForceFirstHour := 1
	if maxdaylessons >= 0 {
		WEIGHT_AG_MaxLessonsPerDay = 1
	}
	if maxdaygaps >= 0 {
		WEIGHT_AG_MaxGapsPerDay = 1
	}
	if maxweekgaps >= 0 {
		WEIGHT_AG_MaxGapsPerWeek = 1
	}
	if maxpm >= 0 {
		WEIGHT_AG_MaxAfternoons = 1
	}

	//TODO: get lunch hours, lunch-break flag
	// lunch := pmon.lunchTimes
	lunch := []bool{false, false, false, false,
		true, true, true,
		false, false, false,
	}
	//TODO: get afternoon start hour
	//pmstart := ttinfo.AfternoonHour0
	pmstart := 6

	slot := agix * ttinfo.SlotsPerWeek
	ndx := 0 // daily gaps in excess of the permitted number
	nw := 0
	nlb := 0  // count missing lunch breaks
	npm := 0  // count active afternoons
	nlx := 0  // excess daily lessons
	nffh := 0 // count broken first hours
	for d := 0; d < ttinfo.NDays; d++ {
		ndi := 0     // gaps on this day
		lb := 0      // number of lunch-break slots
		pending := 0 // count only gaps with a following activity
		lasth := 0   // last active hour of day
		nl := 0      // count lessons

		if firsthour && ttinfo.TtSlots[slot] == 0 {
			nffh++
		}

		for h := 0; h < ttinfo.NHours; h++ {
			aix := ttinfo.TtSlots[slot]
			slot++
			if aix <= 0 {
				if lunch[h] {
					lb++
				}
				if aix == 0 {
					pending++
				}
				continue
			}
			// aix > 0
			nl++
			lasth = h
			ndi += pending
			pending = 0
		}
		if lasth >= pmstart {
			npm++
			// Afternoon lesson(s) on this day
			if withlunch {
				if lb == 0 {
					nlb++
				} else {
					// One lunch break slot doesn't count as a gap.
					//TODO: This should probably not be done if one of the
					// lunch-break slots is blocked (aix == -1)!
					ndi--
				}
			}
		}
		if ndi > maxdaygaps {
			ndx += ndi - maxdaygaps
		}
		if nl > maxdaylessons {
			nlx += nl - maxdaylessons
		}
		nw += ndi // accumulate week gaps
	}
	penalty := ndx * WEIGHT_AG_MaxGapsPerDay
	if nw > maxweekgaps {
		penalty += (nw - maxweekgaps) * WEIGHT_AG_MaxGapsPerWeek
	}
	if npm > maxpm {
		penalty += (npm - maxpm) * WEIGHT_AG_MaxAfternoons
	}
	if nlb != 0 {
		penalty += nlb * WEIGHT_AG_NoLunchBreak
	}
	if nlx != 0 {
		penalty += nlx * WEIGHT_AG_MaxLessonsPerDay
	}
	if nffh != 0 {
		penalty += nffh * WEIGHT_AG_ForceFirstHour
	}
	return penalty
}
