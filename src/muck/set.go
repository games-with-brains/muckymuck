package fbmuck

func MatchControlled(descr int, player dbref, name string) (match dbref) {
	md := NewMatch(descr, player, name, NOTYPE)
	md.MatchAbsolute()
	md.MatchEverything()
	match = md.NoisyMatchResult()
	if match != NOTHING && !controls(player, match) {
		notify(player, "Permission denied. (You don't control what was matched)")
		match = NOTHING
	}
	return
}

func do_name(descr int, player dbref, name, newname string) {
	NoGuest("@name", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			switch {
			case newname == "":
				notify(player, "Give it what new name?")
			case Typeof(thing) == TYPE_PLAYER:
				var password string
				newname = strings.TrimLeftFunc(newname, unicode.IsSpace)
				if terms := strings.SplitN(newname, " ", 2); len(terms) == 2 {
					newname = terms[0]
					password = strings.TrimFunc(terms[1], unicode.IsSpace)
				}
				switch {
				case password == "":
					notify(player, "You must specify a password to change a player name.")
					notify(player, "E.g.: name player = newname password")
				case !check_password(thing, password):
					notify(player, "Incorrect password.")
				case newname != db.Fetch(thing).name && !ok_player_name(newname):
					notify(player, "You can't give a player that name.")
				default:
					/* everything ok, notify */
					log_status("NAME CHANGE: %s(#%d) to %s", db.Fetch(thing).name, thing, newname)
					delete_player(thing)
					ts_modifyobject(thing)
					db.Fetch(thing).name = newname
					add_player(thing)
					notify(player, "Name set.")
				}
			default:
				switch thing.(type) {
				case TYPE_THING:
					if !ok_ascii_thing(newname) {
						notify(player, "Invalid 8-bit name.")
						return
					}
				default:
					if !ok_ascii_other(newname) {
						notify(player, "Invalid 8-bit name.")
						return
					}
				}
				if !ok_name(newname) {
					notify(player, "That is not a reasonable name.")
				} else {
					ts_modifyobject(thing)
					db.Fetch(thing).name = newname
					notify(player, "Name set.")
					db.Fetch(thing).flags |= OBJECT_CHANGED
					switch thing.(type) {
					case TYPE_EXIT:
						if MLevRaw(thing) != NON_MUCKER {
							SetMLevel(thing, NON_MUCKER)
							notify(player, "Action priority Level reset to zero.")
						}
					}
				}
			}
		}
	})
}

func do_describe(descr int, player dbref, name, description string) {
	NoGuest("@describe", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			ts_modifyobject(thing)
			add_property(thing, MESGPROP_DESC, description, 0)
			if(description && *description) {
				notify(player, "Description set.")
			} else {
				notify(player, "Description cleared.")
			}
		}
	})
}

func do_idescribe(descr int, player dbref, name, description string) {
	NoGuest("@idescribe", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			ts_modifyobject(thing)
			add_property(thing, MESGPROP_IDESC, description, 0)
			if description != "" {
				notify(player, "Description set.")
			} else {
				notify(player, "Description cleared.")
			}
		}
	})
}

func do_doing(descr int, player dbref, name, message string) {
	NoGuest("@doing", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			ts_modifyobject(thing)
			add_property(thing, MESGPROP_DOING, message, 0)
			if message != "" {
				notify(player, "Doing set.")
			} else {
				notify(player, "Doing cleared.")
			}
		}
	})
}

func do_fail(descr int, player dbref, name, message string) {
	NoGuest("@fail", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			ts_modifyobject(thing)
			add_property(thing, MESGPROP_FAIL, message, 0)
			if message != "" {
				notify(player, "Message set.")
			} else {
				notify(player, "Message cleared.")
			}
		}
	})
}

func do_success(descr int, player dbref, name, message string) {
	NoGuest("@success", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			ts_modifyobject(thing)
			add_property(thing, MESGPROP_SUCC, message, 0)
			if message != "" {
				notify(player, "Message set.")
			} else {
				notify(player, "Message cleared.")
			}
		}
	})
}

