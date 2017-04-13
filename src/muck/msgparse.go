/* do_parse_mesg()  Parses a message string, expanding {prop}s, {list}s, etc.
 * Written 4/93 - 5/93 by Foxen
 *
 * Args:
 *   int descr             Descriptor that triggered this.
 *   dbref player          Player triggering this.
 *   dbref what            Object that mesg is on.
 *   dbref perms           Object permissions are checked from.
 *   const char *inbuf     A pointer to the input raw mesg string.
 *   char *abuf            Argument string for {&how}.
 *   int mesgtyp           1 for personal messages, 0 for omessages.
 *
 * Returns a pointer to output string.
 */

time_t mpi_prof_start_time;

func safeblessprop(dbref obj, dbref perms, char *buf, int mesgtyp, int set_p) int {
	char *ptr;

	if (!buf)
		return 0;
	buf = strings.TrimLeft(buf, PROPDIR_DELIMITER)
	if len(buf) == 0 {
		return 0
	}

	/* disallow CR's and :'s in prop names. */
	for (ptr = buf; *ptr; ptr++)
		if (*ptr == MPI_LISTSEP || *ptr == PROP_DELIMITER)
			return 0;

	if (!(mesgtyp & MPI_ISBLESSED)) {
		return 0;
	}
	if set_p {
		set_property_flags(obj, buf, PROP_BLESSED)
	} else {
		clear_property_flags(obj, buf, PROP_BLESSED)
	}

	return 1;
}

func safeputprop(dbref obj, dbref perms, char *buf, char *val, int mesgtyp) int {
	char *ptr;

	if (!buf)
		return 0;
	buf = strings.TrimLeft(buf, PROPDIR_DELIMITER)
	if buf == "" {
		return 0
	}

	/* disallow CR's and :'s in prop names. */
	for (ptr = buf; *ptr; ptr++)
		if (*ptr == MPI_LISTSEP || *ptr == PROP_DELIMITER)
			return 0;

	if (Prop_System(buf))
		return 0;
	
	if (!(mesgtyp & MPI_ISBLESSED)) {
		switch {
		case Prop_Hidden(buf), Prop_SeeOnly(buf), strings.Prefix(buf, "_msgmacs/"):
			return 0
		}
	}
	if (val == NULL) {
		remove_property(obj, buf);
	} else {
		add_property(obj, buf, val, 0);
	}
	return 1;
}

func safegetprop_strict(player, what, perms dbref, inbuf string, mesgtyp int, int* blessed) (r string, blessed bool) {
	inbuf = strings.TrimLeft(inbuf, PROPDIR_DELIMITER)
	switch {
	case inbuf == "":
		notify_nolisten(player, "PropFetch: Propname required.", true)
	case Prop_System(inbuf):
		notify_nolisten(player, "PropFetch: Permission denied.", true)
	default:
		if mesgtyp & MPI_ISBLESSED == 0 {
			if Prop_Hidden(inbuf) {
				notify_nolisten(player, "PropFetch: Permission denied.", true)
				return
			}
			if Prop_Private(inbuf) && db.Fetch(perms).Owner != db.Fetch(what).Owner {
				notify_nolisten(player, "PropFetch: Permission denied.", true)
				return
			}
		}

		if r = get_property_class(what, inbuf); r == "" {
			if i := get_property_value(what, inbuf); i == 0 {
				if dd := get_property_dbref(what, inbuf); dd == NOTHING {
					return
				} else {
					r = fmt.Sprintf("#%d", dd)
				}
			} else {
				r = fmt.Sprint(i)
			}
		}

		if r != "" {
			if Prop_Blessed(what, inbuf) {
				blessed = true
			}
		}
	}
	return
}

func safegetprop_limited(player, what, whom, perms dbref, inbuf string, mesgtyp int) (r string, blessed bool) {
	for ; what != NOTHING; what = getparent(what) {
		r, blessed = safegetprop_strict(player, what, perms, inbuf, mesgtyp)
		if db.Fetch(what).Owner == whom || blessed {
			break
		}
	}
	return
}

func safegetprop(player, what, perms dbref, inbuf string, mesgtyp int) (r string, blessed bool) {
	for ; what != NOTHING; what = getparent(what) {
		r, blessed = safegetprop_strict(player, what, perms, inbuf, mesgtyp)
		if r != "" {
			break
		}
	}
	return
}


