package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"fmt"
)

type TtClash struct {
	Course1  *timetable.CourseInfo
	Course2  *timetable.CourseInfo
	Slot     base.TimeSlot
	Resource base.Resource
}

// Test placement of fixed activities.
func test_fixed(ttdata *timetable.TtData) []*TtClash {
	shared_data := ttdata.SharedData

	// Determine resource index areas (atomic groups, teachers, rooms)
	atomicGroupResourceIndex0 := 0
	teacherResourceIndex0 := len(shared_data.AtomicNodes)
	roomResourceIndex0 := teacherResourceIndex0 + len(shared_data.Db.Teachers)
	nResources := roomResourceIndex0 + len(shared_data.Db.Rooms)

	// `resourceWeeks` contains the allocations of the "resources" (atomic
	// groups, teachers, rooms) to activities (indexes). This is organized
	// as an array of "week-chunks" (`HoursPerWeek` entries), one for each
	// resource (atomic groups, teachers, rooms).
	hpw := shared_data.HoursPerWeek
	resourceWeeks := make([]timetable.ActivityIndex, nResources*hpw)

	// Enter resources' blocked slots
	nhours := shared_data.NHours
	rbase := teacherResourceIndex0
	for ix, blocks := range ttdata.TeacherNotAvailable {
		rix := (ix + rbase) * hpw
		for d, blist := range blocks {
			for h, b := range blist {
				if b {
					resourceWeeks[rix+d*nhours+h] = -1
				}
			}
		}
	}
	rbase = roomResourceIndex0
	for ix, blocks := range ttdata.RoomNotAvailable {
		rix := (ix + rbase) * hpw
		for d, blist := range blocks {
			for h, b := range blist {
				if b {
					resourceWeeks[rix+d*nhours+h] = -1
				}
			}
		}
	}
	rbase = atomicGroupResourceIndex0
	classes := shared_data.Db.Classes
	for ix, blocks := range ttdata.ClassNotAvailable {

		cg := classes[ix].ClassGroup
		aglist := shared_data.AtomicGroups[cg]
		fmt.Printf("ATOMIC GROUPS for %s: %v\n", classes[ix].Tag, aglist)
		fmt.Printf(" ::: %v\n", blocks)
		for _, ag := range aglist {
			rix := (int(ag) + rbase) * hpw
			for d, blist := range blocks {
				for h, b := range blist {
					if b {
						resourceWeeks[rix+d*nhours+h] = -1
					}
				}
			}
		}
	}

	// Test for clashes and place if possible.
	allclashes := []*TtClash{}
	for aix, a := range shared_data.Activities {
		if aix == 0 || !a.Fixed {
			continue
		}
		clashes := []*TtClash{}
		timeslot := shared_data.ToTimeSlot(a.Placement)
		slot := int(a.Placement)
		cinfo := a.CourseInfo
		resources := []int{}
		for _, ix := range cinfo.Teachers {
			rix := (int(ix)+teacherResourceIndex0)*hpw + slot
			resources = append(resources, rix)
			if resourceWeeks[rix] != 0 {
				var c0 *timetable.CourseInfo = nil
				if resourceWeeks[rix] > 0 {
					c0 = shared_data.Activities[resourceWeeks[rix]].CourseInfo
				}
				clashes = append(clashes, &TtClash{
					Course1:  c0,
					Course2:  cinfo,
					Slot:     timeslot,
					Resource: shared_data.Db.Teachers[ix],
				})
			}
		}
		for _, ix := range cinfo.AtomicGroups {
			rix := (int(ix)+atomicGroupResourceIndex0)*hpw + slot
			resources = append(resources, rix)
			if resourceWeeks[rix] != 0 {
				a0 := shared_data.Activities[resourceWeeks[rix]]
				clashes = append(clashes, &TtClash{
					Course1:  a0.CourseInfo,
					Course2:  cinfo,
					Slot:     timeslot,
					Resource: shared_data.AtomicNodes[ix],
				})
			}
		}
		if !ttdata.WITHOUT_ROOM_PLACEMENTS {
			// Note that only fixed rooms are tested.
			for _, ix := range cinfo.FixedRooms {
				rix := (int(ix)+roomResourceIndex0)*hpw + slot
				resources = append(resources, rix)
				if resourceWeeks[rix] != 0 {
					a0 := shared_data.Activities[resourceWeeks[rix]]
					clashes = append(clashes, &TtClash{
						Course1:  a0.CourseInfo,
						Course2:  cinfo,
						Slot:     timeslot,
						Resource: shared_data.Db.Rooms[ix],
					})
				}
			}
		}
		if len(clashes) == 0 {
			for _, rix := range resources {
				resourceWeeks[rix] = timetable.ActivityIndex(aix)
			}
		} else {
			allclashes = append(allclashes, clashes...)
		}
	}
	return allclashes
}
