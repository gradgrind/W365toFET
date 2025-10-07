package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"slices"
	"syscall"
	"time"
)

var (
	// The behaviour of the TESTING flag depends on the back-end. It might,
	// for example, use fixed seeds for random number generators so as to
	// produce reproduceable runs.
	TESTING bool
	// This approach relies on parallel processing. If there are too few real
	// processors it will be inefficient:
	MAXPROCESSES                   int
	UNCONSTRAINED_TIMEOUT_FRACTION int
	MIN_UNCONSTRAINED_TIMEOUT      int
	// Below this time a basic constraint type is considered "fast" and need
	// not be sorted for the constraint type accumulation:
	QUICK_BASIC_TIME          int
	NEXT_STAGE_TIMEOUT_FACTOR int // factor * 10
	NEXT_STAGE_TIMEOUT_MIN    int
)

func SetParameterDefault() {
	MAXPROCESSES = max(runtime.NumCPU(), 4)
	UNCONSTRAINED_TIMEOUT_FRACTION = 10
	MIN_UNCONSTRAINED_TIMEOUT = 10
	QUICK_BASIC_TIME = 5
	NEXT_STAGE_TIMEOUT_FACTOR = 15 // => 1.5
	NEXT_STAGE_TIMEOUT_MIN = 20
}

func init() {
	SetParameterDefault()
}

/*
Various strategies are used to try to achieve a – possibly imperfect –
timetable within a specified time. It is impossible to guarantee that all
constraints will be satisfied within a given time, so in order to place
all the activities within this time it may be necessary to drop some of
the constraints.

A certain degree of parallel processing is assumed – less than four processor
cores is likely to result in a very significant slowdown.

The main function (`StartGeneration`) starts a run with the fully constrained
data and a second run with all the "non-basic" constraints removed. Fixed
activity placements and blocked time-slots (for teachers, classes, and rooms)
are regarded as basic, non-negotiable.

TODO: If there are soft constraints, it may make sense to start a further run
with just the hard constraints enabled. The current state of development is
that soft constraints are completely ignored, though they are included in the
fully constrained run.

A `TtInstance` structure is constructed to manage the data for each
timetable generation run, each run having its own goroutine. Each instance
has its own individual timeout to stop it running forever.

Once these initial instances have been started, a "tick-loop" (which is
triggered every second) is entered. This monitors the progress of each active
instance and handles the actions resulting from their completion, whether
successful or not.

Should the fully constrained instance complete successfully within the
allotted time, all other instances are terminated and the result will be as
if only this instance had run.

When the unconstrained instance completes successfully, a series of further
instances is queued for running, each specifying the addition of a list of
(hard) constraints of a single type. Thus for each type of constraint an
instance is constructed. Using timeouts and binary divisions of these lists
an attempt is made to find individual "difficult" constraints, which can then
be disabled in order to get full activity placement within a reasonable time.
Parallel processing can be of some assistance here.

TODO: Should the unconstrained instance fail to complete successfully within
its allotted time, further steps may be taken to trace difficulties within the
activity collection, perhaps identifying "difficult" classes or teachers.

Once the single-constraint-type instances start delivering results, the next
stage can be started, in which the constraint types are added one after the
other to the gradually expanding base. The constraint types are added in order
of their completion in the single-constraint-type trials, so that the less
"difficult" constraints are added first.

Using parallel processing, it is possible that instances will be started
"preemptively", but then turn out to be irrelevant. To handle this, it is
possible to "cancel" an instance, including any instances that have been
started as a consequence. For example, when an instance with a list of
constraints is started, it queues (to be started later, when processors
become available) two further instances, one for each half of the list. If
the instance completes successfully before its timeout is reached, these
subsidiary instances become redundant and can be cancelled.

If the overall timeout is reached (i.e. if the fully constrained instance
has not completed successfuly yet), the "best solution so far" is sought:

If there is an all-hard-constraints-only instance and that has completed
successfuly, its result will be used (the single-constraint accumulation
processes can be cancelled in this case as they are then superfluous).

Otherwise, if the single-constraint accumulation has completed, its result
can be used. If it has not completed, the result will be the last successfully
completed instance. Diagnostic information will also be available (at least
an indication of which constraints were dropped).

If the single-constraint accumulation completes well before the overall
timeout, it may be a sign that its internal timeouts could be lengthened to
(possibly) obtain a result with fewer constraints disabled. Alternatively,
a shorter overall timeout might be considered.

If the single-constraint accumulation doesn't complete before the overall
timeout, that may indicate that a shortening of its internal timeouts
could produce a better result (testing more constraints). Alternatively,
a longer overall timeout might be considered.

There is probably no general "optimum" value for the various timeouts, that
is likely to depend on the data. But perhaps values can be found which are
frequently useful. It might be helpful to use shorter overall timeouts during
the initial phases of testing the data, to identify potential problem areas
without long processing delays. For later phases longer times may be
necessary (depending on the difficulty of the data).
*/

