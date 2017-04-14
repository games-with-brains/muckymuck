package fbmuck

struct mufevent {
	next *mufevent
	event string
	data inst
}

const (
	MUFEVENT_ALL	-1
	MUFEVENT_FIRST	-2
	MUFEVENT_LAST	-3
)

type event_process struct {
	player, prog ObjectID
	filters []string
	deleted bool
	fr *frame
}

var mufevent_processes *chain.Cell

func muf_event_process_free(ptr *mufevent_process) {
	mufevent_processes = ptr.next
}

/* Called when a MUF program enters EVENT_WAITFOR, to register that
 * the program is ready to process MUF events of the given ID type.
 * This duplicates the eventids list for itself, so the caller is
 * responsible for freeing the original eventids list passed.
 */
func muf_event_register_specific(player, prog ObjectID, fr *frame, eventids ...string) {
	newproc := &mufevent_process{
		player: player,
		prog: prog,
		fr: fr,
		filters: make([]string, len(eventids))
	}
	copy(newproc.filters, eventids)
	if ptr := mufevent_processes.End(); ptr == nil {
		mufevent_processes = newproc
	} else {
		ptr.next = newproc
	}
}


/* Called when a MUF program enters EVENT_WAIT, to register that
 * the program is ready to process any type of MUF events.
 */
func muf_event_register(player, prog ObjectID, fr *frame) {
	muf_event_register_specific(player, prog, fr)
}


/* Sends a "READ" event to the first foreground or preempt muf process
 * that is owned by the given player.  Returns 1 if an event was sent,
 * 0 otherwise.
 */
func muf_event_read_notify(descr int, player ObjectID, cmd string) (r bool) {
	for ptr := mufevent_processes; ptr != nil; ptr = ptr.Tail {
		if !ptr.deleted && ptr.player == player && ptr.fr != nil && ptr.fr.multitask != BACKGROUND && (cmd != "" || ptr.fr.wantsblanks) {
			muf_event_add(ptr.fr, "READ", &inst{ data: descr }, 1)
			r = true
			break
		}
	}
	return
}

/* removes the MUF program with the given PID from the EVENT_WAIT queue.
 */
func muf_event_dequeue_pid(pid int) (r int) {
	mufevent_processes.Each(func(proc interface{}) {
		if proc := proc.(event_process); !proc.deleted {
			if proc.fr.pid == pid {
				if !proc.fr.been_background {
					DB.FetchPlayer(proc.player).block = false
				}
				proc.fr.events = nil
				proc.deleted = true
				r++
			}
		}
	})
	return
}

/* Checks the MUF event queue for address references on the stack or
 * ObjectID references on the callstack
 */
func event_has_refs(program ObjectID, proc *mufevent_process) (r bool) {
	if !proc.deleted {
		if fr := proc.fr; fr != nil {
			for loop := 1; loop < fr.caller.top; loop++ {
				if fr.caller.st[loop] == program {
					r = true
					break
				}
			}
			if r == false{
				for loop := 0; loop < fr.argument.top; loop++ {
					if v, ok := fr.argument.st[loop].data.(Address); ok {
						if r = v.progref == program; r {
							break
						}
					}
				}
			}
		}
	}
	return
}

/* Deregisters a program from any instances of it in the EVENT_WAIT queue.
 * killmode values:
 *     0: kill all matching processes (Equivalent to 1)
 *     1: kill all matching MUF processes
 *     2: kill all matching foreground MUF processes
 */
func muf_event_dequeue(prog ObjectID, killmode int) (r int) {
	if killmode == 0 {
		killmode = 1
	}

	mufevent_processes.Each(func(proc interface{}) {
		switch proc := proc.(event_process) {
		case proc.deleted:
		case proc.prog != prog && !event_has_refs(prog, proc) && proc.player != prog:
		case killmode == 2 && proc.fr != nil && proc.fr.multitask == BACKGROUND:
		case killmode == 1 && proc.fr == nil:
		default:
			if proc.fr != nil {
				if !proc.fr.been_background {
					DB.FetchPlayer(proc.player).block = false
				}
				proc.fr.events = nil
				prog_clean(proc.fr)
			}
			proc.deleted = true
			r++
		}
	})
	return
}

func muf_event_pid_frame(pid int) (r *frame) {
	for ptr = mufevent_processes; ptr != nil; ptr = ptr.next {
		if !ptr.deleted && ptr.fr != nil && ptr.fr.pid == pid {
			r = ptr.fr
			break
		}
	}
	return
}

/* Returns true if the given player controls the given PID.
 */
func muf_event_controls(player ObjectID, pid int) (r bool) {
	proc := mufevent_processes
	for ; proc != nil && (proc.deleted || pid != proc.fr.pid); proc = proc.next {}
	switch {
	case !proc == nil:
	case !controls(player, proc.prog) && player != proc.player:
	default:
		r = true
	}
	return
}

/* List all processes in the EVENT_WAIT queue that the given player controls.
 * This is used by the @ps command.
 */
