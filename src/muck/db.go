package fbmuck

func check_remote(player, what ObjectID) {
	o := DB.Fetch(what)
	p := DB.Fetch(player)
	if mlev < JOURNEYMAN && o.Location != player && o.Location != p.Location && what != p.Location && what != player && !controls(ProgUID, what) {
		panic("Mucker Level 2 required to get remote info.")
	}
}

func EachObject(f interface{}) {
	switch f := f.(type) {
	case func(ObjectID):
		for i := ObjectID(0); i < db_top; i++ {
			f(i)
		}
	case func(ObjectID, object):
		for i := ObjectID(0); i < db_top; i++ {
			f(i, DB.Fetch(i))
		}
	case func(object):
		for i := ObjectID(0); i < db_top; i++ {
			f(DB.Fetch(i))
		}
	case func(ObjectID) bool:
		var done bool
		for i := ObjectID(0); !done && i < db_top; i++ {
			done = f(i)
		}
	case func(ObjectID, object) bool:
		var done bool
		for i := ObjectID(0); !done && i < db_top; i++ {
			done = f(i, DB.Fetch(i))
		}
	case func(object) bool:
		var done bool
		for i := ObjectID(0); !done && i < db_top; i++ {
			done = f(DB.Fetch(i))
		}
	default:
		panic(f)
	}
}

func EachObjectInReverse(f interface{}) {
	switch f := f.(type) {
	case func(ObjectID):
		for i := db_top - 1; i > -1; i-- {
			f(i)
		}
	case func(ObjectID, object):
		for i := db_top - 1; i > -1; i-- {
			f(i, DB.Fetch(i))
		}
	case func(object):
		for i := db_top - 1; i > -1; i-- {
			f(DB.Fetch(i))
		}
	case func(ObjectID) bool:
		var done bool
		for i := db_top - 1; !done && i > -1; i-- {
			done = f(i)
		}
	case func(ObjectID, object) bool:
		var done bool
		for i := db_top - 1; !done && i > -1; i++ {
			done = f(i, DB.Fetch(i))
		}
	case func(object) bool:
		var done bool
		for i := db_top - 1; !done && i > -1; i++ {
			done = f(DB.Fetch(i))
		}
	default:
		panic(f)
	}
}


	/* max length of command argument to process_command */
	#define MAX_COMMAND_LEN 2048
	#define BUFFER_LEN ((MAX_COMMAND_LEN)*4)
	#define FILE_BUFSIZ ((BUFSIZ)*8)

	typedef int ObjectID;				/* offset into db */

	#define TIME_INFINITE ((sizeof(time_t) == 4)? 0xefffffff : 0xefffffffffffffff)

	#define DB_READLOCK(x)
	#define DB_WRITELOCK(x)
	#define DB_RELEASE(x)

//	defines for possible data access mods.
const (
	MESGPROP_DESC = "_/de"
	MESGPROP_IDESC = "_/ide"
	MESGPROP_SUCC = "_/sc"
	MESGPROP_OSUCC = "_/osc"
	MESGPROP_FAIL = "_/fl"
	MESGPROP_OFAIL = "_/ofl"
	MESGPROP_DROP = "_/dr"
	MESGPROP_ODROP = "_/odr"
	MESGPROP_DOING = "_/do"
	MESGPROP_OECHO = "_/oecho"
	MESGPROP_PECHO = "_/pecho"
	MESGPROP_LOCK = "_/lok"
	MESGPROP_FLOCK = "@/flk"
	MESGPROP_CONLOCK = "_/clk"
	MESGPROP_CHLOCK = "_/chlk"
	MESGPROP_VALUE = "@/value"
	MESGPROP_GUEST = "@/isguest"
)


	#define DB_PARMSINFO     0x0001
	#define DB_COMPRESSED    0x0002

	#define TYPE_ROOM           0x0
	#define TYPE_THING          0x1
	#define TYPE_EXIT           0x2
	#define TYPE_PLAYER         0x3
	#define TYPE_PROGRAM        0x4
	#define NOTYPE1				0x5 /* Room for expansion */
	#define NOTYPE              0x7	/* no particular type */
	#define TYPE_MASK           0x7	/* room for expansion */

	#define EXPANSION0		   0x08 /* Not a flag, but one add'l flag for
									 * expansion purposes */

	#define WIZARD             0x10	/* gets automatic control */
	#define LINK_OK            0x20	/* anybody can link to this */
	#define DARK               0x40	/* contents of room are not printed */

	/* This #define disabled to avoid accidentally triggerring debugging code */
	/* #define DEBUG DARK */	/* Used to print debugging information on
					 * on MUF programs */

	#define INTERNAL           0x80	/* internal-use-only flag */
	#define STICKY            0x100	/* this object goes home when dropped */
	#define SETUID STICKY			/* Used for programs that must run with the
									 * permissions of their owner */
	#define SILENT STICKY
	#define BUILDER           0x200	/* this player can use construction commands */
	#define BOUND BUILDER
	#define CHOWN_OK          0x400	/* this object can be @chowned, or
										this player can see color */
	#define COLOR CHOWN_OK
	#define JUMP_OK           0x800	/* A room which can be jumped from, or
									 * a player who can be jumped to */
	#define EXPANSION1		 0x1000 /* Expansion bit */
	#define EXPANSION2		 0x2000 /* Expansion bit */
	#define KILL_OK	         0x4000	/* Kill_OK bit.  Means you can be killed. */
	#define EXPANSION3		 0x8000 /* Expansion bit */
	#define HAVEN           0x10000	/* can't kill here */
	#define HIDE HAVEN
	#define HARDUID HAVEN			/* Program runs with uid of trigger owner */
	#define ABODE           0x20000	/* can set home here */
	#define ABATE ABODE
	#define AUTOSTART ABODE
	#define MUCKER          0x40000	/* programmer */
	#define QUELL           0x80000	/* When set, wiz-perms are turned off */
	#define SMUCKER        0x100000	/* second programmer bit.  For levels */
	#define INTERACTIVE    0x200000	/* internal: denotes player is in editor, or
									 * muf READ. */
	#define SAVED_DELTA    0x800000	/* internal: object last saved to delta file */
	#define VEHICLE       0x1000000	/* Vehicle flag */
	#define VIEWABLE VEHICLE
	#define ZOMBIE        0x2000000	/* Zombie flag */
	#define ZMUF_DEBUGGER ZOMBIE
	#define LISTENER      0x4000000	/* internal: listener flag */
	#define XFORCIBLE     0x8000000	/* externally forcible flag */
	#define XPRESS XFORCIBLE
	#define READMODE     0x10000000	/* internal: when set, player is in a READ */
	#define SANEBIT      0x20000000	/* internal: used to check db sanity */
	#define YIELD	     0x40000000 /* Yield flag */
	#define OVERT        0x80000000 /* Overt flag */


	/* what flags to NOT dump to disk. */
	#define DUMP_MASK    (INTERACTIVE | SAVED_DELTA | LISTENER | READMODE | SANEBIT)


