package fbmuck

func can_link_to(who ObjectID, what_type int, where ObjectID) (r bool) {
	switch {
	case where == HOME:
		r = true
	case !where.IsValid():
		r = false
	default:
		switch w := DB.Fetch(where); what_type {
		case Exit:
			//	If the target is LINK_OK, then any exit may be linked there.  Otherwise, only someone who controls the target may link there.
			r = controls(who, where) || w.IsFlagged(LINK_OK)
		case Player:
			/* Players may only be linked to rooms, that are either
			 * controlled by the player or set either L or A. */
			switch w.(type) {
			case Room:
				r = controls(who, where) || Linkable(where)
			}
		case Room:
			//	Rooms may be linked to rooms or things (this sets their dropto location).  Target must be controlled, or be L or A.
			switch w.(type) {
			case Room, Object:
				r = controls(who, where) || Linkable(where)
			}
		case Object:
			//	Things may be linked to rooms, players, or other things (this sets the thing's home).  Target must be controlled, or be L or A.
			switch DB.Fetch(where).(type) {
			case Room, Player, Object:
				r = controls(who, where) || Linkable(where)
			}
		case NOTYPE:
			/* Why is this here? -winged */
			switch w.(type) {
			case Object:
				r = controls(who, where) || w.IsFlagged(LINK_OK)
			default:
				r = controls(who, where) || w.IsFlaggedAnyOf(LINK_OK, ABODE)
			}
		}
	}
	return
}

/* This checks to see if what can be linked to something else by who. */
func can_link(ObjectID who, ObjectID what) bool {
	return controls(who, what) || (TYPEOF(what) == TYPE_EXIT && len(DB.Fetch(what).(Exit).Destinations) == 0)
}

/*
 * Revision 1.2 -- SECURE_TELEPORT
 * you can only jump with an action from rooms that you own
 * or that are jump_ok, and you cannot jump to players that are !jump_ok.
 */

/*
 * could_doit: Checks to see if player could actually do what is proposing
 * to be done: if thing is an exit, this checks to see if the exit will
 * perform a move that is allowed. Then, it checks the @lock on the thing,
 * whether it's an exit or not.
 */
func could_doit(int descr, ObjectID player, ObjectID thing) bool {
	ObjectID source, dest, owner;

	if TYPEOF(thing) == TYPE_EXIT {
			/* If exit is unlinked, can't do it. */
		if len(DB.Fetch(thing).(Exit).Destinations) == 0 {
			return false
		}

		owner = DB.Fetch(thing).Owner
		source = DB.Fetch(player).Location
		dest = *(DB.Fetch(thing).(Exit).Destinations)

		if IsPlayer(dest) {
			/* Check for additional restrictions related to player dests */
			ObjectID destplayer = dest;

			dest = DB.Fetch(dest).Location
			/* If the dest player isn't JUMP_OK, or if the dest player's loc
			 * is set BOUND, can't do it. */
			if !DB.Fetch(destplayer).IsFlagged(JUMP_OK) || DB.Fetch(dest).IsFlagged(BUILDER) {
				return false
			}
		}

		/* for actions */
		if DB.Fetch(thing).Location != NOTHING) && TYPEOF(DB.Fetch(thing).Location) != TYPE_ROOM {
			/* If this is an exit on a Thing or a Player... */

			/* If the destination is a room or player, and the current
			 * location is set BOUND (note: if the player is in a vehicle
			 * set BUILDER this will also return failure) */
			if (Typeof(dest) == TYPE_ROOM || Typeof(dest) == TYPE_PLAYER) && DB.Fetch(source).IsFlagged(BUILDER) {
				return false
			}

			/* If secure_teleport is true, and if the destination is a room */
			if tp_secure_teleport && Typeof(dest) == TYPE_ROOM {
				/* if player doesn't control the source and the source isn't
				 * set Jump_OK, then if the destination isn't HOME,
				 * can't do it.  (Should this include getlink(owner)?  Not
				 * everyone knows that 'home' or '#-3' can be linked to and
				 * be treated specially. -winged) */
				if dest != HOME && !controls(owner, source) && !DB.Fetch(source).IsFlagged(JUMP_OK) {
					return false
				}
			/* FIXME: Add support for in-server banishment from rooms and environments here. */
			}
		}
	}
	/* Check the @lock on the thing, as a final test. */
	return copy_bool(get_property_lock(thing, MESGPROP_LOCK)).Eval(descr, player, thing)
}

func test_lock(descr int, player, thing ObjectID, lockprop string) int {
	return copy_bool(get_property_lock(thing, lockprop)).Eval(descr, player, thing)
}

func test_lock_false_default(descr int, player, thing ObjectID, lockprop string) (ok bool) {
	if lok := get_property_lock(thing, lockprop); !lok.IsTrue() {
		ok = copy_bool(lok).Eval(descr, player, thing))
	}
	return
}

