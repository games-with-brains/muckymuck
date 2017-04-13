package fbmuck

func check_remote(player, what dbref) {
	o := db.Fetch(what)
	p := db.Fetch(player)
	if mlev < JOURNEYMAN && o.Location != player && o.Location != p.Location && what != p.Location && what != player && !controls(ProgUID, what) {
		panic("Mucker Level 2 required to get remote info.")
	}
}

func valid_reference(obj dbref) bool {
	return obj > -1 && obj < db_top
}

func EachObject(f interface{}) {
	switch f := f.(type) {
	case func(dbref):
		for i := dbref(0); i < db_top; i++ {
			f(i)
		}
	case func(dbref, object):
		for i := dbref(0); i < db_top; i++ {
			f(i, db.Fetch(i))
		}
	case func(object):
		for i := dbref(0); i < db_top; i++ {
			f(db.Fetch(i))
		}
	case func(dbref) bool:
		var done bool
		for i := dbref(0); !done && i < db_top; i++ {
			done = f(i)
		}
	case func(dbref, object) bool:
		var done bool
		for i := dbref(0); !done && i < db_top; i++ {
			done = f(i, db.Fetch(i))
		}
	case func(object) bool:
		var done bool
		for i := dbref(0); !done && i < db_top; i++ {
			done = f(db.Fetch(i))
		}
	default:
		panic(f)
	}
}

func EachObjectInReverse(f interface{}) {
	switch f := f.(type) {
	case func(dbref):
		for i := db_top - 1; i > -1; i-- {
			f(i)
		}
	case func(dbref, object):
		for i := db_top - 1; i > -1; i-- {
			f(i, db.Fetch(i))
		}
	case func(object):
		for i := db_top - 1; i > -1; i-- {
			f(db.Fetch(i))
		}
	case func(dbref) bool:
		var done bool
		for i := db_top - 1; !done && i > -1; i-- {
			done = f(i)
		}
	case func(dbref, object) bool:
		var done bool
		for i := db_top - 1; !done && i > -1; i++ {
			done = f(i, db.Fetch(i))
		}
	case func(object) bool:
		var done bool
		for i := db_top - 1; !done && i > -1; i++ {
			done = f(db.Fetch(i))
		}
	default:
		panic(f)
	}
}

/*
	These helper functions operate in two modes:

		if one or more functions are passed in then each of these is executed in sequence for a valid database reference, or NOTHING is returned
		if no functions are passed in then either the valid database reference is returned or a panic occurs
*/
func valid_object(v interface{}, f ...func(dbref)) (obj dbref) {
	switch obj = oper.(dbref); {
	case valid_reference(obj):
		for _, f := range f {
			f(obj)
		}
	case f == nil:
		panic("Not a valid object reference")
	default:
		obj = NOTHING
	}
	return
}

func valid_object_or_home(oper interface{}, f ...func(dbref)) (obj dbref) {
	switch obj = oper.(dbref); {
	case obj == HOME:
		for _, f := range f {
			f(obj)
		}
	case valid_object(oper, f...) != NOTHING:
	default:
		obj = NOTHING
	}
	return
}

