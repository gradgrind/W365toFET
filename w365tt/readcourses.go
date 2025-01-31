package w365tt

import (
	"W365toFET/base"
	"strings"
)

func (db *DbTopLevel) readCourses(newdb *base.DbTopLevel) {
	db.CourseMap = map[Ref]bool{}
	for _, e := range db.Courses {
		subject := db.getCourseSubject(newdb, e.Subjects, e.Id)
		room := db.getCourseRoom(newdb, e.PreferredRooms, e.Id)
		groups := db.getCourseGroups(e.Groups, e.Id)
		teachers := db.getCourseTeachers(e.Teachers, e.Id)
		n := newdb.NewCourse(e.Id)
		n.Subject = subject
		n.Groups = groups
		n.Teachers = teachers
		n.Room = room
		db.CourseMap[e.Id] = true
	}
}

func (db *DbTopLevel) readSuperCourses(newdb *base.DbTopLevel) {
	// In the input from W365 the subjects for the SuperCourses must be
	// taken from the linked EpochPlan.
	// The EpochPlans are otherwise not needed.
	epochPlanSubjects := map[Ref]base.Ref{}
	if db.EpochPlans != nil {
		for _, n := range db.EpochPlans {
			sref, ok := db.SubjectTags[n.Tag]
			if !ok {
				sref = db.makeNewSubject(newdb, n.Tag, n.Name)
			}
			epochPlanSubjects[n.Id] = sref
		}
	}

	sbcMap := map[Ref]*base.SubCourse{}
	for _, spc := range db.SuperCourses {
		// Read the SubCourses.
		for _, e := range spc.SubCourses {
			sbc, ok := sbcMap[e.Id]
			if ok {
				// Assume the SubCourse really is the same.
				sbc.SuperCourses = append(sbc.SuperCourses, spc.Id)
			} else {
				subject := db.getCourseSubject(newdb, e.Subjects, e.Id)
				room := db.getCourseRoom(newdb, e.PreferredRooms, e.Id)
				groups := db.getCourseGroups(e.Groups, e.Id)
				teachers := db.getCourseTeachers(e.Teachers, e.Id)
				// Use a new Id for the SubCourse because it can also be
				// the Id of a Course.
				n := newdb.NewSubCourse("$$" + e.Id)
				n.SuperCourses = []base.Ref{spc.Id}
				n.Subject = subject
				n.Groups = groups
				n.Teachers = teachers
				n.Room = room
				sbcMap[e.Id] = n
			}
		}

		// Now add the SuperCourse.
		subject, ok := epochPlanSubjects[spc.EpochPlan]
		if !ok {
			base.Error.Fatalf("Unknown EpochPlan in SuperCourse %s:\n  %s\n",
				spc.Id, spc.EpochPlan)
		}
		n := newdb.NewSuperCourse(spc.Id)
		n.Subject = subject
		db.CourseMap[n.Id] = true
	}
}

func (db *DbTopLevel) getCourseSubject(
	newdb *base.DbTopLevel,
	srefs []base.Ref,
	courseId base.Ref,
) base.Ref {
	//
	// Deal with the Subjects field of a Course or SubCourse – W365
	// allows multiple subjects.
	// The base db expects one and only one subject (in the Subject field).
	// If there are multiple subjects in the input, these will be converted
	// to a single "composite" subject, using all the subject tags.
	// Repeated use of the same subject list will reuse the created subject.
	//
	msg := "Course %s:\n  Not a Subject: %s\n"
	var subject Ref
	if len(srefs) == 1 {
		wsid := srefs[0]
		_, ok := db.SubjectMap[wsid]
		if !ok {
			base.Error.Fatalf(msg, courseId, wsid)
		}
		subject = wsid
	} else if len(srefs) > 1 {
		// Make a subject name
		sklist := []string{}
		for _, wsid := range srefs {
			// Need Tag/Shortcut field
			stag, ok := db.SubjectMap[wsid]
			if ok {
				sklist = append(sklist, stag)
			} else {
				base.Error.Fatalf(msg, courseId, wsid)
			}
		}
		sktag := strings.Join(sklist, "/")
		wsid, ok := db.SubjectTags[sktag]
		if ok {
			// The name has already been used.
			subject = wsid
		} else {
			// Need a new Subject.
			subject = db.makeNewSubject(newdb, sktag, "Compound Subject")
		}
	} else {
		base.Error.Fatalf("Course/SubCourse has no subject: %s\n", courseId)
	}
	return subject
}

func (db *DbTopLevel) getCourseRoom(
	newdb *base.DbTopLevel,
	rrefs []base.Ref,
	courseId base.Ref,
) base.Ref {
	//
	// Deal with rooms. W365 can have a single RoomGroup or a list of Rooms.
	// If there is a list of Rooms, this is converted to a RoomChoiceGroup.
	// In the end there should be a single Room, RoomChoiceGroup or RoomGroup
	// in the "Room" field.
	// If a list of rooms recurs, the same RoomChoiceGroup is used.
	//
	room := base.Ref("")
	if len(rrefs) > 1 {
		// Make a RoomChoiceGroup
		var estr string
		room, estr = db.makeRoomChoiceGroup(newdb, rrefs)
		if estr != "" {
			base.Error.Printf("In Course %s:\n%s", courseId, estr)
		}
	} else if len(rrefs) == 1 {
		// Check that room is Room or RoomGroup.
		rref0 := rrefs[0]
		_, ok := db.RealRooms[rref0]
		if ok {
			room = rref0
		} else {
			if db.RoomGroupMap[rref0] {
				room = rref0
			} else {
				base.Error.Printf("Invalid room in Course/SubCourse %s:\n  %s\n",
					courseId, rref0)
			}
		}
	}
	return room
}

func (db *DbTopLevel) getCourseGroups(
	grefs []Ref,
	courseId base.Ref,
) []base.Ref {
	//
	// Check the group references and replace Class references by the
	// corresponding whole-class base.Group references.
	//
	glist := []base.Ref{}
	for _, gref := range grefs {
		ngref, ok := db.GroupRefMap[gref]
		if !ok {
			base.Error.Fatalf("Invalid group in Course/SubCourse %s:\n  %s\n",
				courseId, gref)
		}
		glist = append(glist, ngref)
	}
	return glist
}

func (db *DbTopLevel) getCourseTeachers(
	trefs []Ref,
	courseId base.Ref,
) []base.Ref {
	//
	// Check the teacher references.
	//
	tlist := []base.Ref{}
	for _, tref := range trefs {
		_, ok := db.TeacherMap[tref]
		if !ok {
			base.Error.Fatalf("Unknown teacher in Course %s:\n  %s\n",
				courseId, tref)
		}
		tlist = append(tlist, tref)
	}
	return tlist
}
