package fbmuck

/* Commands which involve speaking */

func do_say(ObjectID player, const char *message) {
	if loc := DB.Fetch(player).Location; loc != NOTHING {
		notify(player, fmt.Sprintf("You say, \"%s\"", message))
		notify_except(DB.Fetch(loc).Contents, player, fmt.Sprintf("%s says, \"%s\"", DB.Fetch(player).name, message), player)
	}
}

func do_whisper(int descr, ObjectID player, const char *arg1, const char *arg2) {
	md := NewMatch(descr, player, arg1, IsPlayer).
		MatchNeighbor().
		MatchMe()
	if Wizard(player) && Typeof(player) == TYPE_PLAYER {
		md.MatchAbsolute().MatchPlayer()
	}
	switch who = md.MatchResult(); who {
	case NOTHING:
		notify(player, "Whisper to whom?");
	case AMBIGUOUS:
		notify(player, "I don't know who you mean!");
	default:
		if notify_from(player, who, fmt.Sprintf("%s whispers, \"%s\"", DB.Fetch(player).name, arg2)) {
			notify(player, fmt.Sprintf("You whisper, \"%s\" to %s.", arg2, DB.Fetch(who).name))
		} else {
			notify(player, fmt.Sprintf("%s is not connected.", DB.Fetch(who).name))
		}
	}
}

func do_pose(player ObjectID, message string) {
	if loc := df.Fetch(player).Location; loc != NOTHING {
		notify_except(DB.Fetch(loc).Contents, NOTHING, fmt.Sprintf(DB.Fetch(player).name, " ", message), player)
	}
}

func do_wall(player ObjectID, message string) {
	if Wizard(player) && IsPlayer(player) {
		p := DB.Fetch(player)
		log_status("WALL from %s(%d): %s", p.name, player, message)
		message = fmt.Sprintf("%s shouts, \"%s\"", p.name, message)
		EachObject(func(obj ObjectID) {
			if IsPlayer(obj) {
				notify_from(player, i, message)
			}
		})
	} else {
		notify(player, "But what do you want to do with the wall?")
	}
}

func do_gripe(player ObjectID, message string) {
	if message == "" {
		if Wizard(player) {
			spit_file(player, LOG_GRIPE)
		} else {
			notify(player, "If you wish to gripe, use 'gripe <message>'.")
		}
	} else {
		loc := DB.Fetch(player).Location
		log_gripe("GRIPE from %s(%d) in %s(%d): %s", DB.Fetch(player).name, player, DB.Fetch(loc).name, loc, message)
		notify(player, "Your complaint has been duly noted.")
		wall_wizards(fmt.Sprintf("## GRIPE from %s: %s", DB.Fetch(player).name, message))
	}
}

/* doesn't really belong here, but I couldn't figure out where else */
func do_page(player ObjectID, arg1, arg2 string) {
	switch target := lookup_player(arg1); {
	case !payfor(player, tp_lookup_cost) {
		notify_fmt(player, "You don't have enough %s.", tp_pennies)
	case target == NOTHING:
		notify(player, "I don't recognize that name.")
	case DB.Fetch(target).flags & HAVEN != 0:
		notify(player, "That player does not wish to be disturbed.")
	default:
		var buf string
		if blank(arg2) {
			buf = fmt.Sprintf("You sense that %s is looking for you in %s.", DB.Fetch(player).name, DB.Fetch(DB.Fetch(player).Location).name)
		else
			buf = fmt.Sprintf("%s pages from %s: \"%s\"", DB.Fetch(player).name, DB.Fetch(DB.Fetch(player).Location).name, arg2)
		}
		if notify_from(player, target, buf) {
			notify(player, "Your message has been sent.")
		} else {
			notify(player, fmt.Sprintf("%s is not connected.", DB.Fetch(target).name))
		}
	}
}

func notify_listeners(who, xprog, obj, room ObjectID, msg string, isprivate bool) {
	if obj != NOTHING {
		if tp_listeners && (tp_listeners_obj || Typeof(obj) == TYPE_ROOM) {
			listenqueue(-1, who, room, obj, obj, xprog, "_listen", msg, tp_listen_mlev, 1, 0)
			listenqueue(-1, who, room, obj, obj, xprog, "~listen", msg, tp_listen_mlev, 1, 1)
			listenqueue(-1, who, room, obj, obj, xprog, "~olisten", msg, tp_listen_mlev, 0, 1)
		}

		if tp_zombies && Typeof(obj) == TYPE_THING && !isprivate {
			if DB.Fetch(obj).flags & VEHICLE != 0 && df.Fetch(who).Location == df.Fetch(obj).Location {
				prefix := do_parse_prop(-1, who, obj, MESGPROP_OECHO, "(@Oecho)", MPI_ISPRIVATE)
				if prefix = "" {
					prefix = "Outside>"
				}
				buf := fmt.Sprint(prefix, " ", msg)
				for ref := DB.Fetch(obj).Contents; ref != NOTHING; ref = DB.Fetch(ref).next {
					notify_filtered(who, ref, buf, isprivate)
				}
			}
		}
		switch obj.(type)
		case TYPE_PLAYER, TYPE_THING:
			notify_filtered(who, obj, msg, isprivate)
		}
	}
}

void
notify_except(ObjectID first, ObjectID exception, const char *msg, ObjectID who)
{
	ObjectID room, srch;

	if (first != NOTHING) {

		srch = room = DB.Fetch(first).Location

		if (tp_listeners) {
			notify_from_echo(who, srch, msg, 0);

			if (tp_listeners_env) {
				srch = DB.Fetch(srch).Location
				while (srch != NOTHING) {
					notify_from_echo(who, srch, msg, 0);
					srch = getparent(srch);
				}
			}
		}

		for ; first != NOTHING; first = DB.Fetch(var).next {
			if ((Typeof(first) != TYPE_ROOM) && (first != exception)) {
				/* don't want excepted player or child rooms to hear */
				notify_from_echo(who, first, msg, 0);
			}
		}
	}
}

func parse_oprop(descr int, player, dest, exit ObjectID, propname, prefix, whatcalled string) {
	msg := get_property_class(exit, propname)
	int ival = 0;
	if (Prop_Blessed(exit, propname))
		ival |= MPI_ISBLESSED;

	if (msg)
		parse_omessage(descr, player, dest, exit, msg, prefix, whatcalled, ival);
}

func parse_omessage(descr int, player, dest, exit ObjectID, msg, prefix, whatcalled string, mpiflags int) {
	buf := do_parse_mesg(descr, player, exit, msg, whatcalled, MPI_ISPUBLIC | mpiflags)
	if ptr := pronoun_substitute(descr, player, buf); ptr != "" {
		/*
			TODO: Find out if this should be prefixing with DB.Fetch(player).name,
			or if it should use the prefix argument...  The original code just ignored
			the prefix argument...
		*/
		notify_except(DB.Fetch(dest).Contents, player, prefix_message(ptr, prefix), player)
	}
}

func blank(s string) bool {
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	return len(s) == 0
}