/* sets the drop message for player */
func do_drop_message(descr int, player dbref, name, message string) {
	NoGuest("@drop", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			ts_modifyobject(thing)
			add_property(thing, MESGPROP_DROP, message, 0)
			if message != "" {
				notify(player, "Message set.")
			} else {
				notify(player, "Message cleared.")
			}
		}
	})
}

func do_osuccess(descr int, player dbref, name, message string) {
	NoGuest("@osuccess", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			ts_modifyobject(thing)
			add_property(thing, MESGPROP_OSUCC, message, 0)
			if message != "" {
				notify(player, "Message set.")
			} else {
				notify(player, "Message cleared.")
			}
		}
	})
}

func do_ofail(descr int, player dbref, name, message string) {
	NoGuest("@ofail", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			ts_modifyobject(thing)
			add_property(thing, MESGPROP_OFAIL, message, 0)
			if message != "" {
				notify(player, "Message set.")
			} else {
				notify(player, "Message cleared.")
			}
		}
	})
}

func do_odrop(descr int, player dbref, name, message string) {
	NoGuest("@odrop", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			ts_modifyobject(thing)
			add_property(thing, MESGPROP_ODROP, message, 0)
			if message != "" {
				notify(player, "Message set.")
			} else {
				notify(player, "Message cleared.")
			}
		}
	})
}

func do_oecho(descr int, player dbref, name, message string) {
	NoGuest("@oecho", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			ts_modifyobject(thing)
			add_property(thing, MESGPROP_OECHO, message, 0)
			if message != "" {
				notify(player, "Message set.")
			} else {
				notify(player, "Message cleared.")
			}
		}
	})
}

func do_pecho(descr int, player dbref, name, message string) {
	NoGuest("@pecho", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			ts_modifyobject(thing)
			add_property(thing, MESGPROP_PECHO, message, 0)
			if message != "" {
				notify(player, "Message set.")
			} else {
				notify(player, "Message cleared.")
			}
		}
	})
}

/* sets a lock on an object to the lockstring passed to it.
   If the lockstring is null, then it unlocks the object. */
func setlockstr(descr int, player, thing dbref, keyname string) (r bool) {
	if keyname != "" {
		lock := ParseLock(descr, player, keyname, 0)
		if r = !lock.IsTrue(); r {
			ts_modifyobject(thing)
			set_property(thing, MESGPROP_LOCK, LOCKED)
		}
	} else {
		ts_modifyobject(thing)
		set_property(thing, MESGPROP_LOCK, UNLOCKED)
		db.Fetch(thing).flags |= OBJECT_CHANGED
		r = true
	}
	return
}

void do_conlock(descr int, player dbref, name, keyname string) {
	NoGuest("@conlock", player, func() {
		md := NewMatch(descr, player, name, NOTYPE)
		md.MatchAbsolute()
		md.MatchEverything()
		switch thing := md.MatchResult(); {
		case NOTHING:
			notify(player, "I don't see what you want to set the container-lock on!")
			return
		case AMBIGUOUS:
			notify(player, "I don't know which one you want to set the container-lock on!")
			return
		case !controls(player, thing):
			notify(player, "You can't set the container-lock on that!")
			return
		case keyname == "":
			set_property(thing, "_/clk", UNLOCKED)
			ts_modifyobject(thing)
			notify(player, "Container lock cleared.")
		default:
			if lock := ParseLock(descr, player, keyname, 0); lock.IsTrue() {
				notify(player, "I don't understand that key.")
			} else {
				set_property(thing, "_/clk", lock)
				ts_modifyobject(thing)
				notify(player, "Container lock set.")
			}
		}
	})
}

func do_flock(descr int, player dbref, name, keyname string) {
	NoGuest("@force_lock", player, func() {
		md := NewMatch(descr, player, name, NOTYPE)
		md.MatchAbsolute()
		md.MatchEverything()
		switch thing := md.MatchResult(); {
		case NOTHING:
			notify(player, "I don't see what you want to set the force-lock on!")
		case AMBIGUOUS:
			notify(player, "I don't know which one you want to set the force-lock on!")
		case !controls(player, thing):
			notify(player, "You can't set the force-lock on that!")
		case force_level != 0:
			notify(player, "You can't use @flock from an @force or {force}.")
		case len(keyname) == 0:
			set_property(thing, "@/flk", UNLOCKED)
			ts_modifyobject(thing)
			notify(player, "Force lock cleared.")
		default:
			if lock := ParseLock(descr, player, keyname, 0); lock.IsTrue() {
				notify(player, "I don't understand that key.")
			} else {
				set_property(thing, "@/flk", lock)
				ts_modifyobject(thing)
				notify(player, "Force lock set.")
			}
		}
	})
}

