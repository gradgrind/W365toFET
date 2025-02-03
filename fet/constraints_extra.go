package fet

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"encoding/xml"
	"strconv"
)

type preferredSlots struct {
	XMLName                        xml.Name `xml:"ConstraintActivitiesPreferredTimeSlots"`
	Weight_Percentage              string
	Teacher                        string
	Students                       string
	Subject                        string
	Activity_Tag                   string
	Duration                       string
	Number_of_Preferred_Time_Slots int
	Preferred_Time_Slot            []preferredTime
	Active                         bool
}

type preferredTime struct {
	//XMLName                       xml.Name `xml:"Preferred_Time_Slot"`
	Preferred_Day  string
	Preferred_Hour string
}

type preferredStarts struct {
	XMLName                            xml.Name `xml:"ConstraintActivitiesPreferredStartingTimes"`
	Weight_Percentage                  string
	Teacher                            string
	Students                           string
	Subject                            string
	Activity_Tag                       string
	Duration                           string
	Number_of_Preferred_Starting_Times int
	Preferred_Starting_Time            []preferredStart
	Active                             bool
}

type preferredStart struct {
	//XMLName                       xml.Name `xml:"Preferred_Starting_Time"`
	Preferred_Starting_Day  string
	Preferred_Starting_Hour string
}

type lessonEndsDay struct {
	XMLName           xml.Name `xml:"ConstraintActivityEndsStudentsDay"`
	Weight_Percentage string
	Activity_Id       int
	Active            bool
}

type activityPreferredTimes struct {
	XMLName                        xml.Name `xml:"ConstraintActivityPreferredTimeSlots"`
	Weight_Percentage              string
	Activity_Id                    int
	Number_of_Preferred_Time_Slots int
	Preferred_Time_Slot            []preferredTime
	Active                         bool
}

// addSameStartingTime adds parallel constraints for the lessons
// of the courses in the supplied list.
func (fetinfo *fetInfo) addSameStartingTime(
	clist []Ref,
	weight string, // call with weight2fet(weight), if not "100"
) {
	tclist := &fetinfo.fetdata.Time_Constraints_List
	// Collect parallel activities from each of the courses
	llists := map[int][]ActivityIndex{}
	for _, cref := range clist {
		cinfo := fetinfo.ttinfo.CourseInfo[cref]
		for i, l := range cinfo.Lessons {
			llists[i] = append(llists[i], fetinfo.lessonActivity[l.Id])
		}
	}
	// Add the constraints
	for i := 0; i < len(llists); i++ {
		llist := llists[i]
		tclist.ConstraintActivitiesSameStartingTime = append(
			tclist.ConstraintActivitiesSameStartingTime,
			sameStartingTime{
				Weight_Percentage:    weight,
				Number_of_Activities: len(llist),
				Activity_Id:          llist,
				Active:               true,
			})
	}
}

// TODO: This will be rather difficult to do with the supplied data!
// Unfortunately it has been too far processed for this type of constraint.
// I would need to get ActivityIndexes from the LessonUnitIndexes and
// also filter out duplicate constraints ...
// This would be easier if an earlier form of the data was available!
// addDaysBetween adds MinDaysBetweenActivities constraints for the
// lessons in the supplied ActivityGroup.
func (fetinfo *fetInfo) addDaysBetween(ag *ttbase.ActivityGroup) {

}

