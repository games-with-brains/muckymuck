package fbmuck

import "os"

/***** Insert MFUNs here *****/
func mfn_owner(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("OWNER", mesg_ObjectID_raw(descr, player, what, perms, argv[0]), func(obj ObjectID) {
		if obj == HOME {
			obj = DB.FetchPlayer(player).Home
		}
		r = ref2str(DB.Fetch(obj).Owner)
	})
	return
}

func mfn_controls(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("CONTROLS (1)", mesg_ObjectID_raw(descr, player, what, perms, argv[0]), func(obj ObjectID) {
		if obj == HOME {
			obj = DB.FetchPlayer(player).Home
		}
		obj2 := DB.Fetch(perms).Owner
		if len(argv) >  1 {
			with_useful_object("CONTROLS (2)", mesg_ObjectID_raw(descr, player, what, perms, argv[1]), func(o ObjectID) {
				if obj2 = o; o == HOME {
					obj2 = DB.FetchPlayer(player).Home
				}
				if !IsPlayer(obj2) {
					obj2 = DB.Fetch(obj2).Owner
				}
			})
		} else {
			obj2 = DB.Fetch(perms).Owner
		}
		r = MPIBool(controls(obj2, obj))
	})
	return
}

func mfn_links(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("LINKS", mesg_ObjectID(descr, player, what, perms, argv[0], mesgtyp), func(obj ObjectID) {
		switch o := DB.Fetch(o).(type) {
		case Room:
			r = ref2str(o.ObjectID)
		case Player:
			r = ref2str(o.Home)
		case Object:
			r = ref2str(o.Home)
		case Exit:
			var items []string
			for _, v := range o.Destinations {
				items = append(items, ref2str(v))
			}
			r = strings.Join(items, MPI_LISTSEP)
		}
		if r = "" {
			r = ref2str(obj)
		}
	})
	return
}

func mfn_locked(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("LOCKED (1)", mesg_ObjectID_local(descr, player, what, perms, argv[0], mesgtyp), func(who ObjectID) {
		with_useful_object("LOCKED (2)", mesg_ObjectID_local(descr, player, what, perms, argv[1], mesgtyp), func(obj ObjectID) {
			r = fmt.Sprint(!could_doit(descr, who, obj))
		})
	})
	return
}

func mfn_testlock(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("TESTLOCK (1)", player, func(who ObjectID) {
		obj := mesg_ObjectID_local(descr, player, what, perms, argv[0], mesgtyp)
		if len(argv) > 2 {
			who = mesg_ObjectID_local(descr, player, what, perms, argv[2], mesgtyp)
		}
		with_useful_object("TESTLOCK (2)", obj, func(obj ObjectID) {
			switch {
			case Prop_System(argv[1]):
				ABORT_MPI("TESTLOCK", "Permission denied. (arg2)")
			case mesgtyp & MPI_ISBLESSED == 0 && Prop_Hidden(argv[1]):
				ABORT_MPI("TESTLOCK", "Permission denied. (arg2)")
			case mesgtyp & MPI_ISBLESSED == 0 && Prop_Private(argv[1]) && DB.Fetch(perms).Owner != DB.Fetch(what):
				ABORT_MPI("TESTLOCK", "Permission denied. (arg2)")
			default:
				switch lok := get_property_lock(obj, argv[1]); {
				// FIXME: This case is probably wrong - surely default should be for LOCKED?
				case len(argv) > 3 && lok == UNLOCKED:
					r = argv[3]
				case copy_bool(lok).Eval(descr, who, obj):
					r = "1"
				default:
					r = "0"
				}
			}
		})
	})
	return
}

