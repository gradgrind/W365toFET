package w365tt

import (
	"W365toFET/base"
	"W365toFET/logging"
	"strings"
)

func (db *DbTopLevel) readSubjects(newdb *base.DbTopLevel) {
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
		newdb.Subjects = append(newdb.Subjects, &base.Subject{
			Id:   e.Id,
			Tag:  e.Tag,
			Name: e.Name,
		})
	}
}

func (dbp *DbTopLevel) readCourses() {
	for i := 0; i < len(dbp.Courses); i++ {
		n := dbp.Courses[i]
		dbp.readCourse(n)
	}
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

func (dbp *DbTopLevel) readCourse(course CourseInterface) {
	//
	// Deal with the subject(s) fields. W365 allows multiple subjects, hence
	// the field "Subjects". For the interface there should only be a single
	// subject – in the "Subject"-field. If there is a "Subjects" entry, this
	// is converted to a single subject in the "Subject" field, if necessary
	// creating a new subject. Repeated use of the same subject list will
	// reuse the created subject. There should not be an entry in both
	// "Subject" and "Subjects" field.
	//
	msg1 := "Course %s:\n  Unknown Subject: %s\n"
	msg2 := "Course %s:\n  Not a Subject: %s\n"
	if course.GetSubject() == "" {
		if len(course.getSubjects()) == 1 {
			wsid := course.getSubjects()[0]
			s0, ok := dbp.Elements[wsid]
			if !ok {
				logging.Error.Fatalf(msg1, course.GetId(), wsid)
			}
			if _, ok = s0.(*Subject); !ok {
				logging.Error.Fatalf(msg2, course.GetId(), wsid)
			}
			course.setSubject(wsid)
		} else if len(course.getSubjects()) > 1 {
			// Make a subject name
			sklist := []string{}
			for _, wsid := range course.getSubjects() {
				// Need Tag/Shortcut field
				s0, ok := dbp.Elements[wsid]
				if ok {
					s, ok := s0.(*Subject)
					if !ok {
						logging.Error.Fatalf(msg2, course.GetId(), wsid)
					}
					sklist = append(sklist, s.Tag)
				} else {
					logging.Error.Fatalf(msg1, course.GetId(), wsid)
				}
			}
			skname := strings.Join(sklist, ",")
			stag, ok := dbp.SubjectNames[skname]
			if ok {
				// The Name has already been used.
				course.setSubject(dbp.SubjectTags[stag])
			} else {
				// Need a new Subject.
				course.setSubject(dbp.makeNewSubject("", skname))
			}
		}
	} else {
		if len(course.getSubjects()) != 0 {
			logging.Error.Printf("Course has both Subject AND Subjects: %s\n",
				course.GetId())
		}
		wsid := course.GetSubject()
		s0, ok := dbp.Elements[wsid]
		if ok {
			_, ok = s0.(*Subject)
			if !ok {
				logging.Error.Fatalf(msg2, course.GetId(), wsid)
			}
		} else {
			logging.Error.Fatalf(msg1, course.GetId(), wsid)
		}
	}
	// Clear Subjects field.
	course.setSubjects(nil)

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
