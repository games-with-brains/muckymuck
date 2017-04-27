package fbmuck

/* This file contains code for doing "byte-compilation" of
   mud-forth programs.  As such, it contains many internal
   data structures and other such which are not found in other
   parts of TinyMUCK.                                       */

/* The CONTROL_STACK is a stack for holding previous control statements.
   This is used to resolve forward references for IF/THEN and loops, as well
   as a placeholder for back references for loops. */

#define CTYPE_IF    1
#define CTYPE_ELSE  2
#define CTYPE_BEGIN 3
#define CTYPE_FOR   4			/* Get it?  CTYPE_FOUR!!  HAHAHAHAHA  -Fre'ta */
								/* C-4?  *BOOM!*  -Revar */
#define CTYPE_WHILE 5
#define CTYPE_TRY   6			/* reserved for exception handling */
#define CTYPE_CATCH 7			/* reserved for exception handling */


/* These would be constants, but their value isn't known until runtime. */
static int IN_FORITER;
static int IN_FOREACH;
static int IN_FORPOP;
static int IN_FOR;
static int IN_TRYPOP;

var primitive_list map[string] PROG_PRIMITIVE
func init() {
	primitive_list = make(map[string] PROG_PRIMITIVE)
}

struct CONTROL_STACK {
	short type;
	struct INTERMEDIATE *place;
	struct CONTROL_STACK *next;
	struct CONTROL_STACK *extra;
};

/* This structure is an association list that contains both a procedure
   name and the place in the code that it belongs.  A lookup to the procedure
   will see both it's name and it's number and so we can generate a
   reference to it.  Since I want to disallow co-recursion,  I will not allow
   forward referencing.
   */

struct PROC_LIST {
	const char *name;
	int returntype;
	struct INTERMEDIATE *code;
	struct PROC_LIST *next;
};

/* The intermediate code is code generated as a linked list
   when there is no hint or notion of how much code there
   will be, and to help resolve all references.
   There is always a pointer to the current word that is
   being compiled kept.
   */

#define INTMEDFLG_DIVBYZERO 1
#define INTMEDFLG_MODBYZERO 2
#define INTMEDFLG_INTRY		4

struct INTERMEDIATE {
	int no;						/* which number instruction this is */
	struct inst in;				/* instruction itself */
	short line;					/* line number of instruction */
	short flags;
	struct INTERMEDIATE *next;	/* next instruction */
};


/* The state structure for a compile. */
typedef struct COMPILE_STATE_T {
	struct CONTROL_STACK *control_stack;
	struct PROC_LIST *procs;

	int nowords;				/* number of words compiled */
	struct INTERMEDIATE *curr_word;	/* word being compiled */
	struct INTERMEDIATE *first_word;	/* first word of the list */
	struct INTERMEDIATE *curr_proc;	/* first word of curr. proc. */
	*PublicAPI
	int nested_fors;
	int nested_trys;

	/* Address resolution data.  Used to relink addresses after compile. */
	addrlist []*INTERMEDIATE	/* list of addresses to resolve */
	addroffsets []int			/* list of offsets from instrs */

	/* variable names.  The index into cstat->variables give you what position
	 * the variable holds.
	 */
	const char *variables[MAX_VAR];
	const char *localvars[MAX_VAR];
	const char *scopedvars[MAX_VAR];

	struct line *curr_line;		/* current line */
	int lineno;			/* current line number */
	int start_comment;              /* Line last comment started at */
	int force_comment;              /* Only attempt certain compile. */
	const char *next_char;		/* next char * */
	ObjectID player, program;		/* player and program for this compile */

	int compile_err;			/* 1 if error occured */

	char *line_copy;
	int macrosubs;				/* Safeguard for macro-subst. infinite loops */
	int descr;					/* the descriptor that initiated compiling */
	int force_err_display;		/* If true, always show compiler errors. */
	struct INTERMEDIATE *nextinst;
	defhash map[string] string
} COMPSTATE;


#define free_prog(i) free_prog_real(i,__FILE__,__LINE__);

/* Character defines */
#define BEGINCOMMENT '('
#define ENDCOMMENT ')'
#define BEGINSTRING '"'
#define ENDSTRING '"'
#define BEGINMACRO '.'
#define BEGINDIRECTIVE '$'
#define BEGINESCAPE '\\'

#define SUBSTITUTIONS 20		/* How many nested macros will we allow? */

void
do_abort_compile(COMPSTATE * cstat, const char *c)
{
	static char _buf[BUFFER_LEN];

	if (cstat->start_comment) {
	  _buf = fmt.Sprintf("Error in line %d: %s  Comment starting at line %d.", cstat->lineno, c, cstat->start_comment);
	  cstat->start_comment = 0;
	} else {
	  _buf = fmt.Sprintf("Error in line %d: %s", cstat->lineno, c);
	}
	if (cstat->line_copy) {
		free((void *) cstat->line_copy);
		cstat->line_copy = NULL;
	}
	if (DB.Fetch(cstat.player).IsFlagged(INTERACTIVE) && !DB.Fetch(cstat.player).IsFlagged(READMODE)) || cstat.force_err_display {
		notify_nolisten(cstat->player, _buf, true)
	} else {
		log_muf("%s(#%d) [%s(#%d)] %s(#%d) %s",
			DB.Fetch(DB.Fetch(cstat.program).Owner).name, DB.Fetch(cstat.program).Owner,
			DB.Fetch(cstat.program).name, cstat.program,
			DB.Fetch(cstat.player).name, cstat.player,
			_buf
		);
	}
	cstat->compile_err++;
	if (cstat->compile_err > 1) {
		return;
	}
	if (cstat->nextinst) {
		struct INTERMEDIATE* ptr;
		while (cstat->nextinst)
		{
			ptr = cstat->nextinst;
			cstat->nextinst = ptr->next;
			free(ptr);
		}
		cstat->nextinst = NULL;
	}
	cleanup(cstat)
	cstat.PublicAPI = nil
	free_prog(cstat.program)
	p := DB.Fetch(cstat.program).program
	p.PublicAPI = nil
	p.mcp_binding = nil
	p.proftime.tv_usec = 0
	p.proftime.tv_sec = 0
}

/* abort compile macro */
#define abort_compile(ST,C) { do_abort_compile(ST,C); return 0; }

/* abort compile for void functions */
#define v_abort_compile(ST,C) { do_abort_compile(ST,C); return; }

void compiler_warning(COMPSTATE* cstat, char* text, ...)
{
	char buf[BUFFER_LEN];
	va_list vl;

	va_start(vl, text);
	buf = fmt.Sprintf(text, vl...);
	va_end(vl);

	notify_nolisten(cstat->player, buf, true)
}

/*****************************************************************/


#define ADDRLIST_ALLOC_CHUNK_SIZE 256

func get_address(c *COMPSTATE, dest *INTERMEDIATE, offset int) int {
	for i, v := range c.addrlist {
		if v == dest && c.addroffsets[i] == offset {
			return i
		}
	}
	c.addrlist = append(c.addrlist, dest)
	c.addroffsets = append(c.addroffsets, offset)
	return len(c.addrlist) - 1
}

func fix_addresses(cstat *COMPSTATE) {
	/* renumber the instruction chain */
	for ptr := cstat.first_word; ptr != nil; ptr = ptr.next {
		count++
		ptr.no = count
	}

	/* repoint PublicAPI to targets */
	for pub := cstat.PublicAPI; pub != nil; pub = pub.next {
		i := pub.address.(int)
		pub.address = cstat.addrlist[i].address.(int) + cstat.addroffsets[i]
	}

	/* repoint addresses to targets */
	for ptr := cstat.first_word; ptr != nil; ptr = ptr.next {
		switch ptr.in.data.(type) {
		case Address, PROG_IF, PROG_TRY, PROG_JMP, PROG_EXEC:
			i := ptr.in.data.(int)
			ptr.in.data = cstat.addrlist[i].address.(int) + cstat.addroffsets[i]
		}
	}
}


/*****************************************************************/


func fixpubs(mypubs *PublicAPI, offset *inst) {
	for ; mypubs != nil; mypubs = mypubs.next {
		mypubs.address = offset + mypubs>address.(int)
	}
}


func size_pubs(mypubs *PublicAPI) (r int) {
	for ; mypubs != nil; mypubs = mypubs.next {
		r += sizeof(*mypubs)
	}
	return
}

func expand_def(cstat *COMPSTATE, defname string) (r string) {
	if r = cstat.defhash[defname]; r == "" && defname == BEGINMACRO {
		r = Macros.Expand(defname)
	}
	return
}

func kill_def(cstat *COMPSTATE, defname string) {
	delete(cstat.defhash, defname)
}

func insert_def(cstat *COMPSTATE, defname, deff string) {
	cstat.defhash[defname] = deff
}

func insert_intdef(COMPSTATE * cstat, const char *defname, int deff) {
	insert_def(cstat, defname, fmt.Sprint(deff))
}

func include_defs(cstat *COMPSTATE, i ObjectID) {
	char dirname[BUFFER_LEN];
	char temp[BUFFER_LEN];
	const char *tmpptr;

	dirname := "/_defs/"
	var pptr *Plist
	j := pptr.first_prop(i, dirname, temp)
	for j != nil {
		dirname = "/_defs/" + temp
		tmpptr = get_property_class(i, dirname)
		if tmpptr != nil && *tmpptr != nil {
			insert_def(cstat, temp, (char *) tmpptr)
		}
		j, temp = j.next_prop(pptr)
	}
}


