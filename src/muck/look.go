package fbmuck

#define EXEC_SIGNAL '@'			/* Symbol which tells us what we're looking at is an execution order and not a message.    */

/* prints owner of something */
func print_owner(player, thing dbref) {
	var buf string
	switch Typeof(thing) {
	case TYPE_PLAYER:
		buf = fmt.Sprintf("%s is a player.", db.Fetch(thing).name)
	case TYPE_ROOM, TYPE_THING, TYPE_EXIT, TYPE_PROGRAM:
		buf = fmt.Sprintf("Owner: %s", db.Fetch(db.Fetch(thing).owner).name)
	}
	notify(player, buf)
}

void
exec_or_notify_prop(int descr, dbref player, dbref thing,
					const char *propname, const char *whatcalled)
{
	const char *message = get_property_class(thing, propname);
	int mpiflags = Prop_Blessed(thing, propname)? MPI_ISBLESSED : 0;

	if (message)
		exec_or_notify(descr, player, thing, message, whatcalled, mpiflags);
}

func exec_or_notify(descr int, player, thing dbref, message, whatcalled string, mpiflags int) {
	const char *p;
	char *p2;
	char *p3;
	char buf[BUFFER_LEN];
	char tmpcmd[BUFFER_LEN];
	char tmparg[BUFFER_LEN];

	p = message;

	if (*p == EXEC_SIGNAL) {
		int i;

		if (*(++p) == REGISTERED_TOKEN) {
			buf = p
			p2 = strings.TrimLeftFunc(buf, func(r rune) bool {
				return !unicode.IsSpace(r)
			})
			if len(p) > 0 {
				p = p[1:]
			}

			p3 = strings.TrimLeftFunc(buf, func(r rune) bool {
				return !unicode.IsSpace(r)
			})

			if (*p2) {
				i = find_registered_obj(thing, p2)
			} else {
				i = 0;
			}
		} else {
			i = strings.Atoi(p)
			p = strings.TrimLeftFunc(p, func(r rune) bool {
				return !unicode.IsSpace(r)
			})
			if len(p) > 0 {
				p = p[1:]
			}
		}
		if (i < 0 || i >= db_top || (Typeof(i) != TYPE_PROGRAM)) {
			if (*p) {
				notify(player, p);
			} else {
				notify(player, "You see nothing special.");
			}
		} else {
			tmparg = match_args
			tmpcmd = match_cmdname
			match_args = do_parse_mesg(descr, player, thing, p, whatcalled, MPI_ISPRIVATE | mpiflags)
			match_cmdname = whatcalled
			if tmpfr := interp(descr, player, db.Fetch(player).location, i, thing, PREEMPT, STD_HARDUID, 0); tmpfr != nil {
				interp_loop(player, i, tmpfr, false)
			}
			match_args = tmparg
			match_cmdname = tmpcmd
		}
	} else {
		notify(player, do_parse_mesg(descr, player, thing, p, whatcalled, MPI_ISPRIVATE | mpiflags))
	}
}

static void
look_contents(dbref player, dbref loc, const char *contents_name)
{
	dbref thing;
	dbref can_see_loc;

	/* check to see if he can see the location */
	can_see_loc = (!Dark(loc) || controls(player, loc));

	/* check to see if there is anything there */
	for thing = db.Fetch(loc).contents; thing != NOTHING; thing = db.Fetch(thing).next {
		if (can_see(player, thing, can_see_loc)) {
			/* something exists!  show him everything */
			notify(player, contents_name);
			for thing = db.Fetch(loc).contents; thing != NOTHING; thing = db.Fetch(thing).next {
				if (can_see(player, thing, can_see_loc)) {
					notify(player, unparse_object(player, thing));
				}
			}
			break;				/* we're done */
		}
	}
}

func look_simple(descr int, player, thing dbref) {
	if v := get_property_class(thing, MESGPROP_DESC); v != "" {
		if Prop_Blessed(loc, MESGPROP_DESC) {
			exec_or_notify(descr, player, thing, v, "(@Desc)", MPI_ISBLESSED)
		} else {
			exec_or_notify(descr, player, thing, v, "(@Desc)", 0)
		}
	} else {
		notify(player, "You see nothing special.")
	}
}

func look_room(descr int, player, loc dbref) {
	notify(player, unparse_object(player, loc))

	if _, ok := loc.data.(TYPE_ROOM); ok {
		if v := get_property_class(loc, MESGPROP_DESC); v != "" {
			if Prop_Blessed(loc, MESGPROP_DESC) {
				exec_or_notify(descr, player, loc, v, "(@Desc)", MPI_ISBLESSED)
			} else {
				exec_or_notify(descr, player, loc, v, "(@Desc)", 0)
			}
		}
		/* print appropriate messages if player has the key */
		can_doit(descr, player, loc, 0)
	} else {
		if v := get_property_class(loc, MESGPROP_IDESC); v != "" {
			if Prop_Blessed(loc, MESGPROP_IDESC) {
				exec_or_notify(descr, player, loc, v, "(@Idesc)", MPI_ISBLESSED)
			} else {
				exec_or_notify(descr, player, loc, v, "(@Idesc)", 0)
			}
		}
	}
	ts_useobject(loc)

	look_contents(player, loc, "Contents:")
	if tp_look_propqueues {
		envpropqueue(descr, player, loc, player, loc, NOTHING, "_lookq", fmt.Sprintf("#%d", loc), 1, 1)
	}
}