func do_chlock(descr int, player dbref, name, keyname string) {
	NoGuest("@chown_lock", player, func() {
		md := NewMatch(descr, player, name, NOTYPE)
		md.MatchAbsolute()
		md.MatchEverything()
		switch thing := md.MatchResult(); {
		case thing == NOTHING:
			notify(player, "I don't see what you want to set the chown-lock on!")
		case thing == AMBIGUOUS:
			notify(player, "I don't know which one you want to set the chown-lock on!")
		case !controls(player, thing):
			notify(player, "You can't set the chown-lock on that!")
		case len(keyname) == 0:
			set_property(thing, "_/chlk", UNLOCKED)
			ts_modifyobject(thing)
			notify(player, "Chown lock cleared.")
		default:
			if lock := ParseLock(descr, player, keyname, 0); lock.IsTrue() {
				notify(player, "I don't understand that key.")
			} else {
				set_property(thing, "_/chlk", lock)
				ts_modifyobject(thing)
				notify(player, "Chown lock set.")
			}
		}
	})
}

func do_lock(descr int, player dbref, name, keyname string) {
	NoGuest("@lock", player, func() {
		md := NewMatch(descr, player, name, NOTYPE)
		md.MatchAbsolute()
		md.MatchEverything()
		switch thing := md.MatchResult(); {
		case thing == NOTHING:
			notify(player, "I don't see what you want to lock!")
		case thing == AMBIGUOUS:
			notify(player, "I don't know which one you want to lock!")
		case !controls(player, thing):
			notify(player, "You can't lock that!")
		case len(keyname) != 0:
			if lock := ParseLock(descr, player, keyname, 0); lock.IsTrue() {
				notify(player, "I don't understand that key.")
			} else {
				set_property(thing, MESGPROP_LOCK, lock)
				ts_modifyobject(thing)
				notify(player, "Locked.")
			}
		default:
			do_unlock(descr, player, name)
		}
	})
}

func do_unlock(descr int, player dbref, name string) {
	NoGuest("@unlock", player, func() {
		if thing = MatchControlled(descr, player, name); thing != NOTHING {
			ts_modifyobject(thing)
			set_property(thing, MESGPROP_LOCK, UNLOCKED)
			db.Fetch(thing).flags |= OBJECT_CHANGED
			notify(player, "Unlocked.")
		}
	})
}

func controls_link(dbref who, dbref what) (r bool) {
	p := db.Fetch(what)
	switch p := p.(type) {
	case Exit:
		for i, v := range p.Destinations {
			if controls(who, v) {
				r = true
				break
			}
		}
		r ||= who == db.FetchPlayer(p.Location).Owner
	case Room:
		r = controls(who, p.dbref)
	case Player:
		r = controls(who, p.home)
	case Object:
		r = controls(who, p.home)
	}
	return
}

/* like do_unlink, but if quiet is true, then only error messages are
   printed. */
