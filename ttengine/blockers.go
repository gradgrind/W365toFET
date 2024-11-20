package ttengine

import (
	"W365toFET/base"
	"W365toFET/ttbase"
)

const BLOCKED_ACTIVITY = -1

func (tt *TtCore) addBlockers(
	ttinfo *ttbase.TtInfo,
	t2tt map[Ref]ResourceIndex,
	r2tt map[Ref]ResourceIndex,
) {
	db := ttinfo.Db
	for _, t := range db.Teachers {
		rix, ok := t2tt[t.Id]
		if ok {
			tt.blockResource(rix, t.NotAvailable)
		}
	}
	for _, r := range db.Rooms {
		rix, ok := r2tt[r.Id]
		if ok {
			tt.blockResource(rix, r.NotAvailable)
		}
	}
	for _, cl := range db.Classes {
		na := cl.NotAvailable
		if len(na) == 0 {
			continue
		}
		// Need to block all atomic groups. Their indexes are also their
		// Resource indexes.
		ags := ttinfo.AtomicGroups[cl.ClassGroup]
		for _, ag := range ags {
			tt.blockResource(ag.Index, na)
		}
	}
}

func (tt *TtCore) blockResource(
	rix ResourceIndex,
	timeslots []base.TimeSlot,
) {
	for _, ts := range timeslots {
		p := ts.Day*tt.NHours + ts.Hour
		tt.TtSlots[rix*tt.SlotsPerWeek+p] = BLOCKED_ACTIVITY
	}
}
