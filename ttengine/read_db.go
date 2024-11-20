package ttengine

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"slices"
)

type Ref = base.Ref

type ActivityIndex = int

//type Resource int // or []any (see below)
//type WeekSlot []Resource

type TtCore struct {
	NDays        int
	NHours       int
	SlotsPerWeek int
	Activities   []*Activity // first entry (index 0) free!
	Resources    []any       // pointers to Resources
	TtSlots      []ActivityIndex
}

// I could make just one big vector (slice) and divide it up using the
// access functions.

func readDb(ttinfo *ttbase.TtInfo) *TtCore {
	db := ttinfo.Db
	ndays := len(db.Days)
	nhours := len(db.Hours)

	// Allocate a vector for pointers to all Resources: teachers, (atomic)
	// student groups and (real) rooms.
	// Allocate a vector for pointers to all Activities, keeping the first
	// entry free (0 should be an invalid ActivityIndex).
	// Allocate a vector for a week of time slots for each Resource. Each
	// cell represents a timetable slot for a single resource. If it is
	// occupied – by an ActivityIndex – that indicates which Activity is
	// using the Resource at this time. A value of -1 indicates that the
	// time slot is blocked for this Resource.

	lt := len(db.Teachers)
	lr := len(db.Rooms)
	lw := ndays * nhours

	ags := []*ttbase.AtomicGroup{}
	g2ags := map[Ref][]ResourceIndex{}
	for _, cl := range db.Classes {
		for _, ag := range ttinfo.AtomicGroups[cl.ClassGroup] {
			ags = append(ags, ag)
			// Add to the Group -> index list map
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

	lg := len(ags)

	// If using a single vector for all slots:
	tt := &TtCore{
		NDays:        ndays,
		NHours:       nhours,
		SlotsPerWeek: ndays * nhours,
		// Activities:   set in "addActivities"
		Resources: make([]any, lt+lr+lg),
		TtSlots:   make([]ActivityIndex, (lt+lr+lg)*lw),
	}
	// The slice cells are initialized to 0 or nil, according to slice type.
	// Copy the AtomicGroups to the beginning of the Resources slice.
	i := 0
	for _, ag := range ags {
		tt.Resources[i] = ag
		//fmt.Printf(" :: %+v\n", ag)
		i++
	}

	t2tt := map[Ref]ResourceIndex{}
	for _, t := range db.Teachers {
		t2tt[t.Id] = i
		tt.Resources[i] = t
		i++
	}
	r2tt := map[Ref]ResourceIndex{}
	for _, r := range db.Rooms {
		r2tt[r.Id] = i
		tt.Resources[i] = r
		i++
	}

	// Add the pseudo activities due to the NotAvailable lists
	tt.addBlockers(ttinfo, t2tt, r2tt)

	// Add the Activities
	tt.addActivities(ttinfo, t2tt, r2tt, g2ags)
	return tt
}
