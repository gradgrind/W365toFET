package fet

import (
	"W365toFET/base"
	"W365toFET/timetable"
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
	Activity_Id       timetable.ActivityIndex
	Active            bool
}

type activityPreferredTimes struct {
	XMLName                        xml.Name `xml:"ConstraintActivityPreferredTimeSlots"`
	Weight_Percentage              string
	Activity_Id                    timetable.ActivityIndex
	Number_of_Preferred_Time_Slots int
	Preferred_Time_Slot            []preferredTime
	Active                         bool
}

type sameStartingTime struct {
	XMLName              xml.Name `xml:"ConstraintActivitiesSameStartingTime"`
	Weight_Percentage    string
	Number_of_Activities int
	Activity_Id          []timetable.ActivityIndex
	Active               bool
}

func getExtraConstraints(fetinfo *fetInfo) {
	tclist := &fetinfo.fetdata.Time_Constraints_List
	tt_data := fetinfo.tt_data

	//TODO--
	//for ctype := range clist {
	//	fmt.Printf("CTYPE: %s\n", ctype)
	//}

	for _, dbc := range fetinfo.tt_data.HardMinDaysBetweenLessons {
		tclist.ConstraintMinDaysBetweenActivities = append(
			tclist.ConstraintMinDaysBetweenActivities,
			minDaysBetweenActivities{
				Weight_Percentage:       weight2fet(dbc.Weight),
				Consecutive_If_Same_Day: dbc.ConsecutiveIfSameDay,
				Number_of_Activities:    len(dbc.Activities),
				Activity_Id:             dbc.Activities,
				MinDays:                 dbc.MinDays,
				Active:                  true,
			})
	}
	for _, dbc := range fetinfo.tt_data.SoftMinDaysBetweenLessons {
		tclist.ConstraintMinDaysBetweenActivities = append(
			tclist.ConstraintMinDaysBetweenActivities,
			minDaysBetweenActivities{
				Weight_Percentage:       weight2fet(dbc.Weight),
				Consecutive_If_Same_Day: dbc.ConsecutiveIfSameDay,
				Number_of_Activities:    len(dbc.Activities),
				Activity_Id:             dbc.Activities,
				MinDays:                 dbc.MinDays,
				Active:                  true,
			})
	}

	//TODO: Specification pending
	var doubleBlocked []bool

	for _, clist := range []map[string][]any{
		tt_data.HardConstraints,
		tt_data.SoftConstraints,
	} {
		for _, c := range clist["TtParallelActivities"] {
			cn := c.(*timetable.TtParallelActivities)
			for _, alist := range cn.ActivityGroups {
				tclist.ConstraintActivitiesSameStartingTime = append(
					tclist.ConstraintActivitiesSameStartingTime,
					sameStartingTime{
						Weight_Percentage:    weight2fet(cn.Weight),
						Number_of_Activities: len(alist),
						Activity_Id:          alist,
						Active:               true,
					})
			}
		}

		for _, c := range clist["LessonsEndDay"] {
			cn := c.(*base.LessonsEndDay)
			cinfo := tt_data.Ref2CourseInfo[cn.Course]
			for _, aid := range cinfo.Activities {
				tclist.ConstraintActivityEndsStudentsDay = append(
					tclist.ConstraintActivityEndsStudentsDay,
					lessonEndsDay{
						Weight_Percentage: weight2fet(cn.Weight),
						Activity_Id:       aid,
						Active:            true,
					})
			}
		}

		for _, c := range clist["DoubleLessonNotOverBreaks"] {
			cn := c.(*base.DoubleLessonNotOverBreaks)

			if len(doubleBlocked) != 0 {
				base.Error.Fatalln("Constraint DoubleLessonNotOverBreaks" +
					" specified more than once")
			}

			timeslots := []preferredStart{}
			// Note that a double lesson can't start in the last slot of
			// the day.
			doubleBlocked = make([]bool, tt_data.NHours-1)
			for _, h := range cn.Hours {
				doubleBlocked[h-1] = true
			}
			for d := 0; d < tt_data.NDays; d++ {
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

		for _, c := range clist["BeforeAfterHour"] {
			cn := c.(*base.BeforeAfterHour)
			timeslots := []preferredTime{}
			if cn.After {
				for d := 0; d < tt_data.NDays; d++ {
					for h := cn.Hour + 1; h < tt_data.NHours; h++ {
						timeslots = append(timeslots, preferredTime{
							Preferred_Day:  strconv.Itoa(d),
							Preferred_Hour: strconv.Itoa(h),
						})
					}
				}
			} else {
				for d := 0; d < tt_data.NDays; d++ {
					for h := 0; h < cn.Hour; h++ {
						timeslots = append(timeslots, preferredTime{
							Preferred_Day:  strconv.Itoa(d),
							Preferred_Hour: strconv.Itoa(h),
						})
					}
				}
			}
			for _, k := range cn.Courses {
				cinfo, ok := tt_data.Ref2CourseInfo[k]
				if !ok {
					base.Bug.Fatalf("Invalid course: %s\n", k)
				}
				for _, aid := range cinfo.Activities {
					tclist.ConstraintActivityPreferredTimeSlots = append(
						tclist.ConstraintActivityPreferredTimeSlots,
						activityPreferredTimes{
							Weight_Percentage:              weight2fet(cn.Weight),
							Activity_Id:                    aid,
							Number_of_Preferred_Time_Slots: len(timeslots),
							Preferred_Time_Slot:            timeslots,
							Active:                         true,
						})
				}
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
						Activity_Id:             []timetable.ActivityIndex{l1, l2},
						MinDays:                 cn.DaysBetween,
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
