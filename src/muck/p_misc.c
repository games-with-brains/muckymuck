package fbmuck

func prim_time(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(3)
		t := time.Now().Local()
		push(arg, top, t.Second())
		push(arg, top, t.Minute())
		push(arg, top, t.Hour())
	})
}

func prim_date(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(3)
		t := time.Now().Local()
		push(arg, top, t.Day())
		push(arg, top, t.Month())
		push(arg, top, t.Year())
	})
}

func prim_gmtoffset(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, get_tz_offset())
	})
}

func prim_systime(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, time.Now().Local())
	})
}

func prim_systime_precise(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, time.Now())
	})
}

func prim_timesplit(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		t := op[0].(time.Time).Local()
		CHECKOFLOW(8)
		push(arg, top, t.Second())
		push(arg, top, t.Minute())
		push(arg, top, t.Hour())
		push(arg, top, t.Day())
		push(arg, top, t.Month())
		push(arg, top, t.Year())
		push(arg, top, t.Weekday())
		push(arg, top, t.YearDay())
	})
}

func prim_timefmt(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		format := op[0].(string)
		time := op[1].(int)
		time_tm = localtime(&time)
		CHECKOFLOW(1)
		push(arg, top, format_time(format, time_tm))
	})
}

func prim_userlog(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(tp_userlog_mlev, mlev, 1, top, func(op Array) {
		log_user(player, program, fmt.Sprint(op[0].(string)))
	})
}

func prim_queue(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 3, top, func(op Array) {
		delay := op[0].(int)
		obj := op[1].(ObjectID).ValidObject()
		if Typeof(obj) != TYPE_PROGRAM {
			panic("Object must be a program. (2)")
		}
		command := op[2].(string)

		var temproom ObjectID
		if v, ok := (fr.variables + 1).data.(ObjectID); ok {
			temproom = v
		} else {
			temproom = DB.Fetch(player).Location
		}
		push(arg, top, add_muf_delayq_event(delay, fr.descr, player, temproom, NOTHING, obj, command, "Queued Event.", 0))
	})
}

func prim_kill(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		pid := op[0].(int)
		if pid == fr.pid {
			do_abort_silent()
			push(arg, top, 0)
		} else {
			if mlev < MASTER {
				if !control_process(ProgUID, pid) {
					panic("Permission Denied.")
				}
			}
			push(arg, top, MUFBool(dequeue_process(pid)))
		}
	})
}

func prim_force(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 2, top, func(op Array) {
		if fr.level > 8 {
			panic("Interp call loops not allowed.")
		}
		obj := op[0].(ObjectID)
		command := op[1].(string)
		switch {
		case !obj.IsValid():
			panic("Invalid object to force. (1)")
		case Typeof(obj) != TYPE_PLAYER && Typeof(obj) != TYPE_THING {
			panic("Object to force not a thing or player. (1)")
		case strings.Index(command, '\r') != -1:
			panic("Carriage returns not allowed in command string. (2).")
		case obj == GOD && DB.Fetch(program).Owner != GOD:
			panic("Cannot force god (1).")
		}
		ForceAction(program, func() {
			process_command(ObjectID_first_descr(obj), obj, command)
		})
		for i := 1; i <= fr.caller.top; i++ {
			if _, ok := fr.caller.st[i].(Program); !ok {
				do_abort_silent()
			}
		}
	})
}

func prim_timestamps(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		obj := op[0].(ObjectID).ValidRemoteObject(player, mlev)
		CHECKOFLOW(4)
		push(arg, top, DB.Fetch(obj).Created)
		push(arg, top, DB.Fetch(obj).Modified)
		push(arg, top, DB.Fetch(obj).LastUsed)
		push(arg, top, DB.Fetch(obj).Uses)
	})
}

extern int top_pid;
struct forvars *copy_fors(struct forvars *);
struct tryvars *copy_trys(struct tryvars *);

