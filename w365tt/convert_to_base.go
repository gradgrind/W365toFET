package w365tt

import "W365toFET/base"

func (db *DbTopLevel) ConvertToBase() *base.DbTopLevel {
	newdb := &base.DbTopLevel{}
	elements := map[Ref]any{}

	newdb.Info = base.Info(db.Info)
	for _, e := range db.Days {
		newdb.Days = append(newdb.Days, &base.Day{
			Id:   e.Id,
			Tag:  e.Tag,
			Name: e.Name,
		})
	}
	for _, e := range db.Hours {
		newdb.Hours = append(newdb.Hours, &base.Hour{
			Id:    e.Id,
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
			Id:               e.Id,
			Tag:              e.Tag,
			Name:             e.Name,
			Firstname:        e.Firstname,
			NotAvailable:     tsl,
			MinLessonsPerDay: e.MinLessonsPerDay,
			MaxLessonsPerDay: e.MaxLessonsPerDay,
			MaxDays:          e.MaxDays,
			MaxGapsPerDay:    e.MaxGapsPerDay,
			MaxGapsPerWeek:   e.MaxGapsPerWeek,
			MaxAfternoons:    e.MaxAfternoons,
			LunchBreak:       e.LunchBreak,
		})
	}
	for _, e := range db.Subjects {
		newdb.Subjects = append(newdb.Subjects, &base.Subject{
			Id:   e.Id,
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
			Id:           e.Id,
			Tag:          e.Tag,
			Name:         e.Name,
			NotAvailable: tsl,
		}
		elements[e.Id] = r
		newdb.Rooms = append(newdb.Rooms, r)
	}
	for _, e := range db.RoomGroups {
		rl := []*base.Room{}
		for _, r := range e.Rooms {
			rl = append(rl, elements[r].(*base.Room))
		}
		newdb.RoomGroups = append(newdb.RoomGroups,
			&base.RoomGroup{
				Id:    e.Id,
				Tag:   e.Tag,
				Name:  e.Name,
				Rooms: rl,
			})
	}
	for _, e := range db.RoomChoiceGroups {
		rcg := &base.RoomChoiceGroup{
			Id:   e.Id,
			Tag:  e.Tag,
			Name: e.Name,
		}
		rl := []*base.Room{}
		for _, r := range e.Rooms {
			r2 := elements[r].(*base.Room)
			rl = append(rl, r2)
		}
		rcg.Rooms = rl
		newdb.RoomChoiceGroups = append(newdb.RoomChoiceGroups, rcg)
	}
	groups := []*base.Group{}
	for _, e := range db.Classes {
		tsl := []base.TimeSlot{}
		for _, ts := range e.NotAvailable {
			tsl = append(tsl, base.TimeSlot(ts))
		}
		celement := &base.Class{
			Id:           e.Id,
			Tag:          e.Tag,
			Year:         e.Year,
			Letter:       e.Letter,
			Name:         e.Name,
			NotAvailable: tsl,
			//Divisions: e.Divisions,
			MinLessonsPerDay: e.MinLessonsPerDay,
			MaxLessonsPerDay: e.MaxLessonsPerDay,
			MaxGapsPerDay:    e.MaxGapsPerDay,
			MaxGapsPerWeek:   e.MaxGapsPerWeek,
			MaxAfternoons:    e.MaxAfternoons,
			LunchBreak:       e.LunchBreak,
			ForceFirstHour:   e.ForceFirstHour,
		}
		divs := []*base.Division{}
		for _, div := range e.Divisions {
			delement := &base.Division{
				Name: div.Name,
				//Groups: e.Groups,
			}
			glist := []*base.Group{}
			for _, gref := range div.Groups {
				g := db.Elements[gref].(*Group)
				gelement := &base.Group{
					Id:       gref,
					Tag:      g.Tag,
					Class:    celement,
					Division: delement,
				}
				glist = append(glist, gelement)
				groups = append(groups, gelement)
			}
			delement.Groups = glist
		}
		celement.Divisions = divs
		newdb.Classes = append(newdb.Classes, celement)
	}
	newdb.Groups = groups

	//TODO

	// Courses
	// SuperCourses
	// SubCourses
	// Lessons
	// Constraints

	//

	return newdb
}
