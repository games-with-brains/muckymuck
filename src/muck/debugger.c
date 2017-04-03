func list_proglines(player, program dbref, fr *frame, start, end int) {
	var range []int
	if start == end || end == 0 {
		range += start
	} else {
		range += start
		range += end
	}
	if fr.brkpt.proglines == 0 || program != fr.brkpt.lastproglisted {
		fr.brkpt.proglines = read_program(program)
		fr.brkpt.lastproglisted = program
	}
	struct line *tmpline = db.Fetch(program).sp.(program_specific).first
	db.Fetch(program).sp.(program_specific).first = fr.brkpt.proglines
	tmpflg := db.Fetch(player).flags & INTERNAL != 0
	db.Fetch(player).flags |= INTERNAL
	do_list(player, program, range)
	if !tmpflg {
		db.Fetch(player).flags &= ~INTERNAL
	}
	db.Fetch(program).sp.(program_specific).first = tmpline
	return
}

func show_line_prims(fr *frame, program dbref, pc *inst, maxprims int, markpc bool) string {
	var maxback int
	var linestart, lineend *inst

	thisline := pc.line
	code := db.Fetch(program).sp.(program_specific).code
	end := code + len(code)

	for linestart, maxback = pc, maxprims; linestart > code && linestart.line == thisline && linestart.(type) != MUFProc && --maxback; --linestart {}
	if linestart.line < thisline {
		linestart++
	}

	for lineend, maxback = pc + 1, maxprims; lineend < end && lineend->line == thisline && lineend->type != MUFProc && --maxback; ++lineend {}
	if lineend >= end || lineend->line > thisline || lineend->type == MUFProc {
		lineend--
	}

	if lineend - linestart >= maxprims {
		if pc - (maxprims - 1) / 2 > linestart {
			linestart = pc - (maxprims - 1) / 2
		}
		if linestart + maxprims - 1 < lineend {
			lineend = linestart + maxprims - 1
		}
	}

	if linestart > code && (linestart - 1).line == thisline {
		buf = "..."
	}
	maxback = maxprims
	for linestart <= lineend {
		if buf != "" {
			buf += " "
		}
		if pc == linestart && markpc {
			buf += " {{"
			buf += insttotext(NULL, 0, linestart, program, 1)
			buf += "}} "
		} else {
			buf += insttotext(NULL, 0, linestart, program, 1)
		}
		linestart++
	}
	if lineend < end && (lineend + 1)->line == thisline {
		buf += " ..."
	}
	return
}

func funcname_to_pc(dbref program, const char *name) (r *inst) {
	for _, r = range db.Fetch(program).sp.(program_specific).code {
		if v, ok := r.data.(MUFProc); ok {
			if v.name == name {
				break
			}
		}
	}
	return
}

func linenum_to_pc(dbref program, int whatline) (r *inst) {
	for _, r = range db.Fetch(program).sp.(program_specific).code {
		if r.line == whatline {
			break
		}
	}
	return
}

func unparse_sysreturn(program *dbref, pc *inst) string {
	var ptr *inst
	for ptr = pc - 1; ptr >= db.Fetch(*program).sp.(program_specific).code; ptr-- {
		if _, ok := ptr.data.(MUFProc); ok {
			break
		}
	}
	var fname string
	if p, ok := ptr.data.(MUFProc); ok {
		fname = p.name
	} else {
		fname = "\033[1m???\033[0m"
	}
	return fmt.Sprintf("line \033[1m%d\033[0m, in \033[1m%s\033[0m", pc.line, fname)
}