func _do_unlink(int descr, dbref player, const char *name, int quiet) {
	exit := NewMatch(descr, player, name, IsExit).
		MatchAbsolute().
		MatchPlayer().
		MatchEverything().
		MatchResult()
	switch exit {
	case NOTHING:
		notify(player, "Unlink what?");
	case AMBIGUOUS:
		notify(player, "I don't know which one you mean!");
	default:
		if !controls(player, exit) && !controls_link(player, exit) {
			notify(player, "Permission denied. (You don't control the exit or its link)");
		} else {
			switch Typeof(exit) {
			case TYPE_EXIT:
				p := db.Fetch(exit)
				if len(p.(Exit).Destinations) != 0 {
					add_property(p.Owner, MESGPROP_VALUE, nil, get_property_value(p.Owner, MESGPROP_VALUE) + tp_link_cost)
					db.Fetch(p.Owner).flags |= OBJECT_CHANGED
				}
				ts_modifyobject(exit)
				if len(p.(Exit).Destinations) > 0 {
					p.(Exit).Destinations = nil
					p.flags |= OBJECT_CHANGED
				}
				if !quiet {
					notify(player, "Unlinked.")
				}
				if MLevRaw(exit) != NON_MUCKER {
					SetMLevel(exit, NON_MUCKER)
					p.flags |= OBJECT_CHANGED
					if !quiet {
						notify(player, "Action priority Level reset to 0.")
					}
				}
			case TYPE_ROOM:
				ts_modifyobject(exit)
				db.Fetch(exit).sp = NOTHING
				db.Fetch(exit).flags |= OBJECT_CHANGED
				if !quiet {
					notify(player, "Dropto removed.")
				}
			case TYPE_THING:
				ts_modifyobject(exit)
				db.Fetch(exit).(Thing).home = db.Fetch(exit).Owner
				db.Fetch(exit).(Thing).flags |= OBJECT_CHANGED
				if !quiet {
					notify(player, "Thing's home reset to owner.")
				}
			case TYPE_PLAYER:
				ts_modifyobject(exit)
				p := db.FetchPlayer(exit)
				p.home = tp_player_start
				p.flags |= OBJECT_CHANGED
				if !quiet {
					notify(player, "Player's home reset to default player start room.")
				}
			default:
				notify(player, "You can't unlink that!")
			}
		}
	}
}

func do_unlink(descr int, player dbref, name string) {
	NoGuest("@unlink", player, func() {
		/* do a regular, non-quiet unlink. */
		_do_unlink(descr, player, name, 0)
	})
}

func do_unlink_quiet(descr int, player dbref, name string) {
	_do_unlink(descr, player, name, 1)
}

/*
 * do_relink()
 *
 * re-link an exit object. FIXME: this shares some code with do_link() which
 * should probably be moved into a separate function (is_link_ok() or
 * something like that).
 *
 */
func do_relink(descr int, player dbref, thing_name, dest_name string) {
	var dest dbref

	NoGuest("@relink", player, func() {
		md := NewMatch(descr, player, thing_name, IsExit).
			MatchAllExits().
			MatchNeighbor().
			MatchPossession().
			MatchMe().
			MatchHere().
			MatchAbsolute().
			MatchRegistered()
		if Wizard(db.Fetch(player).Owner) {
			md.MatchPlayer()
		}
		if thing := md.NoisyMatchResult(); thing != NOTHING {
			/* first of all, check if the new target would be valid, so we can avoid breaking the old link if it isn't. */
			switch thing.(type) {
			case TYPE_EXIT:
				/* we're ok, check the usual stuff */
				if len(db.Fetch(thing).(Exit).Destinations) != 0 {
					if !controls(player, thing) {
						notify(player, "Permission denied. (The exit is linked, and you don't control it)")
						return
					}
				} else {
					if !Wizard(db.Fetch(player).Owner) && get_property_value(player, MESGPROP_VALUE) < (tp_link_cost + tp_exit_cost) {
						if cost := tp_link_cost + tp_exit_cost; cost == 1 {
							notify_fmt(player, "It costs %d %s to link this exit.", cost, tp_penny)
						} else {
							notify_fmt(player, "It costs %d %s to link this exit.", cost, tp_pennies)
						}
						return
					} else if !Builder(player) {
						notify(player, "Only authorized builders may seize exits.")
						return
					}
				}

				/* be anal: each and every new links destination has to be ok. Detailed error messages are given by link_exit_dry(). */
				var good_dest []dbref
				if n := link_exit_dry(descr, player, thing, dest_name, good_dest); n == 0 {
					notify(player, "Invalid target.")
					return
				}
			case TYPE_THING, TYPE_PLAYER:
				md := NewMatch(descr, player, dest_name, IsRoom).
					MatchNeighbor().
					MatchAbsolute().
					MatchRegistered().
					MatchMe().
					MatchHere()
				if Typeof(thing) == TYPE_THING {
					md.MatchPossession()
				}
				if dest = md.NoisyMatchResult()); dest != NOTHING {
					if !controls(player, thing) || !can_link_to(player, Typeof(thing), dest) {
						notify(player, "Permission denied. (You can't link to where you want to.")
						return
					}
					if parent_loop_check(thing, dest) {
						notify(player, "That would cause a parent paradox.")
						return
					}
				} else {
					return
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
					return
				case !controls(player, thing), !can_link_to(player, Typeof(thing), dest), thing == dest:
					notify(player, "Permission denied. (You can't link to the dropto like that)")
					return
				}
			case TYPE_PROGRAM:
				notify(player, "You can't link programs to things!")
				return
			default:
				notify(player, "Internal error: weird object type.")
				log_status("PANIC: weird object: Typeof(%d) = %d", thing, Typeof(thing))
				return
			}
			do_unlink_quiet(descr, player, thing_name)
			notify(player, "Attempting to relink...")
			do_link(descr, player, thing_name, dest_name)
		}
	})
}