func muf_event_list(player ObjectID, pat string) (r int) {
	time_t rtime = time((time_t *) NULL);
	time_t etime;

	for proc := mufevent_processes; proc != nil; proc = proc.next {
		if !proc.deleted {
			var pcnt float64
			if proc.fr != nil {
				etime = rtime - proc.fr.started
				if etime > 0 {
					pcnt = proc.fr.totaltime.tv_sec
					pcnt += proc.fr.totaltime.tv_usec / 1000000
					pcnt = pcnt * 100 / etime
					if pcnt > 99.9 {
						pcnt = 99.9
					}
				} else {
					pcnt = 0.0
				}
			}
			pidstr := fmt.Sprint(proc.fr.pid)
			inststr := fmt.Sprint(proc.fr.instcnt / 1000)
			cpustr := fmt.Sprintf("%4.1f", pcnt)
			var progstr, prognamestr string
			if proc.fr != nil {
				progstr = fmt.Sprintf("#%d", proc.fr.caller.st[1])
				prognamestr = fmt.Sprint(DB.Fetch(proc.fr.caller.st[1]).name)
			} else {
				progstr = fmt.Sprintf("#%d", proc.prog)
				prognamestr = fmt.Sprint(DB.Fetch(proc.prog).name)
			}
			buf := fmt.Sprintf(pat, pidstr, "--", time_format_2((long) (rtime - proc.fr.started)), inststr, cpustr, progstr, prognamestr, DB.Fetch(proc.player).name, "EVENT_WAITFOR" )
			if Wizard(DB.Fetch(player).Owner) || DB.Fetch(proc.prog).Owner == DB.Fetch(player).Owner || proc.player == player {
				notify_nolisten(player, buf, true)
			}
			r++
		}
	}
	return
}

/* Given a muf list array, appends pids to it where ref
 * matches the trigger, program, or player.  If ref is #-1
 * then all processes waiting for mufevents are added.
 */
func get_mufevent_pids(stk_array *nw, ObjectID ref) (nw *stk_array) {
	nw = make(Array)
	if ref < 0 {
		proc := mufevent_processes; proc != nil; proc = proc.next {
			if !proc.deleted {
				nw = append(nw, &inst{ data: len(nw) }, &inst{ data: proc.fr.pid })
			}
		}
	} else {
		proc := mufevent_processes; proc != nil; proc = proc.next {
			if !proc.deleted {
				if proc.player == ref || proc.prog == ref || proc.fr.trig == ref {
					nw = append(nw, &inst{ data: len(nw) }, &inst{ data: proc.fr.pid })
				}
			}
		}
	}
	return
}

func get_mufevent_pidinfo(nw *stk_array, pid int) (r *stk_array) {
	time_t      rtime = time(NULL);
	time_t      etime = 0;

	var proc *mufevent_process
	for proc = mufevent_processes; proc != nil && (proc.deleted || proc.fr.pid != pid); proc = proc.next {}
	if proc != nil && proc.fr.pid == pid {
		var pcnt float64
		if proc.fr {
			if etime = rtime - proc.fr.started; etime > 0 {
				pcnt = proc.fr.totaltime.tv_sec
				pcnt += proc.fr.totaltime.tv_usec / 1000000
				if pcnt = pcnt * 100 / etime; pcnt > 100.0 {
					pcnt = 100.0
				}
			} else {
				pcnt = 0.0
			}
		}

		array_setitem(&r, &inst{ data: "PID" }, &inst{ data: proc.fr.pid })
		array_setitem(&r, &inst{ data: "CALLED_PROG" }, &inst{ data: proc.prog })
		array_setitem(&r, &inst{ data: "TRIG" }, &inst{ data: proc.fr.trig })
		array_setitem(&r, &inst{ data: "PLAYER" }, &inst{ data: proc.player })
		array_setitem(&r, &inst{ data: "CALLED_DATA" }, &inst{ data: "EVENT_WAITFOR" })
		array_setitem(&r, &inst{ data: "INSTCNT" }, &inst{ data: proc.fr,instcnt })
		array_setitem(&r, &inst{ data: "DESCR" }, &inst{ data: proc.fr.descr })
		array_setitem(&r, &inst{ data: "CPU" }, &inst{ data: pcnt })
		array_setitem(&r, &inst{ data: "NEXTRUN" }, &inst{ data: -1 })
		array_setitem(&r, &inst{ data: "STARTED" }, &inst{ data: proc.fr.started })
		array_setitem(&r, &inst{ data: "TYPE" }, &inst{ data: "MUFEVENT" })
		array_setitem(&r, &inst{ data: "SUBTYPE" }, &inst{ data: "" })
		arr := make(Array, len(proc.filters))
		copy(arr, proc.filters)
		array_setitem(&nw, &inst{ data: "FILTERS" }, &inst{ data: arr })
	}
	return
}

/* Returns how many events are waiting to be processed.
 */
func muf_event_count(fr *frame) (r int) {
	for ptr := fr.events; ptr != nil; ptr = ptr.next {
		r++
	}
	return
}

/* Returns how many events of the given event type are waiting to be processed.
 * The eventid passed can be an smatch string.
 */
func muf_event_exists(fr *frame, eventid string) (r int) {
	for ptr := fr.events; ptr != nil; ptr = ptr.next {
		if !smatch(eventid, ptr.event) {
			r++
		}
	}
	return
}

