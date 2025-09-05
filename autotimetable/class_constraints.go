package autotimetable

import (
	"W365toFET/timetable"
)

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
