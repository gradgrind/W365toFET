package readxml

import (
	"W365toFET/base"
	"log"
	"slices"
	"strconv"
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

func readLessons(
	// At first, just read in all Lessons to id mapper.
	//outdata *base.DbTopLevel,
	id2node map[Ref]interface{},
	items []Lesson,
) {
	for _, n := range items {
		addId(id2node, &n)
	}
}

// TODO
func makeLessons(
	outdata *base.DbTopLevel,
	id2node map[Ref]interface{},
	courseLessons map[Ref][]int,
	// courseLessons maps course ref -> list of lesson lengths
	scheduled []Ref,
) {
	// Collect scheduled lessons
	lessons := map[Ref][]*Lesson{} // course ref -> lesson list
	for _, sl := range scheduled {
		n, ok := id2node[sl]
		if !ok {
			log.Printf("*ERROR* Lesson in Schedule has no Definition: %s\n",
				sl)
			continue
		}
		np, ok := n.(*Lesson)
		if !ok {
			log.Printf("*ERROR* Bad Lesson in Schedule: %s\n", sl)
			continue
		}
		// Ignore Lessons with no course (they belong to Epochenplan?)
		if np.Course == "" {
			continue
		}
		if _, ok := id2node[np.Course]; !ok {
			log.Printf("*ERROR* Lesson with invalid Course: %s\n", np.Id)
			continue
		}
		lessons[np.Course] = append(lessons[np.Course], np)
	}

	// Generate base.Lessons for the courses, taking into account the
	// already scheduled lessons.
	for cref, llens := range courseLessons {
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
			lid := "#l#" + strconv.Itoa(len(outdata.Lessons))
			outdata.Lessons = append(outdata.Lessons, &base.Lesson{
				Id:       Ref(lid),
				Course:   cref,
				Duration: llen,
				Day:      day,
				Hour:     hour,
				Fixed:    day >= 0,
				Rooms:    []Ref{},
			})
		}
		if len(llist) != 0 {
			log.Fatalf("*ERROR* Didn't consume all lessons in course %s\n",
				cref)
		}
	}
}
