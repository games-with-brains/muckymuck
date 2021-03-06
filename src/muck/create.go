package fbmuck

/* parse_linkable_dest()
 *
 * A utility for open and link which checks whether a given destination
 * string is valid.  It returns a parsed ObjectID on success, and NOTHING
 * on failure.
 */

func parse_linkable_dest(int descr, ObjectID player, ObjectID exit, const char *dest_name) (r ObjectID) {
	r = NewMatch(descr, player, dest_name, NOTYPE).
		MatchAbsolute().
		MatchEverything().
		MatchHome().
		MatchResult()
	switch {
	case r == NOTHING, r == AMBIGUOUS:
		notify(player, fmt.Sprintf("I couldn't find '%s'.", dest_name))
		r = NOTHING
	case !tp_teleport_to_player && Typeof(r) == TYPE_PLAYER:
		notify(player, fmt.Sprintf("You can't link to players.  Destination %s ignored.", unparse_object(player, r)))
		r = NOTHING
	case !can_link(player, exit):
		notify(player, "You can't link that.")
		r = NOTHING
	case !can_link_to(player, Typeof(exit), r):
		notify(player, fmt.Sprintf("You can't link to %s.", unparse_object(player, r)))
		r = NOTHING
	}
	return
}

/* exit_loop_check()
 *
 * Recursive check for loops in destinations of exits.  Checks to see
 * if any circular references are present in the destination chain.
 * Returns true if circular reference found, false if not.
 */
func exit_loop_check(ObjectID source, ObjectID dest) (r bool) {
	switch {
	case Typeof(dest) != TYPE_EXIT:
	case source == dest:
		r = true
	default:
		for _, v := range DB.Fetch(dest).(Exit).Destinations {
			if v == source || (Typeof(v) == TYPE_EXIT && exit_loop_check(source, v)) {
				r = true
				break
			}
		}
	}
	return
}

/* use this to create an exit */
func do_open(descr int, player ObjectID, direction, linkto string) {
	NoGuest("@open", player, func() {
		switch {
		case !Builder(player):
			notify(player, "That command is restricted to authorized builders.")
		case direction == "":
			notify(player, "You must specify a direction or action name to open.")
		case !ok_name(direction):
			notify(player, "That's a strange name for an exit!")
		case !payfor(player, tp_exit_cost):
			notify_fmt(player, "Sorry, you don't have enough %s to open an exit.", tp_pennies)
		default:
			var qname, rname string
			terms := strings.SplitN(linkto, "=", 2)
			if len(terms) == 2 {
				qname = strings.TrimFunc(terms[0], unicode.IsSpace)
				rname = strings.TrimFunc(terms[1], unicode.IsSpace)
			}

			if loc := DB.Fetch(player).Location; {
			case loc == NOTHING:
			case !controls(player, loc):
				notify(player, "Permission denied. (you don't control the location)")
			default:
				/* create the exit */
				exit := new_object()

				/* initialize everything */
				DB.Fetch(exit).NowCalled(direction)
				DB.Fetch(exit).MoveTo(loc)
				DB.Fetch(exit).GiveTo(DB.Fetch(player).Owner)
				DB.Fetch(exit).flags = TYPE_EXIT
				DB.Fetch(exit).(Exit).Destinations = nil

				/* link it in */
				DB.Fetch(exit).next = DB.Fetch(loc).Exits
				DB.Fetch(exit).Touch()
				DB.Fetch(loc).Exits = exit
				DB.Fetch(loc§).Touch()

				/* and we're done */
				notify(player, fmt.Sprintf("Exit opened with number %d.", exit))

				/* check second arg to see if we should do a link */
				if qname != "" {
					notify(player, "Trying to link...")
					if !payfor(player, tp_link_cost) {
						notify_fmt(player, "You don't have enough %s to link.", tp_pennies)
					} else {
						var good_dest []ObjectID
						ndest := link_exit(descr, player, exit, qname, good_dest)
						DB.Fetch(exit).(Exit).Destinations = good_dest
						DB.Fetch(exit).Touch()
					}
				}

				if rname != "" {
					notify(player, fmt.Sprintf("Registered as $%s", rname))
					set_property(player, fmt.Sprintf("_reg/%s", rname), exit)
				}
			}
		}
	})
}

