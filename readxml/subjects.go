package readxml

import (
	"W365toFET/base"
	"slices"
	"strings"
)

func (cdata *conversionData) readSubjects() {
	slices.SortFunc(cdata.xmlin.Subjects, func(a, b Subject) int {
		if a.ListPosition < b.ListPosition {
			return -1
		}
		return 1
	})
	for i := 0; i < len(cdata.xmlin.Subjects); i++ {
		n := &cdata.xmlin.Subjects[i]
		e := cdata.db.NewSubject(n.Id)
		e.Name = n.Name
		e.Tag = n.Shortcut
		cdata.subjectTags[e.Tag] = e.Id
	}
}

func courseSubject(
	cdata *conversionData,
	srefs []Ref,
) Ref {
	//
	// Deal with the Subjects field of a Course – W365 allows multiple
	// subjects.
	// The base db expects one and only one subject (in the Subject field).
	// If there are multiple subjects in the input, these will be converted
	// to a single "composite" subject, using all the subject tags.
	// Repeated use of the same subject list will reuse the created subject.
	//
	db := cdata.db

	msg := "Course %s:\n  Not a Subject: %s\n"
	var subject Ref
	if len(srefs) == 1 {
		wsid := srefs[0]
		_, ok := cdata.subjectMap[wsid]
		if !ok {
			base.Error.Fatalf(msg, courseId, wsid)
		}
		subject = wsid
	} else if len(srefs) > 1 {
		// Make a subject name
		sklist := []string{}
		for _, wsid := range srefs {
			// Need Tag/Shortcut field
			s, ok := db.subjectMap[wsid]
			if ok {
				sklist = append(sklist, s.Tag)
			} else {
				base.Error.Fatalf(msg, courseId, wsid)
			}
		}
		sktag := strings.Join(sklist, ",")
		wsid, ok := cdata.subjectTags[sktag]
		if ok {
			// The name has already been used.
			subject = wsid
		} else {
			// Need a new Subject.
			subject = cdata.makeNewSubject(sktag, "Compound Subject")
			cdata.subjectTags[sktag] = subject
		}
	} else {
		base.Error.Fatalf("Course/SubCourse has no subject: %s\n", courseId)
	}
	return subject
}

func (cdata *conversionData) makeNewSubject(
	tag string,
	name string,
) base.Ref {
	s := cdata.db.NewSubject("")
	s.Tag = tag
	s.Name = name
	return s.Id
}
