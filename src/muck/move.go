package fbmuck

func moveto(what, where dbref) {
	var loc dbref

	if loc = db.Fetch(what).location; loc != NOTHING {
		db.Fetch(loc).contents = remove_first(db.Fetch(loc).contents, what)
		db.Fetch(loc).flags |= OBJECT_CHANGED
	}

	switch where {
	case NOTHING:
		db.Fetch(what).location = NOTHING
		db.Fetch(what).flags |= OBJECT_CHANGED
		return					/* NOTHING doesn't have contents */
	case HOME:
		switch Typeof(what) {
		case TYPE_PLAYER:
			where = db.Fetch(what).sp.(player_specific).home
		case TYPE_THING:
			where = db.Fetch(what).sp.(player_specific).home
			if parent_loop_check(what, where) {
				where = db.Fetch(db.Fetch(what).owner).sp.(player_specific).home
				if parent_loop_check(what, where) {
					where = dbref(tp_player_start)
				}
			}
		case TYPE_ROOM:
			where = GLOBAL_ENVIRONMENT
		case TYPE_PROGRAM:
			where = db.Fetch(what).owner
		}
	default:
		if parent_loop_check(what, where) {
			switch Typeof(what) {
			case TYPE_PLAYER:
				where = db.Fetch(what).sp.(player_specific).home
			case TYPE_THING:
				where = db.Fetch(what).sp.(player_specific).home
				if parent_loop_check(what, where) {
					where = db.Fetch(db.Fetch(what).owner).sp.(player_specific).home
					if parent_loop_check(what, where) {
						where = (dbref) tp_player_start
					}
				}
			case TYPE_ROOM:
				where = GLOBAL_ENVIRONMENT
			case TYPE_PROGRAM:
				where = db.Fetch(what).owner
			}
		}
	}
	db.Fetch(what).next = db.Fetch(where).contents
	db.Fetch(what).flags |= OBJECT_CHANGED
	db.Fetch(where).contents = what
	db.Fetch(where).flags |= OBJECT_CHANGED
	db.Fetch(what).location = where
	db.Fetch(what).flags |= OBJECT_CHANGED
}

func send_contents(int descr, dbref loc, dbref dest) {
	first := db.Fetch(loc).contents
	db.Fetch(loc).contents = NOTHING
	db.Fetch(loc).flags |= OBJECT_CHANGED

	/* blast locations of everything in list */
	for rest := first; rest != NOTHING; rest = db.Fetch(rest).next {
		db.Fetch(rest).location = NOTHING
		db.Fetch(rest).flags |= OBJECT_CHANGED
	}

	for first != NOTHING {
		rest := db.Fetch(first).next
		if Typeof(first) != TYPE_THING && Typeof(first) != TYPE_PROGRAM {
			moveto(first, loc)
		} else {
			where := dest
			if db.Fetch(first).flags & STICKY != 0 {
				where = HOME
			}
			if tp_thing_movement && Typeof(first) == TYPE_THING {
				if parent_loop_check(first, where) {
					enter_room(descr, first, loc, db.Fetch(first).location)
				} else {
					enter_room(descr, first, where, db.Fetch(first).location)
				}
			} else {
				if parent_loop_check(first, where) {
					moveto(first, loc)
				} else {
					moveto(first, where)
				}
			}
		}
		first = rest
	}
	db.Fetch(loc).contents = reverse(db.Fetch(loc).contents)
	db.Fetch(loc).flags |= OBJECT_CHANGED
}

func maybe_dropto(descr int, loc, dropto dbref) {
	if loc != dropto {
		/* check for players */
		for thing := db.Fetch(loc).contents; thing != NOTHING; thing = db.Fetch(thing).next {
			/* Make zombies act like players for dropto processing */
			if Typeof(thing) == TYPE_PLAYER || (Typeof(thing) == TYPE_THING && db.Fetch(thing).flags & ZOMBIE != 0) {
				return
			}
		}

		/* no players, send everything to the dropto */
		send_contents(descr, loc, dropto)
	}
}

/* What are we doing here?  Quick explanation - we want to prevent
   environment loops from happening.  Any item should always be able
   to 'find' its way to room #0.  Since the loop check is recursive,
   we also put in a max iteration check, to keep people from creating
   huge envchains in order to bring the server down.  We have a loop
   if we:
   a) Try to parent to ourselves.
   b) Parent to nothing (not really a loop, but won't get you to #0).
   c) Parent to our own home (not a valid destination).
   d) Find our source room down the environment chain.
   Note: This system will only work if every step _up_ to this point has
   resulted in a consistent (ie: no loops) environment.
*/

