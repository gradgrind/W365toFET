package w365tt

import "W365toFET/base"

func (db *DbTopLevel) readSubjects(newdb *base.DbTopLevel) {
	db.SubjectMap = map[Ref]string{}
	db.SubjectTags = map[string]Ref{}
	for _, e := range db.Subjects {
		// Perform some checks and add to the SubjectTags map.
		_, nok := db.SubjectTags[e.Tag]
		if nok {
			base.Error.Fatalf("Subject Tag (Shortcut) defined twice: %s\n",
				e.Tag)
		}
		db.SubjectTags[e.Tag] = e.Id
		//Copy data to base db.
		n := newdb.NewSubject(e.Id)
		n.Tag = e.Tag
		n.Name = e.Name
		db.SubjectMap[e.Id] = e.Tag
	}
}

func (db *DbTopLevel) makeNewSubject(
	newdb *base.DbTopLevel,
	tag string,
	name string,
) base.Ref {
	s := newdb.NewSubject("")
	s.Tag = tag
	s.Name = name
	// Note that this subject is not needed in db.SubjectMap
	db.SubjectTags[tag] = s.Id
	return s.Id
}