char *
stripspaces(char *buf, int buflen, char *in)
{
	char *ptr;

	for (ptr = in; *ptr == ' '; ptr++) ;
	strcpyn(buf, buflen, ptr);
	ptr = len(buf) + buf - 1;
	while (*ptr == ' ' && ptr > buf)
		*(ptr--) = '\0';
	return buf;
}


char *
string_substitute(const char *str, const char *oldstr, const char *newstr,
                  char *buf, int maxlen)
{
	const char *ptr = str;
	char *ptr2 = buf;
	const char *ptr3;
	int len = len(oldstr);
	int clen = 0;

	if (len == 0) {
		strcpyn(buf, maxlen, str);
		return buf;
	}
	while (*ptr && clen < (maxlen+2)) {
		if strings.HasPrefix(ptr, oldstr) {
			for (ptr3 = newstr; ((ptr2 - buf) < (maxlen - 2)) && *ptr3;)
				*(ptr2++) = *(ptr3++);
			ptr += len;
			clen += len;
		} else {
			*(ptr2++) = *(ptr++);
			clen++;
		}
	}
	*ptr2 = '\0';
	return buf;
}

func get_list_item(player, what, perms dbref, listname string, itemnum, mesgtyp int) (r string, blessed bool) {
	if l := len(listname); listname[l - 1] == NUMBER_TOKEN {
		listname[l - 1] = 0
	}
	buf := fmt.Sprintf("%.512s#/%d", listname, itemnum)
	if r, blessed = safegetprop(player, what, perms, buf, mesgtyp); r == "" {
		buf = fmt.Sprintf("%.512s/%d", listname, itemnum)
		if r, blessed = safegetprop(player, what, perms, buf, mesgtyp); r == "" {
			buf = fmt.Sprintf("%.512s%d", listname, itemnum)
			r, blessed = safegetprop(player, what, perms, buf, mesgtyp)
		}
	}
	return
}

func get_list_count(dbref player, dbref obj, dbref perms, char *listname, int mesgtyp, int* blessed) (i int) {
	char buf[BUFFER_LEN];
	const char *ptr;
	l := len(listname)

	if (listname[len-1] == NUMBER_TOKEN) listname[l-1] = 0;

	buf = fmt.Sprintf("%.512s#", listname)
	ptr, blessed = safegetprop(player, obj, perms, buf, mesgtyp)
	if ptr != "" {
		return strconv.Atoi(ptr)
	}

	buf = fmt.Sprintf("%.512s/#", listname)
	ptr, blessed = safegetprop(player, obj, perms, buf, mesgtyp)
	if ptr != "" {
		return strconv.Atoi(ptr)
	}

	for i = 1; ; i++ {
		ptr, blessed = get_list_item(player, obj, perms, listname, i, mesgtyp)
		if (!ptr)
			return 0;
		if (!*ptr)
			break;
	}
	return
}

func get_concat_list(dbref player, dbref what, dbref perms, dbref obj, char *listname, char *buf, int maxchars, int mode, int mesgtyp, int* blessed) string {
	int i, cnt, len;
	const char *ptr;
	char *pos = buf;

	len = len(listname);
	if (listname[len-1] == NUMBER_TOKEN) listname[len-1] = 0;
	var tmpbless bool
	cnt, tmpbless = get_list_count(what, obj, perms, listname, mesgtyp)

	*blessed = 1;

	if (!tmpbless) {
		*blessed = 0;
	}

	if (cnt == 0) {
		return NULL;
	}
	maxchars -= 2;
	*buf = '\0';
	for (i = 1; ((pos - buf) < (maxchars - 1)) && i <= cnt; i++) {
		ptr, tmpbless = get_list_item(what, obj, perms, listname, i, mesgtyp)
		if (ptr) {
			if (!tmpbless) {
				*blessed = 0;
			}
			ptr = strings.TrimLeftFunc(buf, func(r rune) bool {
				return mode && unicode.IsSpace(r)
			})
			if (pos > buf) {
				if (!mode) {
					*(pos++) = MPI_LISTSEP;
					*pos = '\0';
				} else if (mode == 1) {
					char ch = *(pos - 1);

					if ((pos - buf) >= (maxchars - 2))
						break;
					if (ch == '.' || ch == '?' || ch == '!')
						*(pos++) = ' ';
					*(pos++) = ' ';
					*pos = '\0';
				} else {
					*pos = '\0';
				}
			}
			while (((pos - buf) < (maxchars - 1)) && *ptr)
				*(pos++) = *(ptr++);
			if (mode) {
				while (pos > buf && *(pos - 1) == ' ')
					pos--;
			}
			*pos = '\0';
			if ((pos - buf) >= (maxchars - 1))
				break;
		}
	}
	return (buf);
}

