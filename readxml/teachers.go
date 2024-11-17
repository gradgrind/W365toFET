package readxml

import (
	"fmt"
	"slices"
)

func (cdata *conversionData) readTeachers() {
	db := cdata.db
	slices.SortFunc(cdata.xmlin.Teachers, func(a, b Teacher) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	ndays := len(db.Days)
	nhours := len(db.Hours)
	for i := 0; i < len(cdata.xmlin.Teachers); i++ {
		n := &cdata.xmlin.Teachers[i]
		e := db.NewTeacher(n.Id)
		e.Name = n.Name
		e.Tag = n.Shortcut

		notAvailable := cdata.getAbsences(n.Absences,
			fmt.Sprintf("In Teacher %s (Absences)", n.Id))
		// MaxAfternoons = 0 has a special meaning (all blocked)
		maxpm := n.MaxAfternoons
		if maxpm >= ndays {
			maxpm = -1
		}
		e.NotAvailable = handleZeroAfternoons(db, notAvailable, maxpm)
		if maxpm == 0 {
			maxpm = -1
		}

		maxdays := n.MaxDays
		if maxdays >= ndays {
			maxdays = -1
		}

		// Handle no-lunch-break flag on teacher.
		lb := cdata.withLunchBreak(n.Categories, n.Id)
		maxlpd := n.MaxLessonsPerDay
		if lb {
			if maxlpd >= nhours-1 {
				maxlpd = -1
			}
		} else if maxlpd >= nhours {
			maxlpd = -1
		}

		e.Firstname = n.Firstname
		e.MinLessonsPerDay = n.MinLessonsPerDay
		e.MaxLessonsPerDay = maxlpd
		e.MaxDays = maxdays
		e.MaxGapsPerDay = n.MaxGapsPerDay
		e.MaxGapsPerWeek = -1
		e.MaxAfternoons = maxpm
		e.LunchBreak = lb
	}
}
