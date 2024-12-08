package ttengine

import (
	"W365toFET/ttbase"
	"W365toFET/ttprint"
)

func PrintTT(ttinfo *ttbase.TtInfo, datadir string, name string) {
	for aix := 1; aix < len(ttinfo.Activities); aix++ {
		a := ttinfo.Activities[aix]
		l := a.Lesson
		p := a.Placement
		if p < 0 {
			//TODO?
		} else {
			l.Day = p / ttinfo.NHours
			l.Hour = p % ttinfo.NHours
			//TODO: Rooms will still need some thought ... how to handle
			// the choices, etc.
			// For the moment just incclude the compulsory rooms.
			l.Rooms = ttinfo.CourseInfo[l.Course].Room.Rooms
		}
	}
	plan_name := "Generated Plan"

	flags := map[string]bool{
		"WithTimes":  true,
		"WithBreaks": true,
	}
	ttprint.GenTypstData(ttinfo, datadir, name, plan_name, flags)

	typst := "typst"
	ttprint.MakePdf("print_timetable.typ", datadir, name+"_teachers", typst)
	ttprint.MakePdf("print_timetable.typ", datadir, name+"_classes", typst)
}
