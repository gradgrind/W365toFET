package fet

import (
	"W365toFET/w365tt"
	"encoding/xml"
	"log"
	"slices"
)

type fetActivity struct {
	XMLName           xml.Name `xml:"Activity"`
	Id                int
	Teacher           []string
	Subject           string
	Activity_Tag      string `xml:",omitempty"`
	Students          []string
	Active            bool
	Total_Duration    int
	Duration          int
	Activity_Group_Id int
	Comments          string
}

type fetActivitiesList struct {
	XMLName  xml.Name `xml:"Activities_List"`
	Activity []fetActivity
}

type fetActivityTag struct {
	XMLName   xml.Name `xml:"Activity_Tag"`
	Name      string
	Printable bool
}

type fetActivityTags struct {
	XMLName      xml.Name `xml:"Activity_Tags_List"`
	Activity_Tag []fetActivityTag
}

func gatherCourseInfo(fetinfo *fetInfo) {
	// Gather the Groups, Teachers and "rooms" for the Courses and
	// SuperCourses with lessons (only).
	// Gather the Lessons for these Courses and SuperCourses.
	// Also, the SuperCourses (with lessons) get a list of their
	// SubCourses.
	db := fetinfo.db
	fetinfo.superSubs = make(map[Ref][]Ref)
	fetinfo.courseInfo = make(map[Ref]courseInfo)
	roomData := map[Ref][]Ref{} // course-Ref -> []room-Ref (any sort of "room")

	// Collect Courses with Lessons.
	for i := 0; i < len(db.Lessons); i++ {
		l := &db.Lessons[i]
		lcref := l.Course
		cinfo, ok := fetinfo.courseInfo[lcref]
		if ok {
			// If the course has already been handled, just add the lesson.
			cinfo.lessons = append(cinfo.lessons, l)
			fetinfo.courseInfo[lcref] = cinfo
			continue
		}
		// First encounter with the course.
		var subject Ref
		var groups []Ref
		var teachers []Ref
		var rooms []Ref
		lessons := []*w365tt.Lesson{l}
		actids := []int{}

		c := db.Elements[lcref] // can be Course or SuperCourse
		cnode, ok := c.(*w365tt.Course)
		if ok {
			subject = cnode.Subject
			groups = cnode.Groups
			teachers = cnode.Teachers
			rooms = []Ref{}
			if cnode.Room != "" {
				rooms = append(rooms, cnode.Room)
			}
		} else {
			spc, ok := c.(*w365tt.SuperCourse)
			if !ok {
				log.Fatalf(
					"*ERROR* Invalid Course in Lesson %s:\n  %s\n",
					l.Id, lcref)
			}
			fetinfo.superSubs[lcref] = []Ref{}
			subject = spc.Subject
			groups = []Ref{}
			teachers = []Ref{}
			rooms = []Ref{}
		}
		fetinfo.courseInfo[lcref] = courseInfo{
			subject:  subject,
			groups:   groups,
			teachers: teachers,
			//rooms: filled later
			lessons:    lessons,
			activities: actids,
		}
		roomData[lcref] = rooms
	}

	// Now find the SubCourses, adding their groups, teachers and rooms.
	for _, sbc := range db.SubCourses {
		spc := sbc.SuperCourse
		// Only fill SuperCourses which have Lessons
		cinfo, ok := fetinfo.courseInfo[spc]
		if ok {
			fetinfo.superSubs[spc] = append(fetinfo.superSubs[spc], sbc.Id)

			// Add groups
			if len(sbc.Groups) != 0 {
				cglist := append(cinfo.groups, sbc.Groups...)
				slices.Sort(cglist)
				cglist = slices.Compact(cglist)
				cinfo.groups = make([]Ref, len(cglist))
				copy(cinfo.groups, cglist)
			}

			// Add teachers
			if len(sbc.Teachers) != 0 {
				ctlist := append(cinfo.teachers, sbc.Teachers...)
				slices.Sort(ctlist)
				ctlist = slices.Compact(ctlist)
				cinfo.teachers = make([]Ref, len(ctlist))
				copy(cinfo.teachers, ctlist)
			}

			// Add rooms
			if sbc.Room != "" {
				crlist := append(roomData[spc], sbc.Room)
				slices.Sort(crlist)
				crlist = slices.Compact(crlist)
				roomData[spc] = crlist
			}

			fetinfo.courseInfo[spc] = cinfo
		}
	}

	// Prepare the internal room structure, filtering the room lists of
	// the SuperCourses.
	for cref, crlist := range roomData {
		// Join all Rooms and the Rooms from RoomGroups into a "compulsory"
		// list. Then go through the RoomChoiceGroups. If one contains a
		// compulsory room, ignore the choice.
		// The result is a list of Rooms and a list of Room-choice-lists,
		// which can be converted into a fet virtual room.
		rooms := []Ref{}
		roomChoices := [][]Ref{}
		for _, rref := range crlist {
			rx := fetinfo.db.Elements[rref]
			_, ok := rx.(*w365tt.Room)
			if ok {
				rooms = append(rooms, rref)
			} else {
				rg, ok := rx.(*w365tt.RoomGroup)
				if ok {
					rooms = append(rooms, rg.Rooms...)
				} else {
					rc, ok := rx.(*w365tt.RoomChoiceGroup)
					if !ok {
						log.Fatalf(
							"*BUG* Invalid room in course %s:\n  %s\n",
							cref, rref)
					}
					roomChoices = append(roomChoices, rc.Rooms)
				}
			}
		}
		// Remove duplicates in Room list.
		slices.Sort(rooms)
		rooms = slices.Compact(rooms)
		// Filter choice lists.
		roomChoices = slices.DeleteFunc(roomChoices, func(rcl []Ref) bool {
			for _, rc := range rcl {
				if slices.Contains(rooms, rc) {
					return true
				}
			}
			return false
		})
		cinfo := fetinfo.courseInfo[cref]
		cinfo.room = virtualRoom{
			rooms:       rooms,
			roomChoices: roomChoices,
		}
		fetinfo.courseInfo[cref] = cinfo
	}
}