int
location_loop_check(dbref source, dbref dest)
{   
  unsigned int level = 0;
  unsigned int place = 0;
  dbref pstack[MAX_PARENT_DEPTH+2];

  if (source == dest) {
    return 1;
  }
  pstack[0] = source;
  pstack[1] = dest;

  while (level < MAX_PARENT_DEPTH) {
    dest = db.Fetch(dest).location
    if (dest == NOTHING) {
      return 0;
    }
    if (dest == HOME) {        /* We should never get this, either. */
      return 1;
    }
    if (dest == (dbref) 0) {   /* Reached the top of the chain. */
      return 0;
    }
    /* Check to see if we've found this item before.. */
    for (place = 0; place < (level+2); place++) {
      if (pstack[place] == dest) {
        return 1;
      }
    }
    pstack[level+2] = dest;
    level++;
  }
  return 1;
}

int
parent_loop_check(dbref source, dbref dest)
{   
  unsigned int level = 0;
  unsigned int place = 0;
  dbref pstack[MAX_PARENT_DEPTH+2];

  if (dest == HOME) {
		  switch Typeof(source) {
		  case TYPE_PLAYER:
			  dest = db.Fetch(source).sp.(player_specific).home
		  case TYPE_THING:
			  dest = db.Fetch(source).sp.(player_specific).home
		  case TYPE_ROOM:
			  dest = GLOBAL_ENVIRONMENT;
		  case TYPE_PROGRAM:
			  dest = db.Fetch(source).owner
		  default:
			  return 1;
	  }
  }
  if (location_loop_check(source, dest)) {
	  return 1;
  }

  if (source == dest) {
    return 1;
  }
  pstack[0] = source;
  pstack[1] = dest;

  while (level < MAX_PARENT_DEPTH) {
    /* if (Typeof(dest) == TYPE_THING) {
         dest = db.Fetch(dest).sp.(player_specific).home
       } */
    dest = getparent(dest);
    if (dest == NOTHING) {
      return 0;
    }
    if (dest == HOME) {        /* We should never get this, either. */
      return 1;
    }
    if (dest == (dbref) 0) {   /* Reached the top of the chain. */
      return 0;
    }
    /* Check to see if we've found this item before.. */
    for (place = 0; place < (level+2); place++) {
      if (pstack[place] == dest) {
        return 1;
      }
    }
    pstack[level+2] = dest;
    level++;
  }
  return 1;
}

static int donelook = 0;
func enter_room(int descr, dbref player, dbref loc, dbref exit) {
	dbref old;
	dbref dropto;
	char buf[BUFFER_LEN];

	if loc == HOME {
		loc = db.Fetch(player).sp.(player_specific).home
	}

	/* get old location */
	old = db.Fetch(player).location

	if parent_loop_check(player, loc) {
	  switch (Typeof(player)) {
	  case TYPE_PLAYER:
	    loc = db.Fetch(player).sp.(player_specific).home
	    break;
	  case TYPE_THING:
	    loc = db.Fetch(player).sp.(player_specific).home
	    if parent_loop_check(player, loc) {
	      loc = db.Fetch(db.Fetch(player).owner).sp.(player_specific).home
	      if parent_loop_check(player, loc) {
			  loc = (dbref) tp_player_start;
		  }
	    }
	    break;
	  case TYPE_ROOM:
	    loc = GLOBAL_ENVIRONMENT;
	    break;
	  case TYPE_PROGRAM:
	    loc = db.Fetch(player).owner
	    break;
	  }
	}

	/* check for self-loop */
	/* self-loops don't do move or other player notification */
	/* but you still get autolook and penny check */
	if (loc != old) {

		/* go there */
		moveto(player, loc);

		if (old != NOTHING) {
			propqueue(descr, player, old, exit, player, NOTHING, "_depart", "Depart", 1, 1);
			envpropqueue(descr, player, old, exit, old, NOTHING, "_depart", "Depart", 1, 1);

			propqueue(descr, player, old, exit, player, NOTHING, "_odepart", "Odepart", 1, 0);
			envpropqueue(descr, player, old, exit, old, NOTHING, "_odepart", "Odepart", 1, 0);

			/* notify others unless DARK */
			if !Dark(old) && !Dark(player) && (Typeof(player) != TYPE_THING || ((Typeof(player) == TYPE_THING) && (db.Fetch(player).flags & (ZOMBIE | VEHICLE) != 0))) && (Typeof(exit) != TYPE_EXIT || !Dark(exit)) {
#if !defined(QUIET_MOVES)
				buf = fmt.Sprintf("%s has left.", db.Fetch(player).name)
				notify_except(db.Fetch(old).contents, player, buf, player)
#endif
			}
		}

		/* if old location has STICKY dropto, send stuff through it */
		if old != NOTHING && Typeof(old) == TYPE_ROOM {
			if dropto = db.Fetch(old).sp.(dbref)); dropto != NOTHING && db.Fetch(old).flags & STICKY != 0 {
				maybe_dropto(descr, old, dropto)
			}
		}

		/* tell other folks in new location if not DARK */
		if !Dark(loc) && !Dark(player) && ((Typeof(player) != TYPE_THING) || (Typeof(player) == TYPE_THING && db.Fetch(player).flags & (ZOMBIE | VEHICLE) != 0)) && (Typeof(exit) != TYPE_EXIT || !Dark(exit)) {
#if !defined(QUIET_MOVES)
			buf = fmt.Sprintf("%s has arrived.", db.Fetch(player).name)
			notify_except(db.Fetch(loc).contents, player, buf, player);
#endif
		}
	}
	/* autolook */
	if Typeof(player) != TYPE_THING || (Typeof(player) == TYPE_THING && db.Fetch(player).flags & (ZOMBIE | VEHICLE) != 0) {
		if donelook < 8 {
			donelook++;
			if (can_move(descr, player, tp_autolook_cmd, 1)) {
				do_move(descr, player, tp_autolook_cmd, 1);
			} else {
				do_look_around(descr, player);
			}
			donelook--;
		} else {
			notify(player, "Look aborted because of look action loop.");
		}
	}

	if (tp_penny_rate != 0) {
		/* check for pennies */
		if !controls(player, loc) && get_property_value(db.Fetch(player).owner, MESGPROP_VALUE) <= tp_max_pennies && RANDOM() % tp_penny_rate == 0 {
			notify_fmt(player, "You found one %s!", tp_penny)
			add_property(db.Fetch(player).owner, MESGPROP_VALUE, nil, get_property_value(db.Fetch(player).owner, MESGPROP_VALUE) + 1)
			db.Fetch(db.Fetch(player).owner).flags |= OBJECT_CHANGED
		}
	}

	if (loc != old) {
		envpropqueue(descr, player, loc, exit, player, NOTHING, "_arrive", "Arrive", 1, 1);
		envpropqueue(descr, player, loc, exit, player, NOTHING, "_oarrive", "Oarrive", 1, 0);
	}
}

