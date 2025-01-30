package ttbase

import (
	"W365toFET/base"
)

const BLOCKED_ACTIVITY = -1

func (ttinfo *TtInfo) addBlockers() {
	t2tt := ttinfo.TeacherIndexes
	r2tt := ttinfo.RoomIndexes
	db := ttinfo.Db
	for _, t := range db.Teachers {
		rix, ok := t2tt[t.Id]
		if ok {
			ttinfo.blockResource(rix, t.NotAvailable)
		}
	}
	for _, r := range db.Rooms {
		rix, ok := r2tt[r.Id]
		if ok {
			ttinfo.blockResource(rix, r.NotAvailable)
		}
	}
	for _, cl := range db.Classes {
		na := cl.NotAvailable
		if len(na) == 0 {
			continue
		}
		// Need to block all atomic groups. Their indexes are also their
		// Resource indexes.
		for _, agix := range ttinfo.AtomicGroupIndexes[cl.ClassGroup] {
			ttinfo.blockResource(agix, na)
		}
	}
}

// blockPadding sets the time-slots which are only there as padding (the
// ones at the end of each day) to a blocking value.
func (ttinfo *TtInfo) blockPadding() {
	// Iterate over all resources
	for rix := 0; rix < len(ttinfo.Resources); rix++ {
		// ... and all days
		for d := 0; d < ttinfo.NDays; d++ {
			for h := ttinfo.NHours; h < ttinfo.DayLength; h++ {
				p := d*ttinfo.DayLength + h
				ttinfo.TtSlots[rix*ttinfo.SlotsPerWeek+p] = BLOCKED_ACTIVITY
			}
		}
	}
}

func (ttinfo *TtInfo) blockResource(
	rix ResourceIndex,
	timeslots []base.TimeSlot,
) {
	for _, ts := range timeslots {
		p := ts.Day*ttinfo.DayLength + ts.Hour
		ttinfo.TtSlots[rix*ttinfo.SlotsPerWeek+p] = BLOCKED_ACTIVITY
	}
}