func _link_exit(descr int, player, exit ObjectID, name string, dest_list []ObjectID, dryrun bool) (r int) {
	var prdest, error bool
	for name != "" {
		name = strings.TrimSpace(name)
		p := name
		for name != "" && name != EXIT_DELIMITER {
			name = name[1:]
		}
		q := p[:len(p) - len(name)]
		name = strings.TrimSpace(name)

		if dest := parse_linkable_dest(descr, player, exit, q); dest != NOTHING {
			switch Typeof(dest) {
			case Player, Room, Program:
				if prdest {
					notify(player, fmt.Sprintf("Only one player, room, or program destination allowed. Destination %s ignored.", unparse_object(player, dest)))
					if dryrun {
						error = true
					}
					continue
				}
				dest_list[r] = dest
				r++
				prdest = true
			case Object:
				dest_list[r] = dest
				r++
			case Exit:
				if exit_loop_check(exit, dest) {
					notify(player, fmt.Sprintf("Destination %s would create a loop, ignored.", unparse_object(player, dest)))
					if dryrun {
						error = true
					}
					continue
				}
				dest_list[r] = dest
				r++
			default:
				notify(player, "Internal error: weird object type.")
				log_status("PANIC: weird object: Typeof(%d) = %d", dest, Typeof(dest))
				if dryrun {
					error = true
				}
			}
			if !dryrun {
				if dest == HOME {
					notify(player, "Linked to HOME.")
				} else {
					notify(player, fmt.Sprintf("Linked to %s.", unparse_object(player, dest)))
				}
			}
		}
	}
	if dryrun && error {
		r = 0
	}
	return
}

/*
 * link_exit()
 *
 * This routine connects an exit to a bunch of destinations.
 *
 * 'player' contains the player's name.
 * 'exit' is the the exit whose destinations are to be linked.
 * 'dest_name' is a character string containing the list of exits.
 *
 * 'dest_list' is an array of ObjectID's where the valid destinations are
 * stored.
 *
 */

int
link_exit(int descr, ObjectID player, ObjectID exit, char *dest_name, ObjectID * dest_list)
{
	return _link_exit(descr, player, exit, dest_name, dest_list, 0);
}

/*
 * link_exit_dry()
 *
 * like link_exit(), but only checks whether the link would be ok or not.
 * error messages are still output.
 */
func link_exit_dry(int descr, ObjectID player, ObjectID exit, char *dest_name, ObjectID * dest_list) int {
	return _link_exit(descr, player, exit, dest_name, dest_list, 1);
}

/* do_link
 *
 * Use this to link to a room that you own.  It also sets home for
 * objects and things, and drop-to's for rooms.
 * It seizes ownership of an unlinked exit, and costs 1 penny
 * plus a penny transferred to the exit owner if they aren't you
 *
 * All destinations must either be owned by you, or be LINK_OK.
 */