func mfn_contents(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("CONTENTS", mesg_ObjectID_local(descr, player, what, perms, argv[0], mesgtyp), func(obj ObjectID) {
		T := NOTYPE
		if len(argv) > 1 {
			switch argv[1] {
			case "Room":
				T = TYPE_ROOM
			case "Exit":
				T = TYPE_EXIT	/* won't find any, though */
			case "Player":
				T = TYPE_PLAYER
			case "Program":
				T = TYPE_PROGRAM
			case "Thing":
				T = TYPE_THING
			default:
				ABORT_MPI("CONTENTS", "Type must be 'player', 'room', 'thing', 'program', or 'exit'. (arg2).")
			}
		}
		ownroom := controls(perms, obj)
		var items []string
		for obj = DB.Fetch(obj).Contents; obj != NOTHING; obj = DB.Fetch(obj).next {
			if (T == NOTYPE || Typeof(obj) == T) && (ownroom || controls(perms, obj) || !(DB.Fetch(obj).IsFlagged(DARK) || DB.Fetch(DB.Fetch(obj).Location).IsFlagged(DARK) || (IsProgram(obj) && !DB.Fetch(obj).IsFlagged(LINK_OK)))) && !(IsRoom(obj) && !IsRoom(T)) {
				items = append(items, ref2str(obj))
			}
		}
		r = strings.Join(r, MPI_LISTSEP)
	})
	return
}

func mfn_exits(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("EXITS", mesg_ObjectID(descr, player, what, perms, argv[0], mesgtyp), func(obj ObjectID) {
		switch obj.(type) {
		case TYPE_ROOM, TYPE_THING, TYPE_PLAYER:
			obj = DB.Fetch(obj).Exits
		default:
			obj = NOTHING
		}
		var items []string
		for ; obj != NOTHING; obj = DB.Fetch(obj).next {
			items = append(items, ref2str(obj))
		}
		r = strings.Join(items, MPI_LISTSEP)
	})
	return
}

func mfn_v(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if n := find_mvalue(MPI_VARIABLES, argv[0]); n == nil {
		ABORT_MPI("V", "No such variable defined.")
	} else {
		r = n.value
	}
	return
}

func mfn_set(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	if n := find_mvalue(MPI_VARIABLES, argv[0]); n == nil {
		ABORT_MPI("SET", "No such variable currently defined.")
	} else {
		n.value = argv[1]
	}
	return argv[1]
}

func mfn_ref(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	p := strings.TrimLeftFunc(os.Args[0], unicode.IsSpace)
	var obj ObjectID
	if p[0] == NUMBER_TOKEN && unicode.IsNumber(p[1]) {
		obj = strconv.Atoi(p[1:])
	} else {
		switch obj = mesg_ObjectID_local(descr, player, what, perms, argv[0], mesgtyp); obj {
		case PERMDENIED:
			ABORT_MPI("REF", "Permission denied.")
		case UNKNOWN:
			obj = NOTHING
		}
	}
	return fmt.Sprintf("#%d", obj)
}

func mfn_name(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	switch obj := mesg_ObjectID_raw(descr, player, what, perms, argv[0]); obj {
	case UNKNOWN:
		ABORT_MPI("NAME", "Match failed.")
	case PERMDENIED:
		ABORT_MPI("NAME", "Permission denied.")
	case NOTHING:
		r = "#NOTHING#"
	case AMBIGUOUS:
		r = "#AMBIGUOUS#"
	case HOME:
		r = "#HOME#"
	default:
		r = DB.Fetch(obj).name
		if Typeof(obj) == TYPE_EXIT {
			if items := strings.Split(r, ";", 2); len(items) > 0 {
				r = items[0]
			} else {
				r = ""
			}
		}
	}
	return
}

func mfn_fullname(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	switch obj := mesg_ObjectID_raw(descr, player, what, perms, argv[0]); obj {
	case UNKNOWN:
		ABORT_MPI("NAME", "Match failed.")
	case PERMDENIED:
		ABORT_MPI("NAME", "Permission denied.")
	case NOTHING:
		buf = "#NOTHING#"
	case AMBIGUOUS:
		buf = "#AMBIGUOUS#"
	case HOME:
		buf = "#HOME#"
	default:
		buf = DB.Fetch(obj).name
	}
	return
}