void include_internal_defs(COMPSTATE * cstat) {
	/* Create standard server defines */
	insert_def(cstat, "__version", VERSION);
	insert_def(cstat, "__muckname", tp_muckname);
	insert_intdef(cstat, "__fuzzball__", 1);
	insert_def(cstat, "strip", "striplead striptail");
	insert_def(cstat, "instring", "tolower swap tolower swap instr");
	insert_def(cstat, "rinstring", "tolower swap tolower swap rinstr");
	insert_intdef(cstat, "bg_mode", BACKGROUND);
	insert_intdef(cstat, "fg_mode", FOREGROUND);
	insert_intdef(cstat, "pr_mode", PREEMPT);
	insert_intdef(cstat, "max_variable_count", MAX_VAR);
	insert_intdef(cstat, "sorttype_caseinsens", SORTTYPE_CASEINSENS);
	insert_intdef(cstat, "sorttype_descending", SORTTYPE_DESCENDING);
	insert_intdef(cstat, "sorttype_case_ascend", SORTTYPE_CASE_ASCEND);
	insert_intdef(cstat, "sorttype_nocase_ascend", SORTTYPE_NOCASE_ASCEND);
	insert_intdef(cstat, "sorttype_case_descend", SORTTYPE_CASE_DESCEND);
	insert_intdef(cstat, "sorttype_nocase_descend", SORTTYPE_NOCASE_DESCEND);
	insert_intdef(cstat, "sorttype_shuffle", SORTTYPE_SHUFFLE);

	/* Make defines for compatability to removed primitives */
	insert_def(cstat, "desc", "\"_/de\" getpropstr");
	insert_def(cstat, "succ", "\"_/sc\" getpropstr");
	insert_def(cstat, "fail", "\"_/fl\" getpropstr");
	insert_def(cstat, "drop", "\"_/dr\" getpropstr");
	insert_def(cstat, "osucc", "\"_/osc\" getpropstr");
	insert_def(cstat, "ofail", "\"_/ofl\" getpropstr");
	insert_def(cstat, "odrop", "\"_/odr\" getpropstr");
	insert_def(cstat, "setdesc", "\"_/de\" swap 0 addprop");
	insert_def(cstat, "setsucc", "\"_/sc\" swap 0 addprop");
	insert_def(cstat, "setfail", "\"_/fl\" swap 0 addprop");
	insert_def(cstat, "setdrop", "\"_/dr\" swap 0 addprop");
	insert_def(cstat, "setosucc", "\"_/osc\" swap 0 addprop");
	insert_def(cstat, "setofail", "\"_/ofl\" swap 0 addprop");
	insert_def(cstat, "setodrop", "\"_/odr\" swap 0 addprop");
	insert_def(cstat, "preempt", "pr_mode setmode");
	insert_def(cstat, "background", "bg_mode setmode");
	insert_def(cstat, "foreground", "fg_mode setmode");
	insert_def(cstat, "notify_except", "1 swap notify_exclude");
	insert_def(cstat, "event_wait", "0 array_make event_waitfor");
	insert_def(cstat, "tread", "\"__tread\" timer_start { \"TIMER.__tread\" \"READ\" }list event_waitfor swap pop \"READ\" strcmp if \"\" 0 else read 1 \"__tread\" timer_stop then");
	insert_def(cstat, "truename", "name");

	/* MUF Error defines */
	insert_def(cstat, "err_divzero?", "0 is_set?");
	insert_def(cstat, "err_nan?", "1 is_set?");
	insert_def(cstat, "err_imaginary?", "2 is_set?");
	insert_def(cstat, "err_fbounds?", "3 is_set?");
	insert_def(cstat, "err_ibounds?", "4 is_set?");

	/* Array convenience defines */
	insert_def(cstat, "}array", "} array_make");
	insert_def(cstat, "}list", "} array_make");
	insert_def(cstat, "}dict", "} 2 / array_make_dict");
	insert_def(cstat, "}join", "} array_make \"\" array_join");
	insert_def(cstat, "}cat", "} array_make array_interpret");
	insert_def(cstat, "}tell", "} array_make me @ 1 array_make array_notify");
	insert_def(cstat, "[]", "array_getitem");
	insert_def(cstat, "->[]", "array_setitem");
	insert_def(cstat, "[]<-", "array_appenditem");
	insert_def(cstat, "[..]", "array_getrange");
	insert_def(cstat, "array_diff", "2 array_ndiff");
	insert_def(cstat, "array_union", "2 array_nunion");
	insert_def(cstat, "array_intersect", "2 array_nintersect");

	/* GUI dialog types */
	insert_def(cstat, "d_simple", "\"simple\"");
	insert_def(cstat, "d_tabbed", "\"tabbed\"");
	insert_def(cstat, "d_helper", "\"helper\"");

	/* GUI control types */
	insert_def(cstat, "c_menu",      "\"menu\"");
	insert_def(cstat, "c_datum",     "\"datum\"");
	insert_def(cstat, "c_label",     "\"text\"");
	insert_def(cstat, "c_image",     "\"image\"");
	insert_def(cstat, "c_hrule",     "\"hrule\"");
	insert_def(cstat, "c_vrule",     "\"vrule\"");
	insert_def(cstat, "c_button",    "\"button\"");
	insert_def(cstat, "c_checkbox",  "\"checkbox\"");
	insert_def(cstat, "c_radiobtn",  "\"radio\"");
	insert_def(cstat, "c_password",  "\"password\"");
	insert_def(cstat, "c_edit",      "\"edit\"");
	insert_def(cstat, "c_multiedit", "\"multiedit\"");
	insert_def(cstat, "c_combobox",  "\"combobox\"");
	insert_def(cstat, "c_spinner",   "\"spinner\"");
	insert_def(cstat, "c_scale",     "\"scale\"");
	insert_def(cstat, "c_listbox",   "\"listbox\"");
	insert_def(cstat, "c_tree",      "\"tree\"");
	insert_def(cstat, "c_frame",     "\"frame\"");
	insert_def(cstat, "c_notebook",  "\"notebook\"");

	/* Backwards compatibility for old GUI dialog creation prims */
	insert_def(cstat, "gui_dlog_simple", "d_simple swap 0 array_make_dict gui_dlog_create");
	insert_def(cstat, "gui_dlog_tabbed", "d_tabbed swap \"panes\" over array_keys array_make \"names\" 4 rotate array_vals array_make 2 array_make_dict gui_dlog_create");
	insert_def(cstat, "gui_dlog_helper", "d_helper swap \"panes\" over array_keys array_make \"names\" 4 rotate array_vals array_make 2 array_make_dict gui_dlog_create");

	/* Regex */
	insert_intdef(cstat, "reg_icase",		MUF_RE_ICASE);
	insert_intdef(cstat, "reg_all",			MUF_RE_ALL);
	insert_intdef(cstat, "reg_extended",	MUF_RE_EXTENDED);
}


func init_defs(cstat *COMPSTATE) {
	cstat.defhash = make(map[string] string)

	/* Create standard server defines */
	include_internal_defs(cstat)

	/* Include any defines set in #0's _defs/ propdir. */
	include_defs(cstat, (ObjectID) 0)

	/* Include any defines set in program owner's _defs/ propdir. */
	include_defs(cstat, DB.Fetch(cstat.program).Owner)
}

func uncompile_program(i ObjectID) {
	dequeue_prog(i, 1)
	free_prog(i)
	DB.Fetch(i).(Program) = nil
}

func do_uncompile(ObjectID player) {
	if !Wizard(DB.Fetch(player).Owner) {
		notify_nolisten(player, "Permission denied. (uncompile)", true);
		return;
	}
	EachObject(func(obj ObjectID) {
		if IsProgram(obj) {
			uncompile_program(obj)
		}
	})
	notify_nolisten(player, "All programs decompiled.", true)
}

func free_unused_programs() {
	time_t now = time(nil)
	EachObject(func(obj ObjectID, o *Object) {
		var instances int
		if p := o.program; p.sp != nil {
			instances = p.instances
		}
		if IsProgram(obj) && !o.IsFlagged(ABODE, INTERNAL) && (now - o.LastUsed > tp_clean_interval) && instances == 0 {
			uncompile_program(obj)
		}
	})
}

/* Various flags for the IMMEDIATE instructions */

#define IMMFLAG_REFERENCED	1	/* Referenced by a jump */

/* Checks code for valid fetch-and-clear optim changes, and does them. */
func MaybeOptimizeVarsAt(cstat *COMPSTATE, first *INTERMEDIATE, AtNo int, BangNo int) {
	if !first.IsFlagged(INTMEDFLG_INTRY) {
		var farthest int
		var lvarflag bool
		switch first.in.(type) {
		case PROG_LVAR_AT, PROG_LVAR_AT_CLEAR:
			lvarflag = true
		}

		for curr := first.next; curr != nil; curr = curr.next {
			if curr.IsFlagged(INTMEDFLG_INTRY) {
				break
			}

			switch curr.in.(type) {
			case PROG_PRIMITIVE:
				/* Don't trust any physical @ or !'s in the code, someone
					may be indirectly referencing the scoped variable */
				/* Don't trust any explicit jmp's in the code. */
				switch curr.in.data {
				case AtNo, BangNo, IN_JMP:
					break
				}

				if lvarflag {
					/* For lvars, don't trust the following prims... */
					/*   EXITs escape the code path without leaving lvar scope. */
					/*   EXECUTEs escape the code path without leaving lvar scope. */
					/*   CALLs cause re-entrancy problems. */
					switch curr.in.data {
					case IN_RET, IN_EXECUTE, IN_CALL:
						break
					}
				}
			case PROG_LVAR_AT, PROG_LVAR_AT_CLEAR:
				if lvarflag {
					if curr.in.data == first.in.data {
						/* Can't optimize if references to the variable found before a var! */
						break
					}
				}
			case PROG_SVAR_AT, PROG_SVAR_AT_CLEAR:
				if !lvarflag {
					if curr.in.data.(int) == first.in.data.(int) {
						/* Can't optimize if references to the variable found before a var! */
						break
					}
				}
			case PROG_LVAR_BANG:
				if lvarflag {
					if first.in.data == curr.in.data {
						if addr := curr.address.(int); addr > farthest {
							/* Optimize it! */
							first.in.data = PROG_LVAR_AT_CLEAR(first.in.data)
						}
						break
					}
				}
			case PROG_SVAR_BANG:
				if !lvarflag {
					if first.in.data == curr.in.data {
						if addr := curr.address.(int); addr > farthest {
							/* Optimize it! */
							first.in.data = PROG_SVAR_AT_CLEAR(first.in.data)
						}
						break
					}
				}
			case PROG_EXEC:
				if lvarflag {
					/* Don't try to optimize lvars over execs */
					break
				}
			case PROG_IF, PROG_TRY, PROG_JMP:
				ptr := cstat.addrlist[curr.in.data]
				for i := cstat.addroffsets[curr.in.data]; ptr.next != nil && i > 0; i-- {
					ptr = ptr.next
				}
				if addr := ptr.address.(int); addr <= first.no {
					/* Can't optimize as we've exited the code branch the @ is in. */
					break
				}
				if addr := ptr.address.(int); addr > farthest {
					farthest = addr
				}
			case MUFProc:
				/* Don't try to optimize over functions */
				break
			}
		}
	}
}

func RemoveNextIntermediate(cstat *COMPSTATE, curr *INTERMEDIATE) {
	if curr.next != nil {
		for i, v := range cstat.addrlist {
			if v == curr.next {
				cstat.addrlist[i] = curr
			}
		}
		curr.next = curr.next.next
		cstat.nowords--
	}
}

func RemoveIntermediate(cstat *COMPSTATE, curr *INTERMEDIATE) {
	if curr.next != nil {
		curr.no = curr.next.no
		curr.in.line = curr.next.in.line
		curr.in.data = curr.next.in.data
		curr.next.in.data = 0
		RemoveNextIntermediate(cstat, curr)
	}
}

func ContiguousIntermediates(Flags *int, ptr *INTERMEDIATE, count int) (ok bool) {
	for ok = true ; ok && count > 0; count-- {
		switch {
		case ptr == nil:
			ok = false
		case Flags[ptr.no] & IMMFLAG_REFERENCED != 0:
			ok  = false
		default:
			ptr = ptr.next
		}
	}
	return
}

func IntermediateIsPrimitive(ptr *INTERMEDIATE, primnum int) (ok bool) {
	if ptr != nil {
		if _, ok = ptr.in.data.(PROG_PRIMITIVE) {
			ok = ptr.in.data == primnum
		}
	}
	return
}

func IntermediateIsInteger(ptr *INTERMEDIATE, val int) (ok bool) {
	if ptr != nil {
		ok = ptr.data == val
	}
	return
}

func IntermediateIsString(ptr *INTERMEDIATE, val string) (ok bool) {
	if ptr != nil {
		ok = ptr.in == val
	}
	return
}