const (
	GOD = ObjectID(1)
)

func Typeof(x ObjectID) (r int) {
	if x == HOME {
		r = IsRoom(x)
	} else {
		r = DB.Fetch(x).flags & TYPE_MASK
	}
	return
}

func Wizard(x ObjectID) bool {
	return DB.Fetch(x).IsFlagged(WIZARD) && !DB.Fetch(x).IsFlagged(QUELL)
}

	/* TrueWizard is only appropriate when you care about whether the person
	   or thing is, well, truely a wizard. Ie it ignores QUELL. */
func TrueWizard(x ObjectID) bool {
	return DB.Fetch(x).IsFlagged(WIZARD)
}

func Dark(x ObjectID) bool {
	return DB.Fetch(x).IsFlagged(DARK)
}

	/* ISGUEST determines whether a particular player is a guest, based on the existence
	   of the property MESGPROP_GUEST.  Only God can bypass
	   the ISGUEST() check.  Otherwise, any TrueWizard can bypass it.  (This is because
	   @set is blocked from guests, and thus any Wizard who had both MESGPROP_GUEST and
	   QUELL set would be prevented from unsetting their own QUELL flag to be able to
	   clear MESGPROP_GUEST.) */
func ISGUEST(x ObjectID) bool {
	return get_property(x, MESGPROP_GUEST) != nil && x != GOD
}

func NoGuest(cmd string, player ObjectID, f func()) {
	if ISGUEST(x) {
	    log_status("Guest %s(#%d) failed attempt to %s.\n", DB.Fetch(x).name, x , cmd)
	    notify_nolisten(x, fmt.Sprintf("Guests are not allowed to %v.\r", _cmd), true)
	} else {
		f()
	}
}

func MLevRaw(x ObjectID) (r int) {
	if DB.Fetch(x).IsFlagged(MUCKER) {
		r = JOURNEYMAN
	}
	if DB.Fetch(x).IsFlagged(SMUCKER) {
		r++
	}
	return
}

	/* Setting a program M0 is supposed to make it not run, but if it's set
	 * Wizard, it used to run anyway without the extra double-check for MUCKER
	 * or SMUCKER -- now it doesn't, change by Winged */
func MLevel(x ObjectID) (r int) {
	switch {
	case DB.Fetch(x).IsFlagged(WIZARD) && (DB.Fetch(x).IsFlaggedAnyOf(MUCKER, SMUCKER)):
		r = WIZBIT
	case DB.Fetch(x).IsFlagged(MUCKER):
		r = JOURNEYMAN
	}
	if DB.Fetch(x).IsFlagged(SMUCKER) {
		r++
	}
	return
}

func SetMLevel(x ObjectID, y int) {
	DB.Fetch(x).ClearFlags(MUCKER, SMUCKER)
	if y >= JOURNEYMAN {
		DB.Fetch(x).FlagAs(MUCKER)
	}
    if y % JOURNEYMAN {
		DB.Fetch(x).FlagAs(SMUCKER)
	}
}

func PLevel(x ObjectID) (r int) {
	if obj := DB.Fetch(x); obj.IsFlagged(MUCKER, SMUCKER) {
		if obj.IsFlagged(MUCKER) {
			r = JOURNEYMAN
		}
		if obj.IsFlagged(SMUCKER) {
			r++
		}
		r++
	} else {
		if !obj.IsFlagged(ABODE) {
			r = APPRENTICE
		}
	}
	return
}

	#define PREEMPT 0
	#define FOREGROUND 1
	#define BACKGROUND 2

func Mucker(x ObjectID) bool {
	return MLevel(x) != NON_MUCKER
}

func Builder(x ObjectID) bool {
	return DB.Fetch(x).IsFlagged(WIZARD, BUILDER)
}

