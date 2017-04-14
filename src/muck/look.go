package fbmuck

#define EXEC_SIGNAL '@'			/* Symbol which tells us what we're looking at is an execution order and not a message.    */

/* prints owner of something */
func print_owner(player, thing ObjectID) {
	var buf string
	switch Typeof(thing) {
	case TYPE_PLAYER:
		buf = fmt.Sprintf("%s is a player.", DB.Fetch(thing).name)
	case TYPE_ROOM, TYPE_THING, TYPE_EXIT, TYPE_PROGRAM:
		buf = fmt.Sprintf("Owner: %s", DB.Fetch(DB.Fetch(thing).Owner).name)
	}
	notify(player, buf)
}

void
exec_or_notify_prop(int descr, ObjectID player, ObjectID thing,
					const char *propname, const char *whatcalled)
{
	const char *message = get_property_class(thing, propname);
	int mpiflags = Prop_Blessed(thing, propname)? MPI_ISBLESSED : 0;

	if (message)
		exec_or_notify(descr, player, thing, message, whatcalled, mpiflags);
}

func exec_or_notify(descr int, player, thing ObjectID, message, whatcalled string, mpiflags int) {
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
		if !i.IsValid() || !IsProgram(i) {
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
			if tmpfr := interp(descr, player, DB.Fetch(player).Location, i, thing, PREEMPT, STD_HARDUID, 0); tmpfr != nil {
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
look_contents(ObjectID player, ObjectID loc, const char *contents_name)
{
	ObjectID thing;
	ObjectID can_see_loc;

	/* check to see if he can see the location */
	can_see_loc = (!Dark(loc) || controls(player, loc));

	/* check to see if there is anything there */
	for thing = DB.Fetch(loc).Contents; thing != NOTHING; thing = DB.Fetch(thing).next {
		if (can_see(player, thing, can_see_loc)) {
			/* something exists!  show him everything */
			notify(player, contents_name);
			for thing = DB.Fetch(loc).Contents; thing != NOTHING; thing = DB.Fetch(thing).next {
				if (can_see(player, thing, can_see_loc)) {
					notify(player, unparse_object(player, thing));
				}
			}
			break;				/* we're done */
		}
	}
}

func look_simple(descr int, player, thing ObjectID) {
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

func look_room(descr int, player, loc ObjectID) {
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

func do_look_around(descr int, player ObjectID) {
	if loc = DB.Fetch(player).Location; loc != NOTHING {
		look_room(descr, player, loc)
	}
}

func do_look_at(int descr, ObjectID player, const char *name, const char *detail) {
	if name == "" || name == "here" {
		if thing := DB.Fetch(player).Location; thing != NOTHING {
			look_room(descr, player, thing)
		}
	} else {
		md := NewMatch(descr, player, name, NOTYPE).
			MatchAllExits().
			MatchNeighbor().
			MatchPossession()
		if Wizard(DB.Fetch(player).Owner) {
			md.MatchAbsolute().MatchPlayer()
		}
		switch thing := md.MatchHere().MatchMe().Matchresult(); {
		case thing != NOTHING && thing != AMBIGUOUS && detail == "":
			switch TYPEOF(thing) {
			case TYPE_ROOM:
				if DB.Fetch(player).Location != thing && !can_link_to(player, TYPE_ROOM, thing) {
					notify(player, "Permission denied. (you're not where you want to look, and can't link to it)")
				} else {
					look_room(descr, player, thing)
				}
			case TYPE_PLAYER:
				if DB.Fetch(player).Location != DB.Fetch(thing).Location && !controls(player, thing) {
					notify(player, "Permission denied. (Your location isn't the same as what you're looking at)")
				} else {
					look_simple(descr, player, thing)
					look_contents(player, thing, "Carrying:")
					if tp_look_propqueues {
						envpropqueue(descr, player, thing, player, thing, NOTHING, "_lookq", fmt.Sprintf("#%d", thing), 1, 1)
					}
				}
			case TYPE_THING:
				if DB.Fetch(player).Location != DB.Fetch(thing).Location && DB.Fetch(thing).Location != player && !controls(player, thing) {
					notify(player, "Permission denied. (You're not in the same room as or carrying the object)")
				} else {
					look_simple(descr, player, thing)
					if DB.Fetch(thing).flags & HAVEN == 0 {
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
				thing = DB.Fetch(player).Location
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
func flag_description(thing ObjectID) (r string) {
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

	if DB.Fetch(thing).flags & ~TYPE_MASK != 0 {
		r += "  Flags:"
		if DB.Fetch(thing).flags & WIZARD != 0 {
			r += " WIZARD"
		}
		if DB.Fetch(thing).flags & QUELL != 0 {
			r += " QUELL"
		}
		if DB.Fetch(thing).flags & STICKY != 0 {
			switch Typeof(thing) {
			case TYPE_PROGRAM:
				r += " SETUID"
			case TYPE_PLAYER:
				r += " SILENT"
			default:
				r += " STICKY"
			}
		}
		if DB.Fetch(thing).flags & DARK != 0 {
			if Typeof(thing) == TYPE_PROGRAM {
				r += " DEBUGGING"
			} else {
				r += " DARK"
			}
		}
		if DB.Fetch(thing).flags & LINK_OK != 0 {
			r += " LINK_OK"
		}
		if DB.Fetch(thing).flags & KILL_OK != 0 {
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
		if DB.Fetch(thing).flags & BUILDER != 0 {
			if Typeof(thing) == TYPE_PROGRAM {
				r += " BOUND"
			} else {
				r += " BUILDER"
			}
		}
		if DB.Fetch(thing).flags & CHOWN_OK != 0 {
			if Typeof(thing) == TYPE_PLAYER {
				r += " COLOR"
			} else {
				r += " CHOWN_OK"
			}
		}
		if DB.Fetch(thing).flags & JUMP_OK != 0 {
			r += " JUMP_OK"
		}
		if DB.Fetch(thing).flags & VEHICLE != 0 {
			if Typeof(thing) == TYPE_PROGRAM {
				r += " VIEWABLE"
			} else {
				r += " VEHICLE"
			}
		}
		if tp_enable_match_yield && DB.Fetch(thing).flags & YIELD != 0 {
			r += " YIELD"
		}
		if tp_enable_match_yield && DB.Fetch(thing).flags & OVERT != 0 {
			r += " OVERT"
		}
		if DB.Fetch(thing).flags & XFORCIBLE != 0 {
			if Typeof(thing) == TYPE_EXIT {
				r += " XPRESS"
			} else {
				r += " XFORCIBLE"
			}
		}
		if DB.Fetch(thing).flags & ZOMBIE != 0 {
			r += " ZOMBIE"
		}
		if DB.Fetch(thing).flags & HAVEN != 0 {
			switch Typeof(thing) {
			case TYPE_PROGRAM:
				r += " HARDUID"
			case TYPE_THING:
				r += " HIDE"
			default:
				r += " HAVEN"
			}
		}
		if DB.Fetch(thing).flags & ABODE != 0 {
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

func listprops_wildcard(ObjectID player, ObjectID thing, const char *dir, const char *wild) int {
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
			if (!Prop_System(buf) && ((!Prop_Hidden(buf) && !(PropFlags(propadr) & PROP_SYSPERMS)) || Wizard(DB.Fetch(player).Owner))) {
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

func do_examine(descr int, player ObjectID, name, dir string) {
	var thing, content, exit ObjectID
	int i, cnt;

	if name == "" {
		if thing = DB.Fetch(player).Location; thing == NOTHING {
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
		if Wizard(DB.Fetch(player).Owner) {
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
			buf = fmt.Sprintf("%s  Owner: %s  Parent: %s", unparse_object(player, thing), DB.Fetch(DB.Fetch(thing).Owner).name, unparse_object(player, DB.Fetch(thing).Location)
		case TYPE_THING:
			buf = fmt.Sprintf("%s  Owner: %s  Value: %d", unparse_object(player, thing), DB.Fetch(DB.Fetch(thing).Owner).name, get_property_value(thing, MESGPROP_VALUE))
		case TYPE_PLAYER:
			buf = fmt.Sprintf("%s  %s: %d  ", unparse_object(player, thing), tp_cpennies, get_property_value(thing, MESGPROP_VALUE))
		case TYPE_EXIT, TYPE_PROGRAM:
			buf = fmt.Sprintf("%s  Owner: %s", unparse_object(player, thing), DB.Fetch(DB.Fetch(thing).Owner).name)
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
		/* ex: time_tm = localtime((time_t *)(&(DB.Fetch(thing).Created))); */
		time_tm := localtime((&(DB.Fetch(thing).Created)))
		notify(player, format_time((char *) "Created:  %a %b %e %T %Z %Y", time_tm))
		time_tm = localtime((&(DB.Fetch(thing).Modified)))
		notify(player, format_time((char *) "Modified: %a %b %e %T %Z %Y", time_tm))
		time_tm = localtime((&(DB.Fetch(thing).LastUsed)))
		notify(player, format_time((char *) "Lastused: %a %b %e %T %Z %Y", time_tm))
		if TYPEOF(thing) == TYPE_PROGRAM {
			var i int
			if DB.Fetch(thing).(Program) != nil {
				i = DB.Fetch(thing).(Program).instances
			}
			notify(player, fmt.Sprintf("Usecount: %d     Instances: %d", DB.Fetch(thing).Uses, i))
		} else {
			notify(player, fmt.Sprintf("Usecount: %d", DB.Fetch(thing).Uses))
		}

		notify(player, "[ Use 'examine <object>=/' to list root properties. ]")

		/* show him the contents */
		if DB.Fetch(thing).Contents != NOTHING {
			if TYPEOF(thing) == TYPE_PLAYER {
				notify(player, "Carrying:");
			} else {
				notify(player, "Contents:")
			}
			for content = DB.Fetch(thing).Contents; content != NOTHING; content = DB.Fetch(content).next {
				notify(player, unparse_object(player, content))
			}
		}
		switch o := DB.Fetch(thing).(type) {
		case Room:
			/* tell him about exits */
			if o.Exits != NOTHING {
				notify(player, "Exits:")
				for exit = o.Exits; exit != NOTHING; exit = o.next {
					notify(player, unparse_object(player, exit))
				}
			} else {
				notify(player, "No exits.");
			}

			if o.ObjectID != NOTHING {
				notify(player, fmt.Sprintf("Dropped objects go to: %s", unparse_object(player, o.(ObjectID))))
			}
		case Object:
			/* print home */
			notify(player, fmt.Sprintf("Home: %s", unparse_object(player, o.home)))	/* home */
			/* print location if player can link to it */
			if o.Location != NOTHING && (controls(player, o.Location) || can_link_to(player, NOTYPE, o.Location)) {
				notify(player, fmt.Sprintf("Location: %s", unparse_object(player, o.Location)))
			}
			/* print thing's actions, if any */
			if o.Exits != NOTHING {
				notify(player, "Actions/exits:")
				for exit = o.Exits; exit != NOTHING; exit = o.next {
					notify(player, unparse_object(player, exit))
				}
			} else {
				notify(player, "No actions attached.")
			}
		case Player:
			/* print home */
			notify(player, fmt.Sprintf("Home: %s", unparse_object(player, o.home)))

			/* print location if player can link to it */
			if o.Location != NOTHING && (controls(player, o.Location) || can_link_to(player, NOTYPE, o.Location)) {
				notify(player, fmt.Sprintf("Location: %s", unparse_object(player, o.Location)))
			}
			/* print player's actions, if any */
			if o.Exits != NOTHING {
				notify(player, "Actions/exits:")
				for exit = DB.Fetch(o.Exits); exit != NOTHING; exit = DB.Fetch(exit).next {
					notify(player, unparse_object(player, exit))
				}
			} else {
				notify(player, "No actions attached.")
			}
		case Exit:
			if o.Location != NOTHING {
				notify(player, fmt.Sprintf("Source: %s", unparse_object(player, o.Location)))
			}
			/* print destinations */
			for _, v := range o.Destinations {
				switch v {
				case NOTHING:
				case HOME:
					notify(player, "Destination: *HOME*");
				default:
					notify(player, fmt.Sprintf("Destination: %s", unparse_object(player, v)))
				}
			}
		case Program:
			if len(o.code) > 0 {
				notify(player, fmt.Sprintf("Program compiled size: %d instructions", len(o.code)))
				notify(player, fmt.Sprintf("Cumulative runtime: %v seconds ", o.proftime))
			} else {
				notify(player, fmt.Sprintf("Program not compiled."))
			}
			if loc := o.Location; loc != NOTHING && (controls(player, loc) || can_link_to(player, NOTYPE, loc)) {
				notify(player, fmt.Sprintf("Location: %s", unparse_object(player, loc)))
			}
		}
	}
}

func do_score(player ObjectID) {
	if v := get_property_value(player, MESGPROP_VALUE); v == 1 {
		notify(player, fmt.Sprintf("You have %d %s.", v, tp_penny))
	} else {
		notify(player, fmt.Sprintf("You have %d %s.", v, tp_pennies))
	}
}

func do_inventory(player ObjectID) {
	ObjectID thing;

	if thing := DB.Fetch(player).Contents; thing == NOTHING {
		notify(player, "You aren't carrying anything.")
	} else {
		notify(player, "You are carrying:")
		for ; thing != NOTHING; thing = DB.Fetch(thing).next {
			notify(player, unparse_object(player, thing))
		}
	}
	do_score(player)
}

func init_checkflags(player ObjectID, flags string) (output_type int, check *flgchkdat) {
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


func checkflags(ObjectID what, struct flgchkdat check) (r bool) {
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
	case DB.Fetch(what).flags & check.clearflags != 0:
		fallthrough
	case ~DB.Fetch(what).flags & check.setflags != 0:
		r = false
	}

	if check.forlink {
		switch Typeof(what) {
		case TYPE_ROOM:
			if (DB.Fetch(what).sp == NOTHING) != !check.islinked {
				r = false
			}
		case TYPE_EXIT:
			if (len(DB.Fetch(what).(Exit).Destinations) == 0) != !check.islinked {
				r = false
			}
		case TYPE_PLAYER, TYPE_THING:
			r = check.islinked
		default:
			r = !check.islinked
		}
	}

	if check.forold {
		if (((time(nil)) - DB.Fetch(what).LastUsed) < tp_aging_time) || (((time(nil)) - DB.Fetch(what),modified) < tp_aging_time) != !check.isold {
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

func display_objinfo(player, obj ObjectID, output_type int) {
	buf2 := unparse_object(player, obj)
	switch o := DB.Fetch(obj); output_type {
	case 0:
	case CHECK_OWNERS:
		notify(player, fmt.Sprintf("%-38.512s  %.512s", buf2, unparse_object(player, o.Owner)))
	case CHECK_LINKS:
		switch o := o.(type) {
		case Room:
			notify(player, fmt.Sprintf("%-38.512s  %.512s", buf2, unparse_object(player, o.ObjectID)))
		case Exit:
			switch n := len(o.Destinations); {
			case n == 0:
				notify(player, fmt.Sprintf("%-38.512s  %.512s", buf2, "*UNLINKED*"))
			case n > 1:
				notify(player, fmt.Sprintf("%-38.512s  %.512s", buf2, "*METALINKED*"))
			default:
				notify(player, fmt.Sprintf("%-38.512s  %.512s", buf2, unparse_object(player, o.Destinations[0])))
			}
		case Player:
			notify(player, fmt.Sprintf("%-38.512s  %.512s", buf2, unparse_object(player, o.home)))
		case Object:
			notify(player, fmt.Sprintf("%-38.512s  %.512s", buf2, unparse_object(player, o.home)))
		default:
			notify(player, fmt.Sprintf("%-38.512s  %.512s", buf2, "N/A"))
		}
	case CHECK_LOCATIONS:
		notify(player, fmt.Sprintf("%-38.512s  %.512s", buf2, unparse_object(player, o.Location)))
	case 4:
		return
	default:
		notify(player, buf2)
	}
	notify(player, buf)
}

func do_find(player ObjectID, name, flags string) {
	if !payfor(player, tp_lookup_cost) {
		notify_fmt(player, "You don't have enough %s.", tp_pennies)
	} else {
		var total int
		buf := "*" + name + "*"
		output_type, check := init_checkflags(player, flags)
		EachObject(func(obj ObjectID, o *Object) {
			if (Wizard(DB.Fetch(player).Owner) || o.Owner == DB.Fetch(player).Owner) && checkflags(obj, check) && o.name != "" && (name == "" || !smatch(buf, o.name) {
				display_objinfo(player, obj, output_type)
				total++
			}
		})
		notify(player, "***End of List***")
		notify_fmt(player, "%d objects found.", total)
	}
}

func do_owned(ObjectID player, const char *name, const char *flags) {
	if !payfor(player, tp_lookup_cost) {
		notify_fmt(player, "You don't have enough %s.", tp_pennies)
	} else {
		var victim ObjectID
		output_type, check := init_checkflags(player, flags)
		if Wizard(DB.Fetch(player).Owner) && name != "" {
			if victim = lookup_player(name); victim == NOTHING {
				notify(player, "I couldn't find that player.")
				return
			}
		} else {
			victim = player
		}

		var total int
		EachObject(func(obj ObjectID, o *Object) {
			if o.Owner == DB.Fetch(victim).Owner && checkflags(obj, check) {
				display_objinfo(player, obj, output_type)
				total++
			}
		})
		notify(player, "***End of List***")
		notify_fmt(player, "%d objects found.", total)
	}
}

func do_trace(int descr, ObjectID player, const char *name, int depth) {
	ObjectID thing;
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
		thing = DB.Fetch(thing).Location
	}
	notify(player, "***End of List***");
}

func do_entrances(int descr, ObjectID player, const char *name, const char *flags) {
	var thing ObjectID
	if name == "" {
		thing = DB.Fetch(player).Location
	} else {
		md := NewMatch(descr, player, name, NOTYPE).
			MatchAllExits().
			MatchNeighbor().
			MatchPossession().
			MatchRegistered()
		if Wizard(DB.Fetch(player).Owner) {
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
	case !controls(DB.Fetch(player).Owner, thing):
		notify(player, "Permission denied. (You can't list entrances of objects you don't control)")
	default:
		_, check = init_checkflags(player, flags)
		EachObject(func(obj ObjectID, o *Object) {
			if checkflags(obj, check) {
				switch o := o.(type) {
				case Exit:
					for j := len(o.Destinations) - 1; j > 0; j-- {
						if o.Destinations[j] == thing {
							display_objinfo(player, obj, output_type)
							total++
						}
					}
				case Player:
					if o.home == thing {
						display_objinfo(player, obj, output_type)
						total++
					}
				case Object:
					if o.home == thing {
						display_objinfo(player, obj, output_type)
						total++
					}
				case Room:
					if o == thing {
						display_objinfo(player, obj, output_type)
						total++
					}
				}
			}
		})
		notify(player, "***End of List***")
		notify_fmt(player, "%d objects found.", total)
	}
}

func do_contents(int descr, ObjectID player, const char *name, const char *flags) {
	var total int
	var i, thing ObjectID
	if name == "" {
		thing = DB.Fetch(player).Location
	} else {
		md := NewMatch(descr, player, name, NOTYPE).
			MatchMe().
			MatchHere().
			MatchAllExits().
			MatchNeighbor().
			MatchPossession().
			MatchRegistered()
		if Wizard(DB.Fetch(player).Owner) {
			md.MatchAbsolute().MatchPlayer()
		}
		thing = md.NoisyMatchResult()
	}
	if thing != NOTHING {
		if !controls(DB.Fetch(player).Owner, thing) {
			notify(player, "Permission denied. (You can't get the contents of something you don't control)")
		} else {
			output_type, check := init_checkflags(player, flags)
			for i := DB.Fetch(thing).Contents; i != NOTHING; i = DB.Fetch(i).next {
				if checkflags(i, check) {
					display_objinfo(player, i, output_type)
					total++
				}
			}
			switch TYPEOF(thing) {
			case TYPE_ROOM, TYPE_THING, TYPE_PLAYER:
				for i := DB.Fetch(thing).Exits; i != NOTHING; i = DB.Fetch(i).next {
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

func exit_matches_name(exit ObjectID, name string, exactMatch bool) bool {
	char buf[BUFFER_LEN];
	char *ptr, *ptr2;

	strcpyn(buf, sizeof(buf), DB.Fetch(exit).name)
	for (ptr2 = ptr = buf; *ptr; ptr = ptr2) {
		while (*ptr2 && *ptr2 != ';')
			ptr2++;
		if (*ptr2)
			*ptr2++ = '\0';
		while (*ptr2 == ';')
			ptr2++;
		if (exactMatch ? !strcasecmp(name, ptr) : strings.Prefix(name, ptr)) && len(DB.Fetch(exit).(Exit).Destinations) > 0 && TYPEOF(DB.Fetch(exit).(Exit).Destinations[0]) == TYPE_PROGRAM {
			return true
		}
	}
	return false
}

func ExitMatchExists(player, obj ObjectID, name string, exactMatch bool) (r bool) {
	for exit := DB.Fetch(obj).Exits; exit != NOTHING; exit = DB.Fetch(exit).next {
		if exit_matches_name(exit, name, exactMatch) {
			notify(player, fmt.Sprintf("  %ss are trapped on %.2048s", name, unparse_object(player, obj)))
			r = true
			break
		}
	}
	return false
}

func do_sweep(descr int, player ObjectID, name string) {
	var thing ObjectID
	if name == "" {
		thing = DB.Fetch(player).Location
	} else {
		md := NewMatch(descr, player, name, NOTYPE).
			MatchMe().
			MatchHere().
			MatchAllExits().
			MatchNeighbor().
			MatchPossession().
			MatchRegistered()
		if Wizard(DB.Fetch(player).Owner) {
			md.MatchAbsolute().MatchPlayer()
		}
		thing = md.NoisyMatchResult()
	}
	switch {
	case thing == NOTHING:
		notify(player, "I don't know what object you mean.")
	case name != "" && !controls(DB.Fetch(player).Owner, thing):
		notify(player, "Permission denied. (You can't perform a security sweep in a room you don't own)")
	default:
		buf := fmt.Sprintf("Listeners in %s:", unparse_object(player, thing))
		notify(player, buf)
		for ref := DB.Fetch(thing).Contents; ref != NOTHING; ref = DB.Fetch(ref).next {
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
				if DB.Fetch(ref).flags & (ZOMBIE | LISTENER) != 0 {
					var tellflag bool
					buf := fmt.Sprintf("  %.255s is a", unparse_object(player, ref));
					if DB.Fetch(ref).flags & ZOMBIE != 0 {
						tellflag = true
						if !online(DB.Fetch(ref).Owner) {
							tellflag = false
							buf += " sleeping"
						}
						buf += " zombie"
					}
					if DB.Fetch(ref).flags & LISTENER != 0 && (get_property(ref, "_listen") || get_property(ref, "~listen") || get_property(ref, "~olisten")) {
						buf += " listener"
						tellflag = tell
					}
					buf += " object owned by "
					buf += unparse_object(player, DB.Fetch(ref).Owner)
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
			if DB.Fetch(loc).flags & LISTENER != 0 && (get_property(loc, "_listen") || get_property(loc, "~listen") || get_property(loc, "~olisten")) {
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