package fet

import (
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
func getActivities(fetinfo *fetInfo) {
	tt_data := fetinfo.tt_data
	shared_data := tt_data.SharedData

	// ************* Start with the activity tags
	tags := []fetActivityTag{}
	/* TODO ???
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
	for aid, tt_activity := range shared_data.Activities {
		if aid == 0 {
			continue
		}
		cinfo := tt_activity.CourseInfo
		// Teachers
		tlist := []string{}
		for _, ti := range cinfo.Teachers {
			tlist = append(tlist, shared_data.TeacherNodes[ti].GetResourceTag())
		}
		slices.Sort(tlist)
		// Groups
		glist := []string{}
		for _, cg := range cinfo.Groups {
			glist = append(glist, fetGroupTag(cg))
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
		for _, l := range cinfo.Lessons {
			totalDuration += l.Duration
			//llist = append(llist, l)
		}
		var agid timetable.ActivityIndex = 0
		if len(cinfo.Activities) > 1 {
			agid = cinfo.Activities[0]
		}
		activities = append(activities,
			fetActivity{
				Id:       timetable.ActivityIndex(aid),
				Teacher:  tlist,
				Subject:  cinfo.Subject,
				Students: glist,
				//Activity_Tag:      atag,
				Active:            true,
				Total_Duration:    totalDuration,
				Duration:          tt_activity.Lesson.Duration,
				Activity_Group_Id: agid,
				Comments:          string(tt_activity.Lesson.GetRef()),
			},
		)
	}
	fetinfo.fetdata.Activities_List = fetActivitiesList{
		Activity: activities,
	}
	addPlacementConstraints(fetinfo)
}

func addPlacementConstraints(fetinfo *fetInfo) {
	tt_data := fetinfo.tt_data

	armap := map[int]struct{}{}
	for _, a0 := range tt_data.HardConstraints[timetable.ActivityRooms] {
		armap[int(a0.(*timetable.ActivityRoomConstraint).ActivityIndex)] = struct{}{}
	}

	for _, cinfo := range tt_data.SharedData.CourseInfoList {
		var rooms []string
		// Set "preferred" rooms, if not blocked.
		if !tt_data.WITHOUT_ROOM_PLACEMENTS {
			rooms = fetinfo.getFetRooms(cinfo)
		}

		//--fmt.Printf("COURSE: %s\n", ttinfo.View(cinfo))
		//--fmt.Printf("   --> %+v\n", rooms)

		// Add the constraints.
		scl := &fetinfo.fetdata.Space_Constraints_List
		tcl := &fetinfo.fetdata.Time_Constraints_List
		for i, l := range cinfo.Lessons {
			aid := cinfo.Activities[i]
			_, ok := armap[int(aid)]
			if ok && len(rooms) != 0 {
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

			// The Rooms field of a Lesson item is not used for building
			// FET input files. All room constraints are handled by
			// "ConstraintActivityPreferredRooms".
		}
	}
}
