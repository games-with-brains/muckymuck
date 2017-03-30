#define abort_loop(S, C1, C2) { \
	do_abort_loop(player, program, (S), fr, pc, atop, stop, (C1), (C2)) \
	if fr != nil && fr.trys.top { \
		break \
	} else { \
		return 0 \
	} \
}

#define abort_loop_hard(S, C1, C2) { \
	int tmp = 0 \
	if fr != nil { \
		tmp = fr.trys.top \
		fr.trys.top = 0 \
	} \
	do_abort_loop(player, program, (S), fr, pc, atop, stop, (C1), (C2)) \
	if fr != nil { \
		fr.trys.top = tmp \
	} \
	return 0 \
}

#define POP() (arg + --(*top))

func abort_interp(msg string) {
	do_abort_interp(player, msg, pc, arg, *top, fr, program, __FILE__, __LINE__)
}

func checkop_readonly(nargs int, top *int) {
	if *top < nargs {
		abort_interp("Stack underflow.")
	}
}

func checkop(nargs int, top *int) {
	checkop_readonly(nargs, top)
	if fr.trys.top && *top - fr.trys.st.depth < nargs  {
		abort_interp("Stack protection fault.")
	}
}

#define Min(x,y) ((x < y) ? x : y)
#define ProgMLevel(x) (find_mlev(x, fr, fr.caller.top))

#define ProgUID find_uid(player, fr, fr.caller.top, program)

func MUFBool(x bool) (r int) {
	if x {
		r = 1
	}
	return
}

#define CHECKOFLOW(x) if((*top + (x - 1)) >= STACK_SIZE) \
			  abort_interp("Stack Overflow!");

#define SORTTYPE_CASEINSENS     0x1
#define SORTTYPE_DESCENDING     0x2

#define SORTTYPE_CASE_ASCEND    0
#define SORTTYPE_NOCASE_ASCEND  (SORTTYPE_CASEINSENS)
#define SORTTYPE_CASE_DESCEND   (SORTTYPE_DESCENDING)
#define SORTTYPE_NOCASE_DESCEND (SORTTYPE_CASEINSENS | SORTTYPE_DESCENDING)
#define SORTTYPE_SHUFFLE        4