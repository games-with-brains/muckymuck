package fbmuck

func copyobj(player, old, nu dbref) {
	newp := db.Fetch(nu)
	newp.name = db.Fetch(old).name
	if TYPEOF(old) == TYPE_THING {
		db.Fetch(nu).sp.(player_specific) = new(player_specific)
		db.Fetch(nu).sp.(player_specific).home = player
		add_property(nu, MESGPROP_VALUE, nil, 1)
	}
	newp.properties = copy_prop(old)
	newp.exits = NOTHING
	newp.contents = NOTHING
	newp.next = NOTHING
	newp.location = NOTHING
	moveto(nu, player)
	db.Fetch(nu).flags |= OBJECT_CHANGED
}

func prim_addpennies(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(tp_addpennies_muf_mlev, mlev, 2, top, func(op Array) {
		ref := valid_object(op[0])
		pennies := op[1].(int)
		switch Typeof(ref) {
		case TYPE_PLAYER:
			result = get_property_value(ref, MESGPROP_VALUE)
			if mlev < WIZBIT {
				if pennies > 0 {
					switch {
					case result > result + pennies:
						panic("Would roll over player's score.")
					case result + pennies > tp_max_pennies:
						panic("Would exceed MAX_PENNIES.")
					}
				} else {
					switch {
					case result < result + pennies:
						panic("Would roll over player's score.")
					case result + pennies < 0:
						panic("Result would be negative.")
					}
				}
			}
			result += pennies
			add_property(ref, MESGPROP_VALUE, nil, get_property_value(ref, MESGPROP_VALUE) + pennies)
			db.Fetch(ref).flags |= OBJECT_CHANGED
		case TYPE_THING:
			if mlev < WIZBIT {
				panic("Permission denied.")
			}
			result = get_property_value(ref, MESGPROP_VALUE) + pennies
			if result < 1 {
				panic("Result must be positive.")
			}
			add_property(ref, MESGPROP_VALUE, nil, get_property_value(ref, MESGPROP_VALUE) + pennies)
			db.Fetch(ref).flags |= OBJECT_CHANGED
		default:
			panic("Invalid object type.")
		}
	})
}

func prim_moveto(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		victim := valid_object(op[0])
		dest := valid_object_or_home(op[1])
		switch {
		case fr.level > 8:
			panic("Interp call loops not allowed.")
		case TYPEOF(dest) == TYPE_EXIT:
			panic("Destination argument is an exit.")
		case db.Fetch(victim).flags & JUMP_OK == 0 && !permissions(ProgUID, victim) && mlev < MASTER:
			panic("Object can't be moved.")
		}
		switch TYPEOF(victim) {
		case TYPE_PLAYER:
			switch {
			case TYPEOF(dest) != TYPE_ROOM && !(Typeof(dest) == TYPE_THING && db.Fetch(dest).flags & VEHICLE != 0):
				panic("Bad destination.")
			case parent_loop_check(victim, dest):
				panic("Things can't contain themselves.")
			case mlev > JOURNEYMAN:
			case db.Fetch(db.Fetch(victim).location).flags & JUMP_OK == 0 && !permissions(ProgUID, db.Fetch(victim).location):
				panic("Source not JUMP_OK.")
			case !is_home(oper1) && db.Fetch(dest).flags & JUMP_OK == 0 && !permissions(ProgUID, dest):
				panic("Destination not JUMP_OK.")
			case TYPEOF(dest) == TYPE_THING && db.Fetch(victim).location != db.Fetch(dest).location:
				panic("Not in same location as vehicle.")
			}
			enter_room(fr.descr, victim, dest, program)
		case TYPE_THING:
			switch {
			case parent_loop_check(victim, dest):
				panic("A thing cannot contain itself.")
			case mlev > JOURNEYMAN, Typeof(dest) == TYPE_THING:
			case db.Fetch(victim).flags & VEHICLE != 0 && db.Fetch(dest).flags & VEHICLE != 0:
				panic("Destination doesn't accept vehicles.")
			case db.Fetch(victim).flags & ZOMBIE != 0 && db.Fetch(dest).flags & ZOMBIE != 0:
				panic("Destination doesn't accept zombies.")
			}
			ts_lastuseobject(victim)
		case TYPE_PROGRAM:
			matchroom := NOTHING
			switch {
			case TYPEOF(dest) != TYPE_ROOM && Typeof(dest) != TYPE_PLAYER && Typeof(dest) != TYPE_THING:
				panic("Bad destination.")
			case mlev < MASTER:
				if permissions(ProgUID, dest) {
					matchroom = dest
				}
				if permissions(ProgUID, db.Fetch(victim).location) {
					matchroom = db.Fetch(victim).location
				}
				if matchroom != NOTHING && db.Fetch(matchroom).flags & JUMP_OK == 0 && !permissions(ProgUID, victim) {
					panic("Permission denied.")
				}
			case TYPEOF(victim) == TYPE_THING && (tp_thing_movement || db.Fetch(victim).flags & ZOMBIE != 0):
				enter_room(fr.descr, victim, dest, program)
			default:
				moveto(victim, dest)
			}
		case TYPE_EXIT:
			switch {
			case mlev < MASTER && (!permissions(ProgUID, victim) || !permissions(ProgUID, dest)):
				panic("Permission denied.")
			case dest == HOME, TYPEOF(dest) != TYPE_ROOM && TYPEOF(dest) != TYPE_THING && TYPEOF(dest) != TYPE_PLAYER:
				panic("Bad destination object.")
			case unset_source(ProgUID, db.Fetch(player).location, victim):
				set_source(ProgUID, victim, dest)
				SetMLevel(victim, NON_MUKCER)
			}
		case TYPE_ROOM:
			switch {
			case !tp_thing_movement && Typeof(dest) != TYPE_ROOM:
				panic("Bad destination.")
			case victim == GLOBAL_ENVIRONMENT:
				panic("Permission denied.")
			case dest == HOME:
				/* Allow the owner of the room or the owner of the room's location to reparent the room to #0 */
				if mlev < MASTER && !permissions(ProgUID, victim) && !permissions(ProgUID, db.Fetch(victim).location) {
					panic("Permission denied.")
				}
				dest = GLOBAL_ENVIRONMENT
			case mlev < MASTER && (!permissions(ProgUID, victim) || !can_link_to(ProgUID, NOTYPE, dest)):
				panic("Permission denied.")
			case parent_loop_check(victim, dest):
				panic("Parent room would create a loop.")
			}
			ts_lastuseobject(victim)
			moveto(victim, dest)
		default:
			panic("Invalid object type (1)")
		}
	})
}

func prim_pennies(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		if mlev < tp_pennies_muf_mlev {
			panic("Permission Denied (mlev < tp_pennies_muf_mlev)")
		}
		switch Typeof(obj) {
		case TYPE_PLAYER, TYPE_THING:
			push(arg, top, get_property_value(obj, MESGPROP_VALUE))
		default:
			panic("Invalid argument.")
		}
	})
}

