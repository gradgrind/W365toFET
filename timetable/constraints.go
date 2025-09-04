package timetable

import (
	"W365toFET/base"
	"strings"
)

// This structure is used in the collection of "different-days" constraints.
// It records the parameters for the automatic "different-days" constraint
// and collects overrides for specific courses.
type differentDays struct {
	weight               int
	consecutiveIfSameDay bool
	daysBetween          map[NodeRef][]*base.DaysBetween
}

//TODO: Some more checks on duplicate or inconsistent constraints?

/* `processConstraints` transforms the constraint list from the database
 * into a more convenient form for the timetable at `tt_data.Constraints`.
 *
 * In the basic data the constraints handled here are simply represented
 * as a list of constraint nodes. This function collates them to produce a
 * map of constraint types to a list of those constraint nodes. It also
 * "preprocesses" some of the constraints where this can produce a more
 * convenient structure for their implementation:
 *
 * The constraints AutomaticDifferentDays, DaysBetween and DaysBetweenJoin are
 * processed and combined to be replaced by MinDaysBetweenLessons constraints,
 * which are then available directly as a field in the `TtData` structure.
 *
 * The ParallelCourses constraints are transformed to ParalllelLessons
 * constraints, which are also available directly as a field in the `TtData`
 * structure.
 */
func (tt_data *TtData) processConstraints() {
	db := tt_data.Db

	// Initialize the `differentDays` structure. If an "AutomaticDifferentDays"
	// constraint is present (at most one is permitted), the `weight` and
	// `consecutiveIfSameDay` fields will be set accordingly, otherwise the
	// default weight (`base.MAXWEIGHT`, i.e. a hard constraint) will be used.
	diffDays := differentDays{
		weight:               -1, // uninitialized
		consecutiveIfSameDay: false,
		daysBetween:          map[NodeRef][]*base.DaysBetween{},
	}

	// Collect the "min-days-between" constraints based on activities rather
	// than courses:
	mdba := []MinDaysBetweenLessons{}
	// Collect courses where automatic different-days is overridden:
	ddays := map[NodeRef]bool{}
	// Collect the categorized constraints:
	tt_data.Constraints = map[string][]any{}
	for _, c := range db.Constraints {
		{
			cn, ok := c.(*base.AutomaticDifferentDays)
			if ok { // At most one of these is permitted.
				if diffDays.weight < 0 {
					diffDays.weight = cn.Weight
					diffDays.consecutiveIfSameDay = cn.ConsecutiveIfSameDay
				} else {
					//TODO: Invalid input data ...
					panic("More than one AutomaticDifferentDays constraint")
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
				c1 := tt_data.Ref2CourseInfo[cn.Course1]
				c2 := tt_data.Ref2CourseInfo[cn.Course2]
				for i1, l1 := range c1.Lessons {
					for i2, l2 := range c2.Lessons {
						if l1.Fixed && l2.Fixed {
							// both fixed => no constraint
							continue
						}
						mdba = append(mdba, MinDaysBetweenLessons{
							Weight:               cn.Weight,
							ConsecutiveIfSameDay: cn.ConsecutiveIfSameDay,
							Activities: []ActivityIndex{
								c1.Activities[i1], c2.Activities[i2]},
							MinDays: cn.DaysBetween,
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
				footprint := []int{}         // activity durations
				var alen int = 0             // number of activities in each course
				var alists [][]ActivityIndex // collect the parallel activities
				for i, cref := range cn.Courses {
					cinfo := tt_data.Ref2CourseInfo[cref]
					if i == 0 {
						alen = len(cinfo.Activities)
						alists = make([][]ActivityIndex, alen)
					} else if len(cinfo.Activities) != alen {
						//TODO: This is a data error
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
							//TODO: This is a data error
							clist := []string{}
							for _, cr := range cn.Courses {
								clist = append(clist, string(cr))
							}
							base.Error.Fatalf("Parallel courses have lesson"+
								" mismatch: %s\n",
								strings.Join(clist, ","))
						}
						alists[j] = append(alists[j], cinfo.Activities[j])
					}
				}
				// llists is now a list of lists of parallel activity indexes.
				tt_data.ParallelLessons = append(tt_data.ParallelLessons,
					ParallelLessons{
						Weight:         cn.Weight,
						ActivityGroups: alists,
					})
				continue
			}
		}
		// Collect the other constraints according to type, but unmodified
		ctype := c.CType()
		tt_data.Constraints[ctype] = append(tt_data.Constraints[ctype], c)
	}
	// Resolve the differentDays constraints into days-between-activities
	if diffDays.weight < 0 {
		diffDays.weight = base.MAXWEIGHT
	}
	for _, cinfo := range tt_data.CourseInfoList {
		cref := cinfo.Id
		// Determine groups of activities to couple by means of the fixed flags.
		fixeds := []ActivityIndex{}
		unfixeds := []ActivityIndex{}
		for i, l := range cinfo.Lessons {
			if l.Fixed {
				fixeds = append(fixeds, cinfo.Activities[i])
			} else {
				unfixeds = append(unfixeds, cinfo.Activities[i])
			}
		}

		// Get the constraints for this course
		ddcs, ddcsok := diffDays.daysBetween[cref]
		if len(unfixeds) == 0 || (len(fixeds) == 0 && len(unfixeds) == 1) {
			// No constraints necessary
			if ddcsok {
				//TODO
				base.Warning.Printf("Superfluous DaysBetween constraint on"+
					" course:\n  -- %s", tt_data.View(cinfo))
			}
			continue
		}
		// Collect the activity groups to which the constraint is to be applied
		aidlists := [][]ActivityIndex{}
		if len(fixeds) <= 1 {
			// At most 1 fixed activity, so all activities are relevant
			aidlists = append(aidlists, cinfo.Activities)
		} else {
			// Multiple fixed activities, at least one unfixed one:
			for _, aidf := range fixeds {
				for _, aidu := range unfixeds {
					aidlists = append(aidlists, []ActivityIndex{aidf, aidu})
				}
			}
			if len(unfixeds) > 1 {
				aidlists = append(aidlists, unfixeds)
			}
		}
		// Add the constraints as `MinDaysBetweenLessons`
		if !ddays[cref] &&
			(diffDays.weight != 0 || diffDays.consecutiveIfSameDay) {
			// Add default constraint
			for _, alist := range aidlists {
				if len(alist) > tt_data.NDays {
					//TODO
					base.Warning.Printf("Course has too many lessons for"+
						"DifferentDays constraint:\n  -- %s\n",
						tt_data.View(cinfo))
					continue
				}
				mdba = append(mdba, MinDaysBetweenLessons{
					Weight:               diffDays.weight,
					ConsecutiveIfSameDay: diffDays.consecutiveIfSameDay,
					Activities:           alist,
					MinDays:              1,
				})
			}
		}
		// Generate the additional constraints
		if ddcsok {
			for _, ddc := range ddcs {
				if ddc.Weight != 0 || ddc.ConsecutiveIfSameDay {
					for _, alist := range aidlists {
						if (len(alist)-1)*ddc.DaysBetween >= tt_data.NDays {
							base.Warning.Printf("Course has too many lessons"+
								" for DaysBetween constraint:\n  -- %s\n",
								tt_data.View(cinfo))
							continue
						}
						mdba = append(mdba, MinDaysBetweenLessons{
							Weight:               ddc.Weight,
							ConsecutiveIfSameDay: ddc.ConsecutiveIfSameDay,
							Activities:           alist,
							MinDays:              ddc.DaysBetween,
						})
					}
				}
			}
		}
	}
	tt_data.MinDaysBetweenLessons = mdba
}