func mfn_sublist(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if len(argv) > 1 {
		start := strconv.Atoi(argv[1])
		end := start
		if len(argv) > 2 {
			end = strconv.Atoi(argv[2])
		}

		sep := MPI_LISTSEP
		if len(argv) > 3 {
			if argv[3] == "" {
				ABORT_MPI("SUBLIST", "Can't use null seperator string.")
			} else {
				sep := argv[3]
			}
		}

		items:= strings.Split(argv[0], sep)
		if l := len(items); l > 0 && start != 0 && end != 0 {
			if start > l {
				start = l
			}
			if start < 0 {
				start += l + 1
			}
			if start < 1 {
				start = 1
			}
			if end > l {
				end = l
			}
			if end < 0 {
				end += l + 1
			}
			if end < 1 {
				end = 1
			}

			incr := 1
			if end < start {
				incr = -1
			}

			results := make([]string, 0, end - start + 1)
			switch incr {
			case 1:
				for i := start; i <= end; i++ {
					results = append(results, items[i])
				}
			case -1
				for i := start; i >= end; i-- {
					results = append(results, items[i])
				}
			}
			r = strings.Join(results, sep)
		}
	} else {
		r = argv[0]
	}
	return
}

func mfn_lrand(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if len(argv) > 1 && argv[1] == "" {
		ABORT_MPI("LRAND", "Can't use null seperator string.")
	}
	items := strings.Split(argv[0], argv[1])
	if l := len(items); l > 0 {
		r = items[rand.Intn(l) - 1]
	}
	return
}

func mfn_count(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	if len(argv) > 1 && argv[1] == "" {
		ABORT_MPI("COUNT", "Can't use null seperator string.")
	}
	return fmt.Sprint(strings.Count(argv[0], argv[1]))
}

func mfn_with(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	varname := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
	CHECKRETURN(varname, "WITH", "arg 1")
	value := mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
	CHECKRETURN(value, "WITH", "arg 2")
	new_mvalues(&MPI_VARIABLES, varname)
	set_mvalue(MPI_VARIABLES, varname, value)
	for i, v := range argv[2:] {
		if r = mesg_parse(descr, player, what, perms, v, mesgtyp); r == "" {
			notify(player, fmt.Sprintf("%s %cWITH%c (arg %d)", get_mvalue(MPI_VARIABLES, "how"), MFUN_LEADCHAR, MFUN_ARGEND, i))
			return
		}
	}
	drop_mvalues(&MPI_VARIABLES, varname)
	return
}

func mfn_fold(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	var1 := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
	CHECKRETURN(var1, "FOLD", "arg 1")

	var2 := mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
	CHECKRETURN(var2, "FOLD", "arg 2")
	new_mvalues(&MPI_VARIABLES, var1, var2)

	var sepin string
	if len(argv) > 4 {
		sepin = mesg_parse(descr, player, what, perms, argv[4], mesgtyp)
		CHECKRETURN(sepin, "FOLD", "arg 5")
		if sepin == "" {
			ABORT_MPI("FOLD", "Can't use Null seperator string")
		}
	} else {
		sepin = MPI_LISTSEP
	}

	list := mesg_parse(descr, player, what, perms, argv[2], mesgtyp)
	CHECKRETURN(list, "FOLD", "arg 3")
	items := strings.Split(list, sepin)
	switch len(items) {
	case 0:
		r = mesg_parse(descr, player, what, perms, argv[3], mesgtyp)
		CHECKRETURN(ptr, "FOLD", "arg 4")
		set_mvalue(&MPI_VARIABLES, var1, r)
	case 1:
		set_mvalue(&MPI_VARIABLES, var1, items[0])
		r = mesg_parse(descr, player, what, perms, argv[3], mesgtyp)
		CHECKRETURN(ptr, "FOLD", "arg 4")
		set_mvalue(&MPI_VARIABLES, var1, r)
	default:
		set_mvalue(&MPI_VARIABLES, var1, items[0])
		for i := 1; i < len(items); i++ {
			set_mvalue(&MPI_VARIABLES, var2, items[i])
			r = mesg_parse(descr, player, what, perms, argv[3], mesgtyp)
			CHECKRETURN(ptr, "FOLD", "arg 4")
			set_mvalue(&MPI_VARIABLES, var1, r)
		}
	}
	return
}