func can_doit(descr int, player, thing ObjectID, default_fail_msg string) (r bool) {
	switch p := DB.Fetch(player); {
	case p.Location() == NOTHING:
	case !Wizard(p.Owner) && IsThing(player) && DB.Fetch(thing).IsFlagged(ZOMBIE):
		notify(player, "Sorry, but zombies can't do that.")
	case !could_doit(descr, player, thing):
		if get_property_class(thing, MESGPROP_FAIL) {
			exec_or_notify_prop(descr, player, thing, MESGPROP_FAIL, "(@Fail)")
		} else if (default_fail_msg) {
			notify(player, default_fail_msg)
		}
		if get_property_class(thing, MESGPROP_OFAIL) && !Dark(player) {
			parse_oprop(descr, player, loc, thing, MESGPROP_OFAIL, p.Name(), "(@Ofail)")
		}
	default:
		if get_property_class(thing, MESGPROP_SUCC) {
			exec_or_notify_prop(descr, player, thing, MESGPROP_SUCC, "(@Succ)")
		}
		if get_property_class(thing, MESGPROP_OSUCC) && !Dark(player) {
			parse_oprop(descr, player, loc, thing, MESGPROP_OSUCC, p.Name(), "(@Osucc)")
		}
		r = true
	}
	return
}

func can_see(ObjectID player, ObjectID thing, int can_see_loc) (r bool) {
	switch {
	case player == thing || Typeof(thing) == TYPE_EXIT || Typeof(thing) == TYPE_ROOM:
	case can_see_loc:
		switch Typeof(thing) {
		case TYPE_PROGRAM:
			r = DB.Fetch(thing).IsFlagged(LINK_OK) || controls(player, thing)
		case TYPE_PLAYER:
			r = tp_dark_sleepers && !Dark(thing) && online(thing)
		default:
			r = !Dark(thing) || (controls(player, thing) && !DB.Fetch(player).IsFlagged(STICKY))
		}
	default:
		r = controls(player, thing) && !DB.Fetch(player).IsFlagged(STICKY)
	}
	return
}

func controls(who, what ObjectID) bool {
	/* No one controls invalid objects */
	if what.IsValid() {
		/* Zombies and puppets use the permissions of their owner */
		if Typeof(who) != TYPE_PLAYER {
			who = DB.Fetch(who).Owner
		}
		/* Wizard controls everything */
		if Wizard(who) {
			if DB.Fetch(what).Owner == GOD && who != GOD {
				/* Only God controls God's objects */
				return false
			} else {
				return true
			}
		}

		if tp_realms_control {
			/* Realm Owner controls everything under his environment. */
			/* To set up a Realm, a Wizard sets the W flag on a room.  The
			 * owner of that room controls every Room object contained within
			 * that room, all the way to the leaves of the tree.
			 * -winged */
			for index := what; index != NOTHING; index = DB.Fetch(index).Location {
				if DB.Fetch(index).Owner == who && Typeof(index) == TYPE_ROOM && Wizard(index) {
					/* Realm Owner doesn't control other Player objects */
					if Typeof(what) == TYPE_PLAYER {
						return false
					} else {
						return true
					}
				}
			}
		}

		/* exits are also controlled by the owners of the source and destination */
		/* ACTUALLY, THEY AREN'T.  IT OPENS A BAD MPI SECURITY HOLE. */
		/* any MPI on an exit's @succ or @fail would be run in the context
		 * of the owner, which would allow the owner of the src or dest to
		 * write malicious code for the owner of the exit to run.  Allowing them
		 * control would allow them to modify _ properties, thus enabling the
		 * security hole. -winged */
		/*
		 * if TYPEOF(what) == TYPE_EXIT {
		 *    dest := DB.Fetch(what).(Exit).Destinations
		 *    for i := len(dest) - 1; i > -1; i-- {
		 *        if who == DB.Fetch(dest[i]).Owner {
		 *            return true
		 *        }
		 *    }
		 *    if who == DB.Fetch(DB.Fetch(what).Location).Owner {
		 *        return true
		 *    }
		 * }
		 */

		/* owners control their own stuff */
		return who == DB.Fetch(what).Owner
	}
	return false
}

