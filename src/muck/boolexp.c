package fbmuck

type Lock interface {
	IsTrue() bool
	Unparse(dbref, bool) string
}

func IsTrue(v interface{}) (r bool) {
	switch v := v.(type) {
	case Lock:
		r = v.IsTrue()
	case bool:
		r = v
	}
	return
}

func IsFalse(v interface{}) bool {
	return !IsTrue(v)
}


type Unlocked struct {}
const UNLOCKED = Unlocked{}

func (u Unlocked) IsTrue() bool {
	return true
}

func (u Unlocked) Unparse(player dbref, fullname bool) string {
	return "*UNLOCKED*"
}

func (u Unlocked) Eval(descr int, player, thing dbref) (r bool) {
	return true
}


type Locked struct {}
const LOCKED = Locked{}

func (l Locked) IsTrue() bool {
	return false
}

func (l Locked) Unparse(player dbref, fullname bool) string {
	return "*LOCKED*"
}

func (l Locked) Eval(descr int, player, thing dbref) (r bool) {
	return false
}


type RequireAllKeys []Lock
func (r RequireAllKeys) IsTrue() (ok bool) {
	ok = true
	for _, v := range r {
		if ok &&= v; !ok {
			break
		}
	}
	return
}

func (r RequireAllKeys) Unparse(player dbref, fullname, nested bool) (r string) {
	terms := make([]strings, len(r))
	for i, v := range r {
		terms[i] = v.Unparse(player, fullname)
	}
	r = strings.Join(terms, AND_TOKEN)
	if nested {
		r = "(" + r + ")"
	}
	return
}

func (r RequireAllKeys) Eval(descr int, player, thing dbref) (r bool) {
	return r.IsTrue()
}


type (r RequireAnyKey) []Lock
func (r RequireAnyKey) IsTrue() (ok bool) {
	for _, v := range r {
		if ok ||= v; ok {
			break
		}
	}
	return
}

func (r RequireAnyKey) Unparse(player dbref, fullname bool)  (r string) {
	terms := make([]strings, len(r))
	for i, v := range r {
		if _, ok := v.(RequireAllKeys); ok {
			terms[i] = fmt.Sprintf("(%v)", v.Unparse(player, fullname))
		} else {
			terms[i] = v.Unparse(player, fullname)
		}
	}
	r = strings.Join(terms, OR_TOKEN)
	return
}

func (r RequireAnyKeys) Eval(descr int, player, thing dbref) (r bool) {
	return r.IsTrue()
}


type IgnoreKey struct {
	Lock
}

func (i IgnoreKey) IsTrue() bool {
	return !i.Lock.IsTrue()
}

func (i IgnoreKey) Unparse(player dbref, fullname bool) (r string) {
	switch i.Lock.(type) {
	case RequireAnyKey, RequireAllKeys:
		r = fmt.Sprintf("!(%v)", i.Lock.Unparse(player, fullname))
	default:
		r = fmt.Sprintf("!%v", i.Lock.Unparse(player, fullname))
	}
	return
}

func (i IgnoreKey) Eval(descr int, player, thing dbref) (r bool) {
	return i.IsTrue()
}


type ObjectKey struct {
	dbref
}

func (o ObjectKey) IsTrue() bool {
	return o != NOTHING
}

func (o ObjectKey) Unparse(player dbref, fullname bool) (r string) {
	if fullname {
		r = unparse_object(player, o)
	} else {
		r = fmt.Sprintf("#%d", o)
	}
	return
}

func (o ObjectKey) Eval(descr int, player, thing dbref) (r bool) {
	if o.dbref != NOTHING {
		if _, ok := o.dbref.(TYPE_PROGRAM):
			var real_player dbref
			switch player.(type) {
			case TYPE_PLAYER, TYPE_THING:
				real_player = player
			default:
				real_player = db.Fetch(player).Owner
			}
			if tmpfr := interp(descr, real_player, db.Fetch(player).Location, o.dbref, thing, PREEMPT, STD_HARDUID, 0); tmpfr != nil {
				r = interp_loop(real_player, o.dbref, tmpfr, false) != nil
			}
		}
		r ||= o.dbref == player || o.dbref == db.Fetch(player).Owner || member(o.dbref, db.Fetch(player).Contents) || o.dbref == db.Fetch(player).Location
	}
	return
}