func mfn_for(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	varname := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
	CHECKRETURN(varname, "FOR", "arg 1 (varname)")
	new_mvalues(&MPI_VARIABLES, varname)

	tmp := mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
	CHECKRETURN(tmp, "FOR", "arg 2 (start num)");
	start := strconv.Atoi(tmp)

	tmp = mesg_parse(descr, player, what, perms, argv[2], mesgtyp)
	CHECKRETURN(tmp, "FOR", "arg 3 (end num)")
	end := strconv.Atoi(tmp)

	tmp = mesg_parse(descr, player, what, perms, argv[3], mesgtyp)
	CHECKRETURN(tmp, "FOR", "arg 4 (increment)")
	incr := strconv.Atoi(tmp)

	for i := start; (incr >= 0 && i <= end) || (incr < 0 && i >= end); i += incr {
		set_mvalue(MPI_VARIABLES, varname, fmt.Sprint(i))
		r = mesg_parse(descr, player, what, perms, argv[4], mesgtyp)
		CHECKRETURN(r, "FOR", "arg 5 (repeated command)")
	}
	drop_mvalues(&MPI_VARIABLES, varname)
	return
}

func mfn_foreach(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r) {
	varname := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
	CHECKRETURN(ptr, "FOREACH", "arg 1")
	new_mvalues(&MPI_VARIABLES, varname)

	listbuf := mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
	CHECKRETURN(dptr, "FOREACH", "arg 2")

	var sepin string
	if len(argv) > 3 {
		sepin = mesg_parse(descr, player, what, perms, argv[3], mesgtyp)
		CHECKRETURN(sepinbuf, "FILTER", "arg 4")
		if sepinbuf == "" {
			ABORT_MPI("FILTER", "Can't use Null seperator string")
		}
	} else {
		sepin = MPI_LISTSEP
	}

	var items []string
	for _, v := range strings.Split(list, sepin) {
		set_mvalue(MPI_VARIABLES, varname, v)
		r = mesg_parse(descr, player, what, perms, argv[2], mesgtyp)
		CHECKRETURN(ok, "FOREACH", "arg 3")
	}
	drop_mvalues(&MPI_VARIABLES, varname)
	return
}

func mfn_filter(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	varname := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
	CHECKRETURN(ptr, "FILTER", "arg 1")
	new_mvalues(&MPI_VARIABLES, varname)

	list := mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
	CHECKRETURN(list, "FILTER", "arg 2")

	var sepin string
	if len(argv) > 3 {
		sepin = mesg_parse(descr, player, what, perms, argv[3], mesgtyp)
		CHECKRETURN(sepinbuf, "FILTER", "arg 4")
		if sepinbuf == "" {
			ABORT_MPI("FILTER", "Can't use Null seperator string")
		}
	} else {
		sepin = MPI_LISTSEP
	}

	var sepout string
	if len(argv) > 4 {
		sepout = mesg_parse(descr, player, what, perms, argv[4], mesgtyp)
		CHECKRETURN(sepoutbuf, "FILTER", "arg 5")
	} else {
		sepout = sepin
	}

	var items []string
	for _, v := range strings.Split(list, sepin) {
		set_mvalue(MPI_VARIABLES, varname, v)
		ok := mesg_parse(descr, player, what, perms, argv[2], mesgtyp)
		CHECKRETURN(ok, "FILTER", "arg 3")
		if truestr(ok) {
			items = append(items, v)
		}
	}
	drop_mvalues(&MPI_VARIABLES, varname)
	return strings.Join(items, sepout)
}

func mfn_lremove(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	llist := strings.Split(argv[0], MPI_LISTSEP)
	rlist := strings.Split(argv[1], MPI_LISTSEP)
	return strings.Join(mpi_list_remove(llist, rlist), MPI_LISTSEP)
}

func mfn_lcommon(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	llist := strings.Split(argv[0], MPI_LISTSEP)
	rlist := strings.Split(argv[1], MPI_LISTSEP)
	return strings.Join(mpi_list_common(llist, rlist), MPI_LISTSEP)
}

