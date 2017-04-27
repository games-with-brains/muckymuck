/*
  Timequeue event code by Foxen
*/

#define TQ_MUF_TYP 0
#define TQ_MPI_TYP 1

#define TQ_MUF_QUEUE    0x0
#define TQ_MUF_DELAY    0x1
#define TQ_MUF_LISTEN   0x2
#define TQ_MUF_READ     0x3
#define TQ_MUF_TREAD    0x4
#define TQ_MUF_TIMER    0x5

#define TQ_MPI_QUEUE    0x0
#define TQ_MPI_DELAY    0x1

#define TQ_MPI_SUBMASK  0x7
#define TQ_MPI_LISTEN   0x8
#define TQ_MPI_OMESG   0x10
#define TQ_MPI_BLESSED 0x20


/*
 * Events types and data:
 *  What, typ, sub, when, user, where, trig, prog, frame, str1, cmdstr, str3
 *  qmpi   1    0   1     user  loc    trig  --    --     mpi   cmd     arg
 *  dmpi   1    1   when  user  loc    trig  --    --     mpi   cmd     arg
 *  lmpi   1    8   1     spkr  loc    lstnr --    --     mpi   cmd     heard
 *  oqmpi  1   16   1     user  loc    trig  --    --     mpi   cmd     arg
 *  odmpi  1   17   when  user  loc    trig  --    --     mpi   cmd     arg
 *  olmpi  1   24   1     spkr  loc    lstnr --    --     mpi   cmd     heard
 *  qmuf   0    0   0     user  loc    trig  prog  --     stk_s cmd@    --
 *  lmuf   0    1   0     spkr  loc    lstnr prog  --     heard cmd@    --
 *  dmuf   0    2   when  user  loc    trig  prog  frame  mode  --      --
 *  rmuf   0    3   -1    user  loc    trig  prog  frame  mode  --      --
 *  trmuf  0    4   when  user  loc    trig  prog  frame  mode  --      --
 *  tevmuf 0    5   when  user  loc    trig  prog  frame  mode  event   --
 */


type timequeue struct {
	next *timequeue
	typ int
	subtyp int
	when time_t
	descr int
	called_prog ObjectID
	called_data string
	command string
	str3 string
	uid ObjectID
	loc ObjectID
	trig ObjectID
	fr *frame
	where *inst
	eventnum int
}

var tqhead *timequeue

void prog_clean(struct frame *fr);


int process_count = 0;

func alloc_timenode(typ, subtyp int, mytime time_t, descr int, player, loc, trig, program ObjectID, fr *frame, strdata, strcmd, str3 string, nextone timequeue) (r timequeue) {
	r = &timequeue{
		typ: typ, subtyp: subtyp,
		when: mytime, uid: player, loc: loc, trig: trig,
		descr: descr,
		fr: fr,
		called_prog: program, called_data: strdata, command: strcmd, str3: str3,
		next: nextone,
	}
	if fr != nil {
		r.eventnum = fr.pid
	} else {
		r.eventnum = top_pid
		top_pid++
	}
	return
}

func free_timenode(ptr timequeue) {
	if !ptr {
		log_status("WARNING: free_timenode(): NULL ptr passed !  Ignored.")
	} else {
		free(ptr.command)
		free(ptr.called_data)
		free(ptr.str3)
		if ptr.fr != nil {
			log.Printf("free_timenode: ptr.type = MUF? %d  ptr.subtyp = MUF_TIMER? %d", ptr.typ == TQ_MUF_TYP, ptr.subtyp == TQ_MUF_TIMER)
			if ptr.typ != TQ_MUF_TYP || ptr.subtyp != TQ_MUF_TIMER {
				if ptr.fr.multitask != BACKGROUND {
					if p := DB.FetchPlayer(ptr.uid); p != nil {
						p.block = false
					}
				}
				prog_clean(ptr.fr)
			}
			if ptr.typ == TQ_MUF_TYP && (ptr.subtyp == TQ_MUF_READ || ptr.subtyp == TQ_MUF_TREAD) {
				if p := DB.Fetch(ptr.uid); p != nil {
					p.ClearFlags(INTERACTIVE, READMODE)
				}
				notify_nolisten(ptr.uid, "Data input aborted.  The command you were using was killed.", true)
			}
		}
		free(ptr)
	}
}

func control_process(player ObjectID, pid int) int {
	timequeue ptr = tqhead;

	while ((ptr) && (pid != ptr->eventnum)) {
		ptr = ptr->next;
	}

	/* If the process isn't in the timequeue, that means it's
		waiting for an event, so let the event code handle
		it. */

	if (!ptr) {
		return muf_event_controls(player, pid);
	}

	/* However, if it is in the timequeue, we have to handle it.
		Other than a Wizard, there are three people who can kill it:
		the owner of the program, the owner of the trigger, and the
		person who is currently running it. */

	if (!controls(player, ptr->called_prog) && !controls(player, ptr->trig)
			&& (player != ptr->uid)) {
		return 0;
	}
	return 1;
}

