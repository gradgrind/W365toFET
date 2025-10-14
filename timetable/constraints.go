package timetable

import (
	"W365toFET/base"
	"strings"
)

// These are the constraint types known to the "timetable" package. Some of
// them correspond directly to the constraint in the "base" package, others
// are modified to better suit the needs of the timetable generator.
// For some of them the timetable generator back-end may need a form which is
// more difficult to relate to the original data (such as constraints
// connecting activities rather than courses). For such constraints there
// are functions to perform conversions.
type ConstraintType int

// Run in this directory to generate the String() method:
// stringer --type ConstraintType

const (
	TeacherMinLessonsPerDay ConstraintType = iota
	TeacherMaxLessonsPerDay
	TeacherMaxAfternoons
	TeacherMaxDays
	TeacherLunchBreak
	TeacherMaxGapsPerDay
	TeacherMaxGapsPerWeek

	ClassMinLessonsPerDay
	ClassMaxLessonsPerDay
	ClassMaxAfternoons
	ClassLunchBreak
	ClassForceFirstHour
	ClassMaxGapsPerDay
	ClassMaxGapsPerWeek

	ActivitiesEndDay
	BeforeAfterHour
	DaysBetweenJoin
	MinDaysBetween
	ParallelCourses
	MinHoursFollowing

	DoubleActivityNotOverBreaks //??? This is a one-off, handle specially?

	ActivityRooms

	LastConstraint // not a real constraint, it can be used as the total
	// number of constraints.
)

// Associate constraint names with their indexes
var cnmap map[string]ConstraintType

func init() {
	cnmap = make(map[string]ConstraintType, LastConstraint)
	for cnx := range LastConstraint {
		cnmap[cnx.String()] = cnx
	}
}

const C_GENERAL_DAYS_BETWEEN string = "TtDaysBetween"

//TODO: Some more checks on duplicate or inconsistent constraints?

/* The constraints AutomaticDifferentDays, DaysBetween are processed and
 * combined to be replaced by TtDaysBetween constraints. These also gain
 * activity lists to assist in the implementation of the constraint.
 */
type TtDaysBetween struct {
	Constraint           string
	Weight               int
	Course               NodeRef // Course or SuperCourse
	DaysBetween          int
	ConsecutiveIfSameDay bool
	ActivityLists        [][]ActivityIndex
}

func (c *TtDaysBetween) IsHard() bool {
	// Note that "ConsecutiveIfSameDay" is hard regardless of
	// the weight.
	return c.Weight == base.MAXWEIGHT || c.ConsecutiveIfSameDay
}

/* The DaysBetweenJoin constraints are converted to TtDaysBetweenJoin and
 * gain activity lists to assist in the implementation of the constraint.
 */
type TtDaysBetweenJoin struct {
	Constraint           string
	Weight               int
	Course1              NodeRef // Course or SuperCourse
	Course2              NodeRef // Course or SuperCourse
	DaysBetween          int
	ConsecutiveIfSameDay bool
	ActivityLists        [][]ActivityIndex
}

func (c *TtDaysBetweenJoin) IsHard() bool {
	// Note that "ConsecutiveIfSameDay" is hard regardless of
	// the weight.
	return c.Weight == base.MAXWEIGHT || c.ConsecutiveIfSameDay
}

/* The ParallelCourses constraints are transformed to TtParallelActivities
 * constraints.
 */
type TtParallelActivities struct {
	Constraint     string
	Weight         int
	Courses        []NodeRef // Courses or SuperCourses
	ActivityGroups [][]ActivityIndex
}

func (c *TtParallelActivities) IsHard() bool {
	return c.Weight == base.MAXWEIGHT
}

/* `preprocessConstraints` transforms the constraint list from the the
 * "base" data into a more convenient form for the timetable at
 * `tt_data.HardConstraints` and `tt_data.SoftConstraints`.
 *
 * The result is a map, constraint-type -> list of constraints.
 *
 * Some of the constraints are "preprocessed" to produce a more convenient
 * structure for their implementation.
 */