func send_home(descr int, thing dbref, puppethome int) {
	switch Typeof(thing) {
	case TYPE_PLAYER:
		/* send his possessions home first! */
		/* that way he sees them when he arrives */
		send_contents(descr, thing, HOME)
		enter_room(descr, thing, db.Fetch(thing).sp.(player_specific).home, db.Fetch(thing).location)
	case TYPE_THING:
		if puppethome {
			send_contents(descr, thing, HOME)
		}
		if tp_thing_movement || db.Fetch(thing).flags & (ZOMBIE | LISTENER) != 0 {
			enter_room(descr, thing, db.Fetch(thing).sp.(player_specific).home, db.Fetch(thing).location)
		} else {
			moveto(thing, HOME)		/* home */
		}
	case TYPE_PROGRAM:
		moveto(thing, db.Fetch(thing).owner)
	}
}

func can_move(descr int, player dbref , direction string, lev int) (r bool) {
	if tp_allow_home && direction == "home" {
		r = true
	} else {
		/* otherwise match on exits */
		md := NewMatch(descr, player, direction, TYPE_EXIT)
		md.level = lev
		md.MatchAllExits()
		r = md.LastMatchResult() != NOTHING
	}
	return
}

/*
 * trigger()
 *
 * This procedure triggers a series of actions, or meta-actions
 * which are contained in the 'dest' field of the exit.
 * Locks other than the first one are over-ridden.
 *
 * `player' is the player who triggered the exit
 * `exit' is the exit triggered
 * `pflag' is a flag which indicates whether player and room exits
 * are to be used (non-zero) or ignored (zero).  Note that
 * player/room destinations triggered via a meta-link are
 * ignored.
 *
 */

