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

func (cdata *conversionData) getCourseSubject(c *Course) Ref {
	//
	// Deal with the Subjects field of a Course – W365 allows multiple
	// subjects.
	// The base db expects one and only one subject (in the Subject field).
	// If there are multiple subjects in the input, these will be converted
	// to a single "composite" subject, using all the subject tags.
	// Repeated use of the same subject list will reuse the created subject.
	//
	if c.Subjects == "" {
		base.Error.Fatalf("In Course %s:\n  -- No Subject\n", c.Id)
	}
	slist := []Ref{}
	for _, ref := range splitRefList(c.Subjects) {
		s, ok := cdata.db.Elements[ref]
		if ok {
			_, ok := s.(*base.Subject)
			if ok {
				slist = append(slist, ref)
				continue
			}
		}
		base.Error.Fatalf("In Course %s:\n  -- Invalid Subject: %s\n",
			c.Id, ref)
	}
	if len(slist) == 1 {
		return slist[0]
	}
	sklist := []string{}
	for _, sref := range slist {
		// Need Tag/Shortcut field
		s, ok := cdata.db.Elements[sref]
		if ok {
			e, ok := s.(*base.Subject)
			if ok {
				sklist = append(sklist, e.Tag)
				continue
			}
		}
		base.Error.Fatalf("In Course %s:\n  -- Invalid Subject: %s\n",
			c.Id, sref)
	}
	sktag := strings.Join(sklist, ",")
	sref, ok := cdata.subjectTags[sktag]
	if ok {
		// The name has already been used.
		return sref
	} // Need a new Subject.
	sref = cdata.makeNewSubject(sktag, "Compound Subject")
	cdata.subjectTags[sktag] = sref
	return sref
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