//TODO: rooms, fixed and choices
func (tt_data *TtData) preprocessConstraints() {
	tt_shared_data := tt_data.SharedData
	db := tt_shared_data.Db

	// Add active teacher and class constraints to the
	// `TtData.HardConstraints` structure.
	tt_data.collect_teacher_constraints()
	tt_data.collect_class_constraints()
	tt_data.collect_room_constraints()

	// If an "AutomaticDifferentDays" constraint is present (at most one is
	// permitted), the `auto_weight` and `auto_consec` variables will be set
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
					cn1 := &TtDaysBetween{
						Constraint:           C_GENERAL_DAYS_BETWEEN,
						Weight:               cn.Weight,
						Course:               cref,
						DaysBetween:          cn.DaysBetween,
						ConsecutiveIfSameDay: cn.ConsecutiveIfSameDay,
					}
					cn1.ActivityLists = tt_shared_data.days_between_activities(cn1)
					if c.IsHard() {
						dd_hard = append(dd_hard, cn1)
					} else {
						dd_soft = append(dd_soft, cn1)
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
			cn, ok := c.(*base.DaysBetweenJoin)
			if ok {
				cn1 := &TtDaysBetweenJoin{
					Constraint:           cn.Constraint,
					Weight:               cn.Weight,
					Course1:              cn.Course1,
					Course2:              cn.Course2,
					DaysBetween:          cn.DaysBetween,
					ConsecutiveIfSameDay: cn.ConsecutiveIfSameDay,
					ActivityLists:        tt_shared_data.days_between_join_activities(cn),
				}
				// Note that "ConsecutiveIfSameDay" is hard regardless of
				// the weight.
				if cn1.IsHard() {
					tt_data.HardConstraints[DaysBetweenJoin] = append(
						tt_data.HardConstraints[DaysBetweenJoin], cn1)
				} else {
					tt_data.SoftConstraints[DaysBetweenJoin] = append(
						tt_data.SoftConstraints[DaysBetweenJoin], cn1)
				}
				continue
			}
		}

		{
			cn, ok := c.(*base.ParallelCourses)
			if ok {
				// The courses must have the same number of activities and the
				// lengths of the corresponding activities must also be the same.

				//TODO: later ...
				// A constraint is generated for each activity of the courses.

				// Check activity lengths
				footprint := []int{}         // activity durations
				var alen int = 0             // number of activities in each course
				var alists [][]ActivityIndex // collect the parallel activities
				for i, cref := range cn.Courses {
					cinfo := tt_shared_data.Ref2CourseInfo[cref]
					if i == 0 {
						alen = len(cinfo.TtActivities)
						alists = make([][]ActivityIndex, alen)
					} else if len(cinfo.TtActivities) != alen {
						//TODO: This is a data error
						clist := []string{}
						for _, cr := range cn.Courses {
							clist = append(clist, string(cr))
						}
						base.Error.Fatalf("Parallel courses have different"+
							" activities: %s\n",
							strings.Join(clist, ","))
					}
					for j, l := range cinfo.Activities {
						if i == 0 {
							footprint = append(footprint, l.Duration)
						} else if l.Duration != footprint[j] {
							//TODO: This is a data error
							clist := []string{}
							for _, cr := range cn.Courses {
								clist = append(clist, string(cr))
							}
							base.Error.Fatalf("Parallel courses have activity"+
								" mismatch: %s\n",
								strings.Join(clist, ","))
						}
						alists[j] = append(alists[j], cinfo.TtActivities[j])
					}
				}
				// alists is now a list of lists of parallel activity indexes.
				cpa := &TtParallelActivities{
					Constraint:     cn.Constraint,
					Weight:         cn.Weight,
					Courses:        cn.Courses,
					ActivityGroups: alists,
				}
				if c.IsHard() {
					tt_data.HardConstraints[ParallelCourses] = append(
						tt_data.HardConstraints[ParallelCourses], cpa)
				} else {
					tt_data.SoftConstraints[ParallelCourses] = append(
						tt_data.SoftConstraints[ParallelCourses], cpa)
				}
				continue
			}
		}

		// Collect the other constraints according to type, but unmodified,
		// separating them into hard and soft constraints,
		cname := c.CType()
		ctype, ok := cnmap[cname]
		if !ok {
			//TODO?
			panic("Unknown constraint type: " + cname)
		}
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
	for _, cinfo := range tt_shared_data.CourseInfoList {
		cref := cinfo.Id

		if len(cinfo.Activities) > 1 && !noauto_ddays[cref] {
			cn := &TtDaysBetween{
				Constraint:           C_GENERAL_DAYS_BETWEEN,
				Weight:               auto_weight,
				Course:               cref,
				DaysBetween:          1,
				ConsecutiveIfSameDay: auto_consec,
			}
			cn.ActivityLists = tt_shared_data.days_between_activities(cn)
			if auto_weight == base.MAXWEIGHT || auto_consec {
				dd_hard = append(dd_hard, cn)
			} else {
				dd_soft = append(dd_soft, cn)
			}
		}
	}
	// Now add these as new constraints to the constraint map
	tt_data.HardConstraints[MinDaysBetween] = dd_hard
	tt_data.SoftConstraints[MinDaysBetween] = dd_soft
}

