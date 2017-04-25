package fbmuck

/* Muf Interpreter and dispatcher. */

/* This package performs the interpretation of mud forth programs.
   It is a basically push pop kinda thing, but I'm making some stuff
   inline for maximum efficiency.

   Oh yeah, because it is an interpreted language, please do type
   checking during this time.  While you're at it, any objects you
   are referencing should be checked against db_top.
   */

/* in cases of boolean expressions, we do return a value, the stuff
   that's on top of a stack when a mud-forth program finishes executing.
   In cases where they don't leave a value, we just return a 0.  Note:
   this stuff does not return string or whatnot.  It at most can be
   relied on to return a boolean value.

   interp sets up a player's frames and so on and prepares it for
   execution.
   */

/* The static variable 'err' defined below means to die immediately when
 * set to this value. do_abort_silent() uses this.
 *
 * Otherwise err++ seems popular.
 */
const ERROR_DIE_NOW = -1

const (
	NON_MUCKER = iota
	APPRENTICE
	JOURNEYMAN
	MASTER
	WIZBIT
)

func pop_args(n int, top *int) (r Array) {
	checkop(n, top)
	r = make(Array, n)
	for i := n; i > 0; i-- {
		r[n] = POP().data
	}
	return
}

func apply_primitive(n int, top *int, f func(Array)) {
	defer func() {
		if x := recover(); x != nil {
			abort_interp(x)
		}
	}()
	f(pop_args(n, top))
}

func apply_restricted_primitive(l, mlev, n int, top *int, f func(Array)) {
	apply_primitive(n, top, func(op Array) {
		if mlev < l {
			panic(fmt.Sprintf("Mucker level %v or greater required.", l))
		}
		f(op)
	})
}

/* void    (*prim_func[]) (player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) = */
var prim_func = []func(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	p_null, p_null, p_null, p_null, p_null,  p_null,
	/* JMP, READ,   SLEEP,  CALL,   EXECUTE, RETURN, */
	p_null,           p_null, p_null,
	/* EVENT_WAITFOR, CATCH,  CATCH_DETAILED */

	PRIMS_CONNECTS_FUNCS,
	PRIMS_DB_FUNCS,
	PRIMS_MATH_FUNCS,
	PRIMS_MISC_FUNCS,
	PRIMS_PROPS_FUNCS,
	PRIMS_STACK_FUNCS,
	PRIMS_STRINGS_FUNCS,

	PRIMS_ARRAY_FUNCS,
	PRIMS_FLOAT_FUNCS,
	PRIMS_ERROR_FUNCS,
	PRIMS_MCP_FUNCS,
	PRIMS_REGEX_FUNCS,
	PRIMS_INTERNAL_FUNCS,
	NULL
}

func p_null(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	return
}

func localvars_get(struct frame *fr, ObjectID prog) (r *localvars) {
	if fr != nil {
		for r = fr.lvars; r != nil && r.prog != prog; r = r.next {}
		if r != nil {
			/* Pull this out of the middle of the stack. */
			r.prev = r.next
			if r.next != nil {
				r.next.prev = r.prev
			}
		} else {
			/* Create a new var frame. */
			r = new(localvars)
			r.prog = prog
			r.lvars = make([]inst, MAX_VAR)
		}
		r.next = fr.lvars
		r.prev = &fr.lvars
		fr.lvars = r
		if r.next != nil {
			r.next.prev = &r.next
		}
	}
	return
}

func localvar_dupall(fr, oldfr *frame) {
	if fr != nil && oldfr != nil {
		targ := &fr.lvars
		for orig := oldfr.lvars; orig != nil; orig = orig.next {
			*targ = &localvars{ prog: orig.prog, prev: targ }
			copy((*targ).lvars, orig.lvars)
			targ = &((*targ).next)
		}
	}
}

func scopedvar_addlevel(fr *frame, pc *inst) {
	proc := pc.data.(MUFProc)
	fr.svars = &Scope{ varnames: make([]string, len(proc.varnames)), vars: make([]*inst, len(proc.vars)), next: fr.svars }
	copy(fr.svars.varnames, proc.varnames)
}

func scopedvar_dupall(fr, oldfr *frame) {
	var cur, newsv *Scope
	var prev **Scope
	prev = &fr.svars
	for cur := oldfr.svars; cur != nil; cur = cur.next {
		newsv = &Scope{ varnames: make([]string, len(cur.varnames)), vars: make([]string, len(cur.vars)) }
		copy(newsv.varnames, cur.varnames)
		copy(newsv.vars, cur.vars)
		*prev = newsv
		prev = &newsv.next
	}
}

func scopedvar_get(fr *frame, int level, int varnum) (r *inst) {
	if fr != nil {
		svinfo := fr.svars
		for ; svinfo != nil && level > 0; level-- {
			svinfo = svinfo.next
		}
		if svinfo != nil && varnum > -1 && varnum < svinfo.count {
			r = svinfo.vars[varnum]
		}
	}
	return
}

func scopedvar_getname_byinst(pc *inst, varnum int) (r string) {
	var proc *MUFProc
	var ok bool
	for pc != nil {
		if proc, ok = pc.(MUFProc); ok {
			break
		}
		pc--
	}
	if proc != nil {
		if varnum > -1 && varnum < len(proc.varnames) {
			r = proc.varnames[varnum]
		}
	}
	return
}

func scopedvar_getname(fr *frame, level, varnum int) (r string) {
	if fr != nil {
		svinfo := fr.svars
		for ; svinfo != nil && level > 0; level-- {
			svinfo = svinfo.next
		}
		if svinfo != nil && varnum > -1 && varnum < svinfo.count {
			r = svinfo.varnames[varnum]
		}
	}
	return
}

func scopedvar_getnum(fr *frame, level int, varname string) (r int) {
	r = -1
	if varname != "" && fr != nil {
		svinfo := fr.svars
		for ; svinfo != nil && level > 0; level-- {
			svinfo = svinfo->next
		}
		if svinfo != nil && svinfo.varnames != nil {
			for i, v := range svinfo.varnames {
				if v == varname {
					r = i
					break
				}
			}
		}
	}
	return
}

