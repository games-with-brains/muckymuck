package fbmuck

import "strings"

func with_useful_object(caller string, obj ObjectID, f func()) {
	switch obj {
	case UNKNOWN, AMBIGUOUS, NOTHING, HOME:
		ABORT_MPI(caller, "Match failed.")
	case PERMDENIED:
		ABORT_MPI(caller, "Permission denied.")
	default:
		f()
	}
}

/***** Insert MFUNs here *****/

func mfn_func(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	var namebuf, argbuf, defbuf string

	funcname := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
	CHECKRETURN(funcname, "FUNC", "name argument (1)")

	def := argv[len(argv) - 1]
	for i, v := range argv[1:] {
		argbuf = mesg_parse(descr, player, what, perms, v, mesgtyp)
		CHECKRETURN(argbuf, "FUNC", "variable name argument")
		defbuf = fmt.Sprintf("{with:%s,{:%d},%s}", argbuf, i, def)
	}
	set_mvalue(mpi_functions, funcname, defbuf)
	return
}

func mfn_muckname(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return tp_muckname
}

func mfn_version(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return VERSION
}

func mfn_prop(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if len(argv) == 2 {
		what = mesg_ObjectID(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("PROP", what, func(obj ObjectID) {
		var blessed bool
		if r, blessed = safegetprop(player, obj, perms, argv[0], mesgtyp); r == "" {
			ABORT_MPI("PROP", "Failed read.")
		}
	})
	return
}

func mfn_propbang(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if len(argv) == 2 {
		what = mesg_ObjectID(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("PROP!", obj, func(obj ObjectID) {
		var blessed bool
		if r, blessed = safegetprop_strict(player, what, perms, argv[0], mesgtyp); r = "" {
			ABORT_MPI("PROP!", "Failed read.")
		}
	})
	return
}

func mfn_store(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if len(argv) > 2 {
		what = mesg_ObjectID_strict(descr, player, what, perms, argv[2], mesgtyp)
	}
	with_useful_object("STORE", obj, func(obj ObjectID) {
		if !safeputprop(what, perms, argv[1], argv[0], mesgtyp) {
			ABORT_MPI("BLESS", "Permission denied.")
		}
	})
	return argv[0]
}

func mfn_bless(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if len(argv) > 1 {
		what = mesg_ObjectID_strict(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("BLESS", obj, func(obj ObjectID) {
		if !safeblessprop(obj, perms, argv[0], mesgtyp, 1) {
			ABORT_MPI("BLESS", "Permission denied.")
		}
	})
	return
}

func mfn_unbless(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if len(argv) > 1 {
		what = mesg_ObjectID_strict(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("UNBLESS", what, func(obj ObjectID) {
		if !safeblessprop(obj, perms, argv[0], mesgtyp, 0) {
			ABORT_MPI("UNBLESS", "Permission denied.")
		}
	})
	return
}

func mfn_delprop(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if len(argv) > 1 {
		what = mesg_ObjectID_strict(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("DELPROP", what, func(obj ObjectID) {
		if !safeputprop(obj, perms, argv[0], nil, mesgtyp) {
			ABORT_MPI("DELPROP", "Permission denied.")
		}
	})
	return
}

func mfn_exec(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	obj := what
	if len(argv) == 2 {
		obj = mesg_ObjectID(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("EXEC", obj, func(obj ObjectID) {
		pname := strings.TrimLeft(argv[0], PROPDIR_DELIMITER)
		var blessed bool
		if r, blessed = safegetprop(player, obj, perms, pname, mesgtyp); r = "" {
			ABORT_MPI("EXEC", "Failed read.")
		}
		if blessed {
			mesgtyp |= MPI_ISBLESSED
		} else {
			mesgtyp &= ~MPI_ISBLESSED
		}
		trg := what
		switch {
		case Prop_ReadOnly(pname), Prop_Private(pname), Prop_SeeOnly(pname), Prop_Hidden(pname):
			trg = obj
		}
		r = mesg_parse(descr, player, obj, trg, r, mesgtyp)
		CHECKRETURN(r, "EXEC", "propval")
	})
	return
}

func mfn_execbang(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	obj := what
	if len(argv) == 2 {
		obj = mesg_ObjectID(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("EXEC!", obj, func(obj ObjectID) {
		var blessed bool
		pname = strings.TrimLeft(pname, PROPDIR_DELIMITER)
		if r, blessed = safegetprop_strict(player, obj, perms, argv[0], mesgtyp); r == "" {
			ABORT_MPI("EXEC!", "Failed read.")
		}
		if blessed {
			mesgtyp |= MPI_ISBLESSED
		} else {
			mesgtyp &= ~MPI_ISBLESSED
		}
		trg := what
		switch {
		case Prop_ReadOnly(pname), Prop_Private(pname), Prop_SeeOnly(pname), Prop_Hidden(pname):
			trg = obj
		}
		r = mesg_parse(descr, player, obj, trg, r, mesgtyp)
		CHECKRETURN(r, "EXEC!", "propval")
	})
	return
}

func mfn_index(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	obj := what
	if len(argv) == 2 {
		obj = mesg_ObjectID(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("INDEX", obj, func(obj ObjectID) {
		tmpobj := obj
		var blessed bool
		if r, blessed = safegetprop(player, obj, perms, argv[0], mesgtyp); r != "" {
			obj = tmpobj
			if r, blessed = safegetprop(player, obj, perms, ptr, mesgtyp); blessed {
				mesgtyp |= MPI_ISBLESSED
			} else {
				mesgtyp &= ~MPI_ISBLESSED
			}
			trg := what
			switch {
			case Prop_ReadOnly(r), Prop_Private(r), Prop_SeeOnly(r), Prop_Hidden(r):
				trg = obj
			}
			r = mesg_parse(descr, player, obj, trg, r, mesgtyp)
			CHECKRETURN(ptr, "INDEX", "listval")
		}
	})
	return
}

func mfn_indexbang(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	obj := what
	if len(argv) == 2 {
		obj = mesg_ObjectID(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("INDEX!", obj, func(obj ObjectID) {
		tmpobj := obj
		var blessed bool
		if r, blessed = safegetprop_strict(player, obj, perms, argv[0], mesgtyp); r != "" {
			obj = tmpobj
			r, blessed = safegetprop_strict(player, obj, perms, r, mesgtyp)
			if blessed {
				mesgtyp |= MPI_ISBLESSED
			} else {
				mesgtyp &= ~MPI_ISBLESSED
			}
			trg := what
			switch {
			case Prop_ReadOnly(r), Prop_Private(r), Prop_SeeOnly(r), Prop_Hidden(r):
				trg = obj
			}
			r = mesg_parse(descr, player, obj, trg, r, mesgtyp)
			CHECKRETURN(ptr, "INDEX!", "listval");
		}
	})
	return
}

func mfn_propdir(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	obj := what
	if len(argv) == 2 {
		obj = mesg_ObjectID(descr, player, what, perms, argv, mesgtyp)
	}
	with_useful_object("PROPDIR", obj, func(obj ObjectID) {
		if is_propdir(obj, argv[0]) {
			r = "1"
		} else {
			r = "0"
		}
	})
	return
}

func mfn_listprops(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	obj := what
	if len(argv) > 1 {
		obj = mesg_ObjectID(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("LISTPROPS", obj, func(obj ObjectID) {
		var pattern string
		if len(argv) > 2 {
			pattern = argv[2]
		}
		pname := argv[0]
		if pname[len(pname) - 1] != PROPDIR_DELIMITER {
			pname += PROPDIR_DELIMITER
		}
		var items []string
		for p := next_prop_name(obj, pname); p != ""; p = next_prop_name(obj, p) {
			flag := true
			switch {
			case Prop_System(p):
				flag = false
			case !(mesgtyp & MPI_ISBLESSED):
				if Prop_Hidden(p) {
					flag = false
				}
				if Prop_Private(p) && DB.Fetch(what).Owner != DB.Fetch(obj).Owner {
					flag = false
				}
				if obj != player && DB.Fetch(obj).Owner != DB.Fetch(what).Owner {
					flag = false
				}
			}
			if flag && pattern != "" {
				i := strrchr(p, PROPDIR_DELIMITER)
				if pattern != p[i + 1:] {
					flag = false
				}
			}
			if flag {
				items = append(items, p)
			}
		}
		r = strings.Join(items, MPI_LISTSEP)
	})
	return
}

func mfn_concat(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	obj := what
	if len(argv) == 2 {
		obj = mesg_ObjectID(descr, player, what, perms, argv[1], mesgtyp);
	}
	with_useful_object("CONCAT", obj, func(obj ObjectID) {
		var blessed bool
		r = get_concat_list(player, what, perms, obj, argv[0], r, BUFFER_LEN, 1, mesgtyp, &blessed)
	})
	return
}

func mfn_select(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	char propname[BUFFER_LEN]
	pname := argv[1]
	obj := what
	if len(argv) == 3 {
		obj = mesg_ObjectID(descr, player, what, perms, argv[2], mesgtyp)
	}
	with_useful_object("SELECT", obj, func(obj ObjectID) {
		//	Search contiguously for a bit, looking for a best match. This allows fast hits on LARGE lists.
		targval := strconv.Atoi(argv[0])
	   	i := targval
	   	var blessed bool
	   	for limit := 17; limit > -1 && i > -1 && r == ""; limit-- {
	   		r, blessed = get_list_item(player, obj, perms, pname, i, mesgtyp)
	   		i--
	   	}
	   	if r == "" {
	   		//	If we didn't find it before, search only existing props. This gets fast hits on very SPARSE lists.
	   		/	First, normalize the base propname

	   		var origprop string
	   		for in := argv[1]; in != ""; {
	   			origprop = append(origprop, PROPDIR_DELIMITER)
	   			in = strings.TrimLeft(in, PROPDIR_DELIMITER)
	   			for i, v := range in {
	   				if v == PROPDIR_DELIMITER {
	   					in = in[i + 1:]
	   					break
	   				}
	   				origprop = append(origprop, v)
	   			}
	   		}

	   		var bestname string
	   		var bestval int
	   		var bestobj ObjectID
	   		baselen := len(origprop)
	   		for ; obj != NOTHING; obj = getparent(obj) {
	   			pname = next_prop_name(obj, origprop)
	   			for pname != "" && strings.Prefix(pname, origprop) {
	   				r = pname[baselen:]
	   				if r[0] == NUMBER_TOKEN {
	   					r = r[1:]
	   				}

	   				if r == ""  && is_propdir(obj, pname) {
	   					sublen := len(pname)
	   					progname2 := pname
	   					pname2 := propname2
	   					propname2 += PROPDIR_DELIMITER

	   					pname2 = next_prop_name(obj, pname2)
	   					for pname2 != "" {
	      					r = pname2[sublen:]
	      					if unicode.IsNumber(r) {
	      						i = strconv.Atoi(r)
	      						if bestval < i && i <= targval {
	      							bestval = i
	      							bestobj = obj
	      							strcpyn(bestname, sizeof(bestname), pname2)
	      						}
	      					}
	   						pname2 = next_prop_name(obj, pname2)
	   					}
	   				}
					
	   				//	C STRING MANIPULATION WITH USUAL OBFUSCATING POINTER MANIPULATION

	   				r = pname + baselen
	   				if unicode.IsNumber(r) {
	   					i = strconv.Atoi(r)
	   					if bestval < i && i <= targval {
	   						bestval = i
	   						bestobj = obj
	   						strcpyn(bestname, sizeof(bestname), pname)
	   					}
	   				}
	   				pname = next_prop_name(obj, pname)
	   			}
	   		}
	
	   		if bestname != "" {
	      		r, blessed = safegetprop_strict(player, bestobj, perms, bestname, mesgtyp)
	      		if r == "" {
	      			ABORT_MPI("SELECT", "Failed property read.")
				}
	      	} else {
	   			r = ""
	   		}
	   	}
	})
	return
}

func mfn_list(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	obj := what
	if len(argv) == 2 {
		obj = mesg_ObjectID(descr, player, what, perms, argv[1], mesgtyp);
	}
	with_useful_object("LIST", obj, func(obj ObjectID) {
		var blessed bool
		r = get_concat_list(player, what, perms, obj, argv[0], buf, BUFFER_LEN, 0, mesgtyp, &blessed)
	})
	return
}

func mfn_lexec(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	obj := what
	if len(argv) == 2 {
		obj = mesg_ObjectID(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("LEXEC", obj, func(obj ObjectID) {
		pname := strings.TrimLeft(argv[0], PROPDIR_DELIMITER)
		var blessed bool
		r = get_concat_list(player, what, perms, obj, pname, buf, BUFFER_LEN, 2, mesgtyp, &blessed);
		if blessed {
			mesgtyp |= MPI_ISBLESSED
		} else {
			mesgtyp &= ~MPI_ISBLESSED
		}
		trg := what
		if Prop_ReadOnly(pname) || Prop_Private(pname) || Prop_SeeOnly(pname) || Prop_Hidden(pname) {
			trg = obj
		}
		buf = mesg_parse(descr, player, obj, trg, r, mesgtyp)
		CHECKRETURN(buf, "LEXEC", "listval")
	})
	return
}

func mfn_rand(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	pname := argv[0]
	obj := what
	if len(argv) == 2 {
		obj = mesg_ObjectID(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("RAND", obj, func(obj ObjectID) {
		var blessed bool
		num := get_list_count(what, obj, perms, pname, mesgtyp, &blessed)
		if num == 0 {
			ABORT_MPI("RAND", "Failed list read.")
		}
		if r, blessed = get_list_item(what, obj, perms, pname, (((RANDOM() / 256) % num) + 1), mesgtyp); r == "" {
			ABORT_MPI("RAND", "Failed list read.")
		}
		trg := what
		if blessed {
			mesgtyp |= MPI_ISBLESSED
		} else {
			mesgtyp &= ~MPI_ISBLESSED
		}
		if Prop_ReadOnly(r) || Prop_Private(r) || Prop_SeeOnly(r) || Prop_Hidden(r) {
			trg = obj
		}
		r = mesg_parse(descr, player, obj, trg, r, mesgtyp)
		CHECKRETURN(r, "RAND", "listval")
	})
	return
}

func mfn_timesub(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	period := atoi(argv[0])
	offset := atoi(argv[1])
	pname := argv[2]
	obj := what
	if len(argv) == 4 {
		obj = mesg_ObjectID(descr, player, what, perms, argv[3], mesgtyp)
	}
	with_useful_object("TIMESUB", obj, func(obj ObjectID) {
		var blessed bool
		num := get_list_count(what, obj, perms, pname, mesgtyp, &blessed)
		switch {
		case num == 0:
			ABORT_MPI("TIMESUB", "Failed list read.")
		case period < 1:
			ABORT_MPI("TIMESUB", "Time period too short.")
		}
		if offset = int(((long(time(NULL)) + offset) % period) * num) / period; offset < 0 {
			offset = -offset
		}
		if r, blessed = get_list_item(what, obj, perms, pname, offset + 1, mesgtyp); r == "" {
			ABORT_MPI("TIMESUB", "Failed list read.")
		}
		trg := what
		if blessed {
			mesgtyp |= MPI_ISBLESSED
		} else {
			mesgtyp &= ~MPI_ISBLESSED
		}
		if Prop_ReadOnly(r) || Prop_Private(r) || Prop_SeeOnly(r) || Prop_Hidden(r) {
			trg = obj
		}
		r = mesg_parse(descr, player, obj, trg, r, mesgtyp)
		CHECKRETURN(r, "TIMESUB", "listval")
	})
	return
}

func mfn_nl(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return MPI_LISTSEP
}

func mfn_lit(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return strings.Join(argv, ",")
}

func mfn_eval(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	r = mesg_parse(descr, player, what, perms, strings.Join(argv, ","), (mesgtyp & ~MPI_ISBLESSED))
	CHECKRETURN(ptr, "EVAL", "arg 1");
	return
}

func mfn_evalbang(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	r = mesg_parse(descr, player, what, perms, strings.Join(argv, ","), mesgtyp)
	CHECKRETURN(r, "EVAL!", "arg 1")
	return
}

func mfn_strip(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	r = strings.TrimLeftFunc(argv[0], unicode.IsSpace)
	for i, v := range argv[1:] {
		r = append(r, ",", v)
	}
	return strings.TrimRightFunc(r, unicode.IsSpace)
}

func mfn_mklist(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	return strings.Join(argv, MPI_LISTSEP)
}

func mfn_pronouns(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	obj := player
	if len(argv) > 1 {
		obj = mesg_ObjectID_local(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("PRONOUNS", obj, func(obj ObjectID) {
		r = pronoun_substitute(descr, obj, argv[0])
	})
	return
}

func mfn_ontime(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	r = "-1"
	switch obj := mesg_ObjectID_raw(descr, player, what, perms, argv[0]); obj {
	case UNKNOWN, AMBIGUOUS, NOTHING, HOME:
	case PERMDENIED:
		ABORT_MPI("ONTIME", "Permission denied.")
	default:
		if Typeof(obj) != TYPE_PLAYER {
			obj = DB.Fetch(obj).Owner
		}
		if conn := least_idle_player_descr(obj); conn != 0 {
			r = fmt.Sprint(pontime(conn))
		}
	})
	return
}

func mfn_idle(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	r = "-1"
	switch obj := mesg_ObjectID_raw(descr, player, what, perms, argv[0]); obj {
	case PERMDENIED:
		ABORT_MPI("IDLE", "Permission denied.")
	case UNKNOWN, AMBIGUOUS, NOTHING, HOME:
	default:
		if Typeof(obj) != TYPE_PLAYER {
			obj = DB.Fetch(obj).Owner
		}
		if conn := least_idle_player_descr(obj); conn != 0 {
			r = fmt.Sprint(pidle(conn))
		}
	}
	return
}

func mfn_online(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if mesgtyp & MPI_ISBLESSED == 1 {
		for count := current_descr_count; count > 0; count-- {
			if r != "" {
				r = append(r, MPI_LISTSEP)
			}
			r = append(r, ref2str(pObjectID(count)))
		}
	} else {
		ABORT_MPI("ONLINE", "Permission denied.")
	}
	return
}

func msg_compare(s1, s2 string) (r int) {
	if s1 != "" && s2 != "" && unicode.IsNumber(s1) && unicode.IsNumber(s2) {
		r = strconv.Atoi(s1) - strconv.Atoi(s2)
	} else {
		r = strings.Compare(s1, s2)
	}
	return
}

func mfn_contains(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("CONTAINS (1)", mesg_ObjectID_local(descr, player, what, perms, argv[0], mesgtyp), func(obj1 ObjectID) {
		obj := player
		if len(argv) > 1 {
			obj = mesg_ObjectID_raw(descr, player, what, perms, argv[1])
		}
		with_useful_object("CONTAINS (2)", obj, func(obj2 ObjectID) {
			for obj2 != NOTHING && obj2 != obj1 {
				obj2 = DB.Fetch(obj2).Location
			}
			if obj1 == obj2 {
				r = "1"
			} else {
				r = "0"
			}
		})
	})
	return
}

func mfn_holds(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("HOLDS (1)", mesg_ObjectID_raw(descr, player, what, perms, argv[0]), func(obj1 ObjectID) {
		obj := player
		if len(argv) > 1 {
			obj = mesg_ObjectID_local(descr, player, what, perms, argv[1], mesgtyp)
		}
		with_useful_object("HOLDS (2)", obj, func(obj2 ObjectID) {
			if obj2 == DB.Fetch(obj1).Location {
				r = "1"
			} else {
				r = "0"
			}
		})
	})
	return
}

func mfn_dbeq(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	obj1 := mesg_ObjectID_raw(descr, player, what, perms, argv[0])
	obj2 := mesg_ObjectID_raw(descr, player, what, perms, argv[1])
	switch {
	case obj1 == UNKNOWN || obj1 == PERMDENIED:
		ABORT_MPI("DBEQ", "Match failed (1).")
	case obj2 == UNKNOWN || obj2 == PERMDENIED:
		ABORT_MPI("DBEQ", "Match failed (2).")
	}
	if obj1 == obj2 {
		r = "1"
	} else {
		r = "0"
	}
	return
}

func mfn_ne(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if msg_compare(argv[0], argv[1]) == 0 {
		r = "0"
	} else {
		r = "1"
	}
	return
}

func mfn_eq(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if msg_compare(argv[0], argv[1]) == 0 {
		r = "1"
	} else {
		r = "0"
	}
	return
}

func mfn_gt(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if msg_compare(argv[0], argv[1]) > 0 {
		r = "1"
	} else {
		r = "0"
	}
	return
}

func mfn_lt(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if msg_compare(argv[0], argv[1]) < 0 {
		r = "1"
	} else {
		r = "0"
	}
	return
}

func mfn_ge(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if msg_compare(argv[0], argv[1]) >= 0 {
		r = "1"
	} else {
		r = "0"
	}
	return
}

func mfn_le(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if msg_compare(argv[0], argv[1]) <= 0 {
		r = "1"
	} else {
		r = "0"
	}
	return
}

func mfn_min(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if msg_compare(argv[0], argv[1]) <= 0 {
		r = argv[0]
	} else {
		r = argv[1]
	}
	return
}

func mfn_max(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if msg_compare(argv[0], argv[1]) >= 0 {
		r = argv[0]
	} else {
		r = argv[1]
	}
	return
}

func mfn_isnum(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	if argv[0] != nil && unicode.IsNumber(argv[0]) {
		r = "1"
	} else {
		r = "0"
	}
	return
}

func mfn_isObjectID(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	buf := strings.TrimLeftFunc(argv[0], unicode.IsSpace)
	r = "0"
	if buf[0] == NUMBER_TOKEN && unicode.IsNumber(buf[1]) {
		r = MUFBool(strconv.Atoi(ptr).IsValid())
	}
	return
}

func mfn_inc(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	mpi_use_variable(argv, func(i, v int) {
		r = fmt.Sprint(i + x)
	})
	return
}

func mfn_dec(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	mpi_use_variable(argv, func(i, v int) {
		r = fmt.Sprint(i - x)
	})
	return
}

func mfn_do_int_maths(argv MPIArgs, f func(i, v int) int) string {
	i := strconv.Atoi(argv[0])
	for _, v := range argv[1:] {
		i = f(i, strconv.Atoi(v))
	}
	return fmt.Sprint(i)
}

func mfn_add(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return mfn_do_int_maths(argv, func(i, v int) int {
		return i + v
	})
}

func mfn_subt(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return mfn_do_int_maths(argv, func(i, v int) int {
		return i - v
	})
}

func mfn_mult(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return mfn_do_int_maths(argv, func(i, v int) int {
		return i * v
	})
}

func mfn_div(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return mfn_do_int_maths(argv, func(i, v int) (r int) {
		if v == 0 {
			r = 0
		} else {
			r = i / v
		}
		return
	})
}

func mfn_mod(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return mfn_do_int_maths(argv, func(i, v int) (r int) {
		if v == 0 {
			r = 0
		} else {
			r = i % v
		}
		return
	})
}

func mfn_abs(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	val := strconv.Atoi(argv[0])
	if val < 0 {
		val = -val;
	}
	return fmt.Sprint(val)
}

func mfn_sign(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	switch val := strconv.Atoi(argv[0]) {
	case val < 0:
		return "-1"
	case val > 0:
		return "1"
	}
	return "0"
}

func mfn_dist(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	int a, b, c;
	int a2, b2, c2;
	double result;

	a2 = b2 = c = c2 = 0
	a = strconv.Atoi(argv[0])
	b = strconv.Atoi(argv[1])
	switch len(argv) {
	case 2:
	case 3:
		c = strconv.Atoi(argv[2])
	case 4:
		a2 = strconv.Atoi(argv[2])
		b2 = strconv.Atoi(argv[3])
	case 6:
		c = strconv.Atoi(argv[2])
		a2 = strconv.Atoi(argv[3])
		b2 = strconv.Atoi(argv[4])
		c2 = strconv.Atoi(argv[5])
	default:
		ABORT_MPI("DIST", "Takes 2,3,4, or 6 arguments.")
	}
	a -= a2
	b -= b2
	c -= c2
	result = sqrt((double) (a * a) + (double) (b * b) + (double) (c * c))
	return fmt.Sprintf("%.0f", floor(result + 0.5))
}

func mfn_not(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	if truestr(argv[0]) {
		return "0"
	} else {
		return "1"
	}
}

func mfn_or(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	for i, v := range argv {
		buf := mesg_parse(descr, player, what, perms, argv[i], mesgtyp)
		CHECKRETURN(buf, "OR", fmt.Sprintf("arg %d", i + 1))
		if truestr(buf) {
			return "1"
		}
	}
	return "0"
}

func mfn_xor(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	switch {
	case truestr(argv[0]) && !truestr(argv[1]), !truestr(argv[0]) && truestr(argv[1]):
		return "1"
	}
	return "0"
}

func mfn_and(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	for i, v := range argv {
		r = mesg_parse(descr, player, what, perms, v, mesgtyp)
		CHECKRETURN(r, "AND", fmt.Sprintf("arg %d", i + 1))
		if !truestr(r) {
			return "0"
		}
	}
	return "1"
}

func mfn_dice(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	var offset, total int
	num := 1
	sides := 1

	if len(argv) >= 3 {
		offset = strconv.Atoi(argv[2])
	}
	if len(argv) >= 2 {
		num = strconv.Atoi(argv[1])
	}
	sides = strconv.Atoi(argv[0])
	if num > 8888 {
		ABORT_MPI("DICE", "Too many dice!")
	}
	if sides == 0 {
		return "0"
	}
	for ; num > 0; num-- {
		total += (((RANDOM() / 256) % sides) + 1)
	}
	return fmt.Sprint(total + offset)
}

func mfn_default(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	buf = mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
	CHECKRETURN(buf, "DEFAULT", "arg 1")
	if buf != "" && truestr(buf) {
	} else {
		buf = mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
		CHECKRETURN(buf, "DEFAULT", "arg 2")
	}
	return
}

func mfn_if(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	var fbr string
	if len(argv) == 3 {
		fbr = argv[2]
	}
	buf = mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
	CHECKRETURN(buf, "IF", "arg 1")
	if buf != "" && truestr(buf) {
		buf = mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
		CHECKRETURN(buf, "IF", "arg 2");
	} else if fbr != "" {
		buf = mesg_parse(descr, player, what, perms, fbr, mesgtyp)
		CHECKRETURN(buf, "IF", "arg 3")
	}
	return
}

func mfn_while(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	for {
		buf2 := mesg_parse(descr, player, what, perms, argv[0], mesgtyp)
		CHECKRETURN(buf2, "WHILE", "arg 1");
		if !truestr(buf2) {
			break
		}
		buf = mesg_parse(descr, player, what, perms, argv[1], mesgtyp)
		CHECKRETURN(ptr, "WHILE", "arg 2")
	}
	return
}

func mfn_null(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return ""
}

func mfn_tzoffset(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	return fmt.Sprint(get_tz_offset())
}

func mfn_time(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	time_t lt;
	struct tm *tm;

	lt = time((time_t*) NULL);
	if len(argv) == 1 {
		lt += (3600 * atoi(argv[0]));
		lt += get_tz_offset();
	}
	tm = localtime(&lt);
	return format_time("%T", tm)
}

func mfn_date(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	time_t lt;
	struct tm *tm;

	lt = time((time_t*) NULL);
	if len(argv) == 1 {
		lt += (3600 * strconv.Atoi(argv[0]));
		lt += get_tz_offset();
	}
	tm = localtime(&lt);
	return format_time("%D", tm)
}

func mfn_ftime(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	time_t lt;
	struct tm *tm;

	if len(argv) == 3 {
		lt = atol(argv[2]);
	} else {
		time(&lt);
	}
	if len(argv) > 1 && argv[1]) {
		int offval = atoi(argv[1]);
		if (offval < 25 && offval > -25) {
			lt += 3600 * offval;
		} else {
			lt -= offval;
		}
		lt += get_tz_offset();
	}
	tm = localtime(&lt);
	return format_time(argv[0], tm)
}

func mfn_convtime(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	struct tm otm;
	int mo, dy, yr, hr, mn, sc;

	yr = 70;
	mo = dy = 1;
	hr = mn = sc = 0;
	if (sscanf(argv[0], "%d:%d:%d %d/%d/%d", &hr, &mn, &sc, &mo, &dy, &yr) != 6)
		ABORT_MPI("CONVTIME", "Needs HH:MM:SS MO/DY/YR time string format.");
	if (hr < 0 || hr > 23)
		ABORT_MPI("CONVTIME", "Bad Hour");
	if (mn < 0 || mn > 59)
		ABORT_MPI("CONVTIME", "Bad Minute");
	if (sc < 0 || sc > 59)
		ABORT_MPI("CONVTIME", "Bad Second");
	if (yr < 0 || yr > 99)
		ABORT_MPI("CONVTIME", "Bad Year");
	if (mo < 1 || mo > 12)
		ABORT_MPI("CONVTIME", "Bad Month");
	if (dy < 1 || dy > 31)
		ABORT_MPI("CONVTIME", "Bad Day");
	otm.tm_mon = mo - 1;
	otm.tm_mday = dy;
	otm.tm_hour = hr;
	otm.tm_min = mn;
	otm.tm_sec = sc;
	otm.tm_year = (yr >= 70) ? yr : (yr + 100);
	buf = fmt.Sprint(mktime(&otm))
	return buf;
}

func mfn_ltimestr(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	int tm = strconv.Atol(argv[0])
	int yr, mm, wk, dy, hr, mn;
	char buf2[BUFFER_LEN];

	yr = mm = wk = dy = hr = mn = 0;
	if (tm >= 31556736) {
		yr = tm / 31556736;	        /* Years */
		tm %= 31556736;
	}
	if (tm >= 2621376) {
		mm = tm / 2621376;	        /* Months */
		tm %= 2621376;
	}
	if (tm >= 604800) {
		wk = tm / 604800;	        /* Weeks */
		tm %= 604800;
	}
	if (tm >= 86400) {
		dy = tm / 86400;		/* Days */
		tm %= 86400;
	}
	if (tm >= 3600) {
		hr = tm / 3600;			/* Hours */
		tm %= 3600;
	}
	if (tm >= 60) {
		mn = tm / 60;			/* Minutes */
		tm %= 60;			/* Seconds */
	}

	*buf = '\0';
	if (yr) {
		buf = fmt.Sprintf("%d year%s", yr, (yr == 1) ? "" : "s")
	}
	if (mm) {
		buf2 = fmt.Sprintf("%d month%s", mm, (mm == 1) ? "" : "s")
		if (*buf) {
			strcatn(buf, BUFFER_LEN, ", ");
		}
		strcatn(buf, BUFFER_LEN, buf2);
	}
	if (wk) {
		buf2 = fmt.Sprintf("%d week%s", wk, (wk == 1) ? "" : "s")
		if (*buf) {
			strcatn(buf, BUFFER_LEN, ", ");
		}
		strcatn(buf, BUFFER_LEN, buf2);
	}
	if (dy) {
		buf2 = fmt.Sprintf("%d day%s", dy, (dy == 1) ? "" : "s")
		if (*buf) {
			strcatn(buf, BUFFER_LEN, ", ");
		}
		strcatn(buf, BUFFER_LEN, buf2);
	}
	if (hr) {
		buf2 = fmt.Sprintf("%d hour%s", hr, (hr == 1) ? "" : "s")
		if (*buf) {
			strcatn(buf, BUFFER_LEN, ", ");
		}
		strcatn(buf, BUFFER_LEN, buf2);
	}
	if (mn) {
		buf2 = fmt.Sprintf("%d min%s", mn, (mn == 1) ? "" : "s")
		if (*buf) {
			strcatn(buf, BUFFER_LEN, ", ");
		}
		strcatn(buf, BUFFER_LEN, buf2);
	}
	if (tm || !*buf) {
		buf2 = fmt.Sprintf("%d sec%s", tm, (tm == 1) ? "" : "s")
		if (*buf) {
			strcatn(buf, BUFFER_LEN, ", ");
		}
		strcatn(buf, BUFFER_LEN, buf2);
	}
	return buf;
}

func mfn_timestr(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	int tm = strconv.Atol(argv[0])
	int dy, hr, mn;

	dy = hr = mn = 0;
	if (tm >= 86400) {
		dy = tm / 86400;		/* Days */
		tm %= 86400;
	}
	if (tm >= 3600) {
		hr = tm / 3600;			/* Hours */
		tm %= 3600;
	}
	if (tm >= 60) {
		mn = tm / 60;			/* Minutes */
		tm %= 60;				/* Seconds */
	}

	*buf = '\0';
	if (dy) {
		buf = fmt.Sprintf("%dd %02d:%02d", dy, hr, mn)
	} else {
		buf = fmt.Sprintf("%02d:%02d", hr, mn)
	}
	return buf;
}

func mfn_stimestr(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	int tm = strconv.Atol(argv[0])
	int dy, hr, mn;

	dy = hr = mn = 0;
	if (tm >= 86400) {
		dy = tm / 86400;		/* Days */
		tm %= 86400;
	}
	if (tm >= 3600) {
		hr = tm / 3600;			/* Hours */
		tm %= 3600;
	}
	if (tm >= 60) {
		mn = tm / 60;			/* Minutes */
		tm %= 60;				/* Seconds */
	}

	*buf = '\0';
	if (dy) {
		buf = fmt.Sprintf("%dd", dy)
		return buf;
	}
	if (hr) {
		buf = fmt.Sprintf("%dh", hr)
		return buf;
	}
	if (mn) {
		buf = fmt.Sprintf("%dm", mn)
		return buf;
	}
	buf = fmt.Sprintf("%ds", tm)
	return buf;
}

func mfn_secs(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	var lt time_t
	time(&lt)
	return fmt.Sprint(lt)
}

func mfn_convsecs(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	var lt time_t
	lt := atol(argv[0])
	return fmt.Sprint(ctime(&lt))
}

func mfn_loc(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("LOC", mesg_ObjectID_local(descr, player, what, perms, argv[0], mesgtyp), func(obj ObjectID) {
		r = ref2str(DB.Fetch(obj).Location)
	})
	return
}

func mfn_nearby(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	obj := mesg_ObjectID_raw(descr, player, what, perms, argv[0])
	switch {
	case obj == UNKNOWN || obj == AMBIGUOUS || obj == NOTHING:
		ABORT_MPI("NEARBY", "Match failed (arg1).")
	case obj == PERMDENIED:
		ABORT_MPI("NEARBY", "Permission denied (arg1).")
	case obj == HOME:
		obj = DB.FetchPlayer(player).home
	}
	var obj2 ObjectID
	if len(argv) > 1 {
		obj2 = mesg_ObjectID_raw(descr, player, what, perms, argv[1])
		switch {
		case obj2 == UNKNOWN || obj2 == AMBIGUOUS || obj2 == NOTHING:
			ABORT_MPI("NEARBY", "Match failed (arg2).")
		case obj2 == PERMDENIED:
			ABORT_MPI("NEARBY", "Permission denied (arg2).")
		case obj2 == HOME:
			obj2 = DB.FetchPlayer(player).home
		}
	} else {
		obj2 = what
	}
	if !(mesgtyp & MPI_ISBLESSED) && !isneighbor(obj, what) && !isneighbor(obj2, what) && !isneighbor(obj, player) && !isneighbor(obj2, player) {
		ABORT_MPI("NEARBY", "Permission denied.  Neither object is local.")
	}
	if isneighbor(obj, obj2) {
		r = "1"
	} else {
		r = "0"
	}
	return
}

func mfn_money(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("MONEY", mesg_ObjectID(descr, player, what, perms, argv[0], mesgtyp), func(obj ObjectID) {
		switch TYPEOF(obj) {
		case TYPE_THING, TYPE_PLAYER:
			r = fmt.Sprint(get_property_value(obj, MESGPROP_VALUE))
		default:
			r = "0"
		}
	})
	return
}

func mfn_flags(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("FLAGS", mesg_ObjectID_local(descr, player, what, perms, argv[0], mesgtyp), func(obj ObjectID) {
		r = unparse_flags(obj)
	})
	return 
}

func mfn_tell(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	obj := player
	if len(argv) > 1 {
		obj = mesg_ObjectID_local(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("TELL", obj, func(obj ObjectID) {
		if mesgtyp & MPI_ISLISTENER != 0 && TYPEOF(what) != TYPE_ROOM {
			ABORT_MPI("TELL", "Permission denied.")
		} else {
			buf2 := argv[0]
			items := strings.Split(argv[0], MPI_LISTSEP)
			for _, v := range items {
				var buf string
				if obj != DB.Fetch(perms).Owner && obj != player {
					buf = "> "
				}
				loc := DB.Fetch(what).Location
				name := DB.Fetch(player).name
				switch {
				case Typeof(what) == TYPE_ROOM, DB.Fetch(what).Owner == obj, player == obj, (Typeof(what) == TYPE_EXIT && Typeof(loc) == TYPE_ROOM), strings.Prefix(argv[0], name):
					buf += fmt.Sprintf("%.4093s", v)
				} else {
					buf += name
					if argv[0] != '\'' && !unicode.IsSpace(argv[0]) {
						buf += " "
					}
					buf += fmt.Sprintf("%.4078s", v)
				}
				notify_from_echo(player, obj, buf, 1)
			}
		}
	})
	return argv[0]
}

func mfn_otell(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) string {
	obj := DB.Fetch(player).Location
	if len(argv) > 1 {
		obj = mesg_ObjectID_local(descr, player, what, perms, argv[1], mesgtyp)
	}
	with_useful_object("OTELL", obj, func(obj ObjectID) {
		if mesgtyp & MPI_ISLISTENER != 0 && TYPEOF(what) != TYPE_ROOM {
			ABORT_MPI("OTELL", "Permission denied.")
		} else {
			eobj := player
			if len(argv) > 2 {
				eobj = mesg_ObjectID_raw(descr, player, what, perms, argv[2])
			}
			items := strings.Split(argv[0], MPI_LISTSEP)
			for _, v := range items {
				loc := DB.Fetch(what).Location
				name := DB.Fetch(player).name
				var buf string
				switch {
				case ((DB.Fetch(what).Owner == DB.Fetch(obj).Owner || isancestor(what, obj)) && (TYPEOF(what) == TYPE_ROOM || (TYPEOF(what) == TYPE_EXIT && TYPEOF(loc) == TYPE_ROOM))) || strings.Prefix(argv[0], name):
					buf = v
				case argv[0] == '\'' || unicode.IsSpace(argv[0]):
					buf = fmt.Sprint(DB.Fetch(player).name, v)
				default:
					buf = fmt.Sprint(DB.Fetch(player).name, " ", v)
				}
				for thing := DB.Fetch(obj).Contents; thing != NOTHING; thing = DB.Fetch(thing).next {
					if thing != eobj {
						notify_from_echo(player, thing, buf, 0)
					}
				}
			}
		}
	})
	return argv[0]
}

func mfn_right(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	width := 78
	if len(argv) > 1 {
		width = strconv.Atoi(argv[1])
	}
	filler := " "
	if len(argv) > 2 {
		if filler = argv[2]; filler == "" {
			ABORT_MPI("RIGHT", "Null pad string.")
		}
	}

	for i := len(argv[0]); i < length; i++ {
		buf += filler
	}
	buf += argv[0]
	return
}

func mfn_left(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	width := 78
	if len(argv) > 1 {
		width = strconv.Atoi(argv[1])
	}
	filler := " "
	if len(argv) > 2 {
		if filler = argv[2]; filler == "" {
			ABORT_MPI("LEFT", "Null pad string.")
		}
	}
	buf = argv[0]
	for i := len(argv[0]); i < width; i++ {
		buf += filler
	}
	return
}

func mfn_center(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (buf string) {
	width := 78
	if len(argv) > 1 {
		width = strconv.Atoi(argv[1])
	}
	half_width := width / 2

	filler := " "
	if len(argv) > 2 {
		if filler = argv[2]; filler == "" {
			ABORT_MPI("CENTER", "Null pad string.")
		}
	}

	for i := len(argv[0]) / 2; i < half_width; i++ {
		buf += filler
	}
	buf += argv[0]
	for i := len(buf); i < width; i++ {
		buf += filler
	}
	return
}

func mfn_created(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("CREATED", mesg_ObjectID(descr, player, what, perms, argv[0], mesgtyp), func(obj ObjectID) {
		r = fmt.Sprint(DB.Fetch(obj).Created)
	})
	return
}

func mfn_lastused(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("LASTUSED", mesg_ObjectID(descr, player, what, perms, argv[0], mesgtyp), func(onj ObjectID) {
		r = fmt.Sprint(DB.Fetch(obj).LastUsed)
	})
	return
}

func mfn_modified(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("MODIFIED", mesg_ObjectID(descr, player, what, perms, argv[0], mesgtyp), func(obj ObjectID) {
		r = fmt.Sprint(DB.Fetch(obj).Modified)
	})
	return
}

func mfn_usecount(descr int, player, what, perms ObjectID, argv MPIArgs, mesgtyp int) (r string) {
	with_useful_object("USECOUNT", mesg_ObjectID(descr, player, what, perms, argv[0], mesgtyp), func(obj ObjectID) {
		r = fmt.Sprint(DB.Fetch(obj).Uses)
	})
	return
}