var (
	AtNo PROG_PRIMITIVE = get_primitive("@") /* Wince */
	BangNo PROG_PRIMITIVE = get_primitive("!")
	SwapNo PROG_PRIMITIVE = get_primitive("swap")
	RotNo PROG_PRIMITIVE = get_primitive("rot")
	NotNo PROG_PRIMITIVE = get_primitive("not")
	StrcmpNo PROG_PRIMITIVE = get_primitive("strcmp")
	EqualsNo PROG_PRIMITIVE = get_primitive("=")
	PlusNo PROG_PRIMITIVE = get_primitive("+")
	MinusNo PROG_PRIMITIVE = get_primitive("-")
	MultNo PROG_PRIMITIVE = get_primitive("*")
	DivNo PROG_PRIMITIVE = get_primitive("/")
	ModNo PROG_PRIMITIVE = get_primitive("%")
	DecrNo PROG_PRIMITIVE = get_primitive("--")
	IncrNo PROG_PRIMITIVE = get_primitive("++")
)

func OptimizeIntermediate(cstat *COMPSTATE, force_err_display bool) (r int) {
	/* Code assumes everything is setup nicely, if not, bad things will happen */
	if cstat.first_word != nil {
		old_instr_count := cstat.nowords
		/* renumber the instruction chain */
		var count int
		for curr := cstat.first_word; curr != nil; curr = curr.next {
			curr.no = count
			count++
		}

		Flags := make([]int, count)

		/* Mark instructions which jumps reference */

		for curr := cstat.first_word; curr != nil; curr = curr.next {
			switch a := curr.in.data.(type) {
			case Address, PROG_IF, PROG_TRY, PROG_JMP, PROG_EXEC:
				i := cstat.addrlist[a].address.(int) + cstat.addroffsets[a]
				Flags[i] |= IMMFLAG_REFERENCED
			}
		}

		for curr := cstat.first_word; curr != nil; {
			advance := true
			switch x := curr.in.(type) {
			case PROG_LVAR:
				/* lvar !  ==>  lvar! */
				/* lvar @  ==>  lvar@ */
				if curr.next != nil {
					if _, ok := curr.next.in.data.(PROG_PRIMITIVE); ok {
						if curr.next.in.data == AtNo {
							if ContiguousIntermediates(Flags, curr.next, 1) {
								curr.in.type = PROG_LVAR_AT
								RemoveNextIntermediate(cstat, curr)
								advance = false
								break
							}
						}
						if curr.next.in.data == BangNo {
							if ContiguousIntermediates(Flags, curr.next, 1) {
								curr.in.type = PROG_LVAR_BANG
								RemoveNextIntermediate(cstat, curr)
								advance = false
								break
							}
						}
					}
				}
			case PROG_SVAR:
				/* svar !  ==>  svar! */
				/* svar @  ==>  svar@ */
				if curr.next != nil {
					if _, ok := curr.next.in.data.(PROG_PRIMITIVE); ok {
						if curr.next.in.data == AtNo {
							if ContiguousIntermediates(Flags, curr.next, 1) {
								curr.in.type = PROG_SVAR_AT
								RemoveNextIntermediate(cstat, curr)
								advance = false
								break
							}
						}
						if curr.next.in.data == BangNo {
							if ContiguousIntermediates(Flags, curr.next, 1) {
								curr.in.type = PROG_SVAR_BANG
								RemoveNextIntermediate(cstat, curr)
								advance = false
								break
							}
						}
					}
				}
			case string:
				/* "" strcmp 0 =  ==>  not */
				if IntermediateIsString(curr, "") {
					if ContiguousIntermediates(Flags, curr.next, 3) {
						if IntermediateIsPrimitive(curr.next, StrcmpNo) {
							if IntermediateIsInteger(curr.next.next, 0) {
								if IntermediateIsPrimitive(curr.next.next.next, EqualsNo) {
									curr.in.data = NotNo
									RemoveNextIntermediate(cstat, curr)
									RemoveNextIntermediate(cstat, curr)
									RemoveNextIntermediate(cstat, curr)
									advance = false
									break
								}
							}
						}
					}
				}
			case int:
				/* consolidate constant integer calculations */
				if ContiguousIntermediates(Flags, curr.next, 2) {
					if y, ok := curr.next.in.data.(int); ok {
						/* Int Int +  ==>  Sum */
						if IntermediateIsPrimitive(curr.next.next, PlusNo) {
							x += y
							RemoveNextIntermediate(cstat, curr)
							RemoveNextIntermediate(cstat, curr)
							advance = false
							curr.in = x
							break
						}

						/* Int Int -  ==>  Diff */
						if IntermediateIsPrimitive(curr.next.next, MinusNo) {
							x -= y
							RemoveNextIntermediate(cstat, curr)
							RemoveNextIntermediate(cstat, curr)
							advance = false
							curr.in = x
							break
						}

						/* Int Int *  ==>  Prod */
						if IntermediateIsPrimitive(curr.next.next, MultNo) {
							x *= y
							RemoveNextIntermediate(cstat, curr)
							RemoveNextIntermediate(cstat, curr)
							advance = false
							curr.in = x
							break
						}

						/* Int Int /  ==>  Div  */
						if IntermediateIsPrimitive(curr.next.next, DivNo) {
							if y == 0 {
								if !curr.next.next.IsFlagged(INTMEDFLG_DIVBYZERO) {
									curr.next.next.FlagAs(INTMEDFLG_DIVBYZERO)
									if force_err_display {
										compiler_warning(cstat, "Warning on line %i: Divide by zero", curr.next.next.in.line)
									}
								}
							} else {
								x /= y
								RemoveNextIntermediate(cstat, curr)
								RemoveNextIntermediate(cstat, curr)
								advance = false
							}
							curr.in = x
							break
						}

						/* Int Int %  ==>  Div  */
						if IntermediateIsPrimitive(curr.next.next, ModNo) {
							if in == 0 {
								if !curr.next.next.IsFlagged(INTMEDFLG_MODBYZERO) {
									curr.next.next.FlagAs(INTMEDFLG_MODBYZERO)
									if force_err_display {
										compiler_warning(cstat, "Warning on line %i: Modulus by zero", curr.next.next.in.line)
									}
								}
							} else {
								x %= y
								RemoveNextIntermediate(cstat, curr)
								RemoveNextIntermediate(cstat, curr)
								advance = false
							}
							curr.in.data = x
							break
						}
					}
				}

				/* 0 =  ==>  not */
				if IntermediateIsInteger(curr, 0) {
					if ContiguousIntermediates(Flags, curr.next, 1) {
						if IntermediateIsPrimitive(curr.next, EqualsNo) {
							curr.in.data = NotNo
							RemoveNextIntermediate(cstat, curr)
							advance = false
							break
						}
					}
				}

				/* 1 +  ==>  ++ */
				if IntermediateIsInteger(curr, 1) {
					if ContiguousIntermediates(Flags, curr.next, 1) {
						if IntermediateIsPrimitive(curr.next, PlusNo) {
							curr.in.data = IncrNo
							RemoveNextIntermediate(cstat, curr)
							advance = false
							break
						}
					}
				}

				/* 1 -  ==>  -- */
				if IntermediateIsInteger(curr, 1) {
					if ContiguousIntermediates(Flags, curr.next, 1) {
						if IntermediateIsPrimitive(curr.next, MinusNo) {
							curr.in.data = DecrNo
							RemoveNextIntermediate(cstat, curr)
							advance = false
							break
						}
					}
				}

			case PROG_PRIMITIVE:
					/* rot rot swap  ==>  swap rot */
					if IntermediateIsPrimitive(curr, RotNo) {
						if ContiguousIntermediates(Flags, curr.next, 2) {
							if IntermediateIsPrimitive(curr.next, RotNo) {
								if IntermediateIsPrimitive(curr.next.next, SwapNo) {
									curr.in.data = SwapNo
									curr.next.in.data = RotNo
									RemoveNextIntermediate(cstat, curr.next)
									advance = false
									break
								}
							}
						}
					}
					/* not not if  ==>  if */
					if IntermediateIsPrimitive(curr, NotNo) {
						if ContiguousIntermediates(Flags, curr.next, 2) {
							if IntermediateIsPrimitive(curr.next, NotNo) {
								if _, ok := curr.next.next.in.data.(PROG_IF); ok {
									RemoveIntermediate(cstat, curr)
									RemoveIntermediate(cstat, curr)
									advance = false
									break
								}
							}
						}
					}
					break
			}

			if advance {
				curr = curr.next
			}
		}

		/* Turn all var@'s which have a following var! into a var@-clear */
		for curr := cstat.first_word; curr != nil; curr = curr.next {
			switch curr.in.data.(type) {
			case PROG_SVAR_AT, PROG_LVAR_AT:
				MaybeOptimizeVarsAt(cstat, curr, AtNo, BangNo)
			}
		}
		r = old_instr_count - cstat.nowords		
	}
	return
}

/* Genericized Optimizer ideas:
 *
 * const int OI_ANY = -121314;   // arbitrary unlikely-to-be-needed value.
 *
 * typedef enum {
 *     OI_KEEP,
 *     OI_CHGVAL,
 *     OI_CHGTYPE,
 *     OI_REPLACE,
 *     OI_DELETE
 * } OI_ACTION;
 *
 * OPTIM* option_new();
 * void optim_free(OPTIM* optim);
 * void optim_add_raw  (OPTIM* optim, struct INTERMEDIATE* originst,
 *                      OI_ACTION action, struct INTERMEDIATE* newinst);
 * void optim_add_type (OPTIM* optim, int origtype,
 *                      OI_ACTION action, int newtype);
 * void optim_add_prim (OPTIM* optim, const char* origprim,
 *                      OI_ACTION action, int newval);
 * void optim_add_int  (OPTIM* optim, int origval,
 *                      OI_ACTION action, int newval);
 * void optim_add_str  (OPTIM* optim, const char* origval,
 *                      OI_ACTION action, int newval);
 *
 *
 * OPTIM* optim = optim_new(cstat);
 * optim_add_str (optim, "",       OI_DELETE, 0);
 * optim_add_prim(optim, "strcmp", OI_CHGVAL, get_primitive("not"));
 * optim_add_int (optim, 0,        OI_DELETE, 0);
 * optim_add_prim(optim, "=",      OI_DELETE, 0);
 *
 * OPTIM* optim = optim_new(cstat);
 * optim_add_str(optim, "",        OI_DELETE, 0);
 * optim_add_prim(optim, "strcmp", OI_DELETE, 0);
 * optim_add_prim(optim, "not",    OI_KEEP,   0);
 *
 * OPTIM* optim = optim_new(cstat);
 * optim_add_prim(optim, "rot",  OI_CHGVAL, get_primitive("swap"));
 * optim_add_prim(optim, "rot",  OI_KEEP,   0);
 * optim_add_prim(optim, "swap", OI_DELETE, 0);
 *
 * OPTIM* optim = optim_new(cstat);
 * optim_add_type(optim, PROG_SVAR, OI_CHGTYPE, PROG_SVAR_AT);
 * optim_add_prim(optim, "@",       OI_DELETE,   0);
 *
 * OPTIM* optim = optim_new(cstat);
 * optim_add_type(optim, PROG_SVAR, OI_CHGTYPE, PROG_SVAR_BANG);
 * optim_add_prim(optim, "!",       OI_DELETE,   0);
 *
 * OPTIM* optim = optim_new(cstat);
 * optim_add_int (optim, 0,   OI_DELETE, 0);
 * optim_add_prim(optim, "=", OI_CHGVAL, get_primitive("not"));
 *
 * OPTIM* optim = optim_new(cstat);
 * optim_add_prim(optim, "not",   OI_DELETE, 0);
 * optim_add_prim(optim, "not",   OI_DELETE, 0);
 * optim_add_type(optim, PROG_IF, OI_KEEP,   0);
 *
 * OPTIM* optim = optim_new(cstat);
 * optim_add_int (optim, 0,       OI_DELETE,  0);
 * optim_add_type(optim, PROG_IF, OI_CHGTYPE, PROG_JMP);
 *
 * OPTIM* optim = optim_new(cstat);
 * optim_add_int (optim, 1,       OI_DELETE, 0);
 * optim_add_type(optim, PROG_IF, OI_DELETE, 0);
 *
 * OPTIM* optim = optim_new(cstat);
 * optim_add_int (optim, 1,   OI_DELETE, 0);
 * optim_add_prim(optim, "+", OI_CHGVAL, get_primitive("++"));
 *
 * OPTIM* optim = optim_new(cstat);
 * optim_add_int (optim, 1,   OI_DELETE, 0);
 * optim_add_prim(optim, "-", OI_CHGVAL, get_primitive("--"));
 *
 */


