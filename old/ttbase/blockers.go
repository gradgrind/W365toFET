package ttbase

import (
	"W365toFET/base"
)

const BLOCKED_ACTIVITY = -1

func (ttinfo *TtInfo) addBlockers(
	t2tt map[Ref]ResourceIndex,
	r2tt map[Ref]ResourceIndex,
) {
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
		ags := ttinfo.AtomicGroups[cl.ClassGroup]
		for _, ag := range ags {
			ttinfo.blockResource(ag.Index, na)
		}
	}
}

func (ttinfo *TtInfo) blockResource(
	rix ResourceIndex,
	timeslots []base.TimeSlot,
) {
	for _, ts := range timeslots {
		p := ts.Day*ttinfo.NHours + ts.Hour
		ttinfo.TtSlots[rix*ttinfo.SlotsPerWeek+p] = BLOCKED_ACTIVITY
	}
}