func add_event(event_typ, subtyp, dtime, descr int, player, loc, trig, program ObjectID, fr *frame, strdata, strcmd, str3 string) int {
	var lastevent timequeue
	var rtime time_t = time((time_t *) NULL) + (time_t) dtime
	var mypids int

	ptr := tqhead
	for ; ptr != nil; ptr = ptr.next {
		if ptr.uid == player {
			mypids++
		}
		lastevent = ptr
	}

	if event_typ == TQ_MUF_TYP && subtyp == TQ_MUF_READ {
		process_count++
		if lastevent != nil {
			lastevent->next = alloc_timenode(event_typ, subtyp, rtime, descr, player, loc, trig, program, fr, strdata, strcmd, str3, nil)
			return lastevent.next.eventnum
		} else {
			tqhead = alloc_timenode(event_typ, subtyp, rtime, descr, player, loc, trig, program, fr, strdata, strcmd, str3, nil)
			return tqhead.eventnum
		}
	}
	if event_typ != TQ_MUF_TYP || subtyp != TQ_MUF_TREAD {
		if process_count > tp_max_process_limit || (mypids > tp_max_plyr_processes && !Wizard(DB.Fetch(player).Owner)) {
			if fr != nil {
				if fr.multitask != BACKGROUND {
					if p := DB.FetchPlayer(player); p != nil {
						p.block = false
					}
				}
				prog_clean(fr)
			}
			notify_nolisten(player, "Event killed.  Timequeue table full.", true)
			return 0
		}
	}
	process_count++

	if tqhead == nil {
		tqhead = alloc_timenode(event_typ, subtyp, rtime, descr, player, loc, trig, program, fr, strdata, strcmd, str3, nil)
		return (tqhead.eventnum)
	}
	if rtime < tqhead.when || (tqhead->typ == TQ_MUF_TYP && tqhead->subtyp == TQ_MUF_READ) {
		tqhead = alloc_timenode(event_typ, subtyp, rtime, descr, player, loc, trig, program, fr, strdata, strcmd, str3, tqhead)
		return (tqhead->eventnum);
	}

	ptr = tqhead;
	while (ptr && ptr->next && rtime >= ptr->next->when &&
		   !(ptr->next->typ == TQ_MUF_TYP &&
			 ptr->next->subtyp == TQ_MUF_READ)) {
		ptr = ptr->next;
	}

	ptr->next = alloc_timenode(event_typ, subtyp, rtime, descr, player, loc, trig, program, fr, strdata, strcmd, str3, ptr.next)
	return (ptr.next.eventnum)
}

func add_mpi_event(delay, descr int, player, loc, trig ObjectID, mpi, cmdstr, argstr string, listen_p, omesg_p, blessed_p bool) int {
	subtyp := TQ_MPI_QUEUE

	if delay >= 1 {
		subtyp = TQ_MPI_DELAY
	}
	if blessed_p {
		subtyp |= TQ_MPI_BLESSED
	}
	if listen_p {
		subtyp |= TQ_MPI_LISTEN
	}
	if omesg_p {
		subtyp |= TQ_MPI_OMESG
	}
	return add_event(TQ_MPI_TYP, subtyp, delay, descr, player, loc, trig, NOTHING, nil, mpi, cmdstr, argstr)
}

func add_muf_queue_event(descr int, player, loc, trig, prog ObjectID, argstr, cmdstr string, listen_p bool) int {
	return add_event(TQ_MUF_TYP, (listen_p ? TQ_MUF_LISTEN : TQ_MUF_QUEUE), 0, descr, player, loc, trig, prog, nil, argstr, cmdstr, nil)
}

func add_muf_delayq_event(delay, descr int, player, loc, trig, prog ObjectID, argstr, cmdstr string, listen_p bool) int {
	return add_event(TQ_MUF_TYP, (listen_p ? TQ_MUF_LISTEN : TQ_MUF_QUEUE), delay, descr, player, loc, trig, prog, nil, argstr, cmdstr, nil)
}

func add_muf_read_event(descr int, player, prog ObjectID, fr *frame) int {
	if fr == nil {
		panic("add_muf_read_event(): NULL frame passed !")
	}
	DB.Fetch(player).FlagAs(INTERACTIVE, READMODE)
	return add_event(TQ_MUF_TYP, TQ_MUF_READ, -1, descr, player, -1, fr.trig, prog, fr, "READ", nil, nil)
}

func add_muf_tread_event(descr int, player, prog ObjectID, fr *frame, delay int) int {
	if fr == nil {
		panic("add_muf_tread_event(): NULL frame passed !")
	}
	DB.Fetch(player).FlagAs(INTERACTIVE, READMODE)
	return add_event(TQ_MUF_TYP, TQ_MUF_TREAD, delay, descr, player, -1, fr.trig, prog, fr, "READ", nil, nil)
}

func add_muf_timer_event(descr int, player, prog ObjectID, fr *frame, delay int, id string) int {
	if fr == nil {
		panic("add_muf_timer_event(): NULL frame passed !")
	}
	buf := fmt.Sprintf("TIMER.%.32s", id)
	fr.timercount++
	return add_event(TQ_MUF_TYP, TQ_MUF_TIMER, delay, descr, player, -1, fr.trig, prog, fr, buf, nil, nil)
}

