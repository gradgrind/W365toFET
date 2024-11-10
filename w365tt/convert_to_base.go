package w365tt

import "W365toFET/base"

func (db *DbTopLevel) ConvertToBase() *base.DbTopLevel {
	newdb := &base.DbTopLevel{}

	newdb.Info = base.Info(db.Info)
	for _, e := range db.Days {
		newdb.Days = append(newdb.Days, &base.Day{
			Id:   base.Ref(e.Id),
			Tag:  e.Tag,
			Name: e.Name,
		})
	}
	for _, e := range db.Hours {
		newdb.Hours = append(newdb.Hours, &base.Hour{
			Id:    base.Ref(e.Id),
			Tag:   e.Tag,
			Name:  e.Name,
			Start: e.Start,
			End:   e.End,
		})
	}
	for _, e := range db.Teachers {
		tsl := []base.TimeSlot{}
		for _, ts := range e.NotAvailable {
			tsl = append(tsl, base.TimeSlot(ts))
		}
		newdb.Teachers = append(newdb.Teachers, &base.Teacher{
			Id:               base.Ref(e.Id),
			Tag:              e.Tag,
			Name:             e.Name,
			Firstname:        e.Firstname,
			NotAvailable:     tsl,
			MinLessonsPerDay: int(e.MinLessonsPerDay.(float64)),
			MaxLessonsPerDay: int(e.MaxLessonsPerDay.(float64)),
			MaxDays:          int(e.MaxDays.(float64)),
			MaxGapsPerDay:    int(e.MaxDays.(float64)),
			MaxGapsPerWeek:   int(e.MaxDays.(float64)),
			MaxAfternoons:    int(e.MaxDays.(float64)),
			LunchBreak:       e.LunchBreak,
		})
	}
	for _, e := range db.Subjects {
		newdb.Subjects = append(newdb.Subjects, &base.Subject{
			Id:   base.Ref(e.Id),
			Tag:  e.Tag,
			Name: e.Name,
		})
	}

	//TODO

	//

	return newdb
}