func mfn_lunion(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	llist := strings.Split(argv[0], MPI_LISTSEP)
	rlist := strings.Split(argv[1], MPI_LISTSEP)
	return strings.Join(mpi_list_union(llist, rlist), MPI_LISTSEP)
}

func mfn_lsort(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	switch len(argv) {
	case 1:
		list := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
		CHECKRETURN(ptr, "LSORT", "arg 1")

		litem := strings.Split(list, MPI_LISTSEP)
		for i := 0; i < len(litem); i++ {
			for j := i + 1; j < len(litem); j++ {
				if alphanum_compare(litem[i], litem[j]) > 0 {
					litem[i], litem[j] = litem[j], litem[i]
				}
			}
		}
		buf = strings.Join(litem, MPI_LISTSEP)
	case 4:
		list := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
		CHECKRETURN(ptr, "LSORT", "arg 1")

		lvar := mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
		CHECKRETURN(lvar, "LSORT", "arg 2")

		rvar := mesg_parse(descr, player, what, perms, argv[2], mesgtyp)
		CHECKRETURN(rvar, "LSORT", "arg 3")
		new_mvalues(&MPI_VARIABLES, lvar, rvar)

		litem := strings.Split(list, MPI_LISTSEP)
		for i := 0; i < len(litem); i++ {
			for j := i + 1; j < len(litem); j++ {
				set_mvalue(MPI_VARIABLES, lvar, litem[i])
				set_mvalue(MPI_VARIABLES, rvar, litem[j])
				r := mesg_parse(descr, player, what, perms, argv[3], mesgtyp)
				CHECKRETURN(r, "LSORT", "arg 4")
				if truestr(r) {
					litem[i], litem[j] = litem[j], litem[i]
				}
			}
		}
		buf = strings.Join(litem, MPI_LISTSEP)
		drop_mvalues(&MPI_VARIABLES, lvar)
	default:
		ABORT_MPI("LSORT", "Takes 1 or 4 arguments.")
	}
	return
}

func mfn_lunique(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	items := strings.Split(argv[0], MPI_LISTSEP)
	if len(items) > 0 {
		p := 0
		m := make(map[string] bool)
		for _, v := range items {
			if !m[v] {
				m[v] = true
				items[p] = v
				p++
			}
		}
		items = items[:p]
	}
	return strings.Join(items, MPI_LISTSEP)
}

func mfn_parse(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	varname := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
	CHECKRETURN(varname, "PARSE", "arg 1")
	new_mvalues(&MPI_VARIABLES, varname)

	list := mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
	CHECKRETURN(list, "PARSE", "arg 2")

	var sepin string
	if len(argv) > 3 {
		sepin = mesg_parse(descr, player, what, perms, argv[3], mesgtyp)
		CHECKRETURN(sepin, "PARSE", "arg 4")
		if sepin == "" {
			ABORT_MPI("PARSE", "Can't use Null seperator string")
		}
	} else {
		sepin = MPI_LISTSEP
	}

	var sepout string
	if len(argv) > 4 {
		sepout = mesg_parse(descr, player, what, perms, argv[4], mesgtyp)
		CHECKRETURN(sepout, "PARSE", "arg 5")
	} else {
		sepout = sepin
	}

	for i, v := range strings.Split(list, sepin) {
		set_mvalue(MPI_VARIABLES, varname, v)
		list[i] := mesg_parse(descr, player, what, perms, argv[2], mesgtyp)
		CHECKRETURN(list[i], "PARSE", "arg 3")		
	}
	drop_mvalues(&MPI_VARIABLES, varname)
	return strings.Join(list, sepout)
}

func mfn_smatch(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	if !smatch(argv[1], argv[0]) {
		return "1"
	} else {
		return "0"
	}
}

func mfn_len(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return fmt.Sprint(len(argv[0]))
}

func mfn_subst(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return string_substitute(argv[0], argv[1], argv[2], buf, BUFFER_LEN)
}

