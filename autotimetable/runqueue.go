package autotimetable

import (
	"W365toFET/base"
	"W365toFET/timetable"
	"fmt"
)

type RunQueue struct {
	Queue      []*TtInstance
	Active     map[*TtInstance]struct{}
	MaxRunning int
	Next       int
}

func (rq *RunQueue) add(instance *TtInstance) {
	if rq.Next >= 100 {
		// Reclaim space
		vec2 := rq.Queue[rq.Next:]
		n := len(vec2)
		copy(rq.Queue, vec2)
		rq.Queue = rq.Queue[:n]
		rq.Next = 0
	}
	instance.ProcessingState = -1 // not started yet
	rq.Queue = append(rq.Queue, instance)
	//base.Message.Printf("(TODO) [%d] Queue %s\n",
	//	Ticks, instance.TtData.Description)
}

func (rq *RunQueue) update_instances() {
	// First increment the ticks of active instances.
	for instance := range rq.Active {
		ttdata := instance.TtData
		//base.Message.Printf("(TODO) [%d] ? ACTIVE (%d / %d): %s\n",
		//	Ticks, ttdata.State, instance.ProcessingState, ttdata.Description)
		if ttdata.State != 0 && instance.ProcessingState < 2 {
			// This should only be possible after the call to
			// `timetable.BACKEND.Tick` below.
			panic(fmt.Sprintf("Bug, State = %d", ttdata.State))
		}
		if ttdata.State == 0 {
			ttdata.Ticks++
			// Among other things, update the state:
			timetable.BACKEND.Tick(ttdata)
		} else if instance.ProcessingState < 2 {
			// This should only be possible after the call to
			// `timetable.BACKEND.Tick`.
			panic(fmt.Sprintf("Bug, State = %d", ttdata.State))
		}

		if instance.ProcessingState == 3 {
			// Await completion of the goroutine
			continue
		}

		switch ttdata.State {
		case 0: // running, not finished
			if ttdata.Progress == 100 {
				continue // the state will be changed next time round
			}
			// Check for timeout or getting "stuck"
			t := instance.Timeout
			if t == 0 {
				// Check for lack of progress when there is no timeout
				if ttdata.LastTime < LAST_TIME_0 && ttdata.Ticks >= LAST_TIME_1 {
					// Stop instance
					abort_instance(instance)
				}
				continue
			}

			limit := (ttdata.Ticks * 100) / t
			if ttdata.Progress < limit {
				// Progress is too slow ...
				if ttdata.Progress*2 > limit {
					continue
				}
				base.Message.Printf("(TODO) [%d] Trap %s @ %d (%d): %d\n",
					Ticks, ttdata.Description, ttdata.Ticks, ttdata.Progress,
					len(instance.Constraints))
				abort_instance(instance)
			}

		case 1: // completed successfully
			base.Message.Printf("(TODO) [%d] <<+ %s @ %d\n",
				Ticks, ttdata.Description, ttdata.Ticks)
			instance.ProcessingState = 1

		default: // completed unsuccessfully
			base.Message.Printf("(TODO) [%d] <<- %s @ %d\n",
				Ticks, ttdata.Description, ttdata.Ticks)
			instance.ProcessingState = 2
		}
	}
}

func (rq *RunQueue) update_queue() int {
	// Try to start queued instances
	running := 0
	for instance := range rq.Active {
		if instance.TtData.State != 0 {
			delete(rq.Active, instance)
			if !DEBUG {
				timetable.BACKEND.Clear(instance.TtData)
			}
			continue
		}
		if instance.ProcessingState == 0 || instance.ProcessingState == 3 {
			running++
		}
	}
	for rq.Next < len(rq.Queue) && running < rq.MaxRunning {
		instance := rq.Queue[rq.Next]
		rq.Queue[rq.Next] = nil
		rq.Next++

		if instance.ProcessingState < 0 {
			instance.ProcessingState = 0 // indicate started/running
			rq.Active[instance] = struct{}{}
			running++
		} else {
			if instance.ProcessingState != 3 {
				panic("Bug")
			}
			// Cancelled before starting, skip it
			continue
		}

		base.Message.Printf("(TODO) [%d] >> %s {%d}\n",
			Ticks, instance.TtData.Description, instance.Timeout)
		timetable.BACKEND.Run(instance.TtData, TESTING)
	}
	return len(rq.Active)
}
