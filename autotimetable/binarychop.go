package autotimetable

//TODO: These are just sketches ... not used yet.

//TODO: To assist in reporting the difficult constraints, I suggest
// maintaining a map (or list) of lists where the activation state of
// each constraint is registered, once the test for a constraint type
// are done.

// TODO: Another attempt at binary (?) search patterns ...
// Return a "working" subset
// TODO: Do I still need SearchInfo?
func bs2(
	instance *TtInstance,
	constraint int,
	index0 int,
	number int,
	tag string,
) *TtInstance {
	f := cfmap[constraint]
	if number == 1 {
		inst := newInstance(instance, tag)
		//??
		f(inst, index0, true)

		//TODO ...
		// run trial, waiting ...
		// -> completed

		//TODO: inst.State may be wrong here, because that might be
		// set later in tick loop ...

		switch inst.State {
		case 1:
			// Success: Update "enabled" matrix
			inst.ConstraintEnableMatrix[constraint][index0] = true
			return inst //???
		case 5:
			// Process cancelled
			return nil //???
		default:
			// failed, divide first half
			return instance //???
		}
	}

	// Test the first half
	h := number / 2

	inst := newInstance(instance, tag+"_0")

	i := index0
	for range h {
		f(inst, i, true)
		i++
	}

	//TODO ...
	// run trial, waiting ...
	// -> completed

	switch inst.State {
	case 1:
		// Success: Update "enabled" matrix
		i = index0
		for range h {
			inst.ConstraintEnableMatrix[constraint][i] = true
			i++
		}
	case 5:
		// Process cancelled
		return nil //???
	default:
		// Failed, divide first half
		inst = bs2(instance, constraint, index0, h, inst.TtData.Description)
		if inst == nil {
			return nil
		}
	}

	// Second half
	inst2 := newInstance(inst, instance.TtData.Description+"_1")

	i = h
	for range number - h {
		f(inst, i, true)
		i++
	}

	//TODO ...
	// run trial, waiting ...
	// -> completed

	switch inst2.State {
	case 1:
		// Success: Update "enabled" matrix
		i = h
		for range number - h {
			inst2.ConstraintEnableMatrix[constraint][i] = true
			i++
		}
	case 5:
		// Process cancelled
		return nil //???
	default:
		// Failed, divide second half
		inst2 = bs2(inst, constraint, h, number-h, inst2.TtData.Description)
	}

	return inst2 //??
}

//var TIMEOUT_1 int = 30
//var TIMEOUT_2 int = 10

//TODO: Adjust the sub-timeouts to fit in the overall timeout?

// Deal with a range of constraints of one type.
func binchop2(
	instance *TtInstance,
	constraint int,
	index0 int,
	number int,
	tag string,
) *TtInstance {
	f := cfmap[constraint]

	if number == 1 {
		// No split possible

		//TODO ...

		return instance // failed
	}

	// Run with all constraints enabled
	inst0 := newInstance(instance, tag)
	for i := index0; i < number; i++ {
		f(inst0, i, true)
	}

	// Run immediately: go fet.RunFet(inst0)
	addInstance(inst0, 0)
	// -> completed ???
	// cc = 0 -> success: cancel subroutines? set flags, return inst0
	// cc = 1 -> failed, do nothing except trigger untriggered fail process
	// cc = -1 -> cancelled, return nil

	//TODO: How to run this after a delay, DELAY_1?
	// Test first half
	h := number / 2
	inst1 := binchop2(
		instance,
		constraint,
		index0,
		h,
		tag+".0",
	)

	// Test the second half, starting from the result of the first half,
	// here no delay
	inst2 := binchop2(
		inst1,
		constraint,
		h,
		number-h,
		tag+".1",
	)

	//TODO: Wait for timeout on inst0?

	//TODO ...?
	if inst0.State == 1 {
		return inst0
	}
	if inst0.State < 0 { //??
		return nil
	}
	if inst2.State == 1 {
		return inst2
	}
	return instance
}

// TODO???
/*
func search_constraint_difficulties(
	instance *TtInstance,
	constraint int,
	index0 int,
	number int,

	//? tag string,
) {
	//TODO: It is probably not so good to start follow-ons here until their
	// path has been confirmed correct.

	//TODO: How to pass the activation lists to follow-ons?! I suppose it has
	// to be in the TtInstance.

	if number < 4 {
		// special, linear, treatment
	} else {
		h := number / 2

		{
			inst0 := newInstance(instance, tag) // TODO: tag ...

			si := &timetable.SearchInfo{
				Constraint: constraint, Index0: index0, Enabled: make([]bool, h),
			}
			for i := range h {
				si.Enabled[i] = true
			}
			inst0.SearchInfo = si

			//TODO ...
		}
		{
			inst1 := newInstance(instance, tag) // TODO: tag ...

			si := &timetable.SearchInfo{
				Constraint: constraint, Index0: index0, Enabled: make([]bool, h),
			}
			for i := h; i < number; i++ {
				si.Enabled[i] = true
			}
			inst1.SearchInfo = si

			//TODO: start it ...
		}

		// if fail: split further until succeed (or pass on unchanged)

		// if succeed:

		inst1 := newInstance(instance_0, tag) // TODO: tag ...
		for _, f := range flist[h:] {
			f(inst1)
		}
		//TODO: start it ...

		// evaluate ...
	}
}

func search_instance_succeeded(
	instance_0 *TtInstance,
) *TtInstance {
	si := instance_0.SearchInfo
	i := si.Index0
	for range len(si.Enabled) {
		si.Enabled[i] = true
		i++
	}
	si.Done |= si.Part
	if si.Done == 3 {
		//TODO: Test together.
		// There is a problem, though. What if it fails?
		// Maybe the second half of the test should only be run when the
		// first half has completed?
	}
	return nil
}

func search_instance_failed(
	instance_0 *TtInstance,
) *TtInstance {

}
*/
