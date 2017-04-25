package fbmuck

/* max length of command argument to process_command */
#define MAX_COMMAND_LEN 2048
#define BUFFER_LEN ((MAX_COMMAND_LEN)*4)
#define FILE_BUFSIZ ((BUFSIZ)*8)

typedef int ObjectID;				/* offset into db */

#define TIME_INFINITE ((sizeof(time_t) == 4)? 0xefffffff : 0xefffffffffffffff)

#define DB_READLOCK(x)
#define DB_WRITELOCK(x)
#define DB_RELEASE(x)

/* defines for possible data access mods. */
#define MESGPROP_DESC		"_/de"
#define MESGPROP_IDESC		"_/ide"
#define MESGPROP_SUCC		"_/sc"
#define MESGPROP_OSUCC		"_/osc"
#define MESGPROP_FAIL		"_/fl"
#define MESGPROP_OFAIL		"_/ofl"
#define MESGPROP_DROP		"_/dr"
#define MESGPROP_ODROP		"_/odr"
#define MESGPROP_DOING		"_/do"
#define MESGPROP_OECHO		"_/oecho"
#define MESGPROP_PECHO		"_/pecho"
#define MESGPROP_LOCK		"_/lok"
#define MESGPROP_FLOCK		"@/flk"
#define MESGPROP_CONLOCK	"_/clk"
#define MESGPROP_CHLOCK		"_/chlk"
#define MESGPROP_VALUE		"@/value"
#define MESGPROP_GUEST		"@/isguest"

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


#define GOD ((ObjectID) 1)

#define PREEMPT 0
#define FOREGROUND 1
#define BACKGROUND 2

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

#define MAX_VAR         54		/* maximum number of variables including the
								   * basic ME, LOC, TRIGGER, and COMMAND vars */
#define RES_VAR          4		/* no of reserved variables */

#define STACK_SIZE       1024	/* maximum size of stack */

struct stack_addr {				/* for the system callstack */
	ObjectID progref;				/* program call was made from */
	struct inst *offset;		/* the address of the call */
};

typedef struct inst vars[MAX_VAR];

struct forvars {
	int didfirst;
	struct inst cur;
	struct inst end;
	int step;
	struct forvars *next;
};

struct tryvars {
	int depth;
	int call_level;
	int for_count;
	struct inst *addr;
	struct tryvars *next;
};

struct stack {
	int top;
	struct inst st[STACK_SIZE];
};

struct sysstack {
	int top;
	struct stack_addr st[STACK_SIZE];
};

struct callstack {
	int top;
	ObjectID st[STACK_SIZE];
};

struct localvars {
	struct localvars *next;
	struct localvars **prev;
	ObjectID prog;
	vars lvars;
};

struct forstack {
	int top;
	struct forvars *st;
};

struct trystack {
	int top;
	struct tryvars *st;
};

#define MAX_BREAKS 16
struct debuggerdata {
	unsigned debugging:1;		/* if set, this frame is being debugged */
	unsigned force_debugging:1;	/* if set, debugger is active, even if not set Z */
	unsigned bypass:1;			/* if set, bypass breakpoint on starting instr */
	unsigned isread:1;			/* if set, the prog is trying to do a read */
	unsigned showstack:1;		/* if set, show stack debug line, each inst. */
	unsigned dosyspop:1;		/* if set, fix up system stack before returning. */
	int lastlisted;				/* last listed line */
	char *lastcmd;				/* last executed debugger command */
	short breaknum;				/* the breakpoint that was just caught on */

	ObjectID lastproglisted;		/* What program's text was last loaded to list? */
	struct line *proglines;		/* The actual program text last loaded to list. */

	short count;				/* how many breakpoints are currently set */
	short temp[MAX_BREAKS];		/* is this a temp breakpoint? */
	short level[MAX_BREAKS];	/* level breakpnts.  If -1, no check. */
	struct inst *lastpc;		/* Last inst interped.  For inst changes. */
	struct inst *pc[MAX_BREAKS];	/* pc breakpoint.  If null, no check. */
	int pccount[MAX_BREAKS];	/* how many insts to interp.  -2 for inf. */
	int lastline;				/* Last line interped.  For line changes. */
	int line[MAX_BREAKS];		/* line breakpts.  -1 no check. */
	int linecount[MAX_BREAKS];	/* how many lines to interp.  -2 for inf. */
	ObjectID prog[MAX_BREAKS];		/* program that breakpoint is in. */
};

#define dequeue_prog(x,i) dequeue_prog_real(x,i,__FILE__,__LINE__)

#define STD_REGUID 0
#define STD_SETUID 1
#define STD_HARDUID 2

#define PLAYER_HASH_SIZE   (1024)	/* Table for player lookups */
#define COMP_HASH_SIZE     (256)	/* Table for compiler keywords */
#define DEFHASHSIZE        (256)	/* Table for compiler $defines */

/*
  Usage guidelines:

  To obtain an object pointer use DB.Fetch(i).  Pointers returned by DB.Fetch
  may become invalid after a call to new_object().

  If you have updated an object set the TimeStamps.Changed flag before leaving the routine that did the update.

  Some fields are now handled in a unique way, since they are always memory
  resident, even in the GDBM_DATABASE disk-based muck.  These are: name,
  flags and owner.  Refer to these by DB.Fetch(i).name, DB.Fetch(i).flags and DB.Fetch(i).Owner.

  The programmer is responsible for managing storage for string
  components of entries; db_read will produce malloc'd strings.  Note that db_free and db_read will
  attempt to free any non-NULL string that exists in db when they are invoked.
*/
#endif							/* __DB_H */