func do_chown(int descr, dbref player, const char *name, const char *newowner) {
	dbref thing;
	dbref owner;

	if (!*name) {
		notify(player, "You must specify what you want to take ownership of.");
		return;
	}
	md := NewMatch(descr, player, name, NOTYPE)
	md.MatchEverything()
	md.MatchAbsolute()
	if ((thing = md.NoisyMatchResult()) == NOTHING)
		return;

	if newowner != "me" {
		if ((owner = lookup_player(newowner)) == NOTHING) {
			notify(player, "I couldn't find that player.");
			return;
		}
	} else {
		owner = db.Fetch(player).Owner
	}
	if !Wizard(db.Fetch(player).Owner) && db.Fetch(player).Owner != owner {
		notify(player, "Only wizards can transfer ownership to others.");
		return;
	}
	if Wizard(db.Fetch(player).Owner) && player != GOD && owner == GOD {
		notify(player, "God doesn't need an offering or sacrifice.");
		return;
	}
	if !Wizard(db.Fetch(player).Owner) {
		if TYPEOF(thing) != TYPE_EXIT || (len(db.Fetch(thing).(Exit).Destinations) != 0 && !controls_link(player, thing)) {
			if db.Fetch(thing).flags & CHOWN_OK == 0 || TYPEOF(thing) == TYPE_PROGRAM || !test_lock(descr, player, thing, "_/chlk") {
				notify(player, "You can't take possession of that.")
				return
			}
		}
	}

	if tp_realms_control && !Wizard(db.Fetch(player).Owner) && TrueWizard(thing) && Typeof(thing) == TYPE_ROOM {
		notify(player, "You can't take possession of that.");
		return;
	}

	switch Typeof(thing) {
	case TYPE_ROOM:
		if !Wizard(db.Fetch(player).Owner) && db.Fetch(player).Location != thing {
			notify(player, "You can only chown \"here\".")
			return
		}
		ts_modifyobject(thing)
		db.Fetch(thing).Owner = db.Fetch(owner).Owner
	case TYPE_THING:
		if !Wizard(db.Fetch(player).Owner) && db.Fetch(thing).Location != player {
			notify(player, "You aren't carrying that.")
			return
		}
		ts_modifyobject(thing)
		db.Fetch(thing).Owner = db.Fetch(owner).Owner
	case TYPE_PLAYER:
		notify(player, "Players always own themselves.")
		return
	case TYPE_EXIT, TYPE_PROGRAM:
		ts_modifyobject(thing);
		db.Fetch(thing).Owner = db.Fetch(owner).Owner;
	}
	if owner == player {
		notify(player, "Owner changed to you.")
	} else {
		notify(player, fmt.Sprintf("Owner changed to %s.", unparse_object(player, owner)))
	}
	db.Fetch(thing).flags |= OBJECT_CHANGED
}


/* Note: Gender code taken out.  All gender references are now to be handled
   by property lists...
   Setting of flags and property code done here.  Note that the PROP_DELIMITER
   identifies when you're setting a property.
   A @set <thing>= :clear
   will clear all properties.
   A @set <thing>= type:
   will remove that property.
   A @set <thing>= propname:string
   will add that string property or replace it.
   A @set <thing>= propname:^value
   will add that integer property or replace it.
 */