func prim_fork(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 0, top, func(op Array) {
		CHECKOFLOW(1)
		fr.pc = pc
		tmpfr := new(frame)
		tmpfr.system.top = fr.system.top
		for i := 0; i < fr.system.top; i++ {
			tmpfr.system.st[i] = fr.system.st[i]
		}
		tmpfr.argument.top = fr.argument.top
		for i := 0; i < fr.argument.top; i++ {
			tmpfr.argument.st[i] = fr.argument.st[i].Dup()
		}
		tmpfr.caller.top = fr.caller.top
		for i := 0; i <= fr.caller.top; i++ {
			tmpfr.caller.st[i] = fr.caller.st[i]
			if i > 0 {
				DB.Fetch(fr.caller.st[i]).(Program).instances++
			}
		}

		tmpfr.trys.top = fr.trys.top
		tmpfr.trys.st = copy_trys(fr.trys.st)

		tmpfr.forstack.top = fr.forstack.top
		tmpfr.forstack.st = copy_fors(fr.forstack.st)

		for i := 0; i < MAX_VAR; i++ {
			tmpfr.variables[i] = fr.variables[i].Dup()
		}

		localvar_dupall(tmpfr, fr)
		scopedvar_dupall(tmpfr, fr)

		tmpfr.error.is_flags = fr.error.is_flags
		if fr.rndbuf != nil {
			tmpfr.rndbuf = (void *) malloc(sizeof(unsigned long) * 4)
			memcpy(tmpfr.rndbuf, fr.rndbuf, 16)
		} else {
			tmpfr.rndbuf = nil
		}
		tmpfr.pc = pc
		tmpfr.pc++
		tmpfr.level = fr.level
		tmpfr.already_created = fr.already_created
		tmpfr.trig = fr.trig

		tmpfr.brkpt.breaknum = -1
		tmpfr.brkpt.lastproglisted = NOTHING;
		tmpfr.brkpt.count = 1
		tmpfr.brkpt.temp[0] = 1
		tmpfr.brkpt.level[0] = -1
		tmpfr.brkpt.line[0] = -1
		tmpfr.brkpt.linecount[0] = -2
		tmpfr.brkpt.pc[0] = nil
		tmpfr.brkpt.pccount[0] = -2
		tmpfr.brkpt.prog[0] = program

		tmpfr.pid = top_pid++
		tmpfr.multitask = BACKGROUND
		tmpfr.been_background = 1
		tmpfr.writeonly = 1
		tmpfr.skip_declare = fr.skip_declare
		tmpfr.wantsblanks = fr.wantsblanks
		tmpfr.perms = fr.perms
		tmpfr.descr = fr.descr

		/* child process gets a 0 returned on the stack */
		result = 0
		push(tmpfr.argument.st, &(tmpfr.argument.top), result)
		result = add_muf_delay_event(0, fr.descr, player, NOTHING, NOTHING, program, tmpfr, "BACKGROUND")

		/* parent process gets the child's pid returned on the stack */
		if result == 0 {
			result = -1
		}
		push(arg, top, result)
	})
}

func prim_pid(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, fr.pid)
	})
}

func prim_stats(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		obj := op[0].(ObjectID).ValidPlayer()
		var rooms, exits, things, players, programs int
		EachObject(func(obj ObjectID, o *Object) {
			if obj == NOTHING || o.Owner == obj {
				switch i.(type) {
				case TYPE_ROOM:
					rooms++
				case TYPE_EXIT:
					exits++
				case TYPE_THING:
					things++
				case TYPE_PLAYER:
					players++
				case TYPE_PROGRAM:
					programs++
				}
			}
		})
		CHECKOFLOW(6)
		push(arg, top, rooms + exits + things + players + programs)
		push(arg, top, rooms)
		push(arg, top, exits)
		push(arg, top, things)
		push(arg, top, programs)
		push(arg, top, players)
	}
}

func prim_abort(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		abort_interp(op[0].(string))
	})
}

func prim_ispidp(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		pid := op[0].(int)
		if pid == fr.pid {
			push(arg, top, 1)
		} else {
			if in_timequeue(pid) {
				push(arg, top, 1)
			} else {
				push(arg, top, 0)
			}
		}
	})
}

func prim_parselock(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		CHECKOFLOW(1)
		if lock := op[0].(string); lock == "" {
			push(arg, top, UNLOCKED)
		} else {
			push(arg, top, ParseLock(fr.descr, ProgUID, lock, 0))
		}
	})
}

func prim_unparselock(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if l := op[0].(Lock); l.IsTrue() {
			push(arg, top, "")
		} else {
			push(arg, top, l.Unparse(ProgUID, false))
		}
	})
}

