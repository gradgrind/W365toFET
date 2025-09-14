package timetable

import (
	"W365toFET/base"
	"strings"
)

const ( // New, preprocessed constraint types
	C_GENERAL_DAYS_BETWEEN string = "TtDaysBetween"
	C_PARALLEL_ACTIVITIES  string = "TtParallelActivities"
)

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
 * The constraints AutomaticDifferentDays, DaysBetween and DaysBetweenJoin
 * are processed and combined to be replaced by MinDaysBetweenActivities
 * constraints, which are then available directly as a field in the `TtData`
 * structure.
 *
 * The ParallelCourses constraints are transformed to ParalllelLessons
 * constraints, which are also available directly as a field in the `TtData`
 * structure.
 */

type TtDaysBetween struct {
	Constraint           string
	Weight               int
	Course               NodeRef // Course or SuperCourse
	DaysBetween          int
	ConsecutiveIfSameDay bool
}

func (c *TtDaysBetween) IsHard() bool {
	return c.Weight == base.MAXWEIGHT
}

type TtParallelActivities struct {
	Constraint     string
	Weight         int
	Courses        []NodeRef // Courses or SuperCourses
	ActivityGroups [][]ActivityIndex
}

func (c *TtParallelActivities) IsHard() bool {
	return c.Weight == base.MAXWEIGHT
}