func do_look_around(descr int, player dbref) {
	if loc = db.Fetch(player).location; loc != NOTHING {
		look_room(descr, player, loc)
	}
}

func do_look_at(int descr, dbref player, const char *name, const char *detail) {
	if name == "" || name == "here" {
		if thing := db.Fetch(player).location; thing != NOTHING {
			look_room(descr, player, thing)
		}
	} else {
		md := NewMatch(descr, player, name, NOTYPE).
			MatchAllExits().
			MatchNeighbor().
			MatchPossession()
		if Wizard(db.Fetch(player).owner) {
			md.MatchAbsolute().MatchPlayer()
		}
		switch thing := md.MatchHere().MatchMe().Matchresult(); {
		case thing != NOTHING && thing != AMBIGUOUS && detail == "":
			switch TYPEOF(thing) {
			case TYPE_ROOM:
				if db.Fetch(player).location != thing && !can_link_to(player, TYPE_ROOM, thing) {
					notify(player, "Permission denied. (you're not where you want to look, and can't link to it)")
				} else {
					look_room(descr, player, thing)
				}
			case TYPE_PLAYER:
				if db.Fetch(player).location != db.Fetch(thing).location && !controls(player, thing) {
					notify(player, "Permission denied. (Your location isn't the same as what you're looking at)")
				} else {
					look_simple(descr, player, thing)
					look_contents(player, thing, "Carrying:")
					if tp_look_propqueues {
						envpropqueue(descr, player, thing, player, thing, NOTHING, "_lookq", fmt.Sprintf("#%d", thing), 1, 1)
					}
				}
			case TYPE_THING:
				if db.Fetch(player).location != db.Fetch(thing).location && db.Fetch(thing).location != player && !controls(player, thing) {
					notify(player, "Permission denied. (You're not in the same room as or carrying the object)")
				} else {
					look_simple(descr, player, thing)
					if db.Fetch(thing).flags & HAVEN == 0 {
						look_contents(player, thing, "Contains:")
						ts_useobject(thing)
					}
					if tp_look_propqueues {
						envpropqueue(descr, player, thing, player, thing, NOTHING, "_lookq", fmt.Sprintf("#%d", thing), 1, 1)
					}
				}
			default:
				look_simple(descr, player, thing)
				if TYPEOF(thing) != TYPE_PROGRAM {
					ts_useobject(thing)
				}
				if tp_look_propqueues {
					envpropqueue(descr, player, thing, player, thing, NOTHING, "_lookq", fmt.Sprintf("#%d", thing), 1, 1)
				}
			}
		case thing == NOTHING, detail != "" && thing != AMBIGUOUS:
			var ambiguous bool
			char propname[BUFFER_LEN];
			var propadr, pptr, lastmatch *Plist

			var buf string
			if thing == NOTHING {
				thing = db.Fetch(player).location
				buf = fmt.Sprint(name)
			} else {
				buf = fmt.Sprint(detail)
			}
			lastmatch = nil
			for propadr := pptr.first_prop(thing, "_details/", propname); propadr != nil && ! ambiguous; propadr, propname = propadr.next_prop(pptr) {
				if exit_prefix(propname, buf) {
					if lastmatch != nil {
						lastmatch = nil
						ambiguous = true
					} else {
						lastmatch = propadr
					}
				}
			}
			if v, ok := lastmatch.data.(string); ok {
				if PropFlags(lastmatch) & PROP_BLESSED != 0 {
					exec_or_notify(descr, player, thing, v, "(@detail)", MPI_ISBLESSED)
				} else {
					exec_or_notify(descr, player, thing, v, "(@detail)", 0)
				}
			} else {
				switch {
				case ambiguous:
					notify(player, AMBIGUOUS_MESSAGE)
				case detail != "":
					notify(player, "You see nothing special.")
				default:
					notify(player, NOMATCH_MESSAGE)
				}
			}
		default:
			notify(player, AMBIGUOUS_MESSAGE)
		}
	}
}