func do_set(descr int, player dbref, name, flag string) {
	dbref thing;
	const char *p;
	object_flag_type f;

	NoGuest("@set", player, func() {
		if ((thing = MatchControlled(descr, player, name)) == NOTHING)
			return;
		/* Only God can set anything on any of his stuff */
		if player != GOD && db.Fetch(thing).Owner == GOD {
			notify(player,"Only God may touch God's property.");
			return;
		}

		/* move p past NOT_TOKEN if present */
		p = strings.TrimLeftFunc(flag, func(r rune) bool {
			return unicode.IsSpace(r) || r == NOT_TOKEN
		})
		/* Now we check to see if it's a property reference */
		/* if this gets changed, please also modify boolexp.c */
		if (strchr(flag, PROP_DELIMITER)) {
			/* copy the string so we can muck with it */
			char *type = flag;	/* type */
			char *pname = (char *) strchr(type, PROP_DELIMITER);	/* propname */
			char *x;				/* to preserve string location so we can free it */
			char *temp;
			int ival = 0;

			x = type;
			type = strings.TrimLeftFunc(type, func(r rune) bool {
				return unicode.IsSpace(r) && r != PROP_DELIMITER
			})

			if *type == PROP_DELIMITER {
				/* clear all properties */
				type = strings.TrimLeftFunc(type, unicode.IsSpace)
				if type != "clear" {
					notify(player, "Use '@set <obj>=:clear' to clear all props on an object");
					free((void *)x);
					return;
				}
				remove_property_list(thing, Wizard(db.Fetch(player).Owner));
				ts_modifyobject(thing);
				notify(player, "All user-owned properties removed.");
				free((void *) x);
				return;
			}
			/* get rid of trailing spaces and slashes */
			for (temp = pname - 1; temp >= type && unicode.IsSpace(*temp); temp--) ;
			while (temp >= type && *temp == '/')
				temp--;
			*(++temp) = '\0';

			pname++;				/* move to next character */
			/* while (unicode.IsSpace(*pname) && *pname) pname++; */
			if (*pname == '^' && unicode.IsNumber(pname + 1))
				ival = atoi(++pname);

			if Prop_System(type) || (!Wizard(db.Fetch(player).Owner) && (Prop_SeeOnly(type) || Prop_Hidden(type))) {
				notify(player, "Permission denied. (The property is hidden from you.)")
				free((void *)x)
				return
			}

			if !(*pname) {
				ts_modifyobject(thing)
				remove_property(thing, type)
				notify(player, "Property removed.")
			} else {
				ts_modifyobject(thing)
				if ival != 0 {
					add_property(thing, type, nil, ival)
				} else {
					add_property(thing, type, pname, 0)
				}
				notify(player, "Property set.")
			}
			free((void *) x)
			return
		}
		/* identify flag */
		switch {
		case p == "":
			notify(player, "You must specify a flag to set.")
			return
		case p == "0", p == "M0", strings.Prefix(p, "MUCKER") && *flag == NOT_TOKEN:
			if !Wizard(db.Fetch(player).Owner) {
				if db.Fetch(player).Owner != db.Fetch(thing).Owner || Typeof(thing) != TYPE_PROGRAM {
					notify(player, "Permission denied. (You can't clear that mucker flag)");
					return;
				}
			}
			if (force_level) {
				notify(player, "Can't set this flag from an @force or {force}.");
				return;
			}
			SetMLevel(thing, NON_MUCKER);
			notify(player, "Mucker level set.");
			return;
		case p == "1", p == "M1":
			if !Wizard(db.Fetch(player).Owner) {
				if db.Fetch(player).Owner != db.Fetch(thing).Owner || Typeof(thing) != TYPE_PROGRAM || MLevRaw(player) < APPRENTICE {
					notify(player, "Permission denied. (You may not set that M1)");
					return;
				}
			}
			if (force_level) {
				notify(player, "Can't set this flag from an @force or {force}.");
				return;
			}
			SetMLevel(thing, APPRENTICE);
			notify(player, "Mucker level set.");
			return;
		case p == "2", p == "M2", strings.Prefix(p, "MUCKER") && *flag != NOT_TOKEN:
			if !Wizard(db.Fetch(player).Owner) {
				if db.Fetch(player).Owner != db.Fetch(thing).Owner || Typeof(thing) != TYPE_PROGRAM || MLevRaw(player) < JOURNEYMAN {
					notify(player, "Permission denied. (You may not set that M2)");
					return;
				}
			}
			if force_level {
				notify(player, "Can't set this flag from an @force or {force}.")
			} else {
				SetMLevel(thing, JOURNEYMAN)
				notify(player, "Mucker level set.")
			}
			return;
		case p == "3", p == "M3":
			if !Wizard(db.Fetch(player).Owner) {
				if db.Fetch(player).Owner != db.Fetch(thing).Owner || Typeof(thing) != TYPE_PROGRAM || MLevRaw(player) < MASTER {
					notify(player, "Permission denied. (You may not set that M3)");
					return
				}
			}
			if force_level {
				notify(player, "Can't set this flag from an @force or {force}.");
			} else {
				SetMLevel(thing, MASTER);
				notify(player, "Mucker level set.")
			}
			return
		case p == "4", p == "M4":
			notify(player, "To set Mucker Level 4, set the Wizard bit and another Mucker bit.");
			return;
		case strings.Prefix(p, "WIZARD"):
			if force_level {
				notify(player, "Can't set this flag from an @force or {force}.");
				return
			}
			f = WIZARD
		case strings.Prefix(p, "ZOMBIE"):
			f = ZOMBIE
		case strings.Prefix(p, "VEHICLE"), strings.Prefix(p, "VIEWABLE"):
			if (*flag == NOT_TOKEN && Typeof(thing) == TYPE_THING) {
				for obj := db.Fetch(thing).Contents; obj != NOTHING; obj = db.Fetch(obj).next {
					if TYPEOF(obj) == TYPE_PLAYER {
						notify(player, "That vehicle still has players in it!")
						return
					}
				}
			}
			f = VEHICLE
		case strings.Prefix(p, "LINK_OK"):
			f = LINK_OK
		case strings.Prefix(p, "XFORCIBLE"), strings.Prefix(p, "XPRESS"):
			if force_level {
				notify(player, "Can't set this flag from an @force or {force}.")
				return
			}
			if Typeof(thing) == TYPE_EXIT {
				if !Wizard(db.Fetch(player).Owner) {
					notify(player, "Permission denied. (Only a Wizard may set the M-level of an exit)");
					return;
				}
			}
			f = XFORCIBLE
		case strings.Prefix(p, "KILL_OK"):
			f = KILL_OK
		case strings.Prefix(p, "DARK"), strings.Prefix(p, "DEBUG"):
			f = DARK
		case strings.Prefix(p, "STICKY"), strings.Prefix(p, "SETUID"), strings.Prefix(p, "SILENT"):
			f = STICKY
		case strings.Prefix(p, "QUELL"):
			f = QUELL
		case strings.Prefix(p, "BUILDER"), strings.Prefix(p, "BOUND"):
			f = BUILDER
		case strings.Prefix(p, "CHOWN_OK"), strings.Prefix(p, "COLOR"):
			f = CHOWN_OK
		case strings.Prefix(p, "JUMP_OK"):
			f = JUMP_OK
		case strings.Prefix(p, "HAVEN"), strings.Prefix(p, "HARDUID"):
			f = HAVEN
		case strings.Prefix(p, "ABODE"), strings.Prefix(p, "AUTOSTART"), strings.Prefix(p, "ABATE"):
			f = ABODE
		case strings.Prefix(p, "YIELD") && tp_enable_match_yield && (Typeof(thing) == TYPE_ROOM || Typeof(thing) == TYPE_THING):
			f = YIELD
		case strings.Prefix(p, "OVERT") && tp_enable_match_yield && (Typeof(thing) == TYPE_ROOM || Typeof(thing) == TYPE_THING):
			f = OVERT
		default:
			notify(player, "I don't recognize that flag.")
			return
		}

		switch {
		case restricted(player, thing, f):
			notify(player, "Permission denied. (restricted flag)")
			return
		case f == WIZARD && *flag == NOT_TOKEN && thing == player:
			/* check for stupid wizard */
			notify(player, "You cannot make yourself mortal.")
		case *flag == NOT_TOKEN:
			ts_modifyobject(thing)
			db.Fetch(thing).flags &= ~f
			db.Fetch(thing).flags |= OBJECT_CHANGED
			notify(player, "Flag reset.")
		default:
			ts_modifyobject(thing)
			db.Fetch(thing).flags |= f
			db.Fetch(thing).flags |= OBJECT_CHANGED
			notify(player, "Flag set.")
		}
	})
}