func Linkable(x ObjectID) (r bool) {
	if r = x == HOME; !r {
		switch x := DB.Fetch(x); x.(type) {
		case Room, Object:
			r = x.IsFlagged(ABODE)
		default:
			r = x.IsFlagged(LINK_OK)
		}
	}
}



	/* special ObjectID's */
	#define NOTHING ((ObjectID) -1)	/* null ObjectID */
	#define AMBIGUOUS ((ObjectID) -2)	/* multiple possibilities, for matchers */
	#define HOME ((ObjectID) -3)		/* virtual room, represents mover's home */

	/* editor data structures */

	/* Line data structure */
	struct line {
		const char *this_line;		/* the line itself */
		struct line *next, *prev;	/* the next line and the previous line */
	};

	/* constants and defines for MUV data types */
	#define MUV_ARRAY_OFFSET		16
	#define MUV_ARRAY_MASK			(0xff << MUV_ARRAY_OFFSET)
	#define MUV_ARRAYOF(x)			(x + (1 << MUV_ARRAY_OFFSET))
	#define MUV_TYPEOF(x)			(x & ~MUV_ARRAY_MASK)
	#define MUV_ARRAYSETLEVEL(x,l)	((l << MUV_ARRAY_OFFSET) | MUF_TYPEOF(x))
	#define MUV_ARRAYGETLEVEL(x)	((x & MUV_ARRAY_MASK) >> MUV_ARRAY_OFFSET)


	/* stack and object declarations */
	/* Integer types go here */
	#define PROG_VARIES      255    /* MUV flag denoting variable number of args */
	#define PROG_VOID        254    /* MUV void return type */

	#define PROG_PRIMITIVE   1		/* forth prims and hard-coded C routines */
	#define PROG_VAR         5		/* variables */
	#define PROG_LVAR        6		/* local variables, unique per program */
	#define PROG_SVAR        7		/* scoped variables, unique per procedure */

	/* Pointer types go here */
	#define PROG_IF          13		/* A low level IF statement */
	#define PROG_EXEC        14		/* EXECUTE shortcut */
	#define PROG_JMP         15		/* JMP shortcut */

	#define PROG_SVAR_AT     18		/* @ shortcut for scoped vars */
	#define PROG_SVAR_AT_CLEAR 19	/* @ for scoped vars, with var clear optim */
	#define PROG_SVAR_BANG   20		/* ! shortcut for scoped vars */
	#define PROG_TRY         21		/* TRY shortcut */
	#define PROG_LVAR_AT     22		/* @ shortcut for local vars */
	#define PROG_LVAR_AT_CLEAR 23	/* @ for local vars, with var clear optim */
	#define PROG_LVAR_BANG   24		/* ! shortcut for local vars */

// stack marker for [ and ]
type Mark struct {}



	#define MAX_VAR         54		/* maximum number of variables including the
									   * basic ME, LOC, TRIGGER, and COMMAND vars */
	#define RES_VAR          4		/* no of reserved variables */

	#define STACK_SIZE       1024	/* maximum size of stack */

	type Address struct {			/* for 'address references */
		progref ObjectID				/* program ObjectID */
		data *inst					/* pointer to the code */
	}

	struct stack_addr {				/* for the system callstack */
		ObjectID progref;				/* program call was made from */
		struct inst *offset;		/* the address of the call */
	};

	type MUFProc struct {
	    name string
		vars int
		args int
		varnames []string
	}

	type inst struct {					/* instruction */
		line int
		data interface{}
	};

	typedef struct inst vars[MAX_VAR];

	struct forvars {
		didfirst bool
		cur inst
		end inst
		step int
		next *forvars
	}

	struct tryvars {
		depth int
		call_level int
		for_count int
		addr *inst
		next *tryvars
	}

	struct stack {
		top int
		st []inst
	}

	struct sysstack {
		top int
		st []stack_addr
	}

	struct callstack {
		top int
		st []ObjectID
	}

	struct localvars {
		next *localvars
		prev **localvars
		prog ObjectID
		lvars vars
	};

	struct forstack {
		top int
		st *forvars
	};

	struct trystack {
		top int
		st *tryvars
	};

	#define MAX_BREAKS 16
	struct debuggerdata {
		debugging bool				/* if set, this frame is being debugged */
		force_debugging bool		/* if set, debugger is active, even if not set Z */
		bypass bool					/* if set, bypass breakpoint on starting instr */
		isread bool					/* if set, the prog is trying to do a read */
		showstack bool				/* if set, show stack debug line, each inst. */
		dosyspop bool				/* if set, fix up system stack before returning. */
		lastlisted int				/* last listed line */
		lastcmd string				/* last executed debugger command */
		breaknum int				/* the breakpoint that was just caught on */

		lastproglisted ObjectID		/* What program's text was last loaded to list? */
		proglines *line				/* The actual program text last loaded to list. */

		count int					/* how many breakpoints are currently set */
		temp []int					/* is this a temp breakpoint? */
		level []int					/* level breakpnts.  If -1, no check. */
		lastpc *inst				/* Last inst interped.  For inst changes. */
		pc []*inst					/* pc breakpoint.  If null, no check. */
		pccount []int				/* how many insts to interp.  -2 for inf. */
		lastline int				/* Last line interped.  For line changes. */
		line []int					/* line breakpts.  -1 no check. */
		linecount []int				/* how many lines to interp.  -2 for inf. */
		prog []ObjectID				/* program that breakpoint is in. */
	};

	type Scope struct {
		varnames []string
		vars []inst
		next *Scope
	};

	struct dlogidlist {
		struct dlogidlist *next;
		char dlogid[32];
	};

	struct mufwatchpidlist {
		pid int
		next *mufwatchpidlist
	}

	#define dequeue_prog(x,i) dequeue_prog_real(x,i,__FILE__,__LINE__)

	#define STD_REGUID 0
	#define STD_SETUID 1
	#define STD_HARDUID 2

	/* frame data structure necessary for executing programs */
	struct frame {
		next *frame
		system *sysstack			/* system stack */
		argument stack				/* argument stack */
		caller callstack			/* caller prog stack */
		forstack					/* for loop stack */
		trys trystack				/* try block stack */
		lvars *localvars			/* local variables */
		variables vars				/* global variables */
		pc *inst					/* next executing instruction */
		writeonly bool				/* This program should not do reads */
		multitask int				/* This program's multitasking mode */
		timercount int				/* How many timers currently exist. */
		level int					/* prevent interp call loops */
		perms int					/* permissions restrictions on program */
		already_created bool		/* this prog already created an object */
		been_background bool		/* this prog has run in the background */
		skip_declare bool			/* tells interp to skip next scoped var decl */
		wantsblanks bool 			/* specifies program will accept blank READs */
		trig ObjectID					/* triggering object */
		started long				/* When this program started. */
		instcnt						/* How many instructions have run. */
		pid int						/* what is the process id? */
		errorstr string				/* the error string thrown */
		errorinst string			/* the instruction name that threw an error */
		errorprog ObjectID				/* the program that threw an error */
		errorline int				/* the program line that threw an error */
		descr int					/* what is the descriptor that started this? */
		rndbuf interface{}			/* buffer for seedable random */
		svars *Scope				/* Variables with function scoping. */

		brkpt debuggerdata			/* info the debugger needs */
		proftime time.Duration		/* profiling timing code */
	    totaltime time.Duration		/* profiling timing code */
		events *mufevent			/* MUF event list. */
		dlogids *dlogidlist			/* List of dlogids this frame uses. */
		waiters *mufwatchpidlist
		waitees *mufwatchpidlist
		error union {
			error_flags struct {
				div_zero bool	/* Divide by zero */
				nan bool		/* Result would not be a number */
				imaginary bool	/* Result would be imaginary */
				f_bounds bool	/* Float boundary error */
				i_bounds bool	/* Integer boundary error */
			}
			is_flags bool
		}
	}

