package base

func (db *DbTopLevel) addConstraint(c Constraint) {
	db.Constraints = append(db.Constraints, c)
}

type LessonsEndDay struct {
	Constraint string
	Course     Ref
	Weight     int
}

func (c *LessonsEndDay) CType() string {
	return c.Constraint
}

func (db *DbTopLevel) NewLessonsEndsDay() *LessonsEndDay {
	c := &LessonsEndDay{Constraint: "LessonsEndDay"}
	db.addConstraint(c)
	return c
}