func prim_dbcomp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, op[0].(dbref) == op[1].(dbref))
	})
}

func prim_dbref(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, dbref(op[0].(int)))
	})
}

func prim_contents(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		ref := db.Fetch(obj).contents
		for mlev < JOURNEYMAN && ref != NOTHING && db.Fetch(ref).flags & DARK != 0 && !controls(ProgUID, ref) {
			ref = db.Fetch(ref).next
		}
		push(arg, top, ref)
	})
}

func prim_exits(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		ref := valid_remote_object(player, mlev, op[0])
		if mlev < MASTER && !permissions(ProgUID, ref) {
			panic("Permission denied.")
		}
		switch Typeof(ref) {
		case TYPE_ROOM, TYPE_THING, TYPE_PLAYER:
			ref = db.Fetch(ref).exits
		default:
			panic("Invalid object.")
		}
		push(arg, top, ref)
	})
}

func prim_next(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		ref := db.Fetch(obj).next
		for mlev < JOURNEYMAN && ref != NOTHING && Typeof(ref) != TYPE_EXIT && (db.Fetch(ref).flags & DARK != 0 || Typeof(ref) == TYPE_ROOM) && !controls(ProgUID, ref) {
			ref = db.Fetch(ref).next
		}
		push(arg, top, ref)
	})
}

func prim_nextowned(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(JOURNEYMAN, mlev, 1, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		owner := db.Fetch(obj).owner
		if Typeof(obj) == TYPE_PLAYER {
			obj = 0
		} else {
			obj++
		}
		for ; obj < db_top && (db.Fetch(obj).owner != owner || obj == owner); obj++ {}
		if obj >= db_top {
			obj = NOTHING
		}
		push(arg, top, obj)
	})
}

func prim_name(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, db.Fetch(valid_remote_object(player, mlev, op[0])).name)
	})
}

func prim_setname(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		ref := valid_object(op[0])
		password := op[1].(string)
		if mlev < WIZBIT && !permissions(ProgUID, ref) {
			panic("Permission denied.")
		}

		if Typeof(ref) == TYPE_PLAYER {
			var name string
			password = strings.TrimSpace(password)
			if i := strings.IndexFunc(password, unicode.IsSpace); i > 0 {
				name = password[:i]
				password = strings.TrimSpace(password[i:])
			}
			if i = strings.IndexFunc(password, unicode.IsSpace); i > 0 {
				password = password[:i]
			}
			switch {
			case mlev < WIZBIT:
				panic("Permission denied.");
			case password == "":
				panic("Player namechange requires password.")
			case !check_password(ref, password):
				panic("Incorrect password.")
			case  != db.Fetch(ref).name && !ok_player_name(name):
				panic("You can't give a player that name.")
			}

			log_status("NAME CHANGE (MUF): %s(#%d) to %s", db.Fetch(ref).name, ref, name)
			delete_player(ref)
			db.Fetch(ref).name = name
			add_player(ref)
			ts_modifyobject(ref)
		} else {
			if (Typeof(ref) == TYPE_THING && !ok_ascii_thing(name)) || (Typeof(ref) != TYPE_THING && !ok_ascii_other(name)) {
				panic("Invalid 8-bit name.")
			}
			if !ok_name(name) {
				panic("Invalid name.")
			}
			db.Fetch(ref).name = name
			ts_modifyobject(ref)
			if MLevRaw(ref) != NON_MUCKER {
				SetMLevel(ref, NON_MUCKER)
			}
		}
	})
}

func prim_pmatch(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if name := op[0].(string); name == "me" {
			push(arg, top, player)
		} else {
			push(arg, top, lookup_player(name))
		}
	})
}

func prim_match(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		name := op[0].(string)
		buf := match_args
		tmppp := match_cmdname
		md := NewMatch(fr.descr, player, name, NOTYPE)
		if name[0] == REGISTERED_TOKEN {
			md.MatchRegistered()
		} else {
			md.MatchAllExits().
				MatchNeighbor().
				MatchPossession().
				MatchMe().
				MatchHere().
				MatchHome()
		}
		if Wizard(ProgUID) || mlev >= WIZBIT {
			md.MatchAbsolute().MatchPlayer()
		}
		ref = md.MatchResult()
		match_args = buf
		match_cmdname = tmppp
		push(arg, top, name)
	})
}

func prim_rmatch(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		switch obj.(type) {
		case TYPE_PROGRAM, TYPE_EXIT:
			panic("Invalid argument (1)")
		}
		name := op[1].(dbref)

		buf := match_args
		tmppp := match_cmdname
		md := NewMatch(fr.descr, player, name, TYPE_THING)
		md.RMatch(obj)
		ref := md.MatchResult()
		match_args = buf
		match_cmdname = tmppp
		push(arg, top, ref)
	})
}

func prim_copyobj(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		ref := valid_remote_object(player, mlev, op[0])
		switch {
		case mlev < MASTER:
			switch {
			case fr.already_created:
				panic("Can't create any more objects.")
			case !permissions(ProgUID, ref):
				panic("Permission denied.")
			}
		case Typeof(ref) != TYPE_THING:
			panic("Invalid object type.")
		case !ok_name(db.Fetch(ref).name):
			panic("Invalid name.")
		}
		fr.already_created++
		newobj := new_object()
		*db.Fetch(newobj) = *db.Fetch(ref)
		copyobj(player, ref, newobj)
		push(arg, top, newobj)
	})
}

