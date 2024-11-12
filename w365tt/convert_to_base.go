package w365tt

import "W365toFET/base"

func (db *DbTopLevel) ConvertToBase() *base.DbTopLevel {
	newdb := &base.DbTopLevel{}
	elements := map[Ref]any{}

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
			Id:           base.Ref(e.Id),
			Tag:          e.Tag,
			Name:         e.Name,
			Firstname:    e.Firstname,
			NotAvailable: tsl,
			/* TODO++
			MinLessonsPerDay: int(e.MinLessonsPerDay.(float64)),
			MaxLessonsPerDay: int(e.MaxLessonsPerDay.(float64)),
			MaxDays:          int(e.MaxDays.(float64)),
			MaxGapsPerDay:    int(e.MaxDays.(float64)),
			MaxGapsPerWeek:   int(e.MaxDays.(float64)),
			MaxAfternoons:    int(e.MaxDays.(float64)),
			*/
			LunchBreak: e.LunchBreak,
		})
	}
	for _, e := range db.Subjects {
		newdb.Subjects = append(newdb.Subjects, &base.Subject{
			Id:   base.Ref(e.Id),
			Tag:  e.Tag,
			Name: e.Name,
		})
	}
	for _, e := range db.Rooms {
		tsl := []base.TimeSlot{}
		for _, ts := range e.NotAvailable {
			tsl = append(tsl, base.TimeSlot(ts))
		}
		r := &base.Room{
			Id:           base.Ref(e.Id),
			Tag:          e.Tag,
			Name:         e.Name,
			NotAvailable: tsl,
		}
		elements[e.Id] = r
		newdb.Rooms = append(newdb.Rooms, r)
	}
	for _, e := range db.RoomGroups {
		rg := &base.RoomGroup{
			Id:   base.Ref(e.Id),
			Tag:  e.Tag,
			Name: e.Name,
		}
		rl := []*base.Room{}
		for _, r := range e.Rooms {
			r2 := elements[r].(*base.Room)
			rl = append(rl, elements[r].(*base.Room))
			r2.AddReferrer(rg)
		}
		newdb.RoomGroups = append(newdb.RoomGroups,
			&base.RoomGroup{
				Id:    base.Ref(e.Id),
				Tag:   e.Tag,
				Name:  e.Name,
				Rooms: rl,
			})
	}
	for _, e := range db.RoomChoiceGroups {
		rcg := &base.RoomChoiceGroup{
			Id:   base.Ref(e.Id),
			Tag:  e.Tag,
			Name: e.Name,
		}
		rl := []*base.Room{}
		for _, r := range e.Rooms {
			r2 := elements[r].(*base.Room)
			rl = append(rl, r2)
			r2.AddReferrer(rcg)
		}
		rcg.Rooms = rl
		newdb.RoomChoiceGroups = append(newdb.RoomChoiceGroups, rcg)
	}

	//TODO

	//

	return newdb
}