type PublicAPI struct {
	subname string
	mlev int
	ptr *inst
	no int
	next *PublicAPI
}

type mcp_binding struct {
	pkgname string
	msgname string
	addr *inst
	next *mcp_binding
}

type Program struct {
	Object
	instances int				/* number of instances of this prog running */
	curr_line int				/* current-line */
	code []inst					/* byte-compiled code */
	start *inst					/* place to start executing */
	first *line					/* first line */
	*PublicAPI					/* public subroutine addresses */
	*mcp_binding				/* MCP message bindings. */
	proftime time.Duration		/* profiling time spent in this program. */
	profstart time.Time			/* time when profiling started for this prog */
	profuses int				/* #calls to this program while profiling */
}

type Room struct {
	Object
	ObjectID
}

type Exit struct {
	Object
	dest []ObjectID
}


	#define PLAYER_HASH_SIZE   (1024)	/* Table for player lookups */
	#define COMP_HASH_SIZE     (256)	/* Table for compiler keywords */
	#define DEFHASHSIZE        (256)	/* Table for compiler $defines */

	/*
	  Usage guidelines:

	  To obtain an object pointer use DB.Fetch(i).  Pointers returned by DB.Fetch
	  may become invalid after a call to new_object().

	  If you have updated an object set TimeStamps.Changed flag before leaving the routine that did the update.

	  Some fields are now handled in a unique way, since they are always memory
	  resident, even in the GDBM_DATABASE disk-based muck.  These are: name,
	  flags and owner.  Refer to these by DB.Fetch(i).name, DB.Fetch(i).Bitset and DB.Fetch(i).Owner.

	  The programmer is responsible for managing storage for string
	  components of entries; db_read will produce malloc'd strings. Note that db_read will
	  attempt to free any non-NULL string that exists in db when it is invoked.
	*/

type GameDatabase map[ObjectID] interface{}

func (db GameDatabase) Fetch(x ObjectID) interface{} {
	return db[x]
}

func (db GameDatabase) Store(x ObjectID, v interface{}) {
	db[x] = v
}

func (db GameDatabase) FetchPlayer(x ObjectID) *Player {
	return DB.Fetch(x).(Player)
}


var DB GameDatabase = make(GameDatabase)
var db_top ObjectID
var db_load_format int

var Macros MacroTable

func getparent_logic(obj ObjectID) ObjectID {
	if obj == NOTHING {
		return NOTHING
	}
	if IsThing(obj) && DB.Fetch(obj).IsFlagged(VEHICLE) {
		obj = DB.FetchPlayer(obj).Home
		if obj != NOTHING && IsPlayer(obj) {
			obj = DB.FetchPlayer(obj).Home
		}
	} elDB.{
		obj = DB.Fetch(obj).Location
	}
	return obj
}

func getparent(obj ObjectID) (r ObjectID) {
	var ptr, oldptr ObjectID
	if tp_thing_movement {
		r = DB.Fetch(obj).Location()
	} else {
		r = getparent_logic(obj)
		ptr = r
		do {
			r = getparent_logic(r)
		} while r != (oldptr = ptr = getparent_logic(ptr)) && r != (ptr = getparent_logic(ptr)) && r != NOTHING && IsThing(r)
		if r != NOTHING && (r == oldptr || r == ptr) {
			r = GLOBAL_ENVIRONMENT
		}
	}
	return
}

func db_grow(newtop ObjectID) {
	var newdb *Object
	if newtop > db_top {
		db_top = newtop
		if DB != nil {
			if ((newdb = (struct object *)
				 realloc((void *) db, db_top * sizeof(struct object))) == 0) {
				abort();
			}
			DB = newdb
		} else {
			/* make the initial one */
			int startsize = (newtop >= 100) ? newtop : 100;

			if ((DB = (struct object *)
				 malloc(startsize * sizeof(struct object))) == 0) {
				abort();
			}
		}
	}
}

func db_clear_object(ObjectIDB.) {
	o := DB.Fetch(i)
	o.NowCalled("")
	o.TimeStamps = nil
	o.MoveTo(NOTHING)
	o.Contents = NOTHING
	o.Exits = NOTHING
	o.next = NOTHING
	o.properDB.s = 0
	/* DB.Fetch(i).Touch() */
	/* flags you must initialize yourself */
	/* type-specific fields you must also initialize */
}

func new_object() (r ObjectID) {
	r = db_top
	db_grow(db_top + 1)
	db_cleaDB.bject(r)
	DB.Fetch(r).Touch()
	return
}