type PropertyKey struct {
	*Plist
}

func (p PropertyKey) IsTrue() bool {
	return false
}

func (p PropertyKey) Unparse(player dbref, fullname bool) string {
	return p.key + ":" + p.data.(string)
}

func (p PropertyKey) Eval(descr int, player, thing dbref) (r bool) {
	if v, ok := p.data.(string); ok {
		r = contains_property(descr, player, player, p.key, v, 0)
	}
	return
}


/* Lachesis note on the routines in this package:
 *   eval_booexp does just evaluation.
 *
 *   ParseLock makes potentially recursive calls to several different
 *   subroutines ---
 *        ParseLock_F
 *            This routine does the leaf level parsing and the NOT.
 *        ParseLock_E
 *            This routine does the ORs.
 *        ParseLock_T
 *            This routine does the ANDs.
 */

func copy_bool(l Lock) (r Lock) {
	if l.IsTrue() {
		return UNLOCKED
	}
	switch l := l.(type) {
	case RequireAllKeys:
		r = RequireAllKeys{ copy_bool(l[0]), copy_bool(l[1]) }
	case RequireAnyKey:
		r = RequireAnyKey{ copy_bool(l[0]), copy_bool(l[1]) }
	case IgnoreKey:
		r = IgnoreKey{ copy_bool(l.Lock) }
	case ObjectKey:
		r = ObjectKey{ l.dbref }
	case PropertyKey:
		if l.Plist == nil {
			r = nil
		} else {
			p := &PropertyKey{ Plist: NewPropNode(l.Plist.key) }
			SetPFlagsRaw(p.Plist, PropFlagsRaw(l.Plist))
			p.data = l.data
			r = p
		}
	default:
		panic("copy_bool(): Error in boolexp !")
	}
	return
}

/* If the parser returns UNLOCKED, you lose */
/* UNLOCKED cannot be typed in by the user; use @unlock instead */

/* F -> (E); F -> !F; F -> object identifier */
func ParseLock_F(descr int, lockdef string, player dbref, dbloadp int) (r Lock) {
	switch lockdef = strings.TrimSpace(lockdef); lockdef[0] {
	case '(':
		r = ParseLock_E(descr, lockdef[1:], player, dbloadp)
		lockdef = strings.TrimSpace(lockdef)
		if r.IsTrue() || lockdef[0] != ')' {
			r = UNLOCKED
		}
	case NOT_TOKEN:
		if r = IgnoreKey{ ParseLock_F(descr, lockdef[1:], player, dbloadp) }; r.IsTrue() {
			r = UNLOCKED
		}
	default:
		/* must have hit an object ref */
		/* load the name into our buffer */
		var buf string
		for lockdef != "" && lockdef[0] != AND_TOKEN && lockdef[0] != OR_TOKEN && lockdef[0] != ')' {
			buf += lockdef[0]
			lockdef = lockdef[1:]
		}
		buf = strings.TrimSpace(buf)
		if strings.Index(buf, PROP_DELIMITER) {
			r = parse_boolprop(buf)
		} else {
			if dbloadp {
				if i := strconv.Atoi(buf[1:]); buf[0] == NUMBER_TOKEN && !valid_reference(i) {
					r = ObjectKey{ i }
				} else {
					r = UNLOCKED
				}
			} else {
				r = ObjectKey{
					NewMatch(descr, player, buf, IsThing).
					MatchNeighbor().
					MatchPossession().
					MatchMe().
					MatchHere().
					MatchAbsolute().
					MatchRegistered().
					MatchPlayer().
					MatchResult()
				}

				switch r.dbref {
				case NOTHING:
					notify(player, fmt.Sprintf("I don't see %s here.", buf))
					r = UNLOCKED
				case AMBIGUOUS:
					notify(player, fmt.Sprintf("I don't know which %s you mean!", buf))
					r = UNLOCKED
				}
			}
		}
	}
	return
}

/* T -> F; T -> F & T */
func ParseLock_T(descr int, lockdef string, player dbref, dbloadp int) (r Lock) {
	if r = ParseLock_F(descr, lockdef, player, dbloadp); !r.IsTrue() {
		lockdef = strings.TrimSpace(lockdef)
		if lockdef[0] == AND_TOKEN {
			if r = RequireAllKeys{ r, ParseLock_T(descr, lockdef[1:], player, dbloadp) }; r[1].IsTrue() {
				r = UNLOCKED
			}
		}
	}
}