func do_link(descr int, player ObjectID, thing_name, dest_name string) {
	NoGuest("@link", player, func() {
		md := NewMatch(descr, player, thing_name, IsExit).
			MatchAllExits().
			MatchNeighbor().
			MatchPossession().
			MatchMe().
			MatchHere().
			MatchAbsolute().
			MatchRegistered()
		if Wizard(DB.Fetch(player).Owner) {
			md.MatchPlayer()
		}
		if thing := md.NoisyMatchResult(); thing != NOTHING {
			switch thing.(type) {
			case TYPE_EXIT:
				/* we're ok, check the usual stuff */
				switch {
				case len(DB.Fetch(thing).(Exit).Destinations) != 0:
					if controls(player, thing) {
						notify(player, "That exit is already linked.")
					} else {
						notify(player, "Permission denied. (you don't control the exit to relink)")
					}
				case DB.Fetch(thing).Owner == DB.Fetch(player).Owner && !payfor(player, tp_link_cost):
					if tp_link_cost == 1 {
						notify_fmt(player, "It costs %d %s to link this exit.", tp_link_cost, tp_penny)
					} else {
						notify_fmt(player, "It costs %d %s to link this exit.", tp_link_cost, tp_pennies)
					}
				case !payfor(player, tp_link_cost + tp_exit_cost):
					if tp_link_cost + tp_exit_cost == 1 {
						notify_fmt(player, "It costs %d %s to link this exit.", (tp_link_cost + tp_exit_cost), tp_penny)
					} else {
						notify_fmt(player, "It costs %d %s to link this exit.", (tp_link_cost + tp_exit_cost), tp_pennies)
					}
				case !Builder(player):
					notify(player, "Only authorized builders may seize exits.")
				default:
					owner := DB.Fetch(thing).Owner
					add_property(owner, MESGPROP_VALUE, nil, get_property_value(owner, MESGPROP_VALUE) + tp_exit_cost)
					DB.Fetch(owner).Touch()
					DB.Fetch(thing).GiveTo(DB.Fetch(player).Owner)
					var good_dest []ObjectID
					if n := link_exit(descr, player, thing, dest_name, good_dest); n == 0 {
						notify(player, "No destinations linked.")
						add_property(player, MESGPROP_VALUE, nil, get_property_value(player, MESGPROP_VALUE) + tp_link_cost)
						DB.Fetch(player).Touch()
					} else {
						DB.Fetch(thing).(Exit).Destinations = good_dest
						DB.Fetch(thing).Touch()
					}
				}
			case Object, Player:
				md := NewMatch(descr, player, dest_name, IsRoom).
					MatchNeighbor().
					MatchAbsolute().
					MatchRegistered().
					MatchMe().
					MatchHere()
				if IsThing(thing) {
					md.MatchPossession()
				}
				switch dest := md.NoisyMatchResult(); {
				case dest == NOTHING:
				case !controls(player, thing), !can_link_to(player, Typeof(thing), dest):
					notify(player, "Permission denied. (you don't control the thing, or you can't link to dest)")
				case parent_loop_check(thing, dest):
					notify(player, "That would cause a parent paradox.")
				default:
					p := DB.Fetch(thing)
					p.LiveAt(dest)
					p.Touch()
					notify(player, "Home set.")
				}
			case TYPE_ROOM:			/* room dropto's */
				dest = NewMatch(descr, player, dest_name, IsRoom).
					MatchNeighbor().
					MatchPossession().
					MatchRegistered().
					MatchAbsolute().
					MatchHome().
					NoisyMatchResult()
				switch {
				case dest == NOTHING:
				case !controls(player, thing), !can_link_to(player, Typeof(thing), dest), thing == dest:
					notify(player, "Permission denied. (you don't control the room, or can't link to the dropto)")
				default:
					DB.Store(thing, dest)
					notify(player, "Dropto set.")
					DB.Fetch(thing).Touch()
				}
			case TYPE_PROGRAM:
				notify(player, "You can't link programs to things!")
			default:
				notify(player, "Internal error: weird object type.")
				log_status("PANIC: weird object: Typeof(%d) = %d", thing, Typeof(thing))
			}
		}
	})
}

/*
 * do_dig
 *
 * Use this to create a room.
 */
func do_dig(descr int, player ObjectID, name, pname string) {
	NoGuest("@dig", player, func() {
		switch {
		case !Builder(player):
			notify(player, "That command is restricted to authorized builders.")
		case name == "":
			notify(player, "You must specify a name for the room.")
		case !ok_ascii_other(name):
			notify(player, "Room names are limited to 7-bit ASCII.")
		case !ok_name(name):
			notify(player, "That's a silly name for a room!")
		case !payfor(player, tp_room_cost):
			notify_fmt(player, "Sorry, you don't have enough %s to dig a room.", tp_pennies)
		default:
			room := new_object()

			/* Initialize everything */
			newparent := DB.Fetch(DB.Fetch(player).Location).Location
			for newparent != NOTHING && !DB.Fetch(newparent).IsFlagged(ABODE) {
				newparent = DB.Fetch(newparent).Location
			}
			if newparent == NOTHING {
				newparent = tp_default_room_parent
			}

			DB.Fetch(room).NowCalled(name)
			DB.Fetch(room).MoveTo(newparent)
			DB.Fetch(room).GiveTo(DB.Fetch(player).Owner)
			DB.Fetch(room).Exits = NOTHING
			DB.Fetch(room).sp = NOTHING
			if DB.Fetch(player).IsFlagged(JUMP_OK) {
				DB.Fetch(room).FlagAs(JUMP_OK)
			}
			DB.Fetch(room).flags = TYPE_ROOM
			DB.Fetch(room).next = DB.Fetch(newparent).Contents
			DB.Fetch(room).Touch()
			DB.Fetch(newparent).Contents = room

			DB.Fetch(room).Touch()
			DB.Fetch(newparent).Touch()

			notify(player, fmt.Sprintf("%s created with room number %d.", name, room))

			var qname, rname string
			if terms := strings.SplitN(pname, "=", 2); len(terms) == 2 {
				qname = strings.TrimFunc(terms[0], unicode.IsSpace)
				rname = strings.TrimFunc(terms[1], unicode.IsSpace)
			}

			if qname != "" {
				notify(player, "Trying to set parent...")
				parent := NewMatch(descr, player, qname, IsRoom).
					MatchAbsolute().
					MatchRegistered().
					MatchHere().
					NoisyMatchResult()
				switch {
				case parent == NOTHING, parent == AMBIGUOUS:
					notify(player, "Parent set to default.");
				case !can_link_to(player, Typeof(room), parent), room == parent:
					notify(player, "Permission denied.  Parent set to default.")
				default:
					moveto(room, parent)
					notify(player, fmt.Sprintf("Parent set to %s.", unparse_object(player, parent)))
				}
			}
			if rname != "" {
				set_property(player, fmt.Sprintf("_reg/%s", rname), room)
				notify(player, fmt.Sprintf("Room registered as $%s", rname))
			}
		}
	})
}

