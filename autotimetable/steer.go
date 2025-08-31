package autotimetable

import (
	"W365toFET/base"
	"W365toFET/fet"
	"W365toFET/timetable"
	"os"
	"runtime"
	"time"
)

/*
A `TtRunData` structure is constructed to manage the data being used by the
timetable engine. Each trial run has a `TtInstance` structure to manage the
data specific to this instance.

A goroutine is used to run `checkProgress`, which monitors the state of all
active runs by querying them every second, updating their `TtInstance`
structures when progress is made. It is also possible to register timed
handlers, e.g. a timeout.

The `TtInstance` structures have an `Abort` function, allowing a run to be
terminated before it stops naturally.

Each trial run should have a success path and a failure path. One of these
paths can be started preemptively, perhaps after a certain delay, if
processor units are available. When a trial is resolved, it should be able
to cancel any of the trials on the now invalidated path. Cancellation of a
trial would need to specify which of the paths is to be taken (or none!).

Would it be sensible to limit preemptive spawning of new instances, either by
limiting the total number of active Instances allowed or by setting a minimum
path start delay? The latter alone might not be enough, though.

Perhaps the main goroutine, the one which starts the ball rolling, could
act as a sort of general controller, receiving signals on a channel to abort
runs, determine when the algorithm has finished, etc.
*/

// TODO: At present this only supports a FET back-end. Perhaps a choice should
// be possible ...
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

// TODO: It may well be desirable to be able to override this
var MAXPROCESSES int = runtime.NumCPU() //TODO: use this?

// TODO
// The `success` function is called when the instance completes successfully,
// or after its delay, whichever is sooner. The `failure` function is called
// when the instance fails, or after its delay, whichever is sooner. The
// `other` functions are called after their delays, unless the instance has
// already completed.

// TODO: see runTtEngine
func newInstance(
	rundata *timetable.TtRunData,
	success timetable.TtChainedFunc,
	failure timetable.TtChainedFunc,
	others ...timetable.TtChainedFunc,
) {
	//TODO: danger of race condition here with run counter because any
	// instance could start a new one ... This function should perhaps
	// be handled in the main goroutine, triggered by a channel signal.
	// OR maybe the check-progress goroutine is better? I could pass it
	// a partly initialized `*timetable.TtInstance`.
	// Perhaps the check-progress loop could be in the main goroutine ...
	counter := rundata.RunCounter
	instance := &timetable.TtInstance{
		Id:          counter,
		SuccessPath: success,
		FailurePath: failure,
		OtherPath:   others,
	}

	rundata.Instances = append(rundata.Instances, instance)
	rundata.RunCounter++

	rundata.Active[counter] = struct{}{}
	// Note that the instance already counts as "active" although it hasn't
	// been initialized yet.

	//TODO: The timetable "engine" should be replaceable.
	fet.NewFet(rundata, instance)
}

func startInstance(
	rundata *timetable.TtRunData,
	instance *timetable.TtInstance,
) {
	index := len(rundata.Instances)
	instance.Id = index
	rundata.Instances = append(rundata.Instances, instance)
	rundata.Active[index] = struct{}{}
	// Note that the instance already counts as "active" although it hasn't
	// been initialized yet.

	//TODO: The timetable "engine" should be replaceable.
	fet.NewFet(rundata, instance)
}

