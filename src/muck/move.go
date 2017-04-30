package fbmuck

func moveto(what, where ObjectID) {
	var loc ObjectID
	if loc = DB.Fetch(what).Location; loc != NOTHING {
		DB.Fetch(loc).Contents = remove_first(DB.Fetch(loc).Contents, what)
		DB.Fetch(loc).Touch()
	}

	o := DB.Fetch(what)
	switch where {
	case NOTHING:
		o.MoveTo(NOTHING)
		o.Touch()
		return					/* NOTHING doesn't have contents */
	case HOME:
		switch o := o.(type) {
		case Player:
			where = o.Home
		case Object:
			where = o.Home
			if parent_loop_check(what, where) {
				where = DB.FetchPlayer(o.Owner).Home
				if parent_loop_check(what, where) {
					where = ObjectID(tp_player_start)
				}
			}
		case Room:
			where = GLOBAL_ENVIRONMENT
		case Program:
			where = o.Owner
		}
	default:
		if parent_loop_check(what, where) {
			switch o := o.(type) {
			case Player:
				where = o.Home
			case Object:
				where = o.Home
				if parent_loop_check(what, where) {
					where = DB.FetchPlayer(o.Owner).Home
					if parent_loop_check(what, where) {
						where = tp_player_start
					}
				}
			case Room:
				where = GLOBAL_ENVIRONMENT
			case Program:
				where = o.Owner
			}
		}
	}
	dest := DB.Fetch(where)
	o.next = dest.Contents
	dest.Contents = what
	dest.Touch()
	o.MoveTo(where)
	o.Touch()
}