func (tt_data *TtData) preprocessConstraints() {
	db := tt_data.Db

	// If an "AutomaticDifferentDays" constraint is present (at most one is
	// permitted), the `auto_weight` and `auto_consec` values will be set
	// accordingly, otherwise the default weight (`base.MAXWEIGHT`, i.e.
	// a hard constraint) will be used.

	auto_weight := -1
	auto_consec := false
	noauto_ddays := map[NodeRef]bool{} // collect default overrides

	dd_hard := []any{}
	dd_soft := []any{}

	for _, c := range db.Constraints {
		{
			cn, ok := c.(*base.AutomaticDifferentDays)
			if ok { // At most one of these is permitted.
				if auto_weight < 0 {
					auto_weight = cn.Weight
					auto_consec = cn.ConsecutiveIfSameDay
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
					ddc := &TtDaysBetween{
						Constraint:           C_GENERAL_DAYS_BETWEEN,
						Weight:               cn.Weight,
						Course:               cref,
						DaysBetween:          cn.DaysBetween,
						ConsecutiveIfSameDay: cn.ConsecutiveIfSameDay,
					}
					if c.IsHard() {
						dd_hard = append(dd_hard, ddc)
					} else {
						dd_soft = append(dd_soft, ddc)
					}
					if cn.DaysBetween == 1 {
						// Override default constraint
						noauto_ddays[cref] = true
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

				//TODO: later ...
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
				// alists is now a list of lists of parallel activity indexes.
				cpa := &TtParallelActivities{
					Constraint:     C_PARALLEL_ACTIVITIES,
					Weight:         cn.Weight,
					Courses:        cn.Courses,
					ActivityGroups: alists,
				}
				if c.IsHard() {
					tt_data.HardConstraints[C_PARALLEL_ACTIVITIES] = append(
						tt_data.HardConstraints[C_PARALLEL_ACTIVITIES], cpa)
				} else {
					tt_data.SoftConstraints[C_PARALLEL_ACTIVITIES] = append(
						tt_data.SoftConstraints[C_PARALLEL_ACTIVITIES], cpa)
				}
				continue
			}
		}

		// Collect the other constraints according to type, but unmodified,
		// separating them into hard and soft constraints,
		ctype := c.CType()
		if c.IsHard() {
			tt_data.HardConstraints[ctype] = append(
				tt_data.HardConstraints[ctype], c)
		} else {
			tt_data.SoftConstraints[ctype] = append(
				tt_data.SoftConstraints[ctype], c)
		}
	}
	// Add automatic constraints, where implied.
	if auto_weight < 0 {
		auto_weight = base.MAXWEIGHT
	}
	for _, cinfo := range tt_data.CourseInfoList {
		cref := cinfo.Id

		if len(cinfo.Lessons) > 1 && !noauto_ddays[cref] {
			ddc := &TtDaysBetween{
				Constraint:           C_GENERAL_DAYS_BETWEEN,
				Weight:               auto_weight,
				Course:               cref,
				DaysBetween:          1,
				ConsecutiveIfSameDay: auto_consec,
			}
			if auto_weight == base.MAXWEIGHT {
				dd_hard = append(dd_hard, ddc)
			} else {
				dd_soft = append(dd_soft, ddc)
			}
		}
	}
	// Now add these as new constraints to the constraint map
	tt_data.HardConstraints[C_GENERAL_DAYS_BETWEEN] = dd_hard
	tt_data.SoftConstraints[C_GENERAL_DAYS_BETWEEN] = dd_soft
}

// Called before running the generator back-end to perform constraint
// conversions which have to be done after the constraint selection.
// Currently that is just the generation of the "MinDaysBetweenActivities"
// constraints from the "TtDaysBetween" and the "DaysBetweenJoin" types.
func PrepareSpecialConstraints(tt_data *TtData) {
	for _, c := range tt_data.HardConstraints["TtDaysBetween"] {
		tt_data.days_between_activities(c.(*TtDaysBetween))
	}
	for _, c := range tt_data.SoftConstraints["TtDaysBetween"] {
		tt_data.days_between_activities(c.(*TtDaysBetween))
	}
	for _, c := range tt_data.HardConstraints["DaysBetweenJoin"] {
		tt_data.days_between_join_activities(c.(*base.DaysBetweenJoin))
	}
	for _, c := range tt_data.SoftConstraints["DaysBetweenJoin"] {
		tt_data.days_between_join_activities(c.(*base.DaysBetweenJoin))
	}
}

// Convert a `TtDaysBetween` constraint to be based on activities.
func (tt_data *TtData) days_between_activities(
	constraint *TtDaysBetween,
) {
	cref := constraint.Course
	cinfo := tt_data.Ref2CourseInfo[cref]
	fixeds := []ActivityIndex{}
	unfixeds := []ActivityIndex{}
	for i, l := range cinfo.Lessons {
		if l.Fixed {
			fixeds = append(fixeds, cinfo.Activities[i])
		} else {
			unfixeds = append(unfixeds, cinfo.Activities[i])
		}
	}

	if len(unfixeds) == 0 || (len(fixeds) == 0 && len(unfixeds) == 1) {
		// No constraints necessary
		//TODO
		base.Warning.Printf("Superfluous DaysBetween constraint on"+
			" course:\n  -- %s", tt_data.View(cinfo))
		return
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
	// Add the constraints as `MinDaysBetweenActivities`
	if constraint.Weight != 0 || constraint.ConsecutiveIfSameDay {
		// Add constraint
		for _, alist := range aidlists {
			if len(alist) > tt_data.NDays {
				//TODO
				base.Warning.Printf("Course has too many lessons for"+
					"DifferentDays constraint:\n  -- %s\n",
					tt_data.View(cinfo))
				continue
			}
			if constraint.ConsecutiveIfSameDay ||
				constraint.IsHard() {
				// Note that "ConsecutiveIfSameDay" is hard regardless of
				// the weight.
				tt_data.HardMinDaysBetweenActivities = append(
					tt_data.HardMinDaysBetweenActivities, MinDaysBetweenActivities{
						Weight:               constraint.Weight,
						ConsecutiveIfSameDay: constraint.ConsecutiveIfSameDay,
						Activities:           alist,
						MinDays:              constraint.DaysBetween,
					})
			} else {
				tt_data.SoftMinDaysBetweenActivities = append(
					tt_data.SoftMinDaysBetweenActivities, MinDaysBetweenActivities{
						Weight:               constraint.Weight,
						ConsecutiveIfSameDay: constraint.ConsecutiveIfSameDay,
						Activities:           alist,
						MinDays:              constraint.DaysBetween,
					})
			}
		}
	}
}

// Convert a `DaysBetweenJoin` constraint to be based on activities.
func (tt_data *TtData) days_between_join_activities(
	constraint *base.DaysBetweenJoin,
) {
	c1 := tt_data.Ref2CourseInfo[constraint.Course1]
	c2 := tt_data.Ref2CourseInfo[constraint.Course2]
	for i1, l1 := range c1.Lessons {
		for i2, l2 := range c2.Lessons {
			if l1.Fixed && l2.Fixed {
				// both fixed => no constraint
				continue
			}
			if constraint.IsHard() || constraint.ConsecutiveIfSameDay {
				// Note that "ConsecutiveIfSameDay" is hard regardless of
				// the weight.
				tt_data.HardMinDaysBetweenActivities = append(
					tt_data.HardMinDaysBetweenActivities, MinDaysBetweenActivities{
						Weight:               constraint.Weight,
						ConsecutiveIfSameDay: constraint.ConsecutiveIfSameDay,
						Activities: []ActivityIndex{
							c1.Activities[i1], c2.Activities[i2]},
						MinDays: constraint.DaysBetween,
					})
			} else {
				tt_data.SoftMinDaysBetweenActivities = append(
					tt_data.SoftMinDaysBetweenActivities, MinDaysBetweenActivities{
						Weight:               constraint.Weight,
						ConsecutiveIfSameDay: constraint.ConsecutiveIfSameDay,
						Activities: []ActivityIndex{
							c1.Activities[i1], c2.Activities[i2]},
						MinDays: constraint.DaysBetween,
					})
			}
		}
	}
}
