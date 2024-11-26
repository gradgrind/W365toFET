package ttengine

import (
	"W365toFET/base"
	"slices"
)

// By trying all slots for all (non-fixed) activities just after the fixed
// activities have been placed, each activity can get a list of potentially
// available slots.
//TODO: Also hard different-days constraints can be taken into account.

//TODO: What about parallel courses/ lessons?

//TODO: Should the fixed-lesson placement check for end-of-day problems
// when duration > 1?

func (tt *TtCore) makePossibleSlots() {
	for aix := 1; aix < len(tt.Activities); aix++ {
		a := tt.Activities[aix]
		if a.Fixed {
			// Fixed activities don't need possible slots.
			continue
		}
		plist := []int{}

		banned := []int{}
		ddnew := []ActivityIndex{}
		for _, addix := range a.DifferentDays {
			add := tt.Activities[addix]
			if !add.Fixed {
				// For later tests I only need these ones, as the
				// fixed ones will have already been taken care of.
				ddnew = append(ddnew, addix)
				continue
			}
			// Find day
			p := add.Placement
			if p < 0 {
				//TODO: get course?
				base.Bug.Fatalf("Unplaced fixed Activity: %d\n", addix)
			}
			dx := p / tt.NHours
			if !slices.Contains(banned, dx) {
				banned = append(banned, dx)
			}
		}
		a.DifferentDays = ddnew

		if len(banned) == tt.NDays {
			//TODO: get course?
			base.Error.Fatalf("Activity %d has no available time slots\n",
				aix)
		}
		length := a.Duration
		for d := 0; d < tt.NDays; d++ {
			// Only days without different days activities are possible.
			if slices.Contains(banned, d) {
				continue
			}
			for h := 0; h <= tt.NHours-length; h++ {
				p := d*tt.NHours + h
				if tt.testPlacement(aix, p) {
					plist = append(plist, p)
				}
			}
		}
		a.PossibleSlots = plist
	}
}