func restricted(player, thing ObjectID, flag int) (r bool) {
	switch p := DB.Fetch(player); flag {
	case ABODE:
		if _, ok := DB.Fetch(thing).(Program); ok {
			//	Trying to set a program AUTOSTART requires TrueWizard
			r = !TrueWizard(p.Owner)
		}
	case YIELD:
		r = !Wizard(p.Owner)
	case OVERT:
		r = !Wizard(p.Owner)
	case ZOMBIE:
		switch DB.Fetch(thing).(type) {
		case Player:
			//	Restricting a player from using zombies requires a wizard
			r = !Wizard(p.Owner)
		case Object:
			//	If a player's set Zombie, he's restricted from using them unless he's a wizard, in which case he can do whatever
			r = DB.Fetch(p.Owner).IsFlagged(ZOMBIE) && !Wizard(p.Owner)
		}
	case VEHICLE:
		switch DB.Fetch(thing).(type) {
		case Player:
			/* Restricting a player from using vehicles requires a wizard. */
			r = !Wizard(p.Owner)
		case Object:
			switch {
			case tp_wiz_vehicles:
				r = !Wizard(p.Owner)
			case p.IsFlagged(VEHICLE):
				r = !Wizard(p.Owner)
			}
		}
	case DARK:
		if !Wizard(p.Owner) {
			/* Setting a player dark requires a wizard. */
			switch DB.Fetch(thing).(type) {
			case Player:
				r = true
			case Exit:
				r = !tp_exit_darking
			case Object:
				r = !tp_thing_darking
			}
		}
	case QUELL:
		//	Only God (or God's stuff) can quell or unquell another wizard.
		switch t := DB.Fetch(thing); t.(type) {
		case Player:
			r = TrueWizard(thing) && thing != player
		default:
			r = p.Owner == t.Owner
		}
	case MUCKER, SMUCKER, SMUCKER | MUCKER, BUILDER:
		/* Would someone tell me why setting a program SMUCKER|MUCKER doesn't
		 * go through here? -winged */
		/* Setting a program Bound causes any function called therein to be
		 * put in preempt mode, regardless of the mode it had before.
		 * Since this is just a convenience for atomic-functionwriters,
		 * why is it limited to only a Wizard? -winged */
		/* Setting a player Builder is limited to a Wizard. */
		r = !Wizard(p.Owner)
	case WIZARD:
		if r = Wizard(p.Owner); !r {
			/* To do anything with a Wizard flag requires a Wizard. */
			switch DB.Fetch(thing).(type) {
			case Player:
				/* ...but only God can make a player a Wizard, or re-mort one. */
				r = player != GOD
			}
		}
	}
}

/* Removes 'cost' value from 'who', and returns true if the act has been
 * paid for, else returns false. */
func payfor(who ObjectID, cost int) (r bool) {
	who = DB.Fetch(who).Owner
	/* Wizards don't have to pay for anything. */
	if Wizard(who) {
		r = true
	} else if get_property_value(who, MESGPROP_VALUE) >= cost {
		add_property(who, MESGPROP_VALUE, nil, get_property_value(who, MESGPROP_VALUE) - cost)
		DB.Fetch(who).Touch()
		r = true
	}
	return
}

static inline int
ok_ascii_any(const char *name)
{
	const unsigned char *scan;
	for( scan=(const unsigned char*) name; *scan; ++scan ) {
		if( *scan>127 )
			return 0;
	}
	return 1;
}

int 
ok_ascii_thing(const char *name)
{
	return !tp_7bit_thing_names || ok_ascii_any(name);
}

int
ok_ascii_other(const char *name)
{
	return !tp_7bit_other_names || ok_ascii_any(name);
}
	
func ok_name(name string) (r bool) {
	switch {
	case name == "":
	case name[0] == LOOKUP_TOKEN, name[0] == REGISTERED_TOKEN, name[0] == NUMBER_TOKEN, name[0] == NOT_TOKEN):
	case strings.ContainsAny(name, ARG_DELIMITER + AND_TOKEN + OR_TOKEN + '\r' + ESCAPE_CHAR):
	case name == "me", name == "home", name == "here":
	case tp_reserved_names != nil && tp_reserved_names == name && (!*tp_reserved_names || smatch(tp_reserved_names, name) != 0)):
	default:
		r = true
	}
	return
}

func ok_player_name(name string) (r bool) {
	if r = true; ok_name(name) {
		for scan := name; r && scan != ""; scan = scan[1:] {
			r = !(unicode.IsPrint(scan[0]) && !unicode.IsSpace(scan[0])) && scan[0] != '(' && scan[0] != ')' && scan[0] != '\'' && scan[0] != ',' {	
		}
		if r && (tp_reserved_player_names == "" || smatch(tp_reserved_player_names, name) != 0) {
			r = lookup_player(name) == NOTHING
		}
	}
	return
}

func ok_password(password string) (r bool) {
	/* Password cannot be blank */
	if *password != "" {
		/* Password also cannot contain any nonprintable or space-type characters */
		const char *scan;
		for scan = password; *scan; scan++ {
   			if !(unicode.IsPrint(*scan) && !unicode.IsSpace(*scan)) {
   				return 0
   			}
   		}
		/* Anything else is fair game */
   		return 1
	}
	return 0
}

/* If only paternity checks were this easy in real life... 
 * Returns 1 if the given 'child' is contained by the 'parent'.*/
func isancestor(parent, child ObjectID) bool {
	for child != NOTHING && child != parent {
		child = getparent(child)
	}
	return child == parent
}