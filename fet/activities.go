package fet

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"encoding/xml"
	"slices"
	"strconv"
)

type fetActivity struct {
	XMLName           xml.Name `xml:"Activity"`
	Id                int
	Teacher           []string `xml:",omitempty"`
	Subject           string
	Activity_Tag      string   `xml:",omitempty"`
	Students          []string `xml:",omitempty"`
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

// Generate the fet activties.
func getActivities(fetinfo *fetInfo) []idMap {
	ttinfo := fetinfo.ttinfo
	ref2fet := ttinfo.Ref2Tag

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
	//for i := 0; i <
	for _, cinfo := range ttinfo.CourseInfo {
		// Teachers
		tlist := []string{}
		for _, ti := range cinfo.Teachers {
			tlist = append(tlist, ref2fet[ti])
		}
		slices.Sort(tlist)
		// Groups
		glist := []string{}
		for _, cgref := range cinfo.Groups {
			glist = append(glist, ref2fet[cgref])
		}
		slices.Sort(glist)
		/* ???
		atag := ""
		if slices.Contains(tagged_subjects, sbj) {
			atag = fmt.Sprintf("Tag_%s", sbj)
		}
		*/

		// Generate the Activities for this course (one per Lesson).
		totalDuration := 0
		llist := []*ttbase.Activity{}
		for _, lix := range cinfo.Lessons {
			l := ttinfo.Activities[lix]
			totalDuration += l.Duration
			llist = append(llist, l)
		}
		agid := 0 // first activity should have Id = 1
		if len(llist) > 1 {
			agid = llist[0].Index
		}
		for _, l := range llist {
			aid := l.Index
			activities = append(activities,
				fetActivity{
					Id:       aid,
					Teacher:  tlist,
					Subject:  ref2fet[cinfo.Subject],
					Students: glist,
					//Activity_Tag:      atag,
					Active:            true,
					Total_Duration:    totalDuration,
					Duration:          l.Duration,
					Activity_Group_Id: agid,
					Comments:          string(l.Lesson.Id),
				},
			)
		}
	}

	// Sort Activities
	slices.SortFunc(activities, func(a, b fetActivity) int {
		if a.Id < b.Id {
			return -1
		}
		return 1
	})
	lessonIdMap := []idMap{}
	for _, a := range activities {
		lessonIdMap = append(lessonIdMap, idMap{a.Id, a.Comments})
	}

	fetinfo.fetdata.Activities_List = fetActivitiesList{
		Activity: activities,
	}
	addPlacementConstraints(fetinfo)
	return lessonIdMap
}

func addPlacementConstraints(fetinfo *fetInfo) {
	ttinfo := fetinfo.ttinfo
	ref2fet := ttinfo.Ref2Tag
	for _, cinfo := range ttinfo.LessonCourses {
		// Set "preferred" rooms.
		rooms := fetinfo.getFetRooms(cinfo.Room)
		// Add the constraints.
		scl := &fetinfo.fetdata.Space_Constraints_List
		tcl := &fetinfo.fetdata.Time_Constraints_List
		for _, aid := range cinfo.Lessons {
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
			l := ttinfo.Activities[aid].Lesson
			if l.Day < 0 {
				continue
			}
			if !l.Fixed {
				continue
			}
			tcl.ConstraintActivityPreferredStartingTime = append(
				tcl.ConstraintActivityPreferredStartingTime,
				startingTime{
					Weight_Percentage:  100,
					Activity_Id:        aid,
					Preferred_Day:      strconv.Itoa(l.Day),
					Preferred_Hour:     strconv.Itoa(l.Hour),
					Permanently_Locked: l.Fixed,
					Active:             true,
				},
			)
			if ttinfo.WITHOUT_ROOM_PLACEMENTS || len(l.Rooms) == 0 {
				continue
			}

			// Get room tags of the Lesson's Rooms.
			rlist := []string{}
			for _, rref := range l.Rooms {
				rlist = append(rlist, ref2fet[rref])
			}

			// Special handling for FET's virtual rooms.

			if len(rooms) == 1 {
				// Check for virtual room.
				n, ok := fetinfo.fetVirtualRoomN[rooms[0]]
				if ok {
					if len(rlist) != n {
						base.Error.Printf(
							"Lesson %s:\n  Number of Rooms doesn't"+
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
				base.Error.Printf(
					"Course room is not virtual, but Lesson has"+
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