func (fetinfo *fetInfo) getExtraConstraints() {
	tclist := &fetinfo.fetdata.Time_Constraints_List
	ttinfo := fetinfo.ttinfo

	//TODO--
	//for ctype := range ttinfo.Constraints {
	//	fmt.Printf("CTYPE: %s\n", ctype)
	//}

	// Deal with days-between constraints, first within one activity group
	ttplaces := ttinfo.Placements
	dgc := ttinfo.DayGapConstraints
	for agix, dbclist := range dgc.CourseConstraints {
		// Get the activity group
		ag := ttplaces.ActivityGroups[agix]
		// Only the first course is necessary, as any others in the list are
		// hard-parallel.
		cinfo := ttinfo.CourseInfo[ag.Courses[0]]
		llist := []ActivityIndex{}
		for _, l := range cinfo.Lessons {
			llist = append(llist, fetinfo.lessonActivity[l.Id])
		}
		for _, dbc := range dbclist {
			tclist.ConstraintMinDaysBetweenActivities = append(
				tclist.ConstraintMinDaysBetweenActivities,
				minDaysBetweenActivities{
					Weight_Percentage:       weight2fet(dbc.Weight),
					Consecutive_If_Same_Day: dbc.ConsecutiveIfSameDay,
					Number_of_Activities:    len(llist),
					Activity_Id:             llist,
					MinDays:                 dbc.DayGap,
					Active:                  true,
				})
		}
	}

	// ... then the cross-activity-group days-between constraints

	//TODO
	for agix, xdbcmap := range dgc.CrossCourseConstraints {
		for agix2, dbclist := range xdbcmap {

		}
	}

	// HardParallelCourses
	for _, cagix := range ttplaces.CourseActivityGroup {
		cag := ttplaces.ActivityGroups[cagix]
		courses := cag.Courses
		if len(courses) > 1 {
			fetinfo.addSameStartingTime(courses, "100")
		}
	}
	// SoftParallelConstraints
	for _, pl := range ttinfo.SoftParallelCourses {
		fetinfo.addSameStartingTime(pl.Courses, weight2fet(pl.Weight))
	}

	for _, c := range ttinfo.Constraints["LessonsEndDay"] {
		cn := c.(*base.LessonsEndDay)
		cinfo := ttinfo.CourseInfo[cn.Course]
		for _, l := range cinfo.Lessons {
			tclist.ConstraintActivityEndsStudentsDay = append(
				tclist.ConstraintActivityEndsStudentsDay,
				lessonEndsDay{
					Weight_Percentage: weight2fet(cn.Weight),
					Activity_Id:       fetinfo.lessonActivity[l.Id],
					Active:            true,
				})
		}
	}

	//TODO: Specification pending
	var doubleBlocked []bool
	for _, c := range ttinfo.Constraints["DoubleLessonNotOverBreaks"] {
		cn := c.(*base.DoubleLessonNotOverBreaks)

		if len(doubleBlocked) != 0 {
			base.Error.Fatalln("Constraint DoubleLessonNotOverBreaks" +
				" specified more than once")
		}

		timeslots := []preferredStart{}
		// Note that a double lesson can't start in the last slot of
		// the day.
		doubleBlocked = make([]bool, ttinfo.NHours-1)
		for _, h := range cn.Hours {
			doubleBlocked[h-1] = true
		}
		for d := 0; d < ttinfo.NDays; d++ {
			for h, bl := range doubleBlocked {
				if !bl {
					timeslots = append(timeslots, preferredStart{
						Preferred_Starting_Day:  strconv.Itoa(d),
						Preferred_Starting_Hour: strconv.Itoa(h),
					})
				}
			}
		}
		tclist.ConstraintActivitiesPreferredStartingTimes = append(
			tclist.ConstraintActivitiesPreferredStartingTimes,
			preferredStarts{
				Weight_Percentage:                  weight2fet(cn.Weight),
				Duration:                           "2",
				Number_of_Preferred_Starting_Times: len(timeslots),
				Preferred_Starting_Time:            timeslots,
				Active:                             true,
			})
	}

	for _, c := range ttinfo.Constraints["BeforeAfterHour"] {
		cn := c.(*base.BeforeAfterHour)
		timeslots := []preferredTime{}
		if cn.After {
			for d := 0; d < ttinfo.NDays; d++ {
				for h := cn.Hour + 1; h < ttinfo.NHours; h++ {
					timeslots = append(timeslots, preferredTime{
						Preferred_Day:  strconv.Itoa(d),
						Preferred_Hour: strconv.Itoa(h),
					})
				}
			}
		} else {
			for d := 0; d < ttinfo.NDays; d++ {
				for h := 0; h < cn.Hour; h++ {
					timeslots = append(timeslots, preferredTime{
						Preferred_Day:  strconv.Itoa(d),
						Preferred_Hour: strconv.Itoa(h),
					})
				}
			}
		}
		for _, k := range cn.Courses {
			cinfo, ok := ttinfo.CourseInfo[k]
			if !ok {
				base.Bug.Fatalf("Invalid course: %s\n", k)
			}
			for _, l := range cinfo.Lessons {
				tclist.ConstraintActivityPreferredTimeSlots = append(
					tclist.ConstraintActivityPreferredTimeSlots,
					activityPreferredTimes{
						Weight_Percentage:              weight2fet(cn.Weight),
						Activity_Id:                    fetinfo.lessonActivity[l.Id],
						Number_of_Preferred_Time_Slots: len(timeslots),
						Preferred_Time_Slot:            timeslots,
						Active:                         true,
					})
			}
		}
	}
	/* TODO: Specification pending
	{
		cn, ok := c.(*base.MinHoursFollowing)
		if ok {
			c1 := fetinfo.courseInfo[cn.Course1]
			c2 := fetinfo.courseInfo[cn.Course2]

			//TODO

			mdba := []minDaysBetweenActivities{}
			for _, l1 := range c1.activities {
				for _, l2 := range c2.activities {
					mdba = append(mdba, minDaysBetweenActivities{
						Weight_Percentage:       weight2fet(cn.Weight),
						Consecutive_If_Same_Day: cn.ConsecutiveIfSameDay,
						Number_of_Activities:    2,
						Activity_Id:             []int{l1, l2},
						MinDays:                 cn.DayGap,
						Active:                  true,
					})
				}
			}
			// Append constraints to full list
			tclist.ConstraintMinDaysBetweenActivities = append(
				tclist.ConstraintMinDaysBetweenActivities,
				mdba...)
		}
	}
	*/
}