func mfn_awake(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	r = "0"
	switch obj := mesg_ObjectID_local(descr, player, what, perms, argv[0], mesgtyp); obj {
	case PERMDENIED, AMBIGUOUS, UNKNOWN, NOTHING, HOME:
	default:
		switch {
		case IsThing(obj) && DB.Fetch(obj).IsFlagged(ZOMBIE):
			obj = DB.Fetch(obj).Owner
		case !IsPlayer(obj):
		default:
			r = fmt.Sprint(online(obj)) 
		}
	}
	return
}

func mfn_type(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	r = "Bad"
	obj := mesg_ObjectID_local(descr, player, what, perms, argv[0], mesgtyp)
	switch obj {
	case NOTHING, AMBIGUOUS, UNKNOWN:
	case HOME:
		r = "Room"
	case PERMDENIED:
		ABORT_MPI("TYPE", "Permission Denied.")
	default:
		switch obj.(type) {
		case TYPE_PLAYER:
			r = "Player"
		case TYPE_ROOM:
			r = "Room"
		case TYPE_EXIT:
			r = "Exit"
		case TYPE_THING:
			r = "Thing"
		case TYPE_PROGRAM:
			r = "Program"
		}
	}
	return
}

func mfn_istype(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	var obj ObjectID
	if tp_lazy_mpi_istype_perm {
		obj = mesg_ObjectID_raw(descr, player, what, perms, argv[0])
	} else {
		obj = mesg_ObjectID_local(descr, player, what, perms, argv[0], mesgtyp)
	}

	is_type := func(s string) bool {
		if argv[1] == s {
			return "1"
		} else {
			return "0"
		}
	}

	switch obj {
	case NOTHING, AMBIGUOUS, UNKNOWN:
		return is_type("Bad")
	case PERMDENIED:
		if argv[1] == "Bad" {
			return "1"
		} else {
			ABORT_MPI("TYPE", "Permission Denied.")
		}
	case HOME:
		return is_type("Room")
	}

	switch obj.(type) {
	case TYPE_PLAYER:
		return is_type("Player")
	case TYPE_ROOM:
		return is_type("Room")
	case TYPE_EXIT:
		return is_type("Exit")
	case TYPE_THING:
		return is_type("Thing")
	case TYPE_PROGRAM:
		return is_type("Program")
	}
	return is_type("Bad")
}

func mfn_debugif(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	r = mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
	CHECKRETURN(r, "DEBUGIF", "arg 1")
	if truestr(argv[0]) {
		r = mesg_parse(descr, player, what, perms, argv[1], (mesgtyp | MPI_ISDEBUG))
	} else {
		r = mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
	}
	CHECKRETURN(r, "DEBUGIF", "arg 2")
	return
}

func mfn_debug(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	r = mesg_parse(descr, player, what, perms, argv[0], (mesgtyp | MPI_ISDEBUG))
	CHECKRETURN(r, "DEBUG", "arg 1")
	return
}

func mfn_revoke(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r str) {
	r = mesg_parse(descr, player, what, perms, argv[0], (mesgtyp & ~MPI_ISBLESSED))
	CHECKRETURN(r, "REVOKE", "arg 1")
	return
}

func mfn_timing(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	start_time := time.Now()
	r = mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
	CHECKRETURN(r, "TIMING", "arg 1")
	end_time := time.Now()
	notify_nolisten(player, fmt.Sprintf("Time elapsed: %.6f seconds", end_time - start_time), true)
	return
}

func mfn_delay(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	i := strconv.Atoi(argv[0])
	switch {
	case i < 1:
		i = 1
	case i > 31622400:
		ABORT_MPI("DELAY", "Delaying more than a year in MPI is just silly.")
	}
	cmd := get_mvalue(MPI_VARIABLES, "cmd")
	arg := get_mvalue(MPI_VARIABLES, "arg")
	i = add_mpi_event(i, descr, player, DB.Fetch(player).Location, perms, argv[1], cmd, arg, (mesgtyp & MPI_ISLISTENER != 0), (mesgtyp & MPI_ISPRIVATE == 0), (mesgtyp & MPI_ISBLESSED != 0))
	return fmt.Sprint(i)
}