func do_propset(descr int, player dbref, name, prop string) {
	NoGuest("@propset", player, func() {
		if thing := MatchControlled(descr, player, name); thing != NOTHING {
			terms := strings.Split(prop, PROP_DELIMITER)
			if len(terms) < 3 {
				notify(player, "Not enough arguments to @propset!")
			} else {
				datatype := strings.TrimSpace(terms[0])
				pname := strings.TrimSpace(terms[1])
				value := strings.TrimSpace(terms[2])

				if Prop_System(pname) || (!Wizard(db.Fetch(player).Owner) && (Prop_SeeOnly(pname) || Prop_Hidden(pname))) {
					notify(player, "Permission denied. (can't set a property that's restricted against you)")
					return
				} else {
					switch {
					case datatype == "", strings.HasPrefix(datatype, "string"):
						add_property(thing, pname, value, 0)
						notify(player, "Property set.")
					case strings.HasPrefix(datatype, "integer"):
						if !unicode.IsNumber(value) {
							add_property(thing, pname, nil, strconv.Atoi(value))
							notify(player, "Property set.")
						} else {
							notify(player, "That's not an integer!")
						}
					case strings.HasPrefix(datatype, "float"):
						if f, e := strconv.ParseFloat(value, 64); e == nil {
							set_property(thing, pname, f)
							notify(player, "Property set.")
						} else {
							notify(player, "That's not a floating point number!")
						}
					case strings.HasPrefix(datatype, "dbref"):
						md := NewMatch(descr, player, value, NOTYPE)
						md.MatchAbsolute()
						md.MatchEverything()
						if ref := md.NoisyMatchResult(); ref != NOTHING {
							set_property(thing, pname, ref)
							notify(player, "Property set.")
						}
					case strings.HasPrefix(datatype, "lock"):
						if lock := ParseLock(descr, player, value, 0); lock.IsTrue() {
							notify(player, "I don't understand that lock.")
						} else {
							set_property(thing, pname, lock)
							notify(player, "Property set.")
						}
					case strings.HasPrefix(datatype, "erase"):
						if value == "" {
							remove_property(thing, pname)
							notify(player, "Property erased.")
						} else {
							notify(player, "Don't give a value when erasing a property.")
						}
					default:
						notify(player, "I don't know what type of property you want to set!")
						notify(player, "Valid types are string, integer, float, dbref, lock, and erase.")
					}
				}
			}
		}
	})
}