func prim_set(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		var tmp int
		var result, truwiz bool
		ref := valid_remote_object(player, mlev, op[0])
		flag := op[1].(string)
		if flag == "" {
			panic("Unknown flag.")
		}
		for flag[0] == '!' {
			flag = flag[1:]
			result = !result
		}

		switch {
		case strings.Prefix(flag, "dark"), strings.Prefix(flag, "debug"):
			tmp = DARK
		case strings.Prefix(flag, "abode"), strings.Prefix(flag, "autostart"), strings.Prefix(flag, "abate"):
			tmp = ABODE
		case strings.Prefix(flag, "chown_ok"), strings.Prefix(flag, "color"):
			tmp = CHOWN_OK
		case strings.Prefix(flag, "haven"), strings.Prefix(flag, "harduid"):
			tmp = HAVEN
		case strings.Prefix(flag, "jump_ok"):
			tmp = JUMP_OK
		case strings.Prefix(flag, "link_ok"):
			tmp = LINK_OK
		case strings.Prefix(flag, "kill_ok"):
			tmp = KILL_OK
		case strings.Prefix(flag, "builder"):
			tmp = BUILDER
		case strings.Prefix(flag, "mucker"):
			tmp = MUCKER
		case strings.Prefix(flag, "nucker"):
			tmp = SMUCKER
		case strings.Prefix(flag, "interactive"):
			tmp = INTERACTIVE
		case strings.Prefix(flag, "sticky"), strings.Prefix(flag, "silent"):
			tmp = STICKY
		case strings.Prefix(flag, "wizard"):
			tmp = WIZARD
		case strings.Prefix(flag, "truewizard"):
			tmp = WIZARD
		case strings.Prefix(flag, "xforcible"):
			tmp = XFORCIBLE
		case strings.Prefix(flag, "zombie"):
			tmp = ZOMBIE
		case strings.Prefix(flag, "vehicle"), strings.Prefix(flag, "viewable"):
			tmp = VEHICLE
		case strings.Prefix(flag, "quell"):
			tmp = QUELL
		case tp_enable_match_yield && strings.Prefix(flag, "yield"):
			tmp = YIELD
		case tp_enable_match_yield && strings.Prefix(flag, "overt"):
			tmp = OVERT
		}
		if tmp == 0 {
			panic("Unrecognized flag.")
		}
		if mlev < WIZBIT {
			if !permissions(ProgUID, ref) {
				panic("Permission denied.")
			}
			if ((((tmp == DARK && (Typeof(ref) == TYPE_PLAYER || (!tp_exit_darking && Typeof(ref) == TYPE_EXIT) || (!tp_thing_darking && Typeof(ref) == TYPE_THING)))
								|| (tmp == ZOMBIE && Typeof(ref) == TYPE_THING && db.Fetch(ProgUID).flags & ZOMBIE != 0) || (tmp == ZOMBIE && Typeof(ref) == TYPE_PLAYER)
								|| tmp == BUILDER || tmp == YIELD || tmp == OVERT)) || tmp == WIZARD || tmp == QUELL || tmp == INTERACTIVE
								|| (tmp == ABODE && Typeof(ref) == TYPE_PROGRAM) || tmp == MUCKER || tmp == SMUCKER || tmp == XFORCIBLE) {
				panic("Permission denied.")
			}
		}
		if (tmp == YIELD || tmp == OVERT) && (Typeof(ref) != TYPE_THING && Typeof(ref) != TYPE_ROOM) {
			panic("Permission denied.")
		}
		if result && Typeof(ref) == TYPE_THING && tmp == VEHICLE {
			for obj := db.Fetch(ref).contents; obj != NOTHING; obj = db.Fetch(obj).next {
				if TYPEOF(obj) == TYPE_PLAYER {
					panic("That vehicle still has players in it!")
				}
			}
		}
		if !result {
			db.Fetch(ref).flags |= tmp
			db.Fetch(ref).flags |= OBJECT_CHANGED
		} else {
			db.Fetch(ref).flags &= ~tmp
			db.Fetch(ref).flags |= OBJECT_CHANGED
		}
	})
}

func prim_mlevel(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, MLevRaw(valid_remote_object(player, mlev, op[0])))
	})
}

func prim_flagp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		var tmp int
		var negate, truwiz bool
		ref := valid_remote_object(player, mlev, op[0])
		flag := op[1].(string)
		if flag == "" {
			panic("Unknown flag.")
		}
		for flag[0] == '!' {
			flag = flag[1:]
			negate = true
		}
		switch {
		case strings.Prefix(flag, "dark"), strings.Prefix(flag, "debug"):
			tmp = DARK
		case strings.Prefix(flag, "abode"), strings.Prefix(flag, "autostart"), strings.Prefix(flag, "abate"):
			tmp = ABODE
		case strings.Prefix(flag, "chown_ok"), strings.Prefix(flag, "color"):
			tmp = CHOWN_OK
		case strings.Prefix(flag, "haven"), strings.Prefix(flag, "harduid"):
			tmp = HAVEN
		case strings.Prefix(flag, "jump_ok"):
			tmp = JUMP_OK
		case strings.Prefix(flag, "link_ok"):
			tmp = LINK_OK
		case strings.Prefix(flag, "kill_ok"):
			tmp = KILL_OK
		case strings.Prefix(flag, "builder"):
			tmp = BUILDER
		case strings.Prefix(flag, "mucker"):
			tmp = MUCKER
		case strings.Prefix(flag, "nucker"):
			tmp = SMUCKER
		case strings.Prefix(flag, "interactive"):
			tmp = INTERACTIVE
		case strings.Prefix(flag, "sticky"), strings.Prefix(flag, "silent"):
			tmp = STICKY
		case strings.Prefix(flag, "wizard"):
			tmp = WIZARD
		case strings.Prefix(flag, "truewizard"):
			tmp = WIZARD
			truwiz = true
		case strings.Prefix(flag, "zombie"):
			tmp = ZOMBIE
		case strings.Prefix(flag, "xforcible"):
			tmp = XFORCIBLE
		case strings.Prefix(flag, "vehicle"), strings.Prefix(flag, "viewable"):
			tmp = VEHICLE
		case strings.Prefix(flag, "quell"):
			tmp = QUELL
		case tp_enable_match_yield && strings.Prefix(flag, "yield"):
			tmp = YIELD
		case tp_enable_match_yield && strings.Prefix(flag, "overt"):
			tmp = OVERT
		}
		if negate {
			if !truwiz && tmp == WIZARD {
				push(arg, top, MUFBool(!Wizard(ref)))
			} else {
				push(arg, top, MUFBool(tmp != 0 && db.Fetch(ref).flags & tmp == 0))
			}
		} else {
			if !truwiz && tmp == WIZARD {
				push(arg, top, MUFBool(Wizard(ref)))
			} else {
				push(arg, top, MUFBool(tmp != 0 && db.Fetch(ref).flags & tmp != 0))
			}
		}
	})
}

func apply_predicate(top *int, f func(dbref) int) {
	apply_primitive(1, top, func(op Array) {
		var ok bool
		if ref := valid_object_or_home(op[0], false); ref != NOTHING {
			ok = f(ref)
		}
		push(arg, top, MUFBool(ok))
	})	
}

func prim_playerp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_predicate(top, func(ref dbref) (ok bool) {
		_, ok = ref.(TYPE_PLAYER)
		return
	})
}

func prim_thingp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_predicate(top, func(ref dbref) (ok bool) {
		_, ok = ref.(TYPE_THING)
		return
	})
}

func prim_roomp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_predicate(top, func(ref dbref) (ok bool) {
		_, ok = ref.(TYPE_ROOM)
		return
	})
}

func prim_programp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_predicate(top, func(ref dbref) (ok bool) {
		_, ok = ref.(TYPE_PROGRAM)
		return
	})
}

func prim_exitp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_predicate(top, func(ref dbref) (ok bool) {
		_, ok = ref.(TYPE_EXIT)
		return
	})
}