func trigger(int descr, dbref player, dbref exit, int pflag) {
	int sobjact;				/* sticky object action flag, sends home source obj */
	sobjact = 0;

	var succ bool
	for i, dest := range db.Fetch(exit).sp.exit.dest {
		if dest == HOME {
			dest = db.Fetch(player).sp.(player_specific).home

			/* fix #1112946 temporarily -- premchai21 */
			if Typeof(dest) == TYPE_THING {
				notify(player, "That would be an undefined operation.");
				continue;
			}
		}
		switch Typeof(dest) {
		case TYPE_ROOM:
			if pflag {
				if (parent_loop_check(player, dest)) {
					notify(player, "That would cause a paradox.");
					break;
				}
				if !Wizard(db.Fetch(player).owner) && Typeof(player) == TYPE_THING && db.Fetch(dest).flags & ZOMBIE != 0 {
					notify(player, "You can't go that way.");
					break;
				}
				if db.Fetch(player).flags & VEHICLE != 0 && (db.Fetch(dest).flags | db.Fetch(exit).flags) & VEHICLE != 0 {
					notify(player, "You can't go that way.");
					break;
				}
				if get_property_class(exit, MESGPROP_DROP) {
					exec_or_notify_prop(descr, player, exit, MESGPROP_DROP, "(@Drop)")
				}
				if get_property_class(exit, MESGPROP_ODROP) && !Dark(player) {
					parse_oprop(descr, player, dest, exit, MESGPROP_ODROP, db.Fetch(player).name, "(@Odrop)")
				}
				enter_room(descr, player, dest, exit)
				succ = true
			}
		case TYPE_THING:
			if dest == db.Fetch(exit).location && db.Fetch(dest).flags & VEHICLE != 0 {
				if pflag {
					if parent_loop_check(player, dest) {
						notify(player, "That would cause a paradox.")
						break
					}
					if get_property_class(exit, MESGPROP_DROP) {
						exec_or_notify_prop(descr, player, exit, MESGPROP_DROP, "(@Drop)")
					}
					if get_property_class(exit, MESGPROP_ODROP) && !Dark(player) {
						parse_oprop(descr, player, dest, exit, MESGPROP_ODROP, db.Fetch(player).name, "(@Odrop)")
					}
					enter_room(descr, player, dest, exit)
					succ = true
				}
			} else {
				if TYPEOF(db.Fetch(exit).location) == TYPE_THING {
					if parent_loop_check(dest, db.Fetch(db.Fetch(exit).location).location) {
						notify(player, "That would cause a paradox.")
						break
					}
					if tp_thing_movement {
						enter_room(descr, dest, db.Fetch(db.Fetch(exit).location).location, exit);
					} else {
						moveto(dest, db.Fetch(db.Fetch(exit).location).location)
					}
					if db.Fetch(exit).flags & STICKY == 0 {
						/* send home source object */
						sobjact = 1
					}
				} else {
					if parent_loop_check(dest, db.Fetch(exit).location) {
						notify(player, "That would cause a paradox.")
						break
					}
					if tp_thing_movement {
						enter_room(descr, dest, db.Fetch(exit).location, exit)
					} else {
						moveto(dest, db.Fetch(exit).location)
					}
				}
				if get_property_class(exit, MESGPROP_SUCC) {
					succ = true
				}
			}
			break
		case TYPE_EXIT:		/* It's a meta-link(tm)! */
			ts_useobject(dest)
			trigger(descr, player, db.Fetch(exit).sp.exit.dest[i], 0)
			if get_property_class(exit, MESGPROP_SUCC) {
				succ = true
			}
			break
		case TYPE_PLAYER:
			if pflag && db.Fetch(dest).location != NOTHING {
				if parent_loop_check(player, dest) {
					notify(player, "That would cause a paradox.")
					break
				}
				succ = true
				if db.Fetch(dest).flags & JUMP_OK != 0 {
					if get_property_class(exit, MESGPROP_DROP) {
						exec_or_notify_prop(descr, player, exit, MESGPROP_DROP, "(@Drop)")
					}
					if get_property_class(exit, MESGPROP_ODROP) && !Dark(player) {
						parse_oprop(descr, player, db.Fetch(dest).location, exit, MESGPROP_ODROP, db.Fetch(player).name, "(@Odrop)")
					}
					enter_room(descr, player, db.Fetch(dest).location, exit)
				} else {
					notify(player, "That player does not wish to be disturbed.")
				}
			}
			break
		case TYPE_PROGRAM:
			if tmpfr := interp(descr, player, db.Fetch(player).location, dest, exit, FOREGROUND, STD_REGUID, 0); tmpfr != nil {
				interp_loop(player, dest, tmpfr, false)
			}
			return
		}
	}
	if sobjact {
		send_home(descr, db.Fetch(exit).location, 0)
	}
	if !succ && pflag {
		notify(player, "Done.")
	}
}

func do_move(descr int, player dbref, direction string, lev int) {
	if tp_allow_home && direction == "home" {
		/* send him home */
		/* but steal all his possessions */
		if loc := db.Fetch(player).location; loc != NOTHING {
			notify_except(db.Fetch(loc).contents, player, fmt.Sprintf("%s goes home.", db.Fetch(player).name), player)
		}
		/* give the player the messages */
		notify(player, "There's no place like home...")
		notify(player, "There's no place like home...")
		notify(player, "There's no place like home...")
		notify(player, "You wake up back home, without your possessions.")
		send_home(descr, player, 1)
	} else {
		/* find the exit */
		md := NewMatchCheckKeys(descr, player, direction, TYPE_EXIT)
		md.level = lev
		md.MatchAllExits()
		switch exit := md.MatchResult(); exit {
		case NOTHING:
			notify(player, "You can't go that way.")
		case AMBIGUOUS:
			notify(player, "I don't know which way you mean!")
		default:
			/* we got one */
			/* check to see if we got through */
			ts_useobject(exit)
			loc := db.Fetch(player).location
			if can_doit(descr, player, exit, "You can't go that way.") {
				trigger(descr, player, exit, 1)
			}
		}
	}
}