func unparse_breakpoint(fr *frame, brk int) string {
	static char buf[BUFFER_LEN];
	char buf2[BUFFER_LEN];
	dbref ref;

	buf = fmt.Sprintf("%2d) break", brk + 1);
	if (fr->brkpt.line[brk] != -1) {
		buf2 = fmt.Sprintf(" in line %d", fr->brkpt.line[brk]);
		strcatn(buf, sizeof(buf), buf2);
	}
	if (fr->brkpt.pc[brk] != NULL) {
		ref = fr->brkpt.prog[brk];
		buf2 = fmt.Sprintf(" at %s", unparse_sysreturn(&ref, fr->brkpt.pc[brk] + 1));
		strcatn(buf, sizeof(buf), buf2);
	}
	if (fr->brkpt.linecount[brk] != -2) {
		buf2 = fmt.Sprintf(" after %d line(s)", fr->brkpt.linecount[brk]);
		strcatn(buf, sizeof(buf), buf2);
	}
	if (fr->brkpt.pccount[brk] != -2) {
		buf2 = fmt.Sprintf(" after %d instruction(s)", fr->brkpt.pccount[brk]);
		strcatn(buf, sizeof(buf), buf2);
	}
	if (fr->brkpt.prog[brk] != NOTHING) {
		buf2 = fmt.Sprintf(" in %s(#%d)", db.Fetch(fr.brkpt.prog[brk]).name, fr.brkpt.prog[brk]);
		strcatn(buf, sizeof(buf), buf2);
	}
	if (fr->brkpt.level[brk] != -1) {
		buf2 = fmt.Sprintf(" on call level %d", fr->brkpt.level[brk]);
		strcatn(buf, sizeof(buf), buf2);
	}
	return buf;
}

func muf_backtrace(player, program dbref, count int, fr *frame) {
	var buf, buf2, buf3 []string
	var ptr string
	ref dbref
	var j, cnt, flag int
	var pinst, lastinst *inst
	var lev int

	notify_nolisten(player, "\033[1;33;40mSystem stack backtrace:\033[0m", true)
	i := count;
	if i == 0 {
		i = STACK_SIZE
	}
	ref = program
	j = fr->system.top + 1;
	while (j > 1 && i-- > 0) {
		cnt = 0;
		do {
			lastinst = pinst;
			if (--j == fr->system.top) {
				pinst = fr->pc;
			} else {
				ref = fr->system.st[j].progref;
				pinst = fr->system.st[j].offset;
			}
			ptr = unparse_sysreturn(&ref, pinst);
			cnt++;
		} while (pinst == lastinst && j > 1);
		if (cnt > 1) {
			buf = fmt.Sprintf("     [repeats %d times]", cnt);
			notify_nolisten(player, buf, true);
		}
		lev = fr->system.top - j;
		if (ptr) {
			struct inst* fntop = fr->pc;
			struct inst* varinst;

			for fntop.(type) != MUFProc {
				fntop--
			}

			buf2 = fmt.Sprintf("%.512s\033[1m(\033[0m", ptr)
			for k := 0; k < fntop->data.(MUFProc)->args; k++ {
				const char* nam = scopedvar_getname(fr, lev, k);
				char* val;
				const char* fmt;
				if (!nam) {
					break;
				}
				varinst = scopedvar_get(fr, lev, k);
				val = insttotext(fr, lev, varinst, program, 1);
				if (k) {
					fmt = "\033[1m, %s=\033[0m%s";
				} else {
					fmt = "\033[1m%s=\033[0m%s";
				}
				buf2 = append(buf2, fmt.Sprintf(fmt, nam, val))
			}
			ptr = append(buf2, "\033[1m)\033[0m")
		}
		if pinst != lastinst {
			notify_nolisten(player, fmt.Sprintf("\033[1;33;40m%3d)\033[0m \033[1m%s(#%d)\033[0m %s:", lev, db.Fetch(ref).name, ref, ptr), true)
			flag := db.Fetch(player).flags & INTERNAL != 0
			db.Fetch(player).flags &= ~INTERNAL
			list_proglines(player, ref, fr, pinst.line, 0)
			if flag {
				db.Fetch(player).flags |= INTERNAL
			}
		}
	}
	notify_nolisten(player, "\033[1;33;40m*done*\033[0m", true)
}

func list_program_functions(player, program dbref, arg string) {
	notify_nolisten(player, "*function words*", true)
	for i, v := range db.Fetch(program).sp.(program_specific).code {
		if data, ok := v.(MUFProc); ok {
			if arg == "" || !smatch(arg, data.name) {
				notify_nolisten(player, data.name, true)
			}
		}
	}
	notify_nolisten(player, "*done*", true)
}