func mesg_read_perms(dbref player, dbref perms, dbref obj, int mesgtyp) (r bool) {
	switch {
	case obj == 0, obj == player, obj == perms:
		r = true
	case db.Fetch(perms).Owner == db.Fetch(obj).Owner:
		r = true
	case mesgtyp & MPI_ISBLESSED != 0:
		r = true
	}
	return
}

func isneighbor(d1, d2 dbref) (r bool) {
	if d1 == d2 {
		return true
	}
	if TYPEOF(d1) != TYPE_ROOM && db.Fetch(d1).Location == d2 {
		return true
	}
	if TYPEOF(d2) != TYPE_ROOM && db.Fetch(d2).Location == d1 {
		return true
	}
	if Typeof(d1) != TYPE_ROOM && TYPEOF(d2) != TYPE_ROOM && db.Fetch(d1).Location == db.Fetch(d2).Location {
		return 1
	}
	return
}

func mesg_local_perms(dbref player, dbref perms, dbref obj, int mesgtyp) (r bool) {
	if r = db.Fetch(obj).Location != NOTHING && db.Fetch(perms).Owner == db.Fetch(db.Fetch(obj).Location).Owner; !r {
		r = isneighbor(perms, obj) || isneighbor(player, obj) || mesg_read_perms(player, perms, obj, mesgtyp) {
	}
	return
}

func mesg_dbref_raw(descr int, player, what, perms dbref, buf string) (obj dbref) {
	obj = UNKNOWN
	switch buf {
	case "":
	case "this":
		obj = what
	case "me":
		obj = player
	case "here":
		obj = db.Fetch(player).Location
	case "home":
		obj = HOME
	default:
		obj = NewMatch(descr, player, buf, NOTYPE).
			MatchAbsolute().
			MatchAllExits().
			MatchNeighbor().
			MatchPossession().
			MatchRegistered().
			MatchResult()
		if obj == NOTHING {
			obj = NewMatchRemote(descr, player, what, buf, NOTYPE).
				MatchPlayer().
				MatchAllExits().
				MatchNeighbor().
				MatchPossession().
				MatchRegistered().
				MatchResult()
		}
	}

	if !valid_reference(obj) {
		obj = UNKNOWN
	}
	return obj
}

func mesg_dbref(int descr, dbref player, dbref what, dbref perms, char *buf, int mesgtyp) (r dbref) {
	if r = mesg_dbref_raw(descr, player, what, perms, buf); r != UNKNOWN {
		if !mesg_read_perms(player, perms, r, mesgtyp) {
			r = PERMDENIED
		}
	}
	return
}

func mesg_dbref_strict(int descr, dbref player, dbref what, dbref perms, char *buf, int mesgtyp) (r dbref) {
	if r = mesg_dbref_raw(descr, player, what, perms, buf); r != UNKNOWN {
		if !(mesgtyp & MPI_ISBLESSED) && db.Fetch(perms).Owner != db.Fetch(r).Owner {
			r = PERMDENIED
		}
	}
	return
}

func mesg_dbref_local(int descr, dbref player, dbref what, dbref perms, char *buf, int mesgtyp) (r dbref) {
	if r = mesg_dbref_raw(descr, player, what, perms, buf); r != UNKNOWN {
		if !mesg_local_perms(player, perms, r, mesgtyp) {
			r = PERMDENIED
		}
	}
	return
}

func ref2str(obj dbref) (r string) {
	switch {
	case valid_reference(obj) && IsPlayer(obj):
		r = fmt.Sprintf("*%s", db.Fetch(obj).name)
	case obj == NOTHING, obj == HOME, obj == AMBIGUOUS
		r = fmt.Sprintf("#%d", obj)
	default:
		r = "Bad"
	}
	return
}

func truestr(buf string) bool {
	buf = strings.TrimLeftFunc(buf, unicode.IsSpace)
	if !*buf || (unicode.IsNumber(buf) && !atoi(buf)) {
		return 0
	}
	return 1
}


type MPIValue struct {
	name, value string
	next *MPIValue
}

var MPI_VARIABLES *MPIValue
var mpi_functions *MPIValue

func new_mvalues(env *MPIValue, name ...string) {
	for _, v := range name {
		*env = &MPIValue{ name: name, next: env }
	}
}

func find_mvalue(env *MPIValue, name string) (n *MPIValue) {
	for n = env; n != nil && name != n.name; n = n.next {}
	return
}

func get_mvalue(env *MPIValue, name string) (r string) {
	if n := find_mvalue(env); n != nil {
		r = n.buf
	}
	return 
}

func set_mvalue(env *MPIValue, name, value string) {
	if n := find_mvalue(env); n != nil {
		n.buf = value
	}
	return 
}

func drop_mvalues(env *MPIValue, name) {
	if n := find_mvalue(env); n != nil {
		*env = n.next
	}
}

func free_mvalues(env *MPIValue, downto int) {
	for i := downto; i > 0 && env != nil; env = env.next {
		*env = (*env).next
	}
}

func msg_is_macro(player, what, perms dbref, name string, mesgtyp int) (r bool) {
	if name != "" {
		var blessed bool
		buf := fmt.Sprintf("_msgmacs/%s", name)
		f := get_mvalue(mpi_functions, name)
		if f == "" {
			f, blessed = safegetprop_strict(player, db.Fetch(what).Owner, perms, buf, mesgtyp)
		}
		if f == "" {
			f, blessed = safegetprop_limited(player, obj, db.Fetch(what).Owner, perms, buf, mesgtyp)
		}
		if f == "" {
			f, blessed = safegetprop_strict(player, 0, perms, buf, mesgtyp)
		}
		r = f != ""
	}
	return
}

func msg_unparse_macro(player, what, perms dbref, name string, argv MPIArgs, rest string, mesgtyp int) (r string) {
	const char *ptr;
	char *ptr2;
	dbref obj;
	int i, p = 0;

	buf := rest
	buf2 := fmt.Sprintf("_msgmacs/%s", name)
	f := get_mvalue(mpi_functions, name)
	var blessed bool
	if f == "" {
		f, blessed = safegetprop_strict(player, db.Fetch(what).Owner, perms, buf2, mesgtyp)
	}
	if f == "" {
		f, blessed = safegetprop_limited(player, what, db.Fetch(what).Owner, perms, buf2, mesgtyp)
	}
	if f == "" {
		f, blessed = safegetprop_strict(player, 0, perms, buf2, mesgtyp)
	}
	for f != "" {
		switch f[0] {
		case '\\':
			switch f[1] {
			case 'r':
				r += MPI_LISTSEP
			case '[':
				r += ESCAPE_CHAR
			default:
				r += f[0:2]
			}
			f = f[2:]
		case MFUN_LEADCHAR:
			if f[1] == MFUN_ARGSTART && isdigit(f[2]) && f[3] == MFUN_ARGEND {
				i = f[2] - '1'
				f = f[3:]
				if i >= len(argv) || i < 0 {
					ptr2 = ""
				} else {
					ptr2 = argv[i]
				}
				for ptr2 !+ "" && p < maxchars - 1 {
					r += ptr2[0]
					ptr2 = ptr2[1:]
				}
			} else {
				r += f[0]
				f = f[1:]
			}
		} else {
			r += f[0]
			f = f[1:]
		}
	}
	ptr2 = buf
	for ptr2 != "" {
		r += ptr2[0]
		ptr2++
	}
}

var mpi_messages map[string] int

func init() {
	mpi_messages = make(map[string] int)
}

#define DEFINE_MFUN_LIST
#include "mfunlist.h"


/******** HOOK ********/
func mesg_init() {
	for i, v := range mfun_list {
		mpi_messages[v.name] = i + 1
	}
	mpi_prof_start_time = time(NULL)
}

/******** HOOK ********/
func mesg_args(mesg string, maxargs int) (r []strings, buf string, e error) {
	var in_literal, in_escape bool
	var nesting int
	var word string
	buf := mesg
	l := len(buf)
	for i, v := range buf {
		switch {
		case in_literal:
			if v == MFUN_LITCHAR {
				in_literal = false
			} else {
				word += v
			}
		case in_escape:
			if i < l {
				word += v
			} else {
				e = error.New("incomplete escape character")
				r = append(r, word)
				buf = buf[i:]
				break
			}
		default:
			switch v {
			case '\\':
				if in_escape {
					word += '\\'
				}
				in_escape = !in_escape
			case MFUN_LITCHAR:
				in_literal = true
			case MFUN_LEADCHAR:
				nesting++
			case MFUN_ARGEND:
				if nesting--; nesting < 0 {
					e = error.New("unmatched right parenthesis")
					r = append(r, word)
					buf = buf[i:]
					break
				}
			case MFUN_ARGSEP:
				if nesting < 1 {
					r = append(r, word)
					buf = buf[i:]
					word = ""
				}
			default:
				word += v
			}
		}
	}
	return
}

func cr2slash(in string) (r string) {
	for _, v := range in {
		switch ptr2[0] {
		case MPI_LISTSEP:
			r += "\\r"
		case ESCAPE_CHAR:
			r += "\\["
		case MFUN_LITCHAR:
			r += "\\" + MFUN_LITCHAR
		case '\\':
			r += "\\\\"
		default:
			r += ptr2[0]
		}
	}
	return
}


static int mesg_rec_cnt = 0;
static int mesg_instr_cnt = 0;


/******** HOOK ********/
func mesg_parse(descr int, player, what, perms dbref, inbuf string, mesgtyp int) (r string) {
	char buf[BUFFER_LEN]
	char buf2[BUFFER_LEN]
	char dbuf[BUFFER_LEN]
	char ebuf[BUFFER_LEN]
	const char *ptr;
	char *dptr;
	int p, q, s;
	var showtext, in_literal bool

	var args []string
	mesg_rec_cnt++
	if mesg_rec_cnt > 26 {
		notify_nolisten(player, fmt.Sprintf("%s Recursion limit exceeded.", get_mvalue(MPI_VARIABLES, "how")), true)
		mesg_rec_cnt--
		return
	}
	for wbuf := inbuf; wbuf != ""; wbuf = wbuf[1:] {
		switch v := wbuf[0]; {
		case v == '\\':
			wbuf = wbuf[1:]
// FIXME: filter for escape character without following character
			showtext = true
			switch v := wbuf[0]; v {
			case 'r':
				r += MPI_LISTSEP
			case '[':
				r += ESCAPE_CHAR
			default:
				r += v
			}
		case v == MFUN_LITCHAR:
			in_literal = !in_literal
		case !in_literal && v == MFUN_LEADCHAR:
			if wbuf[1] == MFUN_LEADCHAR {
				showtext = true
				r += v
				wbuf = wbuf[1:]
				ptr = ""
			} else {
				wbuf = wbuf[1:]
				ptr = wbuf
				s = 0
				for v := wbuf[0]; wbuf != "" && v != MFUN_LEADCHAR && !unicode.IsSpace(v) && v != MFUN_ARGSTART && v != MFUN_ARGEND; v = wbuf[0] {
					wbuf = wbuf[1:]
					s++
				}
				if v := wbuf[0]; v != MFUN_ARGSTART && v != MFUN_ARGEND {
					showtext = true
					ptr--
					for i := s + 1; ptr != "" && i > 0; i-- {
						r += ptr[0]
						ptr = ptr[1:]
					}
					p = int(ptr - wbuf) - 1
					ptr = ""	/* unknown substitution type */
				} else {
					if cmdbuf := ptr[:s]; cmdbuf[0] == '&' {
						s = mpi_messages["sublist"]
						switch {
						case s != 0:
							s--
							if mesg_instr_cnt++; mesg_instr_cnt > tp_mpi_max_commands {
								notify_nolisten(player, fmt.Sprintf("%v %v%v%v: Instruction limit exceeded.", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, cmdbuf, MFUN_ARGEND), true)
								mesg_rec_cnt--
								r = ""
								return
							}
							if wbuf[0] == MFUN_ARGEND {
								args = make([]string)
							} else {
								var e error
								oldargs := args
								args, wbuf, e = mesg_args(wbuf[1:])
								copy(oldargs[1:], args)
								args = oldargs
								if e != nil {
									notify_nolisten(player, fmt.Sprintf("%s %c%s%c: End brace not found.", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, cmdbuf, MFUN_ARGEND), true)
									mesg_rec_cnt--
									r = ""
									return
								}
							}
							if get_mvalue(MPI_VARIABLES, cmdbuf[1:]) == "" {
								notify_nolisten(player, fmt.Sprintf("%s %c%s%c: Unrecognized variable.", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, cmdbuf, MFUN_ARGEND), true)
								mesg_rec_cnt--
								r = ""
								return
							}
							args[0] = get_mvalue(MPI_VARIABLES, "how")
							if mesgtyp & MPI_ISDEBUG != 0 {
								dbuf := fmt.Sprintf("%s %*s%c%s%c", get_mvalue(MPI_VARIABLES, "how"), (mesg_rec_cnt * 2 - 4), "", MFUN_LEADCHAR, cmdbuf, MFUN_ARGSTART)
								for _, v := range args[1:] {
									dbuf += MFUN_ARGSEP + MFUN_LITCHAR + cr2slash(v) + MFUN_LITCHAR
								}
								dbuf += MFUN_ARGEND
								notify_nolisten(player, dbuf, true)
							}
							if mfun_list[s].strip_space {
								for i, v := range args[1:] {
									args[i] = strings.TrimFunc(v, unicode.IsSpace)
								}
							}
							if mfun_list[s].preparse {
								for i, v := range args[1:] {
									if r = mesg_parse(descr, player, what, perms, v, buf, mesgtyp); r == "" {
										notify_nolisten(player, fmt.Sprintf("%s %c%s%c (arg %d)", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, cmdbuf, MFUN_ARGEND, i + 1), true)
										mesg_rec_cnt--
										return
									}
									args[i] = buf
								}
							}
							if mesgtyp & MPI_ISDEBUG != 0 {
								dbuf = fmt.Sprintf("%.512s %*s%c%.512s%c", get_mvalue(MPI_VARIABLES, "how"), (mesg_rec_cnt * 2 - 4), "", MFUN_LEADCHAR, cmdbuf, MFUN_ARGSTART)
								for i, v := range argv[1:] {
									dbuf += MFUN_ARGSEP + MFUN_LITCHAR + cr2slash(v) + MFUN_LITCHAR
								}
								dbuf += MFUN_ARGEND
							}
							switch {
							case len(args) < mfun_list[s].minargs:
								notify_nolisten(player, fmt.Sprintf("%s %c%s%c: Too few arguments", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, cmdbuf, MFUN_ARGEND), true)
								mesg_rec_cnt--
								r = ""
								return
							case mfun_list[s].maxargs > 0 && len(args) > mfun_list[s].maxargs:
								notify_nolisten(player, fmt.Sprintf("%s %c%s%c: Too many arguments", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, cmdbuf, MFUN_ARGEND), true)
								mesg_rec_cnt--
								r = ""
								return
							default:
								switch ptr = mfun_list[s].mfn(descr, player, what, perms, argv, mesgtyp); {
								case ptr == nil:
									mesg_rec_cnt--
									r = ""
									return
								case mfun_list[s].postparse:
									if r = mesg_parse(descr, player, what, perms, ptr, buf, mesgtyp); r == "" {
										notify_nolisten(player, ("%s %c%s%c (returned string)", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, cmdbuf, MFUN_ARGEND), true)
										mesg_rec_cnt--
										return
									}
									ptr = dptr
								}
							}
							if mesgtyp & MPI_ISDEBUG != 0 {
								dbuf += " = " + MFUN_LITCHAR + cr2slash(ptr) + MFUN_LITCHAR
								notify_nolisten(player, dbuf, true)
							}
						case msg_is_macro(player, what, perms, cmdbuf, mesgtyp):
							if wbuf[0] == MFUN_ARGEND {
								wbuf = wbuf[1:]
							} else {
								wbuf = wbuf[1:]
								if args, wbuf, e = mesg_args(wbuf); e != nil {
									notify_nolisten(player, fmt.Sprintf("%s %c%s%c: %v", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, cmdbuf, MFUN_ARGEND, e), true)
									mesg_rec_cnt--
									r = ""
									return
								}
							}
							wbuf = msg_unparse_macro(player, what, perms, cmdbuf, argv, wbuf, mesgtyp)
							p--
							ptr = ""
						default:
							notify_nolisten(player, fmt.Sprintf("%s %c%s%c: Unrecognized function.", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, cmdbuf, MFUN_ARGEND), true)
							mesg_rec_cnt--
							r = ""
							return
						}
					} else {
						s = mpi_messages[cmdbuf]
						switch {
						case s != 0:
							s--
							if mesg_instr_cnt++; mesg_instr_cnt > tp_mpi_max_commands {
								notify_nolisten(player, fmt.Sprintf("%v %v%v%v: Instruction limit exceeded.", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, mfun_list[s].name, MFUN_ARGEND), true)
								mesg_rec_cnt--
								r = ""
								return
							}
							if wbuf[0] == MFUN_ARGEND {
								args = make([]string)
							} else {
								if args, wbuf, e = mesg_args(wbuf[1:]); e != nil {
									notify_nolisten(player, fmt.Sprintf("%s %c%s%c: End brace not found.", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, cmdbuf, MFUN_ARGEND), true)
									mesg_rec_cnt--
									r = ""
									return
								}
							}
							if mesgtyp & MPI_ISDEBUG != 0 {
								var dbuf string
								dbuf = fmt.Sprintf("%s %*s%c%s%c%c%s%c", get_mvalue(MPI_VARIABLES, "how"), (mesg_rec_cnt * 2 - 4), "", MFUN_LEADCHAR, mfun_list[s].name, MFUN_ARGSTART, MFUN_LITCHAR + cr2slash(args[0]) + MFUN_LITCHAR)
								for _, v := range args[1:] {
									dbuf += MFUN_ARGSEP + MFUN_LITCHAR + cr2slash(v) + MFUN_LITCHAR
								}
								dbuf += MFUN_ARGEND
								notify_nolisten(player, dbuf, true)
							}
							if mfun_list[s].strip_space {
								var i int
								for i, v := range args {
									args[i] = strings.TrimFunc(v, unicode.IsSpace)
								}
							}
							if mfun_list[s].preparse {
								var i int
								for i, v := range args {
									if r = mesg_parse(descr, player, what, perms, args[i], buf, mesgtyp); r == "" {
										notify_nolisten(player, fmt.Sprintf("%s %c%s%c (arg %d)", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, mfun_list[s].name, MFUN_ARGEND, i + 1), true)
										mesg_rec_cnt--
										return
									}
									args[i] = buf
								}
							}
							if mesgtyp & MPI_ISDEBUG != 0 {
								dbuf = fmt.Sprintf("%.512s %*s%c%.512s%c%c%s%v", get_mvalue(MPI_VARIABLES, "how"), (mesg_rec_cnt * 2 - 4), "", MFUN_LEADCHAR, mfun_list[s].name, MFUN_ARGSTART, MFUN_LITCHAR + cr2slash(args[0]) + MFUN_LITCHAR)
								for i, v := range argv[1:] {
									dbuf += MFUN_ARGSEP + MFUN_LITCHAR + cr2slash(v) + MFUN_LITCHAR
								}
								dbuf += MFUN_ARGEND
							}
							switch {
							case len(args) < mfun_list[s].minargs:
								ebuf = fmt.Sprintf("%s %c%s%c: Too few arguments", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, mfun_list[s].name, MFUN_ARGEND)
								notify_nolisten(player, ebuf, true)
								mesg_rec_cnt--
								r = ""
								return
							case mfun_list[s].maxargs > 0 && len(argv) > mfun_list[s].maxargs:
								ebuf = fmt.Sprintf("%s %c%s%c: Too many arguments", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, mfun_list[s].name, MFUN_ARGEND)
								notify_nolisten(player, ebuf, true)
								mesg_rec_cnt--
								r = ""
								return
							default:
								if ptr = mfun_list[s].mfn(descr, player, what, perms, argv, mesgtyp); ptr == nil {
									mesg_rec_cnt--
									r = ""
									return
								}
								if mfun_list[s].postparse {
									if r = mesg_parse(descr, player, what, perms, ptr, buf, mesgtyp); r == "" {
										notify_nolisten(player, ("%s %c%s%c (returned string)", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, mfun_list[s].name, MFUN_ARGEND), true)
										mesg_rec_cnt--
										return
									}
									ptr = dptr
								}
							}
							if mesgtyp & MPI_ISDEBUG != 0 {
								dbuf += " = " + MFUN_LITCHAR + cr2slash(ptr) + MFUN_LITCHAR
								notify_nolisten(player, dbuf, true)
							}
						case msg_is_macro(player, what, perms, cmdbuf, mesgtyp):
							if wbuf[0] == MFUN_ARGEND {
								wbuf = wbuf[1:]
							} else {
								wbuf = wbuf[1:]
								if args, wbuf, e = mesg_args(wbuf); e != nil {
									notify_nolisten(player, fmt.Sprintf("%s %c%s%c: %v", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, cmdbuf, MFUN_ARGEND, e), true)
									mesg_rec_cnt--
									r = ""
									return
								}
							}
							wbuf = msg_unparse_macro(player, what, perms, cmdbuf, argv, wbuf, mesgtyp)
							p--
							ptr = ""
						default:
							/* unknown function */
							notify_nolisten(player, fmt.Sprintf("%s %c%s%c: Unrecognized function.", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, cmdbuf, MFUN_ARGEND), true)
							mesg_rec_cnt--
							r = ""
							return
						}
					}
				}
				for ptr != "" {
					r += ptr[0]
					ptr = ptr[1:]
				}
			}
		} else {
			r += wbuf[0]
			showtext = true
		}
	}
	if mesgtyp & MPI_ISDEBUG != 0 && showtextflag {
		notify_nolisten(player, fmt.Sprintf("%s %*s`%.512s`", get_mvalue(MPI_VARIABLES, "how"), (mesg_rec_cnt * 2) - 4, "", cr2slash(r)), true)
	}
	mesg_rec_cnt--
	return
}

func do_parse_mesg_2(descr int, player, what, perms dbref, inbuf, abuf string, mesgtyp int) (r string) {
	mfunccnt := len(mpi_functions)
	tmprec_cnt := mesg_rec_cnt
	tmpinst_cnt := mesg_instr_cnt

	howvar := abuf
	if mesgtyp & MPI_NOHOW == 0 {
		if get_mvalue(MPI_VARIABLES, "how") == howvar {
			notify_nolisten(player, fmt.Sprintf("%s Out of MPI variables.", howvar), true)
			return
		}
	}

	cmdvar := match_cmdname
	set_mvalue(MPI_VARIABLES, "cmd", cmdvar)
	tmpcmd := match_cmdname

	argvar := match_args
	set_mvalue(MPI_VARIABLES, "arg", argvar)
	tmparg := match_args

	r = mesg_parse(descr, player, what, perms, inbuf, mesgtyp)
	free_mvalues(&mpi_functions, mfunccnt)
	mesg_rec_cnt = tmprec_cnt
	mesg_instr_cnt = tmpinst_cnt

	match_cmdname = tmpcmd
	match_args = tmparg
	return
}


func do_parse_mesg(descr int, player, what dbref, inbuf, abuf, mesgtyp int) (r string) {
	if tp_do_mpi_parsing {
		/* Quickie additions to do rough per-object MPI profiling */
		st := time.Now()
		tmp := do_parse_mesg_2(descr, player, what, what, inbuf, abuf, mesgtyp)
		et := time.Now()
		if tmp != inbuf {
			if subject := db.Fetch(what); subject != nil {
				subject.time.Duration += et
				subject.MPIUses++
			}
		}
		r = tmp
	} else {
		r = inbuf
	}
	return
}

func do_parse_prop(descr int, player, what dbref, propname string, abuf, mesgtyp int) (r string) {
	if propval := get_property_class(what, propname); propval != nil {
		if Prop_Blessed(what, propname) {
			mesgtyp |= MPI_ISBLESSED
		}
		r = do_parse_mesg(descr, player, what, propval, abuf, mesgtyp)
	}
	return
}