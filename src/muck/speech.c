/* $Header: /cvsroot/fbmuck/fbmuck/src/speech.c,v 1.13 2006/04/19 02:58:54 premchai21 Exp $ */


#include "copyright.h"
#include "config.h"

#include "db.h"
#include "mpi.h"
#include "interface.h"
#include "match.h"
#include "params.h"
#include "tune.h"
#include "props.h"
#include "externs.h"
#include "speech.h"
#include <ctype.h>

/* Commands which involve speaking */

func do_say(dbref player, const char *message) {
	if loc := db.Fetch(player).location; loc != NOTHING {
		notify(player, fmt.Sprintf("You say, \"%s\"", message))
		notify_except(db.Fetch(loc).contents, player, fmt.Sprintf("%s says, \"%s\"", db.Fetch(player).name, message), player)
	}
}

func do_whisper(int descr, dbref player, const char *arg1, const char *arg2) {
	md := NewMatch(descr, player, arg1, TYPE_PLAYER)
	match_neighbor(&md)
	match_me(&md)
	if Wizard(player) && Typeof(player) == TYPE_PLAYER {
		match_absolute(&md)
		match_player(&md)
	}
	switch who = match_result(&md); who {
	case NOTHING:
		notify(player, "Whisper to whom?");
	case AMBIGUOUS:
		notify(player, "I don't know who you mean!");
	default:
		if notify_from(player, who, fmt.Sprintf("%s whispers, \"%s\"", db.Fetch(player).name, arg2)) {
			notify(player, fmt.Sprintf("You whisper, \"%s\" to %s.", arg2, db.Fetch(who).name))
		} else {
			notify(player, fmt.Sprintf("%s is not connected.", db.Fetch(who).name))
		}
	}
}

func do_pose(player dbref, message string) {
	if loc := df.Fetch(player).location; loc != NOTHING {
		notify_except(db.Fetch(loc).contents, NOTHING, fmt.Sprintf(db.Fetch(player).name, " ", message), player)
	}
}

func do_wall(player dbref, message string) {
	if Wizard(player) && Typeof(player) == TYPE_PLAYER {
		log_status("WALL from %s(%d): %s", db.Fetch(player).name, player, message)
		buf := fmt.Sprintf("%s shouts, \"%s\"", db.Fetch(player).name, message)
		for i := 0; i < db_top; i++ {
			if Typeof(i) == TYPE_PLAYER {
				notify_from(player, i, buf)
			}
		}
	} else {
		notify(player, "But what do you want to do with the wall?")
	}
}

func do_gripe(player dbref, message string) {
	if message == "" {
		if Wizard(player) {
			spit_file(player, LOG_GRIPE)
		} else {
			notify(player, "If you wish to gripe, use 'gripe <message>'.")
		}
	} else {
		loc := db.Fetch(player).location
		log_gripe("GRIPE from %s(%d) in %s(%d): %s", db.Fetch(player).name, player, db.Fetch(loc).name, loc, message)
		notify(player, "Your complaint has been duly noted.")
		wall_wizards(fmt.Sprintf("## GRIPE from %s: %s", db.Fetch(player).name, message))
	}
}

/* doesn't really belong here, but I couldn't figure out where else */
func do_page(player dbref, arg1, arg2 string) {
	switch target := lookup_player(arg1); {
	case !payfor(player, tp_lookup_cost) {
		notify_fmt(player, "You don't have enough %s.", tp_pennies)
	case target == NOTHING:
		notify(player, "I don't recognize that name.")
	case db.Fetch(target).flags & HAVEN != 0:
		notify(player, "That player does not wish to be disturbed.")
	default:
		var buf string
		if blank(arg2) {
			buf = fmt.Sprintf("You sense that %s is looking for you in %s.", db.Fetch(player).name, db.Fetch(db.Fetch(player).location).name)
		else
			buf = fmt.Sprintf("%s pages from %s: \"%s\"", db.Fetch(player).name, db.Fetch(db.Fetch(player).location).name, arg2)
		}
		if notify_from(player, target, buf) {
			notify(player, "Your message has been sent.")
		} else {
			notify(player, fmt.Sprintf("%s is not connected.", db.Fetch(target).name))
		}
	}
}

func notify_listeners(who, xprog, obj, room dbref, msg string, isprivate bool) {
	if obj != NOTHING {
		if tp_listeners && (tp_listeners_obj || Typeof(obj) == TYPE_ROOM) {
			listenqueue(-1, who, room, obj, obj, xprog, "_listen", msg, tp_listen_mlev, 1, 0)
			listenqueue(-1, who, room, obj, obj, xprog, "~listen", msg, tp_listen_mlev, 1, 1)
			listenqueue(-1, who, room, obj, obj, xprog, "~olisten", msg, tp_listen_mlev, 0, 1)
		}

		if tp_zombies && Typeof(obj) == TYPE_THING && !isprivate {
			if db.Fetch(obj).flags & VEHICLE != 0 && df.Fetch(who).location == df.Fetch(obj).location {
				prefix := do_parse_prop(-1, who, obj, MESGPROP_OECHO, "(@Oecho)", MPI_ISPRIVATE)
				if prefix = "" {
					prefix = "Outside>"
				}
				buf := fmt.Sprint(prefix, " ", msg)
				for ref := db.Fetch(obj).contents; ref != NOTHING; ref = db.Fetch(ref).next {
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
notify_except(dbref first, dbref exception, const char *msg, dbref who)
{
	dbref room, srch;

	if (first != NOTHING) {

		srch = room = db.Fetch(first).location

		if (tp_listeners) {
			notify_from_echo(who, srch, msg, 0);

			if (tp_listeners_env) {
				srch = db.Fetch(srch).location
				while (srch != NOTHING) {
					notify_from_echo(who, srch, msg, 0);
					srch = getparent(srch);
				}
			}
		}

		for ; first != NOTHING; first = db.Fetch(var).next {
			if ((Typeof(first) != TYPE_ROOM) && (first != exception)) {
				/* don't want excepted player or child rooms to hear */
				notify_from_echo(who, first, msg, 0);
			}
		}
	}
}

func parse_oprop(descr int, player, dest, exit dbref, propname, prefix, whatcalled string) {
	msg := get_property_class(exit, propname)
	int ival = 0;
	if (Prop_Blessed(exit, propname))
		ival |= MPI_ISBLESSED;

	if (msg)
		parse_omessage(descr, player, dest, exit, msg, prefix, whatcalled, ival);
}

func parse_omessage(descr int, player, dest, exit dbref, msg, prefix, whatcalled string, mpiflags int) {
	buf := do_parse_mesg(descr, player, exit, msg, whatcalled, MPI_ISPUBLIC | mpiflags)
	if ptr := pronoun_substitute(descr, player, buf); ptr != "" {
		/*
			TODO: Find out if this should be prefixing with db.Fetch(player).name,
			or if it should use the prefix argument...  The original code just ignored
			the prefix argument...
		*/
		notify_except(db.Fetch(dest).contents, player, prefix_message(ptr, prefix), player)
	}
}

func blank(s string) bool {
	s = strings.TrimLeftFunc(s, unicode.IsSpace)
	return len(s) == 0
}