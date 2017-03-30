fbmuck

//	Property struct
type Plist struct {
	flags byte
	height int8		//	satisfy the avl monster.
	data interface{}
	left, right, dir *Plist
	key [1]byte
};

/* property value types */
#define PROP_DIRTYP   0
#define PROP_STRTYP   2
#define PROP_INTTYP   4
#define PROP_LOKTYP   8
#define PROP_REFTYP   16
#define PROP_FLTTYP   32
#define PROP_TYPMASK  128

/* Property flags.  Unimplemented as yet. */
#define PROP_UREAD       0x0010
#define PROP_UWRITE      0x0020
#define PROP_WREAD       0x0040
#define PROP_WWRITE      0x0080

/* half implemented.  Will be used for stuff like password props. */
#define PROP_SYSPERMS    0x0100

/* Blessed props evaluate with wizbit MPI perms. */
#define PROP_BLESSED     0x1000


/* Macros */
#define SetPFlags(x,y) {(x)->flags = ((x)->flags & PROP_TYPMASK) | (short)y;}
#define PropFlags(x) ((x)->flags & ~PROP_TYPMASK)

#define SetPType(x,y) {(x)->flags = ((x)->flags & ~PROP_TYPMASK) | (short)y;}

#define SetPFlagsRaw(x,y) {(x)->flags = (short)y;}
#define PropFlagsRaw(x) ((x)->flags)

#define Prop_Blessed(obj,propname) (get_property_flags(obj, propname) & PROP_BLESSED)

/* property access macros */
#define Prop_ReadOnly(name) Prop_Check(name, PROP_RDONLY) || Prop_Check(name, PROP_RDONLY2)
#define Prop_Private(name) Prop_Check(name, PROP_PRIVATE)
#define Prop_SeeOnly(name) Prop_Check(name, PROP_SEEONLY)
#define Prop_Hidden(name) Prop_Check(name, PROP_HIDDEN)
#define Prop_System(name) is_prop_prefix(name, "@__sys__")