/*
  Use this to create a program.
  First, find a program that matches that name.  If there's one,
  then we put him into edit mode and do it.
  Otherwise, we create a new object for him, and call it a program.
  */
func do_prog(descr int, player ObjectID, name string) {
	NoGuest("@program", player, func() {
		switch {
		case Typeof(player) != TYPE_PLAYER:
			notify(player, "Only players can edit programs.")
		case !Mucker(player):
			notify(player, "You're no programmer!")
		case name == "":
			notify(player, "No program name given.")
		default:
			i := NewMatch(descr, player, name, IsProgram).
				MatchPossession().
				MatchNeighbor().
				MatchRegistered().
				MatchAbsolute().
				MatchResult()
			switch {
			case i == NOTHING:
				newprog := new_object()
				DB.Fetch(newprog).NowCalled(name)
				add_property(newprog, MESGPROP_DESC, fmt.Sprintf("A scroll containing a spell called %s", name), 0)
				DB.Fetch(newprog).MoveTo(player)
				DB.Fetch(newprog).flags = TYPE_PROGRAM
				mlev := MLevel(player)
				switch {
				case mlev < APPRENTICE:
					mlev = JOURNEYMAN
				case mlev > MASTER:
					mlev = MASTER
				}
				SetMLevel(newprog, mlev)

				DB.Fetch(newprog).GiveTo(DB.Fetch(player).Owner)
				DB.Fetch(newprog).(Program) = new(Program)
				DB.FetchPlayer(player).curr_prog = newprog

				DB.Fetch(newprog).next = DB.Fetch(player).Contents
				DB.Fetch(newprog).Touch()
				DB.Fetch(player).Contents = newprog
				DB.Fetch(newprog).Touch()

				DB.Fetch(player).FlagAs(INTERACTIVE)
				DB.Fetch(player).Touch()
				notify(player, fmt.Sprintf("Program %s created with number %d.", name, newprog))
				notify(player, fmt.Sprintf("Entering editor."))
			case i == AMBIGUOUS:
				notify(player, "I don't know which one you mean!")
			case !IsProgram(i), !controls(player, i):
				notify(player, "Permission denied!")
			case DB.Fetch(i).IsFlagged(INTERNAL):
				notify(player, "Sorry, this program is currently being edited by someone else.  Try again later.")
			} else {
				DB.Fetch(i).(Program).first = read_program(i)
				DB.Fetch(i).FlagAs(INTERNAL)
				DB.FetchPlayer(player).curr_prog = i
				notify(player, "Entering editor.")
				/* list current line */
				do_list(player, i, nil)
				DB.Fetch(i).Touch()
				DB.Fetch(player).FlagAs(INTERACTIVE)
				DB.Fetch(player).Touch()
			}
		}
	})
}

