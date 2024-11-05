package w365tt

import (
	"log"
	"strconv"
	"strings"
)

func (dbp *DbTopLevel) readSubjects() {
	for _, n := range dbp.Subjects {
		_, nok := dbp.SubjectTags[n.Tag]
		if nok {
			log.Fatalf("*ERROR* Subject Tag (Shortcut) defined twice: %s\n",
				n.Tag)
		}
		t, nok := dbp.SubjectNames[n.Name]
		if nok {
			log.Printf("*WARNING* Subject Name defined twice (different"+
				" Tag/Shortcut):\n  %s (%s/%s)\n", n.Name, t, n.Tag)
		} else {
			dbp.SubjectNames[n.Name] = n.Tag
		}
		dbp.SubjectTags[n.Tag] = n.Id
	}
}

func (dbp *DbTopLevel) newSubject() string {
	// A rather primitive new-subject-tag generator
	i := 0
	for {
		i++
		tag := "X" + strconv.Itoa(i)
		_, nok := dbp.SubjectTags[tag]
		if !nok {
			return tag
		}
	}
}

func (dbp *DbTopLevel) readCourses() {
	for i := 0; i < len(dbp.Courses); i++ {
		n := &dbp.Courses[i]
		dbp.readCourse(n)
	}
}

func (dbp *DbTopLevel) readSuperCourses() {
	for _, n := range dbp.SuperCourses {
		s, ok := dbp.Elements[n.Subject]
		if !ok {
			log.Fatalf(
				"*ERROR* SuperCourse %s:\n  Unknown Subject: %s\n",
				n.Id, n.Subject)
		}
		_, ok = s.(*Subject)
		if !ok {
			log.Fatalf(
				"*ERROR* SuperCourse %s:\n  Not a Subject: %s\n",
				n.Id, n.Subject)
		}
	}
}

func (dbp *DbTopLevel) readSubCourses() {
	for i := 0; i < len(dbp.SubCourses); i++ {
		n := &dbp.SubCourses[i]
		s, ok := dbp.Elements[n.SuperCourse]
		if !ok {
			log.Fatalf(
				"*ERROR* SubCourse %s:\n  Unknown SuperCourse: %s\n",
				n.Id, n.SuperCourse)
		}
		_, ok = s.(*SuperCourse)
		if !ok {
			log.Fatalf(
				"*ERROR* SubCourse %s:\n  Not a SuperCourse: %s\n",
				n.Id, n.SuperCourse)
		}
		dbp.readCourse(n)
	}
}

func (dbp *DbTopLevel) readCourse(course CourseInterface) {
	//
	// Deal with the subject(s) fields
	//
	msg1 := "*ERROR* Course %s:\n  Unknown Subject: %s\n"
	msg2 := "*ERROR* Course %s:\n  Not a Subject: %s\n"
	if course.GetSubject() == "" {
		if len(course.getSubjects()) == 1 {
			wsid := course.getSubjects()[0]
			s0, ok := dbp.Elements[wsid]
			if !ok {
				log.Fatalf(msg1, course.GetId(), wsid)
			}
			if _, ok = s0.(*Subject); !ok {
				log.Fatalf(msg2, course.GetId(), wsid)
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
						log.Fatalf(msg2, course.GetId(), wsid)
					}
					sklist = append(sklist, s.Tag)
				} else {
					log.Fatalf(msg1, course.GetId(), wsid)
				}
			}
			skname := strings.Join(sklist, ",")
			stag, ok := dbp.SubjectNames[skname]
			if ok {
				// The Name has already been used.
				course.setSubject(dbp.SubjectTags[stag])
			} else {
				// Need a new Subject.
				stag = dbp.newSubject()
				sref := dbp.NewId()
				i := len(dbp.Subjects)
				dbp.Subjects = append(dbp.Subjects, Subject{
					Id:   sref,
					Tag:  stag,
					Name: skname,
				})
				dbp.AddElement(sref, &dbp.Subjects[i])
				dbp.SubjectTags[stag] = sref
				dbp.SubjectNames[skname] = stag
				course.setSubject(sref)
			}
		}
	} else {
		if len(course.getSubjects()) != 0 {
			log.Printf("*ERROR* Course has both Subject AND Subjects: %s\n",
				course.GetId())
		}
		wsid := course.GetSubject()
		s0, ok := dbp.Elements[wsid]
		if ok {
			_, ok = s0.(*Subject)
			if !ok {
				log.Fatalf(msg2, course.GetId(), wsid)
			}
		} else {
			log.Fatalf(msg1, course.GetId(), wsid)
		}
	}
	// Clear Subjects field.
	course.setSubjects(nil)

	//
	// Deal with groups
	//
	//glist := []Ref{}
	for _, gref := range course.GetGroups() {
		g, ok := dbp.Elements[gref]
		if !ok {
			log.Fatalf("*ERROR* Unknown group in Course %s:\n  %s\n",
				course.GetId(), gref)
			//continue
		}
		// g can be a Group or a Class.
		_, ok = g.(*Group)
		if !ok {
			// Check for class.
			_, ok = g.(*Class)
			if !ok {
				log.Fatalf("*ERROR* Invalid group in Course %s:\n  %s\n",
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
			log.Fatalf("*ERROR* Unknown teacher in Course %s:\n  %s\n",
				course.GetId(), tref)
			//continue
		}
		_, ok = t.(*Teacher)
		if !ok {
			log.Fatalf("*ERROR* Invalid teacher in Course %s:\n  %s\n",
				course.GetId(), tref)
			//continue
		}
		//tlist = append(tlist, tref)
	}

	//
	// Deal with rooms. W365 can have a single RoomGroup or a list of Rooms.
	//
	rref := Ref("")
	if len(course.getPreferredRooms()) > 1 {
		// Make a RoomChoiceGroup
		var estr string
		rref, estr = dbp.makeRoomChoiceGroup(course.getPreferredRooms())
		if estr != "" {
			log.Printf("*ERROR* In Course %s:\n%s", course.GetId(), estr)
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
					log.Printf("*ERROR* Invalid room in Course %s:\n  %s\n",
						course.GetId(), rref0)
				}
			}
		} else {
			log.Printf("*ERROR* Unknown room in Course %s:\n  %s\n",
				course.GetId(), rref0)
		}
	}
	if course.GetRoom() != "" {
		if rref != "" {
			log.Printf(
				"*ERROR* Course has both Room and Rooms entries:\n %s\n",
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
						log.Printf(
							"*ERROR* Invalid room in Course %s:\n  %s\n",
							course.GetId(), course.GetRoom())
						course.setRoom("")
					}
				}
			}

		} else {
			log.Printf("*ERROR* Unknown room in Course %s:\n  %s\n",
				course.GetId(), course.GetRoom())
			course.setRoom("")
		}
	} else {
		course.setRoom(rref)
	}
	course.setPreferredRooms(nil)
}