// Generate the fet activties.
func getActivities(fetinfo *fetInfo) []idMap {

	// ************* Start with the activity tags
	tags := []fetActivityTag{}
	/* ???
	s2tag := map[string]string{}
	for _, ts := range tagged_subjects {
		tag := fmt.Sprintf("Tag_%s", ts)
		s2tag[ts] = tag
		tags = append(tags, fetActivityTag{
			Name: tag,
		})
	}
	*/
	fetinfo.fetdata.Activity_Tags_List = fetActivityTags{
		Activity_Tag: tags,
	}
	// ************* Now the activities
	activities := []fetActivity{}
	lessonIdMap := []idMap{}
	aid := 0
	//for i := 0; i <
	for cref, cinfo := range fetinfo.courseInfo {
		// Teachers
		tlist := []string{}
		for _, ti := range cinfo.teachers {
			tlist = append(tlist, fetinfo.ref2fet[ti])
		}
		// Groups
		glist := []string{}
		for _, cgref := range cinfo.groups {
			glist = append(glist, fetinfo.ref2fet[cgref])
		}
		/* ???
		atag := ""
		if slices.Contains(tagged_subjects, sbj) {
			atag = fmt.Sprintf("Tag_%s", sbj)
		}
		*/

		// Generate the Activities for this course (one per Lesson).
		agid := 0 // first activity should have Id = 1
		if len(cinfo.lessons) > 1 {
			agid = aid + 1
		}
		totalDuration := 0
		for _, l := range cinfo.lessons {
			totalDuration += l.Duration
		}
		for _, l := range cinfo.lessons {
			aid++
			cinfo.activities = append(cinfo.activities, aid)
			activities = append(activities,
				fetActivity{
					Id:       aid,
					Teacher:  tlist,
					Subject:  fetinfo.ref2fet[cinfo.subject],
					Students: glist,
					//Activity_Tag:      atag,
					Active:            true,
					Total_Duration:    totalDuration,
					Duration:          l.Duration,
					Activity_Group_Id: agid,
					Comments:          string(l.Id),
				},
			)
			// Also add to Id-map.
			lessonIdMap = append(
				lessonIdMap, idMap{aid, l.Id})
		}
		fetinfo.courseInfo[cref] = cinfo
	}
	fetinfo.fetdata.Activities_List = fetActivitiesList{
		Activity: activities,
	}
	addPlacementConstraints(fetinfo)
	addDifferentDaysConstraints(fetinfo)
	return lessonIdMap
}