func prim_okp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		var ok bool
		if ref := valid_object(op[0], false); ref != NOTHING {
			ok = f(ref)
		}
		push(arg, top, MUFBool(ok))
	})
}

func prim_location(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		push(arg, top, db.Fetch(obj).location)
	})
}

func prim_owner(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		push(arg, top, db.Fetch(obj).owner)
	})
}

func prim_controls(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		who := valid_object(op[0])
		obj := valid_remote_object(player, mlev, op[1])
		push(arg, top, controls(who, obj))
	})
}

func prim_getlink(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		switch Typeof(obj) {
		case TYPE_PROGRAM:
			panic("Illegal object referenced.")
		case TYPE_EXIT:
			if len(db.Fetch(obj).sp.exit.dest) != 0 {
				ref = db.Fetch(obj).sp.exit.dest[0]
			} else {
				ref = NOTHING
			}
		case TYPE_PLAYER:
			ref = db.Fetch(obj).sp.(player_specific).home
		case TYPE_THING:
			ref = db.Fetch(obj).sp.(player_specific).home
		case TYPE_ROOM:
			ref = db.Fetch(obj).sp.(dbref)
		default:
			ref = NOTHING
		}
		push(arg, top, ref)
	})
}

func prim_getlinks(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		my_obj := valid_remote_object(player, mlev, op[0])
		if Typeof(my_obj) == TYPE_PROGRAM {
			panic("Illegal object referenced.")
		}
		switch Typeof(my_obj) {
		case TYPE_EXIT:
			for _, v := range db.Fetch(my_obj).sp.exit.dest {
				push(arg, top, v)
			}
			push(arg, top, len(db.Fetch(my_obj).sp.exit.dest))
		case TYPE_PLAYER:
			push(arg, top, db.Fetch(my_obj).sp.(player_specific).home)
			push(arg, top, 1)
		case TYPE_THING:
			push(arg, top, db.Fetch(my_obj).sp.(player_specific).home)
			push(arg, top, 1)
		case TYPE_ROOM:
			if ref := db.Fetch(my_obj).sp; ref != NOTHING {
				push(arg, top, 0)
			} else {
				push(arg, top, ref.(dbref))
				push(arg, top, 1)
			}
		default:
			push(arg, top, 0)
		}
	})
}

func prog_can_link_to(mlev int, who dbref, what_type object_flag_type, where dbref) (r bool) {
	switch {
	case where == HOME:
		r = true
	case where < 0, where >= db_top:
	default:
		switch what_type {
		case TYPE_EXIT:
			r = mlev > MASTER || permissions(who, where) || db.Fetch(where).flags & LINK_OK != 0
		case TYPE_PLAYER:
			r = Typeof(where) == TYPE_ROOM && (mlev > MASTER || permissions(who, where) || Linkable(where))
		case TYPE_ROOM:
			r = (Typeof(where) == TYPE_ROOM || Typeof(where) == TYPE_THING) && (mlev > MASTER || permissions(who, where) || Linkable(where))
		case TYPE_THING:
			r = (Typeof(where) == TYPE_ROOM || Typeof(where) == TYPE_PLAYER || Typeof(where) == TYPE_THING) && (mlev > MASTER || permissions(who, where) || Linkable(where))
		case NOTYPE:
			r = mlev > MASTER || permissions(who, where) || (db.Fetch(where).flags & LINK_OK) || (Typeof(where) != TYPE_THING && db.Fetch(where).flags & ABODE != 0)
		}
	}
	return
}

func prim_setlink(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		source := valid_object(op[0])
		if op[1] == NOTHING {
			if mlev < WIZBIT && !permissions(ProgUID, source) {
				panic("Permission denied.")
			}
			switch Typeof(source) {
			case TYPE_EXIT:
				db.Fetch(source).sp.exit.dest = nil
				db.Fetch(source).flags |= OBJECT_CHANGED
				if MLevRaw(source) != NON_MUCKER {
					SetMLevel(source, NON_MUCKER)
				}
			case TYPE_ROOM:
				db.Fetch(source).sp = NOTHING
				db.Fetch(source).flags |= OBJECT_CHANGED
			default:
				panic("Invalid object. (1)")
			}
		} else {
			dest := valid_object_or_home(op[1])
			switch {
			case Typeof(source) == TYPE_PROGRAM:
				panic("Program objects are not linkable. (1)")
			case !prog_can_link_to(mlev, ProgUID, Typeof(source), dest):
				panic("Can't link source to destination.")
			}
			switch Typeof(source) {
			case TYPE_EXIT:
				switch {
				case mlev < WIZBIT && !permissions(ProgUID, source):
					panic("Permission denied.")
				case len(db.Fetch(source).sp.exit.dest) != 0:
					panic("Exit is already linked.")
				case exit_loop_check(source, dest):
					panic("Link would cause a loop.")
				}
				db.Fetch(source).sp.exit.dest = []dbref{ dest }
				db.Fetch(source).flags |= OBJECT_CHANGED
			case TYPE_PLAYER:
				switch {
				case mlev < WIZBIT && !permissions(ProgUID, source):
					panic("Permission denied.")
				case dest == HOME:
					panic("Cannot link player to HOME.")
				}
				db.Fetch(source).sp.(player_specific).home = dest
				db.Fetch(source).flags |= OBJECT_CHANGED
			case TYPE_THING:
				switch {
				case mlev < WIZBIT && !permissions(ProgUID, source):
					panic("Permission denied.")
				case dest == HOME:
					panic("Cannot link thing to HOME.")
				case parent_loop_check(source, dest):
					panic("That would cause a parent paradox.")
				}
				db.Fetch(source).sp.(player_specific).home = dest
				db.Fetch(source).flags |= OBJECT_CHANGED
			case TYPE_ROOM:
				if mlev < WIZBIT && !permissions(ProgUID, source) {
					panic("Permission denied.")
				}
				db.Fetch(source).sp = dest
				db.Fetch(source).flags |= OBJECT_CHANGED
			}
		}
	})
}

func prim_setown(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		obj := valid_object(op[0])
		who := valid_player(op[1])
		switch {
		case Typeof(obj) == TYPE_PLAYER:
			panic("Permission denied: cannot set owner of player. (1)")
		case mlev > MASTER:
		case who != player:
			panic("Permission denied. (2)")
		case db.Fetch(obj).flags & CHOWN_OK == 0, !test_lock(fr.descr, player, obj, "_/chlk"):
			panic("Permission denied. (1)")
		case Typeof(obj) == TYPE_ROOM && db.Fetch(player).location != obj:
			panic("Permission denied: not in room. (1)")
		case Typeof(obj) == TYPE_THING && db.Fetch(obj).location != player:
			panic("Permission denied: object not carried. (1)")
		}
		db.Fetch(obj).owner = db.Fetch(who).owner
		db.Fetch(obj).flags |= OBJECT_CHANGED
	})
}