/* E -> T; E -> T | E */
func ParseLock_E(descr int, lockdef string, player dbref, dbloadp int) (r Lock) {
	if r = ParseLock_T(descr, lockdef, player, dbloadp); !r.IsTrue() {
		if lockdef = strings.TrimSpace(lockdef); lockdef[0] == OR_TOKEN {
			if r = RequireAnyKey{ r, ParseLock_E(descr, lockdef[1:], player, dbloadp) }; r[1].IsTrue() {
				r = UNLOCKED
			}
		}
	}
	return
}

func ParseLock(descr int, player dbref, lockdef string, dbloadp int) Lock {
	return ParseLock_E(descr, &lockdef, player, dbloadp)
}

/* parse a property expression
   If this gets changed, please also remember to modify set.c       */
func parse_boolprop(buf string) (r Lock) {
	var datatype, value string
	if datatype = strings.TrimSpace(buf); datatype[0] == PROP_DELIMITER {
		return UNLOCKED
	}

	if i := strings.Index(buf, PROP_DELIMITER); i != -1 {
		datatype = strings.TrimSpace(buf[:i])
		value = strings.TrimSpace(buf[i:])
	}
	if datatype == "" || value == "" {
		r = UNLOCKED
	} else {
		p := NewPropNode(datatype)
		p.data = value
		r = &PropertyKey{p}
	}
	return
}

func getboolexp1(f *FILE) (b Lock) {
	char buf[BUFFER_LEN];		/* holds string for reading in property */
	int i;						/* index into buf */

	c := getc(f)
	switch (c) {
	case EOF:
		panic("getboolexp1(): unexpected EOF in boolexp !");

	case '\n':
		ungetc(c, f);
		return UNLOCKED

	case '(':
		if c = getc(f); c == '!' {
			b = IgnoreKey{ getboolexp1(f) }
			if getc(f) != ')' {
				goto error
			}
			return
		} else {
			ungetc(c, f)
			switch c = getc(f); c {
			case AND_TOKEN:
				b = RequireAllKeys{ getboolexp1(f), getboolexp1(f) }
			case OR_TOKEN:
				b = RequireAnyKey{ getboolexp1(f), getboolexp1(f) }
			default:
				goto error;
			}
			if getc(f) != ')' {
				goto error
			}
			return
		}

	case '-':
		/* obsolete NOTHING key */
		/* eat it */
		while ((c = getc(f)) != '\n') {
			if (c == EOF) {
				panic("getboolexp1(): unexpected EOF in boolexp !");
			}
		}
		ungetc(c, f);
		return UNLOCKED

	case '[':
		/* property type */
		b = &PropertyKey{}
		i = 0;
		while ((c = getc(f)) != PROP_DELIMITER && i < BUFFER_LEN) {
			buf[i] = c;
			i++;
		}
		if (i >= BUFFER_LEN && c != PROP_DELIMITER)
			goto error;
		buf[i] = '\0';

		b.Plist = NewPropNode(buf)
		p := b->prop_check

		i = 0;
		while ((c = getc(f)) != ']') {
			if (c == '\\')
				c = getc(f);
			buf[i] = c;
			i++;
		}
		buf[i] = '\0';
		if (i >= BUFFER_LEN && c != ']')
			goto error;
		if !unicode.IsNumber(buf) {
			p.data = buf
		} else {
			p.data = strconv.Atol(buf)
		}
		return b;

	default:
		/* better be a dbref */
		ungetc(c, f)
		b = ObjectKey{}

		/* NOTE possibly non-portable code */
		/* Will need to be changed if putref/getref change */
		while (isdigit(c = getc(f))) {
			b.dbref = b.dbref * 10 + c - '0';
		}
		ungetc(c, f)
		return b
	}

  error:
	panic("getboolexp1(): error in boolexp !"); /* bomb out */
	return NULL;
}

func getboolexp(f *FILE) (l Lock) {
	l = getboolexp1(f)
	if getc(f) != '\n' {
		panic("getboolexp(): parse error !")
	}
	return
}