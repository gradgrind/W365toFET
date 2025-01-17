package ttengine

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"cmp"
	"slices"
)

type Penalty int64

type placementMonitor struct {
	stateStack            []*ttState
	ttinfo                *ttbase.TtInfo
	activityPlacementList []int
	unplaced              []ttbase.ActivityIndex
	constraintData        []any // resource index -> constraint data
	resourcePenalties     []Penalty
	score                 Penalty // current total penalty
}

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

// For testing!
func (pmon *placementMonitor) fullIntegrityCheck() {
	ttinfo := pmon.ttinfo
	rmap := make([]bool, len(ttinfo.TtSlots))
	nacts := len(ttinfo.Activities)
	unplaced := map[ttbase.ActivityIndex]bool{}
	for aix := 1; aix < nacts; aix++ {
		a := ttinfo.Activities[aix]
		p := a.Placement
		if p < 0 {
			unplaced[aix] = true
		} else {
			for _, r := range a.Resources {
				rix := r*ttinfo.SlotsPerWeek + p
				for i := 0; i < a.Duration; i++ {
					if ttinfo.TtSlots[rix+i] != aix {
						base.Bug.Fatalf("$!$ Resource: %d Slot: %d --> %d\n"+
							"  -- Activity: %+v\n",
							r, p, ttinfo.TtSlots[rix+i], a)
					}
					rmap[rix+i] = true
				}
			}
		}
	}
	// Check unplaced activities.
	if len(unplaced) != len(pmon.unplaced) {
		base.Bug.Fatalf("$!$ pmon.unplaced (%d) != actually unplaced (%d) \n",
			len(pmon.unplaced), len(unplaced))
	}
	for _, aix := range pmon.unplaced {
		if !unplaced[aix] {
			base.Bug.Fatalf("$!$ Unplaced activity (%d) is actually placed\n",
				aix)
		}
	}
	// Check the unchecked resource allocations.
	for rix, rp := range rmap {
		if !rp {
			aix := ttinfo.TtSlots[rix]
			if aix > 0 {
				base.Bug.Fatalf("$!$ Resource: %d Slot: %d -->"+
					" Activity: %+v\n",
					rix/ttinfo.SlotsPerWeek,
					rix%ttinfo.SlotsPerWeek,
					ttinfo.Activities[aix],
				)
			}
		}
	}
	// Check penalties: 1) Resource penalties
	for rix, rpen := range pmon.resourcePenalties {
		if rpen != pmon.resourcePenalty1(rix) {
			base.Bug.Fatalf("$!$ Resource (%d) has wrong penalty\n", rix)
		}
	}
	//TODO: further penalties
}

type activityPlacement struct {
	placement int
	//fixed bool
	xrooms []ttbase.ResourceIndex
}

type ttState struct {
	placements        []activityPlacement
	unplaced          []ttbase.ActivityIndex
	added             []int64
	count             int64
	score             Penalty
	ttslots           []ttbase.ActivityIndex
	resourcePenalties []Penalty

	lives int
}

func (pmon *placementMonitor) saveState() *ttState {
	//TODO. Currently this is probably saving more than would strictly be
	// necessary. This may be more time-efficient, though?
	ttinfo := pmon.ttinfo
	alist := pmon.ttinfo.Activities
	state := &ttState{
		placements:        make([]activityPlacement, len(alist)),
		unplaced:          make([]ttbase.ActivityIndex, len(pmon.unplaced)),
		ttslots:           make([]ttbase.ActivityIndex, len(ttinfo.TtSlots)),
		resourcePenalties: make([]Penalty, len(pmon.resourcePenalties)),
	}
	for aix := 1; aix < len(alist); aix++ {
		a := alist[aix]
		ap := activityPlacement{
			placement: a.Placement,
			//fixed: a.Fixed,
			xrooms: a.XRooms,
		}
		state.placements[aix] = ap
	}
	copy(state.unplaced, pmon.unplaced)
	state.score = pmon.score
	copy(state.ttslots, ttinfo.TtSlots)
	copy(state.resourcePenalties, pmon.resourcePenalties)
	return state
}

func (pmon *placementMonitor) restoreState(state *ttState) {
	// Restore the pmon-state from the argument.
	// This assumes the length of the activities list is fixed. If new
	// activities are added, or some removed, appropriate changes would
	// need to be made.
	alist := pmon.ttinfo.Activities
	// Integrity check
	if len(alist) != len(state.placements) {
		base.Bug.Fatalln("State restoration: number of activities changed")
	}
	for aix := 1; aix < len(alist); aix++ {
		a := alist[aix]
		ap := state.placements[aix]
		a.Placement = ap.placement
		//a.Fixed = ap.fixed
		a.XRooms = ap.xrooms
	}
	pmon.unplaced = pmon.unplaced[:0]
	pmon.unplaced = append(pmon.unplaced, state.unplaced...)
	pmon.score = state.score

	// Set the resource allocation and penalties
	copy(pmon.ttinfo.TtSlots, state.ttslots)
	copy(pmon.resourcePenalties, state.resourcePenalties)
}

func (pmon *placementMonitor) initConstraintData() {
	ttinfo := pmon.ttinfo
	db := ttinfo.Db

	cdatamap := make([]any, len(ttinfo.Resources))
	for _, c := range db.Classes {
		cgref := c.ClassGroup
		for _, ag := range ttinfo.AtomicGroups[cgref] {
			agix := ag.Index
			cdatamap[agix] = &AGConstraintData{
				lunchbreak:     c.LunchBreak,
				maxdaylessons:  c.MaxLessonsPerDay,
				maxdaygaps:     c.MaxGapsPerDay,
				maxweekgaps:    c.MaxGapsPerWeek,
				maxpm:          c.MaxAfternoons,
				forcefirsthour: c.ForceFirstHour,
			}
			//fmt.Printf(" ++ %s: %+v\n", ag.Tag, cdatamap[agix])
		}
	}
	pmon.constraintData = cdatamap
}