func file_and_line() (f string, l int) {
	var ok bool
	if _, f, l, ok = runtime.Caller(1); !ok {
		panic()
	}
	return
}

int top_pid = 1;

struct forvars *for_pool = NULL;
struct forvars **last_for = &for_pool;
struct tryvars *try_pool = NULL;
struct tryvars **last_try = &try_pool;

func purge_for_pool() {
	/* This only purges up to the most recently used. */
	/* Purge this a second time to purge all. */
	var cur, next *forvars
	cur = *last_for
	*last_for = nil
	last_for = &for_pool

	for cur != nil {
		next = cur.next
		free(cur);
		cur = next;
	}
}

void
purge_try_pool(void)
{
	/* This only purges up to the most recently used. */
	/* Purge this a second time to purge all. */
	struct tryvars *cur, *next;

	cur = *last_try;
	*last_try = NULL;
	last_try = &try_pool;

	while (cur) {
		next = cur->next;
		free(cur);
		cur = next;
	}
}

func interp(int descr, ObjectID player, ObjectID location, ObjectID program, ObjectID source, int nosleeps, int whichperms, int forced_pid) *frame {
	if MLevel(program) == NON_MUCKER || MLevel(DB.Fetch(program).Owner) == NON_MUCKER || (source != NOTHING && !TrueWizard(DB.Fetch(source).Owner) && !can_link_to(DB.Fetch(source).Owner, TYPE_EXIT, program)) {
		notify_nolisten(player, "Program call: Permission denied.", true)
		return 0
	}
	fr := &frame{ descr: descr, multitask: nosleeps, perms: whichperms, been_background: nosleeps == 2, trig: source }
	if forced_pid != 0 {
		fr.pid = forced_pid
	} else {
		fr.pid = top_pid
		top_pid++
	}
	fr.caller.st[0] = source
	fr.caller.st[1] = program

	fr.system.top = 1
	fr.system.st[0].progref = 0
	fr.system.st[0].offset = 0

	fr.forstack.top = 0
	fr.forstack.st = nil
	fr.trys.top = 0
	fr.trys.st = nil

	fr.errorprog = NOTHING

	fr.rndbuf = nil
	fr.dlogids = nil

	fr.argument.top = 0
	fr.pc = DB.Fetch(program).(Program).start
	fr.writeonly = source == -1 || Typeof(source) == TYPE_ROOM || (Typeof(source) == TYPE_PLAYER && !online(source))) || DB.Fetch(player).flags & READMODE != 0
	fr.error.is_flags = 0

	/* set basic local variables */

	for i := 0; i < MAX_VAR; i++ {
		fr.variables[i].data = 0
	}

	fr.brkpt.force_debugging = 0
	fr.brkpt.debugging = 0
	fr.brkpt.bypass = 0
	fr.brkpt.isread = 0
	fr.brkpt.showstack = 0
	fr.brkpt.dosyspop = 0
	fr.brkpt.lastline = 0
	fr.brkpt.lastpc = 0
	fr.brkpt.lastlisted = 0
	fr.brkpt.lastcmd = ""
	fr.brkpt.breaknum = -1

	fr.brkpt.lastproglisted = NOTHING
	fr.brkpt.proglines = nil

	fr.brkpt.count = 1
	fr.brkpt.temp[0] = 1
	fr.brkpt.level[0] = -1
	fr.brkpt.line[0] = -1
	fr.brkpt.linecount[0] = -2
	fr.brkpt.pc[0] = nil
	fr.brkpt.pccount[0] = -2
	fr.brkpt.prog[0] = program

	fr.proftime.tv_sec = 0
	fr.proftime.tv_usec = 0
	fr.totaltime.tv_sec = 0
	fr.totaltime.tv_usec = 0

	fr.variables[0].data = player
	fr.variables[1].data = location
	fr.variables[2].data = source
	fr.variables[3].data = match_cmdname

	if DB.Fetch(program).(Program).code {
		DB.Fetch(program).(Program).profuses++
	}
	DB.Fetch(program).(Program).instances++
	push(fr.argument.st, &(fr.argument.top), match_args)
	return fr
}

static int err;
int already_created;

func copy_fors(forstack *forvars) (out *forvars) {
	var nu, last *forvars
	for in := forstack; in != nil; in = in.next {
		if !for_pool {
			nu = new(forvars)
		} else {
			nu = for_pool
			if *last_for == for_pool.next {
				last_for = &for_pool
			}
			for_pool = nu.next
		}
		nu.didfirst = in.didfirst
		nu.cur = in.cur
		nu.end = in.end
		nu.step = in.step
		nu.next = nil

		if out == nil {
			out = nu
			last = out
		} else {
			last.next = nu
			last = nu
		}
	}
	return
}

struct forvars *
push_for(struct forvars *forstack)
{
	struct forvars *nu;

	if (!for_pool) {
		nu = (struct forvars *) malloc(sizeof(struct forvars));
	} else {
		nu = for_pool;
		if (*last_for == for_pool->next) {
			last_for = &for_pool;
		}
		for_pool = nu->next;
	}
	nu->next = forstack;
	return nu;
}

struct forvars *
pop_for(struct forvars *forstack)
{
	struct forvars *newstack;

	if (!forstack) {
		return NULL;
	}
	newstack = forstack->next;
	forstack->next = for_pool;
	for_pool = forstack;
 	if (last_for == &for_pool) {
 		last_for = &(for_pool->next);
 	}
	return newstack;
}