func mfn_kill(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	switch i := strconv.Atoi(argv[0]); {
	case i == 0:
		i = dequeue_prog(perms, 0)
	case i > 0:
		if in_timequeue(i) {
			if !control_process(perms, i) {
				ABORT_MPI("KILL", "Permission denied.")
			}
			i = dequeue_process(i)
		} else {
			i = 0
		}
	default:
		ABORT_MPI("KILL", "Invalid process ID.")
	}
	return fmt.Sprint(i)
}

static int mpi_muf_call_levels = 0

func mfn_muf(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	obj := mesg_ObjectID_raw(descr, player, what, perms, argv[0])
	switch {
	case obj == UNKNOWN:
		ABORT_MPI("MUF", "Match failed.")
	case obj <= NOTHING || TYPEOF(obj) != TYPE_PROGRAM:
		ABORT_MPI("MUF", "Bad program reference.")
	case !DB.Fetch(obj).IsFlagged(LINK_OK) && !controls(perms, obj):
		ABORT_MPI("MUF", "Permission denied.")
	case mesgtyp & (MPI_ISLISTENER | MPI_ISLOCK) && MLevel(obj) < MASTER:
		ABORT_MPI("MUF", "Permission denied.")
	}
	
	if mpi_muf_call_levels++; mpi_muf_call_levels > 18 {
		ABORT_MPI("MUF", "Too many call levels.")
	}

	match_args = argv[1]
	match_cmdname = fmt.Sprintf("%s(MPI)", get_mvalue(MPI_VARIABLES, "how"))
	if tmpfr := interp(descr, player, DB.Fetch(player).Location, obj, perms, PREEMPT, STD_HARDUID, 0); tmpfr {
		rv = interp_loop(player, obj, tmpfr, true)
	}
	mpi_muf_call_levels--

	if rv != nil {
		switch rv := rv.data.(type) {
		case string:
			buf = rv
		case int:
			buf = fmt.Sprint(rv)
		case float64:
			buf = fmt.Sprintf("%.15g", rv)
		case ObjectID:
			buf = ref2str(rv)
		}
	}
	return
}

func mfn_force(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	with_useful_object("FORCE", mesg_ObjectID_raw(descr, player, what, perms, argv[0]), func(obj ObjectID) {
		switch {
		case Typeof(obj) != TYPE_THING && Typeof(obj) != TYPE_PLAYER:
			ABORT_MPI("FORCE", "Bad object reference. (arg1)")
		case argv[1] == "":
			ABORT_MPI("FORCE", "Null command string. (arg2)")
		case !tp_zombies && mesgtyp & MPI_ISBLESSED == 0:
			ABORT_MPI("FORCE", "Permission Denied.")
		}
		if mesgtyp & MPI_ISBLESSED == 0 {
			loc := DB.Fetch(obj).Location
			if IsThing(obj) {
				switch {
				case DB.Fetch(obj).IsFlagged(DARK):
					ABORT_MPI("FORCE", "Cannot force a dark puppet.")
				case DB.Fetch(obj).IsFlagged(ZOMBIE):
					ABORT_MPI("FORCE", "Permission denied.")
				case loc != NOTHING && DB.Fetch(loc).IsFlagged(ZOMBIE) && IsRoom(loc):
					ABORT_MPI("FORCE", "Cannot force a Puppet in a no-puppets room.")
				}
				objname := strings.TrimSpace(DB.Fetch(obj).name)
				if lookup_player(objname) != NOTHING {
					ABORT_MPI("FORCE", "Cannot force a thing named after a player.")
				}
			}
			switch {
			case !DB.Fetch(obj).IsFlagged(XFORCIBLE):
				ABORT_MPI("FORCE", "Permission denied: forced object not @set Xforcible.")
			case !test_lock_false_default(descr, perms, obj, "@/flk"):
				ABORT_MPI("FORCE", "Permission denied: Object not force-locked to trigger.")
			}
		}
		switch {
		case obj == GOD:
			ABORT_MPI("FORCE", "Permission denied: You can't force God.")
		case force_level > tp_max_force_level - 1:
			ABORT_MPI("FORCE", "Permission denied: You can't force recursively.")
		}
		objname := strings.TrimSpace(DB.Fetch(obj).name)
		for i, v := range strings.Split(argv[1], MPI_LISTSEP) {
			if lookup_player(objname) != NOTHING && Typeof(obj) != TYPE_PLAYER {
				ABORT_MPI("FORCE", "Cannot force a thing named after a player. [2]")
			}
			ForceAction(what, func() {
				if objname != "" {
					process_command(ObjectID_first_descr(obj), obj, ptr)
				}
			})
		}
	})
	return
}

