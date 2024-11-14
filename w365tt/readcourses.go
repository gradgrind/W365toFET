package w365tt

import (
	"W365toFET/base"
	"W365toFET/logging"
	"strings"
)

func (db *DbTopLevel) readSubjects(newdb *base.DbTopLevel) {
	db.SubjectMap = map[Ref]*base.Subject{}
	db.SubjectTags = map[string]Ref{}
	db.SubjectNames = map[string]string{}
	for _, e := range db.Subjects {
		// Perform some checks and add to the SubjectNames
		// and SubjectTags maps.
		_, nok := db.SubjectTags[e.Tag]
		if nok {
			logging.Error.Fatalf("Subject Tag (Shortcut) defined twice: %s\n",
				e.Tag)
		}
		t, nok := db.SubjectNames[e.Name]
		if nok {
			logging.Warning.Printf("Subject Name defined twice (different"+
				" Tag/Shortcut):\n  %s (%s/%s)\n", e.Name, t, e.Tag)
		} else {
			db.SubjectNames[e.Name] = e.Tag
		}
		db.SubjectTags[e.Tag] = e.Id
		//Copy data to base db.
		s := &base.Subject{
			Id:   e.Id,
			Tag:  e.Tag,
			Name: e.Name,
		}
		newdb.Subjects = append(newdb.Subjects, s)
		db.SubjectMap[e.Id] = s
	}
}

func (db *DbTopLevel) readCourses(newdb *base.DbTopLevel) {
	for _, e := range db.Courses {
		subject := db.getCourseSubject(newdb, e.Subjects, e.Id)
		room := db.getCourseRoom(newdb, e.PreferredRooms, e.Id)
		newdb.Courses = append(newdb.Courses, &base.Course{
			Id:       e.Id,
			Subject:  subject,
			Groups:   e.Groups,
			Teachers: e.Teachers,
			Room:     room,
		})
	}
	// TODO?
	// dbp.readCourse(n)
}

func (dbp *DbTopLevel) readSuperCourses() {
	// If there are SuperCourses without subjects, the subjects will be
	// taken from the EpochPlans, which are then no longer needed.
	epochPlanSubjects := map[Ref]Ref{}
	if dbp.EpochPlans != nil {
		for _, n := range dbp.EpochPlans {
			sref, ok := dbp.SubjectTags[n.Tag]
			if !ok {
				sref = dbp.makeNewSubject(n.Tag, n.Name)
			}
			epochPlanSubjects[n.Id] = sref
		}
	}
	dbp.EpochPlans = nil

	for i := 0; i < len(dbp.SuperCourses); i++ {
		n := dbp.SuperCourses[i]
		if n.Subject == "" {
			n.Subject = epochPlanSubjects[n.EpochPlan]
		}
	}
}

func (dbp *DbTopLevel) readSubCourses() {
	for i := 0; i < len(dbp.SubCourses); i++ {
		n := dbp.SubCourses[i]
		for _, spcref := range n.SuperCourses {
			s, ok := dbp.Elements[spcref]
			if !ok {
				logging.Error.Fatalf(
					"SubCourse %s:\n  Unknown SuperCourse: %s\n",
					n.Id, spcref)
			}
			_, ok = s.(*SuperCourse)
			if !ok {
				logging.Error.Fatalf(
					"SubCourse %s:\n  Not a SuperCourse: %s\n",
					n.Id, spcref)
			}
		}
		dbp.readCourse(n)
	}
}

func (db *DbTopLevel) getCourseSubject(
	newdb *base.DbTopLevel,
	srefs []base.Ref,
	courseId base.Ref,
) base.Ref {
	//
	// Deal with the Subjects field of a Course (or SuperCourse) – W365
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
			logging.Error.Fatalf(msg, courseId, wsid)
		}
		subject = wsid
	} else if len(srefs) > 1 {
		// Make a subject name
		sklist := []string{}
		for _, wsid := range srefs {
			// Need Tag/Shortcut field
			s, ok := db.SubjectMap[wsid]
			if ok {
				sklist = append(sklist, s.Tag)
			} else {
				logging.Error.Fatalf(msg, courseId, wsid)
			}
		}
		sktag := strings.Join(sklist, ",")
		wsid, ok := db.SubjectTags[sktag]
		if ok {
			// The Name has already been used.
			subject = wsid
		} else {
			// Need a new Subject.
			subject := db.NewId()
			sbj := &base.Subject{
				Id:   subject,
				Tag:  sktag,
				Name: "Compound Subject",
			}
			newdb.Subjects = append(newdb.Subjects, sbj)
		}
	} else {
		logging.Error.Fatalf("Course has no subject: %s\n", courseId)
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
	// in the "Room" field. The "PreferredRooms" field in cleared.
	// If a list of rooms recurs, the same RoomChoiceGroup is used.
	//
	room := base.Ref{""}
	if len(rrefs) > 1 {
		// Make a RoomChoiceGroup
		var estr string
		room, estr = db.makeRoomChoiceGroup(newdb, rrefs)
		if estr != "" {
			logging.Error.Printf("In Course %s:\n%s", courseId, estr)
		}
	} else if len(rrefs) == 1 {
		// Check that room is Room or RoomGroup.
		rref0 := rrefs[0]
		r, ok := db.RealRooms[rref0]
		if ok {
			room = rref0
		} else {
			//TODO: How to test for RoomGroup
			
			else {
					logging.Error.Printf("Invalid room in Course %s:\n  %s\n",
					course.GetId(), rref0)
			}
		}
		
	}

	return room
}