func prim_newobject(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		loc := valid_remote_object(player, mlev, op[0])
		name := op[1].(string)
		switch {
		case mlev > JOURNEYMAN:
		case fr.already_created:
			panic("An object was already created this program run.")
		case !permissions(ProgUID, loc) {
			panic("Permission denied.")
		case Typeof(loc) != TYPE_ROOM:
			panic("Invalid argument (1)")
		case !ok_ascii_other(name), !ok_name(name):
			panic("Invalid name. (2)")
		}
		ref := new_object()
		db.Fetch(ref).name = name
		db.Fetch(ref).sp.(player_specific) = new(player_specific)
		db.Fetch(ref).location = loc
		db.Fetch(ref).owner = db.Fetch(ProgUID).owner
		add_property(ref, MESGPROP_VALUE, nil, 1)
		db.Fetch(ref).exits = NOTHING
		db.Fetch(ref).flags = TYPE_THING
		if l := db.Fetch(player).location; l != NOTHING && controls(player, l) {
			db.Fetch(ref).sp.(player_specific).home = l
		} else {
			db.Fetch(ref).sp.(player_specific).home = db.Fetch(player).sp.(player_specific).home
		}
		CHECKOFLOW(3)
		db.Fetch(ref).next = db.Fetch(loc).contents
		db.Fetch(ref).flags |= OBJECT_CHANGED
		db.Fetch(loc).contents = ref

		db.Fetch(ref).flags |= OBJECT_CHANGED
		db.Fetch(loc).flags |= OBJECT_CHANGED
		push(arg, top, ref)
	})
}

func prim_newroom(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		loc := valid_object(op[0])
		name := op[1].(string)
		switch {
		case mlev > JOURNEYMAN:
		case fr.already_created:
			panic("An object was already created this program run.")
		case !permissions(ProgUID, loc) {
			panic("Permission denied.")
		case Typeof(loc) != TYPE_ROOM:
			panic("Invalid argument (1)")
		case !ok_ascii_other(name), !ok_name(name):
			panic("Invalid name. (2)")
		}
		ref := new_object()
		db.Fetch(ref).name = name
		db.Fetch(ref).location = loc
		db.Fetch(ref).owner = db.Fetch(ProgUID).owner
		db.Fetch(ref).exits = NOTHING
		db.Fetch(ref).sp = NOTHING
		db.Fetch(ref).flags = TYPE_ROOM | (db.Fetch(player).flags & JUMP_OK)
		CHECKOFLOW(3)
		db.Fetch(ref).next = loc.contents
		db.Fetch(ref).flags |= OBJECT_CHANGED
		loc.contents = ref
		db.Fetch(ref).flags |= OBJECT_CHANGED
		db.Fetch(loc).flags |= OBJECT_CHANGED
		push(arg, top, ref)
	}
}

func prim_newexit(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 2, top, func(op Array) {
		loc := valid_remote_object(player, mlev, op[0])
		name := op[1].(string)
		switch {
		case Typeof(loc) == TYPE_ROOM, Typeof(loc) == TYPE_THING:
		default:
			panic("Invalid argument (1)")
		}
		switch {
		case mlev < WIZBIT && !permissions(ProgUID, loc):
			panic("Permission denied.")
		case !ok_ascii_other(name), !ok_name(name):
			panic("Invalid name. (2)")
		}
		ref := new_object()
		db.Fetch(ref).name = name
		db.Fetch(ref).location = loc
		db.Fetch(ref).owner = db.Fetch(ProgUID).owner
		db.Fetch(ref).flags = TYPE_EXIT
		db.Fetch(ref).sp.exit.dest = nil

		/* link it in */
		CHECKOFLOW(3)
		db.Fetch(ref).next = db.Fetch(loc).exits
		db.Fetch(ref).flags |= OBJECT_CHANGED
		db.Fetch(loc).exits = ref

		db.Fetch(loc).flags |= OBJECT_CHANGED
		push(arg, top, loc)
	})
}

func prim_lockedp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		if fr.level > 8 {
			panic("Interp call loops not allowed.")
		}
		p := valid_remote_player(player, mlev, op[0])
		obj := valid_remote_object(player, mlev, op[1])
		push(arg, top, MUFBool(!could_doit(fr.descr, p, obj)))
	})
}

func prim_recycle(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		ref := valid_object(op[0])
		switch {
		case mlev < WIZBIT && !permissions(ProgUID, result):
			panic("Permission denied.")
		case ref == tp_player_start, ref == GLOBAL_ENVIRONMENT:
			panic("Cannot recycle that room.")
		case Typeof(ref) == TYPE_PLAYER:
			panic("Cannot recycle a player.")
		case ref == program:
			panic("Cannot recycle currently running program.")
		case Typeof(ref) == TYPE_EXIT && !unset_source(player, db.Fetch(player).location, ref) {
			panic("Cannot recycle old style exits.")
		}
		for i := 0; i < fr.caller.top; i++ {
			if fr.caller.st[i] == ref {
				panic("Cannot recycle active program.")
			}
		}
		recycle(fr.descr, player, ref)
	})
}

func prim_setlockstr(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		ref := valid_object(op[0])
		if mlev < WIZBIT && !permissions(ProgUID, ref) {
			panic("Permission denied.")
		}
		push(arg, top, setlockstr(fr.descr, player, ref, op[1].(string)))
	})
}

func prim_getlockstr(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		ref := op[0].(dbref)
		check_remote(player, ref)
		if mlev < MASTER && !permissions(ProgUID, ref) {
			panic("Permission denied.")
		}
		push(arg, top, get_property_lock(ref, MESGPROP_LOCK).Unparse(player, false))
	})
}

func prim_part_pmatch(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		push(arg, top, partial_pmatch(POP().data.(string)))
	})
}


func prim_checkpassword(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 2, top, func(op Array) {
		push(arg, top, check_password(valid_player(op[0]), op[1].(string)))
	})
}

func prim_movepennies(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(tp_movepennies_muf_mlev, mlev, 3, top, func(op Array) {
		donor := valid_object(op[0])
		recipient := valid_object(op[1])
		pennies := op[2].(int)
		if pennies < 0 {
			panic("Invalid argument. (3)")
		}
		if mlev < WIZBIT {
			switch {
			case Typeof(donor) != TYPE_PLAYER:
				panic("Permission denied. (2)")
			case get_property_value(donor, MESGPROP_VALUE) < get_property_value(donor, MESGPROP_VALUE) - pennies:
				panic("Would roll over player's score. (1)")
			case get_property_value(donor, MESGPROP_VALUE) - pennies < 0:
				panic("Result would be negative. (1)")
			case Typeof(recipient) != TYPE_PLAYER:
				panic("Permission denied. (2)")
			case get_property_value(recipient, MESGPROP_VALUE) > get_property_value(recipient, MESGPROP_VALUE) + pennies:
				panic("Would roll over player's score. (2)")
			case get_property_value(recipient, MESGPROP_VALUE) + pennies > tp_max_pennies:
				panic("Would exceed MAX_PENNIES. (2)")
			}
		}
		add_property(donor, MESGPROP_VALUE, nil, get_property_value(donor, MESGPROP_VALUE) - pennies)
		db.Fetch(donor).flags |= OBJECT_CHANGED
		add_property(recipient, MESGPROP_VALUE, nil, get_property_value(recipient, MESGPROP_VALUE) + pennies)
		db.Fetch(recipient).flags |= OBJECT_CHANGED
	})
}


