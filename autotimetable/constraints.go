package autotimetable

import "W365toFET/timetable"

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
