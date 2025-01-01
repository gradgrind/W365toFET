package ttengine

import (
	"W365toFET/ttbase"
	"W365toFET/ttprint"
)

func PrintTT(ttinfo *ttbase.TtInfo, datadir string, name string) {
	for aix := 1; aix < len(ttinfo.Activities); aix++ {
		a := ttinfo.Activities[aix]
		l := a.Lesson
		if l == nil {
			continue
		}
		p := a.Placement
		if p < 0 {
			//TODO?
		} else {
			l.Day = p / ttinfo.NHours
			l.Hour = p % ttinfo.NHours
			//TODO: Rooms will still need some thought ... how to handle
			// the choices, etc.
			// For the moment just include the compulsory rooms.
			l.Rooms = ttinfo.CourseInfo[l.Course].Room.Rooms
		}
	}

	// Generate Typst data
	ttprint.GenTimetables(ttinfo, datadir, name, nil, "typst")
}