//	overall control code.  Does piece-meal tokenization parsing and backward checking.
func do_compile(descr int, player, program ObjectID, force_err_display bool) {
	p := DB.Fetch(program).program
	cstat := &COMPSTATE{
		force_err_display: force_err_display,
		descr: descr,
		curr_line: p.first,
		lineno: 1,
		force_comment: tp_muf_comments_strict,
		player: player,
		program: program,
	}
	cstat.variables[0] = "ME"
	cstat.variables[1] = "LOC"
	cstat.variables[2] = "TRIGGER"
	cstat.variables[3] = "COMMAND"
	for i := RES_VAR; i < MAX_VAR; i++ {
		cstat.variables[i] = ""
	}
	for i := 0; i < MAX_VAR; i++ {
		cstat.localvars[i] = ""
		cstat.scopedvars[i] = ""
	}
	if cstat.curr_line != nil {
		cstat.next_char = cstat.curr_line.this_line
	}
	init_defs(&cstat)

	/* free old stuff */
	dequeue_prog(cstat.program, 1)
	free_prog(cstat.program)
	p.PublicAPI = nil
	p.mcp_binding = nil
	p.mcp_binding = nil
	p.proftime.tv_usec = 0
	p.proftime.tv_sec = 0
	p.profstart = time(NULL)
	p.profuses = 0

	if cstat.curr_line == nil {
		v_abort_compile(&cstat, "Missing program text.")
	}

	/* do compilation */
	for token := next_token(cstat); token != nil; token := next_token(cstat) {
		if cstat.compile_err {
			return
		}
		new_word := next_word(cstat, token)

		if cstat.compile_err {
			return
		}

		if new_word != nil {
			if cstat.first_word == nil {
				cstat.curr_word = new_word
				cstat.first_word = cstat.curr_word
			else {
				cstat.curr_word.next = new_word
				cstat.curr_word = cstat.curr_word.next
			}
		}
		for cstat.curr_word != nil && cstat.curr_word.next != nil {
			cstat.curr_word = cstat.curr_word.next
		}
	}

	if cstat.curr_proc != nil {
		v_abort_compile(&cstat, "Unexpected end of file.")
	}

	if cstat.procs == nil {
		v_abort_compile(&cstat, "Missing procedure definition.")
	}

	if tp_optimize_muf {
		maxpasses := 4
		passcount := 0
		optimcount := 0
		optcnt := 1

		for ; optcnt > 0 && maxpasses > 0; maxpasses-- {
			optcnt = OptimizeIntermediate(&cstat, force_err_display)
			optimcount += optcnt
			passcount++
		}

		if force_err_display && optimcount > 0 {
			notify_nolisten(cstat.player, fmt.Sprintf("Program optimized by %d instructions in %d passes.", optimcount, passcount), true)
		}
	}

	/* do copying over */
	fix_addresses(&cstat)
	copy_program(&cstat)
	fixpubs(cstat.PublicAPI, p.code)
	p.PublicAPI = cstat.PublicAPI
	cstat.nextinst = nil
	if !cstat.compile_err {
		set_start(&cstat)
		cleanup(&cstat)
		p.instances = 0
		/* restart AUTOSTART program. */
		if DB.Fetch(cstat.program).IsFlagged(ABODE) && TrueWizard(DB.Fetch(cstat.program).Owner) {
			add_muf_queue_event(-1, DB.Fetch(cstat.program).Owner, NOTHING, NOTHING, cstat.program, "Startup", "Queued Event.", 0)
		}

		if force_err_display {
			notify_nolisten(cstat.player, "Program compiled successfully.", true)
		}
	}
}

func next_word(cstat *COMPSTATE, token string) (r *INTERMEDIATE) {
	switch {
	case token == "":
	case call(cstat, token):
		r = call_word(cstat, token)
	case scopedvar(cstat, token):
		r = svar_word(cstat, token)
	case localvar(cstat, token):
		r = lvar_word(cstat, token)
	case variable(cstat, token):
		r = var_word(cstat, token)
	case special(token):
		r = process_special(cstat, token)
	case primitive(token):
		r = primitive_word(cstat, token)
	case string(token):
		r = string_word(cstat, token + 1)
	case unicode.IsNumber(token):
		r = number_word(cstat, token)
	case ifloat(token):
		r = float_word(cstat, token)
	case object(token):
		r = object_word(cstat, token)
	case quoted(cstat, token):
		r = quoted_word(cstat, token + 1)
	default:
		abort_compile(cstat, fmt.Sprintf("Unrecognized word %s.", token))
	}
	return
}

/* Little routine to do the line_copy handling right */
func advance_line(cstat *COMPSTATE) {
	cstat.curr_line = cstat.curr_line.next
	cstat.lineno++
	cstat.macrosubs = 0
	if cstat.curr_line != "" {
		cstat.line_copy = cstat.curr_line.this_line
		cstat.next_char = cstat.line_copy
	} else {
		cstat.line_copy = ""
		cstat.next_char = ""
	}
}

/* Skips comments, grabs strings, returns NULL when no more tokens to grab. */
func next_token_raw(cstat *COMPSTATE) (buf string) {
	if cstat.curr_line != "" && cstat.next_char != "" {
		switch cstat.next_char = strings.TrimSpace(cstat.next_char); {
		case cstat.next_char == "":
			advance_line(cstat)
			buf = next_token_raw(cstat)
		case cstat.next_char == BEGINCOMMENT:
			cstat.start_comment = cstat.lineno
			if cstat.force_comment == 1 {
				do_comment(cstat, -1)
			} else {
				do_comment(cstat, 0)
			}
			cstat.start_comment = 0
			buf = next_token_raw(cstat)
		case cstat.next_char == BEGINSTRING:
			buf = do_string(cstat)
		default:
			if i := strings.IndexFunc(cstat.next_char, unicode.IsSpace); i > 0 {
				buf = cstat.next_char[:i]
				cstat.next_char = cstat.next_char[i:]
			}
		}
	}
	return
}

func next_token(cstat *COMPSTATE) (r string) {
	switch temp := next_token_raw(cstat); {
	case temp == "":
	case temp[0] == BEGINDIRECTIVE:
		do_directive(cstat, temp)
		r = next_token(cstat)
	case temp[0] == BEGINESCAPE:
		if len(temp) > 1 {
			temp = temp[1:]
		}
		r = temp
	default:
		if expansion := expand_def(cstat, temp); r != "" {
			cstat.macrosubs++
			if cstat.macrosubs > SUBSTITUTIONS {
				abort_compile(cstat, "Too many macro substitutions.")
			} else {
				cstat.line_copy = expansion + cstat.next_char
				cstat.next_char = cstat.line_copy
				r = next_token(cstat)
			}
		} else {
			r = temp
		}
	}
	return
}

/* skip comments, recursive style */
func do_new_comment(cstat *COMPSTATE, depth int) (r int) {
	switch {
	case cstat.next_char == "" || cstat.next_char != BEGINCOMMENT:
		r = 2
	case depth >= 7: /*arbitrary*/
		r = 3
	default:
		for cstat.next_char = cstat.next_char[1:]; cstat.next_char != ENDCOMMENT; {
			switch cstat.next_char {
			case "":
				for cstat.next_char == "" {
					advance_line(cstat)
					if cstat.curr_line == nil {
						return 1
					}
				}
			case BEGINCOMMENT:
				if r = do_new_comment(cstat, depth + 1); r != 0 {
					return
				}
			default:
				cstat.next_char = cstat.next_char[1:]
			}
		}

		cstat.next_char = cstat.next_char[1:]  /* Advance past ENDCOMMENT */
		var in_str bool
		for ptr := cstat.next_char; ptr != ""; ptr = ptr[1:] {
			if in_str {
				if ptr == ENDSTRING {
					in_str = false
				}
			} else {
				switch ptr {
				case BEGINSTRING:
					in_str = true
				case ENDSTRING:
					in_str = true
					break
				}
			}
		}
		if in_str {
			compiler_warning(cstat, "Warning on line %i: Unterminated string may indicate unterminated comment. Comment starts on line %i.", cstat.lineno, cstat.start_comment)
		}
		if cstat.next_char == "" {
			advance_line(cstat)
		}
		if depth > 0 && cstat.curr_line == nil {	/* EOF? Don't care if done (depth==0) */
			r = 1
		}
	}
	return
}

/* skip comments */
func do_comment(cstat *COMPSTATE, depth int) {
	var next_char int		/* Save state if needed. */
	var lineno int
	var curr_line *line
	var macrosubs bool

	if depth == 0 {
		if cstat.force_comment == 0 {
			if cstat.line_copy != "" {
				next_char = cstat.next_char - cstat.line_copy
			}
			macrosubs = cstat.macrosubs
			lineno = cstat.lineno
			curr_line = cstat.curr_line
		}

		if r := do_new_comment(cstat, 0); r != 0 {
			if cstat.force_comment != 0 {
				switch r {
				case 1:
					v_abort_compile(cstat, "Unterminated comment.");
				case 2:
					v_abort_compile(cstat, "Expected comment.");
				case 3:
					v_abort_compile(cstat, "Comments nested too deep (more than 7 levels).");
				}
				return
			} else {
				/* Set back up, drop through for retry. */
				cstat.line_copy = ""
				cstat.curr_line = curr_line
				cstat.macrosubs = macrosubs
				cstat.lineno = lineno
				if cstat.curr_line != nil {
					cstat.line_copy = cstat.curr_line.this_line
					cstat.next_char = cstat.line_copy[next_char:]
				} else {
					cstat.next_char = ""
				}
			}
		} else {
			/* Comment hunt worked, new-style. */
			return
		}
	}
}

func is_preprocessor_conditional(token string) (r bool) {
	switch token {
	case "$ifdef", "$ifndef", "$iflib", "$ifnlib", "$ifver", "$iflibver", "$ifnver", "$ifnlibver", "$ifcancall", "$ifncancall":
		r = true
	}
	return
}

func do_directive_match(cstat *COMPSTATE, name string) (i int) {
	tempa := match_args
	tempb := match_cmdname
	i = NewMatch(cstat.descr, cstat.player, name, NOTYPE).
		MatchRegistered().
		MatchAbsolute().
		MatchMe().
		MatchResult()
	match_args = tempa
	match_cmdname = tempb
	return
}

/* handle compiler directives */
func do_directive(cstat *COMPSTATE, direct string) {
	switch temp := direct[1:]; temp {
	case "":
		v_abort_compile(cstat, "I don't understand that compiler directive!")
	case "define":
		name := next_token_raw(cstat)
		if tmpname == "" {
			v_abort_compile(cstat, "Unexpected end of file looking for $define name.")
		}
		var definition, term string
		for term = next_token_raw(cstat); term != "" && term != "$enddef"; term = next_token_raw(cstat) {
			for cp := term; cp != ""; cp = cp[1:] {
				if term[0] == BEGINSTRING && cp != term && (cp[0] == ENDSTRING || cp[0] == BEGINESCAPE) {
					definition += BEGINESCAPE
				}
				definition += cp[0]
			}
			if term[0] == BEGINSTRING {
				definition += ENDSTRING
			}
			definition += ' '
		}
		if term == "" {
			v_abort_compile(cstat, "Unexpected end of file in $define definition.")
		}
		insert_def(cstat, name, definition)
	case "cleardefs":
		cstat.defhash = make(map[string] string)
		include_internal_defs(cstat)
		cstat.next_char = strings.TrimSpace(cstat.next_char)
		nextToken := cstat.next_char
		name := nextToken
		cstat.next_char = ""
		advance_line(cstat)
		if name == "" || MLevel(DB.Fetch(cstat.program).Owner) < WIZBIT {
			include_defs(cstat, DB.Fetch(cstat.program).Owner)
			include_defs(cstat, ObjectID(0))
		}
	case "enddef":
		v_abort_compile(cstat, "$enddef without a previous matching $define.")
	case "def":
		name := next_token_raw(cstat)
		if name == "" {
			v_abort_compile(cstat, "Unexpected end of file looking for $def name.")
		}
		insert_def(cstat, name, cstat.next_char)
		cstat.next_char = ""
		advance_line(cstat)
	case "pubdef":
		switch name := next_token_raw(cstat); {
		case name == "":
			v_abort_compile(cstat, "Unexpected end of file looking for $pubdef name.")
		case name != ":" && (strings.ContainsAny(name, "/:") || Prop_SeeOnly(name) || Prop_Hidden(name) || Prop_System(name)):
			v_abort_compile(cstat, "Invalid $pubdef name.  No /, :, @ nor ~ are allowed.")
		case name == ":":
			remove_property(cstat.program, "/_defs")
		default:
			var propname string
			cstat.next_char = strings.TrimSpace(cstat.next_char)
			defstr := cstat.next_char
			doitset := true
			if name[0] == '\\' {
				name = name[1:]
				propname = fmt.Sprintf("/_defs/%s", name)
				doitset = get_property_class(cstat.program, propname) != ""
			} else {
				propname = fmt.Sprintf("/_defs/%s", name)
			}
			if doitset {
				if defstr != "" {
					add_property(cstat.program, propname, defstr, 0)
				} else {
					remove_property(cstat.program, propname)
				}
			}
		}
		cstat.next_char = ""
		advance_line(cstat)
	case "libdef":
		switch name := next_token_raw(cstat); {
		case name == "":
			v_abort_compile(cstat, "Unexpected end of file looking for $libdef name.")
		case strings.ContainsAny(name, "/:") || Prop_SeeOnly(name) || Prop_Hidden(name) || Prop_System(name):
			v_abort_compile(cstat, "Invalid $libdef name.  No /, :, @, nor ~ are allowed.")
		default:
			var propname string
			cstat.next_char = strings.TrimSpace(cstat.next_char)
			doitset := true
			if name == '\\' {
				name = name[1:]
				propname = fmt.Sprintf("/_defs/%s", name)
				doitset = get_property_class(cstat.program, propname) != ""
			} else {
				propname = fmt.Sprintf("/_defs/%s", name);
			}
			definition := fmt.Sprintf("#%i \"%s\" call", cstat.program, name)
			if doitset {
				if defstr != "" {
					add_property(cstat.program, propname, definition, 0)
				} else {
					remove_property(cstat.program, propname)
				}
			}
		}
		cstat.next_char = ""
		advance_line(cstat)
	case "include":
		var i ObjectID
		if name := next_token_raw(cstat); name == "":
			v_abort_compile(cstat, "Unexpected end of file while doing $include.")
		} else {
			i = do_directive_match(cstat, name)
			if !i.IsValid() {
				v_abort_compile(cstat, "I don't understand what object you want to $include.")
			}
			include_defs(cstat, i)
		}
	case "undef":
		name := next_token_raw(cstat)
		if name == "" {
			v_abort_compile(cstat, "Unexpected end of file looking for name to $undef.")
		}
		kill_def(cstat, name)
	case "echo":
		notify_nolisten(cstat.player, cstat.next_char, true)
		cstat.next_char = ""
		advance_line(cstat)
	case "abort":
		cstat.next_char = spaces.TrimSpace(cstat.next_char)
		if name := cstat.next_char; name == "" {
			v_abort_compile(cstat, name)
		} else {
			v_abort_compile(cstat, "Forced abort for the compile.")
		}
	case "version":
		name := next_token_raw(cstat)
		if !ifloat(name) {
			v_abort_compile(cstat, "Expected a floating point number for the version.")
		}
		cstat.next_char = ""
		add_property(cstat.program, "_version", name, 0)
		advance_line(cstat)
	case "lib-version":
		name := next_token_raw(cstat)
		if !ifloat(name) {
			v_abort_compile(cstat, "Expected a floating point number for the version.")
		}
		cstat.next_char = ""
		add_property(cstat.program, "_lib-version", name, 0)
		advance_line(cstat)
	case "author":
		cstat.next_char = strings.TrimSpace(cstat.next_char)
		name := cstat.next_char
		cstat.next_char = ""
		add_property(cstat.program, "_author", name, 0)
		advance_line(cstat)
	case "note":
		cstat.next_char = strings.TrimSpace(cstat.next_char)
		name := cstat.next_char
		cstat.next_char = ""
		add_property(cstat.program, "_note", name, 0)
		advance_line(cstat)
	case "ifdef", "ifndef":
		invert_flag := temp == "ifndef"
		switch name := next_token_raw(cstat); name == "" {
			v_abort_compile(cstat, "Unexpected end of file looking for $ifdef condition.");
		}
		if name == '"' {
			temp = name[1:]
		} else {
			temp = name
		}
		for i = 1; i < len(temp) && temp[i] != '=' && temp[i] != '>' && temp[i] != '<'; i++ {}
		name = temp[i:]
		switch temp[i] {
		case '>':
			i = 1
		case '=':
			i = 0
		case '<':
			i = -1
		default:
			i = -2
		}
		*tmpname = '\0';
		tmpname++;
		definition := expand_def(cstat, temp)
		if i == -2 {
			if definition == "" {
				j = 1
			}
		} else {
			if definition == "" {
				j = 1
			} else {
				j = strings.Compare(definition, name)
				switch {
				case i == 0 && j == 0, i * j > 0:
					j = 0
				default:
					j = 1
				}
			}
		}
		if invert_flag {
			if j = 0 {
				j = 1
			} else {
				j = 0
			}
		}
		if j != 0 {
			i = 0
			for definition = next_token_raw(cstat); definition != "" && (i != 0 || (definition != "$else" && definition != "$endif")); definition = next_token_raw(cstat) {
				switch {
				case is_preprocessor_conditional(definition):
					i++
				case definition == "$endif":
					i--
				}
			}
			if definition == "" {
				v_abort_compile(cstat, "Unexpected end of file in $ifdef clause.")
			}
		}
	case "ifcancall", "ifncancall":
		var i ObjectID
		switch name := next_token_raw(cstat); name {
		case "":
			v_abort_compile(cstat, "Unexpected end of file for ifcancall.")
		case "this":
			i = cstat.program
		default:
			i = do_directive_match(cstat, name)
		}
		if !i.IsValid() {
			v_abort_compile(cstat, "I don't understand what program you want to check in ifcancall.")
		}
		name := next_token_raw(cstat)
		if name == "" {
			v_abort_compile(cstat, "I don't understand what function you want to check for.")
		}
		cstat.next_char = ""
		advance_line(cstat)
		if program := DB.Fetch(i).program; program.code == nil {
			tmpline := program.first
			program.first = read_program(i)
			do_compile(cstat.descr, DB.Fetch(i).Owner, i, 0)
			program.first = tmpline
		}
		j := 0
		if MLevel(DB.Fetch(i).Owner) > NON_MUCKER && (MLevel(DB.Fetch(cstat.program).Owner) >= WIZBIT || DB.Fetch(i).Owner == DB.Fetch(cstat.program).Owner || Linkable(i)) {
			pbs := DB.Fetch(i).(Program).PublicAPI
			for ; pbs != nil && name != pbs.subname; pbs = pbs.next {}
			if pbs != nil && MLevel(DB.Fetch(cstat.program).Owner) >= pbs.mlev {
				j = 1
			}
		}
		if temp == "ifncancall" {
			if j == 0 {
				j = 1
			} else {
				j = 0
			}
		}
		if j == 0 {
			i = 0
			for tmpptr = next_token_raw(cstat); tmpptr != "" && (i != 0 || (tmpptr != "$else" && tmpptr != "$endif")); tmpptr = next_token_raw(cstat) {
				switch {
				case is_preprocessor_conditional(tmpptr):
					i++
				case tmpptr == "$endif":
					i--
				}
			}
			if tmpptr == "" {
				v_abort_compile(cstat, "Unexpected end of file in $ifcancall clause.")
			}
		}
	case "ifver", "iflibver", "ifnver", "ifnlibver":
		double verflt = 0;
		double checkflt = 0;
		int needFree = 0;

		if name := next_token_raw(cstat); name == "" {
			v_abort_compile(cstat, "Unexpected end of file while doing $ifver.")
		} else {
			var i ObjectID
			if name == "this" {
				i = cstat.program			
			} else {
				i = do_directive_match(cstat, name)
			}
			if !i.IsValid() {
				v_abort_compile(cstat, "I don't understand what object you want to check with $ifver.")
			}
			var property string
			switch temp {
			case "ifver", "ifnver":
				property = get_property_class(i, "_version")
			default:
				property = get_property_class(i, "_lib-version")
			}
			if property == "" {
				property = "0.0"
			}
			if name = next_token_raw(cstat); name == "" {
				v_abort_compile(cstat, "I don't understand what version you want to compare to with $ifver.")
			}
			if property == "" || !ifloat(property) {
				verflt = 0.0
			} else {
				sscanf(property, "%lg", &verflt)
			}
			if name == "" || !ifloat(name) {
				checkflt = 0.0
			} else {
				sscanf(name, "%lg", &checkflt)
			}
			cstat.next_char = ""
			advance_line(cstat)
			if checkflt <= verflt {
				j = 1
			} else {
				j = 0
			}
			switch temp {
			case "ifnver", "ifnlibver":
				if j = 0 {
					j = 1
				} else {
					j = 0
				}
			}
			if j == 0 {
				i = 0
				for tmpptr = next_token_raw(cstat); tmpptr != "" && (i != 0 || (tmpptr != "$else" && tmpptr != "$endif")); tmpptr = next_token_raw(cstat) {
					switch {
					case is_preprocessor_conditional(tmpptr):
						i++
					case tmpptr == "$endif":
						i--
					}
				}
				if tmpptr == "" {
					v_abort_compile(cstat, "Unexpected end of file in $ifver clause.")
				}
			}
		}
	case "iflib", "ifnlib":
		name := next_token_raw(cstat)
		if name == "" {
			v_abort_compile(cstat, "Unexpected end of file in $iflib/$ifnlib clause.")
		}
		i = do_directive_match(cstat, name)
		if !i.IsValid() || !IsProgram(i) {
			j = 1
		} else {
			j = 0
		}
		if temp == "ifnlib" {
			if j = 0 {
				j = 1
			} else {
				j = 0
			}
		}
		if j == 0 {
			i = 0
			for tmpptr = next_token_raw(cstat); tmpptr != "" && (i != 0 || (tmpptr != "$else" && tmpptr != "$endif")); tmpptr = next_token_raw(cstat) {
				switch {
				case is_preprocessor_conditional(tmpptr):
					i++
				case tmpptr == "$endif":
					i--
				}
			}
			if tmpptr == "" {
				v_abort_compile(cstat, "Unexpected end of file in $iflib clause.")
			}
		}
	case "else":
		i = 0
		for tmpptr = next_token_raw(cstat); tmpptr != "" && (i != 0 || (tmpptr != "$else" && tmpptr != "$endif")); tmpptr = next_token_raw(cstat) {
			switch {
			case is_preprocessor_conditional(tmpptr):
				i++
			case tmpptr == "$endif":
				i--
			}
		}
		if tmpptr == "" {
			v_abort_compile(cstat, "Unexpected end of file in $else clause.")
		}
	case "endif":
	case "pragma":
		/* FIXME - move pragmas to its own section for easy expansion. */
		if cstat.next_char = strings.TrimSpace(cstat.next_char); cstat.next_char == "" {
			v_abort_compile(cstat, "Pragma requires at least one argument.")
		}
		switch tmpptr = next_token_raw(cstat); param {
		case "":
			v_abort_compile(cstat, "Pragma requires at least one argument.")
		case "comment_strict":
			/* Do non-recursive comments (old style) */
			cstat.force_comment = 1
		case "comment_recurse":
			/* Do recursive comments ((new) style) */
			cstat.force_comment = 2
		case "comment_loose" {
			/* Try to compile with recursive and non-recursive comments
			doing recursive first, then strict on a comment-based
			compile error.  Only throw an error if both fail.  This is
			the default mode. */
			cstat.force_comment = 0
		default:
			/* If the pragma is not recognized, it is ignored, with a warning. */
			compiler_warning(cstat, "Warning on line %i: Pragma %.64s unrecognized.  Ignoring.", cstat.lineno, tmpptr)
			cstat.next_char = ""
		}
		if cstat.next_char != "" {
			compiler_warning(cstat, "Warning on line %i: Ignoring extra pragma arguments: %.256s", cstat.lineno, cstat.next_char)
			advance_line(cstat)
		}
	default:
		v_abort_compile(cstat, "Unrecognized compiler directive.")
	}
}


/* return string */
func do_string(cstat *COMPSTATE) (buf string) {
	buf += cstat.next_char[0]
	cstat.next_char = cstat.next_char[1:]
	i := 1
	for quoted := false; (quoted || cstat.next_char[0] != ENDSTRING) && cstat.next_char[0]; cstat.next_char = cstat.next_char[1:] {
		switch c := cstat.next_char[0]; {
		case c == '\\' && !quoted:
			quoted = true
		case c == 'r' && quoted:
			buf += '\r'
			quoted = false
		case c == '[' && quoted:
			buf += ESCAPE_CHAR
			quoted = false
		default:
			buf += c
			quoted = false
		}
	}
	if cstat.next_char == "" {
		abort_compile(cstat, "Unterminated string found at end of line.")
	}
	cstat.next_char = cstat.next_char[1:]
	return
}

func add_word(cstat *COMPSTATE, v interface{}) (r *INTERMEDIATE) {
	r = new_inst(cstat)
	r.no = cstat.nowords
	cstat.nowords++
	r.in = &inst{ line: cstat.lineno, data: v }
	return
}

/* process special.  Performs special processing.
   It sets up FOR and IF structures.  Remember --- for those, we've got to set aside an extra argument space.         */
func process_special(cstat *COMPSTATE, token string) (r *INTERMEDIATE) {
	switch token {
	case ":":
		var argsflag bool

		switch proc_name := next_token(cstat);
		case cstat.curr_proc != nil:
			abort_compile(cstat, "Definition within definition.")
		case proc_name == "":
			abort_compile(cstat, "Unexpected end of file within procedure.")
		default:
			if proc_name[len(proc_name) - 1] == '[' {
				argsflag = true
				proc_name = proc_name[:len(proc_name) - 1]
			}
			if proc_name == "" {
				abort_compile(cstat, "Bad procedure name.")
			} else {
				proc := MUFProc{ name: proc_name }
				r = add_word(cstat, &proc)
				cstat.curr_proc = r
				if argsflag {
					var outflag bool
					var varspec string
					for argsdone := false; !argsdone; {
						switch varspec = next_token(cstat); {
						case varspec == "":
							abort_compile(cstat, "Unexpected end of file within procedure arguments declaration.")
						case varspec == "]":
							argsdone = true
						case varspec == "--":
							outflag = true
						}
						if !outflag {
							varname := strchr(varspec, ':')
							if varname {
								varname++
							} else {
								varname = varspec
							}
							if varname != "" {
								if add_scopedvar(cstat, varname) < 0 {
									abort_compile(cstat, "Variable limit exceeded.")
								}
								proc.vars++
								proc.args++
							}
						}
					}
				}

				cstat.procs = &PROC_LIST{
					name: proc_name,
					code: r,
					next: cstat.procs,
				}
			}
		}
	case ";":
		switch {
		case cstat.control_stack != nil:
			abort_compile(cstat, "Unexpected end of procedure definition.")
		case cstat.curr_proc == nil:
			abort_compile(cstat, "Procedure end without body.")
		default:
			r = add_word(cstat, PROG_PRIMITIVE(IN_RET))
			proc := cstat.curr_proc.in.data.(MUFProc)
			if varcnt := proc.vars; varcnt != 0 {
			    proc.varnames = make([]string, varcnt)
			    for i, _ := range proc.varnames {
					proc.varnames[i] = cstat.scopedvars[i]
					cstat.scopedvars[i] = ""
			    }
			}
			cstat.curr_proc = nil
		}
	case "IF":
		r = add_word(cstat, PROG_IF(add_control_structure(cstat, CTYPE_IF, nu)))
	case "ELSE":
		switch innermost_control_type(cstat) {
		case CTYPE_IF:
		case CTYPE_TRY:
			abort_compile(cstat, "Unterminated TRY-CATCH block at ELSE.")
		case CTYPE_CATCH:
			abort_compile(cstat, "Unterminated CATCH-ENDCATCH block at ELSE.")
		case CTYPE_FOR, CTYPE_BEGIN:
			abort_compile(cstat, "Unterminated Loop at ELSE.")
		default:
			abort_compile(cstat, "ELSE without IF.")
		}
		r = add_word(cstat, PROG_JMP(0))
		pop_control_structure(cstat, CTYPE_IF, 0).in.data = get_address(cstat, r, 1)
		add_control_structure(cstat, CTYPE_ELSE, r)
	case "THEN":
		switch innermost_control_type(cstat) {
		case CTYPE_IF, CTYPE_ELSE:
		case CTYPE_TRY:
			abort_compile(cstat, "Unterminated TRY-CATCH block at THEN.")
		case CTYPE_CATCH:
			abort_compile(cstat, "Unterminated CATCH-ENDCATCH block at THEN.")
		case CTYPE_FOR, CTYPE_BEGIN:
			abort_compile(cstat, "Unterminated Loop at THEN.")
		default:
			abort_compile(cstat, "THEN without IF.")
		}
		prealloc_inst(cstat)
		pop_control_structure(cstat, CTYPE_IF, CTYPE_ELSE).in.data = get_address(cstat, cstat.nextinst, 0)
	case "BEGIN":
		prealloc_inst(cstat)
		add_control_structure(cstat, CTYPE_BEGIN, cstat.nextinst)
	case "FOR":
		r = add_word(cstat, PROG_PRIMITIVE(IN_FOR))
		r.next = add_word(cstat, PROG_PRIMITIVE(IN_FORITER))
		r.next.next = add_word(cstat, PROG_IF(0))
		add_control_structure(cstat, CTYPE_FOR, r.next)
		cstat.nested_fors++
	case "FOREACH":
		r = add_word(cstat, PROG_PRIMITIVE(IN_FOREACH))
		r.next = add_word(cstat, PROG_PRIMITIVE(IN_FORITER))
		r.next.next = add_word(cstat, PROG_IF(0))
		add_control_structure(cstat, CTYPE_FOR, r.next)
		cstat.nested_fors++
	case "UNTIL":
		prealloc_inst(cstat)
		switch innermost_control_type(cstat) {
		case CTYPE_FOR:
			cstat.nested_fors--
			resolve_loop_addrs(cstat, get_address(cstat, cstat.nextinst, 1))
			r = add_word(cstat, PROG_IF(get_address(cstat, pop_control_structure(cstat, CTYPE_BEGIN, CTYPE_FOR), 0))
			r.next = add_word(cstat, PROG_PRIMITIVE(IN_FORPOP))
		case CTYPE_BEGIN:
			resolve_loop_addrs(cstat, get_address(cstat, cstat.nextinst, 1))
			r = add_word(cstat, PROG_IF(get_address(cstat, pop_control_structure(cstat, CTYPE_BEGIN, CTYPE_FOR), 0))
		case CTYPE_TRY:
			abort_compile(cstat, "Unterminated TRY-CATCH block at UNTIL.")
		case CTYPE_CATCH:
			abort_compile(cstat, "Unterminated CATCH-ENDCATCH block at UNTIL.")
		case CTYPE_IF, CTYPE_ELSE:
			abort_compile(cstat, "Unterminated IF-THEN at UNTIL.")
		default:
			abort_compile(cstat, "Loop start not found for UNTIL.")
		}
	case "WHILE":
		if !in_loop(cstat) {
			abort_compile(cstat, "Can't have a WHILE outside of a loop.")
		} else {
			var first *INTERMEDIATE
			for trycount := count_trys_inside_loop(cstat); trycount > 0; trycount-- {
				if r == nil {
					r = add_words(cstat, PROG_PRIMITIVE(IN_TRYPOP))
					first = r
				} else {
					r.next = add_words(cstat, PROG_PRIMITIVE(IN_TRYPOP))
					r = r.next
				}
			}
			if r == nil {
				r = add_words(cstat, PROG_PRIMITIVE(IN_TRYPOP))
				first = r
			} else {
				r.next = add_words(cstat, PROG_PRIMITIVE(IN_TRYPOP))
				r = r.next
			}
			add_loop_exit(cstat, r)
			r = first
		}
	case "BREAK":
		if !in_loop(cstat) {
			abort_compile(cstat, "Can't have a BREAK outside of a loop.")
		} else {
			var first *INTERMEDIATE
			for trycount := count_trys_inside_loop(cstat); trycount > 0; trycount-- {
				if r == nil {
					r = add_words(cstat, PROG_PRIMITIVE(IN_TRYPOP))
					first = r
				} else {
					r.next = add_words(cstat, PROG_PRIMITIVE(IN_TRYPOP))
					r = r.next
				}
			}
			if r == nil {
				r = add_words(cstat, PROG_JMP(0))
				first = r
			} else {
				r.next = add_words(cstat, PROG_JMP(0))
				r = r.next
			}
			add_loop_exit(cstat, r)
			r = first
		}
	case "CONTINUE":
		if !in_loop(cstat) {
			abort_compile(cstat, "Can't CONTINUE outside of a loop.")
		} else {
			beef := locate_control_structure(cstat, CTYPE_FOR, CTYPE_BEGIN)
			var first *INTERMEDIATE
			for trycount := count_trys_inside_loop(cstat); trycount > 0; trycount-- {
				if r == nil {
					r = add_words(cstat, PROG_PRIMITIVE(IN_TRYPOP))
					first = r
				} else {
					r.next = add_words(cstat, PROG_PRIMITIVE(IN_TRYPOP))
					r = r.next
				}
			}
			if r == nil {
				r = add_words(cstat, PROG_JMP(get_address(cstat, beef, 0)))
				first = r
			} else {
				r.next = add_words(cstat, PROG_JMP(get_address(cstat, beef, 0)))
				r = r.next
			}
			r = first
		}
	case "REPEAT":
		switch innermost_control_type(cstat) {
		case CTYPE_FOR:
			cstat.nested_fors--
			prealloc_inst(cstat)
			resolve_loop_addrs(cstat, get_address(cstat, cstat.nextinst, 1))
			r = add_word(cstat, PROG_JMP(get_address(cstat, pop_control_structure(cstat, CTYPE_BEGIN, CTYPE_FOR), 0)))
			r.next = add_word(cstat, PROG_PRIMITIVE(IN_FORPOP))
		case CTYPE_BEGIN:
			prealloc_inst(cstat)
			resolve_loop_addrs(cstat, get_address(cstat, cstat.nextinst, 1))
			r = add_word(cstat, PROG_JMP(get_address(cstat, pop_control_structure(cstat, CTYPE_BEGIN, CTYPE_FOR), 0)))
		case CTYPE_TRY:
			abort_compile(cstat, "Unterminated TRY-CATCH block at REPEAT.")
		case CTYPE_CATCH:
			abort_compile(cstat, "Unterminated CATCH-ENDCATCH block at REPEAT.")
		case CTYPE_IF, CTYPE_ELSE:
			abort_compile(cstat, "Unterminated IF-THEN at REPEAT.")
		default:
			abort_compile(cstat, "Loop start not found for REPEAT.")
		}
	case "TRY":
		r = add_word(cstat, PROG_TRY(0))
		add_control_structure(cstat, CTYPE_TRY, r)
		cstat.nested_trys++
	case "CATCH", "CATCH_DETAILED":
		switch innermost_control_type(cstat) {
		case CTYPE_TRY, CTYPE_CATCH:
			r = add_word(cstat, PROG_PRIMITIVE(IN_TRYPOP))
			r.next = add_word(cstat, PROG_JMP(0))
			if token == "CATCH_DETAILED" {
				r.next.next = add_word(cstat, PROG_PRIMITIVE(IN_CATCH_DETAILED))
			} else {
				r.next.next = add_word(cstat, PROG_PRIMITIVE(IN_CATCH))
			}
			pop_control_structure(cstat, CTYPE_TRY, 0).in.data = get_address(cstat, r.next.next, 0)
			cstat.nested_trys--
			add_control_structure(cstat, CTYPE_CATCH, r.next)
		case CTYPE_FOR, CTYPE_BEGIN:
			abort_compile(cstat, "Unterminated Loop at CATCH.")
		case CTYPE_IF, CTYPE_ELSE:
			abort_compile(cstat, "Unterminated IF-THEN at CATCH.")
		default:
			abort_compile(cstat, "No TRY found for CATCH.")
		}
	case "ENDCATCH":
		switch innermost_control_type(cstat) {
		case CTYPE_CATCH:
			prealloc_inst(cstat)
			pop_control_structure(cstat, CTYPE_CATCH, 0).in.data = get_address(cstat, cstat.nextinst, 0)
		case CTYPE_FOR, CTYPE_BEGIN:
			abort_compile(cstat, "Unterminated Loop at ENDCATCH.")
		case CTYPE_IF, CTYPE_ELSE:
			abort_compile(cstat, "Unterminated IF-THEN at ENDCATCH.")
		default:
			abort_compile(cstat, "No CATCH found for ENDCATCH.")
		}
	case "CALL":
		r = add_word(cstat, PROG_PRIMITIVE(IN_CALL))
	case "WIZCALL", "PUBLIC":
		var wizflag bool
		if token == "WIZCALL" {
			wizflag = true
		}
		if cstat.curr_proc {
			abort_compile(cstat, "PUBLIC  or WIZCALL declaration within procedure.")
		}
		tok := next_token(cstat);
		if tok == nil || !call(cstat, tok) {
			abort_compile(cstat, "Subroutine unknown in PUBLIC or WIZCALL declaration.")
		}

		p := cstat.procs
		for ; p != nil && p.name != tok; p = p.next {}
		if p == nil {
			abort_compile(cstat, "Subroutine unknown in PUBLIC or WIZCALL declaration.")
		}
		if cstat.PublicAPI == nil {
			cstat.PublicAPI = &PublicAPI{ subname: tok, address: get_address(cstat, p.code, 0) }
			if wizflag {
				cstat.PublicAPI.mlev = WIZBIT
			} else {
				cstat.PublicAPI.mlev = APPRENTICE
			}
		} else {
			for pub := cstat.PublicAPI; pub != nil; pub = pub.next {
				switch {
				case tok == pub.subname:
					abort_compile(cstat, "Function already declared public.")
				case pub.next == nil {
					pub.next = &PublicAPI{ subname: tok }
					pub = pub.next
					pub.address = get_address(cstat, p.code, 0)
					if wizflag {
						pub.mlev = WIZBIT
					} else {
						pub.mlev = APPRENTICE
					}
					pub = nil
				}
			}
		}
	case "VAR":
		if tok := next_token(cstat); cstat.curr_proc == nil {
			switch {
			case tok == nil:
				abort_compile(cstat, "Unexpected end of program.")
			case add_variable(cstat, tok) == 0:
				abort_compile(cstat, "Variable limit exceeded.")
			}
		} else {
			switch {
			case tok == nil:
				abort_compile(cstat, "Unexpected end of program.")
			case add_scopedvar(cstat, tok) < 0:
				abort_compile(cstat, "Variable limit exceeded.")
			}
			cstat.curr_proc.in.data.(MUFProc).vars++
		}
	case "VAR!":
		if cstat.curr_proc == nil {
			abort_compile(cstat, "VAR! used outside of procedure.");
		} else {
			switch tok := next_token(cstat); {
			case tok == nil:
				abort_compile(cstat, "Unexpected end of program.")
			case add_scopedvar(cstat, tok) < 0:
				abort_compile(cstat, "Variable limit exceeded.")
			}
			proc := cstat.curr_proc.in.data.(MUFProc)
			r = add_word(cstat, PROG_SVAR_BANG(proc.vars))
			proc.vars++
		}
	case "LVAR":
		if cstat.curr_proc {
			abort_compile(cstat, "Local variable declared within procedure.")
		} else {
			switch tok := next_token(cstat); {
			case tok == nil, add_localvar(cstat, tok) == -1:
				abort_compile(cstat, "Local variable limit exceeded.")
			}
		}
	} else {
		abort_compile(cstat, fmt.Sprintf("Unrecognized special form %s found. (%d)", token, cstat.lineno))
	}
}

/* return primitive word. */
func primitive_word(cstat *COMPSTATE, token string) (r *INTERMEDIATE) {
	switch pnum := get_primitive(token); pnum {
	case IN_RET, IN_JMP:
		head := add_word(cstat, -1)
		curr := head
		for i := 0; i < cstat.nested_trys; i++ {
			curr.next = add_word(cstat, IN_TRYPOP)
			curr = cur.next
		}
		for i := 0; i < cstat.nested_fors; i++ {
			curr.next = add_word(cstat, IN_FORPOP)
			curr = cur.next
		}
		if head.next == nil {
			r = add_word(cstat, pnum)
		} else {
			curr.next = add_word(cstat, pnum)
			r = head.next
		}
	default:
		r = add_word(cstat, pnum)
	}
	return
}

/* return self pushing word (string) */
func string_word(cstat *COMPSTATE, token string) *INTERMEDIATE {
	return add_word(cstat, token)
}

/* return self pushing word (float) */
func float_word(cstat *COMPSTATE, token string) (r *INTERMEDIATE) {
	r = new_inst(cstat)
	r.no = cstat.nowords
	cstat.nowords++
	r.in = &inst{ line: cstat.lineno, data: sscanf(token, "%lg", &(r.in.data)) }
	return
}

/* return self pushing word (number) */
func number_word(cstat *COMPSTATE, token string) *INTERMEDIATE {
	return add_word(cstat, strconv.Atoi(token))
}

/* do a subroutine call --- push address onto stack, then make a primitive CALL. */
func call_word(cstat *COMPSTATE, token string) *INTERMEDIATE {
	var p *PROC_LIST
	for p = cstat.procs; p != nil && p.name != token; p = p.next {}
	return add_word(cstat, PROG_EXEC(get_address(cstat, p.code, 0))
}

func quoted_word(cstat *COMPSTATE, token string) *INTERMEDIATE {
	var p *PROC_LIST
	for p = cstat.procs; p != nil && p.name != token; p = p.next {}
	return add_word(cstat, get_address(cstat, p.code, 0))
}

/* returns number corresponding to variable number.
   We assume that it DOES exist */
func var_word(cstat *COMPSTATE, token string) *INTERMEDIATE {
	var i int
	for ; i < MAX_VAR && cstat.variables[i] != token; i++ {}
	return add_word(cstat, PROG_VAR(i))
}

func svar_word(cstat *COMPSTATE, token string) *INTERMEDIATE {
	var i int
	for ; i < MAX_VAR && cstat.scopedvars[i] != token; i++ {}
	return add_word(cstat, PROG_SVAR(i))
}

func lvar_word(cstat *COMPSTATE, token string) *INTERMEDIATE {
	var i int
	for ; i < MAX_VAR && cstat.localvars[i] != token; i++ {}
	return add_word(cstat, PROG_LVAR(i))
}

/* check if object is in database before putting it in */
func object_word(cstat *COMPSTATE, token string) *INTERMEDIATE {
	return add_word(cstat, ObjectID(strconv.Atoi(token[1:])))
}

/* support routines for internal data structures. */

/* add if to control stack */
func add_control_structure(cstat *COMPSTATE, typ int, place *INTERMEDIATE) {
	cstat.control_stack = &CONTROL_STACK{
		place: place,
		type: typ,
		next: cstat.control_stack,
	}
}

/* add while to current loop's list of exits remaining to be resolved. */
func add_loop_exit(cstat *COMPSTATE, place *INTERMEDIATE) {
	loop := cstat.control_stack
	for ; loop != nil && loop.type != CTYPE_BEGIN && loop.type != CTYPE_FOR; loop = loop.next {}
	if loop != nil {
		loop.extra = &CONTROL_STACK{
			place: place,
			type: CTYPE_WHILE,
			extra: loop.extra,
		}
	}
}

/* Returns true if a loop start is in the control structure stack. */
func in_loop(cstat *COMPSTATE) bool {
	loop := cstat.control_stack
	for ; loop != nil && loop.type != CTYPE_BEGIN && loop.type != CTYPE_FOR; loop = loop.next {}
	return loop != nil
}

/* Returns the type of the innermost nested control structure. */
func innermost_control_type(cstat *COMPSTATE) (r int) {
	if ctrl := cstat.control_stack; ctrl != nil {
		r = ctrl.type
	}
	return
}

/* Returns number of TRYs before topmost Loop */
func count_trys_inside_loop(cstat *COMPSTATE) (r int) {
	for loop := cstat.control_stack; loop != nil; loop = loop.next {
		if loop.type == CTYPE_FOR || loop.type == CTYPE_BEGIN {
			break
		}
		if loop.type == CTYPE_TRY {
			r++
		}
	}
	return
}

/* returns topmost begin or for off the stack */
func locate_control_structure(COMPSTATE* cstat, int type1, int type2) (r *INTERMEDIATE) {
	for loop := cstat.control_stack; loop != nil; loop = loop.next {
		if loop.type == type1 || loop.type == type2 {
			r = loop.place
			break
		}
	}
	return
}

/* checks if topmost loop stack item is a for */
func innermost_control_place(ctat *COMPSTATE, type1 int) (r *INTERMEDIATE) {
	switch ctrl := cstat.control_stack; {
	case ctrl == nil, ctrl.type != type1:
	default:
		r = ctrl.place
	}
	return
}

/* Pops off the innermost control structure and returns the place. */
func pop_control_structure(cstat *COMPSTATE, type1, type2 int) (r *INTERMEDIATE) {
	switch ctrl := cstat.control_stack; {
	case ctrl == nil, ctrl.type != type1 && ctrl.type != type2:
	default:
		r = ctrl.place
		cstat.control_stack = ctrl.next
	}
	return
}

/* pops first while off the innermost control structure, if it's a loop. */
func pop_loop_exit(cstat *COMPSTATE) (r *INTERMEDIATE) {
	switch parent := cstat.control_stack; {
	case parent == nil, parent.type != CTYPE_BEGIN && parent.type != CTYPE_FOR, parent.extra == nil, parent.extra.type != CTYPE_WHILE:
	default:
		r = parent.extra.place
		parent.extra = parent.extra.extra
	}
}

func resolve_loop_addrs(cstat *COMPSTATE, where int) {
	var exit *INTERMEDIATE
	for exit = pop_loop_exit(cstat); exit != nil; exit = pop_loop_exit(cstat) {
		exit.in.data = where
	}
	if exit = innermost_control_place(cstat, CTYPE_FOR); exit != nil {
		exit.next.in.data = where
	}
}

/* adds variable.  Return 0 if no space left */
func add_variable(cstat *COMPSTATE, varname string) (i int) {
	for i = RES_VAR; i < MAX_VAR && cstat.variables[i] != nil; i++ {}
	if i == MAX_VAR {
		i = 0
	} else {
		cstat.variables[i] = varname
	}
	return
}

/* adds local variable.  Return 0 if no space left */
func add_scopedvar(COMPSTATE * cstat, const char *varname) (i int) {
	for i = 0; i < MAX_VAR && cstat.scopedvars[i] != nil; i++ {}
	if i == MAX_VAR {
		i = -1
	} else {
		cstat.scopedvars[i] = varname
	}
	return
}

func add_localvar(cstat *COMPSTATE, varname string) (i int) {
	for i = 0; i < MAX_VAR && cstat.localvars[i] != nil; i++ {}
	if i == MAX_VAR {
		i = -1
	} else {
		cstat.localvars[i] = varname
	}
	return
}

/* predicates for procedure calls */
func special(token string) (r bool) {
	switch token {
	case ":", ";", "IF", "ELSE", "THEN", "BEGIN", "FOR", "FOREACH", "UNTIL", "WHILE", "BREAK", "CONTINUE", "REPEAT", "TRY", "CATCH", "CATCH_DETAILED", "ENDCATCH", "CALL", "PUBLIC", "WIZCALL", "LVAR", "VAR!", "VAR":
		r = true
	}
	return
}

/* see if procedure call */
func call(cstat *COMPSTATE, token string) (r bool) {
	struct PROC_LIST *i;

	for i := cstat.procs; i !- nil; i = i.next {
		if i.name == token {
			r = true
			break
		}
	}
	return
}

/* see if it's a quoted procedure name */
int
quoted(COMPSTATE * cstat, const char *token)
{
	return (*token == '\'' && call(cstat, token + 1));
}

/* see if it's an object # */
int
object(const char *token)
{
	if (*token == NUMBER_TOKEN && unicode.IsNumber(token + 1))
		return 1;
	else
		return 0;
}

/* see if string */
int
string(const char *token)
{
	return (token[0] == '"');
}

func variable(COMPSTATE * cstat, const char *token) (r bool) {
	for i := 0; i < MAX_VAR && cstat.variables[i] != nil; i++ {
		if token == cstat.variables[i] {
			r = true
			break
		}
	}
	return
}

func scopedvar(cstat *COMPSTATE, token string) (r bool) {
	for i := 0; i < MAX_VAR && cstat.scopedvars[i] != nil; i++ {
		if token == cstat.scopedvars[i] {
			r = true
			break
		}
	}
	return
}

func localvar(cstat *COMPSTATE, token string) (r bool) {
	for i := 0; i < MAX_VAR && cstat.localvars[i] != nil; i++ {
		if token == cstat->localvars[i] {
			r = true
			break
		}
	}
	return
}

/* see if token is primitive */
func primitive(token string) bool {
	primnum := get_primitive(token)
	return primnum != 0 && primnum <= (BASE_MAX - PRIMS_INTERNAL_CNT)
}

/* return primitive instruction */
func get_primitive(token string) PROG_PRIMITIVE {
	return primitive_list[token]
}

func append_intermediate_chain(chain, add *INTERMEDIATE) {
	for chain.next != nil {
		chain = chain.next
	}
	chain.next = add
}

func cleanup(c *COMPSTATE) {
	c.first_word = nil
	c.control_stack = nil
	c.procs = nil
	c.defhash = make(map[string] string)
	c.addroffsets = nil
	c.addrlist = nil
	for i := RES_VAR; i < MAX_VAR && c.variables[i] != nil; i++ {
		c.variables[i] = nil
	}
	for i := 0; i < MAX_VAR && c.scopedvars[i] != nil; i++ {
		c.scopedvars[i] = nil
	}
	for i := 0; i < MAX_VAR && c.localvars[i] != nil; i++ {
		c.localvars[i] = nil
	}
}



/* copy program to an array */
func copy_program(cstat *COMPSTATE) {
	if !cstat.first_word {
		v_abort_compile(cstat, "Nothing to compile.")
	}
	code := make([]*inst, cstat.nowords + 1)
	curr := cstat.first_word
	for i := 0; curr != nil; i++ {
		code[i].line = curr.in.line
		switch code[i].(type) {
		case PROG_PRIMITIVE, int, PROG_SVAR, PROG_SVAR_AT, PROG_SVAR_AT_CLEAR, PROG_SVAR_BANG, PROG_LVAR, PROG_LVAR_AT, PROG_LVAR_AT_CLEAR, PROG_LVAR_BANG, PROG_VAR, float64, string, ObjectID:
			code[i].data = curr.in.data
		case MUFProc:
			data := curr.in.data.(MUFProc)
			proc := &MUFProc {
				name: data.name
				vars: data.vars
				args: data.args
			}
			copy(proc.varnames, data.varnames)
			code[i].data = proc
		case Address:
			code[i].data = alloc_addr(cstat, curr.in.data.(int), code)
		case PROG_IF, PROG_JMP, PROG_EXEC, PROG_TRY:
			code[i].data = code + curr.in.data.(int)
		default:
			v_abort_compile(cstat, "Unknown type compile!  Internal error.")
			break
		}
		curr = curr.next
	}
	DB.Fetch(cstat.program).(Program) = code
}

func set_start(cstat *COMPSTATE) {
	DB.Fetch(cstat.program).(Program).siz = cstat.nowords

	/* address instr no is resolved before this gets called. */
	DB.Fetch(cstat.program).(Program).start = DB.Fetch(cstat.program).(Program).code + cstat.procs.code.no
}

func prealloc_inst(cstat *COMPSTATE) (r *INTERMEDIATE) {
	/* only allocate at most one extra instr */
	if cstat.nextinst == nil {
		r = &INTERMEDIATE{ no: cstat.nowords }
		if cstat.nested_trys > 0 {
			r.FlagAs(INTMEDFLG_INTRY)
		}
		if cstat.nextinst == nil {
			cstat.nextinst = r
		} else {
			for ptr := cstat.nextinst; ptr.next != nil; ptr = ptr.next {}
			ptr.next = r
		}
	}
	return
}

func new_inst(COMPSTATE * cstat) (r *INTERMEDIATE) {
	if r = cstat.nextinst; r == nil {
		r = new(INTERMEDIATE)
	}
	cstat.nextinst = r.next
	r.next = nil
	if cstat.nested_trys > 0 {
		r.FlagAs(INTMEDFLG_INTRY)
	}
	return
}

/* allocate an address */
func alloc_addr(cstat *COMPSTATE, offset int, codestart *inst) *Address {
	return &Address{ cstat.program, data: codestart + offset }
}

func free_prog_real(prog ObjectID, file, line string) {
	p := DB.Fetch(prog).program
	if p.code != nil {
		var instances int
		if p.sp != nil {
			instances = p.instances
		}
		if instances != nil {
			log_status("WARNING: freeing program %s with %d instances reported from %s:%d", unparse_object(GOD, prog), instances, file, line)
		}
		if i := scan_instances(prog); i != 0 {
			log_status("WARNING: freeing program %s with %d instances found from %s:%d", unparse_object(GOD, prog), i, file, line)
		}
		for i, v := range p.code {
			if _, ok := v.(Address); ok {
				p.code[i].data = nil
			} else {
				p.code[i] = nil
			}
		}
	}
	p.code = nil
	p.start = 0
}

func init_primitives() {
	primitive_list = make(map[string] PROG_PRIMITIVE)
	for i := BASE_MIN; i <= BASE_MAX; i++ {
		primitive_list[base_inst[val - BASE_MIN]] = val
	}
	IN_FORPOP = get_primitive(" FORPOP")
	IN_FORITER = get_primitive(" FORITER")
	IN_FOR = get_primitive(" FOR")
	IN_FOREACH = get_primitive(" FOREACH")
	IN_TRYPOP = get_primitive(" TRYPOP")
	log_status("MUF: %d primitives exist.", BASE_MAX)
}