// Convert a `TtDaysBetween` constraint to be based on activities.
func (tt_shared_data *TtSharedData) days_between_activities(
	constraint *TtDaysBetween,
) [][]ActivityIndex {
	allist := [][]ActivityIndex{}
	cref := constraint.Course
	cinfo := tt_shared_data.Ref2CourseInfo[cref]
	fixeds := []ActivityIndex{}
	unfixeds := []ActivityIndex{}
	for i, l := range cinfo.Activities {
		if l.Fixed {
			fixeds = append(fixeds, cinfo.TtActivities[i])
		} else {
			unfixeds = append(unfixeds, cinfo.TtActivities[i])
		}
	}

	if len(unfixeds) == 0 || (len(fixeds) == 0 && len(unfixeds) == 1) {
		// No constraints necessary
		//TODO
		base.Warning.Printf("Ignoring superfluous DaysBetween constraint on"+
			" course:\n  -- %s", tt_shared_data.View(cinfo))
		return allist
	}
	// Collect the activity groups to which the constraint is to be applied
	aidlists := [][]ActivityIndex{}
	if len(fixeds) <= 1 {
		// At most 1 fixed activity, so all activities are relevant
		aidlists = append(aidlists, cinfo.TtActivities)
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
			if len(alist) > tt_shared_data.NDays {
				//TODO
				base.Warning.Printf("Course has too many activities for"+
					"DifferentDays constraint:\n  -- %s\n",
					tt_shared_data.View(cinfo))
				continue
			}
			allist = append(allist, alist)
		}
	}
	return allist
}

// Construct the activity relationships for a `DaysBetweenJoin` constraint.
func (tt_shared_data *TtSharedData) days_between_join_activities(
	constraint *base.DaysBetweenJoin,
) [][]ActivityIndex {
	c1 := tt_shared_data.Ref2CourseInfo[constraint.Course1]
	c2 := tt_shared_data.Ref2CourseInfo[constraint.Course2]
	allist := [][]ActivityIndex{}
	for i1, l1 := range c1.Activities {
		for i2, l2 := range c2.Activities {
			if l1.Fixed && l2.Fixed {
				// both fixed => no constraint
				continue
			}
			allist = append(allist, []ActivityIndex{
				c1.TtActivities[i1], c2.TtActivities[i2]})
		}
	}
	return allist
}
