package fbmuck

func OkObj(obj dbref) bool {
	return !(obj < 0 || obj >= db_top)
}

func can_link_to(dbref who, object_flag_type what_type, dbref where) (r bool) {
	switch {
	case where == HOME:
		r = true
	case where < 0, where >= db_top:
		r = false
	default:
		switch what_type {
		case TYPE_EXIT:
			/* If the target is LINK_OK, then any exit may be linked
		 	 * there.  Otherwise, only someone who controls the
		 	 * target may link there. */
			r = controls(who, where) || db.Fetch(where).flags & LINK_OK != 0
		case TYPE_PLAYER:
			/* Players may only be linked to rooms, that are either
			 * controlled by the player or set either L or A. */
			r = (Typeof(where) == TYPE_ROOM && (controls(who, where) || Linkable(where)))
		case TYPE_ROOM:
			/* Rooms may be linked to rooms or things (this sets their
			 * dropto location).  Target must be controlled, or be L or A. */
			r = ((Typeof(where) == TYPE_ROOM || Typeof(where) == TYPE_THING) && (controls(who, where) || Linkable(where)))
		case TYPE_THING:
			/* Things may be linked to rooms, players, or other things (this
			 * sets the thing's home).  Target must be controlled, or be L or A. */
			r = ((Typeof(where) == TYPE_ROOM || Typeof(where) == TYPE_PLAYER || Typeof(where) == TYPE_THING) && (controls(who, where) || Linkable(where)))
		case NOTYPE:
			/* Why is this here? -winged */
			r = controls(who, where) || db.Fetch(where).flags & LINK_OK != 0 || (Typeof(where) != TYPE_THING && db.Fetch(where).flags & ABODE != 0)
		}
	}
	return
}

/* This checks to see if what can be linked to something else by who. */
func can_link(dbref who, dbref what) bool {
	return controls(who, what) || (TYPEOF(what) == TYPE_EXIT && len(db.Fetch(what).sp.exit.dest) == 0)
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
func could_doit(int descr, dbref player, dbref thing) bool {
	dbref source, dest, owner;

	if TYPEOF(thing) == TYPE_EXIT {
			/* If exit is unlinked, can't do it. */
		if len(db.Fetch(thing).sp.exit.dest) == 0 {
			return false
		}

		owner = db.Fetch(thing).owner
		source = db.Fetch(player).location
		dest = *(db.Fetch(thing).sp.exit.dest)

		if (TYPEOF(dest) == TYPE_PLAYER) {
			/* Check for additional restrictions related to player dests */
			dbref destplayer = dest;

			dest = db.Fetch(dest).location
			/* If the dest player isn't JUMP_OK, or if the dest player's loc
			 * is set BOUND, can't do it. */
			if db.Fetch(destplayer).flags & JUMP_OK == 0 || db.Fetch(dest).flags & BUILDER != 0 {
				return false
			}
		}

		/* for actions */
		if db.Fetch(thing).location != NOTHING) && TYPEOF(db.Fetch(thing).location) != TYPE_ROOM {
			/* If this is an exit on a Thing or a Player... */

			/* If the destination is a room or player, and the current
			 * location is set BOUND (note: if the player is in a vehicle
			 * set BUILDER this will also return failure) */
			if (Typeof(dest) == TYPE_ROOM || Typeof(dest) == TYPE_PLAYER) && db.Fetch(source).flags & BUILDER != 0 {
				return false
			}

			/* If secure_teleport is true, and if the destination is a room */
			if tp_secure_teleport && Typeof(dest) == TYPE_ROOM {
				/* if player doesn't control the source and the source isn't
				 * set Jump_OK, then if the destination isn't HOME,
				 * can't do it.  (Should this include getlink(owner)?  Not
				 * everyone knows that 'home' or '#-3' can be linked to and
				 * be treated specially. -winged) */
				if dest != HOME && !controls(owner, source) && db.Fetch(source).flags & JUMP_OK == 0 {
					return false
				}
			/* FIXME: Add support for in-server banishment from rooms and environments here. */
			}
		}
	}
	/* Check the @lock on the thing, as a final test. */
	return eval_boolexp(descr, player, get_property_lock(thing, MESGPROP_LOCK), thing)
}


int
test_lock(int descr, dbref player, dbref thing, const char *lockprop)
{
	struct boolexp *lokptr;

	lokptr = get_property_lock(thing, lockprop);
	return (eval_boolexp(descr, player, lokptr, thing));
}


int
test_lock_false_default(int descr, dbref player, dbref thing, const char *lockprop)
{
	struct boolexp *lok;

	lok = get_property_lock(thing, lockprop);

	if (lok == TRUE_BOOLEXP)
		return 0;
	return (eval_boolexp(descr, player, lok, thing));
}