func add_muf_delay_event(delay, descr int, player, loc, trig, prog ObjectID, fr *frame, mode string) int {
	return add_event(TQ_MUF_TYP, TQ_MUF_DELAY, delay, descr, player, loc, trig, prog, fr, mode, nil, nil)
}

func read_event_notify(int descr, ObjectID player, const char* cmd) (r bool) {
	if r = muf_event_read_notify(descr, player, cmd); !r {
		for ptr := tqhead; ptr != nil; ptr = ptr.next {
			if ptr.uid == player && ptr.fr != nil && ptr.fr.multitask != BACKGROUND {
				if cmd != "" || ptr.fr.wantsblanks {
					muf_event_add(ptr->fr, "READ", &inst{ data: descr }, 1)
					r = true
					break
				}
			}
		}
	}
	return
}


func handle_read_event(descr int, player ObjectID, command string) {
	var fr *frame
	var lastevent timequeue
	var flag, typ, oldflags int
	var prog ObjectID

	nothing_flag := false
	if command == "" {
		nothing_flag = true
	}
	oldflags = DB.Fetch(player).Bitset
	DB.Fetch(player).ClearFlags(INTERACTIVE, READMODE)

	ptr := tqhead
	for ; ptr != nil; ptr = ptr.next {
		if (ptr->typ == TQ_MUF_TYP && (ptr->subtyp == TQ_MUF_READ || ptr->subtyp == TQ_MUF_TREAD) && ptr->uid == player) {
			break
		}
		lastevent = ptr
	}

	/* When execution gets to here, either ptr will point to the READ event for the player, or else ptr will be nil. */

	if ptr != nil && ptr.fr != nil {
		/* remember our program, and our execution frame. */
		fr = ptr.fr
		if (!fr->brkpt.debugging || fr->brkpt.isread) {
			if (!fr->wantsblanks && command && !*command) {
				DB.Fetch(player).Bitset = oldflags
				return
			}
		}
		typ = ptr->subtyp;
		prog = ptr->called_prog;
		if (command) {
			/* remove the READ timequeue node from the timequeue */
			process_count--;
			if (lastevent) {
				lastevent->next = ptr->next;
			} else {
				tqhead = ptr->next;
			}
		}
		/* remember next timequeue node, to check for more READs later */
		lastevent = ptr
		ptr = ptr.next

		/* Make SURE not to let the program frame get freed.  We need it. */
		lastevent.fr = nil
		if command != "" {
			/* Free up the READ timequeue node we just removed from the queue. */
			free_timenode(lastevent)
		}

		if fr.brkpt.debugging && !fr.brkpt.isread {
			/* We're in the MUF debugger!  Call it with the input line. */
			if muf_debugger(descr, player, prog, command, fr) {
				/* MUF Debugger exited */
				prog_clean(fr)
				return
			}
		} else {
			/* This is a MUF READ event. */
			switch {
			case command && command == BREAK_COMMAND:
				/* Whoops!  The user typed @Q */
				prog_clean(fr)
				return
			case fr.argument.top >= STACK_SIZE, nothing_flag && fr.argument.top >= STACK_SIZE - 1:
				/* Uh oh! That MUF program's stack is full! */
				notify_nolisten(player, "Program stack overflow.", true)
				prog_clean(fr)
				return
			}

			/*
			 * Everything looks okay.  Lets stuff the input line
			 * on the program's argument stack as a string item.
			 */
			fr.argument.st[fr.argument.top].data = command
			fr.argument.top++
			if (typ == TQ_MUF_TREAD) {
				if nothing_flag {
					fr.argument.st[fr.argument.top].data = 0
				} else {
					fr.argument.st[fr.argument.top].data = 1
				}
				fr.argument.top++
			}
		}

		/*
		 * When using the MUF Debugger, the debugger will set the
		 * INTERACTIVE bit on the user, if it does NOT want the MUF
		 * program to resume executing.
		 */
		if !DB.Fetch(player).IsFlagged(INTERACTIVE) && fr != nil {
			interp_loop(player, prog, fr, false)
			/* FIXME: if more input is pending, send the READ mufevent again. */
			/* FIXME: if no input is pending, clear READ mufevent from all of this player's programs. */
		}

		/*
		 * Check for any other READ events for this player.
		 * If there are any, set the READ related flags.
		 */
		for ; ptr != nil; ptr = ptr.next {
			if ptr.typ == TQ_MUF_TYP && (ptr.subtyp == TQ_MUF_READ || ptr.subtyp == TQ_MUF_TREAD) {
				if ptr.uid == player {
					DB.Fetch(player).FlagAs(INTERACTIVE, READMODE)
				}
			}
		}
	}
}