func do_edit(descr int, player ObjectID, name string) {
	NoGuest("@edit", player, func() {
		switch {
		case Typeof(player) != TYPE_PLAYER:
			notify(player, "Only players can edit programs.")
		case !Mucker(player):
			notify(player, "You're no programmer!")
		case name == "":
			notify(player, "No program name given.")
		default:
			i := NewMatch(descr, player, name, IsProgram).
				MatchPossession().
				MatchNeighbor().
				MatchRegistered().
				MatchAbsolute().
				NoisyMatchResult()
			switch {
			case i == NOTHING, i == AMBIGUOUS:
			case Typeof(i) != TYPE_PROGRAM, !controls(player, i):
				notify(player, "Permission denied!")
			case DB.Fetch(i).IsFlagged(INTERNAL):
				notify(player, "Sorry, this program is currently being edited by someone else.  Try again later.")
			default:
				DB.Fetch(i).FlagAs(INTERNAL)
				DB.Fetch(i).(Program).first = read_program(i)
				DB.FetchPlayer(player).curr_prog = i
				notify(player, "Entering editor.")
				/* list current line */
				do_list(player, i, nil)
				DB.Fetch(player).FlagAs(INTERACTIVE)
				DB.Fetch(i).Touch()
				DB.Fetch(player).Touch()
			}
		}
	})
}

func do_mcpedit(descr int, player ObjectID, name string) {
	NoGuest("@mcpedit", player, func() {
		if mfr := descr_mcpframe(descr); mfr == nil {
			do_edit(descr, player, name)
		} else {
			switch supp := mcp_frame_package_supported(mfr, "dns-org-mud-moo-simpleedit"); {
			case supp.minor == 0 && supp.major == 0:
				do_edit(descr, player, name)
			case name == "":
				notify(player, "No program name given.")
			default:
				prog := NewMatch(descr, player, name, IsProgram).
					MatchPossession().
					MatchNeighbor().
					MatchRegistered().
					MatchAbsolute().
					MatchResult()
				switch prog {
				case NOTHING:
					/* FIXME: must arrange this to query user. */
					notify(player, "I don't see that here!")
				case AMBIGUOUS:
					notify(player, "I don't know which one you mean!")
				default:
					mcpedit_program(descr, player, prog, name)
				}
			}
		}
	})
}

func do_mcpprogram(descr int, player ObjectID, name string) {
	NoGuest("@mcpprogram", player, func() {
		switch {
		case Typeof(player) != TYPE_PLAYER:
			notify(player, "Only players can edit programs.")
		case !Mucker(player):
			notify(player, "You're no programmer!")
		case name == "":
			notify(player, "No program name given.")
		default:
			prog := NewMatch(descr, player, name, IsProgram).
				MatchPossession().
				MatchNeighbor().
				MatchRegistered().
				MatchAbsolute().
				MatchResult()

			switch prog {
			case AMBIGUOUS:
				notify(player, "I don't know which one you mean!")
			case NOTHING:
				prog = new_object()
				DB.Fetch(prog).NowCalled(name)
				add_property(prog, MESGPROP_DESC, fmt.Sprintf("A scroll containing a spell called %s", name), 0)
				DB.Fetch(prog).MoveTo(player)
				DB.Fetch(prog).flags = TYPE_PROGRAM

				mlev := MLevel(player)
				switch {
				case mlev < APPRENTICE:
					mlev = JOURNEYMAN
				case mlev > MASTER:
					mlev = MASTER
				}
				SetMLevel(prog, mlev)

				DB.Fetch(prog).GiveTo(DB.Fetch(player).Owner)
				DB.Fetch(prog).(Program) = new(Program)
				DB.FetchPlayer(player).curr_prog = prog

				DB.Fetch(prog).next = DB.Fetch(player).Contents
				DB.Fetch(prog).Touch()
				DB.Fetch(player).Contents = prog
				DB.Fetch(prog).Touch()
				DB.Fetch(player).Touch()
				notify(player, fmt.Sprintf("Program %s created with number %d.", name, prog))
				fallthrough
			default:
				mcpedit_program(descr, player, prog, name)		
			}
		}
	})
}