func prim_findnext(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(JOURNEYMAN, mlev, 4, top, func(op Array) {
		item := op[0].(dbref)
		owner := op[1].(dbref)
		name := op[2].(string)
		flags := op[3].(string)
		switch {
		case item < NOTHING, item >= db_top:
			panic("Bad object. (1)")
		case owner < NOTHING, owner >= db_top:
			panic("Bad object. (2)")
		case mlev < MASTER && owner == NOTHING:
			panic("Permission denied.  Owner inspecific searches require Mucker Level 3.")
		case mlev < MASTER && owner != ProgUID:
			panic("Permission denied.  Searching for other people's stuff requires Mucker Level 3.")
		case item == NOTHING:
			item = 0
		default:
			item++
		}
		buf := name
		ref := NOTHING
		_, check := init_checkflags(player, flags)
		for i := item; i < db_top && ref == NOTHING; i++ {
			if (owner == NOTHING || db.Fetch(i).owner == owner) && checkflags(i, check) && db.Fetch(i).name && (name == "" || !smatch(buf, db.Fetch(i).name)) {
				ref = i
			}
		}
		push(arg, top, ref)
	})
}

/* ============================ */
/* = More ProtoMuck prims     = */
/* ============================ */

func prim_nextentrance(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 0, top, func(op Array) {
		linkref := valid_object_or_home(op[0])
		ref := valid_object(op[1])
		if linkref == HOME {
			linkref = db.Fetch(player).sp.(player_specific).home
		}
		var foundref bool
		for ref++; ref < db_top && !foundref ; ref++ {
			switch Typeof(valid_object(ref)) {
			case TYPE_PLAYER:
				foundref = db.Fetch(ref).sp.(player_specific).home == linkref
			case TYPE_ROOM:
				foundref = db.Fetch(ref).sp == linkref
			case TYPE_THING:
				foundref = db.Fetch(ref).sp.(player_specific).home == linkref
			case TYPE_EXIT:
				count := len(db.Fetch(ref).sp.exit.dest)
				for i := 0; i < count && !foundref; i++ {
					foundref = db.Fetch(ref).sp.exit.dest[i] == linkref
				}
			}
		}
		if !foundref {
			ref = NOTHING
		}
		push(arg, top, ref)
	})
}

func prim_newplayer(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 2, top, func(op Array) {
		name := op[0].(string)
		password := op[1].(string)
		switch {
		case !ok_player_name(name):
			panic("Invalid player name. (1)")
		case !ok_password(password):
			panic("Invalid password. (1)")
		}
		newplayer := create_player(name, password)
		log_status("PCREATED[MUF]: %s(%d) by %s(%d)", db.Fetch(newplayer).name, newplayer, db.Fetch(player).name, player)
		push(arg, top, newplayer)
	})
}

func prim_copyplayer(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 3, top, func(op Array) {
		ref := valid_remote_player(player, mlev, op[0])
		name := op[1].(string)
		password := op[2].(string)
		switch {
		case !ok_player_name(name):
			panic("Invalid player name. (2)")
		case !ok_password(password):
			panic("Invalid password. (2)")
		}

		/* else he doesn't already exist, create him */
		newplayer := create_player(name, password)

		/* initialize everything */
		db.Fetch(newplayer).flags = db.Fetch(ref).flags

		newp := db.Fetch(newplayer)
		newp.properties = copy_prop(ref)
		newp.exits = NOTHING
		newp.contents = NOTHING
		newp.next = NOTHING
		db.Fetch(newplayer).sp.(player_specific).home = db.Fetch(ref).sp.(player_specific).home
		add_property(newplayer, MESGPROP_VALUE, nil, get_property_value(newplayer, MESGPROP_VALUE) + get_property_value(ref, MESGPROP_VALUE))
		moveto(newplayer, db.Fetch(ref).sp.(player_specific).home)

		/* link him to player_start */
		log_status("PCREATE[MUF]: %s(%d) by %s(%d)", db.Fetch(newplayer).name, newplayer, db.Fetch(player).name, player)
    	push(arg, top, newplayer)
	})
}

func prim_toadplayer(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 2, top, func(op Array) {
		recipient := valid_remote_player(player, mlev, op[0])
		victim := valid_remote_player(player, mlev, op[1])
		switch {
		case victim == recipient:
			panic("Victim and recipient must be different players.")
		case get_property_class( victim, "@/precious"):
			panic("That player is precious.")
		case victim == GOD:
			panic("God may not be toaded. (2)")
		case db.Fetch(victim).flags & WIZARD != 0:
			panic("You can't toad a wizard.")
		default:
			send_contents(fr->descr, victim, HOME);
			for stuff := 0; stuff < db_top; stuff++ {
			    if db.Fetch(stuff).owner == victim {
					switch Typeof(stuff) {
					case TYPE_PROGRAM:
						dequeue_prog(stuff, 0)  /* dequeue player's progs */
						db.Fetch(stuff).flags &= ~(ABODE | WIZARD)
						SetMLevel(stuff, NON_MUCKER)
						db.Fetch(stuff).owner = recipient
						db.Fetch(stuff).flags |= OBJECT_CHANGED
					case TYPE_ROOM, TYPE_THING, TYPE_EXIT:
						db.Fetch(stuff).owner = recipient
						db.Fetch(stuff).flags |= OBJECT_CHANGED
					}
			    }
			    if Typeof(stuff) == TYPE_THING && db.Fetch(stuff).sp.(player_specific).home == victim {
					db.Fetch(stuff).sp.(player_specific).home = tp_player_start
			    }
			}
			db.Fetch(victim).sp.(player_specific).password = ""
			dequeue_prog(victim, 0);  /* dequeue progs that player's running */

			log_status("TOADED[MUF]: %s(%d) by %s(%d)", db.Fetch(victim).name, victim, db.Fetch(player).name, player)

			delete_player(victim)
			db.Fetch(victim).name = fmt.Sprintf("A slimy toad named %s", db.Fetch(victim).name)
			db.Fetch(victim).flags |= OBJECT_CHANGED
			boot_player_off(victim)

			db.Fetch(victim).sp.(player_specific).descrs = nil
	
			ignore_remove_from_all_players(victim)
			ignore_flush_cache(victim)

			db.Fetch(victim).sp.(player_specific) = new(player_specific)
			db.Fetch(victim).sp.(player_specific).home = db.Fetch(recipient).sp.(player_specific).home

			/* reset name */
			db.Fetch(victim).flags = (db.Fetch(victim).flags & ~TYPE_MASK) | TYPE_THING
			db.Fetch(victim).owner = recipient
			add_property(victim, MESGPROP_VALUE, nil, 1)
		}
	})
}

