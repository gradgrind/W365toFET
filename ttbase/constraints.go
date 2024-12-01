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

func (ttinfo *TtInfo) processConstraints() {
	// Some constraints can be "preprocessed" into more convenient structures.
	db := ttinfo.Db
	diffDays := differentDays{
		weight: -1, // uninitialized
	}
	mdba := []MinDaysBetweenLessons{}
	ddays := map[Ref]bool{} // collect diff days override flags
	ttinfo.Constraints = map[string][]any{}
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
					if cn.DaysBetween == 1 {
						// Override default constraint
						ddays[cref] = true
					}
				}
				continue
			}
		}
		{
			cn, ok := c.(*base.DaysBetweenJoin)
			if ok {
				c1 := ttinfo.CourseInfo[cn.Course1]
				c2 := ttinfo.CourseInfo[cn.Course2]
				for _, l1ref := range c1.Lessons {
					l1fixed := ttinfo.Activities[l1ref].Fixed
					for _, l2ref := range c2.Lessons {
						if l1fixed && ttinfo.Activities[l2ref].Fixed {
							// both fixed => no constraint
							continue
						}
						mdba = append(mdba, MinDaysBetweenLessons{
							Weight:               cn.Weight,
							ConsecutiveIfSameDay: cn.ConsecutiveIfSameDay,
							Lessons:              []int{l1ref, l2ref},
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
		// Determine groups of lessons to couple by means of the fixed flags.
		fixeds := []int{}
		unfixeds := []int{}
		for _, lref := range cinfo.Lessons {
			if ttinfo.Activities[lref].Fixed {
				fixeds = append(fixeds, lref)
			} else {
				unfixeds = append(unfixeds, lref)
			}
		}

		ddcs, ddcsok := diffDays.daysBetween[cref]
		if len(unfixeds) == 0 || (len(fixeds) == 0 && len(unfixeds) == 1) {
			// No constraints necessary
			if ddcsok {
				base.Warning.Printf("Superfluous DaysBetween constraint on"+
					" course:\n  -- %s", ttinfo.View(cinfo))
			}
			continue
		}
		aidlists := [][]int{}
		if len(fixeds) <= 1 {
			aidlists = append(aidlists, cinfo.Lessons)
		} else {
			for _, aid := range fixeds {
				aids := []int{aid}
				aids = append(aids, unfixeds...)
				aidlists = append(aidlists, aids)
			}
		}

		// Add constraints

		if !ddays[cref] && diffDays.weight != 0 {
			// Add default constraint
			for _, alist := range aidlists {
				if len(alist) > ttinfo.NDays {
					base.Warning.Printf("Course has too many lessons for"+
						"DifferentDays constraint:\n  -- %s\n",
						ttinfo.View(cinfo))
					continue
				}
				mdba = append(mdba, MinDaysBetweenLessons{
					Weight:               diffDays.weight,
					ConsecutiveIfSameDay: diffDays.consecutiveIfSameDay,
					Lessons:              alist,
					MinDays:              1,
				})
			}
		}

		if ddcsok {
			// Generate the additional constraints
			for _, ddc := range ddcs {
				if ddc.Weight != 0 {
					for _, alist := range aidlists {
						if (len(alist)-1)*ddc.DaysBetween >= ttinfo.NDays {
							base.Warning.Printf("Course has too many lessons"+
								" for DaysBetween constraint:\n  -- %s\n",
								ttinfo.View(cinfo))
							continue
						}
						mdba = append(mdba, MinDaysBetweenLessons{
							Weight:               ddc.Weight,
							ConsecutiveIfSameDay: ddc.ConsecutiveIfSameDay,
							Lessons:              alist,
							MinDays:              ddc.DaysBetween,
						})
					}
				}
			}
		}
	}
	ttinfo.MinDaysBetweenLessons = mdba
}