#ifdef VERBOSE_EXAMINE
func flag_description(thing dbref) (r string) {
	r = "Type: "
	switch Typeof(thing) {
	case TYPE_ROOM:
		r += "ROOM"
	case TYPE_EXIT:
		r += "EXIT/ACTION"
	case TYPE_THING:
		r += "THING"
	case TYPE_PLAYER:
		r += "PLAYER"
	case TYPE_PROGRAM:
		r += "PROGRAM"
	default:
		r += "***UNKNOWN TYPE***"
	}

	if db.Fetch(thing).flags & ~TYPE_MASK != 0 {
		r += "  Flags:"
		if db.Fetch(thing).flags & WIZARD != 0 {
			r += " WIZARD"
		}
		if db.Fetch(thing).flags & QUELL != 0 {
			r += " QUELL"
		}
		if db.Fetch(thing).flags & STICKY != 0 {
			switch Typeof(thing) {
			case TYPE_PROGRAM:
				r += " SETUID"
			case TYPE_PLAYER:
				r += " SILENT"
			default:
				r += " STICKY"
			}
		}
		if db.Fetch(thing).flags & DARK != 0 {
			if Typeof(thing) == TYPE_PROGRAM {
				r += " DEBUGGING"
			} else {
				r += " DARK"
			}
		}
		if db.Fetch(thing).flags & LINK_OK != 0 {
			r += " LINK_OK"
		}
		if db.Fetch(thing).flags & KILL_OK != 0 {
			r += " KILL_OK"
		}
		if MLevRaw(thing) != NON_MUCKER {
			r += " MUCKER"
			switch MLevRaw(thing) {
			case APPRENTIVE:
				r += "1"
			case JOURNEYMAN:
				r += "2"
			case MASTER:
				r += "3"
			}
		}
		if db.Fetch(thing).flags & BUILDER != 0 {
			if Typeof(thing) == TYPE_PROGRAM {
				r += " BOUND"
			} else {
				r += " BUILDER"
			}
		}
		if db.Fetch(thing).flags & CHOWN_OK != 0 {
			if Typeof(thing) == TYPE_PLAYER {
				r += " COLOR"
			} else {
				r += " CHOWN_OK"
			}
		}
		if db.Fetch(thing).flags & JUMP_OK != 0 {
			r += " JUMP_OK"
		}
		if db.Fetch(thing).flags & VEHICLE != 0 {
			if Typeof(thing) == TYPE_PROGRAM {
				r += " VIEWABLE"
			} else {
				r += " VEHICLE"
			}
		}
		if tp_enable_match_yield && db.Fetch(thing).flags & YIELD != 0 {
			r += " YIELD"
		}
		if tp_enable_match_yield && db.Fetch(thing).flags & OVERT != 0 {
			r += " OVERT"
		}
		if db.Fetch(thing).flags & XFORCIBLE != 0 {
			if Typeof(thing) == TYPE_EXIT {
				r += " XPRESS"
			} else {
				r += " XFORCIBLE"
			}
		}
		if db.Fetch(thing).flags & ZOMBIE != 0 {
			r += " ZOMBIE"
		}
		if db.Fetch(thing).flags & HAVEN != 0 {
			switch Typeof(thing) {
			case TYPE_PROGRAM:
				r += " HARDUID"
			case TYPE_THING:
				r += " HIDE"
			default:
				r += " HAVEN"
			}
		}
		if db.Fetch(thing).flags & ABODE != 0 {
			switch Typeof(thing) {
			case TYPE_PROGRAM:
				r += " AUTOSTART"
			case TYPE_EXIT:
				r += " ABATE"
			default:
				r += " ABODE"
			}
		}
	}
	return
}

#endif							/* VERBOSE_EXAMINE */

func listprops_wildcard(dbref player, dbref thing, const char *dir, const char *wild) int {
	char propname[BUFFER_LEN];
	char wld[BUFFER_LEN];
	char buf[BUFFER_LEN];
	char buf2[BUFFER_LEN];
	char *ptr, *wldcrd = wld;
	int i, cnt = 0;
	int recurse = 0;

	strcpyn(wld, sizeof(wld), wild);
	i = len(wld);
	if (i && wld[i - 1] == PROPDIR_DELIMITER)
		strcatn(wld, sizeof(wld), "*");
	wldcrd = strings.TrimLeft(wld, PROPDIR_DELIMITER)
	if strings.Compare(wldcrd, "**") {
		recurse = 1
	}

	for (ptr = wldcrd; *ptr && *ptr != PROPDIR_DELIMITER; ptr++) ;
	if (*ptr)
		*ptr++ = '\0';

	var pptr *Plist
	propadr := pptr.first_prop(thing, dir, propname)
	for propadr != nil {
		if !smatch(wldcrd, propname) {
			buf = fmt.Sprint(dir, PROPDIR_DELIMITER, propname)
			if (!Prop_System(buf) && ((!Prop_Hidden(buf) && !(PropFlags(propadr) & PROP_SYSPERMS)) || Wizard(db.Fetch(player).owner))) {
				if (!*ptr || recurse) {
					cnt++;
					displayprop(player, thing, buf, buf2, sizeof(buf2));
					notify(player, buf2);
				}
				if (recurse)
					ptr = "**";
				cnt += listprops_wildcard(player, thing, buf, ptr);
			}
		}
		propadr, propname = propadr.next_prop(pptr)
	}
	return cnt;
}

