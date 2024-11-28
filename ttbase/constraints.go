package ttbase

import (
	"W365toFET/base"
	"strings"
)

func processConstraints(ttinfo *TtInfo) {
	db := ttinfo.Db
	var doubleBlocked []bool
	for _, c := range db.Constraints {
		{
			cn, ok := c.(*base.AutomaticDifferentDays)
			if ok {
				if ttinfo.autoDifferentDays == nil {
					ttinfo.autoDifferentDays = cn
				} else {
					base.Bug.Fatalln(
						"More than one AutomaticDifferentDays constraint")
				}
				continue
			}
		}
		//TODO: Perhaps I should keep lists of lesson indexes together with the
		// constraint (parameters).
		{
			cn, ok := c.(*base.DaysBetween)
			if ok {
				for _, cref := range cn.Courses {
					ttinfo.daysBetween[cref] = append(
						ttinfo.daysBetween[cref], cn)
				}
				continue
			}
		}
		{
			cn, ok := c.(*base.DaysBetweenJoin)
			if ok {
				c1 := ttinfo.CourseInfo[cn.Course1]
				c2 := ttinfo.CourseInfo[cn.Course2]
				mdba := []minDaysBetweenActivities{}
				for _, l1 := range c1.activities {
					for _, l2 := range c2.activities {
						mdba = append(mdba, minDaysBetweenActivities{
							Weight_Percentage:       weight2fet(cn.Weight),
							Consecutive_If_Same_Day: cn.ConsecutiveIfSameDay,
							Number_of_Activities:    2,
							Activity_Id:             []int{l1, l2},
							MinDays:                 cn.DaysBetween,
							Active:                  true,
						})
					}
				}
				// Append constraints to full list
				tclist.ConstraintMinDaysBetweenActivities = append(
					tclist.ConstraintMinDaysBetweenActivities,
					mdba...)
				continue
			}
		}
		{
			cn, ok := c.(*base.LessonsEndDay)
			if ok {
				cinfo, ok := fetinfo.courseInfo[cn.Course]
				if !ok {
					base.Bug.Fatalf("Invalid course: %s\n", cn.Course)
				}
				for _, aid := range cinfo.activities {
					tclist.ConstraintActivityEndsStudentsDay = append(
						tclist.ConstraintActivityEndsStudentsDay,
						lessonEndsDay{
							Weight_Percentage: weight2fet(cn.Weight),
							Activity_Id:       aid,
							Active:            true,
						})
				}
				continue
			}
		}
		{
			cn, ok := c.(*base.BeforeAfterHour)
			if ok {
				timeslots := []preferredTime{}
				if cn.After {
					for _, dd := range fetinfo.days {
						for h := cn.Hour + 1; h < len(fetinfo.hours); h++ {
							timeslots = append(timeslots, preferredTime{
								Preferred_Day:  dd,
								Preferred_Hour: fetinfo.hours[h],
							})
						}
					}
				} else {
					for _, dd := range fetinfo.days {
						for h := 0; h < cn.Hour; h++ {
							timeslots = append(timeslots, preferredTime{
								Preferred_Day:  dd,
								Preferred_Hour: fetinfo.hours[h],
							})
						}
					}
				}
				for _, k := range cn.Courses {
					cinfo, ok := fetinfo.courseInfo[k]
					if !ok {
						base.Bug.Fatalf("Invalid course: %s\n", k)
					}
					for _, aid := range cinfo.activities {
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
				continue
			}
		}
		{
			cn, ok := c.(*base.ParallelCourses)
			if ok {
				// The courses must have the same number of lessons and the
				// lengths of the corresponding lessons must also be the same.
				// A constraint is generated for each lesson of the courses.

				// Check lesson lengths
				footprint := []int{} // lesson sizes
				ll := 0              // number of lessons in each course
				var alists [][]int   // collect the parallel activities
				for i, cref := range cn.Courses {
					cinfo := ttinfo.CourseInfo[cref]
					if i == 0 {
						ll = len(cinfo.Lessons)
						alists = make([][]int, ll)
					} else if len(cinfo.Lessons) != ll {
						clist := []string{}
						for _, cr := range cn.Courses {
							clist = append(clist, string(cr))
						}
						base.Error.Fatalf("Parallel courses have different"+
							" lessons: %s\n",
							strings.Join(clist, ","))
					}
					for j, l := range cinfo.Lessons {
						if i == 0 {
							footprint = append(footprint, l.Duration)
						} else if l.Duration != footprint[j] {
							clist := []string{}
							for _, cr := range cn.Courses {
								clist = append(clist, string(cr))
							}
							base.Error.Fatalf("Parallel courses have lesson"+
								" mismatch: %s\n",
								strings.Join(clist, ","))
						}
						alists[j] = append(alists[j], cinfo.activities[j])
					}
				}
				for _, alist := range alists {
					tclist.ConstraintActivitiesSameStartingTime = append(
						tclist.ConstraintActivitiesSameStartingTime,
						sameStartingTime{
							Weight_Percentage:    weight2fet(cn.Weight),
							Number_of_Activities: len(alist),
							Activity_Id:          alist,
							Active:               true,
						})
				}
				continue
			}
		}
		{
			cn, ok := c.(*base.DoubleLessonNotOverBreaks)
			if ok {
				if len(doubleBlocked) != 0 {
					base.Error.Fatalln("Constraint DoubleLessonNotOverBreaks" +
						" specified more than once")
				}
				timeslots := []preferredStart{}
				// Note that a double lesson can't start in the last slot of
				// the day.
				doubleBlocked = make([]bool, len(fetinfo.hours)-1)
				for _, h := range cn.Hours {
					doubleBlocked[h-1] = true
				}
				for _, dd := range fetinfo.days {
					for h, bl := range doubleBlocked {
						if !bl {
							timeslots = append(timeslots, preferredStart{
								Preferred_Starting_Day:  dd,
								Preferred_Starting_Hour: fetinfo.hours[h],
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
				continue
			}
		}
		/*
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
								MinDays:                 cn.DaysBetween,
								Active:                  true,
							})
						}
					}
					// Append constraints to full list
					tclist.ConstraintMinDaysBetweenActivities = append(
						tclist.ConstraintMinDaysBetweenActivities,
						mdba...)
					continue
				}
			}
		*/
	}
}