func do_leave(descr int, player dbref) {
	loc := db.Fetch(player).location
	dest := db.Fetch(loc).location
	switch {
	case loc == NOTHING, Typeof(loc) == TYPE_ROOM:
		notify(player, "You can't go that way.")
	case db.Fetch(loc).flags & VEHICLE == 0:
		notify(player, "You can only exit vehicles.")
	case Typeof(dest) != TYPE_ROOM && Typeof(dest) != TYPE_THING:
		notify(player, "You can't exit a vehicle inside of a player.")
	case parent_loop_check(player, dest):
		notify(player, "You can't go that way.")
	default:
		notify(player, "You exit the vehicle.");
		enter_room(descr, player, dest, loc);
	}
}

func do_get(descr int, player dbref, what, obj string) {
	dbref thing, cont;
	int cando;

	md := NewMatchCheckKeys(descr, player, what, TYPE_THING)
	md.MatchNeighbor()
	md.MatchPossession()
	if Wizard(db.Fetch(player).owner) {
		md.MatchAbsolute();	/* the wizard has long fingers */
	}

	if thing = md.NoisyMatchResult(); thing != NOTHING {
		cont = thing
		if (obj && *obj) {
			md := NewMatchCheckKeys(descr, player, obj, TYPE_THING)
			md.RMatch(cont)
			if Wizard(db.Fetch(player).owner) {
				md.MatchAbsolute();	/* the wizard has long fingers */
			}
			if thing = md.NoisyMatchResult(); thing == NOTHING {
				return
			}
			if (Typeof(cont) == TYPE_PLAYER) {
				notify(player, "You can't steal things from players.");
				return;
			}
			if (!test_lock_false_default(descr, player, cont, "_/clk")) {
				notify(player, "You can't open that container.");
				return;
			}
		}
		if Typeof(player) != TYPE_PLAYER {
			if Typeof(db.Fetch(thing).location) != TYPE_ROOM {
				if db.Fetch(player).owner != db.Fetch(thing).owner {
					notify(player, "Zombies aren't allowed to be thieves!");
					return;
				}
			}
		}
		if db.Fetch(thing).location == player {
			notify(player, "You already have that!")
			return
		}
		if (Typeof(cont) == TYPE_PLAYER) {
			notify(player, "You can't steal stuff from players.");
			return;
		}
		if (parent_loop_check(thing, player)) {
			notify(player, "You can't pick yourself up by your bootstraps!");
			return;
		}
		switch (Typeof(thing)) {
		case TYPE_THING:
			ts_useobject(thing);
		case TYPE_PROGRAM:
			if (obj && *obj) {
				cando = could_doit(descr, player, thing);
				if (!cando)
					notify(player, "You can't get that.");
			} else {
				cando = can_doit(descr, player, thing, "You can't pick that up.");
			}
			if (cando) {
				if (tp_thing_movement && (Typeof(thing) == TYPE_THING)) {
					enter_room(descr, thing, player, db.Fetch(thing).location)
				} else {
					moveto(thing, player);
				}
				notify(player, "Taken.");
			}
			break;
		default:
			notify(player, "You can't take that!");
			break;
		}
	}
}