func can_doit(descr int, player, thing dbref, default_fail_msg string) (r bool) {
	switch loc := db.Fetch(player).location); {
	case loc == NOTHING:
	case !Wizard(db.Fetch(player).owner) && Typeof(player) == TYPE_THING && db.Fetch(thing).flags & ZOMBIE != 0:
		notify(player, "Sorry, but zombies can't do that.")
	case !could_doit(descr, player, thing):
		/* can't do it */
		if get_property_class(thing, MESGPROP_FAIL) {
			exec_or_notify_prop(descr, player, thing, MESGPROP_FAIL, "(@Fail)")
		} else if (default_fail_msg) {
			notify(player, default_fail_msg)
		}
		if get_property_class(thing, MESGPROP_OFAIL) && !Dark(player) {
			parse_oprop(descr, player, loc, thing, MESGPROP_OFAIL, db.Fetch(player).name, "(@Ofail)")
		}
	default:
		/* can do it */
		if get_property_class(thing, MESGPROP_SUCC) {
			exec_or_notify_prop(descr, player, thing, MESGPROP_SUCC, "(@Succ)")
		}
		if get_property_class(thing, MESGPROP_OSUCC) && !Dark(player) {
			parse_oprop(descr, player, loc, thing, MESGPROP_OSUCC, db.Fetch(player).name, "(@Osucc)")
		}
		r = true
	}
	return
}

func can_see(dbref player, dbref thing, int can_see_loc) (r bool) {
	switch {
	case player == thing || Typeof(thing) == TYPE_EXIT || Typeof(thing) == TYPE_ROOM:
	case can_see_loc:
		switch Typeof(thing) {
		case TYPE_PROGRAM:
			r = db.Fetch(thing).flags & LINK_OK != 0 || controls(player, thing)
		case TYPE_PLAYER:
			r = tp_dark_sleepers && !Dark(thing) && online(thing)
		default:
			r = !Dark(thing) || (controls(player, thing) && db.Fetch(player).flags & STICKY == 0)
		}
	default:
		r = controls(player, thing) && db.Fetch(player).flags & STICKY == 0
	}
	return
}