func mcpedit_program(descr int, player, prog ObjectID, name string) {
	if mfr := descr_mcpframe(descr); mfr == nil {
		do_edit(descr, player, name)
	} else {
		switch supp := mcp_frame_package_supported(mfr, "dns-org-mud-moo-simpleedit"); {
		case supp.minor == 0 && supp.major == 0:
			do_edit(descr, player, name)
		case Typeof(player) != TYPE_PLAYER:
			show_mcp_error(mfr, "@mcpedit", "Only players can edit programs.")
		case !Mucker(player):
			show_mcp_error(mfr, "@mcpedit", "You're no programmer!")
		case name == "":
			show_mcp_error(mfr, "@mcpedit", "No program name given.");
		case Typeof(prog) != TYPE_PROGRAM, !controls(player, prog):
			show_mcp_error(mfr, "@mcpedit", "Permission denied!")
		case DB.Fetch(prog).IsFlagged(INTERNAL):
			show_mcp_error(mfr, "@mcpedit", "Sorry, this program is currently being edited by someone else.  Try again later.");
		default:
			DB.Fetch(prog).(Program).first = read_program(prog)
			DB.FetchPlayer(player).curr_prog = prog
			refstr := fmt.Sprintf("%d.prog.", prog)
			namestr := fmt.Sprintf("a program named %s(%d)", DB.Fetch(prog).name, prog)
			msg := &McpMesg{ package: "dns-org-mud-moo-simpleedit", mesgname: "content" }
			mcp_mesg_arg_append(&msg, "reference", refstr)
			mcp_mesg_arg_append(&msg, "type", "muf-code")
			mcp_mesg_arg_append(&msg, "name", namestr)
			for curr := DB.Fetch(prog).(Program).first; curr != nil; curr = curr.next {
				mcp_mesg_arg_append(&msg, "content", curr.this_line)
			}
			mcp_frame_output_mesg(mfr, &msg)
			DB.Fetch(prog).(Program).first = nil
		}
	}
}

/*
 * copy a single property, identified by its name, from one object to
 * another. helper routine for copy_props (below).
 */
func copy_one_prop(source ObjectID, propname string) (r interface{}) {
	if currprop := get_property(source, propname); currprop != nil {
		switch v := currprop.(type) {
		case string:
			newprop.data = v
		case int:
			newprop.data = v
		case float64:
			newprop.data = v
		case ObjectID:
			newprop.data = v
		case Lock:
			newprop.data = copy_bool(v)
		}
	}
	return
}

/*
 * copy a property (sub)tree from one object to another one. this is a
 * helper routine used by do_clone, based loosely on listprops_wildcard from
 * look.c.
 */
func copy_props(player, source, destination ObjectID, dir string) {
	/* loop through all properties in the current propdir */
	var pptr *Plist
	var propname string
	for propadr := pptr.first_prop(source, dir, propname); propadr != nil; propadr, propname = propadr.next_prop(pptr) {
		prop := dir + PROPDIR_DELIMITER + propname
		if tp_verbose_clone && Wizard(DB.Fetch(player).Owner) {
			notify(player, fmt.Sprintf("copying property %s", prop))
		}
		set_property(destination, buf, copy_one_prop(source, prop))
		copy_props(player, source, destination, prop)
	}
}

/*
 * do_clone
 *
 * Use this to clone an object.
 */
func do_clone(descr int, player ObjectID, name string) {
	NoGuest("@clone", player, func() {
		switch {
		case !Builder(player):
			notify(player, "That command is restricted to authorized builders.")
		case name == "":
			notify(player, "Clone what?")
		default:
			/* All OK so far, so try to find the thing that should be cloned. We do not allow rooms, exits, etc. to be cloned for now. */
			thing := NewMatch(descr, player, name, IsThing).
				MatchPossession().
				MatchNeighbor().
				MatchRegistered().
				MatchAbsolute().
				NoisyMatchResult()
	
			switch {
			case thing == NOTHING:
			case thing == AMBIGUOUS:
 				notify(player, "I don't know which one you mean!")
			case !IsThing(thing):
				notify(player, "That is not a cloneable object.")
			case !ok_name(DB.Fetch(thing).name):
				/* check the name again, just in case reserved name patterns have changed since the original object was created. */
				notify(player, "You cannot clone something with such a weird name!")
			case !controls(player, thing):
				notify(player, "Permission denied. (you can't clone this)")
			default:
				cost := OBJECT_GETCOST(get_property_value(thing, MESGPROP_VALUE))
				if cost < tp_object_cost {
					cost = tp_object_cost
				}
				if !payfor(player, cost) {
					notify_fmt(player, "Sorry, you don't have enough %s.", tp_pennies)
				} else {
					if tp_verbose_clone {
						notify(player, fmt.Sprintf("Now cloning %s...", unparse_object(player, thing)))
					}
		
					clonedthing := new_object()
					DB.Fetch(clonedthing).NowCalled(DB.Fetch(thing).name)
					DB.Store(clonedthing, Player)
					DB.Fetch(clonedthing).MoveTo(player)
					DB.Fetch(clonedthing).GiveTo(DB.Fetch(player).Owner)
					add_property(clonedthing, MESGPROP_VALUE, nil, get_property_value(thing, MESGPROP_VALUE))

					/* FIXME: should we clone attached actions? */
					DB.Fetch(clonedthing).Exits = NOTHING
					DB.Fetch(clonedthing).Bitset = DB.Fetch(thing).Bitset

					copy_props(player, thing, clonedthing, "")
					if get_property_value(thing, MESGPROP_VALUE) > tp_max_object_endowment {
						add_property(thing, MESGPROP_VALUE, nil, tp_max_object_endowment)
					}
					DB.FetchPlayer(clonedthing).LiveAt(DB.FetchPlayer(thing).Home)
					DB.Fetch(clonedthing).next = DB.Fetch(player).Contents
					DB.Fetch(clonedthing).Touch()

					DB.Fetch(player).Contents = clonedthing
					DB.Fetch(player).Touch()

					notify(player, fmt.Sprintf("%s created with number %d.", DB.Fetch(thing).name, clonedthing))
					DB.Fetch(clonedthing).Touch()
				}
			}
		} 
	})
}

