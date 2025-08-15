package timetable

/* A TtState item records the placement state of all activities and resources
 * at a particular time. It can be used for saving and restoring the state.
 */

type TtState struct {
	ActivitySlots []TimeSlot
	ResourceWeeks []ActivityIndex
}

// Make a new TtState item specifying the current placement state.
func (tt_data *TtData) SaveState() TtState {
	a := append([]TimeSlot{}, tt_data.ActivitySlots...)
	r := append([]ActivityIndex{}, tt_data.ResourceWeeks...)
	return TtState{a, r}
}

// Restore from a saved state. Subsequently the saved state should be
// regarded as invalid, because the restoration doesn't clone it. Thus
// it is probable that the "saved" state will be changed when the
// current state is changed.
func (tt_data *TtData) RestoreStateMove(state TtState) {
	tt_data.ActivitySlots = state.ActivitySlots
	tt_data.ResourceWeeks = state.ResourceWeeks
}

// Restore from a saved state, leaving the the saved state as an independent
// structure which can be used again.
func (tt_data *TtData) RestoreStateClone(state TtState) {
	tt_data.ActivitySlots = append([]TimeSlot{}, state.ActivitySlots...)
	tt_data.ResourceWeeks = append([]ActivityIndex{}, state.ResourceWeeks...)
}