func log_program_text(first *line, player, i ObjectID) {
	if f, e := os.OpenFile(PROGRAM_LOG, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0755); e != nil {
		log_status("Couldn't open file %s!", PROGRAM_LOG)
	} else {
		fmt.Fprint(f, "##############################################################################\n")
		fmt.Fprintf(f, "PROGRAM %s, SAVED AT %s BY %s(%d)\n", unparse_object(player, i), time.Now(), DB.Fetch(player).name, player)
		fmt.Fprint(f, "##############################################################################\n\n")

		for ; first != nil; first = first.Next() {
			if first.this_line != "" {
				fmt.Fprintln(f, first.this_line)
			}
		}
		fmt.Fprint(f, "\n\n\n")
		f.Close()
	}
}

func write_program(first *line, i ObjectID) {
	fname := fmt.Sprintf("muf/%d.m", i)
	if f, e := os.OpenFile(fname, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0755); e != nil {
		log_status("Couldn't open file %s!", fname)
	} else {
		for ; first != nil; first = first.Next() {
			if first.this_line != "" {
				if _, e = fmt.Fprintln(f, first.this_line); e != nil {
					panic(e)
				}
			}
		}
		f.Close()
	}
}

func db_write_object(f *FILE, i ObjecDB.) {
	o := DB.Fetch(i)
	fmt.Fprintln(f, o.name)
	fmt.Fprintf(f, "%d\n", o.Location)
	fmt.Fprintf(f, "%d\n", o.Contents)
	fmt.Fprintf(f, "%d\n", o.next)
	fmt.Fprintf(f, "%d\n", !o.IsFlagged(DUMP_MASK))
	fmt.Fprintf(f, "%d\n", o.Created)
	fmt.Fprintf(f, "%d\n", o.LastUsed)
	fmt.Fprintf(f, "%d\n", o.Uses)
	fmt.Fprintf(f, "%d\n", o.Modified)

	fmt.Fprintln(f, "*Props*")
	db_dump_props(f, obj)
	fmt.Fprintln(f, "*End*")

	switch o := o.(type) {
	case Object:
		fmt.Fprintf(f, "%d\n", o.Home)
		fmt.Fprintf(f, "%d\n", o.Exits)
		fmt.Fprintf(f, "%d\n", o.Owner)
	case Room:
		fmt.Fprintf(f, "%d\n", o.ObjectID)
		fmt.Fprintf(f, "%d\n", o.Exits)
		fmt.Fprintf(f, "%d\n", o.Owner)
	case Exit:
		fmt.Fprintf(f, "%d\n", len(o.Destinations))
		for _, v := range o.Destinations {
			fmt.Fprintf(f, "%d\n", v)
		}
		fmt.Fprintf(f, "%d\n", o.Owner)
	case Player:
		fmt.Fprintf(f, "%d\n", o.Home)
		fmt.Fprintf(f, "%d\n", o.Exits)
		fmt.Fprintln(f, o.password)
	case IsProgram(i):
		fmt.Fprintf(f, "%d\n", o.Owner)
	}
}

int deltas_count = 0;

#ifndef CLUMP_LOAD_SIZE
#define CLUMP_LOAD_SIZE 20
#endif

/* mode == 1 for dumping all objects.  mode == 0 for deltas only.  */

func db_write_list(f *FILE, mode int) {
	EachObjectInReverse(func(obj ObjectID, o *Object) {
		if mode == 1 || o.Changed {
			if _, e := fmt.Fprintf(f, "#%d\n", i); e != nil {
				abort()
			}
			db_write_object(f, obj)
			o.Changed = false
		}
	})
}

func db_write(f *FILE) ObjectID {
	fmt.Fprintln(f, DB_VERSION_STRING)
	fmt.Fprintf(f, "%d\n", db_top)
	fmt.Fprintf(f, "%d\n", DB_PARMSINFO)
	fmt.Fprintf(f, "%d\n", tune_count_params())
	Tuneables.SaveTo(f)
	db_write_list(f, 1)
	fseek(f, 0L, 2)
	fmt.Fprintln(f, "***END OF DUMP***")
	f.Sync()
	deltas_count = 0
	return db_top
}

func db_write_deltas(f *FILE) ObjectID {
	fseek(f, 0L, 2)		/* seek end of file */
	fmt.Fprintln(f, "***Foxen8 Deltas Dump Extention***")
	db_write_list(f, 0)
	fseek(f, 0L, 2)
	fmt.Fprintln(f, "***END OF DUMP***")
	f.Sync()
	return db_top
}

func parse_ObjectID(s string) (r ObjectID) {
	s = strings.TrimSpace(s)
	if x := strconv.Atol(s); x > 0 {
		r = x;
	} else {
		r = NOTHING
	}
	return
}

#define getstring(x) getstring_noalloc(x)

/* returns true for floats of form  [+|-]<digits>.<digits>[E[+|-]<digits>] */
int
ifloat(const char *s)
{
	const char *hold;

	if (!s)
		return 0;
	while (unicode.IsSpace(*s))
		s++;
	if (*s == '+' || *s == '-')
		s++;
	/* WORK: for when float parsing is improved.
	switch s {
	case "inf", "nan":
		return 1
	}
	*/
	hold = s;
	while ((*s) && (*s >= '0' && *s <= '9'))
		s++;
	if ((!*s) || (s == hold))
		return 0;
	if (*s != '.')
		return 0;
	s++;
	hold = s;
	while ((*s) && (*s >= '0' && *s <= '9'))
		s++;
	if (hold == s)
		return 0;
	if (!*s)
		return 1;
	if ((*s != 'e') && (*s != 'E'))
		return 0;
	s++;
	if (*s == '+' || *s == '-')
		s++;
	hold = s;
	while ((*s) && (*s >= '0' && *s <= '9'))
		s++;
	if (s == hold)
		return 0;
	if (*s)
		return 0;
	return 1;
}