func SteerGeneration(tt_data_0 *timetable.TtData, stempath string) {

	// `stempath` provides the path to the source file, including the stem
	// (without file-type extension) of the file name. A new working directory
	// will be created in the same directory.
	workingdir := stempath + "_fet"
	os.RemoveAll(workingdir)
	err := os.Mkdir(workingdir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	stop := make(chan bool)
	new_instance := make(chan *timetable.TtInstance)

	rundata := &timetable.TtRunData{
		TtData_0:    tt_data_0,
		TtData:      tt_data_0,
		WorkingDir:  workingdir,
		NewInstance: new_instance,
		Stop:        stop,
		//RunCounter: 0,
		Active: map[int]struct{}{},
	}

	//TODO: A run with all constraints enabled, no timeout
	runTtEngine(rundata, 0)

	// From now use modified TtData
	// Copy original DbTopLevel (shallow copy only!)
	db0 := tt_data_0.Db
	db_1 := *db0
	db := &db_1

	// Copy original TtData (shallow copy only!)
	tt_data_1 := *tt_data_0
	tt_data := &tt_data_1
	tt_data.Db = db

	// Remove constraints
	tt_data.Constraints = map[string][]any{}
	tt_data.MinDaysBetweenLessons = nil
	tt_data.ParallelLessons = nil
	tt_data.WITHOUT_ROOM_PLACEMENTS = true

	rundata.TtData = tt_data

	// First run with no constraints except the hard-blocked time slots and
	// the fixed activities.
	// Regenerate the teachers list without constraints
	new_teachers := make([]*base.Teacher, len(db.Teachers))
	for i, t0p := range db.Teachers {
		t := *t0p
		t.MinLessonsPerDay = -1 // unconstrained
		t.MaxLessonsPerDay = -1 // unconstrained
		t.MaxDays = -1          // unconstrained
		t.MaxGapsPerDay = -1    // unconstrained
		t.MaxGapsPerWeek = -1   // unconstrained
		t.MaxAfternoons = -1    // unconstrained
		t.LunchBreak = false
		new_teachers[i] = &t
	}
	db.Teachers = new_teachers
	// Regenerate the classes list without constraints
	new_classes := make([]*base.Class, len(db.Classes))
	for i, c0p := range db.Classes {
		c := *c0p
		c.MinLessonsPerDay = -1 // unconstrained
		c.MaxLessonsPerDay = -1 // unconstrained
		c.MaxGapsPerDay = -1    // unconstrained
		c.MaxGapsPerWeek = -1   // unconstrained
		c.MaxAfternoons = -1    // unconstrained
		c.LunchBreak = false
		c.ForceFirstHour = false
		new_classes[i] = &c
	}
	db.Classes = new_classes

	//***********************************

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			//fmt.Println("checkProgress done!")
			//TODO: tidying up?
			return
		case new_instance := <-new_instance:
			runInstance(new_instance)
		case <-ticker.C:
			// Update the progress records of the currently active
			// subprocesses.
			for i := range rundata.Active {
				instance := rundata.Instances[i]
				h := instance.UpdateHandler
				if h != nil {
					// The handler should only be set when the the process is
					// fully running
					h(instance)
				}
			}
		}
	}

	//***********************************

	go checkProgress(stop, rundata)

	//TODO... ???

	runTtEngine(rundata, 30)

	/* ???

	// Add the pseudo activities due to the NotAvailable lists of classes,
	// teachers and rooms.
	tt_data.BlockResources()

	TODO?
	// Get preliminary constraint info – needed for the call to addActivity
	//ttinfo.processConstraints()

	TODO?
	// Add the remaining Activity information
	//ttinfo.addActivityInfo(t2tt, r2tt, g2ags)

	*/
	stop <- true
}

// TODO: At what stage should the goroutine be started, avoid race conditions
// with RunCounter, Active handler, etc.
// TODO: Take available CPUs into account.
func runTtEngine(rundata *timetable.TtRunData, timeout int) {
	//TODO: timeout

	instance := &timetable.TtInstance{}
	rundata.Instances = append(rundata.Instances, instance)
	rundata.Active[rundata.RunCounter] = struct{}{}
	// Note that the instance already counts as "active" although it hasn't
	// been initialized yet.

	//TODO: The timetable "engine" should be replaceable.
	fet.NewFet(rundata, instance)

	// Update `RunCounter` AFTER call to NewFet so that this can access the
	// appropriate `RunCounter` value directly
	rundata.RunCounter++
}

// Keep track of the subprocesses: their progress and completion state.
// It should be possible to register handlers for particular times and to
// issue cancellations. Timeouts should be possible.
// The number of active processes should be recorded, so that a limit can
// be set.
func checkProgress(stop chan bool, rundata *timetable.TtRunData) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			//fmt.Println("checkProgress done!")
			return
		case <-ticker.C:
			// Update the progress records of the currently active
			// subprocesses.
			for i := range rundata.Active {
				instance := rundata.Instances[i]
				h := instance.UpdateHandler
				if h != nil {
					// The handler should only be set when the the process is
					// fully running
					h(instance)
				}
			}
		}
	}
}

// TODO
func EndInstance(
	instance *timetable.TtInstance,
	successPath bool,
	failurePath bool,
) {

}