func do_drop(descr int, player dbref, name, obj string) {
	var cont, thing dbref
	char buf[BUFFER_LEN];

	if loc := db.Fetch(player).location; loc != NOTHING {
		md := NewMatch(descr, player, name, NOTYPE)
		md.MatchPossession()
		if thing = md.NoisyMatchResult(); thing == NOTHING || thing == AMBIGUOUS {
			return
		}
		cont = loc;
		if obj != "" {
			md := NewMatch(descr, player, obj, NOTYPE)
			md.MatchPossession()
			md.MatchNeighbor()
			if Wizard(db.Fetch(player).owner) {
				md.MatchAbsolute()	/* the wizard has long fingers */
			}
			if cont = md.NoisyMatchResult(); cont == NOTHING || thing == AMBIGUOUS {
				return
			}
		}
		switch Typeof(thing) {
		case TYPE_THING:
			ts_useobject(thing);
		case TYPE_PROGRAM:
			switch {
			case db.Fetch(thing).location != player:
				/* Shouldn't ever happen. */
				notify(player, "You can't drop that.")
			case Typeof(cont) != TYPE_ROOM && Typeof(cont) != TYPE_PLAYER && Typeof(cont) != TYPE_THING:
				notify(player, "You can't put anything in that.")
			case Typeof(cont) != TYPE_ROOM && !test_lock_false_default(descr, player, cont, "_/clk"):
				notify(player, "You don't have permission to put something in that.")
			case parent_loop_check(thing, cont):
				notify(player, "You can't put something inside of itself.")
			default:
				if Typeof(cont) == TYPE_ROOM && db.Fetch(thing).flags & STICKY != 0 && Typeof(thing) == TYPE_THING {
					send_home(descr, thing, 0);
				} else {
					immediate_dropto := TYPEOF(cont) == TYPE_ROOM && db.Fetch(cont).sp != NOTHING && db.Fetch(cont).flags & STICKY == 0
					if tp_thing_movement && TYPEOF(thing) == TYPE_THING {
						enter_room(descr, thing, immediate_dropto ? db.Fetch(cont).sp.(dbref) : cont, player)
					} else {
						moveto(thing, immediate_dropto ? db.Fetch(cont).sp.(dbref) : cont)
					}
				}
				switch {
				case TYPEOF(cont) == TYPE_THING:
					notify(player, "Put away.")
				case TYPEOF(cont) == TYPE_PLAYER:
					notify_fmt(cont, "%s hands you %s", db.Fetch(player).name, db.Fetch(thing).name)
					notify_fmt(player, "You hand %s to %s", db.Fetch(thing).name, db.Fetch(cont).name)
				default:
					if get_property_class(thing, MESGPROP_DROP) {
						exec_or_notify_prop(descr, player, thing, MESGPROP_DROP, "(@Drop)");
					} else {
						notify(player, "Dropped.")
					}
					if get_property_class(loc, MESGPROP_DROP) {
						exec_or_notify_prop(descr, player, loc, MESGPROP_DROP, "(@Drop)")
					}
					if get_property_class(thing, MESGPROP_ODROP) {
						parse_oprop(descr, player, loc, thing, MESGPROP_ODROP, db.Fetch(player).name, "(@Odrop)")
					} else {
						buf = fmt.Sprintf("%s drops %s.", db.Fetch(player).name, db.Fetch(thing).name)
						notify_except(db.Fetch(loc).contents, player, buf, player)
					}
					if get_property_class(loc, MESGPROP_ODROP) {
						parse_oprop(descr, player, loc, loc, MESGPROP_ODROP, db.Fetch(thing).name, "(@Odrop)")
					}
				}
			}
		default:
			notify(player, "You can't drop that.")
		}
	}
}

func do_recycle(descr int, player dbref, name string) {
	var buf [BUFFER_LEN]byte

	NoGuest("@recycle", player, func() {
		md := NewMatch(descr, player, name, TYPE_THING).
			MatchAllExits().
			MatchNeighbor().
			MatchPossession().
			MatchRegistered().
			MatchHere().
			MatchAbsolute()
		if thing := md.NoisyMatchResult(); thing != NOTHING {
			switch {
			case player != GOD && db.Fetch(thing).owner == GOD:
				notify(player, "Only God may reclaim God's property.")
			case !controls(player, thing):
				notify(player, "Permission denied. (You don't control what you want to recycle)")
			default:
				switch Typeof(thing) {
				case TYPE_ROOM:
					switch {
					case db.Fetch(thing).owner != db.Fetch(player).owner:
						notify(player, "Permission denied. (You don't control the room you want to recycle)")
						return
					case thing == tp_player_start:
						notify(player, "That is the player start room, and may not be recycled.")
						return
					case thing == GLOBAL_ENVIRONMENT:
						notify(player, "If you want to do that, why don't you just delete the database instead?  Room #0 contains everything, and is needed for database sanity.");
						return
					}
				case TYPE_THING:
					switch {
					case db.Fetch(thing).owner != db.Fetch(player).owner:
						notify(player, "Permission denied. (You can't recycle a thing you don't control)")
						return
					case thing == player:
						/* player may be a zombie or puppet */
						buf = fmt.Sprintf("%.512s's owner commands it to kill itself.  It blinks a few times in shock, and says, \"But.. but.. WHY?\"  It suddenly clutches it's heart, grimacing with pain..  Staggers a few steps before falling to it's knees, then plops down on it's face.  *thud*  It kicks its legs a few times, with weakening force, as it suffers a seizure.  It's color slowly starts changing to purple, before it explodes with a fatal *POOF*!", db.Fetch(thing).name)
						notify_except(db.Fetch(db.Fetch(thing).location).contents, thing, buf, player)
						notify(db.Fetch(player).owner, buf)
						notify(db.Fetch(player).owner, "Now don't you feel guilty?")
					}
				case TYPE_EXIT:
					switch {
					case db.Fetch(thing).owner != db.Fetch(player).owner:
						notify(player, "Permission denied. (You may not recycle an exit you don't own)")
						return
					case !unset_source(player, db.Fetch(player).location, thing):
						notify(player, "You can't do that to an exit in another room.")
						return
					}
				case TYPE_PLAYER:
					notify(player, "You can't recycle a player!")
					return
				case TYPE_PROGRAM:
					if db.Fetch(thing).owner != db.Fetch(player).owner {
						notify(player, "Permission denied. (You can't recycle a program you don't own)")
						return
					}
					SetMLevel(thing, NON_MUCKER)
					if db.Fetch(thing).sp.(program_specific) != nil && db.Fetch(thing).sp.(program_specific).instances > 0 {
						dequeue_prog(thing, 0)
					}
				}
				notify(player, fmt.Sprintf("Thank you for recycling %.512s (#%d).", db.Fetch(thing).name, thing))
				recycle(descr, player, thing)
			}
		}
	})
}