func prim_instances(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		ref := valid_object(op[0])
		if Typeof(ref) != TYPE_PROGRAM {
			panic("Object must be a program.")
		}
		if p := db.Fetch(ref).sp.program; p.sp != nil {
			push(arg, top, p.sp.instances)
		} else {
			push(arg, top, 0)
		}
	})
}

func prim_compiledp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		ref := valid_object(op[0])
		if Typeof(ref) != TYPE_PROGRAM {
			panic("Object must be a program.")
		}
		push(arg, top, len(db.Fetch(ref).sp.(program_specific).code))
	})
}

func prim_newpassword(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 2, top, func(op Array) {
		switch obj := valid_remote_player(player, mlev, op[0]); {
		case obj == GOD:
			panic("God cannot be newpassworded.")
		case player != GOD && TrueWizard(obj) && player != obj:
			panic("Only God can change a wizards password")
		}
		set_password(obj, op[1].(string))
	})
}

func prim_newprogram(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 1, top, func(op Array) {
		name := op[0].(string)
		if !ok_ascii_other(name) || !ok_name(name) {
			panic("Invalid name (2)")
		}
		newprog := new_object()
		db.Fetch(newprog).name = name
		add_property(newprog, MESGPROP_DESC, fmt.Sprintf("A scroll containing a spell called %s", name), 0)
		db.Fetch(newprog).location = player
		db.Fetch(newprog).flags = TYPE_PROGRAM
		l := MLevel(player)
		switch {
		case l < APPRENTICE:
	    	l = APPRENTICE
		case l > MASTER:
	    	l = MASTER
		}
		SetMLevel(newprog, APPRENTICE)
		db.Fetch(newprog).owner = db.Fetch(player).owner
		db.Fetch(newprog).sp.(program_specific) = new(program_specific)
		db.Fetch(player).sp.(player_specific).curr_prog = newprog
		db.Fetch(newprog).next = db.Fetch(player).contents
		db.Fetch(newprog).flags |= OBJECT_CHANGED
		db.Fetch(player).contents = newprog
		db.Fetch(newprog).flags |= OBJECT_CHANGED
		db.Fetch(player).flags |= OBJECT_CHANGED
		push(arg, top, newprog)
	})
}

func prim_compile(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 2, top, func(op Array) {
		obj := valid_object(op[0])
		display_errors := op[1].(int)
		var i int
		p := db.Fetch(obj).sp.program
		if p.sp != nil {
			i = p.sp.instances
		}
		switch {
		case Typeof(obj) != TYPE_PROGRAM:
			panic("No program dbref given. (1)")
		case i > 0:
			panic("That program is currently in use.")
		}
		tmpline := p.sp.first
		p.sp.first = read_program(obj)
		do_compile(fr.descr, player, obj, display_errors)
		p.sp.first = tmpline
		push(arg, top, len(p.sp.code))
	})
}

func prim_uncompile(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 1, top, func(op Array) {
		obj := valid_object(op[0])
		var i int
		if p := db.Fetch(obj).sp.program; p.sp != nil {
			i = p.sp.instances
		}
		switch {
		case Typeof(obj) != TYPE_PROGRAM:
			panic("No program dbref given.")
		case i > 0:
			panic("That program is currently in use.")
		}
		uncompile_program(obj)
	})
}

func prim_getpids(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		obj := op[0].(dbref)
		nw := get_pids(obj)
		if program == obj {
			array_appenditem(&nw, &inst{ data: fr.pid })
		}
		push(arg, top, nw)
	})
}

func prim_getpidinfo(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		pid := op[0].(int)
		if pid == fr.pid {
			var cpu float64
			var etime time_t
			if etime = time(NULL) - fr.started; etime > 0 {
				cpu = ((fr.totaltime.tv_sec + (fr.totaltime.tv_usec / 1000000.0)) * 100.0) / etime
				if cpu > 100.0 {
					cpu = 100.0
				}
			} else {
				cpu = 0.0f
			}
			push(arg, top, Dictionary{
				"PID": fr.pid,
				"INSTCNT": fr.instcnt,
				"DESCR": fr.descr,
				"NEXTRUN": 0,
				"STARTED": fr.started,
				"CALLED_PROG": program,
				"TRIG": fr.trig,
				"PLAYER": player,
				"CPU": cpu,
				"CALLED_DATA": "",
				"TYPE": "MUF",
				"SUBTYPE": "",
			})
		} else {
			push(arg, top, get_pidinfo(pid))
		}
	})
}

func prim_contents_array(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		ref := valid_remote_object(player, mlev, op[0])
		var nu Array
		switch Typeof(ref) {
		case TYPE_PROGRAM, TYPE_EXIT:
		default:
			var count int
			for ref = db.Fetch(oper1.data.objref).contents; (ref >= 0) && (ref < db_top); ref = db.Fetch(ref).next {
				count++
			}
			nw = make(stk_array, count)
			for ref = db.Fetch(oper1.data.objref).contents, count = 0; (ref >= 0) && (ref < db_top); ref = db.Fetch(ref).next {
				nw[count] = ref
				count++
			}
		}
		push(arg, top, nw)
	})
}

func prim_exits_array(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		ref := valid_remote_object(player, mlev, op[0])
		if mlev < MASTER && !permissions(ProgUID, ref) {
			panic("Permission denied.")
		}
		var nu Array
		switch Typeof(ref) {
		case TYPE_PROGRAM, TYPE_EXIT:
		default:
			var count int
			for ref = db.Fetch(oper1.data.objref).exits; ref >= 0 && ref < db_top; ref = db.Fetch(ref).next {
				count++
			}
			nw = make(Array, count)
			for ref = db.Fetch(oper1.data.objref).exits, count = 0; ref >= 0 && ref < db_top; ref = db.Fetch(ref).next {
				nw[count] = ref
				count++
			}
		}
		push(arg, top, nw)
	})
}

func prim_getlinks_array(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		ref := valid_remote_object(player, mlev, op[0])
		var nw Array
		switch Typeof(ref) {
		case TYPE_ROOM:
			push(arg, top, Array{ db.Fetch(ref).sp.(dbref) })
		case TYPE_THING:
			push(arg, top, Array{ db.Fetch(ref).sp.(player_specific).home })
		case TYPE_PLAYER:
			push(arg, top, Array{ db.Fetch(ref).sp.(player_specific).home })
		case TYPE_EXIT:
			obj := db.Fetch(ref)
			nw := make(Array, len(obj.sp.exit.dest))
			copy(nw, sp.exit.dest)
			push(arg, top, nw)
		}
	})
}

