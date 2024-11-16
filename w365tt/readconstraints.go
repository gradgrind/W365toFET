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

func (db *DbTopLevel) readConstraints(newdb *base.DbTopLevel) {
	for _, e := range db.Constraints {
		switch e["constraint"] {
		case "MARGIN_HOUR":
			c := newdb.NewLessonsEndsDay()
			c.Course = a2r(e["course"])
			c.Weight = a2i(e["weight"])
		}
	}
}
