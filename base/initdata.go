package base

import "W365toFET/w365tt"

//TODO: The conversion should be in w365tt

//TODO: Read more directly to this structure, rather than writing back
// to the input structure,

func MoveDb(db0 *w365tt.DbTopLevel) *DbTopLevel {
	db := &DbTopLevel{}

	db.Info = Info{
		Institution:        db0.Info.Institution,
		FirstAfternoonHour: db0.Info.FirstAfternoonHour,
		MiddayBreak:        db0.Info.MiddayBreak,
		Reference:          db0.Info.Reference,
	}
	for _, d := range db0.Days {
		db.Days = append(db.Days, Day{
			Id: Ref(d.Id), Name: d.Name, Tag: d.Tag})
	}

	for _, d := range db0.Hours {
		db.Hours = append(db.Hours, Hour{
			Id: Ref(d.Id), Name: d.Name, Tag: d.Tag,
			Start: d.Start, End: d.End})
	}

	for _, d := range db0.Teachers {
		ts := []TimeSlot{}
		for _, ts0 := range d.NotAvailable {
			ts = append(ts, TimeSlot{ts0.Day, ts0.Hour})
		}
		db.Teachers = append(db.Teachers, Teacher{
			Id: Ref(d.Id), Name: d.Name, Tag: d.Tag,
			Firstname: d.Firstname, NotAvailable: ts,
			MinLessonsPerDay: int(d.MinLessonsPerDay.(float64)),
			MaxLessonsPerDay: int(d.MaxLessonsPerDay.(float64)),
			MaxDays:          int(d.MaxDays.(float64)),
			MaxGapsPerDay:    int(d.MaxGapsPerDay.(float64)),
			MaxGapsPerWeek:   int(d.MaxGapsPerWeek.(float64)),
			MaxAfternoons:    int(d.MaxAfternoons.(float64)),
			LunchBreak:       d.LunchBreak,
		})
	}

	for _, d := range db0.Subjects {
		db.Subjects = append(db.Subjects, Subject{
			Id: Ref(d.Id), Name: d.Name, Tag: d.Tag})
	}

	for _, d := range db0.Rooms {
		ts := []TimeSlot{}
		for _, ts0 := range d.NotAvailable {
			ts = append(ts, TimeSlot{ts0.Day, ts0.Hour})
		}
		db.Rooms = append(db.Rooms, Room{
			Id: Ref(d.Id), Name: d.Name, Tag: d.Tag,
			NotAvailable: ts,
		})
	}

	for _, d := range db0.RoomGroups {
		refs := []Ref{}
		for _, r0 := range d.Rooms {
			refs = append(refs, Ref(r0))
		}
		db.RoomGroups = append(db.RoomGroups, RoomGroup{
			Id: Ref(d.Id), Name: d.Name, Tag: d.Tag,
			Rooms: refs,
		})
	}

	for _, d := range db0.RoomChoiceGroups {
		refs := []Ref{}
		for _, r0 := range d.Rooms {
			refs = append(refs, Ref(r0))
		}
		db.RoomChoiceGroups = append(db.RoomChoiceGroups, RoomChoiceGroup{
			Id: Ref(d.Id), Name: d.Name, Tag: d.Tag,
			Rooms: refs,
		})
	}

	for _, d := range db0.Classes {
		ts := []TimeSlot{}
		for _, ts0 := range d.NotAvailable {
			ts = append(ts, TimeSlot{ts0.Day, ts0.Hour})
		}
		divs := []Division{}
		for _, div0 := range d.Divisions {
			glist := []Ref{}
			for _, g0 := range div0.Groups {
				glist = append(glist, Ref(g0))
			}
			divs = append(divs, Division{
				Name:   d.Name,
				Groups: glist,
			})
		}
		db.Classes = append(db.Classes, Class{
			Id: Ref(d.Id), Name: d.Name, Tag: d.Tag,
			Year: d.Year, Letter: d.Letter, NotAvailable: ts,
			Divisions:        divs,
			MinLessonsPerDay: int(d.MinLessonsPerDay.(float64)),
			MaxLessonsPerDay: int(d.MaxLessonsPerDay.(float64)),
			MaxGapsPerDay:    int(d.MaxGapsPerDay.(float64)),
			MaxGapsPerWeek:   int(d.MaxGapsPerWeek.(float64)),
			MaxAfternoons:    int(d.MaxAfternoons.(float64)),
			LunchBreak:       d.LunchBreak,
			ForceFirstHour:   d.ForceFirstHour,
		})
	}

	for _, d := range db0.Groups {
		db.Groups = append(db.Groups, Group{
			Id: Ref(d.Id), Tag: d.Tag})
	}

	for _, d := range db0.Courses {
		glist := []Ref{}
		for _, g0 := range d.Groups {
			glist = append(glist, Ref(g0))
		}
		tlist := []Ref{}
		for _, t0 := range d.Teachers {
			tlist = append(tlist, Ref(t0))
		}
		db.Courses = append(db.Courses, Course{
			Id: Ref(d.Id), Subject: Ref(d.Subject),
			Groups: glist, Teachers: tlist, Room: Ref(d.Room)})
	}

	for _, d := range db0.SuperCourses {
		db.SuperCourses = append(db.SuperCourses, SuperCourse{
			Id: Ref(d.Id), Subject: Ref(d.Subject)})
	}

	for _, d := range db0.SubCourses {
		glist := []Ref{}
		for _, g0 := range d.Groups {
			glist = append(glist, Ref(g0))
		}
		tlist := []Ref{}
		for _, t0 := range d.Teachers {
			tlist = append(tlist, Ref(t0))
		}
		db.SubCourses = append(db.SubCourses, SubCourse{
			Id: Ref(d.Id), Subject: Ref(d.Subject),
			SuperCourse: Ref(d.SuperCourse),
			Groups:      glist, Teachers: tlist, Room: Ref(d.Room)})
	}

	for _, d := range db0.Lessons {
		rlist := []Ref{}
		for _, r0 := range d.Rooms {
			rlist = append(rlist, Ref(r0))
		}
		db.Lessons = append(db.Lessons, Lesson{
			Id: Ref(d.Id), Course: Ref(d.Course),
			Duration: d.Duration, Day: d.Day, Hour: d.Hour,
			Fixed: d.Fixed, Rooms: rlist,
		})
	}

	db.Constraints = db0.Constraints

	return db
}