func prim_entrances_array(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		ref := valid_object(op[0])
    	var nw Array
		for i := 0; i < db_top; i++ {
        	switch Typeof(i) {
           	case TYPE_EXIT:
				nw = append(nw, db.Fetch(i).sp.exit.dest...)
            case TYPE_PLAYER:
                if db.Fetch(i).sp.(player_specific).home == ref {
					nw = append(nw, i)
				}
            case TYPE_THING:
                if db.Fetch(i).sp.(player_specific).home == ref {
					nw = append(nw, i)
				}
            case TYPE_ROOM:
                if db.Fetch(i).sp == ref {
					nw = append(nw, i)
				}
			}
    	}
		push(arg, top, nw)
	})
}

func prim_program_getlines(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		ref := valid_object(op[0])
		start := op[1].(int)
		end := op[2].(int)
		switch {
		case Typeof(ref) != TYPE_PROGRAM:
			panic("Non-program object. (1)")
		case mlev < WIZBIT && !controls(ProgUID, ref) && db.Fetch(ref).flags & VEHICLE == 0:
			panic("Permission denied.")
		case start < 0, end < 0:
			panic("Line indexes must be non-negative.")
		}
		if start == 0 {
			start = 1
		}
		if end > 0 && start > end {
			panic("Illogical line range.")
		}

		/* we make two passes over our linked list's data,
		 * first we figure out how many lines are
		 * actually there. This is so we only allocate
		 * our array once, rather re-allocating 4000 times
		 * for a 4000-line program listing, while avoiding
		 * letting calls like '#xxx 1 999999 program_getlines'
		 * taking up tons of memory.
		 */
		first := read_program(ref)
		curr := first

		/* find our line */
		var ary Array
		var i int
		for i = 1; curr != nil && i < start; i++ {
			curr = curr.next
		}
		if curr != nil {
			segment := curr	/* we need to keep this line */
			/* continue our looping */
			for ; curr && (end == 0 || i < end); i++ {
				curr = curr.next
			}
			count := i - start + 1
		
			if curr == nil {
				/* if we don't have curr, we counted one beyond the end of the program, so we account for that. */
			    count--
			}
			ary = make(Array, 0, count)
			for curr = segment; count > 0; curr = curr.next {
				ary = append(ary, curr.this_line)
				count--
			}
		}
		push(arg, top, ary)
	})
}

func prim_program_setlines(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 2, top, func(op Array) {
		obj := valid_object(op[0])
		listing := op[1].(Array)

		switch {
		case Typeof(obj) != TYPE_PROGRAM:
			panic("Non-program object. (1)")
		case !array_is_homogenous(listing, ""):
			panic("Argument not an array of strings. (2)")
		case !controls(ProgUID, obj) {
			panic("Permission denied.")
		case db.Fetch(obj).flags & INTERNAL != 0:
			panic("Program already being edited.")
		}

		var lines, prev *line
		for _, v := range listing {
			ln := new(line)
			if _, ok  := v.data.(string); ok {
				ln.this_line = v.data.(string)
			}

			if prev != nil {
				prev.next = ln
				ln.prev	= prev
			} else {
				lines = ln
			}
			prev = ln
		}

		write_program(lines, obj)
		log_status("PROGRAM SAVED: %s by %s(%d)", unparse_object(player, obj), db.Fetch(player).name, player)
		if tp_log_programs {
			log_program_text(lines, player, obj)
		}
		db.Fetch(program).flags |= OBJECT_CHANGED
	})
}

func prim_setlinks_array(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		source := valid_object(op[0])
		links := op[1].(Array)
		switch dest_count := len(links); {
		case !array_is_homogenous(links, dbref(0)):
			panic("Argument not an array of dbrefs. (2)")
		case mlev < WIZBIT && !permissions(ProgUID, source):
			panic("Permission denied. (1)")
		case dest_count > 1 && Typeof(source) != TYPE_EXIT:
			panic("Only exit may be linked to multiple destinations.")
		default:
			var found_prp bool
			for _, v := range links {
				where := valid_object_or_home(v.(dbref))
				if !prog_can_link_to(mlev, ProgUID, Typeof(source), where) {
					pamic("Can't link source to destination. (2)")
				}

				switch source := source.(type) {
				case TYPE_EXIT:
					switch where := where.(type) {
					case TYPE_THING:
					case TYPE_PLAYER, TYPE_ROOM, TYPE_PROGRAM:
						if found_prp {
							panic("Only one player, room, or program destination allowed.")
						}
						found_prp = true
					case TYPE_EXIT:
						if exit_loop_check(source, where) {
							panic("Destination would create loop.")
						}
					default:
						panic("Invalid object. (2)")
					}
				case TYPE_PLAYER:
					if where == HOME {
						panic("Cannot link player to HOME.")
					}
				case TYPE_THING:
					switch {
					case where == HOME:
						panic("Cannot link thing to HOME.")
					case parent_loop_check(source, where):
						panic("That would case a parent paradox.")
					}
				case TYPE_ROOM:
				default:
					panic("Invalid object. (1)")
				}
			}

			if source, ok := what.(TYPE_EXIT); ok {
				if MLevRaw(source) != NON_MUCKER {
					SetMLevel(source, NON_MUCKER)
				}
				db.Fetch(source).sp.exit.dest = nil
			}

			if dest_count < 1 {
				switch source := source.(type) {
				case TYPE_EXIT:
					db.Fetch(source).sp.exit.dest = nil
					db.Fetch(source).flags |= OBJECT_CHANGED
				case TYPE_ROOM:
					db.Fetch(source).sp = NOTHING
					db.Fetch(source).flags |= OBJECT_CHANGED
				default:
					panic("Only exits and rooms may be linked to nothing. (1)")
				}
			} else {
				switch source := source.(type) {
				case TYPE_EXIT:
					dests := make([]dbref, dest_count, dest_count)
					for i, v := range links {
						dests[i] = v.(dbref)
					}
					db.Fetch(source).sp.exit.dest = dests
					db.Fetch(source).flags |= OBJECT_CHANGED
				case TYPE_ROOM:
					db.Fetch(source).sp = links[0].(dbref)
					db.Fetch(source).flags |= OBJECT_CHANGED
				case TYPE_PLAYER:
					db.Fetch(source).sp.(player_specific).home = links[0].(dbref)
				case TYPE_THING:
					db.Fetch(source).sp.(player_specific).home = links[0].(dbref)
				default:
					panic("Invalid object. (1)")
				}
			}
			db.Fetch(source).flags |= OBJECT_CHANGED
		}
	})
}