func do_examine(descr int, player dbref, name, dir string) {
	var thing, content, exit dbref
	int i, cnt;
	struct tm *time_tm;			/* used for timestamps */

	if name == "" {
		if thing = db.Fetch(player).location; thing == NOTHING {
			return
		}
	} else {
		md := NewMatch(descr, player, name, NOTYPE).
			MatchAllExits().
			MatchNeighbor().
			MatchPossession().
			MatchAbsolute().
			MatchRegistered()

		/* only Wizards can examine other players */
		if Wizard(db.Fetch(player).owner) {
			md.MatchPlayer()
		}
		md.MatchHere().MatchMe()

		/* get result */
		if thing = md.NoisyMatchResult(); thing == NOTHING {
			return
		}
	}

	if !can_link(player, thing) {
		print_owner(player, thing)
	} else {
		if (*dir) {
			/* show him the properties */
			if cnt := listprops_wildcard(player, thing, "", dir); cnt == 1 {
				notify(player, fmt.Sprintf(cnt, "property listed."))
			} else {
				notify(player, fmt.Sprint(cnt, "properties listed."))
			}
			return;
		}
		var buf string
		switch thing.(type) {
		case TYPE_ROOM:
			buf = fmt.Sprintf("%s  Owner: %s  Parent: %s", unparse_object(player, thing), db.Fetch(db.Fetch(thing).owner).name, unparse_object(player, db.Fetch(thing).location)
		case TYPE_THING:
			buf = fmt.Sprintf("%s  Owner: %s  Value: %d", unparse_object(player, thing), db.Fetch(db.Fetch(thing).owner).name, get_property_value(thing, MESGPROP_VALUE))
		case TYPE_PLAYER:
			buf = fmt.Sprintf("%s  %s: %d  ", unparse_object(player, thing), tp_cpennies, get_property_value(thing, MESGPROP_VALUE))
		case TYPE_EXIT, TYPE_PROGRAM:
			buf = fmt.Sprintf("%s  Owner: %s", unparse_object(player, thing), db.Fetch(db.Fetch(thing).owner).name)
		}
		notify(player, buf)
#ifdef VERBOSE_EXAMINE
		notify(player, flag_description(thing));
#endif							/* VERBOSE_EXAMINE */

		if get_property_class(thing, MESGPROP_DESC) {
			notify(player, get_property_class(thing, MESGPROP_DESC))
		}
		notify(player, fmt.Sprintf("Key: %s", get_property_lock(thing, MESGPROP_LOCK).Unparse(player, true)))
		notify(player, fmt.Sprintf("Chown_OK Key: %s", get_property_lock(thing, "_/chlk").Unparse(player, true)))
		notify(player, fmt.Sprintf("Container Key: %s", get_property_lock(thing, "_/clk").Unparse(player, true)))
		notify(player, fmt.Sprintf("Force Key: %s", get_property_lock(thing, "@/flk").Unparse(player, true)))

		if get_property_class(thing, MESGPROP_SUCC) {
			notify(player, fmt.Sprintf("Success: %s", get_property_class(thing, MESGPROP_SUCC)))
		}
		if get_property_class(thing, MESGPROP_FAIL) {
			notify(player, fmt.Sprintf("Fail: %s", get_property_class(thing, MESGPROP_FAIL)))
		}
		if get_property_class(thing, MESGPROP_DROP) {
			notify(player, fmt.Sprintf("Drop: %s", get_property_class(thing, MESGPROP_DROP)))
		}
		if get_property_class(thing, MESGPROP_OSUCC) {
			notify(player, fmt.Sprintf("Osuccess: %s", get_property_class(thing, MESGPROP_OSUCC)))
		}
		if get_property_class(thing, MESGPROP_OFAIL) {
			notify(player, fmt.Sprintf("Ofail: %s", get_property_class(thing, MESGPROP_OFAIL)))
		}
		if get_property_class(thing, MESGPROP_ODROP) {
			notify(player, fmt.Sprintf("Odrop: %s", get_property_class(thing, MESGPROP_ODROP)))
		}
		if tp_who_doing && get_property_class(thing, MESGPROP_DOING) {
			notify(player, fmt.Sprintf("Doing: %s", get_property_class(thing, MESGPROP_DOING)))
		}
		if get_property_class(thing, MESGPROP_OECHO) {
			notify(player, fmt.Sprintf("Oecho: %s", get_property_class(thing, MESGPROP_OECHO)))
		}
		if get_property_class(thing, MESGPROP_PECHO) {
			notify(player, fmt.Sprintf("Pecho: %s", get_property_class(thing, MESGPROP_PECHO)))
		}
		if get_property_class(thing, MESGPROP_IDESC) {
			notify(player, fmt.Sprintf("Idesc: %s", get_property_class(thing, MESGPROP_IDESC)))
		}

		/* Timestamps */
		/* ex: time_tm = localtime((time_t *)(&(db.Fetch(thing).ts.created))); */
		time_tm = localtime((&(db.Fetch(thing).ts.created)))
		notify(player, format_time((char *) "Created:  %a %b %e %T %Z %Y", time_tm))
		time_tm = localtime((&(db.Fetch(thing).ts.modified)))
		notify(player, format_time((char *) "Modified: %a %b %e %T %Z %Y", time_tm))
		time_tm = localtime((&(db.Fetch(thing).ts.lastused)))
		notify(player, format_time((char *) "Lastused: %a %b %e %T %Z %Y", time_tm))
		if TYPEOF(thing) == TYPE_PROGRAM {
			var i int
			if db.Fetch(thing).sp.(program_specific) != nil {
				i = db.Fetch(thing).sp.(program_specific).instances
			}
			notify(player, fmt.Sprintf("Usecount: %d     Instances: %d", db.Fetch(thing).ts.usecount, i))
		} else {
			notify(player, fmt.Sprintf("Usecount: %d", db.Fetch(thing).ts.usecount))
		}

		notify(player, "[ Use 'examine <object>=/' to list root properties. ]")

		/* show him the contents */
		if db.Fetch(thing).contents != NOTHING {
			if TYPEOF(thing) == TYPE_PLAYER {
				notify(player, "Carrying:");
			} else {
				notify(player, "Contents:")
			}
			for content = db.Fetch(thing).contents; content != NOTHING; content = db.Fetch(content).next {
				notify(player, unparse_object(player, content))
			}
		}
		switch thing.(type) {
		case TYPE_ROOM:
			/* tell him about exits */
			if db.Fetch(thing).exits != NOTHING {
				notify(player, "Exits:");
				for exit = db.Fetch(thing).exits; exit != NOTHING; exit = db.Fetch(exit)next {
					notify(player, unparse_object(player, exit));
				}
			} else {
				notify(player, "No exits.");
			}

			/* print dropto if present */
			if db.Fetch(thing).sp != NOTHING {
				notify(player, fmt.Sprintf("Dropped objects go to: %s", unparse_object(player, db.Fetch(thing).sp.(dbref))))
			}
		case TYPE_THING:
			/* print home */
			notify(player, fmt.Sprintf("Home: %s", unparse_object(player, db.Fetch(thing).sp.(player_specific).home)))	/* home */
			/* print location if player can link to it */
			if db.Fetch(thing).location != NOTHING && (controls(player, db.Fetch(thing).location) || can_link_to(player, NOTYPE, db.Fetch(thing).location)) {
				notify(player, fmt.Sprintf("Location: %s", unparse_object(player, db.Fetch(thing).location)))
			}
			/* print thing's actions, if any */
			if db.Fetch(thing).exits != NOTHING {
				notify(player, "Actions/exits:")
				for exit = db.Fetch(thing).exits; exit != NOTHING; exit = db.Fetch(exit).next {
					notify(player, unparse_object(player, exit))
				}
			} else {
				notify(player, "No actions attached.")
			}
		case TYPE_PLAYER:
			/* print home */
			notify(player, fmt.Sprintf("Home: %s", unparse_object(player, db.Fetch(thing).sp.(player_specific).home)))

			/* print location if player can link to it */
			if db.Fetch(thing).location != NOTHING && (controls(player, db.Fetch(thing).location) || can_link_to(player, NOTYPE, db.Fetch(thing).location)) {
				notify(player, fmt.Sprintf("Location: %s", unparse_object(player, db.Fetch(thing).location)))
			}
			/* print player's actions, if any */
			if db.Fetch(thing).exits != NOTHING {
				notify(player, "Actions/exits:")
				for exit = db.Fetch(thing.exits); exit != NOTHING; exit = db.Fetch(exit).next {
					notify(player, unparse_object(player, exit))
				}
			} else {
				notify(player, "No actions attached.")
			}
		case TYPE_EXIT:
			if db.Fetch(thing).location != NOTHING {
				notify(player, fmt.Sprintf("Source: %s", unparse_object(player, db.Fetch(thing).location)))
			}
			/* print destinations */
			for _, v := range db.Fetch(thing).sp.exit.dest {
				switch v {
				case NOTHING:
				case HOME:
					notify(player, "Destination: *HOME*");
				default:
					notify(player, fmt.Sprintf("Destination: %s", unparse_object(player, v)))
				}
			}
		case TYPE_PROGRAM:
			if len(db.Fetch(thing).sp.(program_specific).code) > 0 {
				struct timeval tv = db.Fetch(thing).sp.(program_specific).proftime
				notify(player, fmt.Sprintf("Program compiled size: %d instructions", len(db.Fetch(thing).sp.(program_specific).code)))
				notify(player, fmt.Sprintf("Cumulative runtime: %d.%06d seconds ", int(tv.tv_sec), int(tv.tv_usec)))
			} else {
				notify(player, fmt.Sprintf("Program not compiled."))
			}
			/* print location if player can link to it */
			if db.Fetch(thing).location != NOTHING && (controls(player, db.Fetch(thing).location) || can_link_to(player, NOTYPE, db.Fetch(thing).location)) {
				notify(player, fmt.Sprintf("Location: %s", unparse_object(player, db.Fetch(thing).location)))
			}
		}
	}
}

func do_score(player dbref) {
	if v := get_property_value(player, MESGPROP_VALUE); v == 1 {
		notify(player, fmt.Sprintf("You have %d %s.", v, tp_penny))
	} else {
		notify(player, fmt.Sprintf("You have %d %s.", v, tp_pennies))
	}
}

func do_inventory(player dbref) {
	dbref thing;

	if thing := db.Fetch(player).contents; thing == NOTHING {
		notify(player, "You aren't carrying anything.")
	} else {
		notify(player, "You are carrying:")
		for ; thing != NOTHING; thing = db.Fetch(thing).next {
			notify(player, unparse_object(player, thing))
		}
	}
	do_score(player)
}

func init_checkflags(player dbref, flags string) (output_type int, check *flgchkdat) {
	check = new(flgchkdat)

	char buf[BUFFER_LEN];
	char *cptr;
	int output_type = 0;
	int mode = 0;

	buf := flags
	for (cptr = buf; *cptr && (*cptr != '='); cptr++) ;
	if (*cptr == '=')
		*(cptr++) = '\0';
	flags = buf
	cptr = strings.TrimLeftFunc(cptr, unicode.IsSpace)
	switch {
	case strings.Prefix("owners", cptr):
		output_type = 1
	case strings.Prefix("locations", cptr):
		output_type = 3
	case strings.Prefix("links", cptr):
		output_type = 2
	case strings.Prefix("count", cptr):
		output_type = 4
	default:
		output_type = 0
	}

	for flags != "" {
		if mode != 0 {
			switch strings.ToUpper(*flags) {
			case '!':
				mode = 0
			case 'R':
				check.isnotroom = true
			case 'T':
				check.isnotthing = true
			case 'E':
				check.isnotexit = true
			case 'P':
				check.isnotplayer = true
			case 'F':
				check.isnotprog = true
			case 'U':
				check.forlink = true
				check.islinked = true
			case '@':
				check.forold = true
				check.isold = false
			case '0':
				check.isnotzero = true
			case '1':
				check.isnotone = true
			case '2':
				check.isnottwo = true
			case '3':
				check.isnotthree = true
			case 'M':
				check.forlevel = true
				check.islevel = 0
			case 'A':
				check.clearflags |= ABODE
			case 'B':
				check.clearflags |= BUILDER
			case 'C':
				check.clearflags |= CHOWN_OK
			case 'D':
				check.clearflags |= DARK
			case 'H':
				check.clearflags |= HAVEN
			case 'J':
				check.clearflags |= JUMP_OK
			case 'K':
				check.clearflags |= KILL_OK
			case 'L':
				check.clearflags |= LINK_OK
			case 'O':
				if tp_enable_match_yield {
					check.clearflags |= OVERT
				}
			case 'Q':
				check.clearflags |= QUELL
			case 'S':
				check.clearflags |= STICKY
			case 'V':
				check.clearflags |= VEHICLE
			case 'Y':
				if tp_enable_match_yield {
					check.clearflags |= YIELD
				}
			case 'Z':
				check.clearflags |= ZOMBIE
			case 'W':
				check.clearflags |= WIZARD
			case 'X':
				check.clearflags |= XFORCIBLE
			case ' ':
				mode = 2
			}
		} else {
			switch strings.ToUpper(*flags) {
			case '!':
				mode = 2
			case 'R':
				check.fortype = true
				check.istype = TYPE_ROOM
			case 'T':
				check.fortype = true
				check.istype = TYPE_THING
			case 'E':
				check.fortype = true
				check.istype = TYPE_EXIT
			case 'P':
				check.fortype = true
				check.istype = TYPE_PLAYER
			case 'F':
				check.fortype = true
				check.istype = TYPE_PROGRAM
			case 'U':
				check.forlink = true
				check.islinked = false
			case '@':
				check.forold = true
				check.isold = true
			case '0':
				check.forlevel = true
				check.islevel = 0
			case '1':
				check.forlevel = true
				check.islevel = 1
			case '2':
				check.forlevel = true
				check.islevel = 2;
				}
			case '3':
				check.forlevel = true
				check.islevel = 3
			case 'M':
				check.isnotzero = true
			case 'A':
				check.setflags |= ABODE
			case 'B':
				check.setflags |= BUILDER
			case 'C':
				check.setflags |= CHOWN_OK
			case 'D':
				check.setflags |= DARK
			case 'H':
				check.setflags |= HAVEN
			case 'J':
				check.setflags |= JUMP_OK
			case 'K':
				check.setflags |= KILL_OK
			case 'L':
				check.setflags |= LINK_OK
			case 'O':
				if tp_enable_match_yield {
					check.setflags |= OVERT
				}
			case 'Q':
				check.setflags |= QUELL
			case 'S':
				check.setflags |= STICKY
			case 'V':
				check.setflags |= VEHICLE
			case 'Y':
				if tp_enable_match_yield {
					check.setflags |= YIELD
				}
			case 'Z':
				check.setflags |= ZOMBIE
			case 'W':
				check.setflags |= WIZARD
			case 'X':
				check.setflags |= XFORCIBLE
			case ' ':
			}
		}
		if mode != 0 {
			mode--
		}
		flags++
	}
	return output_type;
}