struct tryvars *
copy_trys(struct tryvars *trystack)
{
	struct tryvars *in;
	struct tryvars *out = NULL;
	struct tryvars *nu;
	struct tryvars *last = NULL;

	for (in = trystack; in; in = in->next) {
		if (!try_pool) {
			nu = (struct tryvars*) malloc(sizeof(struct tryvars));
		} else {
			nu = try_pool;
			if (*last_try == try_pool->next) {
				last_try = &try_pool;
			}
			try_pool = nu->next;
		}

		nu->depth      = in->depth;
		nu->call_level = in->call_level;
		nu->for_count  = in->for_count;
		nu->addr       = in->addr;
		nu->next = NULL;

		if (!out) {
			last = out = nu;
		} else {
			last->next = nu;
			last = nu;
		}
	}
	return out;
}

struct tryvars *
push_try(struct tryvars *trystack)
{
	struct tryvars *nu;

	if (!try_pool) {
		nu = (struct tryvars*) malloc(sizeof(struct tryvars));
	} else {
		nu = try_pool;
		if (*last_try == try_pool->next) {
			last_try = &try_pool;
		}
		try_pool = nu->next;
	}
	nu->next = trystack;
	return nu;
}

func pop_try(trystack *tryvars) (r *tryvars) {
	if trystack != nil {
		r = trystack.next
		trystack.next = try_pool
		try_pool = trystack
	 	if last_try == &try_pool {
	 		last_try = &(try_pool.next)
	 	}
	}
	return
}

/* clean up lists from watchpid and sends event */
func watchpid_process(fr *frame) {
	if fr != nil {
		temp1 := inst{ data: fr.pid }
		for fr.waitees != nil {
			cur := fr.waitees
			fr.waitees = cur.next

			frame := timequeue_pid_frame(cur.pid)
			if frame != nil {
				for curptr := &frame.waiters; *curptr != nil; curptr = curptr.next {
					if curptr.pid == fr.pid {
						cur = *curptr
						*curptr = (*curptr).next
						break
					}
				}
			}
		}

		for fr.waiters != nil {
			buf := fmt.Sprintf("PROC.EXIT.%d", fr.pid)
			cur := fr.waiters
			fr.waiters = cur.next

			frame := timequeue_pid_frame(cur.pid)
			if frame != nil {
				muf_event_add(frame, buf, &temp1, 0)
				for curptr := &frame.waitees; *curptr != nil; curptr = &(*curptr).next {
					if (*curptr).pid == fr.pid {
						cur = *curptr
						*curptr = (*curptr).next
						break
					}
				}
			}
		}
	}
}

/* clean up the stack. */
func prog_clean(fr *frame) {
	if !fr {
		log_status("WARNING: prog_clean(): Tried to free a NULL frame !  Ignored.")
	} else {
		watchpid_process(fr)
		fr.system.top = 0
		for i := 0; i < fr.argument.top; i++{
			fr.argument.st[i] = nil
		}

		log.Printf("prog_clean: fr.caller.top=%d\n", fr.caller.top, 0)
		for i := 1; i <= fr.caller.top; i++ {
			log.Printf("Decreasing instances of fr.caller.st[%d](#%d)\n", i, fr.caller.st[i])
			DB.Fetch(fr.caller.st[i]).(Program).instances--
		}

		for i := 0; i < MAX_VAR; i++ {
			fr.variables[i] = nil
		}

		fr.lvars = nil
		for ; fr.svars != nil; fr.svars = fr.svars.next {}

		if fr.forstack.st != nil {
			struct forvars **loop = &(fr.forstack.st)

			for *loop != nil {
				loop = &((*loop).next)
			}
			*loop = for_pool
			if last_for == &for_pool {
				last_for = loop
			}
			for_pool = fr.forstack.st
			fr.forstack.st = nil
			fr.forstack.top = 0
		}

		if fr.trys.st != nil {
			struct tryvars **loop = &(fr.trys.st)

			for *loop != nil {
				loop = &((*loop).next)
			}
			*loop = try_pool
			if last_try == &try_pool {
				last_try = loop
			}
			try_pool = fr.trys.st
			fr.trys.st = nil
			fr.trys.top = 0
		}

		fr.argument.top = 0
		fr.pc = 0
		if fr.brkpt.lastcmd {
			free(fr.brkpt.lastcmd)
		}
		if fr.brkpt.proglines {
			fr.brkpt.proglines = nil
		}
		if fr.rndbuf {
			delete_seed(fr.rndbuf)
		}
		muf_dlog_purge(fr)
		dequeue_timers(fr.pid, nil)
		fr.events = nil
		err = 0
	}
}


func reload(fr *frame, atop, stop int) {
	fr.argument.top = atop
	fr.system.top = stop
}

func ValueIsFalse(v interface{}) (r bool) {
	switch v.(type) {
	case int:
		r = v == 0
	case float64:
		r = v == 0
	case string:
		r = len(v) == 0
	case Mark:
		r = true
	case stk_array:
		r = v.Len() == 0
	case Lock:
		r = v.IsTrue()
	case ObjectID:
		r = v == NOTHING
	}
	return
}

func (from *inst) Dup() (to *inst) {
	if from != nil {
		to = new(inst)
		switch data := from.data.(type) {
		case MUFProc:
		    if data != nil {
				to.data = &MUFProc { name: data.name, args: data.args, varnames: make([]string{}, len(data.varnames)) }
				copy(to.data.varnames, data.varnames)
			}
		case addr:
			DB.Fetch(from.data.(addr).progref).(Program).instances++
		case Lock:
		    if !data.IsTrue() {
				to.data = copy_bool(from.data)
			}
		}
	}
}

func copyvars(from, to *vars) {
	assert(from && to);
	for i := 0; i < MAX_VAR; i++ {
		(*to)[i] = (*from)[i]
	}
}

func calc_profile_timing(prog ObjectID, fr *frame) {
	tv := time.Now() - fr.proftime
	DB.Fetch(prog).(Program).proftime += tv
	fr.totaltime += tv
}

var interp_depth int