func prim_prettylock(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, op[0].(Lock).Unparse(ProgUID, true))
	})
}

func prim_testlock(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		obj := op[0].(ObjectID).ValidRemoteObject(player, mlev)
		switch {
		case fr.level > 8:
			panic("Interp call loops not allowed.")
		case Typeof(obj) != TYPE_PLAYER && Typeof(obj) != TYPE_THING:
			panic("Invalid object type (1).")
		}
		push(arg, top, copy_bool(op[1].(Lock)).Eval(fr.descr, obj, player))
	})
}

func prim_sysparm(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		CHECKOFLOW(1)
		if parm := op[0].(string); parm != "" {
			if player == GOD {
				push(arg, top, Tuneables.GetAs(MLEV_GOD, parm))
			} else {
				push(arg, top, Tuneables.GetAs(MLEV_WIZARD, parm))
			}
		}
	})
}

func prim_cancallp(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		obj := op[0].(ObjectID).ValidObject()
		funcname := op[1].(string)
		if Typeof(obj) != TYPE_PROGRAM {
			panic("Object is not a MUF Program. (1)")
		}

		p := DB.Fetch(obj)
		if p.(Program).code == nil {
			tmpline := p.(Program).first
			p.(Program).first = read_program(obj)
			do_compile(-1, p.Owner, obj, 0)
			p.(Program).first = tmpline
		}

		result := 0
		if ProgMLevel(obj) > NON_MUCKER && (mlev >= WIZBIT || p.Owner == ProgUID || Linkable(obj)) {
			pbs := p.(Program).PublicAPI
			for ; pbs != nil && funcname != pbs.subname; pbs = pbs.next {}
			if pbs != nil && mlev >= pbs.mlev {
				result = 1
			}
		}
		CHECKOFLOW(1)
		push(arg, top, result)
	})
}

func prim_setsysparm(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 2, top, func(op Array) {
		if force_level != 0 {
			panic("Cannot be forced.")
		}
		parm := op[0].(string)
		security := TUNE_MLEV(player)
		oldvalue := Tuneables.GetAs(security, parm)
		newvalue := op[1].(string)
		switch result := Tuneables.SetAs(security, parm, newvalue); result {
		case TUNESET_SUCCESS:
			log_status("TUNED (MUF): %s(%d) tuned %s from '%s' to '%s'", DB.Fetch(player).name, player, parm, oldvalue, newvalue)
		case TUNESET_UNKNOWN:
			panic("Unknown parameter. (1)")
		case TUNESET_SYNTAX:
			panic("Bad parameter syntax. (2)")
		case TUNESET_BADVAL:
			panic("Bad parameter value. (2)")
		case TUNESET_DENIED:
			panic("Permission denied. (1)")
		}
	})
}

func prim_sysparm_array(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if player == GOD {
			push(arg, top, Tuneables.ArrayAs(MLEV_GOD, op[0].(string))
		} else {
			push(arg, top, Tuneables.ArrayAs(MLEV_WIZARD, op[0].(string)))
		}
	})
}

func prim_timer_start(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		if fr.timercount > tp_process_timer_limit {
			panic("Too many timers!")
		}
		timer := op[1].(string)
	    dequeue_timers(fr.pid, timer)
		add_muf_timer_event(fr.descr, player, program, fr, op[0].(int), timer)
	})
}

func prim_timer_stop(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
	    dequeue_timers(fr.pid, op[0].(string))
	})
}

func prim_event_exists(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, muf_event_exists(fr, op[0].(string)))
	})
}

func prim_event_count(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, muf_event_count(fr))
	})
}

func prim_event_send(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 3, top, func(op Array) {
		pid := op[0].(int)
		eventid := op[1].(string)

		var destfr frame
		if pid == fr.pid {
			destfr = fr
		} else {
			destfr = timequeue_pid_frame(pid)
		}

		if destfr != 0 {
			muf_event_add(destfr, fmt.Sprintf("USER.%.32s", eventid), Dictionary{
				"data": op[2],
				"caller_pid": fr.pid,
				"descr": fr.descr,
				"caller_prog": program,
				"trigger": fr.trig,
				"prog_uid": ProgUID,
				"player": player,
			}, 0)
		}
	})
}

