package w365tt

import (
	"W365toFET/base"
)

func a2r(r any) base.Ref {
	return base.Ref(r.(string))
}

func a2i(i any) int {
	return int(i.(float64))
}

func a2rr(rr any) []base.Ref {
	rlist := []Ref{}
	for _, r := range rr.([]any) {
		rlist = append(rlist, a2r(r))
	}
	return rlist
}

func (db *DbTopLevel) readConstraints(newdb *base.DbTopLevel) {
	for _, e := range db.Constraints {
		switch e["constraint"] {
		case "MARGIN_HOUR":
			c := newdb.NewLessonsEndDay()
			c.Weight = a2i(e["weight"])
			c.Course = a2r(e["course"])
		case "BEFORE_AFTER_HOUR":
			c := newdb.NewBeforeAfterHour()
			c.Weight = a2i(e["weight"])
			c.Courses = a2rr(e["courses"])
			c.After = e["after"].(bool)
			c.Hour = a2i(e["hour"])
		}
	}
}