func checkflags(dbref what, struct flgchkdat check) (r bool) {
	r = true
	switch {
	case check.fortype && Typeof(what) != check.istype:
		fallthrough
	case check.isnotroom && Typeof(what) == TYPE_ROOM:
		fallthrough
	case check.isnotexit && Typeof(what) == TYPE_EXIT:
		fallthrough
	case check.isnotthing && Typeof(what) == TYPE_THING:
		fallthrough
	case check.isnotplayer && Typeof(what) == TYPE_PLAYER:
		fallthrough
	case check.isnotprog && Typeof(what) == TYPE_PROGRAM:
		fallthrough
	case check.forlevel && MLevRaw(what) != check.islevel:
		fallthrough
	case check.isnotzero && MLevRaw(what) == NON_MUCKER:
		fallthrough
	case check.isnotone && MLevRaw(what) == APPRENTICE:
		fallthrough
	case check.isnottwo && MLevRaw(what) == JOURNEYMAN:
		fallthrough
	case check.isnotthree && MLevRaw(what) == MASTER:
		fallthrough
	case db.Fetch(what).flags & check.clearflags != 0:
		fallthrough
	case ~db.Fetch(what).flags & check.setflags != 0:
		r = false
	}

	if check.forlink {
		switch Typeof(what) {
		case TYPE_ROOM:
			if (db.Fetch(what).sp == NOTHING) != !check.islinked {
				r = false
			}
		case TYPE_EXIT:
			if (len(db.Fetch(what).sp.exit.dest) == 0) != !check.islinked {
				r = false
			}
		case TYPE_PLAYER, TYPE_THING:
			r = check.islinked
		default:
			r = !check.islinked
		}
	}

	if check.forold {
		if (((time(nil)) - db.Fetch(what).ts.lastused) < tp_aging_time) || (((time(nil)) - db.Fetch(what).ts.modified) < tp_aging_time) != !check.isold {
			r = false
		}
	}
	return
}

