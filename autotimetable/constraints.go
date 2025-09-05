package autotimetable

import "W365toFET/timetable"

/*
The idea is for each constraint to have a function to switch the constraint
on/off. The constraint type is indexed, the individual instances of a
constraint type also have indexes, so two indexes are needed to refer to
a single constraint.

The cumulative effects of the constraint switches are to be found in the
`TtData` supplied in the current `TtInstance`, which is used as the base
upon which the switch functions work.

Of course, only constraints which are actually specified in the source data
need to be tested. So it might be worth filtering the lists before starting
the test sequences (which may use a binary search method).

//TODO: This is a more general comment:
Although a binary search can be relatively efficient, this efficiency will
be reduced if more than one of the components is difficult, especially if the
difficulty arises from the combination, which they mostly do. On the other
hand, checking all combinations is not feasible (because of the enormous
number). The hope is that by offering some assistance in narrowing down the
difficult constraints to particular types, and perhaps individuals within
those types, that the user can find a way to adjust the constraints to make
the construction of the timetable possible.
*/

//++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

// TODO: This probably needs to encompass ALL constraints ...
const (
	TMinLessonsPerDay int = iota
	TMaxLessonsPerDay
	TMaxAfternoons
	TMaxDays
	TLunchBreak
	TMaxGapsPerDay
	TMaxGapsPerWeek

	CMinLessonsPerDay
	CMaxLessonsPerDay
	CMaxAfternoons
	CLunchBreak
	CForceFirstHour
	CMaxGapsPerDay
	CMaxGapsPerWeek

	// ...

	LastConstraint // can be used for sizing function map, etc.
)

var cfmap [LastConstraint]func(*timetable.TtInstance, int, bool)