func (dbp *DbTopLevel) readCourse(course CourseInterface) {

	//
	// Deal with groups.
	//
	//glist := []Ref{}
	for _, gref := range course.GetGroups() {
		g, ok := dbp.Elements[gref]
		if !ok {
			logging.Error.Fatalf("Unknown group in Course %s:\n  %s\n",
				course.GetId(), gref)
			//continue
		}
		// g can be a Group or a Class.
		_, ok = g.(*Group)
		if !ok {
			// Check for class.
			_, ok = g.(*Class)
			if !ok {
				logging.Error.Fatalf("Invalid group in Course %s:\n  %s\n",
					course.GetId(), gref)
				//continue
			}
		}
		//glist = append(glist, gref)
	}

	//
	// Deal with teachers
	//
	//tlist := []Ref{}
	for _, tref := range course.GetTeachers() {
		t, ok := dbp.Elements[tref]
		if !ok {
			logging.Error.Fatalf("Unknown teacher in Course %s:\n  %s\n",
				course.GetId(), tref)
			//continue
		}
		_, ok = t.(*Teacher)
		if !ok {
			logging.Error.Fatalf("Invalid teacher in Course %s:\n  %s\n",
				course.GetId(), tref)
			//continue
		}
		//tlist = append(tlist, tref)
	}

	//
	// Deal with rooms. W365 can have a single RoomGroup or a list of Rooms.
	// If there is a list of Rooms, this is converted to a RoomChoiceGroup.
	// In the end there should be a single Room, RoomChoiceGroup or RoomGroup
	// in the "Room" field. The "PreferredRooms" field in cleared.
	// If a list of rooms recurs, the same RoomChoiceGroup is used.
	//
	rref := Ref("")
	if len(course.getPreferredRooms()) > 1 {
		// Make a RoomChoiceGroup
		var estr string
		rref, estr = dbp.makeRoomChoiceGroup(course.getPreferredRooms())
		if estr != "" {
			logging.Error.Printf("In Course %s:\n%s", course.GetId(), estr)
		}
	} else if len(course.getPreferredRooms()) == 1 {
		// Check that room is Room or RoomGroup.
		rref0 := course.getPreferredRooms()[0]
		r, ok := dbp.Elements[rref0]
		if ok {
			_, ok = r.(*Room)
			if ok {
				rref = rref0
			} else {
				_, ok = r.(*RoomGroup)
				if ok {
					rref = rref0
				} else {
					logging.Error.Printf("Invalid room in Course %s:\n  %s\n",
						course.GetId(), rref0)
				}
			}
		} else {
			logging.Error.Printf("Unknown room in Course %s:\n  %s\n",
				course.GetId(), rref0)
		}
	}
	if course.GetRoom() != "" {
		if rref != "" {
			logging.Error.Printf(
				"Course has both Room and Rooms entries:\n %s\n",
				course.GetId())
		}
		r, ok := dbp.Elements[course.GetRoom()]
		if ok {
			_, ok = r.(*Room)
			if !ok {
				_, ok = r.(*RoomGroup)
				if !ok {
					_, ok = r.(*RoomChoiceGroup)
					if !ok {
						logging.Error.Printf(
							"Invalid room in Course %s:\n  %s\n",
							course.GetId(), course.GetRoom())
						course.setRoom("")
					}
				}
			}

		} else {
			logging.Error.Printf("Unknown room in Course %s:\n  %s\n",
				course.GetId(), course.GetRoom())
			course.setRoom("")
		}
	} else {
		course.setRoom(rref)
	}
	course.setPreferredRooms(nil)
}
