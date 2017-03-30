package fbmuck

const (
	MFUN_LITCHAR = '`'
	MFUN_LEADCHAR = '{'
	MFUN_ARGSTART = ':'
	MFUN_ARGSEP = ','
	MFUN_ARGEND = '}'
	MPI_LISTSEP = '\r'
)

#define UNKNOWN ((dbref)-88)
#define PERMDENIED ((dbref)-89)

#define CHECKRETURN(vari,funam,num) if vari == "" { \
	buf = fmt.Sprintf("%s %c%s%c (%s)", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, funam, MFUN_ARGEND, num); \
	notify_nolisten(player, buf, true); \
	return nil \
}

#define ABORT_MPI(funam,mesg) { \
	buf = fmt.Sprintf("%s %c%s%c: %s", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, funam, MFUN_ARGEND, mesg); \
	notify_nolisten(player, buf, true);  \
	return nil \
}

type MPIArgs []string

func ForceAction(program dbref, f func()) {
	force_prog = program
	force_level++
	f()
	force_level--
	force_prog = NOTHING
}

func mpi_use_variable(argv MPIArgs, f func(i, v int)) {
	if v := find_mvalue(MPI_VARIABLES, argv[0]); v == nil {
		ABORT_MPI("INC", "No such variable currently defined.")
	} else  {
		x := 1
		if len(argv) > 1 {
			x = strconv.Atoi(argv[1])
		}
		f(strconv.Atoi(v.buf), x)
	}
}

func mpi_list_commas(list []string, sep string) (r string) {
	switch l := len(list); l {
	case 0:
	case 1:
		r = items[0]
	case 2:
		buf = items[0] + sep + items[1]
	default:
		buf = strings.Join(items[:l - 2], ", ") + sep + items[l - 1]
	}
}

func mpi_list_remove(llist, rlist []string) (r []string) {
	m := map[string] bool
	for _, v := range llist {
		m[v] = true
	}
	for _, v := range rlist {
		m[v] = false
	}
	for k, v := range m {
		if v {
			r = append(r, k)
		}
	}
	return
}

func mpi_list_common(llist, rlist []string) (r []string) {
	switch {
	case len(llist) == 0, len(rlist) == 0:
	default:
		lm := make(map[string] bool)
		for _, v := range llist {
			lm[v] = true
		}

		rm := make(map[string] bool)
		for _, v := range rlist {
			rm[v] = true
		}

		for k, v := range lm {
			if rm[k] {
				r = append(r, k)
			}
		}
	}
	return
}

func mpi_list_union(llist, rlist []string) (r []string) {
	switch {
	case len(llist) == 0:
		r = rlist
	case len(rlist) == 0:
		r = llist
	default:
		if len(rlist) < len(llist) {
			llist, rlist = rlist, llist	
		}
		r = make([]string, len(llist), len(llist) * 1.25)
		copy(r, llist)
		p := 0
		m := make(map[string] bool)
		for _, v := range r {
			m[v] = true
		}
		for _, v := range rlist {
			if !m[v] {
				r = append(r, v)
			}
		}
	}
	return
}