func debug_printvar(player, program dbref, fr *frame, arg string) {
	int i;
	char buf[BUFFER_LEN];

	if arg != "" {
		var lflag, sflag bool
		varnum := scopedvar_getnum(fr, 0, arg)
		if varnum != -1 {
			sflag = true
		} else {
			switch arg[0] {
			case 'L', 'l':
				arg = arg[1:]
				if arg[0] == 'V' || arg[0] == 'v' {
					arg = arg[1:]
				}
				lflag = true
				varnum = scopedvar_getnum(fr, 0, arg);
			case 'S', 's':
				arg = arg[1:]
				if arg[0] == 'V' || arg[0] == 'v' {
					arg = arg[1:]
				}
				sflag = true
			case 'V', 'v':
				arg = arg[1:]
			}
		}
		switch {
		case varnum > -1:
			i = varnum
		case unicode.IsNumber(arg):
			i = strconv.Atoi(arg)
		default:
			notify_nolisten(player, "I don't know which variable you mean.", true)
			return
		}
		switch {
		case i < 0, i >= MAX_VAR:
			notify_nolisten(player, "Variable number out of range.", true)
		case sflag:
			tmp := scopedvar_get(fr, 0, i)
			if tmp == nil {
				notify_nolisten(player, "Scoped variable number out of range.", true)
			} else {
				notify_nolisten(player, insttotext(fr, 0, tmp, -1, 1), true)
			}
		case lflag:
			lvars := localvars_get(fr, program)
			notify_nolisten(player, insttotext(fr, 0, &(lvars->lvars[i]), -1, 1), true)
		default:
			notify_nolisten(player, insttotext(fr, 0, &(fr->variables[i]), -1, 1), true)
		}
	} else {
		notify_nolisten(player, "I don't know which variable you mean.", true);
	}
}

func push_arg(player dbref, fr *frame, arg string) {
	var num int
	var inum float64

	switch {
	case fr.argument.top >= STACK_SIZE:
		notify_nolisten(player, "That would overflow the stack.", true)
	case unicode.IsNumber(arg):
		num = strconv.Atoi(arg)
		push(fr.argument.st, &fr.argument.top, num)
		notify_nolisten(player, "Integer pushed.", true)
	case ifloat(arg):
		inum = strconv.Atof(arg)
		push(fr.argument.st, &fr.argument.top, inum)
		notify_nolisten(player, "Float pushed.", true)
	case arg[0] == NUMBER_TOKEN:
		/* push a dbref */
		if !unicode.IsNumber(arg[1]) {
			notify_nolisten(player, "I don't understand that dbref.", true)
			return
		}
		num = strconv.Atoi(arg[1:])
		push(fr.argument.st, &fr.argument.top, num)
		notify_nolisten(player, "Dbref pushed.", true)
	case arg[0] == '"':
		buf := strings.SplitN(arg[1:], "\"", 1)[0]
		push(fr.argument.st, &fr.argument.top, buf)
		notify_nolisten(player, "String pushed.", true)
	default:
		var lflag, sflag bool
		varnum := scopedvar_getnum(fr, 0, arg)
		if varnum != -1 {
			sflag = true
		} else {
			switch arg[0] {
			case 'S', 's':
				arg = arg[1:]
				if arg[0] == 'V' || arg[0] == 'v' {
					arg = arg[1:]
				}
				sflag = true
				varnum = scopedvar_getnum(fr, 0, arg)
			case 'L', 'l':
				arg = arg[1:]
				if arg[0] == 'V' || arg[0] == 'v' {
					arg = arg[1:]
				}
				lflag = true
			case 'V', 'v':
				arg = arg[1:]
			}
		}
		switch {
		case varnum > -1:
			num = varnum
		case unicode.IsNumber(arg):
			num = strconv.Atoi(arg)
		default:
			notify_nolisten(player, "I don't understand what you want to push.", true)
			return
		}
		switch {
		case lflag:
			push(fr.argument.st, &fr.argument.top, num)
			notify_nolisten(player, "Local variable pushed.", true)
		case sflag:
			push(fr.argument.st, &fr.argument.top, num)
			notify_nolisten(player, "Scoped variable pushed.", true)
		default:
			push(fr.argument.st, &fr.argument.top, num)
			notify_nolisten(player, "Global variable pushed.", true)
		}
	}
}

struct inst primset[5];
static struct MUFProc temp_muf_proc_data = {
    "__Temp_Debugger_Proc",
	0,
	0,
	NULL
};

