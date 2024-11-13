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
		newdb.RoomGroups = append(newdb.RoomGroups,
			&base.RoomGroup{
				Id:    e.Id,
				Tag:   e.Tag,
				Name:  e.Name,
				Rooms: e.Rooms,
			})
	}
	for _, e := range db.RoomChoiceGroups {
		newdb.RoomChoiceGroups = append(newdb.RoomChoiceGroups,
			&base.RoomChoiceGroup{
				Id:    e.Id,
				Tag:   e.Tag,
				Name:  e.Name,
				Rooms: e.Rooms,
			})
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
		celement.Divisions = []base.Division{}
		for _, div := range e.Divisions {
			celement.Divisions = append(celement.Divisions, base.Division{
				Name:   div.Name,
				Groups: div.Groups,
			})
		}
		newdb.Classes = append(newdb.Classes, celement)
	}
	newdb.Groups = groups

	for _, e := range db.Courses {
		newdb.Courses = append(newdb.Courses, &base.Course{
			Id:       e.Id,
			Subject:  e.Subject,
			Groups:   e.Groups,
			Teachers: e.Teachers,
			Room:     e.Room,
		})
	}
	for _, e := range db.SuperCourses {
		newdb.SuperCourses = append(newdb.SuperCourses, &base.SuperCourse{
			Id:      e.Id,
			Subject: e.Subject,
		})
	}
	for _, e := range db.SubCourses {
		newdb.SubCourses = append(newdb.SubCourses, &base.SubCourse{
			Id:          e.Id,
			SuperCourse: e.SuperCourse,
			Subject:     e.Subject,
			Groups:      e.Groups,
			Teachers:    e.Teachers,
			Room:        e.Room,
		})
	}
	for _, e := range db.Lessons {
		newdb.Lessons = append(newdb.Lessons, &base.Lesson{
			Id:       e.Id,
			Course:   e.Course,
			Duration: e.Duration,
			Day:      e.Day,
			Hour:     e.Hour,
			Fixed:    e.Fixed,
			Rooms:    e.Rooms,
		})
	}

	newdb.Constraints = db.Constraints

	return newdb
}
