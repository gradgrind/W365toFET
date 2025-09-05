package autotimetable

import (
	"W365toFET/timetable"
)

/*
The idea is to provide for each constraint type a function to switch the
constraint on. As it must be possible to combine the switches, the `TtData`
of the input instance needs to be modified, rather than starting from
scratch with the `TtData_0`.

A further search, when a constraint type has been identified as potentially
difficult, would be for individual classes.

Of course, only constraints which are actually specified need to be activated.
At least at the individual classes level, it might be worth filtering the
list before starting the binary search.

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
// These functions enable/disable a particular class constraint

func init() {
	cfmap[CMinLessonsPerDay] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MinLessonsPerDay =
				instance.TtData_0.Db.Classes[class].MinLessonsPerDay
		} else {
			instance.TtData.Db.Classes[class].MinLessonsPerDay = -1
		}
	}
	cfmap[CMaxLessonsPerDay] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MaxLessonsPerDay =
				instance.TtData_0.Db.Classes[class].MaxLessonsPerDay
		} else {
			instance.TtData.Db.Classes[class].MaxLessonsPerDay = -1
		}
	}
	cfmap[CMaxAfternoons] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MaxAfternoons =
				instance.TtData_0.Db.Classes[class].MaxAfternoons
		} else {
			instance.TtData.Db.Classes[class].MaxAfternoons = -1
		}
	}
	cfmap[CLunchBreak] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].LunchBreak =
				instance.TtData_0.Db.Classes[class].LunchBreak
		} else {
			instance.TtData.Db.Classes[class].LunchBreak = false
		}
	}
	cfmap[CForceFirstHour] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].ForceFirstHour =
				instance.TtData_0.Db.Classes[class].ForceFirstHour
		} else {
			instance.TtData.Db.Classes[class].ForceFirstHour = false
		}
	}
	cfmap[CMaxGapsPerDay] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MaxGapsPerDay =
				instance.TtData_0.Db.Classes[class].MaxGapsPerDay
		} else {
			instance.TtData.Db.Classes[class].MaxGapsPerDay = -1
		}
	}
	cfmap[CMaxGapsPerWeek] = func(
		instance *timetable.TtInstance,
		class int,
		enable bool,
	) {
		if enable {
			instance.TtData.Db.Classes[class].MaxGapsPerWeek =
				instance.TtData_0.Db.Classes[class].MaxGapsPerWeek
		} else {
			instance.TtData.Db.Classes[class].MaxGapsPerWeek = -1
		}
	}

}

//TODO: Map constraint indexes to constraint names?
