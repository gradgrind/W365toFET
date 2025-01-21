package w365tt

import (
	"W365toFET/base"
)

func a2r(r any) Ref {
	return Ref(r.(string))
}

func a2i(i any) int {
	return int(i.(float64))
}

func a2rr(rr any) []Ref {
	rlist := []Ref{}
	for _, r := range rr.([]any) {
		rlist = append(rlist, a2r(r))
	}
	return rlist
}

func a2ii(ii any) []int {
	ilist := []int{}
	for _, i := range ii.([]any) {
		ilist = append(ilist, a2i(i))
	}
	return ilist
}

func (db *DbTopLevel) readConstraints(newdb *base.DbTopLevel) {
	for _, e := range db.Constraints {
		switch e["Constraint"] {
		case "MARGIN_HOUR":
			c := newdb.NewLessonsEndDay()
			c.Weight = a2i(e["Weight"])
			c.Course = a2r(e["Course"])
		case "BEFORE_AFTER_HOUR":
			c := newdb.NewBeforeAfterHour()
			c.Weight = a2i(e["Weight"])
			c.Courses = a2rr(e["Courses"])
			c.After = e["After"].(bool)
			c.Hour = a2i(e["Hour"])
		case "AUTOMATIC_DIFFERENT_DAYS":
			c := newdb.NewAutomaticDifferentDays()
			c.Weight = a2i(e["Weight"])
			c.ConsecutiveIfSameDay = e["ConsecutiveIfSameDay"].(bool)
		case "DAYS_BETWEEN":
			c := newdb.NewDaysBetween()
			c.Weight = a2i(e["Weight"])
			c.DayGap = a2i(e["DaysBetween"])
			c.Courses = a2rr(e["Courses"])
			c.ConsecutiveIfSameDay = e["ConsecutiveIfSameDay"].(bool)
		case "DAYS_BETWEEN_JOIN":
			c := newdb.NewDaysBetweenJoin()
			c.Weight = a2i(e["Weight"])
			c.DayGap = a2i(e["DaysBetween"])
			c.Course1 = a2r(e["Course1"])
			c.Course2 = a2r(e["Course2"])
			c.ConsecutiveIfSameDay = e["ConsecutiveIfSameDay"].(bool)
		case "MIN_HOURS_FOLLOWING":
			c := newdb.NewMinHoursFollowing()
			c.Weight = a2i(e["Weight"])
			c.Hours = a2i(e["Hours"])
			c.Course1 = a2r(e["Course1"])
			c.Course2 = a2r(e["Course2"])
		case "DOUBLE_LESSON_NOT_OVER_BREAKS":
			c := newdb.NewDoubleLessonNotOverBreaks()
			c.Weight = a2i(e["Weight"])
			c.Hours = a2ii(e["Hours"])
		case "PARALLEL_COURSES":
			c := newdb.NewParallelCourses()
			c.Weight = a2i(e["Weight"])
			c.Courses = a2rr(e["Courses"])
		default:
			base.Warning.Printf("Unrecognized constraint: %+v\n", e)
		}
	}
}