func addPlacementConstraints(fetinfo *fetInfo) {
	for _, cinfo := range fetinfo.courseInfo {
		// Set "preferred" rooms.
		rooms := getFetRooms(fetinfo, cinfo.room)
		// Add the constraints.
		scl := &fetinfo.fetdata.Space_Constraints_List
		tcl := &fetinfo.fetdata.Time_Constraints_List
		for i, aid := range cinfo.activities {
			if len(rooms) != 0 {
				scl.ConstraintActivityPreferredRooms = append(
					scl.ConstraintActivityPreferredRooms,
					roomChoice{
						Weight_Percentage:         100,
						Activity_Id:               aid,
						Number_of_Preferred_Rooms: len(rooms),
						Preferred_Room:            rooms,
						Active:                    true,
					},
				)
			}
			l := cinfo.lessons[i]
			if l.Day < 0 {
				continue
			}
			if fetinfo.ONLY_FIXED && !l.Fixed {
				continue
			}
			tcl.ConstraintActivityPreferredStartingTime = append(
				tcl.ConstraintActivityPreferredStartingTime,
				startingTime{
					Weight_Percentage:  100,
					Activity_Id:        aid,
					Preferred_Day:      fetinfo.days[l.Day],
					Preferred_Hour:     fetinfo.hours[l.Hour],
					Permanently_Locked: l.Fixed,
					Active:             true,
				},
			)
			if fetinfo.WITHOUT_ROOM_PLACEMENTS || len(l.Rooms) == 0 {
				continue
			}

			// Get room tags of the Lesson's Rooms.
			rlist := []string{}
			for _, rref := range l.Rooms {
				rlist = append(rlist, fetinfo.ref2fet[rref])
			}

			// Special handling for FET's virtual rooms.

			if len(rooms) == 1 {
				// Check for virtual room.
				n, ok := fetinfo.fetVirtualRoomN[rooms[0]]
				if ok {
					if len(rlist) != n {
						log.Printf(
							"*ERROR* Lesson %s:\n  Number of Rooms doesn't"+
								" match virtual room (%s) in course. \n",
							l.Id, rooms[0])
						continue
					}
					scl.ConstraintActivityPreferredRoom = append(
						scl.ConstraintActivityPreferredRoom,
						placedRoom{
							Weight_Percentage:    100,
							Activity_Id:          aid,
							Room:                 rooms[0],
							Number_of_Real_Rooms: len(rlist),
							Real_Room:            rlist,
							Permanently_Locked:   false,
							Active:               true,
						},
					)
					continue
				}
			}

			if len(rlist) != 1 {
				log.Printf(
					"*ERROR* Course room is not virtual, but Lesson has"+
						" more than one Room:\n  %s", l.Id)
				continue
			}

			scl.ConstraintActivityPreferredRoom = append(
				scl.ConstraintActivityPreferredRoom,
				placedRoom{
					Weight_Percentage:  100,
					Activity_Id:        aid,
					Room:               rlist[0],
					Permanently_Locked: false,
					Active:             true,
				},
			)
		}
	}
}