func do_abort_loop(player, program ObjectID, msg string, fr *frame, pc *inst, atop int, stop bool, clinst1, clinst2 *inst) {
	if !fr {
		panic("localvars_get(): NULL frame passed !");
	}
	if fr.trys.top != nil {
		fr.errorstr = msg
		if pc != nil {
			fr.errorinst = insttotext(fr, 0, pc, program, 1)
			fr.errorline = pc.line
		} else {
			fr.errorinst = nil
			fr.errorline = -1
		}
		fr.errorprog = program
		err++
	} else {
		if pc != nil {
			calc_profile_timing(program, fr)
		}
	}
	*clinst1 = nil
	*clinst2 = nil
	reload(fr, atop, stop)
	fr.pc = pc
	if fr.trys.top == nil {
		if pc != nil {
			interp_err(player, program, pc, fr.argument.st, fr.argument.top, fr.caller.st[1], insttotext(fr, 0, pc, program, 1), msg)
			if controls(player, program) {
				muf_backtrace(player, program, STACK_SIZE, fr)
			}
		} else {
			notify_nolisten(player, msg, true)
		}
		interp_depth--
		prog_clean(fr)
		DB.FetchPlayer(player).block = false
	}
}

func interp_loop(player, program ObjectID, fr *frame, has_return bool) (r *inst) {
	register struct inst *temp1;
	register struct inst *temp2;
	int i = 0, tmp, mlev;
	char dbuf[BUFFER_LEN];

	instno_debug_line := get_primitive("debug_line")
	interp_depth++
	fr.level = interp_depth		/* increment interp level */

	/* load everything into local stuff */
	pc := fr.pc
	atop := fr.argument.top
	stop := fr.system.top
	arg := fr.argument.st
	sys := fr.system.st
	writeonly := fr.writeonly
	already_created := false
	fr.brkpt.isread = false

	if pc == nil {
		tmpline := DB.Fetch(program).(Program).first
		DB.Fetch(program).(Program).first = (line*)(read_program(program))
		do_compile(-1, DB.Fetch(program).Owner, program, 0)
		DB.Fetch(program).(Program).first = tmpline
		fr.pc = DB.Fetch(program).(Program).start
		if pc = fr.pc; pc == nil {
			abort_loop_hard("Program not compilable. Cannot run.", nil, nil)
		}
		DB.Fetch(program).(Program).profuses++
		DB.Fetch(program).(Program).instances++
	}
	ts_useobject(program)
	err = 0

	instr_count := 0
	mlev = ProgMLevel(program)
	fr.proftime = time.Now()

	/* This is the 'natural' way to exit a function */
	for stop != nil {
		/* Abort program if player/thing running it is recycled */
		if !player.IsValid() || (!IsProgram(player) && !IsThing(player)) {
			reload(fr, atop, stop)
			prog_clean(fr)
			interp_depth--
			calc_profile_timing(program, fr)
			return nil
		}
		fr.instcnt++
		instr_count++

		if fr.multitask == PREEMPT || DB.Fetch(program).flags & BUILDER != 0 {
			if mlev >= WIZBIT {
				if tp_max_ml4_preempt_count {
					if instr_count >= tp_max_ml4_preempt_count {
						abort_loop_hard("Maximum preempt instruction count exceeded", NULL, NULL)
					}
				} else {
					instr_count = 0
				}
			} else {
				/* else make sure that the program doesn't run too long */
				if instr_count >= tp_max_instr_count {
					abort_loop_hard("Maximum preempt instruction count exceeded", NULL, NULL)
				}
			}
		} else {
			/* if in FOREGROUND or BACKGROUND mode, '0 sleep' every so often. */
			if (fr.instcnt > tp_instr_slice * 4) && (instr_count >= tp_instr_slice) {
				fr.pc = pc
				reload(fr, atop, stop)
				DB.FetchPlayer(player).block = !fr.been_background
				if fr.multitask == FOREGROUND {
					add_muf_delay_event(0, fr.descr, player, NOTHING, NOTHING, program, fr, "FOREGROUND")
				} else {
					add_muf_delay_event(0, fr.descr, player, NOTHING, NOTHING, program, fr, "BACKGROUND")
				}
				interp_depth--
				calc_profile_timing(program,fr)
				return nil
			}
		}
		fr.brkpt.debugging = (DB.Fetch(program).flags) & ZOMBIE != 0 || fr.brkpt.force_debugging) && !fr.been_background && controls(player, program)
		if DB.Fetch(program).flags & DARK != 0 || (fr.brkpt.debugging && fr.brkpt.showstack && !fr.brkpt.bypass) {
			if pc.(type) != PROG_PRIMITIVE || pc.data.(int) != instno_debug_line {
				notify_nolisten(player, debug_inst(fr, 0, pc, fr.pid, arg, dbuf, sizeof(dbuf), atop, program), true)
			}
		}
		if fr.brkpt.debugging {
			breakflag := false
			switch {
			case stop == 1 && !fr.brkpt.bypass && pc.(type) == PROG_PRIMITIVE && pc.data.(int) == IN_RET:
				/* Program is about to EXIT */
				notify_nolisten(player, "Program is about to EXIT.", true)
				breakflag = true
			case fr.brkpt.count != 0:
				for i := 0; i < fr.brkpt.count; i++ {
					if ((!fr->brkpt.pc[i] || pc == fr->brkpt.pc[i]) &&
								(fr->brkpt.line[i] == -1 || (fr->brkpt.lastline != pc->line && fr->brkpt.line[i] == pc->line)) &&
								(fr->brkpt.level[i] == -1 || stop <= fr->brkpt.level[i]) &&
								(fr->brkpt.prog[i] == NOTHING || fr->brkpt.prog[i] == program) &&
								(fr->brkpt.linecount[i] == -2 || (fr->brkpt.lastline != pc->line && fr->brkpt.linecount[i]-- <= 0)) &&
								(fr->brkpt.pccount[i] == -2 || (fr->brkpt.lastpc != pc && fr->brkpt.pccount[i]-- <= 0))) {
						if fr.brkpt.bypass {
							if fr.brkpt.pccount[i] == -1 {
								fr.brkpt.pccount[i] = 0
							}
							if fr.brkpt.linecount[i] == -1 {
								fr.brkpt.linecount[i] = 0
							}
						} else {
							breakflag = true
							break
						}
					}
				}
			}
			if breakflag {
				char *m;
				char buf[BUFFER_LEN];

				if fr.brkpt.dosyspop {
					stop--
					program = sys[stop].progref
					pc = sys[stop].offset
				}
				add_muf_read_event(fr.descr, player, program, fr)
				reload(fr, atop, stop)
				fr.pc = pc
				fr.brkpt.isread = false
				fr.brkpt.breaknum = i
				fr.brkpt.lastlisted = 0
				fr.brkpt.bypass = false
				fr.brkpt.dosyspop = false
				p := DB.Fetch(player)
				p.curr_prog = program
				p.block = false
				interp_depth--
				if !fr.brkpt.showstack {
					m = debug_inst(fr, 0, pc, fr.pid, arg, dbuf, sizeof(dbuf), atop, program)
					notify_nolisten(player, m, true)
				}
				if pc <= &(DB.Fetch(program).(Program).code[0]) || (pc - 1).line != pc.line {
					list_proglines(player, program, fr, pc.line, 0)
				} else {
					m = show_line_prims(fr, program, pc, 15, 1)
					buf = fmt.Sprintf("     %s", m)
					notify_nolisten(player, buf, true)
				}
				calc_profile_timing(program,fr)
				return nil
			}
			fr.brkpt.lastline = pc.line
			fr.brkpt.lastpc = pc
			fr.brkpt.bypass = false
		}
		switch mlev {
		case 0, 1:
			if fr.instcnt > tp_max_instr_count {
				abort_loop_hard("Maximum total instruction count exceeded.", nil, nil)
			}
		case 2:
			if fr.instcnt > tp_max_instr_count * 4 {
				abort_loop_hard("Maximum total instruction count exceeded.", nil, nil)
			}
		}
		switch pc.data.(type) {
		case int, float64, Address, ObjectID, PROG_VAR, PROG_LVAR, PROG_SVAR, string, Lock, Mark, stk_array:
			if atop >= STACK_SIZE {
				abort_loop("Stack overflow.", NULL, NULL)
			}
			arg[atop] = pc
			pc++
			atop++
		case PROG_LVAR_AT:
			switch {
			case atop >= STACK_SIZE:
				abort_loop("Stack overflow.", NULL, NULL)
			case pc.data.(int) >= MAX_VAR, pc.data.(int) < 0:
				abort_loop("Scoped variable number out of range.", NULL, NULL)
			default:
				lv := localvars_get(fr, program)
				tmp := &(lv.lvars[pc.data.(int)])
				arg[atop] = tmp
				pc++
				atop++
			}
		case PROG_LVAR_AT_CLEAR:
			switch {
			case atop >= STACK_SIZE:
				abort_loop("Stack overflow.", NULL, NULL)
			case pc.data.(int) >= MAX_VAR, pc.data.(int) < 0:
				abort_loop("Scoped variable number out of range.", NULL, NULL)
			default:
				lv := localvars_get(fr, program)
				tmp := &(lv.lvars[pc.data.(int)])
				arg[atop] = tmp
				pc++
				atop++
			}
		case PROG_LVAR_BANG:
			switch {
			case atop < 1:
				abort_loop("Stack Underflow.", NULL, NULL)
			case fr.trys.top && atop - fr.trys.st.depth < 1:
				abort_loop("Stack protection fault.", NULL, NULL)
			case pc.data.(int) >= MAX_VAR || pc.data.(int) < 0:
				abort_loop("Scoped variable number out of range.", NULL, NULL)
			default:
				lv := localvars_get(fr, program)
				the_var := &(lv.lvars[pc.data.(int)])
				atop--
				temp1 = arg + atop
				*the_var = *temp1
				pc++
			}
		case PROG_SVAR_AT:
			switch tmp := scopedvar_get(fr, 0, pc.data.(int)); {
			case tmp == nil:
				abort_loop("Scoped variable number out of range.", NULL, NULL)
			case atop >= STACK_SIZE:
				abort_loop("Stack overflow.", NULL, NULL)
			default:
				arg[atop] = tmp
				pc++
				atop++
			}
		case PROG_SVAR_AT_CLEAR:
			switch tmp := scopedvar_get(fr, 0, pc.data.(int)); {
			case tmp == nil:
				abort_loop("Scoped variable number out of range.", NULL, NULL)
			case atop >= STACK_SIZE:
				abort_loop("Stack overflow.", NULL, NULL)
			default:
				arg[atop] = tmp
				pc++
				atop++
			}
		case PROG_SVAR_BANG:
			switch the_var := scopedvar_get(fr, 0, pc.data.(int)); {
			case !the_var:
				abort_loop("Scoped variable number out of range.", NULL, NULL)
			case atop < 1:
				abort_loop("Stack Underflow.", NULL, NULL)
			case fr.trys.top != nil && atop - fr.trys.st.depth < 1:
				abort_loop("Stack protection fault.", NULL, NULL)
			default:
				atop--
				temp1 = arg + atop
				*the_var = *temp1
				pc++
			}
		case MUFProc:
			switch i := pc.args; {
			case atop < i:
				abort_loop("Stack Underflow.", nil, nil)
			case fr.trys.top && atop - fr.trys.st.depth < i:
				abort_loop("Stack protection fault.", nil, nil)
			default:
				if fr.skip_declare {
					fr.skip_declare = false
				} else {
					scopedvar_addlevel(fr, pc)
				}
				for ; i > 0; i-- {
					atop--
					temp1 = arg + atop
					if !scopedvar_get(fr, 0, i) {
						abort_loop_hard("Internal error: Scoped variable number out of range in FUNCTION init.", temp1, nil)
					}
				}
				pc++
			}
		case PROG_IF:
			switch {
			case atop < 1:
				abort_loop("Stack Underflow.", NULL, NULL)
			case fr.trys.top && atop - fr.trys.st.depth < 1:
				abort_loop("Stack protection fault.", NULL, NULL)
			default:
				atop--
				temp1 = arg + atop
				if ValueIsFalse(temp1) {
					pc = pc.data.call
				} else {
					pc++
				}
			}
		case PROG_EXEC:
			switch {
			case stop >= STACK_SIZE:
				abort_loop("System Stack Overflow", NULL, NULL)
			default:
				sys[stop].progref = program
				sys[stop].offset = pc + 1
				stop++
				pc = pc.data.call
				fr.skip_declare = false  /* Make sure we DON'T skip var decls */
			}
		case PROG_JMP:
			/* Don't need to worry about skipping scoped var decls here. */
			/* JMP to a function header can only happen in IN_JMP */
			pc = pc.data.call
		case PROG_TRY:
			switch {
			case atop < 1:
				abort_loop("Stack Underflow.", NULL, NULL)
			case fr.trys.top && atop - fr.trys.st.depth < 1:
				abort_loop("Stack protection fault.", NULL, NULL)
			default:
				atop--
				switch temp1 = (arg + atop).data.(int); {
				case temp1.data.(int) < 0:
					abort_loop("Argument is not a positive integer.", temp1, NULL)
				case fr.trys.top && atop - fr.trys.st.depth < temp1:
					abort_loop("Stack protection fault.", NULL, NULL)
				case temp1 > atop:
					abort_loop("Stack Underflow.", temp1, NULL)
				}
				fr.trys.top++
				fr.trys.st = push_try(fr.trys.st)
				fr.trys.st.depth = atop - temp1.data.(int)
				fr.trys.st.call_level = stop
				fr.trys.st.for_count = 0
				fr.trys.st.addr = pc.data.call
				pc++
			}
		case PROG_PRIMITIVE:
			/*
			 * All pc modifiers and stuff like that should stay here,
			 * everything else call with an independent dispatcher.
			 */
			switch pc.data.(int) {
			case IN_JMP:
				switch {
				case atop < 1:
					abort_loop("Stack underflow.  Missing address.", NULL, NULL)
				case fr.trys.top && atop - fr.trys.st.depth < 1:
					abort_loop("Stack protection fault.", NULL, NULL)
				default:
					atop--
					addr := (arg + atop).data.(Address)
					switch {
					case !addr.progref.IsValid(), !IsProgram(addr.progref):
						abort_loop_hard("Internal error.  Invalid address.", temp1, NULL)
					case program != addr.progref:
						abort_loop("Destination outside current program.", temp1, NULL)
					default:
						if _, ok := addr.data.(MUFProc); ok  {
							fr.skip_declare = true
						}
						pc = addr.data
					}
				}
			case IN_EXECUTE:
				switch {
				case atop < 1:
					abort_loop("Stack Underflow. Missing address.", NULL, NULL)
				case fr.trys.top && atop - fr.trys.st.depth < 1:
					abort_loop("Stack protection fault.", NULL, NULL)
				default:
					atop--
					addr := (arg + atop).data.(Address)
					switch {
					case !addr.progref.IsValid(), !IsProgram(addr.progref):
						abort_loop_hard("Internal error.  Invalid address.", temp1, NULL)
					case stop >= STACK_SIZE:
						abort_loop("System Stack Overflow", temp1, NULL)
					default:
						sys[stop].progref = program
						sys[stop].offset = pc + 1
						stop++
						if program != addr.progref {
							program = addr.progref
							fr.caller.top++
							fr.caller.st[fr.caller.top] = program
							mlev = ProgMLevel(program)
							DB.Fetch(program).(Program).instances++
						}
						pc = addr.data
					}
				}
			case IN_CALL:
				switch {
				case atop < 1:
					abort_loop("Stack Underflow. Missing ObjectID argument.", NULL, NULL)
				case fr.trys.top && atop - fr.trys.st.depth < 1:
					abort_loop("Stack protection fault.", NULL, NULL)
				default:
					atop--
					temp1 = arg + atop
					temp2 = ""
					if _, ok := temp1.(ObjectID); !ok {
						temp2 = temp1.data.(string)
						switch {
						case atop < 1:
							abort_loop("Stack Underflow. Missing ObjectID of func.", temp1, NULL)
						case fr.trys.top && atop - fr.trys.st.depth < 1:
							abort_loop("Stack protection fault.", NULL, NULL)
						}
						atop--
						temp1 = arg + atop
						if temp2 == "" {
							abort_loop("Null string not allowed. (2)", temp1, temp2)
						}
					}
					obj := temp1.(ObjectID).ValidObject()
					switch {
					case Typeof(obj) != TYPE_PROGRAM:
						abort_loop("Invalid object.", obj, temp2)
					}
					if DB.Fetch(obj).(Program).code == nil {
						tmpline := DB.Fetch(obj).(Program).first
						DB.Fetch(obj).(Program).first = read_program(obj)
						do_compile(-1, DB.Fetch(obj).Owner, obj, 0)
						DB.Fetch(obj).(Program).first = tmpline
						if DB.Fetch(obj).(Program).code == nil {
							abort_loop("Program not compilable.", obj, temp2)
						}
					}
					switch {
					case ProgMLevel(obj) == NON_MUCKER:
						abort_loop("Permission denied", obj, temp2)
					case mlev < WIZBIT && DB.Fetch(obj).Owner != ProgUID && !Linkable(obj):
						abort_loop("Permission denied", obj, temp2)
					case stop >= STACK_SIZE:
						abort_loop("System Stack Overflow", obj, temp2)
					}
					sys[stop].progref = program
					sys[stop].offset = pc + 1
					if temp2 == nil {
						pc = DB.Fetch(obj).(Program).start
					} else {
						for pbs := DB.Fetch(obj).(Program).PublicAPI; pbs != nil && temp2.data.(string) != pbs.subname; pbs = pbs.next {}
						switch {
						case pbs == nil:
							abort_loop("PUBLIC or WIZCALL function not found. (2)", temp2, temp2)
						case mlev < pbs.mlev:
							abort_loop("Insufficient permissions to call WIZCALL function. (2)", temp2, temp2)
						}
						pc = pbs.address.(*inst)
					}
					stop++
					if obj != program {
						calc_profile_timing(program, fr)
						fr.proftime = time.Now()
						program = obj
						fr.caller.top++
						fr.caller.st[fr.caller.top] = program
						DB.Fetch(program).(Program).instances++
						mlev = ProgMLevel(program)
					}
					DB.Fetch(program).(Program).profuses++
					ts_useobject(program)
				}
			case IN_RET:
				if stop > 1 && program != sys[stop - 1].progref {
					if !sys[stop - 1].progref.IsValid() || !IsProgram(sys[stop - 1].progref) {
						abort_loop_hard("Internal error.  Invalid address.", nil, nil)
					}
					calc_profile_timing(program, fr)
					fr.proftime = time.Now()
					DB.Fetch(program).(Program).instances--
					program = sys[stop - 1].progref
					mlev = ProgMLevel(program)
					fr.caller.top--
				}
				if fr.svars != nil {
					fr.svars = fr.svars.next
				}
				stop--
				pc = sys[stop].offset
			case IN_CATCH:
				if fr.trys.top == nil {
					abort_loop_hard("Internal error.  TRY stack underflow.", NULL, NULL)
				}
				depth := fr->trys.st->depth;
				for atop > depth {
					atop--
				}
				for ; fr.trys.st.for_count > 0; fr.trys.st.for_count-- {
					fr.forstack.top--
					fr.forstack.st = pop_for(fr.forstack.st)
				}
				fr.trys.top--
				fr.trys.st = pop_try(fr.trys.st)
				if fr.errorstr != "" {
					arg[atop].data = fr.errorstr
					atop++
					fr.errorstr = ""
				} else {
					arg[atop].data = ""
					atop++
				}
				fr.errorinst = nil
				reload(fr, atop, stop)
				pc++
			case IN_CATCH_DETAILED:
				if fr.trys.top == nil {
					abort_loop_hard("Internal error.  TRY stack underflow.", NULL, NULL)
				}
				depth := fr.trys.st.depth
				for atop > depth {
					atop--
				}
				for ; fr.trys.st.for_count > 0; fr.trys.st.for_count-- {
					fr.forstack.top--
					fr.forstack.st = pop_for(fr.forstack.st)
				}
				fr.trys.top--
				fr.trys.st = pop_try(fr.trys.st)
				nu := make(Dictionary)
				if fr.errorstr != "" {
					nu["error"] = fr.errorstr
					fr.errorstr = ""
				}
				if fr.errorinst != nil {
					nu["instr"] = fr.errorinst
					fr.errorinst = nil
				}
				nu["line"] = fr.errorline
				nu["program"] = fr.errorprog
				arg[atop].data = nu
				atop++
				reload(fr, atop, stop)
				pc++
			case IN_EVENT_WAITFOR:
				switch {
				case atop < 1:
					abort_loop("Stack Underflow. Missing eventID list array argument.", NULL, NULL)
				case fr.trys.top && atop - fr.trys.st.depth < 1:
					abort_loop("Stack protection fault.", NULL, NULL)
				default:
					atop--
					temp1 := arg + atop
					arr := temp1.data.(Array)
					if !array_is_homogenous(arr, "") {
						panic("Argument must be a list array of eventid strings.", temp1, NULL)
					}
					fr.pc = pc + 1
					reload(fr, atop, stop)
					events := make(Array, len(arr))
					for i, v := range arr {
						for j, _ := range events {
							if events[j] == v {
								events[i] = v
								break
							}
						}
					}
					muf_event_register_specific(player, program, fr, events...)
					DB.FetchPlayer(player).block = !fr.been_background
					interp_depth--
					calc_profile_timing(program, fr)
				}
				return nil
			case IN_READ:
				switch {
				case writeonly:
					abort_loop("Program is write-only.", NULL, NULL)
				case fr.multitask == BACKGROUND:
					abort_loop("BACKGROUND programs are write only.", NULL, NULL)
				default:
					reload(fr, atop, stop)
					fr.brkpt.isread = true
					fr.pc = pc + 1
					p := DB.FetchPlayer(player)
					p.curr_prog = program
					p.block = false
					add_muf_read_event(fr.descr, player, program, fr)
					interp_depth--
					calc_profile_timing(program, fr)
				}
				return nil
			case IN_SLEEP:
				switch {
				case atop < 1:
					abort_loop("Stack Underflow.", NULL, NULL)
				case fr.trys.top && atop - fr.trys.st.depth < 1:
					abort_loop("Stack protection fault.", NULL, NULL)
				default:
					atop--
					temp1 = (arg + atop).data.(int)
					fr.pc = pc + 1
					reload(fr, atop, stop)
					if temp1 < 0 {
						abort_loop("Timetravel beyond scope of muf.", temp1, NULL)
					}
					add_muf_delay_event(temp1, fr.descr, player, NOTHING, NOTHING, program, fr, "SLEEPING")
					DB.FetchPlayer(player).block = !fr.been_background
					interp_depth--
					calc_profile_timing(program, fr)
				}
				return nil
			default:
				reload(fr, atop, stop)
				tmp = atop
				prim_func[pc.data.(int) - 1](player, program, mlev, pc, arg, &tmp, fr)
				atop = tmp
				pc++
			}
		default:
			pc = nil
			abort_loop_hard("Program internal error. Unknown instruction type.", nil, nil)
		}						/* switch */
		if err != 0 {
			switch {
			case err == ERROR_DIE_NOW, fr.trys.top == nil:
				reload(fr, atop, stop)
				prog_clean(fr)
				DB.FetchPlayer(player).block = false
				interp_depth--
				calc_profile_timing(program, fr)
				return nil
			default:
				for fr.trys.st.call_level < stop {
					if stop > 1 && program != sys[stop - 1].progref {
						if !sys[stop - 1].progref.IsValid() || !IsProgram(sys[stop - 1].progref) {
							abort_loop_hard("Internal error.  Invalid address.", nil, nil)
						}
						calc_profile_timing(program, fr)
						fr.proftime = time.Now()
						DB.Fetch(program).(Program).instances--
						program = sys[stop - 1].progref
						mlev = ProgMLevel(program)
						fr.caller.top--
					}
					if fr.svars != nil {
						fr.svars = fr.svars.next
					}
					stop--
				}
				pc = fr.trys.st.addr
				err = 0
			}
		}
	}
	DB.FetchPlayer(player).block = false
	if atop == 0 {
		reload(fr, atop, stop)
		prog_clean(fr)
		interp_depth--
		calc_profile_timing(program,fr)
	} else {
		if has_return {
			retval := arg[atop - 1]
			r = &retval
		} else {
			if !ValueIsFalse(arg + atop - 1) {
				r = (struct inst *) 1
			} else {
				r = nil
			}
		}
		reload(fr, atop, stop)
		prog_clean(fr)
		interp_depth--
		calc_profile_timing(program,fr)
	}
	return
}