func valid_player(oper interface{}, f ...func(dbref)) (obj dbref) {
	switch obj = oper.(dbref); {
	case valid_reference(obj) && IsPlayer(obj):
		for _, f := raneg f {
			f(obj)
		}
	case f == nil:
		panic("Not a valid object reference")
	default:
		obj = NOTHING
	return
}

func valid_remote_object(player dbref, mlev int, oper interface{}, f ...func(dbref)) (obj dbref) {
	obj = valid_object(oper, f...)
	check_remote(obj)
	return
}

func valid_remote_player(player dbref, mlev int, oper interface{}, f ...func(dbref)) (obj dbref) {
	obj = valid_player(oper, f...)
	check_remote(obj)
	return
}



	/* max length of command argument to process_command */
	#define MAX_COMMAND_LEN 2048
	#define BUFFER_LEN ((MAX_COMMAND_LEN)*4)
	#define FILE_BUFSIZ ((BUFSIZ)*8)

	typedef int dbref;				/* offset into db */

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
	#define OBJECT_CHANGED 0x400000	/* internal: when an object is dbdirty()ed,
									 * set this */
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
	#define DUMP_MASK    (INTERACTIVE | SAVED_DELTA | OBJECT_CHANGED | LISTENER | READMODE | SANEBIT)


	typedef long object_flag_type;

const (
	GOD = dbref(1)
)

func Typeof(x dbref) (r int) {
	if x == HOME {
		r = TYPE_ROOM
	} else {
		r = db.Fetch(x).flags & TYPE_MASK
	}
	return
}

func Wizard(x dbref) bool {
	return db.Fetch(x).flags & WIZARD != 0 && db.Fetch(x).flags & QUELL == 0
}

	/* TrueWizard is only appropriate when you care about whether the person
	   or thing is, well, truely a wizard. Ie it ignores QUELL. */
func TrueWizard(x dbref) bool {
	return db.Fetch(x).flags & WIZARD != 0
}

func Dark(x dbref) bool {
	return db.Fetch(x).flags & DARK != 0
}

	/* ISGUEST determines whether a particular player is a guest, based on the existence
	   of the property MESGPROP_GUEST.  Only God can bypass
	   the ISGUEST() check.  Otherwise, any TrueWizard can bypass it.  (This is because
	   @set is blocked from guests, and thus any Wizard who had both MESGPROP_GUEST and
	   QUELL set would be prevented from unsetting their own QUELL flag to be able to
	   clear MESGPROP_GUEST.) */
func ISGUEST(x dbref) bool {
	return get_property(x, MESGPROP_GUEST) != nil && x != GOD
}

func NoGuest(cmd string, player dbref, f func()) {
	if ISGUEST(x) {
	    log_status("Guest %s(#%d) failed attempt to %s.\n", db.Fetch(x).name, x , cmd)
	    notify_nolisten(x, fmt.Sprintf("Guests are not allowed to %v.\r", _cmd), true)
	} else {
		f()
	}
}

func MLevRaw(x dbref) (r int) {
	if db.Fetch(x).flags & MUCKER != 0 {
		r = JOURNEYMAN
	}
	if db.Fetch(x).flags & SMUCKER != 0 {
		r++
	}
	return
}

	/* Setting a program M0 is supposed to make it not run, but if it's set
	 * Wizard, it used to run anyway without the extra double-check for MUCKER
	 * or SMUCKER -- now it doesn't, change by Winged */
func MLevel(x dbref) (r int) {
	switch {
	case db.Fetch(x).flags & WIZARD != 0 && (db.Fetch(x).flags & MUCKER != 0 || db.Fetch(x).flags & SMUCKER != 0):
		r = WIZBIT
	case db.Fetch(x).flags & MUCKER != 0:
		r = JOURNEYMAN
	}
	if db.Fetch(x).flags & SMUCKER != 0 {
		r++
	}
	return
}

func SetMLevel(x dbref, y int) {
	db.Fetch(x).flags &= ~(MUCKER | SMUCKER)
	if y >= JOURNEYMAN {
		db.Fetch(x).flags |= MUCKER
	}
    if y % JOURNEYMAN {
		db.Fetch(x).flags |= SMUCKER
	}
}

func PLevel(x dbref) (r int) {
	if db.Fetch(x).flags & (MUCKER | SMUCKER) != 0 {
		if db.Fetch(x).flags & MUCKER != 0 {
			r = JOURNEYMAN
		}
		if db.Fetch(x).flags & SMUCKER != 0 {
			r++
		}
		r++
	} else {
		if db.Fetch(x).flags & ABODE == 0 {
			r = APPRENTICE
		}
	}
	return
}

	#define PREEMPT 0
	#define FOREGROUND 1
	#define BACKGROUND 2

func Mucker(x dbref) bool {
	return MLevel(x) != NON_MUCKER
}

func Builder(x dbref) bool {
	return db.Fetch(x).flags & (WIZARD | BUILDER) != 0
}

func Linkable(x dbref) (r bool) {
	switch {
	case x == HOME:
		r = true
	case Typeof(x) == TYPE_ROOM || Typeof(x) == TYPE_THING:
		r = db.Fetch(x).flags & ABODE != 0
	default:
		r = db.Fetch(x).flags & LINK_OK != 0
	}
}



	/* special dbref's */
	#define NOTHING ((dbref) -1)	/* null dbref */
	#define AMBIGUOUS ((dbref) -2)	/* multiple possibilities, for matchers */
	#define HOME ((dbref) -3)		/* virtual room, represents mover's home */

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
		progref dbref				/* program dbref */
		data *inst					/* pointer to the code */
	}

	struct stack_addr {				/* for the system callstack */
		dbref progref;				/* program call was made from */
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
		st []dbref
	}

	struct localvars {
		next *localvars
		prev **localvars
		prog dbref
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

		lastproglisted dbref		/* What program's text was last loaded to list? */
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
		prog []dbref				/* program that breakpoint is in. */
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
		trig dbref					/* triggering object */
		started long				/* When this program started. */
		instcnt						/* How many instructions have run. */
		pid int						/* what is the process id? */
		errorstr string				/* the error string thrown */
		errorinst string			/* the instruction name that threw an error */
		errorprog dbref				/* the program that threw an error */
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
	*Object
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

type Player struct {
	*Object
	home dbref
	curr_prog dbref
	insert_mode bool
	block bool
	password string
	descrs []int
	ignore_cache []dbref
	ignore_last dbref
}

type Room struct {
	*Object
	dbref
}

type Exit struct {
	*Object
	dest []dbref
}

	struct macrotable {
		name string
		definition string
		implementor dbref
		left *macrotable
		right *macrotable
	};

	#define PLAYER_HASH_SIZE   (1024)	/* Table for player lookups */
	#define COMP_HASH_SIZE     (256)	/* Table for compiler keywords */
	#define DEFHASHSIZE        (256)	/* Table for compiler $defines */

	/*
	  Usage guidelines:

	  To obtain an object pointer use db.Fetch(i).  Pointers returned by db.Fetch
	  may become invalid after a call to new_object().

	  If you have updated an object set OBJECT_CHANGED flag before leaving the routine that did the update.

	  Some fields are now handled in a unique way, since they are always memory
	  resident, even in the GDBM_DATABASE disk-based muck.  These are: name,
	  flags and owner.  Refer to these by db.Fetch(i).name, db.Fetch(i).flags and db.Fetch(i).Owner.

	  The programmer is responsible for managing storage for string
	  components of entries; db_read will produce malloc'd strings. Note that db_read will
	  attempt to free any non-NULL string that exists in db when it is invoked.
	*/

type DB map[dbref] interface{}

func (db DB) Fetch(x dbref) interface{} {
	return db[x]
}

func (db DB) Store(x dbref, v interface{}) {
	DB[x] = v
}

func (db DB) FetchPlayer(x dbref) *Player {
	return db.Fetch(x).(Player)
}


var db = make(DB)
var db_top dbref
var db_load_format int

struct macrotable *macrotop;

func getparent_logic(obj dbref) dbref {
	if obj == NOTHING {
		return NOTHING
	}
	if TYPEOF(obj) == TYPE_THING && db.Fetch(obj).flags & VEHICLE != 0 {
		obj = db.Fetch(obj).(Player).home
		if obj != NOTHING && TYPEOF(obj) == TYPE_PLAYER {
			obj = db.Fetch(obj).(Player).home
		}
	} else {
		obj = db.Fetch(obj).Location
	}
	return obj
}

func getparent(obj dbref) dbref {
	var ptr, oldptr dbref
	if tp_thing_movement {
		obj = db.Fetch(obj).Location
	} else {
		ptr = getparent_logic(obj)
		do {
			obj = getparent_logic(obj)
		} while obj != (oldptr = ptr = getparent_logic(ptr)) && obj != (ptr = getparent_logic(ptr)) && obj != NOTHING && TYPEOF(obj) == TYPE_THING
		if obj != NOTHING && (obj == oldptr || obj == ptr) {
			obj = GLOBAL_ENVIRONMENT
		}
	}
	return obj
}

func db_grow(newtop dbref) {
	var newdb *Object
	if newtop > db_top {
		db_top = newtop
		if db != nil {
			if ((newdb = (struct object *)
				 realloc((void *) db, db_top * sizeof(struct object))) == 0) {
				abort();
			}
			db = newdb;
		} else {
			/* make the initial one */
			int startsize = (newtop >= 100) ? newtop : 100;

			if ((db = (struct object *)
				 malloc(startsize * sizeof(struct object))) == 0) {
				abort();
			}
		}
	}
}

func db_clear_object(dbref i) {
	o := db.Fetch(i)
	o.name = ""
	o.TimeStamps = nil
	o.Location = NOTHING
	o.Contents = NOTHING
	o.Exits = NOTHING
	o.next = NOTHING
	o.properties = 0
	/* db.Fetch(i).flags |= OBJECT_CHANGED */
	/* flags you must initialize yourself */
	/* type-specific fields you must also initialize */
}

func new_object() (r dbref) {
	r = db_top
	db_grow(db_top + 1)
	db_clear_object(r)
	db.Fetch(r).flags |= OBJECT_CHANGED
	return
}

func putref(f *FILE, ref dbref) {
	if fprintf(f, "%d\n", ref) < 0 {
		abort()
	}
}

func putstring(f *FILE, s string) {
	if s != "" {
		if fputs(s, f) == EOF {
			abort()
		}
	}
	if putc('\n', f) == EOF {
		abort()
	}
}

func putproperties_rec(f *FILE, dir string, obj dbref) {
	val pptr *Plist
	char name[BUFFER_LEN]

	_, pref := pptr.first_prop_nofetch(obj, dir, name)
	for pref != nil {
		p := pref;
		p.db_putprop(f, dir)
		buf := dir + name
		if p.dir != nil {
			buf += "/"
			putproperties_rec(f, buf, obj)
		}
		pref, name = pref.next_prop(pptr)
	}
}

func putproperties(f *FILE, obj dbref) {
	putstring(f, "*Props*");
	db_dump_props(f, obj);
	/* putproperties_rec(f, "/", obj); */
	putstring(f, "*End*");
}

extern FILE *input_file;
extern FILE *delta_infile;
extern FILE *delta_outfile;

func macrodump(node *macrotable, f *FILE) {
	if node != nil {
		macrodump(node.left, f)
		putstring(f, node.name)
		putstring(f, node.definition)
		putref(f, node.implementor)
		macrodump(node.right, f)
	}
}

char *
file_line(FILE * f)
{
	char buf[BUFFER_LEN];
	int len;

	if (!fgets(buf, BUFFER_LEN, f))
		return NULL;
	len = len(buf);
	if (buf[len-1] == '\n') {
		buf[--len] = '\0';
	}
	if (buf[len-1] == '\r') {
		buf[--len] = '\0';
	}
	return buf
}

void
foldtree(struct macrotable *center)
{
	int count = 0;
	struct macrotable *nextcent = center;

	for (; nextcent; nextcent = nextcent->left)
		count++;
	if (count > 1) {
		for (nextcent = center, count /= 2; count--; nextcent = nextcent->left) ;
		if (center->left)
			center->left->right = NULL;
		center->left = nextcent;
		foldtree(center->left);
	}
	for (count = 0, nextcent = center; nextcent; nextcent = nextcent->right)
		count++;
	if (count > 1) {
		for (nextcent = center, count /= 2; count--; nextcent = nextcent->right) ;
		if (center->right)
			center->right->left = NULL;
		foldtree(center->right);
	}
}

int
macrochain(struct macrotable *lastnode, FILE * f)
{
	char *line, *line2;
	struct macrotable *newmacro;

	if (!(line = file_line(f)))
		return 0;
	line2 = file_line(f);

	newmacro = (struct macrotable *) new_macro(line, line2, getref(f));
	free(line);
	free(line2);

	if (!macrotop)
		macrotop = (struct macrotable *) newmacro;
	else {
		newmacro->left = lastnode;
		lastnode->right = newmacro;
	}
	return (1 + macrochain(newmacro, f));
}

void
macroload(FILE * f)
{
	int count = 0;

	macrotop = NULL;
	count = macrochain(macrotop, f);
	for (count /= 2; count--; macrotop = macrotop->right) ;
	foldtree(macrotop);
	return;
}

func log_program_text(first *line, player, i dbref) {
	var f *FILE
	lt := time(NULL)

	fname := PROGRAM_LOG
	f = fopen(fname, "ab");
	if (!f) {
		log_status("Couldn't open file %s!", fname)
		return;
	}

	fputs("#######################################", f);
	fputs("#######################################\n", f);
	fprintf(f, "PROGRAM %s, SAVED AT %s BY %s(%d)\n", unparse_object(player, i), ctime(&lt), db.Fetch(player).name, player)
	fputs("#######################################", f);
	fputs("#######################################\n\n", f);

	for first != nil; first = first.next {
		if first.this_line {
			fputs(first.this_line, f)
			fputc('\n', f)
		}
	}
	fputs("\n\n\n", f)
	fclose(f)
}

func write_program(struct line *first, dbref i) {
	FILE *f;

	fname := fmt.Sprintf("muf/%d.m", (int) i);
	f = fopen(fname, "wb");
	if (!f) {
		log_status("Couldn't open file %s!", fname);
		return;
	}
	while (first) {
		if (!first->this_line)
			continue;
		if (fputs(first->this_line, f) == EOF) {
			abort();
		}
		if (fputc('\n', f) == EOF) {
			abort();
		}
		first = first->next;
	}
	fclose(f);
}

func db_write_object(f *FILE, i dbref) {
	o := db.Fetch(i)
	putstring(f, o.name)
	putref(f, o.Location)
	putref(f, o.Contents)
	putref(f, o.next)
	putref(f, o.flags & ~DUMP_MASK)		/* write non-internal flags */
	putref(f, o.Created)
	putref(f, o.LastUsed)
	putref(f, o.Uses)
	putref(f, o.Modified)
	putproperties(f, i)

	switch {
	case IsThing(i):
		putref(f, o.(Player).home)
		putref(f, o.Exits)
		putref(f, o.Owner)
	case IsRoom(i):
		putref(f, o.(dbref))
		putref(f, o.Exits)
		putref(f, o.Owner)
	case IsExit(i):
		putref(f, len(o.(Exit).Destinations))
		for _, v := range o.(Exit).Destinations {
			putref(f, v)
		}
		putref(f, o.Owner)
	case IsPlayer(i):
		putref(f, o.(Player).home)
		putref(f, o.Exits)
		putstring(f, o.(Player).password)
	case IsProgram(i):
		putref(f, o.Owner)
	}
}

int deltas_count = 0;

#ifndef CLUMP_LOAD_SIZE
#define CLUMP_LOAD_SIZE 20
#endif

/* mode == 1 for dumping all objects.  mode == 0 for deltas only.  */

func db_write_list(f *FILE, mode int) {
	EachObjectInReverse(func(obj dbref, o *Object) {
		if mode == 1 || o.flags & OBJECT_CHANGED != 0 {
			if fprintf(f, "#%d\n", i) < 0 {
				abort()
			}
			db_write_object(f, obj)
			o.flags &= ~OBJECT_CHANGED	/* clear changed flag */
		}
	})
}

func db_write(f *FILE) dbref {
	putstring(f, DB_VERSION_STRING)
	putref(f, db_top)
	putref(f, DB_PARMSINFO)
	putref(f, tune_count_parms())
	tune_save_parms_to_file(f)
	db_write_list(f, 1)
	fseek(f, 0L, 2)
	putstring(f, "***END OF DUMP***")
	fflush(f)
	deltas_count = 0
	return db_top
}

func db_write_deltas(f *FILE) dbref {
	fseek(f, 0L, 2)		/* seek end of file */
	putstring(f, "***Foxen8 Deltas Dump Extention***")
	db_write_list(f, 0)
	fseek(f, 0L, 2)
	putstring(f, "***END OF DUMP***")
	fflush(f)
	return db_top
}

func parse_dbref(s string) (r dbref) {
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

func getproperties(FILE * f, dbref obj, const char *pdir) {
	char buf[BUFFER_LEN * 3], *p;
	int datalen;

	/* get rid of first line */
	fgets(buf, sizeof(buf), f);

	if buf != "Props*\n" {
		/* initialize first line stuff */
		fgets(buf, sizeof(buf), f);
		for {
			/* fgets reads in \n too! */
			if buf == "***Property list end ***\n" || buf == "*End*\n" {
				break
			}
			p = strchr(buf, PROP_DELIMITER);
			*(p++) = '\0';		/* Purrrrrrrrrr... */
			datalen = len(p);
			p[datalen - 1] = '\0';

			if ((*p == '^') && (unicode.IsNumber(p + 1))) {
				add_prop_nofetch(obj, buf, NULL, atol(p + 1))
			} else {
				if (*buf) {
					add_prop_nofetch(obj, buf, p, 0)
				}
			}
			fgets(buf, sizeof(buf), f);
		}
	} else {
		db_getprops(f, obj, pdir);
	}
}

func db_free() {
	db = nil
	db_top = 0
	player_list = make(map[string] dbref)
	primitive_list = make(map[string] PROG_PRIMITIVE)
}

func read_program(i dbref) *line {
	char buf[BUFFER_LEN];
	first *line
	prev *line = NULL
	FILE *f;
	int len;

	first = NULL;
	buf = fmt.Sprintf("muf/%d.m", (int) i);
	f = fopen(buf, "rb");
	if (!f)
		return 0;

	while (fgets(buf, BUFFER_LEN, f)) {
		nu := new(line)
		len = len(buf);
		if (len > 0 && buf[len - 1] == '\n') {
			buf[len - 1] = '\0';
			len--;
		}
		if (len > 0 && buf[len - 1] == '\r') {
			buf[len - 1] = '\0';
			len--;
		}
		if (!*buf)
			strcpyn(buf, sizeof(buf), " ");
		nu->this_line = buf
		if (!first) {
			prev = nu;
			first = nu;
		} else {
			prev->next = nu;
			nu->prev = prev;
			prev = nu;
		}
	}

	fclose(f);
	return first;
}

#define getstring_oldcomp_noalloc(foo) getstring_noalloc(foo)

func db_read_object_old(f *FILE, o *Object, objno dbref) {
	db_clear_object(objno)
	db.Fetch(objno).flags = 0
	db.Fetch(objno).name = getstring(f)
	add_prop_nofetch(objno, MESGPROP_DESC, getstring_oldcomp_noalloc(f), 0)
	db.Fetch(objno).flags |= OBJECT_CHANGED
	o.Location = getref(f)
	o.Contents = getref(f)
	exits := getref(f)
	o.next = getref(f)
	set_property_nofetch(objno, MESGPROP_LOCK, getboolexp(f))
	db.Fetch(objno).flags |= OBJECT_CHANGED

	add_prop_nofetch(objno, MESGPROP_FAIL, getstring_oldcomp_noalloc(f), 0)
	db.Fetch(objno).flags |= OBJECT_CHANGED
	add_prop_nofetch(objno, MESGPROP_SUCC, getstring_oldcomp_noalloc(f), 0)
	db.Fetch(objno).flags |= OBJECT_CHANGED
	add_prop_nofetch(objno, MESGPROP_OFAIL, getstring_oldcomp_noalloc(f), 0)
	db.Fetch(objno).flags |= OBJECT_CHANGED
	add_prop_nofetch(objno, MESGPROP_OSUCC, getstring_oldcomp_noalloc(f), 0)
	db.Fetch(objno).flags |= OBJECT_CHANGED


	db.Fetch(objno).Owner = getref(f)
	pennies := getref(f)

	/* timestamps mods */
	o.Created = time(NULL)
	o.LastUsed = time(NULL)
	o.Uses = 0
	o.Modified = time(NULL)

	db.Fetch(objno).flags |= getref(f)
	/*
	 * flags have to be checked for conflict --- if they happen to coincide
	 * with chown_ok flags and jump_ok flags, we bump them up to the
	 * corresponding HAVEN and ABODE flags
	 */
	if db.Fetch(objno).flags & CHOWN_OK != 0 {
		db.Fetch(objno).flags &= ~CHOWN_OK
		db.Fetch(objno).flags |= HAVEN
	}
	if db.Fetch(objno).flags & JUMP_OK != 0 {
		db.Fetch(objno).flags &= ~JUMP_OK
		db.Fetch(objno).flags |= ABODE
	}
	password := getstring(f)
	switch db.Fetch(objno).flags & TYPE_MASK != 0 {
	case TYPE_THING:
		db.Fetch(objno).(Player) = new(Player)
		db.Fetch(objno).(Player).home = exits
		add_prop_nofetch(objno, MESGPROP_VALUE, "", pennies)
		o.Exits = NOTHING
	case TYPE_ROOM:
		o.sp = o.Location
		o.Location = NOTHING
		o.Exits = exits
	case TYPE_EXIT:
		if o.Location == NOTHING {
			o.(Exit).Destinations = nil
		} else {
			o.(Exit).Destinations = []dbref{ o.Location }
		}
		o.Location = NOTHING
	case TYPE_PLAYER:
		db.Fetch(objno).(Player) = &Player{ home: exits, curr_prog: NOTHING, ignore_last: NOTHING }
		o.Exits = NOTHING
		add_prop_nofetch(objno, MESGPROP_VALUE, "", pennies)
		set_password_raw(objno, "")
		set_password(objno, password);
	}
}

func db_read_object_new(f *FILE, o *Object, objno dbref) {
	int j;
	const char *password;

	db_clear_object(objno);
	db.Fetch(objno).flags = 0
	db.Fetch(objno).name = getstring(f)
	add_prop_nofetch(objno, MESGPROP_DESC, getstring_noalloc, 0)
	db.Fetch(objno).flags |= OBJECT_CHANGED

	o->location = getref(f);
	o->contents = getref(f);
	/* o->exits = getref(f); */
	o->next = getref(f);
	set_property_nofetch(objno, MESGPROP_LOCK, getboolexp(f))
	db.Fetch(objno).flags |= OBJECT_CHANGED

	add_prop_nofetch(objno, MESGPROP_FAIL, getstring_oldcomp_noalloc(f), 0)
	db.Fetch(objno).flags |= OBJECT_CHANGED
	add_prop_nofetch(objno, MESGPROP_SUCC, getstring_oldcomp_noalloc(f), 0)
	db.Fetch(objno).flags |= OBJECT_CHANGED
	add_prop_nofetch(objno, MESGPROP_OFAIL, getstring_oldcomp_noalloc(f), 0)
	db.Fetch(objno).flags |= OBJECT_CHANGED
	add_prop_nofetch(objno, MESGPROP_OSUCC, getstring_oldcomp_noalloc(f), 0)
	db.Fetch(objno).flags |= OBJECT_CHANGED

	/* timestamps mods */
	o.Created = time(NULL)
	o.LastUsed = time(NULL)
	o.Uses = 0;
	o.Modified = time(NULL)

	db.Fetch(objno).flags |= getref(f)

	/*
	 * flags have to be checked for conflict --- if they happen to coincide
	 * with chown_ok flags and jump_ok flags, we bump them up to the
	 * corresponding HAVEN and ABODE flags
	 */
	if db.Fetch(objno).flags & CHOWN_OK != 0 {
		db.Fetch(objno).flags &= ~CHOWN_OK;
		db.Fetch(objno).flags |= HAVEN;
	}
	if db.Fetch(objno).flags & JUMP_OK != 0 {
		db.Fetch(objno).flags &= ~JUMP_OK;
		db.Fetch(objno).flags |= ABODE;
	}
	/* o->password = getstring(f); */
	switch db.Fetch(objno).flags & TYPE_MASK {
	case TYPE_THING:
		db.Fetch(objno).(Player) = new(Player)
		db.Fetch(objno).(Player).home = getref(f)
		o.Exits = getref(f)
		db.Fetch(objno).Owner = getref(f)
		add_prop_nofetch(objno, MESGPROP_VALUE, "", getref(f))
	case TYPE_ROOM:
		o.sp = getref(f)
		o.Exits = getref(f)
		db.Fetch(objno).Owner = getref(f)
	case TYPE_EXIT:
		o.(Exit).Destinations = make([]dbref, getref(f))
		for i, _ := range o.(Exit).Destinations {
			o.(Exit).Destinations[i] = getref(f)
		}
		db.Fetch(objno).Owner = getref(f)
	case TYPE_PLAYER:
		db.Fetch(objno).(Player) = &Player{ home: getref(f), curr_prog: NOTHING, ignore_last: NOTHING }
		o.Exits = getref(f)
		add_prop_nofetch(objno, MESGPROP_VALUE, "", getref(f))
		password = getstring(f)
		set_password_raw(objno, "")
		set_password(objno, password)
	}
}

/* Reads in Foxen, Foxen[2-8], WhiteFire, Mage or Lachesis DB Formats */
func db_read_object_foxen(f *FILE, o *Object, objno dbref, dtype int, read_before bool) {
	int c, prop_flag = 0;

	if read_before {
		*(db.Fetch(objno)) = nil
	}
	db_clear_object(objno)

	db.Fetch(objno).flags = 0
	db.Fetch(objno).name = getstring(f)
	if dtype <= 3 {
		add_prop_nofetch(objno, MESGPROP_DESC, getstring_oldcomp_noalloc(f), 0)
		db.Fetch(objno).flags |= OBJECT_CHANGED
	}
	o.Location = getref(f)
	o.Contents = getref(f)
	o.next = getref(f)
	if dtype < 6 {
		set_property_nofetch(objno, MESGPROP_LOCK, getboolexp(f))
		db.Fetch(objno).flags |= OBJECT_CHANGED
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
		add_prop_nofetch(objno, MESGPROP_FAIL, getstring_oldcomp_noalloc(f), 0)
		db.Fetch(objno).flags |= OBJECT_CHANGED

		add_prop_nofetch(objno, MESGPROP_SUCC, y, 0)
		db.Fetch(objno).flags |= OBJECT_CHANGED

		add_prop_nofetch(objno, MESGPROP_DROP, getstring_oldcomp_noalloc(f), 0)
		db.Fetch(objno).flags |= OBJECT_CHANGED

		add_prop_nofetch(objno, MESGPROP_OFAIL, getstring_oldcomp_noalloc(f), 0)
		db.Fetch(objno).flags |= OBJECT_CHANGED

		add_prop_nofetch(objno, MESGPROP_OSUCC, getstring_oldcomp_noalloc(f), 0)
		db.Fetch(objno).flags |= OBJECT_CHANGED

		add_prop_nofetch(objno, MESGPROP_ODROP, getstring_oldcomp_noalloc(f), 0)
		db.Fetch(objno).flags |= OBJECT_CHANGED
	}
	tmp := getref(f)			/* flags list */
	if dtype >= 4 {
		tmp &= ~DUMP_MASK
	}
	db.Fetch(objno).flags |= tmp
	db.Fetch(objno).flags &= ~SAVED_DELTA
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

	switch db.Fetch(objno).flags & TYPE_MASK != 0 {
	case TYPE_THING:
		db.Fetch(objno).(Player) = new(Player)
		var home dbref
		if prop_flag {
			home = getref(f)
		} else {
			home = j
		}
		db.Fetch(objno).(Player).home = home
		o.Exits = getref(f)
		db.Fetch(objno).Owner = getref(f)
		if dtype < 10 {
			add_prop_nofetch(objno, MESGPROP_VALUE, "", getref(f))
		}
	case TYPE_ROOM:
		if prop_flag {
			o.sp = getref(f)
		} else {
			o.sp = j
		}
		o.Exits = getref(f)
		db.Fetch(objno).Owner = getref(f)
	case TYPE_EXIT:
		if prop_flag {
			o.(Exit).Destinations = make([]dbref, getref(f))
		} else {
			o.(Exit).Destinations = make([]dbref, j)
		}
		for i, _ := range o.(Exit).Destinations {
			o.(Exit).Destinations[i] = getref(f)
		}
		db.Fetch(objno).Owner = getref(f)
	case TYPE_PLAYER:
		if prop_flag {
			db.Fetch(objno).(Player) = &Player{ home: getref(f), curr_prog: NOTHING, ignore_last: NOTHING }
		} else {
			db.Fetch(objno).(Player) = &Player{ home: j, curr_prog: NOTHING, ignore_last: NOTHING }
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
	case TYPE_PROGRAM:
		db.Fetch(objno).(Program) = new(Program)
		db.Fetch(objno).Owner = getref(f)
		db.Fetch(objno).flags &= ~INTERNAL

		if dtype < 8 && db.Fetch(objno).flags & LINK_OK != 0 {
			/* set Viewable flag on Link_ok programs. */
			db.Fetch(objno).flags |= VEHICLE
		}
		if dtype < 5 && MLevel(objno) == NON_MUCKER {
			SetMLevel(objno, JOURNEYMAN)
		}
	}
}

func autostart_progs() {
	if !db_conversion_flag {
		EachObject(func(obj dbref, o *Object) {
			if IsProgram(i) {
				if o.flags & ABODE != 0 && TrueWizard(o.Owner) {
					/* pre-compile AUTOSTART programs. */
					/* They queue up when they finish compiling. */
					/* Uncomment when db.Fetch "does" something. */
					tmp := o.(Program).first
					o.(Program).first = read_program(i)
					do_compile(-1, o.Owner, i, 0)
					o.(Program).first = tmp
				}
			}
		})
	}
}

func db_read(FILE * f) dbref {
	dbref grow, thisref;
	o *Object
	const char *version;
	int doing_deltas;
	int main_db_format = 0;
	int parmcnt;
	char c;

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
		tune_load_parms_from_file(f, NOTHING, parmcnt)
	}

	/* grow the db up front */
	if dbflags & DB_ID_GROW != 0 {
		db_grow(grow)
	}

	doing_deltas = dbflags & DB_ID_DELTAS
	if doing_deltas {
		if db == nil {
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

			/* read it in */
			o = db.Fetch(thisref)
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
			if IsPlayer(thisref) {
				db.Fetch(thisref).Owner = thisref
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
						tune_load_parms_from_file(f, NOTHING, parmcnt)
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