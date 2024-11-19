package ttengine

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"fmt"
	"slices"
)

type Resource int // or []any (see below)
type WeekSlot []Resource

type TtCore struct {
	NDays      int
	NHours     int
	Activities []Activity
	Resources  []WeekSlot
}

// I could make just one big vector (slice) and divide it up using the
// access functions.

/*
func readDb(db *base.DbTopLevel) *TtCore {
	ttinfo := ttbase.MakeTtInfo(db)
	ndays := len(db.Days)
	nhours := len(db.Hours)
	ttls := ttinfo.TtLessons

	//TODO

	// Allocate a vector with entries for all resources: teachers, (atomic)
	// student groups and (real) rooms. It might be a good idea to leave the
	// first entry (index 0) free.
	// Each entry is a vector for the time slots in a school week. Each slot
	// can contain a reference to an activity, indicating that this time
	// slot is blocked for this resource by the given activity. The references
	// could be pointers or indexes, but the value 0 should be reserved to
	// indicate "free". There would be a special reference for slots blocked
	// by NotAvailable constraints.

	lt := len(db.Teachers)
	lr := len(db.Rooms)

	nag := 0
	for gref, ags := range ttinfo.AtomicGroups {

	}
	//TODO: lg := number of atomic groups (incl. for classes with no divisions)

	lw := ndays * nhours

	// If using a single vector for all slots:

	tt := &TtCore{
		NDays:      ndays,
		NHours:     nhours,
		Activities: make([]Activity, len(ttls)),
		Resources:  make([]WeekSlot, (lt+lr+lg)*lw+1),
		// The slots are initialized to 0 (or nil for type "any").
	}

	return tt
}
*/

func handleAtomicGroups(
	db *base.DbTopLevel,
	agmap map[base.Ref][]*ttbase.AtomicGroup,
) {
	ags := []*ttbase.AtomicGroup{}
	g2ags := map[base.Ref][]ttbase.ResourceIndex{}
	for _, cl := range db.Classes {
		for _, ag := range agmap[cl.ClassGroup] {
			ags = append(ags, ag)
			//TODO: Get the Group -> index list map
			g2ags[cl.ClassGroup] = append(g2ags[cl.ClassGroup],
				ag.Index)
			for _, gref := range ag.Groups {
				g2ags[gref] = append(g2ags[gref], ag.Index)
			}
		}
	}
	// Sort the AtomicGroups
	slices.SortFunc(ags, func(a, b *ttbase.AtomicGroup) int {
		if a.Index < b.Index {
			return -1
		}
		return 1
	})

	for _, ag := range ags {
		fmt.Printf(" :: %+v\n", ag)
	}

	//TODO: Save ags and g2ags ...
}