const(
	CHECK_OWNERS = 1
	CHECK_LINKS = 2
	CHECK_LOCATIONS = 3
)

func display_objinfo(player, obj dbref, output_type int) {
	var buf string
	buf2 := unparse_object(player, obj)
	switch output_type {
	case 0:
	case CHECK_OWNERS:
		buf = fmt.Sprintf("%-38.512s  %.512s", buf2, unparse_object(player, db.Fetch(obj).owner));
	case CHECK_LINKS:
		switch Typeof(obj) {
		case TYPE_ROOM:
			buf = fmt.Sprintf("%-38.512s  %.512s", buf2, unparse_object(player, db.Fetch(obj).sp.(dbref)))
		case TYPE_EXIT:
			switch n := len(db.Fetch(obj).sp.exit.dest); {
			case n == 0:
				buf = fmt.Sprintf("%-38.512s  %.512s", buf2, "*UNLINKED*")
			case n > 1:
				buf = fmt.Sprintf("%-38.512s  %.512s", buf2, "*METALINKED*")
			default:
				buf = fmt.Sprintf("%-38.512s  %.512s", buf2, unparse_object(player, db.Fetch(obj).sp.exit.dest[0]))
			}
		case TYPE_PLAYER:
			buf = fmt.Sprintf("%-38.512s  %.512s", buf2, unparse_object(player, db.Fetch(obj).sp.(player_specific).home))
		case TYPE_THING:
			buf = fmt.Sprintf("%-38.512s  %.512s", buf2, unparse_object(player, db.Fetch(obj).sp.(player_specific).home))
		default:
			buf = fmt.Sprintf("%-38.512s  %.512s", buf2, "N/A")
		}
	case CHECK_LOCATIONS:
		buf = fmt.Sprintf("%-38.512s  %.512s", buf2, unparse_object(player, db.Fetch(obj).location))
	case 4:
		return
	default:
		buf = buf2
	}
	notify(player, buf)
}

