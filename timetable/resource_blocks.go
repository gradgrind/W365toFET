package timetable

import "W365toFET/base"

const BLOCKED_ACTIVITY = -1

// `blockResource` blocks the given resource (index) in the given time slots.
func (tt_data *TtData) blockResource(
	rix ResourceIndex,
	timeslots []base.TimeSlot,
) {
	tt_shared_data := tt_data.SharedData
	for _, ts := range timeslots {
		p := ts.Day*tt_shared_data.NHours + ts.Hour
		tt_data.ResourceWeeks[rix*tt_shared_data.HoursPerWeek+p] =
			BLOCKED_ACTIVITY
	}
}

// `BlockResources` blocks the resource in the time slots specified in the
// NotAvailable fields of their nodes in the main data structure
// (base.DbTopLevel).
func (tt_data *TtData) BlockResources() {
	tt_shared_data := tt_data.SharedData
	db := tt_shared_data.Db
	for _, tnode := range db.Teachers {
		rix, ok := tt_shared_data.TeacherIndex[tnode.Id]
		if ok {
			tt_data.blockResource(rix, tnode.NotAvailable)
		} else {
			base.Bug.Fatalf("ResourceBlocking, unknown teacher: %s", tnode.Id)
		}
	}

	for _, rnode := range db.Rooms {
		rix, ok := tt_shared_data.RoomIndex[rnode.Id]
		if ok {
			tt_data.blockResource(rix, rnode.NotAvailable)
		} else {
			base.Bug.Fatalf("ResourceBlocking, unknown room: %s", rnode.Id)
		}
	}

	for _, cnode := range db.Classes {
		na := cnode.NotAvailable
		if len(na) == 0 {
			continue
		}
		for _, rix := range tt_shared_data.AtomicGroups[cnode.ClassGroup] {
			tt_data.blockResource(rix, na)
		}
	}
}