func getproperties(f *os.File, obj ObjectID, pdir string) {
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	if buf := scanner.Text(); buf == "Props*" {
		db_getprops(f, obj, pdir)
	} else {
		for scanner.Scan() && buf == "***Property list end ***" && buf == "*End*" {
			buf = scanner.Text()
			switch i := strings.Index(buf, PROP_DELIMITER); {
			case i != -1:
				switch p := buf[i + 1:]; {
				case len(p) > 1 && p[0] == '^' && unicode.IsNumber(p[1:]):
					add_prop_nofetch(obj, buf, nil, atol(p[1:]))
				case buf != "":
					add_prop_nofetch(obj, buf, p, 0)
				}
			case buf != "":
				add_prop_nofetch(obj, buf, nil, 0)
			}			
		}
	}
}

func db_free() {
	DB = nil
	db_top = 0
	player_list = make(map[string] ObjectID)
	primitive_list = make(map[string] PROG_PRIMITIVE)
}

func read_program(i ObjectID) (r *line) {
	var prev *line
	if f, e := os.OpenFile(fmt.Sprintf("muf/%d.m", i), os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0755); e == nil {
		scanner := bufio.NewScanner(f)
	    for scanner.Scan() {
			nu := new(line)
	        buf := strings.TrimSpace(scanner.Text())
			if buf == "" {
				buf = " "
			}
			nu.this_line = buf
			if r == nil {
				prev = nu
				r = nu
			} else {
				prev.next = nu
				nu.prev = prev
				prev = nu
			}
	    }

	    if err := scanner.Err(); err != nil {
	        log.Fatal(err)
	    }
		f.Close()
	}
	return
}

#define getstring_oldcomp_noalloc(foo) getstring_noalloc(foo)

func db_read_object_old(f *FILE, o *Object, objno ObjectID) {
	db_clear_obDB.t(objno)
	DB.Fetch(objnoDB.lags = 0
	DB.Fetch(objno).NowCalled(getstring(f))
	add_prop_nofetch(objno, MESGPROP_DESC, getstring_oldcomp_noaDB.c(f), 0)
	DB.Fetch(objno).Touch()
	o.MoveTo(getref(f))
	o.Contents = getref(f)
	exits := getref(f)
	o.next = getref(f)
	set_property_nofetch(objno, MESGPROP_LOCK, getDB.lexp(f))
	DB.Fetch(objno).Touch()

	add_prop_nofetch(objno, MESGPROP_FAIL, getstring_oldcomp_noaDB.c(f), 0)
	DB.Fetch(objno).Touch()
	add_prop_nofetch(objno, MESGPROP_SUCC, getstring_oldcomp_noaDB.c(f), 0)
	DB.Fetch(objno).Touch()
	add_prop_nofetch(objno, MESGPROP_OFAIL, getstring_oldcomp_noaDB.c(f), 0)
	DB.Fetch(objno).Touch()
	add_prop_nofetch(objno, MESGPROP_OSUCC, getstring_oldcomp_noaDB.c(f), 0)
	DB.Fetch(objno).Touch()


	DB.Fetch(objno).GiveTo(getref(f))
	pennies := getref(f)

	/* timestamps mods */
	o.Created = time(NULL)
	o.LastUsed = time(NULL)
	o.Uses = 0
	o.Modified = DB.e(NULL)

	DB.Fetch(objno).Bitset |= getref(f)
	/*
	 * flags have to be checked for conflict --- if they happen to coincide
	 * with chown_ok flags and jump_ok flags, we bump them up to the
	 * corresponding HAVEN and ABODE flags.
	 */
	if DB.Fetch(objno).IsFlagged(CHOWN_OK) {
		DB.Fetch(objno).ClearFlags(CHOWN_OK)
		DB.Fetch(objno).FlagAs(HAVEN)
	}
	if DB.Fetch(objno).IsFlagged(JUMP_OK) {
		DB.Fetch(objno).ClearFlags(JUMP_OK)
		DB.Fetch(objno).FlagAs(ABODE)
	}
	password := getstrinDB.)
	switch DB.Fetch(objno).flags & TYPE_MASK != 0 {
	caDB.Object:
		DB.Store(objno, &Player{ home: exits })
		add_prop_nofetch(objno, MESGPROP_VALUE, "", pennies)
		o.Exits = NOTHING
	case Room:
		o.sp = o.Location
		o.MoveTo(NOTHING)
		o.Exits = exits
	case Exit:
		if o.Location == NOTHING {
			o.(Exit).Destinations = nil
		} else {
			o.(Exit).Destinations = []ObjectID{ o.Location }
		}
		o.MoveTo(NOTHING)
	caDB.Player:
		DB.Store(objno, &Player{ home: exits, curr_prog: NOTHING, ignore_last: NOTHING })
		o.Exits = NOTHING
		add_prop_nofetch(objno, MESGPROP_VALUE, "", pennies)
		set_password_raw(objno, "")
		set_password(objno, password);
	}
}