func interp_err(player, program ObjectID, pc, arg *inst, atop int, origprog ObjectID, msg1, msg2 string) {
	err++
	var buf string
	if DB.Fetch(origprog).Owner == DB.Fetch(player).Owner {
		buf = "Program Error.  Your program just got the following error."
		notify_nolisten(player, buf, true)
	} else {
		buf = fmt.Sprintf("Programmer Error.  Please tell %s what you typed, and the following message.", DB.Fetch(DB.Fetch(origprog).Owner).name)
		notify_nolisten(player, buf, true)
	}
	if pc != nil {
		buf = fmt.Sprintf("%s(#%d), line %d; %s: %s", DB.Fetch(program).name, program, pc.line, msg1, msg2)
		notify_nolisten(player, buf, true)
	} else {
		buf = fmt.Sprintf("%s(#%d), line %d; %s: %s", DB.Fetch(program).name, program, -1, msg1, msg2)
		notify_nolisten(player, buf, true)
	}
	lt := time(nil)
	tbuf := format_time("%c", localtime(&lt))
	errcount := get_property_value(origprog, ".debug/errcount")
	errcount++
	add_property(origprog, ".debug/errcount", nil, errcount)
	add_property(origprog, ".debug/lasterr", buf, 0)
	add_property(origprog, ".debug/lastcrash", nil, (int)lt)
	add_property(origprog, ".debug/lastcrashtime", tbuf, 0)
	if origprog != program {
		errcount = get_property_value(program, ".debug/errcount")
		errcount++
		add_property(program, ".debug/errcount", nil, errcount)
		add_property(program, ".debug/lasterr", buf, 0)
		add_property(program, ".debug/lastcrash", nil, (int)lt)
		add_property(program, ".debug/lastcrashtime", tbuf, 0)
	}
}