func set_flags_from_tunestr(obj dbref, tunestr string) {
	for f := db.Fetch(obj).flags; len(tunestr) > 0; tunestr = tunestr[1:] {
		switch pcc := strings.ToUpper(tunestr[0]); pcc {
		case '\n', '\r':
			break
		case '0':
			SetMLevel(obj, NON_MUCKER)
		case '1':
			SetMLevel(obj, APPRENTICE)
		case '2':
			SetMLevel(obj, JOURNEYMAN)
		case '3':
			SetMLevel(obj, MASTER)
		case 'A':
			f |= ABODE
		case 'B':
			f |= BUILDER
		case 'C':
			f |= CHOWN_OK
		case 'D':
			f |= DARK
		case 'H':
			f |= HAVEN
		case 'J':
			f |= JUMP_OK
		case 'K':
			f |= KILL_OK
		case 'L':
			f |= LINK_OK
		case 'M':
			SetMLevel(obj, JOURNEYMAN)
		case 'Q':
			f |= QUELL
		case 'S':
			f |= STICKY
		case 'V':
			f |= VEHICLE
		case 'W':
			/* f |= WIZARD;     This is very bad to auto-set. */
		case 'X':
			f |= XFORCIBLE;
		case 'Y':
			f |= YIELD
		case 'O':
			f |= OVERT
		case 'Z':
			f |= ZOMBIE
		}
	}
	ts_modifyobject(obj)
	f |= OBJECT_CHANGED
}