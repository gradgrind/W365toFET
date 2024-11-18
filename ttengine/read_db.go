package ttengine

import "W365toFET/base"

type Resource int // or []any (see below)
type WeekSlot []Resource

type TtBase struct {
	Resources []WeekSlot
}

// I could make just one big vector (slice) and divide it up using the
// access functions.

func readDb(db *base.DbTopLevel) *TtBase {
	tt := &TtBase{}

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

	/*
		lt := len(db.Teachers)
		lr := len(db.Rooms)
		lg := number of atomic groups (incl. for classes with no divisions)

		lw := ndays * nhours

		// If using a single vector for all slots:
		tt.Resources := make([]Resource, (lt + lr + lg) * lw + 1)
		// The slots are initialized to 0 (or nil for type "any").

	*/
	return tt
}