func push(stack *inst, top *int, res interface{}) {
	stack[*top].data = res;
	(*top)++;
}

func is_home(oper *inst) (r bool) {
	if v, ok := oper.data.(ObjectID); ok {
		r = v == HOME
	}
	return
}

func permissions(player, thing ObjectID) (ok bool) {
	if thing == player || thing == HOME {
		ok = true
	} else {
		switch t := DB.Fetch(thing); .(type) {
		case Exit:
			ok = t.Owner == DB.Fetch(player).Owner || t.Owner == NOTHING
		case Room, Object, Program:
			ok = t.Owner == DB.Fetch(player).Owner
		}
	}
	return
}

func find_mlev(prog ObjectID, fr *frame, st int) ObjectID {
	if DB.Fetch(prog).flags & STICKY != 0 && DB.Fetch(prog).flags & HAVEN != 0 {
		if st > 1 && TrueWizard(DB.Fetch(prog).Owner) {
			return find_mlev(fr.caller.st[st - 1], fr, st - 1)
		}
	}
	if MLevel(prog) < MLevel(DB.Fetch(prog).Owner) {
		return MLevel(prog)
	}
	return MLevel(DB.Fetch(prog).Owner)
}

func find_uid(player ObjectID, fr *frame, st int, program ObjectID) ObjectID {
	if DB.Fetch(program).flags & STICKY != 0 || fr.perms == STD_SETUID {
		if DB.Fetch(program).flags & HAVEN != 0 {
			if st > 1 && TrueWizard(DB.Fetch(program).Owner) {
				return find_uid(player, fr, st - 1, fr.caller.st[st - 1])
			}
			return DB.Fetch(program).Owner
		}
		return DB.Fetch(program).Owner
	}
	if ProgMLevel(program) < JOURNEYMAN {
		return DB.Fetch(program).Owner
	}
	if DB.Fetch(program).flags & HAVEN != 0 || fr.perms == STD_HARDUID {
		if fr.trig == NOTHING {
			return DB.Fetch(program).Owner
		}
		return DB.Fetch(fr.trig).Owner
	}
	return DB.Fetch(player).Owner
}

func do_abort_interp(player ObjectID, msg string, pc, arg *inst, atop int, fr *frame, program ObjectID, file string, line int) {
	if fr.trys.top {
		fr.errorstr = msg
		if pc != nil {
			fr.errorinst = insttotext(fr, 0, pc, program, 1)
			fr.errorline = pc.line
		} else {
			fr.errorinst = nil
			fr.errorline = -1
		}
		fr.errorprog = program
		err++
	} else {
		fr.pc = pc
		calc_profile_timing(program, fr)
		interp_err(player, program, pc, arg, atop, fr.caller.st[1], insttotext(fr, 0, pc, program, 1), msg)
		if controls(player, program) {
			muf_backtrace(player, program, STACK_SIZE, fr)
		}
	}
}

/*
 * Errors set with this will not be caught.
 *
 * This will always result in program termination the next time
 * interp_loop() checks for this.
 */
func do_abort_silent() {
	err = ERROR_DIE_NOW
}