func prim_pname_okp(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if ok_player_name(op[0].(string)) {
			push(arg, top, 1)
		} else {
			push(arg, top, 0)
		}
	})
}

func prim_name_okp(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		name := op[0].(string)
		if ok_ascii_other(name) && ok_name(name) {
			push(arg, top, 1)
		} else {
			push(arg, top, 0)
		}
	})
}

func prim_ext_name_okp(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		var ok1, ok2 func(string) bool
		switch v := op[0].(type) {
		case string:
			switch strings.ToLower(v) {
			case "e", "exit", "r", "room", "f", "muf", "program":
				ok1 = ok_ascii_other
				ok2 = ok_name
			case "t", "thing":
				ok1 = ok_ascii_thing
				ok2 = ok_name
			case "p", "player":
				ok1 = ok_player_name
			default:
				panic("String must be a valid object type (2)." )
			}
		case ObjectID:
			switch v.(ObjectID).ValidObject().(type) {
			case Exit, Room, Program:
				ok1 = ok_ascii_other
				ok2 = ok_name
			case Object:
				ok1 = ok_ascii_thing
				ok2 = ok_name
			case Player:
				ok1 = ok_player_name
			}
		default:
			panic("Dbref or object type name expected (2).");
		}
		result := ok1 != nil && ok1(op[1])
		if ok2 != nil && !result {
			result = ok2(op[1])
		}
		if result {
			push(arg, top, 1)
		} else {
			push(arg, top, 0)
		}
	})
}

func prim_force_level(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	CHECKOFLOW(1)
	push(arg, top, force_level)
}

func prim_watchpid(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		pid := op[0].(int)
		if pid == fr.pid {
			panic("Narcissistic processes not allowed.")
		}
		frame := timequeue_pid_frame(pid)
		if frame != nil {
			struct mufwatchpidlist **cur;
			struct mufwatchpidlist *waitee;

			for cur = &frame.waiters; *cur != nil; cur = &(*cur).next {
				if (*cur).pid == pid {
					break
				}
			}

			if *cur == nil {
				*cur = &mufwatchpidlist{ pid: fr.pid }
			} else {
				(*cur).pid = fr.pid
				(*cur).next = nil
			}
			waitee = &mufwatchpidlist{ next: fr.waitees, pid: pid }
			fr.waitees = waitee
		} else {
			muf_event_add(fr, fmt.Sprintf("PROC.EXIT.%d", pid), oper1, 0)
		}
	})
}

func prim_read_wants_blanks(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	fr.wantsblanks = true
}

func prim_debugger_break(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		b := fr.brkpt
		if b.count >= MAX_BREAKS {
			panic("Too many breakpoints set.")
		}
		b.force_debugging = true
		switch {
		case b.count != 1, b.temp[0] != 1, b.level[0] != -1, b.line[0] != -1, b.linecount[0] != -2, b.pc[0] != nil, b.pccount != -2, b.prog[0] != program:
			/* No initial breakpoint.  Lets make one. */
			b.count++
			i := b.count
			b.temp[i] = 1
			b.level[i] = -1
			b.line[i] = -1
			b.linecount[i] = -2
			b.pc[i] = nil
			b.pccount[i] = 0
			b.prog[i] = NOTHING
		}
	})
}

func prim_ignoringp(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 2, top, func(op Array) {
		push(arg, top, ignore_is_ignoring(op[0].(ObjectID).ValidObject(), op[1].(ObjectID).ValidObject()))
	})
}

func prim_ignore_add(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 2, top, func(op Array) {
		ignore_add_player(op[0].(ObjectID).ValidObject(), op[1].(ObjectID).ValidObject())
	})
}

func prim_ignore_del(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 2, top, func(op Array) {
		ignore_remove_player(op[0].(ObjectID).ValidObject(), op[1].(ObjectID).ValidObject())
	})
}

func prim_debug_on(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	DB.Fetch(program).FlagAs(DARK)
	DB.Fetch(program).Touch()
}

func prim_debug_off(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	DB.Fetch(program).ClearFlags(DARK)
	DB.Fetch(program).Touch()
}

func prim_debug_line(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	if !DB.Fetch(program).IsFlagged(DARK) && controls(player, program) {
		notify_nolisten(player, debug_inst(fr, 0, pc, fr.pid, arg, buf, sizeof(buf), *top, program), true)
	}
}