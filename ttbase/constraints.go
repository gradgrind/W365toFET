package ttbase

import (
	"W365toFET/base"
	"strings"
)

type differentDays struct {
	weight               int
	consecutiveIfSameDay bool
	daysBetween          map[Ref][]*base.DaysBetween
	daysBetweenJoin      map[Ref][]*base.DaysBetweenJoin
}

type minDaysBetweenActivities struct {
	weight               int
	consecutiveIfSameDay bool
	lessons              []int
	minDays              int
}

func processConstraints(ttinfo *TtInfo) {
	// Some constraints can be "preprocessed" into more convenient structures.
	db := ttinfo.Db
	diffDays := differentDays{
		weight: -1, // uninitialized
	}
	mdba := []minDaysBetweenActivities{}
	for _, c := range db.Constraints {
		{
			cn, ok := c.(*base.AutomaticDifferentDays)
			if ok {
				if diffDays.weight < 0 {
					diffDays.weight = cn.Weight
					diffDays.consecutiveIfSameDay = cn.ConsecutiveIfSameDay
					diffDays.daysBetween = map[Ref][]*base.DaysBetween{}
					diffDays.daysBetweenJoin = map[Ref][]*base.DaysBetweenJoin{}
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
					diffDays.daysBetween[cref] = append(
						diffDays.daysBetween[cref], cn)
				}
				continue
			}
		}
		{
			cn, ok := c.(*base.DaysBetweenJoin)
			if ok {
				c1 := ttinfo.CourseInfo[cn.Course1]
				c2 := ttinfo.CourseInfo[cn.Course2]
				for _, l1 := range c1.Lessons {
					for _, l2 := range c2.Lessons {
						mdba = append(mdba, minDaysBetweenActivities{
							weight:               cn.Weight,
							consecutiveIfSameDay: cn.ConsecutiveIfSameDay,
							lessons:              []int{l1, l2},
							minDays:              cn.DaysBetween,
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
				var llists [][]int   // collect the parallel lessons
				for i, cref := range cn.Courses {
					cinfo := ttinfo.CourseInfo[cref]
					if i == 0 {
						ll = len(cinfo.Lessons)
						llists = make([][]int, ll)
					} else if len(cinfo.Lessons) != ll {
						clist := []string{}
						for _, cr := range cn.Courses {
							clist = append(clist, string(cr))
						}
						base.Error.Fatalf("Parallel courses have different"+
							" lessons: %s\n",
							strings.Join(clist, ","))
					}
					for j, lix := range cinfo.Lessons {
						l := ttinfo.TtLessons[lix].Lesson
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
						llists[j] = append(llists[j], lix)
					}
				}
				// llists is now a list of lists of parallel TtLesson indexes.
				cn.Activities = llists
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