/* Adds a MUF event to the event queue for the given program instance.
 * If the exclusive flag is true, and if an item of the same event type
 * already exists in the queue, the new one will NOT be added.
 */
func muf_event_add(fr *frame, event string, val *inst, exclusive bool) {
	ptr := fr.events
	if exclusive {
		for ptr != nil && ptr.next != nil; ptr = ptr.next {
			if event == ptr.event {
				return
			}
		}

		if ptr != nil && event == ptr.event {
			return
		}
	}

	if ptr == nil {
		fr.events = &mufevent{ event: event, data: val }
	} else {
		ptr.next = &mufevent{ event: event, data: val }
	}
}

/* Removes the first event of one of the specified types from the event queue
 * of the given program instance.
 * Returns a pointer to the removed event to the caller.
 * Returns NULL if no matching events are found.
 */
func muf_event_pop_specific(fr *frame, events []string) (r *mufevent) {
	for _, v := range events {
		if fr.events != nil && !smatch(v, fr.events.event) {
			r = fr.events
			fr.events = r.next
			return r
		}
	}

	for ptr := fr.events; ptr != nil && ptr.next; ptr = ptr.next {
		for _, v := range events {
			if !smatch(v, ptr.next.event) {
				r = ptr.next
				ptr.next = r.next
				return r
			}
		}
	}
	return
}

/* Removes a given MUF event type from the event queue of the given
 * program instance.  If which is MUFEVENT_ALL, all instances are removed.
 * If which is MUFEVENT_FIRST, only the first instance is removed.
 * If which is MUFEVENT_LAST, only the last instance is removed.
 */
func muf_event_remove(fr *frame, event string, which int) {
	for fr.events != nil && event == fr.events.event {
		if which != MUFEVENT_LAST {
			fr.events = fr.events.next
			if which == MUFEVENT_FIRST {
				return
			}
		}
	}

	for ptr := fr.events; ptr != nil && ptr.next != nil; ptr = ptr.next {
		if event == ptr.next.event {
			if which == MUFEVENT_LAST {
				ptr = ptr.next
			} else {
				ptr.next = ptr.next.next
				if which == MUFEVENT_FIRST {
					return
				}
			}
		}
	}
}

/* This pops the top muf event off of the given program instance's
 * event queue, and returns it to the caller.
 */
func muf_event_pop(struct frame *fr) (r *mufevent) {
	if fr.events != nil {
		r = fr.events
		fr.events = fr.events.next
	}
	return
}

/* For all program instances who are in the EVENT_WAIT queue,
 * check to see if they have any items in their event queue.
 * If so, then process one each.  Up to ten programs can have
 * events processed at a time.
 *
 * This also needs to make sure that background processes aren't
 * waiting for READ events.
 */
func muf_event_process() {
	var proc *mufevent_process
	limit := 10
	for proc = mufevent_processes; proc != nil && limit > 0; proc = proc.next {
		if !proc.deleted && proc.fr != nil {
			var ev *mufevent
			if len(proc.filters) > 0 {
				/* Make sure it's not waiting for a READ event, if it's backgrounded */
				if proc.fr.been_background {
					for _, v := range proc.filters {
						if strcasecmp(v, "READ") == 0 {
							/* It's a backgrounded process, waiting for a READ...
							* should we throw an error?  Should we push a null event onto the list?  At this point, I'm
							* pushing a READ event with descr = -1, so that it will at least get out of its loop. -winged
							*/
							muf_event_add(proc.fr, "READ", &inst{ data: -1 }, 0)
							break
						}
					}
				}

				/* Search prog's event list for the apropriate event type. */

				/* HACK:  This is probably inefficient to be walking this queue over and over. Hopefully it's usually a short list.
				 * Would it be more efficient to use a hash table?  It'd be more wasteful of memory, I think. -winged
				 */
				ev = muf_event_pop_specific(proc.fr, proc.filters)
			} else {
				/* Pop first event off of prog's event queue. */
				ev = muf_event_pop(proc.fr)
			}
			if ev != nil {
				limit--
				if proc.fr.argument.top + 1 >= STACK_SIZE {
					notify_nolisten(proc.player, "Program stack overflow.", true)
					prog_clean(proc.fr)
				} else {
					player := DB.FetchPlayer(proc.player)
					current_program := player.curr_prog
					block := player.block
					is_bg := proc.fr.multitask == BACKGROUND
					proc.fr.argument.st[proc.fr.argument.top] = ev.data
					proc.fr.argument.top++
					push(proc.fr.argument.st, &(proc.fr.argument.top), ev.event)
					interp_loop(proc.player, proc.prog, proc.fr, false)
					if is_bg {
						player.block = block
						player.curr_prog = current_program
					}
				}
				proc.fr = nil  /* We do NOT want to free this program after every EVENT_WAIT. */
				proc.deleted = true
			}
		}
	}
	for proc = mufevent_processes; proc != nil; {
		next = proc.next
		if proc.deleted {
			muf_event_process_free(proc)
		}
		proc = next
	}
}