func send_contents(int descr, ObjectID loc, ObjectID dest) {
	first := DB.Fetch(loc).Contents
	DB.Fetch(loc).Contents = NOTHING
	DB.Fetch(loc).Touch()

	/* blast locations of everything in list */
	for rest := first; rest != NOTHING; rest = DB.Fetch(rest).next {
		DB.Fetch(rest).MoveTo(NOTHING)
		DB.Fetch(rest).Touch()
	}

	for first != NOTHING {
		rest := DB.Fetch(first).next
		if Typeof(first) != TYPE_THING && Typeof(first) != TYPE_PROGRAM {
			moveto(first, loc)
		} else {
			where := dest
			if DB.Fetch(first).IsFlagged(STICKY) {
				where = HOME
			}
			if tp_thing_movement && IsThing(first) {
				if parent_loop_check(first, where) {
					enter_room(descr, first, loc, DB.Fetch(first).Location)
				} else {
					enter_room(descr, first, where, DB.Fetch(first).Location)
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
	DB.Fetch(loc).Contents = reverse(DB.Fetch(loc).Contents)
	DB.Fetch(loc).Touch()
}

func maybe_dropto(descr int, loc, dropto ObjectID) {
	if loc != dropto {
		/* check for players */
		for thing := DB.Fetch(loc).Contents; thing != NOTHING; thing = DB.Fetch(thing).next {
			/* Make zombies act like players for dropto processing */
			if IsPlayer(thing) || (IsThing(thing) && DB.Fetch(thing).IsFlagged(ZOMBIE)) {
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
location_loop_check(ObjectID source, ObjectID dest)
{   
  unsigned int level = 0;
  unsigned int place = 0;
  ObjectID pstack[MAX_PARENT_DEPTH+2];

  if (source == dest) {
    return 1;
  }
  pstack[0] = source;
  pstack[1] = dest;

  while (level < MAX_PARENT_DEPTH) {
    dest = DB.Fetch(dest).Location
    if (dest == NOTHING) {
      return 0;
    }
    if (dest == HOME) {        /* We should never get this, either. */
      return 1;
    }
    if (dest == (ObjectID) 0) {   /* Reached the top of the chain. */
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

func parent_loop_check(source, dest ObjectID) bool {   
	if dest == HOME {
		switch source := DB.Fetch(source).(type) {
		case Player:
			dest = source.Home
		case Object:
			dest = source.Home
		case Room:
			dest = GLOBAL_ENVIRONMENT
		case Program:
			dest = source.Owner
		default:
			return true
		}
	}

	switch {
	case location_loop_check(source, dest):
		return true
	case source == dest:
		return true
	}

	pstack := []ObjectID{ source, dest }
	for len(pstack) < MAX_PARENT_DEPTH {
		switch dest = getparent(dest); {
		case dest == NOTHING:
			return false
		case dest == HOME:
			/* We should never get this, either. */
			return true
		case dest == 0:
			/* Reached the top of the chain. */
			return false
		default:
			/* Check to see if we've found this item before.. */
			for place := 0; place < len(pstack) + 2; place++ {
				if pstack[place] == dest {
					return true
				}
			}
			pstack = append(pstack, dest)
		}
	}
	return true
}

static int donelook = 0;
func enter_room(descr int, player, loc, exit ObjectID) {
	var old, dropto ObjectID

	p := DB.Fetch(player)
	if loc == HOME {
		loc = p.Home
	}

	/* get old location */
	old = p.Location

	if parent_loop_check(player, loc) {
		switch p := p.(type) {
		case Player:
			loc = p.Home
		case Object:
			loc = p.Home
			if parent_loop_check(player, loc) {
				loc = DB.FetchPlayer(p.Owner).Home
				if parent_loop_check(player, loc) {
					loc = tp_player_start
				}
			}
		case Room:
			loc = GLOBAL_ENVIRONMENT
		case Program:
			loc = p.Owner
		}
	}

	/* check for self-loop */
	/* self-loops don't do move or other player notification */
	/* but you still get autolook and penny check */
	if loc != old {
		moveto(player, loc)
		if old != NOTHING {
			propqueue(descr, player, old, exit, player, NOTHING, "_depart", "Depart", 1, 1);
			envpropqueue(descr, player, old, exit, old, NOTHING, "_depart", "Depart", 1, 1);

			propqueue(descr, player, old, exit, player, NOTHING, "_odepart", "Odepart", 1, 0);
			envpropqueue(descr, player, old, exit, old, NOTHING, "_odepart", "Odepart", 1, 0);

			/* notify others unless DARK */
			if !Dark(old) && !Dark(player) && (!IsThing(player) || (IsThing(player) && p.IsFlagged(ZOMBIE, VEHICLE)) && (!IsExit(exit) || !Dark(exit)) {
#if !defined(QUIET_MOVES)
				notify_except(DB.Fetch(old).Contents, player, fmt.Sprintf("%s has left.", p.name), player)
#endif
			}
		}

		/* if old location has STICKY dropto, send stuff through it */
		if old != NOTHING && IsRoom(old) {
			if dropto = DB.Fetch(old).(ObjectID)); dropto != NOTHING && DB.Fetch(old).IsFlagged(STICKY) {
				maybe_dropto(descr, old, dropto)
			}
		}

		/* tell other folks in new location if not DARK */
		if !Dark(loc) && !Dark(player) && (!IsThing(player) || (IsThing(player) && DB.Fetch(player).IsFlagged(ZOMBIE, VEHICLE)) && (!IsExit(exit) || !Dark(exit)) {
#if !defined(QUIET_MOVES)
			notify_except(DB.Fetch(loc).Contents, player, fmt.Sprintf("%s has arrived.", p.name), player)
#endif
		}
	}
	/* autolook */
	if !IsThing(player) || (IsThing(player) && p.IsFlagged(ZOMBIE | VEHICLE)) {
		if donelook < 8 {
			donelook++
			if can_move(descr, player, tp_autolook_cmd, 1) {
				do_move(descr, player, tp_autolook_cmd, 1)
			} else {
				do_look_around(descr, player)
			}
			donelook--
		} else {
			notify(player, "Look aborted because of look action loop.")
		}
	}

	if tp_penny_rate != 0 {
		/* check for pennies */
		if !controls(player, loc) && get_property_value(p.Owner, MESGPROP_VALUE) <= tp_max_pennies && rand.Int() % tp_penny_rate == 0 {
			notify_fmt(player, "You found one %s!", tp_penny)
			add_property(p.Owner, MESGPROP_VALUE, nil, get_property_value(p.Owner, MESGPROP_VALUE) + 1)
			DB.Fetch(p.Owner).Touch()
		}
	}

	if loc != old {
		envpropqueue(descr, player, loc, exit, player, NOTHING, "_arrive", "Arrive", 1, 1)
		envpropqueue(descr, player, loc, exit, player, NOTHING, "_oarrive", "Oarrive", 1, 0)
	}
}

func send_home(descr int, thing ObjectID, puppethome int) {
	switch o := DB.Fetch(thing).(type) {
	case Player:
		/* send his possessions home first! */
		/* that way he sees them when he arrives */
		send_contents(descr, thing, HOME)
		enter_room(descr, thing, o.Home, o.Location)
	case Object:
		if puppethome {
			send_contents(descr, thing, HOME)
		}
		if tp_thing_movement || o.IsFlagged(ZOMBIE, LISTENER) {
			enter_room(descr, thing, o.Home, o.Location)
		} else {
			moveto(thing, HOME)		/* home */
		}
	case Program:
		moveto(thing, o.Owner)
	}
}

func can_move(descr int, player ObjectID , direction string, lev int) (r bool) {
	if tp_allow_home && direction == "home" {
		r = true
	} else {
		/* otherwise match on exits */
		md := NewMatch(descr, player, direction, IsExit)
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

func trigger(int descr, ObjectID player, ObjectID exit, int pflag) {
	int sobjact;				/* sticky object action flag, sends home source obj */
	sobjact = 0;

	var succ bool
	e := DB.Fetch(exit).(Exit)
	p := DB.FetchPlayer(player)
	for i, dest := range e.Destinations {
		if dest == HOME {
			dest = p.Home

			/* fix #1112946 temporarily -- premchai21 */
			if IsThing(dest) {
				notify(player, "That would be an undefined operation.")
				continue
			}
		}
		switch dest := DB.Fetch(dest).(type) {
		case Room:
			switch {
			case !pflag:
			case parent_loop_check(player, dest):
				notify(player, "That would cause a paradox.")
				break
			case !Wizard(p.Owner) && IsThing(player) && dest.IsFlagged(ZOMBIE):
				notify(player, "You can't go that way.")
				break
			case p.IsFlagged(VEHICLE) && (dest.Bitset | e.Bitset).IsFlagged(VEHICLE):
				notify(player, "You can't go that way.")
				break
			default:
				if get_property_class(exit, MESGPROP_DROP) {
					exec_or_notify_prop(descr, player, exit, MESGPROP_DROP, "(@Drop)")
				}
				if get_property_class(exit, MESGPROP_ODROP) && !Dark(player) {
					parse_oprop(descr, player, dest, exit, MESGPROP_ODROP, p.name, "(@Odrop)")
				}
				enter_room(descr, player, dest, exit)
				succ = true
			}
		case Object:
			if dest == e.Location && dest.IsFlagged(VEHICLE) {
				switch {
				case !pflag:
				case parent_loop_check(player, dest):
					notify(player, "That would cause a paradox.")
					break
				default:
					if get_property_class(exit, MESGPROP_DROP) {
						exec_or_notify_prop(descr, player, exit, MESGPROP_DROP, "(@Drop)")
					}
					if get_property_class(exit, MESGPROP_ODROP) && !Dark(player) {
						parse_oprop(descr, player, dest, exit, MESGPROP_ODROP, p.name, "(@Odrop)")
					}
					enter_room(descr, player, dest, exit)
					succ = true
				}
			} else {
				if IsThing(e.Location) {
					switch {
					case parent_loop_check(dest, DB.Fetch(e.Location).Location):
						notify(player, "That would cause a paradox.")
						break
					default:
						if tp_thing_movement {
							enter_room(descr, dest, DB.Fetch(e.Location).Location, exit);
						} else {
							moveto(dest, DB.Fetch(e.Location).Location)
						}
						if !e.IsFlagged(STICKY) {
							/* send home source object */
							sobjact = 1
						}
					}
				} else {
					switch {
					case parent_loop_check(dest, e.Location):
						notify(player, "That would cause a paradox.")
						break
					default:
						if tp_thing_movement {
							enter_room(descr, dest, e.Location, exit)
						} else {
							moveto(dest, e.Location)
						}
					}
				}
				if get_property_class(exit, MESGPROP_SUCC) {
					succ = true
				}
			}
			break
		case Exit:		/* It's a meta-link(tm)! */
			ts_useobject(dest)
			trigger(descr, player, e.Destinations[i], 0)
			if get_property_class(exit, MESGPROP_SUCC) {
				succ = true
			}
			break
		case Player:
			if pflag && dest.Location != NOTHING {
				if parent_loop_check(player, dest) {
					notify(player, "That would cause a paradox.")
					break
				}
				succ = true
				if dest.IsFlagged(JUMP_OK) {
					if get_property_class(exit, MESGPROP_DROP) {
						exec_or_notify_prop(descr, player, exit, MESGPROP_DROP, "(@Drop)")
					}
					if get_property_class(exit, MESGPROP_ODROP) && !Dark(player) {
						parse_oprop(descr, player, dest.Location, exit, MESGPROP_ODROP, p.name, "(@Odrop)")
					}
					enter_room(descr, player, dest.Location, exit)
				} else {
					notify(player, "That player does not wish to be disturbed.")
				}
			}
			break
		case Program:
			if tmpfr := interp(descr, player, p.Location, dest, exit, FOREGROUND, STD_REGUID, 0); tmpfr != nil {
				interp_loop(player, dest, tmpfr, false)
			}
			return
		}
	}
	if sobjact {
		send_home(descr, e.Location, 0)
	}
	if !succ && pflag {
		notify(player, "Done.")
	}
}

func do_move(descr int, player ObjectID, direction string, lev int) {
	if tp_allow_home && direction == "home" {
		/* send him home */
		/* but steal all his possessions */
		if loc := DB.Fetch(player).Location; loc != NOTHING {
			notify_except(DB.Fetch(loc).Contents, player, fmt.Sprintf("%s goes home.", DB.Fetch(player).name), player)
		}
		/* give the player the messages */
		notify(player, "There's no place like home...")
		notify(player, "There's no place like home...")
		notify(player, "There's no place like home...")
		notify(player, "You wake up back home, without your possessions.")
		send_home(descr, player, 1)
	} else {
		/* find the exit */
		md := NewMatchCheckKeys(descr, player, direction, IsExit)
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
			loc := DB.Fetch(player).Location
			if can_doit(descr, player, exit, "You can't go that way.") {
				trigger(descr, player, exit, 1)
			}
		}
	}
}

func do_leave(descr int, player ObjectID) {
	loc := DB.Fetch(player).Location
	dest := DB.Fetch(loc).Location
	switch {
	case loc == NOTHING, IsRoom(loc):
		notify(player, "You can't go that way.")
	case !DB.Fetch(loc).IsFlagged(VEHICLE):
		notify(player, "You can only exit vehicles.")
	case !IsRoom(dest) && !IsThing(dest):
		notify(player, "You can't exit a vehicle inside of a player.")
	case parent_loop_check(player, dest):
		notify(player, "You can't go that way.")
	default:
		notify(player, "You exit the vehicle.");
		enter_room(descr, player, dest, loc);
	}
}

func do_get(descr int, player ObjectID, what, obj string) {
	ObjectID thing, cont;
	int cando;

	md := NewMatchCheckKeys(descr, player, what, IsThing)
	md.MatchNeighbor()
	md.MatchPossession()
	if Wizard(DB.Fetch(player).Owner) {
		md.MatchAbsolute();	/* the wizard has long fingers */
	}

	if thing = md.NoisyMatchResult(); thing != NOTHING {
		cont = thing
		if (obj && *obj) {
			md := NewMatchCheckKeys(descr, player, obj, IsThing)
			md.RMatch(cont)
			if Wizard(DB.Fetch(player).Owner) {
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
			if Typeof(DB.Fetch(thing).Location) != TYPE_ROOM {
				if DB.Fetch(player).Owner != DB.Fetch(thing).Owner {
					notify(player, "Zombies aren't allowed to be thieves!");
					return;
				}
			}
		}
		if DB.Fetch(thing).Location == player {
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
					enter_room(descr, thing, player, DB.Fetch(thing).Location)
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

func do_drop(descr int, player ObjectID, name, obj string) {
	var cont, thing ObjectID
	char buf[BUFFER_LEN];

	if loc := DB.Fetch(player).Location; loc != NOTHING {
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
			if Wizard(DB.Fetch(player).Owner) {
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
			case DB.Fetch(thing).Location != player:
				/* Shouldn't ever happen. */
				notify(player, "You can't drop that.")
			case Typeof(cont) != TYPE_ROOM && Typeof(cont) != TYPE_PLAYER && Typeof(cont) != TYPE_THING:
				notify(player, "You can't put anything in that.")
			case Typeof(cont) != TYPE_ROOM && !test_lock_false_default(descr, player, cont, "_/clk"):
				notify(player, "You don't have permission to put something in that.")
			case parent_loop_check(thing, cont):
				notify(player, "You can't put something inside of itself.")
			default:
				if IsRoom(cont) && DB.Fetch(thing).IsFlagged(STICKY) && IsThing(thing) {
					send_home(descr, thing, 0);
				} else {
					immediate_dropto := IsRoom(cont) && DB.Fetch(cont).sp != NOTHING && !DB.Fetch(cont).IsFlagged(STICKY)
					if tp_thing_movement && IsThing(thing) {
						if immediate_dropto {
							enter_room(descr, thing, DB.Fetch(cont).(ObjectID), player)
						} else {
							enter_room(descr, thing, cont, player)
						}
					} else {
						if immediate_dropto {
							moveto(thing, DB.Fetch(cont).(ObjectID))
						} else {
							moveto(thing, cont)
						}
					}
				}
				switch {
				case TYPEOF(cont) == TYPE_THING:
					notify(player, "Put away.")
				case TYPEOF(cont) == TYPE_PLAYER:
					notify_fmt(cont, "%s hands you %s", DB.Fetch(player).name, DB.Fetch(thing).name)
					notify_fmt(player, "You hand %s to %s", DB.Fetch(thing).name, DB.Fetch(cont).name)
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
						parse_oprop(descr, player, loc, thing, MESGPROP_ODROP, DB.Fetch(player).name, "(@Odrop)")
					} else {
						buf = fmt.Sprintf("%s drops %s.", DB.Fetch(player).name, DB.Fetch(thing).name)
						notify_except(DB.Fetch(loc).Contents, player, buf, player)
					}
					if get_property_class(loc, MESGPROP_ODROP) {
						parse_oprop(descr, player, loc, loc, MESGPROP_ODROP, DB.Fetch(thing).name, "(@Odrop)")
					}
				}
			}
		default:
			notify(player, "You can't drop that.")
		}
	}
}

func do_recycle(descr int, player ObjectID, name string) {
	var buf [BUFFER_LEN]byte

	NoGuest("@recycle", player, func() {
		md := NewMatch(descr, player, name, IsThing).
			MatchAllExits().
			MatchNeighbor().
			MatchPossession().
			MatchRegistered().
			MatchHere().
			MatchAbsolute()
		if thing := md.NoisyMatchResult(); thing != NOTHING {
			switch {
			case player != GOD && DB.Fetch(thing).Owner == GOD:
				notify(player, "Only God may reclaim God's property.")
			case !controls(player, thing):
				notify(player, "Permission denied. (You don't control what you want to recycle)")
			default:
				switch Typeof(thing) {
				case TYPE_ROOM:
					switch {
					case DB.Fetch(thing).Owner != DB.Fetch(player).Owner:
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
					case DB.Fetch(thing).Owner != DB.Fetch(player).Owner:
						notify(player, "Permission denied. (You can't recycle a thing you don't control)")
						return
					case thing == player:
						/* player may be a zombie or puppet */
						buf = fmt.Sprintf("%.512s's owner commands it to kill itself.  It blinks a few times in shock, and says, \"But.. but.. WHY?\"  It suddenly clutches it's heart, grimacing with pain..  Staggers a few steps before falling to it's knees, then plops down on it's face.  *thud*  It kicks its legs a few times, with weakening force, as it suffers a seizure.  It's color slowly starts changing to purple, before it explodes with a fatal *POOF*!", DB.Fetch(thing).name)
						notify_except(DB.Fetch(DB.Fetch(thing).Location).Contents, thing, buf, player)
						notify(DB.Fetch(player).Owner, buf)
						notify(DB.Fetch(player).Owner, "Now don't you feel guilty?")
					}
				case TYPE_EXIT:
					switch {
					case DB.Fetch(thing).Owner != DB.Fetch(player).Owner:
						notify(player, "Permission denied. (You may not recycle an exit you don't own)")
						return
					case !unset_source(player, DB.Fetch(player).Location, thing):
						notify(player, "You can't do that to an exit in another room.")
						return
					}
				case TYPE_PLAYER:
					notify(player, "You can't recycle a player!")
					return
				case TYPE_PROGRAM:
					if DB.Fetch(thing).Owner != DB.Fetch(player).Owner {
						notify(player, "Permission denied. (You can't recycle a program you don't own)")
						return
					}
					SetMLevel(thing, NON_MUCKER)
					if DB.Fetch(thing).(Program) != nil && DB.Fetch(thing).(Program).instances > 0 {
						dequeue_prog(thing, 0)
					}
				}
				notify(player, fmt.Sprintf("Thank you for recycling %.512s (#%d).", DB.Fetch(thing).name, thing))
				recycle(descr, player, thing)
			}
		}
	})
}

var depth int = 0

func recycle(descr int, player, thing ObjectID) {
	char buf[2048]
	int looplimit

	depth++
	var rest ObjectID
	if force_level {
		if thing == force_prog {
			log_status("SANITYCHECK: Was about to recycle FORCEing object #%d!", thing)
			notify(player, "ERROR: Cannot recycle an object FORCEing you!")
			return
		}

		switch p := DB.Fetch(thing).(type) {
		case Program:
			if i := p.instances; i != 0 {
				log_status("SANITYCHECK: Trying to recycle a running program (#%d) from FORCE!", thing)
				notify(player, "ERROR: Cannot recycle a running program from FORCE.")
				return
			}
		}
	}
	/* dequeue any MUF or MPI events for the given object */
	dequeue_prog(thing, 0)
	o := DB.Fetch(thing)
	switch o := o.(type) {
	case Room:
		if !Wizard(o.Owner) {
			add_property(o.Owner, MESGPROP_VALUE, nil, get_property_value(o.Owner, MESGPROP_VALUE) + tp_room_cost)
		}
		DB.Fetch(o.Owner).Touch()
		for first := o.Exits; first != NOTHING; first = rest {
			p := DB.Fetch(first)
			rest = p.next
			switch p.Location {
			case NOTHING, thing:
				recycle(descr, player, first)
			}
		}
		notify_except(o.Contents, NOTHING, "You feel a wrenching sensation...", player)
	case Object:
		if !Wizard(o.Owner) {
			add_property(o.Owner, MESGPROP_VALUE, nil, get_property_value(o.Owner, MESGPROP_VALUE) + get_property_value(thing, MESGPROP_VALUE))
		}
		DB.Fetch(o.Owner).Touch()
		for first := o.Exits; first != NOTHING; first = rest {
			p := DB.Fetch(first)
			rest = p.next
			switch p.Location {
			case NOTHING, thing:
				recycle(descr, player, first)
			}
		}
	case Exit:
		if !Wizard(o.Owner) {
			add_property(o.Owner, MESGPROP_VALUE, nil, get_property_value(o.Owner, MESGPROP_VALUE) + tp_exit_cost)
		}
		if !Wizard(o.Owner) && len(o.Destinations) != 0 {
			add_property(o.Owner, MESGPROP_VALUE, nil, get_property_value(o.Owner, MESGPROP_VALUE) + tp_link_cost)
		}
		DB.Fetch(o.Owner).Touch()
	case Program:
		os.Remove(fmt.Sprintf("muf/%v.m", thing))
	}

	t := DB.Fetch(thing)
	first := t.Contents
	EachObject(func(obj ObjectID, o *Object) {
		switch rest := DB.Fetch(rest).(type) {
		case Room:
			if o.sp == thing {
				o.sp = NOTHING
				o.Touch()
			}
			if o.Exits == thing {
				o.Exits = t.next
				o.Touch()
			}
			if o.Owner == thing {
				o.GiveTo(GOD)
				o.Touch()
			}
		case Object:
			if o.Home == thing {
				if p := DB.FetchPlayer(o.Owner); p.Home == thing {
					p.LiveAt(tp_player_start)
				}
				loc := DB.FetchPlayer(o.Owner).Home
				if parent_loop_check(rest, loc) {
					loc = o.Owner
					if parent_loop_check(rest, loc) {
						loc = ObjectID(0)
					}
				}
				o.LiveAt(loc)
				o.Touch()
			}
			if o.Exits == thing {
				o.Exits = t.next
				o.Touch()
			}
			if o.Owner == thing {
				o.GiveTo(GOD)
				o.Touch()
			}
		case Exit:
			var i, j int
			for _, v := range o.Destinations {
				if v != thing {
					o.Destinations[j] = v
					j++
				}
				if j < len(o.Destinations) {
					add_property(o.Owner, MESGPROP_VALUE, nil, get_property_value(o.Owner, MESGPROP_VALUE) + tp_link_cost)
					DB.Fetch(o.Owner).Touch()
					for x, _ := range o.Destinations[j:] {
						o.Destinations[x] = nil
					}
					o.Destinations = o.Destinations[:j]
					o.Touch()
				}
			}
			if o.Owner == thing {
				o.GiveTo(GOD)
				o.Touch()
			}
		case Player:
			if IsProgram(thing) && o.IsFlagged(INTERACTIVE) && o.(Player).curr_prog == thing {
				if o.IsFlagged(READMODE) {
					notify(rest, "The program you were running has been recycled.  Aborting program.")
				} else {
					DB.Fetch(first).(Program).first = nil
					o.(Player).insert_mode = false
					DB.Fetch(thing).ClearFlags(INTERNAL)
					o.ClearFlags(INTERACTIVE)
					o.(Player).curr_prog = NOTHING
					notify(rest, "The program you were editing has been recycled.  Exiting Editor.")
				}
			}
			if o.Home == thing {
				o.LiveAt(tp_player_start)
				o.Touch()
			}
			if o.Exits == thing {
				o.Exits = t.next
				o.Touch()
			}
			if o.curr_prog == thing {
				o.curr_prog = 0
			}
		case IsProgram(rest):
			if o.Owner == thing {
				o.GiveTo(GOD)
				o.Touch()
			}
		}
		if o.Contents == thing {
			o.Contents = t.next
			o.Touch()
		}
		if o.next == thing {
			o.next = t.next
			o.Touch()
		}
	})

	EachObjectInReverse(func(obj ObjectID, o *Object) bool {
		if IsPlayer(first) || (IsThing(first) && (DB.Fetch(first).IsFlagged(ZOMBIE, VEHICLE) || tp_thing_movement)) {
			enter_room(descr, first, HOME, DB.Fetch(thing).Location)
			/* If the room is set to drag players back, there'll be no reasoning with it.  DRAG the player out. */
			if DB.Fetch(first).Location == thing {
				notify_fmt(player, "Escaping teleport loop!  Going home.")
				moveto(first, HOME)
			}
		} else {
			moveto(first, HOME)
		}
		first = t.Contents
		return first == NOTHING
	})

	moveto(thing, NOTHING)
	depth--

	DB.Store(thing, nil)
	db_clear_object(thing)
	t.Touch()
}