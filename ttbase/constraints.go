package ttbase

import (
	"W365toFET/base"
	"strings"
)

type differentDays struct {
	weight               int
	consecutiveIfSameDay bool
	daysBetween          map[Ref][]*base.DaysBetween
}

type MinDaysBetweenLessons struct {
	// Result of processing constraints DifferentDays and DaysBetween
	Weight               int
	ConsecutiveIfSameDay bool
	Lessons              []int
	MinDays              int
}

type ParallelLessons struct {
	Weight       int
	LessonGroups [][]ActivityIndex
}

func processConstraints(ttinfo *TtInfo) {
	// Some constraints can be "preprocessed" into more convenient structures.
	db := ttinfo.Db
	diffDays := differentDays{
		weight: -1, // uninitialized
	}
	mdba := []MinDaysBetweenLessons{}
	for _, c := range db.Constraints {
		{
			cn, ok := c.(*base.AutomaticDifferentDays)
			if ok {
				if diffDays.weight < 0 {
					diffDays.weight = cn.Weight
					diffDays.consecutiveIfSameDay = cn.ConsecutiveIfSameDay
					diffDays.daysBetween = map[Ref][]*base.DaysBetween{}
				} else {
					base.Bug.Fatalln(
						"More than one AutomaticDifferentDays constraint")
				}
				continue
			}
		}
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
						mdba = append(mdba, MinDaysBetweenLessons{
							Weight:               cn.Weight,
							ConsecutiveIfSameDay: cn.ConsecutiveIfSameDay,
							Lessons:              []int{l1, l2},
							MinDays:              cn.DaysBetween,
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
						l := ttinfo.Activities[lix].Lesson
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
				ttinfo.ParallelLessons = append(ttinfo.ParallelLessons,
					ParallelLessons{
						Weight:       cn.Weight,
						LessonGroups: llists,
					})
				continue
			}
		}
		// Collect the other constraints according to type
		ctype := c.CType()
		ttinfo.Constraints[ctype] = append(ttinfo.Constraints[ctype], c)
	}
	// Resolve the differentDays constraints into days-between-lessons.
	if diffDays.weight < 0 {
		diffDays.weight = base.MAXWEIGHT
	}
	for cref, cinfo := range ttinfo.CourseInfo {
		ddcs, ok := diffDays.daysBetween[cref]
		if ok {
			// Generate the constraints in the list
			for _, ddc := range ddcs {
				if ddc.Weight != 0 {
					mdba = append(mdba, MinDaysBetweenLessons{
						Weight:               ddc.Weight,
						ConsecutiveIfSameDay: ddc.ConsecutiveIfSameDay,
						Lessons:              cinfo.Lessons,
						MinDays:              ddc.DaysBetween,
					})

				}
			}
		} else if diffDays.weight != 0 {
			// Generate the default constraint
			mdba = append(mdba, MinDaysBetweenLessons{
				Weight:               diffDays.weight,
				ConsecutiveIfSameDay: diffDays.consecutiveIfSameDay,
				Lessons:              cinfo.Lessons,
				MinDays:              1,
			})
		}
	}
	ttinfo.MinDaysBetweenLessons = mdba
}
