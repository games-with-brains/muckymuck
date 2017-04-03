fbmuck

//	Property struct
type Plist struct {
	data interface{}
	height int8			//	satisfy the avl monster
	left, right, dir *Plist
	key [1]byte
}

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
#define Prop_ReadOnly(name) \
    (Prop_Check(name, PROP_RDONLY) || Prop_Check(name, PROP_RDONLY2))
#define Prop_Private(name) Prop_Check(name, PROP_PRIVATE)
#define Prop_SeeOnly(name) Prop_Check(name, PROP_SEEONLY)
#define Prop_Hidden(name) Prop_Check(name, PROP_HIDDEN)
#define Prop_System(name) is_prop_prefix(name, "@__sys__")

func (avl *Plist) find(key string) (r *Plist) {
	for avl != nil {
		switch cmpval := strings.Compare(key, avl.key); {
		case cmpval > 0:
			avl = avl.right
		case cmpval < 0:
			avl = avl.left
		default:
			break
		}
	}
	return avl
}

func (node *Plist) height_of() (r int) {
	if node != nil {
		r = node.height
	}
	return
}

func (node *Plist) height_diff() (r int) {
	if node != nil {
		r = node.right.height_of() - node.left.height_of()
	}
	return
}

func max(a, b int) (r int) {
	r = a
	if b > r {
		r = b
	}
	return
}

func (node *Plist) fixup_height() {
	if node != nil {
		node.height = 1 + max(node.left.height_of(), node.right.height_of())
	}
}

func (a *Plist) rotate_left_single() (r *Plist) {
	r = a.right
	a.right = r.left
	r.left = a

	a.fixup_height()
	r.fixup_height()
	return
}

func (a *Plist) rotate_left_double() (r *Plist) {
	b := a.right
	r = b.left
	a.right = r.left
	b.left = r.right
	r.left = a
	r.right = b

	a.fixup_height()
	b.fixup_height()
	r.fixup_height()
	return
}

func (a *Plist) rotate_right_single() (r *Plist) {
	r = a.left
	a.left = r.right
	r.right = a
	a.fixup_height()
	r.fixup_height()
	return
}

func (a *Plist) rotate_right_double() (r *Plist) {
	b := a.left
	r == b.right

	a.left = r.right
	b.right = r.left
	r.right = a
	r.left = b

	a.fixup_height()
	b.fixup_height()
	r.fixup_height()
	return c
}

func (a *Plist) balance_node() *Plist {
	dh := a.height_diff()
	if abs(dh) < 2 {
		a.fixup_height()
	} else {
		switch {
		case dh == 2:
			if a.right.height_diff() >= 0 {
				a = a.rotate_left_single()
			} else {
				a = a.rotate_left_double()
			}
		case a.left.height_diff() <= 0:
			a = a.rotate_right_single()
		} else {
			a = a.rotate_right_double()
		}
	}
	return a
}

func NewPropNode(name string) (r *Plist) {
	r = &Plist{ height: 1, key: name }
	SetPFlagsRaw(r, PROP_DIRTYP)
	return
}

func (p *Plist) clear_propnode() {
	p.data = nil
	SetPType(p, PROP_DIRTYP)
}

var insert_balancep bool

func (avl *Plist) insert(key string) (r *Plist) {
	if p := avl; p != nil {
		switch cmp := strings.Compare(key, p.key); {
		case cmp > 0:
			r = p.right.insert(key)
		case cmp < 0:
			r = p.left.insert(key)
		default:
			insert_balancep = false
			r = p
		}
		if insert_balancep {
			*avl = p.balance_node()
		}
	} else {
		*avl = allNewPropNodey)
		insert_balancep = true
		r = p
	}
	return
}

func (avl *Plist) getmax() *Plist {
	if avl != nil && avl.right != nil {
		return avl.right.getmax()
	}
	return avl
}

func (root *Plist) remove_propnode(key string) (save *Plist) {
	avl := *root
	save = avl
	if avl != nil {
		switch cmpval := strings.Compare(key, avl.key); {
		case cmpval < 0:
			save = avl.left.remove_propnode(key)
		case cmpval > 0:
			save = avl.right.remove_propnode(key)
		case avl.left == nil:
			avl = avl.right
		case avl.right == nil:
			avl = avl.left
		default:
			tmp := avl.left.remove_propnode(avl.left.getmax().key)
			if tmp == nil {
				//	this shouldn't be possible.
				panic("remove_propnode() returned nil!")
			}
			tmp.left = avl.left
			tmp.right = avl.right
			avl = tmp
		}
		if save != nil {
			save.left = nil
			save.right = nil
		}
		*root = avl.balance_node()
	}
	return save
}

func (avl *Plist) delnode(key string) *Plist {
	avl.remove_propnode(key)
	return avl
}

func (list *Plist) locate_prop(name string) *Plist {
	return list.find(name)
}

//	if *list is NULL, create a new propdir, then insert the prop
func (list *Plist) new_prop(name string) *Plist {
	return list.insert(name)
}

//	when last prop in dir is deleted, destroy propdir & change *list to nil
func (list *Plist) delete_prop(name string) *Plist {
	*list = *list.delnode(name)
	return *list
}

func (list *Plist) first_node() (r *Plist) {
	for r = list; r.left != nil; r = r.left {}
	return
}


func (ptr *Plist) next_node(name string) (r *Plist) {
	if ptr != nil && len(name) > 0 {
		switch cmpval := strings.Compare(name, ptr.key); {
		case cmpval < 0:
			r = ptr.left.next_node(name)
			if r == nil {
				r = ptr
			}
		case cmpval > 0:
			r = ptr.right.next_node(name)
		case ptr.right != nil:
			r = ptr.right
			for r.left != nil {
				r = from.left
			}
		}
	}
	return
}

//	copies properties
func (old *Plist) copy_proplist(obj dbref) (nu *Plist) {
	if old != nil {
		p := nu.new_prop(old.key)
		SetPFlagsRaw(p, PropFlagsRaw(old))
		switch v := old.data.(type) {
		case Lock:			// FIXME: lock
			p.data = copy_bool(v)
		case PROP_DIRTYP:
			p.data = 0
		default:
			p.data = old.data
		}
		p.dir = old.dir.copy_proplist(obj)
		p.left = old.left.copy_proplist(obj)
		p.right = old.right.copy_proplist(obj)
	}
}

func Prop_Check(name, what string) (r bool) {
	if r = name == what; !r {
		for ((name = strchr(name, PROPDIR_DELIMITER))) {
			if name[1] == what {
				r = true
				break
			}
			name = name[1:]
		}
	}
	return false
}