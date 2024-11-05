package w365tt

import "log"

func (dbp *DbTopLevel) readLessons() {
	for i := 0; i < len(dbp.Lessons); i++ {
		n := &dbp.Lessons[i]
		// The course must be Course or Supercourse.
		c, ok := dbp.Elements[n.Course]
		if !ok {
			log.Fatalf(
				"*ERROR* Lesson %s:\n  Unknown course: %s\n",
				n.Id, n.Course)
		}
		_, ok = c.(*Course)
		if !ok {
			_, ok = c.(*SuperCourse)
			if !ok {
				log.Fatalf(
					"*ERROR* Lesson %s:\n  Not a SuperCourse: %s\n",
					n.Id, n.Course)
			}
		}
		// Check the Rooms.
		reflist := []Ref{}
		for _, rref := range n.Rooms {
			r, ok := dbp.Elements[rref]
			if ok {
				_, ok = r.(*Room)
				if ok {
					reflist = append(reflist, rref)
					continue
				}
			}
			log.Printf(
				"*ERROR* Invalid Room in Lesson %s:\n  %s\n",
				n.Id, rref)
		}
		n.Rooms = reflist

	}
}