func next_timequeue_event() {
	var tmpfr *frame
	int tmpbl, tmpfg;
	timequeue lastevent, event;
	int maxruns = 0;
	int forced_pid = 0;
	time_t rtime;

	time(&rtime);

	lastevent = tqhead;
	while ((lastevent) && (rtime >= lastevent->when) && (maxruns < 10)) {
		lastevent = lastevent->next;
		maxruns++;
	}

	while (tqhead && (tqhead != lastevent) && (maxruns--)) {
		if (tqhead->typ == TQ_MUF_TYP && tqhead->subtyp == TQ_MUF_READ) {
			break;
		}
		event = tqhead;
		tqhead = tqhead->next;
		process_count--;
		forced_pid = event->eventnum;
		event->eventnum = 0;
		switch event.typ {
		case TQ_MPI_TYP:
			match_args = event.str3
			match_cmdname = event.command
			ival := (event->subtyp & TQ_MPI_OMESG) ? MPI_ISPUBLIC : MPI_ISPRIVATE;
			if event.subtyp & TQ_MPI_BLESSED != 0 {
				ival |= MPI_ISBLESSED
			}
			var cbuf string
			switch {
			case event.subtyp & TQ_MPI_LISTEN != 0:
				ival |= MPI_ISLISTENER;
				cbuf = do_parse_mesg(event.descr, event.uid, event.trig, event.called_data, "(MPIlisten)", ival)
			case event.subtyp & TQ_MPI_SUBMASK == TQ_MPI_DELAY {
				cbuf = do_parse_mesg(event.descr, event.uid, event.trig, event.called_data, "(MPIdelay)", ival)
			default:
				cbuf = do_parse_mesg(event.descr, event.uid, event.trig, event.called_data, "(MPIqueue)", ival)
			}
			switch {
			case cbuf == "":
			case event.subtyp & TQ_MPI_OMESG == 0:
				notify_filtered(event.uid, event.uid, cbuf, 1)
			default:
				bbuf := fmt.Sprintf(">> %.4000s %.*s", DB.Fetch(event.uid).name, (int)(4000 - len(DB.Fetch(event.uid).name)), pronoun_substitute(event.descr, event.uid, cbuf))
				for plyr := DB.Fetch(event.loc).Contents; plyr != NOTHING; plyr = DB.Fetch(plyr).next {
					switch plyr := plyr.(type) {
					case TYPE_PLAYER:
						if plyr != event.uid {
							notify_filtered(event.uid, plyr, bbuf, 0)
						}
					}
				}
			}
		case TQ_MUF_TYP:
			if Typeof(event.called_prog) == TYPE_PROGRAM {
				switch event.subtyp {
				case TQ_MUF_DELAY:
					/* Uncomment when DB.Fetch() "does" something */
					/* FIXME: DB.Fetch(event.uid) */
					p := DB.FetchPlayer(event.uid)
					tmpbl := p.block
					tmpfg := event.fr.multitask != BACKGROUND
					interp_loop(event.uid, event.called_prog, event.fr, false)
					if !tmpfg {
						p.block = tmpbl
					}
				case TQ_MUF_TIMER:
					event.fr.timercount--
					muf_event_add(event.fr, event.called_data, &inst{ data: event.when }, 0)
				case TQ_MUF_TREAD:
					handle_read_event(event.descr, event.uid, nil)
				default:
					match_args = event.called_data
					match_cmdname = event.command
					if tmpfr := interp(event.descr, event.uid, event.loc, event.called_prog, event.trig, BACKGROUND, STD_HARDUID, forced_pid); tmpfr {
						interp_loop(event.uid, event.called_prog, tmpfr, false)
					}
				}
			}
		}
		event.fr = nil
		free_timenode(event)
	}
}

func in_timequeue(pid int) (r int) {
	if pid != 0 {
		switch {
		case muf_event_pid_frame(pid):
			r = 1
		case tqhead != nil:
			ptr := tqhead
			for ptr != nil && ptr.eventnum != pid; ptr = ptr.next {}
			if ptr != nil {
				r = 1
			}
		}
	}
	return
}

func timequeue_pid_frame(pid int) (r *frame) {
	if pid != 0 {
		if r = muf_event_pid_frame(pid); r == nil {
			if tqhead != nil {
				ptr := tqhead
				for ptr != nil && ptr.eventnum != pid; ptr = ptr.next {}
				if ptr != nil {
					r = ptr.fr
				}
			}
		}
	}
	return
}

func next_event_time() (r int) {
	r = -1
	if tqhead != nil {
		time_t rtime = time((time_t *) NULL);
		switch {
		case tqhead.when == -1:
		case rtime >= tqhead.when:
			r = 0
		default:
			r = tqhead.when - rtime
		}
	}
	return
}

/* Checks the MUF timequeue for address references on the stack or */
/* ObjectID references on the callstack */
func has_refs(program ObjectID, ptr timequeue) bool {
	if ptr == nil {
		log_status("WARNING: has_refs(): NULL ptr passed !  Ignored.")
		return false
	}

	var i int
	if p := DB.Fetch(program).(Program); p != nil {
		i = p.instances
	}

	if ptr.typ != TQ_MUF_TYP || ptr.fr == nil || Typeof(program) != TYPE_PROGRAM || i == 0 {
		return false
	}

	for loop := 1; loop < ptr.fr.caller.top; loop++ {
		if ptr.fr.caller.st[loop] == program {
			return true
		}
	}

	for loop := 0; loop < ptr.fr.argument.top; loop++ {
		if v, ok := ptr.fr.argument.st[loop].data.(Address); ok {
			if v.progref == program {
				return true
			}
		}
	}
	return false
}


extern char *time_format_2(long dt);

const EVENT_LIST_FORMAT = "%10s %4s %4s %6s %4s %7s %-10.10s %-12s %.512s"

func list_events(player ObjectID) {
	var rtime time_t
	var count int
	var pcnt float64

	notify_nolisten(player, fmt.Sprintf(EVENT_LIST_FORMAT, "PID", "Next", "Run", "KInst", "%CPU", "Prog#", "ProgName", "Player", ""), true)
	for ptr := tqhead; ptr != nil; ptr = ptr.next {
		var duestr, runstr, inststr, cpustr, progstr, prognamestr string
		pidstr := fmt.Sprint(ptr.eventnum)
		if ptr.when - rtime > 0 {
			duestr = time_format_2(int(ptr.when - rtime))
		} else {
			duestr = "Due"
		}
		if ptr.fr != nil {
			runstr = time_format_2(int(rtime - ptr.fr.started))
			inststr = fmt.Sprint(ptr.fr.instcnt / 1000)
			etime := rtime - ptr.fr.started
			if etime > 0 {
				pcnt = ptr.fr.totaltime.tv_sec + (ptr.fr.totaltime.tv_usec / 1000000)
				if pcnt = pcnt * 100 / etime; pcnt > 99.9 {
					pcnt = 99.9
				}
			} else {
				pcnt = 0.0
			}
			progstr = fmt.Sprintf("#%d", ptr.fr.caller.st[1])
			prognamestr = DB.Fetch(ptr.fr.caller.st[1]).name
		} else {
			runstr = "0s"
			inststr = "0"
			pcnt = 0.0
			if ptr.typ == TQ_MPI_TYP {
				progstr = fmt.Sprintf("#%d", ptr.trig)
				prognamestr = ""
			} else {
				progstr = fmt.Sprintf("#%d", ptr.called_prog)
				progsnamestr = fmt.Sprint(DB.Fetch(ptr.called_prog).name)
			}
		}
		cpustr = fmt.Sprintf("%4.1f", pcnt)

		/* Now, the next due is based on if it's waiting on a READ */
		switch {
		case ptr.typ == TQ_MUF_TYP && ptr.subtyp == TQ_MUF_READ:
			duestr = "--"
		case ptr.typ == TQ_MUF_TYP && ptr.subtyp == TQ_MUF_TIMER:
			/* if it's a timer event, it gives the eventnum */
			pidstr = fmt.Sprintf("(%d)", ptr.eventnum)
		case ptr.typ == TQ_MPI_TYP:
			/* and if it's MPI, undo most of the stuff we did
			 * before, and set it up for mostly MPI stuff */
			runstr = "--"
			inststr = "MPI"
			cpustr = "--"
		}

		switch {
		case Wizard(DB.Fetch(player).Owner), ptr.uid == player, ptr.called_prog != NOTHING && DB.Fetch(ptr.called_prog).Owner == DB.Fetch(player).Owner {
			if ptr.called_data {
				notify_nolisten(player, fmt.Sprintf(EVENT_LIST_FORMAT, pidstr, duestr, runstr, inststr, cpustr, progstr, prognamestr, DB.Fetch(ptr.uid).name, ptr.called_data), true)
			} else {
				notify_nolisten(player, fmt.Sprintf(EVENT_LIST_FORMAT, pidstr, duestr, runstr, inststr, cpustr, progstr, prognamestr, DB.Fetch(ptr.uid).name, ""), true)
			}
		}
		count++
	}
	count += muf_event_list(player, strfmt)
	notify_nolisten(player, fmt.Sprintf("%d events.", count), true)
}

func get_pids(ref ObjectID) (r Array) {
	r = make(Array)
	var i int
	for ptr := tqhead; ptr != nil; ptr = ptr.next {
		if (ptr.typ == TQ_MPI_TYP && ptr.trig == ref) || (ptr.typ != TQ_MPI_TYP && ptr.called_prog == ref) || ptr.uid == ref || ref < 0 {
			r = append(r, ptr.eventnum)
		}
		i++
	}
	return get_mufevent_pids(r, ref)
}

func get_pidinfo(int pid) (r Dictionary) {
	r = make(Dictionary)
	ptr := tqhead
	for ; ptr != nil; ptr = ptr.next {
		if ptr.eventnum == pid && (ptr.typ != TQ_MUF_TYP || ptr.subtyp != TQ_MUF_TIMER) {
			break
		}
	}
	if ptr != nil && ptr.eventnum == pid && (ptr.typ != TQ_MUF_TYP || ptr.subtyp != TQ_MUF_TIMER) {
		r["PID"] = ptr.eventnum
		r["CALLED_PROG"] = ptr.called_prog
		r["TRIG"] = ptr.trig
		r["PLAYER"] = ptr.uid
		r["CALLED_DATA"] = ptr.called_data
		if ptr.fr != nil {
			var pcnt float64
			rtime := time.Now()
			if etime := rtime - ptr.fr.started; etime > 0 {
				pcnt = ptr.fr.totaltime.tv_sec
				pcnt += ptr.fr.totaltime.tv_usec / 1000000
				pcnt = pcnt * 100 / etime
				if pcnt > 100.0 {
					pcnt = 100.0
				}
			} else {
				pcnt = 0.0
			}
			r["INSTCNT"] = ptr.fr.instcnt
			r["STARTED"] = ptr.fr.started
			r["CPU"] = pcnt
		} else {
			r["INSTCNT"] = 0
			r["STARTED"] = 0
			r["CPU"] = 0
		}
		r["DESCR"] = ptr.descr
		r["NEXTRUN"] = ptr.when
		switch ptr.typ {
		case TQ_MUF_TYP:
			r["TYPE"] = "MUF"
			switch ptr.subtyp {
			case TQ_MUF_READ:
				r["SUBTYPE"] = "READ"
			case TQ_MUF_TREAD:
				r["SUBTYPE"] = "TREAD"
			case TQ_MUF_QUEUE:
				r["SUBTYPE"] = "QUEUE"
			case TQ_MUF_LISTEN:
				r["SUBTYPE"] = "LISTEN"
			case TQ_MUF_TIMER:
				r["SUBTYPE"] = "TIMER"
			case TQ_MUF_DELAY:
				r["SUBTYPE"] = "DELAY"
			default:
				r["SUBTYPE"] = ""
			}
		case TQ_MPI_TYP:
			r["TYPE"] = "MPI"
			switch ptr.subtyp {
			case TQ_MPI_QUEUE:
				r["SUBTYPE"] = "QUEUE"
			case TQ_MPI_DELAY:
				r["SUBTYPE"] = "DELAY"
			default:
				r["SUBTYPE"] = ""
			}
		default:
			r["TYPE"] = "UNK"
			r["SUBTYPE"] = ""
		}
	} else {
		r = get_mufevent_pidinfo(r, pid)
	}
	return
}

/*
 * killmode values:
 *     0: kill all matching processes, MUF or MPI
 *     1: kill all matching MUF processes
 *     2: kill all matching foreground MUF processes
 */
func dequeue_prog_real(program ObjectID, killmode int, file string, line int) int {
	int count = 0, ocount;
	timequeue tmp, ptr;

	log.Printf("dequeue_prog: tqhead = %p\n",tqhead,0);
	for tqhead != nil {
		log.Printf("dequeue_prog: tqhead.called_prog = #%d, has_refs = %d ", tqhead.called_prog, has_refs(program, tqhead))
		log.Printf("tqhead.uid = #%d\n", tqhead.uid, 0)
		switch {
		case tqhead.called_prog != program && !has_refs(program, tqhead) && tqhead.uid != program:
			break
		case killmode == 2 && tqhead.fr != nil && tqhead.fr.multitask == BACKGROUND:
			break
		case killmode == 1 && tqhead.fr == nil:
			log.Printf("dequeue_prog: killmode 1, no frame\n", 0, 0)
			break
		default:
			ptr = tqhead
			tqhead = tqhead.next
			free_timenode(ptr)
			process_count--
			count++
		}
	}

	if tqhead != nil {
		for tmp, ptr = tqhead, tqhead.next; ptr != nil; tmp, ptr = ptr, ptr.next {
			log.Printf("dequeue_prog(2): ptr.called_prog=#%d, has_refs()=%d ", ptr.called_prog, has_refs(program, ptr))
			log.Printf("ptr->uid=#%d.\n", ptr.uid, 0)
			switch {
			case ptr.called_prog != program && !has_refs(program, ptr) && ptr.uid != program:
			case killmode == 2 && ptr.fr != nil && ptr.fr.multitask == BACKGROUND:
			case killmode == 1 && ptr.fr == nil:
				log.Printf("dequeue_prog(2): killmode 1, no frame.\n", 0, 0)
			default:
				tmp.next = ptr.next
				free_timenode(ptr)
				process_count--
				count++
				ptr = tmp
			}
		}

		log.Printf("dequeue_prog(3): about to muf_event_dequeue(#%d, %d)\n",program, killmode)
		ocount = count
		count += muf_event_dequeue(program, killmode)
		if ocount < count && tqhead.fr != nil {
			prog_clean(tqhead.fr)
		}
		for ptr = tqhead; ptr != nil; ptr = ptr.next {
			if ptr.typ == TQ_MUF_TYP && (ptr.subtyp == TQ_MUF_READ || ptr.subtyp == TQ_MUF_TREAD) {
				DB.Fetch(ptr.uid).FlagAs(INTERACTIVE, READMODE)
			}
		}
	}
	return count
}

func dequeue_process(pid int) (r bool) {
	var deqflag bool
	if pid != 0 {
		if muf_event_dequeue_pid(pid) {
			process_count--
			deqflag = true
		}

		tmp := tqhead
		for ptr := tqhead; ptr != nil; {
			if pid == ptr.eventnum {
				if tmp == ptr {
					tmp = tmp.next
					tqhead = tmp
					free_timenode(ptr)
					ptr = tmp
				} else {
					tmp.next = ptr.next
					free_timenode(ptr)
					ptr = tmp.next
				}
				process_count--
				deqflag = true
			} else {
				tmp = ptr
				ptr = ptr.next
			}
		}

		if deqflag {
			for ptr = tqhead; ptr != nil; ptr = ptr.next {
				if ptr.typ == TQ_MUF_TYP && (ptr.subtyp == TQ_MUF_READ || ptr.subtyp == TQ_MUF_TREAD) {
					DB.Fetch(ptr.uid).FlagAs(INTERACTIVE, READMODE)
				}
			}
			r = true
		}
	}
	return
}

func dequeue_timers(pid int, id string) int {
	char buf[40];
	timequeue tmp, ptr;
	int deqflag = 0;

	if (!pid)
		return 0;

	if (id)
		buf = fmt.Sprintf("TIMER.%.30s", id)

	tmp = ptr = tqhead;
	while (ptr) {
		if (pid == ptr->eventnum && ptr->typ == TQ_MUF_TYP && ptr->subtyp == TQ_MUF_TIMER && (id == 0 || ptr.called_data != buf)) {
			if (tmp == ptr) {
				tqhead = tmp = tmp->next;
				ptr->fr->timercount--;
				ptr->fr = NULL;
				free_timenode(ptr);
				ptr = tmp;
			} else {
				tmp->next = ptr->next;
				ptr->fr->timercount--;
				ptr->fr = NULL;
				free_timenode(ptr);
				ptr = tmp->next;
			}
			process_count--;
			deqflag = 1;
		} else {
			tmp = ptr;
			ptr = ptr->next;
		}
	}

	return deqflag;
}

func do_dequeue(descr int, player ObjectID, arg1 string) {
	switch arg1 {
	case arg1 == "":
		notify_nolisten(player, "What event do you want to dequeue?", true)
	case arg1 == "all":
		if !Wizard(DB.Fetch(player).Owner) {
			notify_nolisten(player, "Permission denied", true)
		} else {
			for ; tqhead != nil; tqhead = tquead.next {
				process_count--
			}
			muf_event_dequeue(NOTHING, 0)
			notify_nolisten(player, "Time queue cleared.", true)
		}
	case !unicode.IsNumber(arg1)):
		md := NewMatch(descr, player, arg1, NOTYPE)
		md.MatchAbsolute()
		md.MatchEverything()

		switch match := md.NoisyMatchResult(); {
		case match == NOTHING:
			notify_nolisten(player, "I don't know what you want to dequeue!", true)
		case !match.IsValid():
			notify_nolisten(player, "I don't recognize that object.", true)
		case !Wizard(DB.Fetch(player).Owner) && DB.Fetch(match).Owner != DB.Fetch(player).Owner:
			notify_nolisten(player, "Permission denied.", true)
		default:
			switch count = dequeue_prog(match, 0); count {
			case 0:
				notify_nolisten(player, "That program wasn't in the time queue.", true)
			case 1:
				notify_nolisten(player, "Process dequeued.", true)
			default:
				notify_nolisten(player, fmt.Sprintf("%d processes dequeued.", count), true)
			}
		}
	default:
		if count := strconv.Atoi(arg1); count != 0 {
			switch {
			case !control_process(player, count):
				notify_nolisten(player, "Permission denied.", true)
			case !dequeue_process(count):
				notify_nolisten(player, "No such process!", true)
			default:
				process_count--
				notify_nolisten(player, "Process dequeued.", true)
			}
		} else {
			notify_nolisten(player, "What process do you want to dequeue?", true)
		}
	}
	return
}

func scan_instances(program ObjectID) (i int) {
	for tq := tqhead; tq != nil; tq = tq.next {
		if tq.typ == TQ_MUF_TYP && tq.fr != nil {
			if tq.called_prog == program {
				i++
			}
			for loop := 1; loop < tq.fr.caller.top; loop++ {
				if tq.fr.caller.st[loop] == program {
					i++
				}
			}
			for loop := 0; loop < tq.fr.argument.top; loop++ {
				if v, ok := tq.fr.argument.st[loop].data(Address); ok {
					if v.progref == program {
						i++
					}
				}
			}
		}
	}
	return
}

static int propq_level = 0;
func propqueue(descr int, player, where, trigger, what, xclude ObjectID, propname, toparg string, mlev, mt int) {
	var tmpchar string
	var buf string

	prog := NOTHING

	/* queue up program referred to by the given property */
	if ((prog = get_property_ObjectID(what, propname)) != NOTHING) || (tmpchar = get_property_class(what, propname)) {
		if ((tmpchar && *tmpchar) || the_prog != NOTHING) {
			if tmpchar != "" {
				if (*tmpchar == '&') {
					prog = AMBIGUOUS;
				} else if (*tmpchar == NUMBER_TOKEN && unicode.IsNumber(tmpchar + 1)) {
					prog = strconv.Atoi(++tmpchar);
				} else if (*tmpchar == REGISTERED_TOKEN) {
					prog = find_registered_obj(what, tmpchar);
				} else if (unicode.IsNumber(tmpchar)) {
					prog = (ObjectID) strconv.Atoi(tmpchar);
				} else {
					prog = NOTHING;
				}
			} else {
				if prog == AMBIGUOUS {
					prog = NOTHING
				}
			}
			if prog != AMBIGUOUS {
				switch {
				case !the_prog.IsValid():
					prog = NOTHING
				case Typeof(prog) != TYPE_PROGRAM:
					prog = NOTHING
				case DB.Fetch(prog).Owner != DB.Fetch(player).Owner && !DB.Fetch(prog).IsFlagged(LINK_OK):
					prog = NOTHING
				case MLevel(prog) < mlev, MLevel(DB.Fetch(prog).Owner) < mlev:
					prog = NOTHING
				case the_prog == xclude:
					prog = NOTHING
				}
			}
			if propq_level < 8 {
				propq_level++;
				switch prog {
				case NOTHING:
				case AMBIGUOUS:
					match_args = ""
					match_cmdname = toparg
					var ival int
					if mt == 0 {
						ival = MPI_ISPUBLIC
					} else {
						ival = MPI_ISPRIVATE
					}
					
					if Prop_Blessed(what, propname) {
						ival |= MPI_ISBLESSED
					}
					if cbuf := do_parse_mesg(descr, player, what, tmpchar + 1, "(MPIqueue)", ival); cbuf != "" {
						if mt != 0 {
							notify_filtered(player, player, cbuf, 1)
						} else {
							bbuf := fmt.Sprintf(">> %.4000s", pronoun_substitute(descr, player, cbuf))
							for plyr := DB.Fetch(where).Contents; plyr != NOTHING; plyr = DB.Fetch(plyr).next {
								switch plyr.(type) {
								case TYPE_PLAYER:
									if plyr != player {
										notify_filtered(player, plyr, bbuf, 0)
									}
								}
							}
						}
					}
				default:
					if toparg != "" {
						match_args = toparg
					} else {
						match_args = ""
					}
					match_cmdname = "Queued event."
					if tmpfr := interp(descr, player, where, prog, trigger, BACKGROUND, STD_HARDUID, 0); tmpfr != nil {
						interp_loop(player, prog, tmpfr, false)
					}
				}
				propq_level--
			} else {
				notify_nolisten(player, "Propqueue stopped to prevent infinite loop.", true)
			}
		}
	}
	buf = propname
	if is_propdir(what, buf) {
		buf += "/"
		for name := next_prop_name(what, buf); name != ""; name = next_prop_name(what, buf) {
			propqueue(descr, player, where, trigger, what, xclude, name, toparg, mlev, mt)
		}
	}
}

func envpropqueue(descr int, player, where, trigger, what, xclude ObjectID, propname, toparg string, mlev, mt int) {
	for ; what != NOTHING; what = getparent(what) {
		propqueue(descr, player, where, trigger, what, xclude, propname, toparg, mlev, mt)
	}
}

func listenqueue(descr int, player, where, trigger, what, xclude ObjectID, propname, toparg string, mlev, mt, mpi_p int) {
	if DB.Fetch(what).IsFlagged(LISTENER) || DB.Fetch(DB.Fetch(what).Owner).IsFlagged(ZOMBIE) {
		var buf string

		/* queue up program referred to by the given property */
		prog := get_property_ObjectID(what, propname)
		tmpchar := get_property_class(what, propname)
		if prog != NOTHING || tmpchar != "" {
			if tmpchar != "" {
				if i := strings.Index(tmpchar, "="); i > -1 {
					buf = tmpchar[:i]
					if !smatch(buf, toparg) {
						tmpchar = tmpchar[i:]
					} else {
						tmpchar = ""
					}
				}
			}

			if tmpchar != "" || prog != NOTHING {
				switch {
				case tmpchar != "":
					switch {
					case tmpchar[0] == '&':
						prog = AMBIGUOUS
					case tmpchar[0] == NUMBER_TOKEN && unicode.IsNumber(tmpchar[1]):
						tmpchar = tmpchar[1:]
						prog = ObjectID(strconv.Atoi(tmpchar))
					case tmpchar[0] == REGISTERED_TOKEN:
						prog = find_registered_obj(what, tmpchar)
					case unicode.IsNumber(tmpchar[0]):
						prog = ObjectID(strconv.Atoi(tmpchar[0]))
					default:
						prog = NOTHING
					}
				case prog == AMBIGUOUS:
					prog = NOTHING
				}
				if prog != AMBIGUOUS {
					switch {
					case !prog.IsValid():
						prog = NOTHING
					case Typeof(prog) != TYPE_PROGRAM:
						prog = NOTHING
					case DB.Fetch(prog).Owner != DB.Fetch(player).Owner && !DB.Fetch(prog).IsFlagged(LINK_OK):
						prog = NOTHING
					case MLevel(prog) < mlev:
						prog = NOTHING
					case MLevel(DB.Fetch(prog).Owner) < mlev:
						prog = NOTHING
					case the_prog == xclude:
						prog = NOTHING
					}
				}
				switch prog {
				case NOTHING:
				case AMBIGUOUS:
					if mpi_p != nil {
						if mt != 0 {
							add_mpi_event(1, descr, player, where, trigger, tmpchar[1:], "Listen", toparg, 1, false, Prop_Blessed(what, propname))
						} else {
							add_mpi_event(1, descr, player, where, trigger, tmpchar[1:], "Olisten", toparg, 1, true, Prop_Blessed(what, propname))
						}
					}
				default:
					add_muf_queue_event(descr, player, where, trigger, prog, toparg, "(_Listen)", 1)
				}
			}
		}
		buf = propname
		if is_propdir(what, buf) {
			buf += "/"
			for name := next_prop_name(what, buf); name != ""; name = next_prop_name(what, buf) {
				listenqueue(descr, player, where, trigger, what, xclude, name, toparg, mlev, mt, mpi_p)
			}
		}
	}
}