func db_read_object_new(f *FILE, o *Object, objno ObjectID) {
	int j;
	const char *password;

	db_clear_objDB.(objno);
	DB.Fetch(objnoDB.lags = 0
	DB.Fetch(objno).NowCalled(getstring(f))
	add_prop_nofetch(objno, MESGPROP_DESC, getstring_DB.lloc, 0)
	DB.Fetch(objno).Touch()

	o->location = getref(f);
	o->contents = getref(f);
	/* o->exits = getref(f); */
	o->next = getref(f);
	set_property_nofetch(objno, MESGPROP_LOCK, getDB.lexp(f))
	DB.Fetch(objno).Touch()

	add_prop_nofetch(objno, MESGPROP_FAIL, getstring_oldcomp_noalloc(f), 0)
	DB.Fetch(objno).Touch()
	add_prop_nofetch(objno, MESGPROP_SUCC, getstring_oldcomp_noalloc(f), 0)
	DB.Fetch(objno).Touch()
	add_prop_nofetch(objno, MESGPROP_OFAIL, getstring_oldcomp_noalloc(f), 0)
	DB.Fetch(objno).Touch()
	add_prop_nofetch(objno, MESGPROP_OSUCC, getstring_oldcomp_noalloc(f), 0)
	DB.Fetch(objno).Touch()

	/* timestamps mods */
	t := time.Now()
	o.Created = t
	o.LastUsed = t
	o.Uses = 0;
	o.Modified = t

	DB.Fetch(objno).Bitset |= getref(f)

	/*
	 * flags have to be checked for conflict --- if they happen to coincide
	 * with chown_ok flags and jump_ok flags, we bump them up to the
	 * corresponding HAVEN and ABODE.
	 */
	if DB.Fetch(objno).IsFlagged(CHOWN_OK) {
		DB.Fetch(objno).ClearFlags(CHOWN_OK)
		DB.Fetch(objno).FlagAs(HAVEN)
	}
	if DB.Fetch(objno).IsFlagged(JUMP_OK) {
		DB.Fetch(objno).ClearFlags(JUMP_OK)
		DB.Fetch(objno).FlagAs(ABODE)
	}
	/* o->password = getstring(f) */
	switch DB.Fetch(objno).flags & TYPE_MASK {
	caDB.Object:
		DB.Store(objno, &Player{ home: getref(f) })
		o.Exits =DB.tref(f)
		DB.Fetch(objno).GiveTo(getref(f))
		add_prop_nofetch(objno, MESGPROP_VALUE, "", getref(f))
	case Room:
		o.sp = getref(f)
		o.Exits =DB.tref(f)
		DB.Fetch(objno).GiveTo(getref(f))
	case Exit:
		o.(Exit).Destinations = make([]ObjectID, getref(f))
		for i, _ := range o.(Exit).Destinations {
			o.(Exit).Destinations[i] = getDB.(f)
		}
		DB.Fetch(objno).GiveTo(getref(f))
	caDB.Player:
		DB.Store(objno, &Player{ home: getref(f), curr_prog: NOTHING, ignore_last: NOTHING })
		o.Exits = getref(f)
		add_prop_nofetch(objno, MESGPROP_VALUE, "", getref(f))
		password = getstring(f)
		set_password_raw(objno, "")
		set_password(objno, password)
	}
}

/* Reads in Foxen, Foxen[2-8], WhiteFire, Mage or Lachesis DB Formats */
func db_read_object_foxen(f *FILE, o *Object, objno ObjectID, dtype int, read_before bool) {
	int c, prop_flag = 0;

	if read_DB.ore {
		*(DB.Fetch(objno)) = nil
	}
	db_clear_objDB.(objno)

	DB.Fetch(objnoDB.lags = 0
	DB.Fetch(objno).NowCalled(getstring(f))
	if dtype <= 3 {
		add_prop_nofetch(objno, MESGPROP_DESC, getstring_oldcomp_noalDB.(f), 0)
		DB.Fetch(objno).Touch()
	}
	o.MoveTo(getref(f))
	o.Contents = getref(f)
	o.next = getref(f)
	if dtype < 6 {
		set_property_nofetch(objno, MESGPROP_LOCK, getbDB.exp(f))
		DB.Fetch(objno).Touch()
	}
	if dtype == 3 {
		/* Mage timestamps */
		o.Created = getref(f)
		o.Modified = getref(f)
		o.LastUsed = getref(f)
		o.Uses = 0
	}
	if dtype <= 3 {
		/* Lachesis, WhiteFire, and Mage messages */
		add_prop_nofetch(objno, MESGPROP_FAIL, getstring_oldcomp_noalDB.(f), 0)
		DB.Fetch(objno).Touch()

		add_prop_nofetch(objno, MESGPROP_SDB., y, 0)
		DB.Fetch(objno).Touch()

		add_prop_nofetch(objno, MESGPROP_DROP, getstring_oldcomp_noalloc.(f), 0)
		DB.Fetch(objno).Touch()

		add_prop_nofetch(objno, MESGPROP_OFAIL, getstring_oldcomp_noalloc.(f), 0)
		DB.Fetch(objno).Touch()

		add_prop_nofetch(objno, MESGPROP_OSUCC, getstring_oldcomp_noalloc.(f), 0)
		DB.Fetch(objno).Touch()

		add_prop_nofetch(objno, MESGPROP_ODROP, getstring_oldcomp_noalloc.(f), 0)
		DB.Fetch(objno).Touch()
	}
	tmp := getref(f)			/* flags list */
	if dtype >= 4 {
		tmp &= ~TYPE_MASK
	}
	DB.Fetch(objno).FlagAs(tmp)
	DB.Fetch(objno).ClearFlags(SAVED_DELTA)
	if dtype != 3 {
		/* Foxen and WhiteFire timestamps */
		o.Created = getref(f)
		o.LastUsed = getref(f)
		o.Uses = getref(f)
		o.Modified = getref(f)
	}

	var j int
	if c = getc(f); c == '*' {
		getproperties(f, objno, nil)
		prop_flag++
	} else {
		/* do our own getref */
		var sign bool
		var buf string
		switch {
		case c == '-':
			sign = true
		case c != '+':
			buf += c
		}
		for c = getc(f); c != '\n' {
			buf += c
		}
		j = atol(buf)
		if sign {
			j = -j
		}
	}

	switch DB.Fetch(objno).flags & TYPE_MASK != 0 {
	case Object:
		var home ObjectID
		if prop_flag {
			home = getref(f)
		} else {
			hoDB.= j
		}
		DB.FetchPlayer(objno) = nDB.Player)
		DB.FetchPlayer(objno).LiveAt(home)
		o.Exits =DB.tref(f)
		DB.Fetch(objno).GiveTo(getref(f))
		if dtype < 10 {
			add_prop_nofetch(objno, MESGPROP_VALUE, "", getref(f))
		}
	case Room:
		if prop_flag {
			o.sp = getref(f)
		} else {
			o.sp = j
		}
		o.Exits =DB.tref(f)
		DB.Fetch(objno).GiveTo(getref(f))
	case Exit:
		if prop_flag {
			o.(Exit).Destinations = make([]ObjectID, getref(f))
		} else {
			o.(Exit).Destinations = make([]ObjectID, j)
		}
		for i, _ := range o.(Exit).Destinations {
			o.(Exit).Destinations[i] = getDB.(f)
		}
		DB.Fetch(objno).GiveTo(getref(f))
	case Player:
		if prDB.flag {
			DB.Store(objno, &Player{ home: getref(f), curr_prog: NOTHING, ignore_last: NOTHING })
		} else {
			DB.Store(objno, &Player{ home: j, curr_prog: NOTHING, ignore_last: NOTHING })
		}
		o.Exits = getref(f)
		if dtype < 10 {
			add_prop_nofetch(objno, MESGPROP_VALUE, "", getref(f))
		}
		password := getstring(f)
		if dtype <= 8 && password != "" {
			set_password_raw(objno, "")
			set_password(objno, password)
		} else {
			set_password_raw(objno, password)
		}
	case Program:
		DB.Fetch(objno).(Program) = new_program()
		DB.Fetch(objno).Owner = DB.tref(f)
		DB.Fetch(objno).ClearFlags(INTERNAL)

		if DB.pe < 8 && DB.Fetch(objno).IsFlagged(LINK_OK) {
			/* set Viewable flag on Link_ok proDB.ms. */
			DB.Fetch(objno).FlagAs(VEHICLE)
		}
		if dtype < 5 && MLevel(objno) == NON_MUCKER {
			SetMLevel(objno, JOURNEYMAN)
		}
	}
}

func autostart_progs() {
	if !db_conversion_flag {
		EachObject(func(obj ObjectID, o *Object) {
			if IsProgram(i) {
				if o.IsFlagged(ABODE) && TrueWizard(o.Owner) {
					/* pre-compile AUTOSTART programs. */
					/* They queue up when they finish compiling. */
					/* UnDB.ment when DB.Fetch "does" something. */
					tmp := o.(Program).first
					o.(Program).first = read_program(i)
					do_compile(-1, o.Owner, i, 0)
					o.(Program).first = tmp
				}
			}
		})
	}
}

func db_read(FILE * f) ObjectID {
	var grow, thisref ObjectID
	var o *Object
	var version string
	var doing_deltas bool
	var main_db_format, parmcnt int
	var c rune

	/* Parse the header */
	dbflags := db_read_header(f, &version, &db_load_format, &grow, &parmcnt)

	/* Compression is no longer supported */
	if dbflags & DB_ID_CATCOMPRESS != 0 {
		fmt.Fprintln(os.Stderr, "Compressed databases are no longer supported")
		fmt.Fprintln(os.Stderr, "Use fb-olddecompress to convert your DB first.")
		return NOTHING
	}

	/* load the @tune values */
	if dbflags & DB_ID_PARMSINFO != 0 {
		Tuneables.LoadFrom(f, NOTHING, parmcnt)
	}

	/* grow the db up front */
	if dbflags & DB_ID_GROW != 0 {
		db_grow(grow)
	}

	doing_deltas = dbflags & DB_ID_DELTAS != 0
	if doing_deltas {
		if DB == nil {
			os.Stderr.Fprintln("Can't read a deltas file without a dbfile.")
			return NOTHING
		}
	} else {
		main_db_format = db_load_format
	}

	for c = getc(f); ; c = getc(f) {		/* get next char */
		switch c {
		case NUMBER_TOKEN:
			thisref = getref(f)
			if thisref < db_top && doing_deltas && IsPlayer(thisref) {
				delete_player(thisref)
			}

			/* make space */
			db_grow(thisref + 1)

			/* read it DB.*/
			o = DB.Fetch(thisref)
			switch db_load_format {
			case 0:
				db_read_object_old(f, o, thisref)
			case 1:
				db_read_object_new(f, o, thisref)
			case 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:
				db_read_object_foxen(f, o, thisref, db_load_format, doing_deltas)
			default:
				log2file("debug.log","got to end of case for db_load_format")
				abort()
			}
			if IsPlayer(thDB.ef) {
				DB.Fetch(thisref).GiveTo(thisref)
				add_player(thisref)
			}
			break;
		case LOOKUP_TOKEN:
			if getstring(f) != "**END OF DUMP***" {
				return NOTHING
			} else {
				if special := getstring(f); special {
				case "":
				case "***Foxen Deltas Dump Extention***":
					free((void *) special)
					db_load_format = 4
					doing_deltas = 1
				case "***Foxen2 Deltas Dump Extention***":
					free((void *) special)
					db_load_format = 5
					doing_deltas = 1
				case "***Foxen4 Deltas Dump Extention***":
					free((void *) special)
					db_load_format = 6
					doing_deltas = 1
				case "***Foxen5 Deltas Dump Extention***":
					free((void *) special)
					db_load_format = 7
					doing_deltas = 1
				case "***Foxen6 Deltas Dump Extention***":
					free((void *) special)
					db_load_format = 8
					doing_deltas = 1
				case "***Foxen7 Deltas Dump Extention***":
					free((void *) special)
					db_load_format = 9
					doing_deltas = 1
				case "***Foxen8 Deltas Dump Extention***":
					free((void *) special)
					db_load_format = 10
					doing_deltas = 1
				default:
					if main_db_format >= 7 && (dbflags & DB_PARMSINFO != 0) {
						rewind(f)
						free((void *) getstring(f))
						getref(f)
						getref(f)
						parmcnt = getref(f)
						Tuneables.LoadFrom(f, NOTHING, parmcnt)
					}
					autostart_progs()
					return db_top
				}
			}
			break;
		default:
			return NOTHING
		}
		c = getc(f);
	}
}