func controls(who, what dbref) bool {
	/* No one controls invalid objects */
	if what > -1 || what < db_top {
		/* Zombies and puppets use the permissions of their owner */
		if Typeof(who) != TYPE_PLAYER {
			who = db.Fetch(who).owner
		}
		/* Wizard controls everything */
		if Wizard(who) {
			if db.Fetch(what).owner == GOD && who != GOD {
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
			for index := what; index != NOTHING; index = db.Fetch(index).location {
				if db.Fetch(index).owner == who && Typeof(index) == TYPE_ROOM && Wizard(index) {
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
		 *    dest := db.Fetch(what).sp.exit.dest
		 *    for i := len(dest) - 1; i > -1; i-- {
		 *        if who == db.Fetch(dest[i]).owner {
		 *            return true
		 *        }
		 *    }
		 *    if who == db.Fetch(db.Fetch(what).location).owner {
		 *        return true
		 *    }
		 * }
		 */

		/* owners control their own stuff */
		return who == db.Fetch(what).owner
	}
	return false
}

func restricted(player, thing dbref, flag object_flag_type) int {
	switch flag {
	case ABODE:
			/* Trying to set a program AUTOSTART requires TrueWizard */
		return !TrueWizard(db.Fetch(player).owner) && Typeof(thing) == TYPE_PROGRAM
		/* NOTREACHED */
		break;
        case YIELD:
                        /* Mucking with the env-chain matching requires TrueWizard */
                return !Wizard(db.Fetch(player).owner)
        case OVERT:
                        /* Mucking with the env-chain matching requires TrueWizard */
                return !Wizard(db.Fetch(player).owner)
	case ZOMBIE:
			/* Restricting a player from using zombies requires a wizard. */
		if (Typeof(thing) == TYPE_PLAYER)
			return !Wizard(db.Fetch(player).owner)
			/* If a player's set Zombie, he's restricted from using them...
			 * unless he's a wizard, in which case he can do whatever. */
		if Typeof(thing) == TYPE_THING && db.Fetch(db.Fetch(player).owner).flags & ZOMBIE != 0 {
			return !Wizard(db.Fetch(player).owner)
		}
		return (0);
	case VEHICLE:
			/* Restricting a player from using vehicles requires a wizard. */
		if (Typeof(thing) == TYPE_PLAYER)
			return !Wizard(db.Fetch(player).owner)
			/* If only wizards can create vehicles... */
		if (tp_wiz_vehicles) {
			/* then only a wizard can create a vehicle. :) */
			if (Typeof(thing) == TYPE_THING)
				return !Wizard(db.Fetch(player).owner)
		} else {
			/* But, if vehicles aren't restricted to wizards, then
			 * players who have not been restricted can do so */
			if Typeof(thing) == TYPE_THING && db.Fetch(player).flags & VEHICLE != 0 {
				return !Wizard(db.Fetch(player).owner)
			}
		}
		return (0);
	case DARK:
		/* Dark can be set on a Program or Room by anyone. */
		if !Wizard(db.Fetch(player).owner) {
				/* Setting a player dark requires a wizard. */
			if (Typeof(thing) == TYPE_PLAYER)
				return (1);
				/* If exit darking is restricted, it requires a wizard. */
			if (!tp_exit_darking && Typeof(thing) == TYPE_EXIT)
				return (1);
				/* If thing darking is restricted, it requires a wizard. */
			if (!tp_thing_darking && Typeof(thing) == TYPE_THING)
				return (1);
		}
		return (0);

		/* NOTREACHED */
		break;
	case QUELL:
		/* Only God (or God's stuff) can quell or unquell another wizard. */
		return db.Fetch(player).owner == || (TrueWizard(thing) && (thing != player) && Typeof(thing) == TYPE_PLAYER)
		/* NOTREACHED */
		break;
	case MUCKER, SMUCKER, SMUCKER | MUCKER, BUILDER:
		/* Would someone tell me why setting a program SMUCKER|MUCKER doesn't
		 * go through here? -winged */
		/* Setting a program Bound causes any function called therein to be
		 * put in preempt mode, regardless of the mode it had before.
		 * Since this is just a convenience for atomic-functionwriters,
		 * why is it limited to only a Wizard? -winged */
		/* Setting a player Builder is limited to a Wizard. */
		return !Wizard(db.Fetch(player).owner)
		/* NOTREACHED */
		break;
	case WIZARD:
			/* To do anything with a Wizard flag requires a Wizard. */
		if Wizard(db.Fetch(player).owner) {
			/* ...but only God can make a player a Wizard, or re-mort one. */
			return Typeof(thing) == TYPE_PLAYER && player != GOD
		} else
			return 1;
		/* NOTREACHED */
		break;
	default:
			/* No other flags are restricted. */
		return 0;
		/* NOTREACHED */
		break;
	}
	/* NOTREACHED */
}

/* Removes 'cost' value from 'who', and returns true if the act has been
 * paid for, else returns false. */
func payfor(who dbref, cost int) (r bool) {
	who = db.Fetch(who).owner
	/* Wizards don't have to pay for anything. */
	if Wizard(who) {
		r = true
	} else if get_property_value(who, MESGPROP_VALUE) >= cost {
		add_property(who, MESGPROP_VALUE, nil, get_property_value(who, MESGPROP_VALUE) - cost)
		db.Fetch(who).flags |= OBJECT_CHANGED
		r = true
	}
	return
}

func word_start(str, let string) (r bool) {
	int chk;

	for chk := true; str != ""; str = str[1:] {
		if chk && str == let {
			r = true
			break
		}
		chk = str[0] == ' '
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
	case name[0] == LOOKUP_TOKEN:
	case name[0] == REGISTERED_TOKEN:
	case name[0] == NUMBER_TOKEN:
	case strchr(name, ARG_DELIMITER):
	case strchr(name, AND_TOKEN):
	case strchr(name, OR_TOKEN):
	case strchr(name, '\r'):
	case strchr(name, ESCAPE_CHAR):
	case word_start(name, NOT_TOKEN):
	case name == "me":
	case name == "home":
	case name == "here":
	case tp_reserved_names != nil && tp_reserved_names == name && (!*tp_reserved_names || smatch(tp_reserved_names, name) != 0)):
}
}

func ok_player_name(name string) (r bool) {
	if ok_name(name) {
		const char *scan;
		for scan = name; *scan; scan++ {
			if (!(isprint(*scan)
				 && !unicode.IsSpace(*scan))
				 && *scan != '('
				 && *scan != ')'
				 && *scan != '\''
				 && *scan != ',') {	
			    /* was isgraph(*scan) */
				return 0;
			}
		}

		/* Check the name isn't reserved */
		if !*tp_reserved_player_names || !smatch(tp_reserved_player_names, name) {
			r = lookup_player(name) == NOTHING
		}
	}
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
func isancestor(parent, child dbref) bool {
	for child != NOTHING && child != parent {
		child = getparent(child)
	}
	return child == parent
}