func mfn_midstr(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	pos1 := strconv.Atoi(argv[1])
	pos2 := pos1
	if len(argv) > 2 {
		pos2 = strconv.Atoi(argv[2])
	}
	if l := len(argv[0]); pos1 != 0 && pos2 != 0 {
		if pos1 > l {
			pos1 = l
		}
		if pos1 < 0 {
			pos1 += l + 1
		}
		if pos1 < 1 {
			pos1 = 1
		}
		if pos2 > l {
			pos2 = l
		}
		if pos2 < 0 {
			pos2 += l + 1
		}
		if pos2 < 1 {
			pos2 = 1
		}
		if pos2 > pos1 {
			pos1, pos2 = pos2, pos1
		}
		buf = argv[0][pos1:pos2]
	}
	return
}

func mfn_instr(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	if argv[1] == "" {
		ABORT_MPI("INSTR", "Can't search for a null string.")
	}
	return fmt.Sprint(strings.Index(argv[0], argv1) + 1)
}

func mfn_lmember(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	/* {lmember:list,item,delim} */
	var delim string
	if len(argv) < 3 {
		delim = MPI_LISTSEP
	} else {
		delim = argv[2]
	}
	if delim == "" {
		ABORT_MPI("LMEMBER", "List delimiter cannot be a null string.")
	}

	var n int
	for i, v := range strings.Split(delim) {
		if v == argv[1] {
			n = i + 1
				break
		}
	}
	return fmt.Sprint(n)
}

func mfn_tolower(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return strings.ToLower(argv[0])
}

func mfn_toupper(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return strings.ToUpper(argv[0])
}

func mfn_commas(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	switch len(argv) {
	case 1:
		list := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
		CHECKRETURN(ptr, "COMMAS", "arg 1")
		mpi_list_commas(strings.Split(list, MPI_LISTSEP), " and ")
	case 2:
		list := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
		CHECKRETURN(ptr, "COMMAS", "arg 1")
		items := strings.Split(list, MPI_LISTSEP)
		if l := len(items); l != 0 {
			sep := mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
			CHECKRETURN(sep, "COMMAS", "arg 2")
			mpi_list_commas(strings.Split(list, MPI_LISTSEP), sep)
		}
	case 4:
		list := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
		CHECKRETURN(ptr, "COMMAS", "arg 1")
		items := strings.Split(list, MPI_LISTSEP)
		if l := len(items); l != 0 {
			sep := mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
			CHECKRETURN(sep, "COMMAS", "arg 2")
			varname := mesg_parse(descr, player, what, perms, argv[2], mesgtyp)
			CHECKRETURN(varname, "COMMAS", "arg 3")
			new_mvalues(MPI_VARIABLES, varname)

			for i, v := range items {
				set_mvalue(MPI_VARIABLES, varname, v)
				items[i] = mesg_parse(descr, player, what, perms, argv[3], mesgtyp)
				CHECKRETURN(items[i], "COMMAS", "arg 3")
			}
			mpi_list_commas(strings.Split(list, MPI_LISTSEP), sep)
			drop_mvalues(&MPI_VARIABLES, varname)
		}
	default:
		ABORT_MPI("COMMAS", "Takes 1, 2, or 4 arguments.")
	}
	return
}

func mfn_escape(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	r = MFUN_LITCHAR
	for _, v := range argv {
		switch v {
		case '\\':
		case MFUN_LITCHAR:
			r += '\\' + v
		default:
			r += v
		}
	}
	r += MFUN_LITCHAR
	return
}