func muf_debugger(descr int, player, program dbref, text string, fr *frame) (r bool) {
	char buf2[BUFFER_LEN];
	char *ptr2
	struct inst *pinst;
	int i, j, cnt;

	cmd := strings.TrimSpace(text)
	if i := strings.IndexFunc(cmd, unicode.IsSpace); i != -1 {
		arg = cmd[i:]
		cmd = cmd[:i]
	}
	if cmd == "" && fr.brkpt.lastcmd != "" {
		cmd = fr.brkpt.lastcmd
	} else {
		fr.brkpt.lastcmd = cmd
	}
	/* delete triggering breakpoint, if it's only temp. */
	j = fr->brkpt.breaknum
	if j >= 0 && fr.brkpt.temp[j] {
		for j++; j < fr.brkpt.count; j++ {
			fr.brkpt.temp[j - 1] = fr.brkpt.temp[j]
			fr.brkpt.level[j - 1] = fr.brkpt.level[j]
			fr.brkpt.line[j - 1] = fr.brkpt.line[j]
			fr.brkpt.linecount[j - 1] = fr.brkpt.linecount[j]
			fr.brkpt.pc[j - 1] = fr.brkpt.pc[j]
			fr.brkpt.pccount[j - 1] = fr.brkpt.pccount[j]
			fr.brkpt.prog[j - 1] = fr.brkpt.prog[j]
		}
		fr.brkpt.count--
	}
	fr.brkpt.breaknum = -1

	switch cmd {
	case "cont":
	case "finish":
		if fr.brkpt.count >= MAX_BREAKS {
			notify_nolisten(player, "Cannot finish because there are too many breakpoints set.", true)
			add_muf_read_event(descr, player, program, fr)
		} else {
			j = fr.brkpt.count++
			fr.brkpt.temp[j] = 1
			fr.brkpt.level[j] = fr.system.top - 1
			fr.brkpt.line[j] = -1
			fr.brkpt.linecount[j] = -2
			fr.brkpt.pc[j] = nil
			fr.brkpt.pccount[j] = -2
			fr.brkpt.prog[j] = program
			fr.brkpt.bypass = 1
		}
	case "stepi":
		if i = strconv.Atoi(arg); i == 0 {
			i = 1
		}
		if fr.brkpt.count >= MAX_BREAKS {
			notify_nolisten(player, "Cannot stepi because there are too many breakpoints set.", true)
			add_muf_read_event(descr, player, program, fr)
		} else {
			j = fr.brkpt.count++
			fr.brkpt.temp[j] = 1
			fr.brkpt.level[j] = -1
			fr.brkpt.line[j] = -1
			fr.brkpt.linecount[j] = -2
			fr.brkpt.pc[j] = nil
			fr.brkpt.pccount[j] = i
			fr.brkpt.prog[j] = NOTHING
			fr.brkpt.bypass = 1
		}
	case "step"
		if i = strconv.Atoi(arg); i == 0 {
			i = 1
		}
		if fr.brkpt.count >= MAX_BREAKS {
			notify_nolisten(player, "Cannot step because there are too many breakpoints set.", true)
			add_muf_read_event(descr, player, program, fr)
		} else {
			j = fr.brkpt.count++
			fr.brkpt.temp[j] = 1
			fr.brkpt.level[j] = -1
			fr.brkpt.line[j] = -1
			fr.brkpt.linecount[j] = i
			fr.brkpt.pc[j] = nil
			fr.brkpt.pccount[j] = -2
			fr.brkpt.prog[j] = NOTHING
			fr.brkpt.bypass = 1
		}
	case "nexti":
		if i = strconv.Atoi(arg); i == 0 {
			i = 1
		}
		if fr.brkpt.count >= MAX_BREAKS {
			notify_nolisten(player, "Cannot nexti because there are too many breakpoints set.", true)
			add_muf_read_event(descr, player, program, fr)
		} else {
			j = fr.brkpt.count++
			fr.brkpt.temp[j] = 1
			fr.brkpt.level[j] = fr.system.top
			fr.brkpt.line[j] = -1
			fr.brkpt.linecount[j] = -2
			fr.brkpt.pc[j] = nil
			fr.brkpt.pccount[j] = i
			fr.brkpt.prog[j] = program
			fr.brkpt.bypass = 1
		}
	case "next":
		if i = strconv.Atoi(arg); i == 0 {
			i = 1
		}
		if fr.brkpt.count >= MAX_BREAKS {
			notify_nolisten(player, "Cannot next because there are too many breakpoints set.", true)
			add_muf_read_event(descr, player, program, fr)
		} else {
			j = fr.brkpt.count++
			fr.brkpt.temp[j] = 1
			fr.brkpt.level[j] = fr.system.top
			fr.brkpt.line[j] = -1
			fr.brkpt.linecount[j] = i
			fr.brkpt.pc[j] = nil
			fr.brkpt.pccount[j] = -2
			fr.brkpt.prog[j] = program
			fr.brkpt.bypass = 1
		}
	case "exec":
		if fr.brkpt.count >= MAX_BREAKS {
			notify_nolisten(player, "Cannot finish because there are too many breakpoints set.", true)
			add_muf_read_event(descr, player, program, fr)
		} else {
			switch pinst = funcname_to_pc(program, arg); {
			case pinst == nil:
				notify_nolisten(player, "I don't know a function by that name.", true)
				add_muf_read_event(descr, player, program, fr)
			case fr.system.top >= STACK_SIZE:
				notify_nolisten(player, "That would exceed the system stack size for this program.", true)
				add_muf_read_event(descr, player, program, fr)
			default:
				fr.system.st[fr.system.top].progref = program
				fr.system.st[fr.system.top].offset = fr.pc
				top++
				fr.pc = pinst
				j = fr.brkpt.count++
				fr.brkpt.temp[j] = 1
				fr.brkpt.level[j] = fr.system.top - 1
				fr.brkpt.line[j] = -1
				fr.brkpt.linecount[j] = -2
				fr.brkpt.pc[j] = nil
				fr.brkpt.pccount[j] = -2
				fr.brkpt.prog[j] = program
				fr.brkpt.bypass = 1
			}
		}
	case "prim":
		switch {
		case fr.brkpt.count >= MAX_BREAKS:
			notify_nolisten(player, "Cannot finish because there are too many breakpoints set.", true)
			add_muf_read_event(descr, player, program, fr)
		case !primitive(arg):
			notify_nolisten(player, "I don't recognize that primitive.", true)
			add_muf_read_event(descr, player, program, fr)
		case fr.system.top >= STACK_SIZE:
			notify_nolisten(player, "That would exceed the system stack size for this program.", true)
			add_muf_read_event(descr, player, program, fr)
		default:
			primset[0].line = 0
			primset[0].data.(MUFProc) = &temp_muf_proc_data
			primset[0].data.(MUFProc).vars = 0
			primset[0].data.(MUFProc).args = 0
			primset[0].data.(MUFProc).varnames = nil
			primset[1].line = 0;
			primset[1].data = get_primitive(arg)
			primset[2].line = 0
			primset[2].data = IN_RET

			fr.system.st[fr.system.top].progref = program
			fr.system.st[fr.system.top++].offset = fr.pc
			fr.pc = &primset[1]
			j = fr.brkpt.count++
			fr.brkpt.temp[j] = 1
			fr.brkpt.level[j] = -1
			fr.brkpt.line[j] = -1
			fr.brkpt.linecount[j] = -2
			fr.brkpt.pc[j] = &primset[2]
			fr.brkpt.pccount[j] = -2
			fr.brkpt.prog[j] = program
			fr.brkpt.bypass = 1
			fr.brkpt.dosyspop = 1
		}
	case "break":
		add_muf_read_event(descr, player, program, fr)
		if fr.brkpt.count >= MAX_BREAKS {
			notify_nolisten(player, "Too many breakpoints set.", true)
		} else {
			if unicode.IsNumber(arg) {
				i = strconv.Atoi(arg)
			} else {
				if pinst = funcname_to_pc(program, arg); pinst == 0 {
					notify_nolisten(player, "I don't know a function by that name.", true)
					return 0
				} else {
					i = pinst.line
				}
			}
			if i == 0 {
				i = fr.pc.line
			}
			j = fr.brkpt.count++
			fr.brkpt.temp[j] = 0
			fr.brkpt.level[j] = -1
			fr.brkpt.line[j] = i
			fr.brkpt.linecount[j] = -2
			fr.brkpt.pc[j] = nil
			fr.brkpt.pccount[j] = -2
			fr.brkpt.prog[j] = program
			notify_nolisten(player, "Breakpoint set.", true)
		}
	case "delete":
		add_muf_read_event(descr, player, program, fr)
		switch i = strconv.Atoi(arg); {
		case i == 0:
			notify_nolisten(player, "Which breakpoint did you want to delete?", true)
		case i < 1 || i > fr.brkpt.count:
			notify_nolisten(player, "No such breakpoint.", true)
		default:
			j = i - 1
			for j++; j < fr.brkpt.count; j++ {
				fr.brkpt.temp[j - 1] = fr.brkpt.temp[j]
				fr.brkpt.level[j - 1] = fr.brkpt.level[j]
				fr.brkpt.line[j - 1] = fr.brkpt.line[j]
				fr.brkpt.linecount[j - 1] = fr.brkpt.linecount[j]
				fr.brkpt.pc[j - 1] = fr.brkpt.pc[j]
				fr.brkpt.pccount[j - 1] = fr.brkpt.pccount[j]
				fr.brkpt.prog[j - 1] = fr.brkpt.prog[j]
			}
			fr.brkpt.count--
			notify_nolisten(player, "Breakpoint deleted.", true)
		}
	case "breaks":
		notify_nolisten(player, "Breakpoints:", true)
		for i := 0; i < fr.brkpt.count; i++ {
			notify_nolisten(player, unparse_breakpoint(fr, i), true)
		}
		notify_nolisten(player, "*done*", true)
		add_muf_read_event(descr, player, program, fr)
	case "where":
		i = strconv.Atoi(arg)
		muf_backtrace(player, program, i, fr)
		add_muf_read_event(descr, player, program, fr)
	case "stack":
		notify_nolisten(player, "*Argument stack top*", true)
		i = strconv.Atoi(arg)
		if i == 0 {
			i = STACK_SIZE
		}
		var ptr string
		for j := fr.argument.top; j > 0 && i > 0; i-- {
			cnt = 0
			do {
				buf = ptr
				j--
				ptr = insttotext(nil, 0, &fr.argument.st[j], program, 1)
				cnt++
			} while ptr == buf && j > 0
			if cnt > 1 {
				notify_fmt(player, "     [repeats %d times]", cnt)
			}
			if ptr != buf {
				notify_fmt(player, "%3d) %s", j + 1, ptr)
			}
		}
		notify_nolisten(player, "*done*", true)
		add_muf_read_event(descr, player, program, fr)
	case "list", "listi":
		var startline, endline int
		add_muf_read_event(descr, player, program, fr)
		if ptr2 = (char *) strchr(arg, ','); ptr != nil {
			*ptr2++ = '\0';
		} else {
			ptr2 = ""
		}
		if *arg != nil {
			if fr.brkpt.lastlisted {
				startline = fr.brkpt.lastlisted + 1
			} else {
				startline = fr.pc.line
			}
			endline = startline + 15
		} else {
			if !unicode.IsNumber(arg) {
				if pinst = funcname_to_pc(program, arg); pinst == nil {
					notify_nolisten(player, "I don't know a function by that name. (starting arg, 1)", true)
					return 0
				} else {
					startline = pinst.line
					endline = startline + 15
				}
			} else {
				if *ptr2 != nil {
					endline = startline = atoi(arg)
				} else {
					startline = atoi(arg) - 7
					endline = startline + 15
				}
			}
		}
		if *ptr2 != nil {
			if !unicode.IsNumber(ptr2) {
				if pinst = funcname_to_pc(program, ptr2); pinst == nil {
					notify_nolisten(player, "I don't know a function by that name. (ending arg, 1)", true)
					return 0
				} else {
					endline = pinst.line
				}
			} else {
				endline = strconv.Atoi(ptr2)
			}
		}
		p := db.Fetch(program).sp.program
		if i = (p.sp.code + len(p.sp.code) - 1).line; startline > i {
			notify_nolisten(player, "Starting line is beyond end of program.", true)
		} else {
			if startline < 1 {
				startline = 1
			}
			if endline > i {
				endline = i
			}
			if endline < startline {
				endline = startline
			}
			notify_nolisten(player, "Listing:", true)
			if cmd == "listi" {
				for i := startline; i <= endline; i++ {
					if pinst = linenum_to_pc(program, i); pinst != nil {
						if i == fr.pc.line {
							notify_nolisten(player, fmt.Sprintf("line %d: %s", i, show_line_prims(fr, program, fr.pc, STACK_SIZE, 1)), true)
						} else {
							notify_nolisten(player, fmt.Sprintf("line %d: %s", i, show_line_prims(fr, program, pinst, STACK_SIZE, 0)), true)
						}
					}
				}
			} else {
				list_proglines(player, program, fr, startline, endline)
			}
			fr.brkpt.lastlisted = endline
			notify_nolisten(player, "*done*", true)
		}
	case "quit":
		notify_nolisten(player, "Halting execution.", true)
		return 1
	case "trace":
		add_muf_read_event(descr, player, program, fr)
		switch arg {
		case "on":
			fr.brkpt.showstack = true
			notify_nolisten(player, "Trace turned on.", true)
		case "off":
			fr.brkpt.showstack = false
			notify_nolisten(player, "Trace turned off.", true)
		default:
			notify_nolisten(player, fmt.Sprintf("Trace is currently %s.", fr.brkpt.showstack ? "on" : "off"), true)
		}
	case "words":
		list_program_functions(player, program, arg)
		add_muf_read_event(descr, player, program, fr)
	case "print":
		debug_printvar(player, program, fr, arg)
		add_muf_read_event(descr, player, program, fr)
	case "push":
		push_arg(player, fr, arg)
		add_muf_read_event(descr, player, program, fr)
	case "pop":
		add_muf_read_event(descr, player, program, fr)
		if fr.argument.top < 1 {
			notify_nolisten(player, "Nothing to pop.", true)
			return 0
		}
		fr.argument.top--
		notify_nolisten(player, "Stack item popped.", true)
	case "help":
		notify_nolisten(player, "cont            continues execution until a breakpoint is hit.", true)
		notify_nolisten(player, "finish          completes execution of current function.", true)
		notify_nolisten(player, "step [NUM]      executes one (or NUM, 1) lines of muf.", true)
		notify_nolisten(player, "stepi [NUM]     executes one (or NUM, 1) muf instructions.", true)
		notify_nolisten(player, "next [NUM]      like step, except skips CALL and EXECUTE.", true)
		notify_nolisten(player, "nexti [NUM]     like stepi, except skips CALL and EXECUTE.", true)
		notify_nolisten(player, "break LINE#     sets breakpoint at given LINE number.", true)
		notify_nolisten(player, "break FUNCNAME  sets breakpoint at start of given function.", true)
		notify_nolisten(player, "breaks          lists all currently set breakpoints.", true)
		notify_nolisten(player, "delete NUM      deletes breakpoint by NUM, as listed by 'breaks'", true)
		notify_nolisten(player, "where [LEVS]    displays function call backtrace of up to num levels deep.", true)
		notify_nolisten(player, "stack [NUM]     shows the top num items on the stack.", true)
		notify_nolisten(player, "print v#        displays the value of given global variable #.", true)
		notify_nolisten(player, "print lv#       displays the value of given local variable #.", true)
		notify_nolisten(player, "trace [on|off]  turns on/off debug stack tracing.", true)
		notify_nolisten(player, "list [L1,[L2]]  lists source code of given line range.", true)
		notify_nolisten(player, "list FUNCNAME   lists source code of given function.", true)
		notify_nolisten(player, "listi [L1,[L2]] lists instructions in given line range.", true)
		notify_nolisten(player, "listi FUNCNAME  lists instructions in given function.", true)
		notify_nolisten(player, "words           lists all function word names in program.", tue)
		notify_nolisten(player, "words PATTERN   lists all function word names that match PATTERN.", true)
		notify_nolisten(player, "exec FUNCNAME   calls given function with the current stack data.", true)
		notify_nolisten(player, "prim PRIMITIVE  executes given primitive with current stack data.", true)
		notify_nolisten(player, "push DATA       pushes an int, dbref, var, or string onto the stack.", true)
		notify_nolisten(player, "pop             pops top data item off the stack.", true)
		notify_nolisten(player, "help            displays this help screen.", true)
		notify_nolisten(player, "quit            stop execution here.", true)
		add_muf_read_event(descr, player, program, fr)
	default:
		notify_nolisten(player, "I don't understand that debugger command. Type 'help' for help.", true)
		add_muf_read_event(descr, player, program, fr)
	}
	return 0
}