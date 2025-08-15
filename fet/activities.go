package fet

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"encoding/xml"
	"slices"
	"strconv"
)

type fetActivity struct {
	XMLName           xml.Name `xml:"Activity"`
	Id                timetable.ActivityIndex
	Teacher           []string `xml:",omitempty"`
	Subject           string
	Activity_Tag      string   `xml:",omitempty"`
	Students          []string `xml:",omitempty"`
	Active            bool
	Total_Duration    int
	Duration          int
	Activity_Group_Id timetable.ActivityIndex
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
	tt_data := fetinfo.tt_data
	db := tt_data.Db
	//ref2fet := tt_data.Db.Ref2Tag

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
	for _, cinfo := range tt_data.CourseInfoList {
		// Teachers
		tlist := []string{}
		for _, ti := range cinfo.Teachers {
			tlist = append(tlist, db.Ref2Tag(ti))
		}
		slices.Sort(tlist)
		// Groups
		glist := []string{}
		for _, cgref := range cinfo.Groups {
			glist = append(glist, db.Ref2Tag(cgref))
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
		//llist := []*ttbase.Activity{}
		for _, l := range cinfo.Activities {
			totalDuration += l.Duration
			//llist = append(llist, l)
		}
		var agid timetable.ActivityIndex = 0
		if len(cinfo.Activities) > 1 {
			agid = cinfo.ActivityGroup[0]
		}
		for i, l := range cinfo.Activities {
			aid := cinfo.ActivityGroup[i]
			activities = append(activities,
				fetActivity{
					Id:       aid,
					Teacher:  tlist,
					Subject:  db.Ref2Tag(cinfo.Subject),
					Students: glist,
					//Activity_Tag:      atag,
					Active:            true,
					Total_Duration:    totalDuration,
					Duration:          l.Duration,
					Activity_Group_Id: agid,
					Comments:          string(l.Id),
				},
			)
		}
	}

	// Sort Activities - TODO: is this necessary?
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
	tt_data := fetinfo.tt_data
	db := tt_data.Db
	for _, cinfo := range tt_data.CourseInfoList {
		// Set "preferred" rooms.
		rooms := fetinfo.getFetRooms(cinfo.Rooms)

		//--fmt.Printf("COURSE: %s\n", ttinfo.View(cinfo))
		//--fmt.Printf("   --> %+v\n", rooms)

		// Add the constraints.
		scl := &fetinfo.fetdata.Space_Constraints_List
		tcl := &fetinfo.fetdata.Time_Constraints_List
		for i, l := range cinfo.Activities {
			aid := cinfo.ActivityGroup[i]
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

			//TODO:
			//if tt_data.WITHOUT_ROOM_PLACEMENTS {
			//	continue
			//}
			if len(l.Rooms) == 0 {
				continue
			}

			// Get room tags of the Lesson's Rooms.
			rlist := []string{}
			for _, rref := range l.Rooms {
				rlist = append(rlist, db.Ref2Tag(rref))
			}

			// Special handling for FET's virtual rooms.

			if len(rooms) == 1 {
				// Check for virtual room.
				n, ok := fetinfo.fetVirtualRoomN[rooms[0]]
				if ok {
					if len(rlist) != n {
						// FET can't cope with this.
						// A warning should have been issued in ttbase.
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