var Descriptions map[string]string = map[string]string{
	"COMPLETE":           "All constraints active",
	"ONLY_BLOCKED_SLOTS": "All constraints – except blocked slots – disabled",
}

func StartGeneration(tt_data_0 *timetable.TtData, TIMEOUT int) {
	tt_shared_data := tt_data_0.SharedData

	// Catch termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//TODO-- Wait group to ensure all goroutines finish before exiting
	//var wg sync.WaitGroup

	runqueue := RunQueue{
		Queue:      nil,
		Active:     map[*TtInstance]struct{}{},
		MaxRunning: MAXPROCESSES, //TODO????
		Next:       0,
	}

	// `workingdir` provides the path to a working directory which can be used
	// freely during processing. It may or may not already exist, existing
	// contents need not be preserved during processing.
	workingdir := tt_shared_data.WorkingDir

	tt_data_0.Description = "COMPLETE"

	// Provide an empty working directory.
	os.RemoveAll(workingdir)
	err := os.Mkdir(workingdir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	// Global data
	Ticks = 0
	TtData_0 = tt_data_0

	// First run: all constraints enabled.
	// On successful completion, all other instances should be stopped.
	// If it fails, just this instance should be wound up. Otherwise it
	// should run until it times out, at which point any other active
	// instances should be stopped and the "best" solution at this point
	// chosen.
	full_instance := &TtInstance{
		Tagged:  true,
		Timeout: TIMEOUT,
		TtData:  tt_data_0,
	}

	// Add to run queue
	runqueue.add(full_instance)

	//TODO: Instance without soft constraints (if any)?

	// Unconstrained instance
	tt_data := new_ttdata(tt_data_0, "ONLY_BLOCKED_SLOTS")
	null_instance := &TtInstance{
		Tagged: true,

		Timeout: max(TIMEOUT/UNCONSTRAINED_TIMEOUT_FRACTION,
			MIN_UNCONSTRAINED_TIMEOUT),

		TtData: tt_data,
		HardConstraintEnabled: setup_hard_constraint_map(
			tt_data_0.HardConstraints),
	}
	disable_all_constraints(null_instance.TtData)
	// Add to run queue
	runqueue.add(null_instance)

	// Start stage 0
	stage := 0

	// *** Ticker loop ***
	var basic_constraints []*TtInstance
	var constraint_list []*TtInstance
	var current_instance *TtInstance
	ticker := time.NewTicker(time.Second)
	//TODO? defer tidy(runqueue, ticker)
	defer ticker.Stop()

	//tickloop:
	for runqueue.update_queue() != 0 {
		select {

		case ossig := <-sigChan:
			//TODO: seek the best solution so far?
			base.Message.Printf("*** SIGNAL *** %+v", ossig)
			runqueue.disable()
			cancel_instance(full_instance)
			//cancel_instance(hard_only_instance)
			cancel_instance(null_instance)
			cancel_instance(current_instance)
			stage = -1
			continue

		case <-ticker.C:
			Ticks++
			runqueue.update_instances()
			if stage == -1 {
				continue
			}

			if Ticks == TIMEOUT {
				runqueue.disable()
				//base.Message.Printf(
				//	"(TODO) [%d] TIMEOUT\n", Ticks)
				if full_instance.ProcessingState == 0 {
					cancel_instance(full_instance)
				}
				//cancel_instance(hard_only_instance)
				cancel_instance(null_instance)

				//TODO: Can current_instance be nil here?

				if current_instance.Result == nil {
					cancel_instance(current_instance)
				}
				for _, i := range constraint_list {
					cancel_instance(i)
				}
				stage = -1
				continue
			}

			if full_instance.ProcessingState == 1 {
				// Cancel all other runs and return this instance as result.
				runqueue.disable()
				//cancel_instance(hard_only_instance)
				cancel_instance(null_instance)
				cancel_instance(current_instance)
				for _, i := range constraint_list {
					cancel_instance(i)
				}
				full_instance.Result = full_instance
				current_instance = full_instance
				stage = -1
				continue
			}

			//TODO: Special treatment if there are no constraints to add?

			if stage == 0 {
				// During stage 0 only `full_instance` and `null_instance`
				// (and perhaps hard_only_instance) are running.
				switch null_instance.ProcessingState {
				case 0:
					if null_instance.TtData.Ticks == null_instance.Timeout {
						abort_instance(null_instance)
					}
				case 1:
					// The null instance completed successfully.
					current_instance = null_instance
					// Start trials of single constraint types.
					basic_constraints = get_basic_constraints(
						null_instance, null_instance.TtData.Ticks)
					// Queue instances for running
					for _, bc := range basic_constraints {
						runqueue.add(bc)
					}
					constraint_list = slices.Clone(basic_constraints)
					base.Message.Printf("(TODO) [%d] CONSTRAINT-TYPES: %d\n",
						Ticks, len(basic_constraints))
					stage = 1
				default:
					// The null instance failed.
					stage = 10
					base.Message.Printf(
						"(TODO) [%d] Unconstrained instance failed", Ticks)

					//TODO: Seek problems in the unconstrained data.
					panic("TODO")
				}
				continue
			}

			if stage == 1 {
				// During stage 1 we are accumulating single-constraint-type
				// instances: check their states.
				inext := -1
				next_timeout := 0
				for i, instance := range constraint_list {
					if instance.ProcessingState == 1 {
						// completed successfully
						inext = i
						break
					}
				}
				if inext >= 0 {
					// Update current_instance ...

					//TODO: Clear old current_instance data?

					current_instance = constraint_list[inext]
					base.Message.Printf("+++ %s\n",
						current_instance.TtData.Description)
					next_timeout = max(
						current_instance.TtData.Ticks*NEXT_STAGE_TIMEOUT_FACTOR/10,
						NEXT_STAGE_TIMEOUT_MIN)

					constraint_list = slices.Delete(
						constraint_list, inext, inext+1)

					if len(constraint_list) == 0 {
						//TODO?
						break xxx
					}

					//TODO: renew instances
				}

				split_instances := []*TtInstance{}
				new_constraint_list := []*TtInstance{}
				for _, instance := range constraint_list {
					//TODO ...
					if instance.ProcessingState == 2 {
						// failed (or timed out): split the instance

						// Split if more than one instance in list
						if len(instance.Constraints) > 1 {
							nhalf := len(instance.Constraints) / 2
							split_instances = append(split_instances,
								new_instance(
									current_instance,
									instance.TtData.Description+"~0",
									instance.ConstraintType,
									instance.Constraints[:nhalf],
									instance.Timeout))
							split_instances = append(split_instances,
								new_instance(
									current_instance,
									instance.TtData.Description+"~1",
									instance.ConstraintType,
									instance.Constraints[nhalf:],
									instance.Timeout))
						}
					} else {
						if inext >= 0 {
							if instance.ProcessingState == 0 {
								abort_instance(instance)
							}
							instance = new_instance(
								current_instance,
								instance.TtData.Description+"+",
								instance.ConstraintType,
								instance.Constraints,
								next_timeout)
							runqueue.add(instance)
						}
						new_constraint_list = append(
							new_constraint_list, instance)
					}
				}
				constraint_list = append(new_constraint_list,
					split_instances...)

				for _, instance := range split_instances {
					runqueue.add(instance)
				}

				//TODO?

				continue
			}

			//TODO--

			// This bit handles the phase where constraint types are being
			// added step by step.

			// The basic idea is to try with all the constraints in the list
			// enabled, up to a time limit. If this fails, try recursively
			// with the first half, then add the second half recursively to
			// that result (with a newly calculated timeout). The recursion
			// stops when there is only one constraint in the list, it either
			// being added or not (if timed out).
			// Especially if there is a difficult constraint early in the
			// list, this can take a long time, with multiple failed runs.
			// By queueing the halves when the instance starts, it might be
			// possible to get some speed-up when multiple processors are
			// available, at least with favourable data.

			//TODO: A "stuck" analysis on running instances might help to
			// reduce processing time?

		}
	} // tickloop: end

	//TODO: Consider also the possibility that there may be no (or only one)
	// basic constraint types.

	result := current_instance.Result
	ttdata := result.TtData

	/*TODO++?
	// Remove temporary data for all instances except the result
	for i := 0; i < runqueue.Next; i++ {
		ttd := runqueue.Queue[i].TtData
		if ttd != ttdata {
			timetable.BACKEND.Clear(ttd)
		}
	}
	*/

	nn := 0
	nall := 0
	for i, clist := range result.HardConstraintEnabled {
		n := 0
		for _, b := range clist {
			if b {
				n++
			}
		}
		fmt.Printf("$ CONSTRAINT %d: %d / %d\n", i, n, len(clist))
		nn += n
		nall += len(clist)
	}
	fmt.Printf("$ ALL CONSTRAINTS: %d / %d\n", nn, nall)

	//TODO
	base.Message.Printf("(TODO) RESULT: %s\n", ttdata.Description)
}

// Cancelling an instance will abort it if it is running.
// The `ProcessingState` is set to 3 to indicate that a queued instance
// is not to be started. Also its subsidiary instances will be cancelled.
func cancel_instance(instance *TtInstance) *TtInstance {
	if instance != nil {
		//base.Message.Printf("CANCEL %s / %d\n",
		//	instance.TtData.Description, instance.ProcessingState)
		switch instance.ProcessingState {
		case 0:
			abort_instance(instance)
		case 3:
			return instance.Result
		}
		instance.ProcessingState = 3
		if instance.Result != nil {
			return instance.Result
		}
		// Cancel subsidiary instances
		r := cancel_instance(instance.Instance1)
		r2 := cancel_instance(instance.Instance2)
		if r == nil {
			r = r2
			if r == nil {
				r = instance.BaseInstance
			}
		}
		instance.Result = r
		return r
	}
	return nil
}

func abort_instance(instance *TtInstance) {
	if !instance.Stopped {
		timetable.BACKEND.Abort(instance.TtData)
		instance.Stopped = true
	}
}

func new_instance(
	instance_0 *TtInstance,
	descriptor string,
	constraint_type timetable.ConstraintType,
	constraint_indexes []int,
	timeout int,
) *TtInstance {
	// Copy original TtData (shallow copy only!)
	ttdata := new_ttdata(instance_0.TtData, descriptor)

	// Make a deep copy of the hard constraint matrix
	hcmat0 := instance_0.HardConstraintEnabled
	hcmat := make([][]bool, len(hcmat0))
	for i, c := range hcmat0 {
		hcmat[i] = slices.Clone(c)
	}

	// Make a new `TtInstance`
	instance := &TtInstance{
		Timeout:               timeout,
		TtData:                ttdata,
		HardConstraintEnabled: hcmat,

		// Base data for this instance:
		BaseInstance:   instance_0,
		ConstraintType: constraint_type,
		Constraints:    constraint_indexes,

		// Run time
		Stopped:   false,
		Instance1: nil,
		Instance2: nil,
		Result:    nil,
	}

	// Enable the constraints in `ttdata` and `hcmat`
	//fmt.Printf("§ENABLE %s: %v\n", constraint_type.String(), constraint_indexes)

	// Mark the constraints in the matrix
	cmap := instance.HardConstraintEnabled[constraint_type]
	for _, i := range constraint_indexes {
		cmap[i] = true
	}
	// Reconstruct the constraint list
	newlist := []any{}
	for i, c := range TtData_0.HardConstraints[constraint_type] {
		if cmap[i] {
			newlist = append(newlist, c)
		}
	}
	instance.TtData.HardConstraints[constraint_type] = newlist

	return instance
}

func new_ttdata(
	ttdata_0 *timetable.TtData,
	descriptor string,
) *timetable.TtData {
	ttdata := &timetable.TtData{
		Description:             descriptor,
		SharedData:              ttdata_0.SharedData,
		TeacherNotAvailable:     ttdata_0.TeacherNotAvailable,
		ClassNotAvailable:       ttdata_0.ClassNotAvailable,
		RoomNotAvailable:        ttdata_0.RoomNotAvailable,
		WITHOUT_ROOM_PLACEMENTS: ttdata_0.WITHOUT_ROOM_PLACEMENTS,
		State:                   0,
	}

	// Make a deeper copy of the constraints so that these can be
	// switched on or off without affecting those in the original `TtData`.

	// Make a copy of the constraints lists
	hcmap := make(map[timetable.ConstraintType][]any,
		len(ttdata_0.HardConstraints))
	for k, v := range ttdata_0.HardConstraints {
		hcmap[k] = slices.Clone(v)
	}
	ttdata.HardConstraints = hcmap

	scmap := make(map[timetable.ConstraintType][]any,
		len(ttdata_0.SoftConstraints))
	for k, v := range ttdata_0.SoftConstraints {
		scmap[k] = slices.Clone(v)
	}
	ttdata.SoftConstraints = scmap
	return ttdata
}

/*
// This will probably need some tuning. There is probably no such thing as
// an optimal algorithm, that depends very much on the data.
func timed_out(instance *TtInstance) bool {
	if instance.Timeout < 0 {
		// This allows a complete override of the timeout feature.
		// Unlike Timeout = 0 it will not try to catch runs which get
		// stuck at low completion levels.
		return false
	}
	if instance.Timeout != 0 && instance.Ticks > instance.Timeout {
		return true
	}
	delta := instance.Ticks - instance.LastTime
	if delta < 5 {
		return false
	}
	// The acceptable time should depend on Progress.
	if instance.Progress < 80 && delta*2 > instance.Progress {
		return true
	}
	if instance.Progress < 95 && delta > 400 {
		return true
	}
	return false
}
*/

//
/*
// TODO:  Is this old bit fetching file names from db.ModuleData still
// needed somehow?

	fetfile := stempath
	mapfile := stempath
	thisdir := filepath.Dir(stempath)
	moduleData := db.ModuleData
	fetData, ok := moduleData["FetData"].(map[string]string)
	if ok {
		var f string
		f, ok = fetData["FetFile"]
		if ok {
			fetfile = filepath.Join(thisdir, f)
		}
		f, ok = fetData["MapFile"]
		if ok {
			mapfile = filepath.Join(thisdir, f)
		}
	}
	fetfile += ".fet"
	mapfile += ".map"
*/