var depth int = 0

func recycle(descr int, player, thing dbref) {
	dbref first
	dbref rest
	char buf[2048]
	int looplimit

	depth++
	if force_level {
		if thing == force_prog {
			log_status("SANITYCHECK: Was about to recycle FORCEing object #%d!", thing)
			notify(player, "ERROR: Cannot recycle an object FORCEing you!")
			return
		}

		var i int
		if db.Fetch(thing).sp.(program_specific) != nil {
			i = db.Fetch(thing).sp.(program_specific).instances
		}
		if Typeof(thing) == TYPE_PROGRAM && i != 0 {
			log_status("SANITYCHECK: Trying to recycle a running program (#%d) from FORCE!", thing)
			notify(player, "ERROR: Cannot recycle a running program from FORCE.")
			return
		}
	}
	/* dequeue any MUF or MPI events for the given object */
	dequeue_prog(thing, 0)
	switch thing.(type) {
	case TYPE_ROOM:
		if !Wizard(db.Fetch(thing).owner) {
			add_property(db.Fetch(thing).owner, MESGPROP_VALUE, nil, get_property_value(db.Fetch(thing).owner, MESGPROP_VALUE) + tp_room_cost)
		}
		db.Fetch(db.Fetch(thing).owner).flags |= OBJECT_CHANGED
		for first := db.Fetch(thing).exits; first != NOTHING; first = rest {
			rest = db.Fetch(first).next
			if db.Fetch(first).location == NOTHING || db.Fetch(first).location == thing {
				recycle(descr, player, first)
			}
		}
		notify_except(db.Fetch(thing).contents, NOTHING, "You feel a wrenching sensation...", player)
	case TYPE_THING:
		if !Wizard(db.Fetch(thing).owner) {
			add_property(db.Fetch(thing).owner, MESGPROP_VALUE, nil, get_property_value(db.Fetch(thing).owner, MESGPROP_VALUE) + get_property_value(thing, MESGPROP_VALUE))
		}
		db.Fetch(db.Fetch(thing).owner).flags |= OBJECT_CHANGED
		for first := db.Fetch(thing).exits; first != NOTHING; first = rest {
			rest = db.Fetch(first).next
			if db.Fetch(first).location == NOTHING || db.Fetch(first).location == thing {
				recycle(descr, player, first)
			}
		}
	case TYPE_EXIT:
		if !Wizard(db.Fetch(thing).owner) {
			add_property(db.Fetch(thing).owner, MESGPROP_VALUE, nil, get_property_value(db.Fetch(thing).owner, MESGPROP_VALUE) + tp_exit_cost)
		}
		if !Wizard(db.Fetch(thing).owner) && len(db.Fetch(thing).sp.exit.dest) != 0 {
			add_property(db.Fetch(thing).owner, MESGPROP_VALUE, nil, get_property_value(db.Fetch(thing).owner, MESGPROP_VALUE) + tp_link_cost)
		}
		db.Fetch(db.Fetch(thing).owner).flags |= OBJECT_CHANGED
	case TYPE_PROGRAM:
		unlink(fmt.Sprintf("muf/%v.m", thing))
	}

	for rest := 0; rest < db_top; rest++ {
		switch TYPEOF(rest) {
		case TYPE_ROOM:
			if db.Fetch(rest).sp == thing {
				db.Fetch(rest).sp = NOTHING
				db.Fetch(rest).flags |= OBJECT_CHANGED
			}
			if db.Fetch(rest).exits == thing {
				db.Fetch(rest).exits = db.Fetch(thing).next
				db.Fetch(rest).flags |= OBJECT_CHANGED
			}
			if db.Fetch(rest).owner == thing {
				db.Fetch(rest).owner = GOD
				db.Fetch(rest).flags |= OBJECT_CHANGED
			}
		case TYPE_THING:
			if db.Fetch(rest).sp.(player_specific).home == thing {
				if db.Fetch(db.Fetch(rest).owner).sp.(player_specific).home == thing {
					db.Fetch(db.Fetch(rest).owner).sp.(player_specific).home = tp_player_start
				}
				loc := db.Fetch(db.Fetch(rest).owner).sp.(player_specific).home
				if parent_loop_check(rest, loc) {
					loc = db.Fetch(rest).owner
					if parent_loop_check(rest, loc) {
						loc = dbref(0)
					}
				}
				db.Fetch(rest).sp.(player_specific).home = loc
				db.Fetch(rest).flags |= OBJECT_CHANGED
			}
			if db.Fetch(rest).exits == thing {
				db.Fetch(rest).exits = db.Fetch(thing).next
				db.Fetch(rest).flags |= OBJECT_CHANGED
			}
			if db.Fetch(rest).owner == thing {
				db.Fetch(rest).owner = GOD
				db.Fetch(rest).flags |= OBJECT_CHANGED
			}
		case TYPE_EXIT:
			var i, j int
			for ; i < len(db.Fetch(rest).sp.exit.dest); i++ {
				if db.Fetch(rest).sp.exit.dest[i] != thing {
					db.Fetch(rest).sp.exit.dest[j] = db.Fetch(rest).sp.exit.dest[i]
					j++
				}
				if j < len(db.Fetch(rest).sp.exit.dest) {
					add_property(db.Fetch(rest).owner, MESGPROP_VALUE, nil, get_property_value(db.Fetch(rest).owner, MESGPROP_VALUE) + tp_link_cost)
					db.Fetch(db.Fetch(rest).owner).flags |= OBJECT_CHANGED
					for x, _ := range db.Fetch(rest).sp.exit.dest[j:] {
						db.Fetch(rest).sp.exit.dest[x] = nil
					}
					db.Fetch(rest).sp.exit.dest = db.Fetch(rest).sp.exit.dest[:j]
					db.Fetch(rest).flags |= OBJECT_CHANGED
				}
			}
			if db.Fetch(rest).owner == thing {
				db.Fetch(rest).owner = GOD
				db.Fetch(rest).flags |= OBJECT_CHANGED
			}
		case TYPE_PLAYER:
			if TYPEOF(thing) == TYPE_PROGRAM && db.Fetch(rest).flags & INTERACTIVE != 0 && db.Fetch(rest).sp.(player_specific).curr_prog == thing {
				if db.Fetch(rest).flags & READMODE != 0 {
					notify(rest, "The program you were running has been recycled.  Aborting program.")
				} else {
					db.Fetch(first).sp.(program_specific).first = nil
					db.Fetch(rest).sp.(player_specific).insert_mode = false
					db.Fetch(thing).flags &= ~INTERNAL
					db.Fetch(rest).flags &= ~INTERACTIVE
					db.Fetch(rest).sp.(player_specific).curr_prog = NOTHING
					notify(rest, "The program you were editing has been recycled.  Exiting Editor.")
				}
			}
			if db.Fetch(rest).sp.(player_specific).home == thing {
				db.Fetch(rest).sp.(player_specific).home = tp_player_start
				db.Fetch(rest).flags |= OBJECT_CHANGED
			}
			if db.Fetch(rest).exits == thing {
				db.Fetch(rest).exits = db.Fetch(thing).next
				db.Fetch(rest).flags |= OBJECT_CHANGED
			}
			if db.Fetch(rest).sp.(player_specific).curr_prog == thing {
				db.Fetch(rest).sp.(player_specific).curr_prog = 0
			}
		case TYPE_PROGRAM:
			if db.Fetch(rest).owner == thing {
				db.Fetch(rest).owner = GOD
				db.Fetch(rest).flags |= OBJECT_CHANGED
			}
		}
		if db.Fetch(rest).contents == thing {
			db.Fetch(rest).contents = db.Fetch(thing).next
			db.Fetch(rest).flags |= OBJECT_CHANGED
		}
		if db.Fetch(rest).next == thing {
			db.Fetch(rest).next = db.Fetch(thing).next
			db.Fetch(rest).flags |= OBJECT_CHANGED
		}
	}

	first = db.Fetch(thing).contents
	for looplimit = db_top; (looplimit > 0 && first != NOTHING); looplimit-- {
		if TYPEOF(first) == TYPE_PLAYER || (TYPEOF(first) == TYPE_THING && (db.Fetch(first).flags & (ZOMBIE | VEHICLE) != 0 || tp_thing_movement)) {
			enter_room(descr, first, HOME, db.Fetch(thing).location)
			/* If the room is set to drag players back, there'll be no reasoning with it.  DRAG the player out. */
			if db.Fetch(first).location == thing {
				notify_fmt(player, "Escaping teleport loop!  Going home.")
				moveto(first, HOME)
			}
		} else {
			moveto(first, HOME)
		}
		first = db.Fetch(thing).contents
	}

	moveto(thing, NOTHING)
	depth--

	db_free_object(thing)
	db_clear_object(thing)
	db.Fetch(thing).flags |= OBJECT_CHANGED
}