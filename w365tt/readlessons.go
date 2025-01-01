package w365tt

import (
	"W365toFET/base"
)

func (db *DbTopLevel) readLessons(newdb *base.DbTopLevel) {
	for _, e := range db.Lessons {
		// The course must be Course or Supercourse.
		_, ok := db.CourseMap[e.Course]
		if !ok {
			base.Error.Fatalf(
				"Lesson %s:\n  Invalid course: %s\n",
				e.Id, e.Course)
		}
		// Check the Rooms.
		reflist := []base.Ref{}
		for _, rref := range e.Rooms {
			_, ok := db.RealRooms[rref]
			if ok {
				reflist = append(reflist, rref)
			} else {
				base.Error.Printf(
					"Invalid Room in Lesson %s:\n  %s\n",
					e.Id, rref)
			}
		}
		n := newdb.NewLesson(e.Id)
		n.Course = e.Course
		n.Duration = e.Duration
		n.Day = e.Day
		n.Hour = e.Hour
		n.Fixed = e.Fixed
		n.Rooms = reflist
		n.Flags = e.Flags
		n.Background = e.Background
		n.Footnote = e.Footnote
	}
}