func do_find(player dbref, name, flags string) {
	if !payfor(player, tp_lookup_cost) {
		notify_fmt(player, "You don't have enough %s.", tp_pennies)
	} else {
		var total int
		buf := "*" + name + "*"
		output_type, check := init_checkflags(player, flags)
		for i := 0; i < db_top; i++ {
			if (Wizard(db.Fetch(player).owner) || db.Fetch(i).owner == db.Fetch(player).owner) && checkflags(i, check) && db.Fetch(i).name != "" && (name == "" || !smatch(buf, db.Fetch(i).name) {
				display_objinfo(player, i, output_type)
				total++
			}
		}
		notify(player, "***End of List***")
		notify_fmt(player, "%d objects found.", total)
	}
}

func do_owned(dbref player, const char *name, const char *flags) {
	if !payfor(player, tp_lookup_cost) {
		notify_fmt(player, "You don't have enough %s.", tp_pennies)
	} else {
		var victim dbref
		output_type, check := init_checkflags(player, flags)
		if Wizard(db.Fetch(player).owner) && name != "" {
			if victim = lookup_player(name); victim == NOTHING {
				notify(player, "I couldn't find that player.")
				return
			}
		} else {
			victim = player
		}

		var total int
		for i := 0; i < db_top; i++ {
			if db.Fetch(i).owner == db.Fetch(victim).owner && checkflags(i, check) {
				display_objinfo(player, i, output_type)
				total++
			}
		}
		notify(player, "***End of List***")
		notify_fmt(player, "%d objects found.", total)
	}
}

func do_trace(int descr, dbref player, const char *name, int depth) {
	dbref thing;
	int i;

	md := NewMatch(descr, player, name, NOTYPE).
		MatchAbsolute().
		MatchHere().
		MatchMe().
		MatchNeighbor().
		MatchPossession().
		MatchRegistered()
	if thing = md.NoisyMatchResult(); thing == NOTHING || thing == AMBIGUOUS {
		return
	}
	for (i = 0; (!depth || i < depth) && thing != NOTHING; i++) {
		if controls(player, thing) || can_link_to(player, NOTYPE, thing) {
			notify(player, unparse_object(player, thing))
		} else {
			notify(player, "**Missing**")
		}
		thing = db.Fetch(thing).location
	}
	notify(player, "***End of List***");
}

func do_entrances(int descr, dbref player, const char *name, const char *flags) {
	var thing dbref
	if name == "" {
		thing = db.Fetch(player).location
	} else {
		md := NewMatch(descr, player, name, NOTYPE).
			MatchAllExits().
			MatchNeighbor().
			MatchPossession().
			MatchRegistered()
		if Wizard(db.Fetch(player).owner) {
			md.MatchAbsolute().MatchPlayer()
		}
		md.MatchHere()
		md.MatchMe()
		thing = md.NoisyMatchResult()
	}
	var total int
	switch output_type, check := init_checkflags(player, flags); {
	case thing == NOTHING:
		notify(player, "I don't know what object you mean.")
	case !controls(db.Fetch(player).owner, thing):
		notify(player, "Permission denied. (You can't list entrances of objects you don't control)")
	default:
		_, check = init_checkflags(player, flags)
		for i := 0; i < db_top; i++ {
			if checkflags(i, check) {
				switch TYPEOF(i) {
				case TYPE_EXIT:
					for j := len(db.Fetch(i).sp.exit.dest); j > 0; j-- {
						if db.Fetch(i).sp.exit.dest[j] == thing {
							display_objinfo(player, i, output_type)
							total++
						}
					}
				case TYPE_PLAYER:
					if db.Fetch(i).sp.(player_specific).home == thing {
						display_objinfo(player, i, output_type)
						total++
					}
				case TYPE_THING:
					if db.Fetch(i).sp.(player_specific).home == thing {
						display_objinfo(player, i, output_type)
						total++
					}
				case TYPE_ROOM:
					if db.Fetch(i).sp == thing {
						display_objinfo(player, i, output_type)
						total++
					}
				}
			}
		}
		notify(player, "***End of List***")
		notify_fmt(player, "%d objects found.", total)
	}
}

func do_contents(int descr, dbref player, const char *name, const char *flags) {
	var total int
	var i, thing dbref
	if name == "" {
		thing = db.Fetch(player).location
	} else {
		md := NewMatch(descr, player, name, NOTYPE).
			MatchMe().
			MatchHere().
			MatchAllExits().
			MatchNeighbor().
			MatchPossession().
			MatchRegistered()
		if Wizard(db.Fetch(player).owner) {
			md.MatchAbsolute().MatchPlayer()
		}
		thing = md.NoisyMatchResult()
	}
	if thing != NOTHING {
		if !controls(db.Fetch(player).owner, thing) {
			notify(player, "Permission denied. (You can't get the contents of something you don't control)")
		} else {
			output_type, check := init_checkflags(player, flags)
			for i := db.Fetch(thing).contents; i != NOTHING; i = db.Fetch(i).next {
				if checkflags(i, check) {
					display_objinfo(player, i, output_type)
					total++
				}
			}
			switch TYPEOF(thing) {
			case TYPE_ROOM, TYPE_THING, TYPE_PLAYER:
				for i := db.Fetch(thing).exits; i != NOTHING; i = db.Fetch(i).next {
					if checkflags(i, check) {
						display_objinfo(player, i, output_type)
						total++
					}
				}
			}
			notify(player, "***End of List***")
			notify_fmt(player, "%d objects found.", total)
		}
	}
}

func exit_matches_name(exit dbref, name string, exactMatch bool) bool {
	char buf[BUFFER_LEN];
	char *ptr, *ptr2;

	strcpyn(buf, sizeof(buf), db.Fetch(exit).name)
	for (ptr2 = ptr = buf; *ptr; ptr = ptr2) {
		while (*ptr2 && *ptr2 != ';')
			ptr2++;
		if (*ptr2)
			*ptr2++ = '\0';
		while (*ptr2 == ';')
			ptr2++;
		if (exactMatch ? !strcasecmp(name, ptr) : strings.Prefix(name, ptr)) && len(db.Fetch(exit).sp.exit.dest) > 0 && TYPEOF(db.Fetch(exit).sp.exit.dest[0]) == TYPE_PROGRAM {
			return true
		}
	}
	return false
}

func ExitMatchExists(player, obj dbref, name string, exactMatch bool) (r bool) {
	for exit := db.Fetch(obj).exits; exit != NOTHING; exit = db.Fetch(exit).next {
		if exit_matches_name(exit, name, exactMatch) {
			notify(player, fmt.Sprintf("  %ss are trapped on %.2048s", name, unparse_object(player, obj)))
			r = true
			break
		}
	}
	return false
}

func do_sweep(descr int, player dbref, name string) {
	var thing dbref
	if name == "" {
		thing = db.Fetch(player).location
	} else {
		md := NewMatch(descr, player, name, NOTYPE).
			MatchMe().
			MatchHere().
			MatchAllExits().
			MatchNeighbor().
			MatchPossession().
			MatchRegistered()
		if Wizard(db.Fetch(player).owner) {
			md.MatchAbsolute().MatchPlayer()
		}
		thing = md.NoisyMatchResult()
	}
	switch {
	case thing == NOTHING:
		notify(player, "I don't know what object you mean.")
	case name != "" && !controls(db.Fetch(player).owner, thing):
		notify(player, "Permission denied. (You can't perform a security sweep in a room you don't own)")
	default:
		buf := fmt.Sprintf("Listeners in %s:", unparse_object(player, thing))
		notify(player, buf)
		for ref := db.Fetch(thing).contents; ref != NOTHING; ref = db.Fetch(ref).next {
			switch Typeof(ref) {
			case TYPE_PLAYER:
				if !Dark(thing) || online(ref) {
					if online(ref) {
						notify(player, fmt.Sprintf("  %s is a player.", unparse_object(player, ref)))
					} else {
						notify(player, fmt.Sprintf("  %s is a sleeping player"))
					}
				}
			case TYPE_THING:
				if db.Fetch(ref).flags & (ZOMBIE | LISTENER) != 0 {
					var tellflag bool
					buf := fmt.Sprintf("  %.255s is a", unparse_object(player, ref));
					if db.Fetch(ref).flags & ZOMBIE != 0 {
						tellflag = true
						if !online(db.Fetch(ref).owner) {
							tellflag = false
							buf += " sleeping"
						}
						buf += " zombie"
					}
					if db.Fetch(ref).flags & LISTENER != 0 && (get_property(ref, "_listen") || get_property(ref, "~listen") || get_property(ref, "~olisten")) {
						buf += " listener"
						tellflag = tell
					}
					buf += " object owned by "
					buf += unparse_object(player, db.Fetch(ref).owner)
					buf += "."
					if tellflag {
						notify(player, buf)
					}
				}
				ExitMatchExists(player, ref, "page", false)
				ExitMatchExists(player, ref, "whisper", false)
				if !ExitMatchExists(player, ref, "pose", true) && !ExitMatchExists(player, ref, "pos", true) {
					ExitMatchExists(player, ref, "po", true)
				}
				ExitMatchExists(player, ref, "say", false)
			}
		}
		var flag bool
		for loc := thing; loc != NOTHING; loc = getparent(loc) {
			if !flag {
				notify(player, "Listening rooms down the environment:")
				flag = true
			}
			if db.Fetch(loc).flags & LISTENER != 0 && (get_property(loc, "_listen") || get_property(loc, "~listen") || get_property(loc, "~olisten")) {
				notify(player, fmt.Sprintf("  %s is a listening room.", unparse_object(player, loc)))
			}
			ExitMatchExists(player, loc, "page", false)
			ExitMatchExists(player, loc, "whisper", false)
			if !ExitMatchExists(player, loc, "pose", true) && !ExitMatchExists(player, loc, "pos", true) {
				ExitMatchExists(player, loc, "po", true)
			}
			ExitMatchExists(player, loc, "say", false)
		}
		notify(player, "**End of list**")
	}
}