/*
 * do_create
 *
 * Use this to create an object.
 */
func do_create(player ObjectID, name, acost string) {
	NoGuest("@create", player, func() {
		var qname, rname string
		if terms := strings.SplitN(acost, "=", 2); len(terms) == 2 {
			qname = strings.TrimFunc(terms[0], unicode.IsSpace)
			rname = strings.TrimFunc(terms[1], unicode.IsSpace)
		}
		switch {
		case !Builder(player):
			notify(player, "That command is restricted to authorized builders.")
		case name == "":
			notify(player, "Create what?")
		case !ok_ascii_thing(name):
			notify(player, "Thing names are limited to 7-bit ASCII.")
		case !ok_name(name):
			notify(player, "That's a silly name for a thing!")
		case cost < 0:
			notify(player, "You can't create an object for less than nothing!")
		default:
			cost := strconv.Atoi(qname)
			if cost < tp_object_cost {
				cost = tp_object_cost
			}
			if !payfor(player, cost) {
				notify_fmt(player, "Sorry, you don't have enough %s.", tp_pennies)
			} else {
				p := DB.Fetch(player)
				thing := new_object()
				DB.Store(thing, &Object{
					name: name,
					location: player,
					owner: p.Owner,
					exits: NOTHING,
				})
				add_property(thing, MESGPROP_VALUE, nil, OBJECT_ENDOWMENT(cost))
				if get_property_value(thing, MESGPROP_VALUE) > tp_max_object_endowment {
					add_property(thing, MESGPROP_VALUE, nil, tp_max_object_endowment)
				}
				if loc := p.Location(); loc != NOTHING && controls(player, loc) {
					DB.Fetch(thing).(HasHome).LiveAt(loc)
				} else {
					DB.Fetch(thing).(HasHome).LiveAt(player)
				}
				DB.Fetch(thing).next = p.Contents
				DB.Fetch(thing).Touch()
				p.Contents = thing
				DB.Fetch(player).Touch()
				notify(player, fmt.Sprintf("%s created with number %d.", name, thing))
				DB.Fetch(thing).Touch()
				if rname != "" {
					notify(player, fmt.Sprintf("Registered as $%s", rname))
					set_property(player, fmt.Sprintf("_reg/%s", rname), thing)
				}		
			}
		}
	})
}

/*
 * parse_source()
 *
 * This is a utility used by do_action and do_attach.  It parses
 * the source string into a ObjectID, and checks to see that it
 * exists.
 *
 * The return value is the ObjectID of the source, or NOTHING if an
 * error occurs.
 *
 */
func parse_source(int descr, ObjectID player, const char *source_name) (r ObjectID) {
	r = NewMatch(descr, player, source_name, NOTYPE).
		MatchNeighbor().
		MatchMe().
		MatchHere().
		MatchPossession().
		MatchRegistered().
		MatchAbsolute().
		NoisyMatchResult()

	switch {
	case r == NOTHING:
	case !controls(player, r):
		notify(player, "Permission denied. (you don't control the attachment point)")
	case Typeof(r) == TYPE_EXIT:
		notify(player, "You can't attach an action to an action.")
	case Typeof(source) == TYPE_PROGRAM:
		notify(player, "You can't attach an action to a program.")
	}
	return
}

/*
 * set_source()
 *
 * This routine sets the source of an action to the specified source.
 * It is called by do_action and do_attach.
 *
 */
