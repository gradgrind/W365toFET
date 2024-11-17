package readxml

import (
	"W365toFET/base"
	"log"
	"slices"
)

//TODO: The Course field can be a Course or a SuperCourse.
// Only lessons placed in a schedule are present, and only one schedule
// should be used. Perhaps regard all schedulued lessons as fixed regardless
// of the flag?
// Non-scheduled lessons must be discovered (from the Courses and SuperCourses)
// and then added.
// Bear in mind that lessons have no length field, so multi-hour lessons are
// made up of single lessons. It might be worth replacing these by real
// multi-hour lessons? Of course, they must be reverted to W365 form to pass
// them back , but this should be possible ...

func (cdata *conversionData) makeLessons(scheduled []Ref) {
	// Collect scheduled lessons.
	schedmap := map[Ref]bool{}
	for _, ref := range scheduled {
		schedmap[ref] = true
	}
	lessons := map[Ref][]*Lesson{}
	for i := 0; i < len(cdata.xmlin.Lessons); i++ {
		n := &cdata.xmlin.Lessons[i]
		_, ok := schedmap[n.Id]
		if !ok {
			continue
		}

		// Ignore Lessons with no course (they belong to Epochenplan?)
		cid := n.Course
		if cid == "" {
			continue
		}
		e, ok := cdata.db.Elements[cid]
		if ok {
			_, ok = e.(*base.Course)
			if !ok {
				_, ok = e.(*base.SuperCourse)
				if !ok {
					base.Error.Fatalf("Lesson %s has invalid Course\n", n.Id)
				}
			}
			lessons[cid] = append(lessons[cid], n)
			continue
		}
		base.Error.Fatalf("Lesson %s has unknown Course\n", n.Id)
	}

	// Generate base.Lessons for the courses, taking into account the
	// already scheduled lessons.
	for cref, llens := range cdata.courseLessons {
		// Regard all supplied Lessons as fixed? If not, the others should
		// perhaps be ignored. The state of development of W365 on which
		// this model is based would require new lessons anyway.
		// Sort the lesson times.
		llist := lessons[cref]
		slices.SortFunc(llist, func(a, b *Lesson) int {
			if a.Day < b.Day || (a.Day == b.Day && a.Hour < b.Hour) {
				return -1
			}
			return 1
		})

		// Sort the lengths and deal with the largest first.
		slices.Sort(llens)
		i := len(llens)
		for i > 0 {
			i--
			llen := llens[i]
			day := -1
			hour := -1
			for j := 0; j < len(llist); j++ {
				ll := llist[j]
				d := ll.Day
				h := ll.Hour
				for t := 1; t < llen; t++ {
					if j+t >= len(llist) {
						goto add_lesson
					}
					llx := llist[j+t]
					if llx.Day != d || llx.Hour != h+t {
						goto next
					}
				}
				// found
				day = d
				hour = h
				// remove used Lessons
				llist = slices.Delete(llist, j, j+llen)
				break
			next:
			}
		add_lesson:
			l := cdata.db.NewLesson("")
			l.Course = cref
			l.Duration = llen
			l.Day = day
			l.Hour = hour
			l.Fixed = day >= 0
			l.Rooms = []Ref{}
		}
		if len(llist) != 0 {
			log.Fatalf("*ERROR* Didn't consume all lessons in course %s\n",
				cref)
		}
	}
}