func set_source(player, action, source ObjectID) {
	switch s := DB.Fetch(source); s.(type) {
	case Room, Object, Player:
		a := DB.Fetch(action)
		a.next = s.Exits
		s.Exits = action
		s.Touch()
		a.MoveTo(source)
		a.Touch()
	default:
		notify(player, "Internal error: weird object type.")
		log_status("PANIC: tried to source %d to %d: type: %T", action, source, s)
	}
}

func unset_source(player, loc, action ObjectID) bool {
	if oldsrc := DB.Fetch(action).Location(); oldsrc == NOTHING {
		/* old-style, sourceless exit */
		if !member(action, DB.Fetch(loc).Exits) {
			return false
		}
		l := DB.Fetch(player).Location()
		DB.Fetch(l).Exits = remove_first(DB.Fetch(l).Exits, action)
		DB.Fetch(l).Touch()
	} else {
		switch o := DB.Fetch(oldsrc); o.(type) {
		case Player, Room, Object:
			o.Exits = remove_first(o.Exits, action)
			o.Touch()
		default:
			log_status("PANIC: source of action #%d was type: %T.", action, o)
			return false
		}
	}
	return true
}

/*
 * do_action()
 *
 * This routine attaches a new existing action to a source object,
 * where possible.
 * The action will not do anything until it is LINKed.
 *
 */
func do_action(descr int, player ObjectID, action_name, source_name string) {
	NoGuest("@action", player, func() {
		if !Builder(player) {
			notify(player, "That command is restricted to authorized builders.")
		} else {
			var qname, rname string
			switch terms := strings.SplitN(source_name, "=", 2); len(terms) {
			case 2:
				qname = strings.TrimFunc(terms[0], unicode.IsSpace)
				rname = strings.TrimFunc(terms[1], unicode.IsSpace)
			}

			switch {
			case action_name == "", qname == "":
				notify(player, "You must specify an action name and a source object.")
			case !ok_ascii_other(action_name):
				notify(player, "Action names are limited to 7-bit ASCII.")
			case !ok_name(action_name):
				notify(player, "That's a strange name for an action!")
			default:
				if source := parse_source(descr, player, qname); source != NOTHING {
					if !payfor(player, tp_exit_cost) {
						notify_fmt(player, "Sorry, you don't have enough %s to make an action.", tp_pennies)
					} else {
						action := new_object()
						DB.Fetch(action).NowCalled(action_name)
						DB.Fetch(action).MoveTo(NOTHING)
						DB.Fetch(action).GiveTo(DB.Fetch(player).Owner)
						DB.Fetch(action).(Exit).Destinations = nil
						DB.Fetch(action).flags = TYPE_EXIT

						set_source(player, action, source)
						notify(player, fmt.Sprintf("Action created with number %d and attached.", action))
						DB.Fetch(action).Touch()

						if rname != "" {
							notify(player, fmt.Sprintf("Registered as $%s", rname))
							set_property(player, fmt.Sprintf("_reg/%s", rname), action)
						}
					}
				}
			}
		}
	})
}

/*
 * do_attach()
 *
 * This routine attaches a previously existing action to a source object.
 * The action will not do anything unless it is LINKed.
 *
 */
func do_attach(descr int, player ObjectID, action_name, source_name string) {
	NoGuest("@attach", player, func() {
		if loc := DB.Fetch(player).Location); loc != NOTHING {
			switch {
			case !Builder(player):
				notify(player, "That command is restricted to authorized builders.")
			case action_name == "", source_name == "":
				notify(player, "You must specify an action name and a source object.")
			default:
				md := NewMatch(descr, player, action_name, IsExit).
					MatchAllExits().
					MatchRegistered().
					MatchAbsolute()

				if action := md.NoisyMatchResult(); action != NOTHING {
					switch source := parse_source(descr, player, source_name); {
					case Typeof(action) != TYPE_EXIT:
						notify(player, "That's not an action!")
					case !controls(player, action):
						notify(player, "Permission denied. (you don't control the action you're trying to reattach)")
					case source == NOTHING, Typeof(source) == TYPE_PROGRAM, !unset_source(player, loc, action):
					default:
						set_source(player, action, source)
						notify(player, "Action re-attached.")
						if MLevRaw(action) != NON_MUCKER {
							SetMLevel(action, NON_MUCKER)
							notify(player, "Action priority Level reset to zero.")
						}
					}
				}
			}
		}
	})
}