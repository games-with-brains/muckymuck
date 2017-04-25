package fbmuck

func WizardPlayerOnly(p ObjectID, f func(*Player)) {
	if Wizard(player) {
		if p, ok := DB.Fetch(player).(*Player); ok {
			f(p)
			return
		}
	}
	notify(player, "Only a Wizard player can do that")
}

func do_teleport(descr int, player ObjectID, arg1, arg2 string) {
	var victim, destination ObjectID
	var to string

	/* get victim, destination */
	if arg2 == "" {
		victim = player
		to = arg1
	} else {
		victim = NewMatch(descr, player, arg1, NOTYPE).
			MatchNeighbor().
			MatchPossession().
			MatchMe().
			MatchHere().
			MatchAbsolute().
			MatchRegistered().
			MatchPlayer().
			NoisyMatchResult()

		if victim == NOTHING {
			return
		}
		to = arg2
	}
	if player != GOD && DB.Fetch(victim).Owner == GOD {
		notify(player, "God has already set that where He wants it to be.")
		return
	}

	/* get destination */
	md := NewMatch(descr, player, to, IsPlayer).
		MatchPossession().
		MatchMe().
		MatchHere().
		MatchHome().
		MatchAbsolute().
		MatchRegistered()
	if Wizard(DB.Fetch(player).Owner) {
		md.MatchNeighbor().MatchPlayer()
	}
	switch destination = md.MatchResult(); desination {
	case NOTHING:
		notify(player, "Send it where?")
	case AMBIGUOUS:
		notify(player, "I don't know which destination you mean!")
	case HOME:
		v := DB.FetchPlayer(victim)
		switch victim.(type) {
		case Player:
			destination = v.Home
			if parent_loop_check(victim, destination) {
				destination = DB.Fetch(v.Owner).Home
			}
		case Object:
			destination = v.Home
			if parent_loop_check(victim, destination) {
				destination = DB.Fetch(v.Owner).Home
				if parent_loop_check(victim, destination) {
					destination = ObjectID(0)
				}
			}
		case Room:
			destination = GLOBAL_ENVIRONMENT
		case Program:
			destination = v.Owner
		default:
			destination = tp_player_start
		}
	default:
		switch {
		case IsPlayer(victim):
			v := DB.FetchPlayer(victim)
			switch {
			case !controls(player, victim), !controls(player, destination), !controls(player, v.Location), (IsThing(destination) && !controls(player, df.Fetch(destination).Location)):
				notify(player, "Permission denied. (must control victim, dest, victim's loc, and dest's loc)")
			case !IsRoom(destination) && !IsThing(destination):
				notify(player, "Bad destination.")
			case !Wizard(victim) && IsThing(destination) && DB.Fetch(destination).flags & VEHICLE == 0:
				notify(player, "Destination object is not a vehicle.")
			case parent_loop_check(victim, destination):
				notify(player, "Objects can't contain themselves.")
			default:
				notify(victim, "You feel a wrenching sensation...")
				enter_room(descr, victim, destination, v.Location)
				notify(player, "Teleported.")
			}
		case IsThing(victim):
			if parent_loop_check(victim, destination) {
				notify(player, "You can't make a container contain itself!")
				break
			}
			fallthrough
		case IsProgram(victim):
			switch {
			case !IsRoom(destination) && !IsPlayer(destination) && !IsThing(destination):
				notify(player, "Bad destination.")
			case !((controls(player, destination) || can_link_to(player, NOTYPE, destination)) && (controls(player, victim) || controls(player, DB.Fetch(victim).Location))):
				notify(player, "Permission denied. (must control dest and be able to link to it, or control dest's loc)")
			default:
				/* check for non-sticky dropto */
				if IsRoom(destination) && DB.Fetch(destination).sp != NOTHING && DB.Fetch(destination).flags & STICKY == 0 {
					destination = DB.Fetch(destination).(ObjectID)
				}
				if tp_thing_movement && IsThing(victim) {
					enter_room(descr, victim, destination, DB.Fetch(victim).Location)
				} else {
					moveto(victim, destination)
				}
				notify(player, "Teleported.")
			}
		case IsRoom(victim):
			switch {
			case !IsRoom(destination):
				notify(player, "Bad destination.")
			case !controls(player, victim), !can_link_to(player, NOTYPE, destination), victim == GLOBAL_ENVIRONMENT:
				notify(player, "Permission denied. (Can't move #0, dest must be linkable, and must control victim)")
			case parent_loop_check(victim, destination):
				notify(player, "Parent would create a loop.")
			default:
				moveto(victim, destination)
				notify(player, "Parent set.")
			}
		default:
			notify(player, "You can't teleport that.")
		}
	}
	return
}

int blessprops_wildcard(ObjectID player, ObjectID thing, const char *dir, const char *wild, int blessp) {
	var propname string
	var wld []byte
	var buf []byte
	var buf2 []byte
	ptr, wldcrd = wld
	pptr *Plist
	var i, cnt int
	var recurse int

	if player != GOD && DB.Fetch(thing).Owner == GOD {
		notify(player,"Only God may touch what is God's.");
		return 0;
	}

	strcpyn(wld, sizeof(wld), wild);
	i = len(wld);
	if (i && wld[i - 1] == PROPDIR_DELIMITER)
		strcatn(wld, sizeof(wld), "*");
	for wldcrd = wld; *wldcrd == PROPDIR_DELIMITER; wldcrd++ {}
	if wldcrd != "**" {
		recurse = 1
	}

	for (ptr = wldcrd; *ptr && *ptr != PROPDIR_DELIMITER; ptr++) ;
	if (*ptr)
		*ptr++ = '\0';

	propadr := pptr.first_prop(thing, dir, propname)
	for propadr != nil {
		if !smatch(wldcrd, propname) {
			buf = fmt.Sprint(dir, PROPDIR_DELIMITER, propname)
			if (!Prop_System(buf) && ((!Prop_Hidden(buf) && !(PropFlags(propadr) & PROP_SYSPERMS)) || Wizard(DB.Fetch(player).Owner))) {
				if (!*ptr || recurse) {
					cnt++;
					if (blessp) {
						set_property_flags(thing, buf, PROP_BLESSED);
						buf2 = fmt.Sprintf("Blessed %s", buf)
					} else {
						clear_property_flags(thing, buf, PROP_BLESSED);
						buf2 = fmt.Sprintf("Unblessed %s", buf)
					}
					notify(player, buf2);
				}
				if (recurse)
					ptr = "**";
				cnt += blessprops_wildcard(player, thing, buf, ptr, blessp);
			}
		}
		propadr, propname = propadr.next_prop(pptr)
	}
	return cnt;
}

func do_unbless(descr int, player ObjectID, what, propname string) {
	WizardPlayerOnly(player, func(p *Player) {
		if propname == "" {
			notify(player, "Usage is @unbless object=propname.")
		} else {
			md := NewMatch(descr, player, what, NOTYPE)
			md.MatchEverything()
			switch victim = md.NoisyMatchResult(); {
			case victim == NOTHING:
			case !Wizard(p.Owner):
				notify(player, "Permission denied. (You're not a wizard)")
			default:
				if cnt := blessprops_wildcard(player, victim, "", propname, 0); cnt == 1 {
					notify(player, fmt.Sprintf("%d property unblessed.", cnt))
				} else {
					notify(player, fmt.Sprintf("%d properties unblessed.", cnt))
				}
			}
		}
	})
}

func do_bless(descr int, player ObjectID, what, propname string) {
	WizardPlayerOnly(player, func(p *Player) {
		switch {
		case force_level:
			notify(player, "Can't @force an @bless.")
		case propname == "":
			notify(player, "Usage is @bless object=propname.")
		default:
			/* get victim */
			md := NewMatch(descr, player, what, NOTYPE)
			md.MatchEverything()
			switch victim = md.NoisyMatchResult(); {
			case victim == NOTHING:
			case player != GOD && DB.Fetch(victim).Owner == GOD:
				notify(player, "Only God may touch God's stuff.")
			case !Wizard(p.Owner):
				notify(player, "Permission denied. (you're not a wizard)")
			default:
				if cnt := blessprops_wildcard(player, victim, "", propname, 1); cnt == 1 {
					notify(player, fmt.Sprintf("%d property blessed.", cnt))
				} else {
					notify(player, fmt.Sprintf("%d properties blessed.", cnt))
				}
			}
		}
	})
}

func do_force(descr int, player ObjectID, what, command string) {
	switch {
	case force_level > tp_max_force_level - 1:
		notify(player, "Can't force recursively.")
		return
	case !tp_zombies && (!Wizard(player) || !IsPlayer(player)):
		notify(player, "Zombies are not enabled here.")
		return
	}

	/* get victim */
	victim := NewMatch(descr, player, what, NOTYPE).
		MatchNeighbor().
		MatchPossession().
		MatchMe().
		MatchHere().
		MatchAbsolute().
		MatchRegistered().
		MatchPlayer().
		NoisyMatchResult()
	p := DB.FetchPlayer(player)
	v := DB.FetchPlayer(victim)
	terms := strings.SplitN(v.name, " ", 2)
	switch {
	case victim == NOTHING:
	case !IsPlayer(victim) && !IsThing(victim):
		notify(player, "Permission Denied -- Target not a player or thing.")
	case victim == GOD:
		notify(player, "You cannot force God to do anything.")
	case !Wizard(player) && v.flags & XFORCIBLE == 0:
		notify(player, "Permission denied: forced object not @set Xforcible.")
	case !Wizard(player) && !test_lock_false_default(descr, player, victim, "@/flk"):
		notify(player, "Permission denied: Object not force-locked to you.")
	case !Wizard(player) && IsThing(victim) && v.Location != NOTHING && DB.Fetch(v.Location).flags & ZOMBIE != 0 && IsRoom(v.Location):
		notify(player, "Sorry, but that's in a no-puppet zone.")
	case !Wizard(p.Owner) && IsThing(victim) && p.flags & ZOMBIE != 0:
		notify(player, "Permission denied -- you cannot use zombies.")
	case !Wizard(p.Owner) && IsThing(victim) && p.flags & DARK != 0:
		notify(player, "Permission denied -- you cannot force dark zombies.")
	case !Wizard(p.Owner) && IsThing(victim) && terms > 0 && lookup_player(terms[0]) != NOTHING:
		notify(player, "Puppet cannot share the name of a player.")
	default:
		log_status("FORCED: %s(%d) by %s(%d): %s", v.Name(), victim, p.Name(), player, command)
		/* force victim to do command */
		ForceAction(NOTHING, func() {
			process_command(ObjectID_first_descr(victim), victim, command)
		})
	}
}

func do_stats(player ObjectID, name string) {
	var rooms, exits, things, players, programs, garbage, total, altered, oldobjs int
	currtime := time.Now()
	owner := NOTHING

	if !Wizard(DB.FetchPlayer(player).Owner) && len(name) == 0 {
		notify(player, fmt.Sprintf("The universe contains %d objects.", db_top))
	} else {
		total = rooms = exits = things = players = programs = 0;
		if len(name) > 0 {
			owner = lookup_player(name)
			if owner == NOTHING {
				notify(player, "I can't find that player.")
				return
			}
			if p := DB.FetchPlayer(player); !Wizard(p.Owner) && p.Owner != owner {
				notify(player, "Permission denied. (you must be a wizard to get someone else's stats)")
				return
			}
			EachObject(func(obj ObjectID, o *Object) {
				if o.Owner == owner {
					if o.Changed {
						altered++
					}
					/* if unused for 90 days, inc oldobj count */
					if (currtime - o.LastUsed) > tp_aging_time {
						oldobjs++
					}

					switch {
					case IsRoom(obj):
						total++
						rooms++
					case IsExit(obj):
						total++
						exits++
					case IsThing(obj):
						total++
						things++
					case IsPlayer(obj):
						total++
						players++
					case IsProgram(obj):
						total++
						programs++
					}
				}
			})
		} else {
			EachObject(func(obj ObjectID, o *Object) {
				if o.Changed {
					altered++
				}
				/* if unused for 90 days, inc oldobj count */
				if (currtime - o.LastUsed) > tp_aging_time {
					oldobjs++
				}

				switch {
				case IsRoom(obj):
					total++
					rooms++
				case IsExit(obj):
					total++
					exits++
				case IsThing(obj):
					total++
					things++
				case IsPlayer(obj):
					total++
					players++
				case IsProgram(obj):
					total++
					programs++
				}
			})
		}
		notify_fmt(player, "%7d room%s        %7d exit%s        %7d thing%s", rooms, (rooms == 1) ? " " : "s", exits, (exits == 1) ? " " : "s", things, (things == 1) ? " " : "s")
		notify_fmt(player, "%7d program%s     %7d player%s      %7d garbage", programs, (programs == 1) ? " " : "s", players, (players == 1) ? " " : "s", garbage)
		notify_fmt(player, "%7d total object%s                     %7d old & unused", total, (total == 1) ? " " : "s", oldobjs)

		struct tm *time_tm
		time_t lasttime = (time_t) get_property_value(0, "_sys/lastdumptime")

		time_tm = localtime(&lasttime)

		if altered == 1 {
			notify_fmt(player, "%7d unsaved object%s     Last dump: %s", altered, "", format_time("%a %b %e %T %Z", time_tm))
		} else {
			notify_fmt(player, "%7d unsaved object%s     Last dump: %s", altered, "s", format_time("%a %b %e %T %Z", time_tm))
		}
	}
}

func do_boot(player ObjectID, name string) {
	WizardPlayerOnly(player, func(p *Player) {
		switch victim := lookup_player(name); victim {
		case NOTHING:
			notify(player, "That player does not exist.")
		case GOD:
			notify(player, "You can't boot God!")
		default:
			switch v := DB.Fetch(victim); v.(type) {
			case Player:
				notify(victim, "You have been booted off the game.")
				if boot_off(victim) {
					log_status("BOOTED: %s(%d) by %s(%d)", v.Name(), victim, p.Name(), player)
					if player != victim {
						notify(player, fmt.Sprintf("You booted %s off!", v.Name()))
					}
				} else {
					notify(player, fmt.Sprintf("%s is not connected.", v.Name()))
				}
			default:
				notify(player, "You can only boot players!")
			}
		}
	})
}

func do_toad(descr int, player ObjectID, name, recip string) {
	WizardPlayerOnly(player, func(p *Player) {
		switch victim := lookup_player(name); victim {
		case NOTHING:
			notify(player, "That player does not exist.")
		case GOD:
			notify(player, "You cannot @toad God.")
			if player != GOD {
				log_status("TOAD ATTEMPT: %s(#%d) tried to toad God.", p.name, player)
			}
		case player:
			//	We don't want the last wizard to be toaded, in any case, so only someone else can do it.
			notify(player, "You cannot toad yourself.  Get someone else to do it for you.")
		default:
			var recipient ObjectID
			if recip == "" {
				/* FIXME: Make me a tunable parameter! */
				recipient = GOD
			} else {
				recipient = lookup_player(recip)
				if recipient == NOTHING || recipient == victim {
					notify(player, "That recipient does not exist.")
					return
				}
			}

			switch {
			case !IsPlayer(victim):
				notify(player, "You can only turn players into toads!")
			case player != GOD && TrueWizard(victim):
				notify(player, "You can't turn a Wizard into a toad.")
			default:
				send_contents(descr, victim, HOME)
				dequeue_prog(victim, 0)							//	Dequeue the programs that the player's running
				EachObject(func(obj ObjectID, o *Object) {
					if o.Owner == victim {
						switch o.(type) {
						case Program:
							dequeue_prog(obj, 0)				//	dequeue player's progs
							if TrueWizard(recipient) {
								o.flags &= ~(ABODE | WIZARD)
								SetMLevel(obj, APPRENTICE)
							}
						case Room, Object, Exit:
							o.GiveTo(recipient)
							o.Touch()
						}
					}
					if _, ok o.(Object); ok {
						if o.Home == victim {
							//	FIXME: Set a tunable "lost and found" area!
							p.LiveAt(tp_player_start)
						}
					}
				})

				v := DB.FetchPlayer(victim)
				notify(victim, "You have been turned into a toad.")
				notify(player, fmt.Sprintf("You turned %s into a toad!", v.name))
				log_status("TOADED: %s(%d) by %s(%d)", v.name, victim, p.name, player)

				/* reset name */
				delete_player(victim)
				boot_player_off(victim)
				ignore_remove_from_all_players(victim)
				ignore_flush_cache(victim)
				DB.Store(victim, &Object{
					name: fmt.Sprintf("A slimy toad named %s", v.name),
					home: p.Home(),
					owner: player,
				})
				v.Touch()
				add_property(victim, MESGPROP_VALUE, NULL, 1)		/* don't let him keep his immense wealth */
			}
		}
	})
}

func do_newpassword(player ObjectID, name, password string) {
	WizardPlayerOnly(player, func(p *Player) {
		switch victim := lookup_player(name); {
		case victim == NOTHING:
			notify(player, "No such player.")
		case password != "" && !ok_password(password):
			/* Wiz can set null passwords, but not bad passwords */
			notify(player, "Bad password")
		case victim == GOD:
			notify(player, "You can't change God's password!")
		case TrueWizard(victim) && player != GOD:
			notify(player, "Only God can change a wizard's password.")
		default:
			set_password(victim, password)
			DB.FetchPlayer(victim).Touch()
			notify(player, "Password changed.")
			notify(victim, fmt.Sprintf("Your password has been changed by %s.", p.Name()))
			log_status("NEWPASS'ED: %s(%d) by %s(%d)", DB.FetchPlayer(victim).name, victim, p.Name(), player)
		}
	}
}

func do_pcreate(player ObjectID, user, password string) {
	WizardPlayerOnly(player, func(p *Player) {
		if newguy := create_player(user, password); newguy == NOTHING {
			notify(player, "Create failed.")
		} else {
			log_status("PCREATED %s(%d) by %s(%d)", DB.FetchPlayer(newguy).name, newguy, p.Name(), player)
			notify(player, fmt.Sprintf("Player %s created as object #%d.", user, newguy))
		}
	})
}

func do_serverdebug(descr int, player ObjectID, arg1, arg2 string) {
	switch {
	case !Wizard(DB.FetchPlayer(player).Owner):
		notify(player, "Permission denied. (@dbginfo is a wizard-only command)")
	case arg1 == "":
		notify(player, "Usage: @dbginfo [cache|guitest|misc]")
	default:
		if strings.Prefix(arg1, "guitest") {
			do_post_dlog(descr, arg2)
		}
		notify(player, "Done.")
	}
}


long max_open_files(void);		/* from interface.c */

func do_muf_topprofs(player ObjectID, arg1 string) {
	struct profnode {
		struct profnode *next;
		ObjectID  prog;
		double proftime;
		double pcnt;
		long   comptime;
		long   usecount;
	} *tops = NULL;

	struct profnode *curr = NULL;
	int nodecount = 0;
	ObjectID i = NOTHING;
	int count = atoi(arg1);
	current_systime := time.Now()

	switch {
	case !Wizard(DB.FetchPlayer(player).Owner):
		notify(player, "Permission denied. (MUF profiling stats are wiz-only)");
		return
	case arg1 == "reset":
		EachObjectInReverse(func(obj ObjectID, o *Object) {
			if p := o.(Program); IsProgram(obj) {
				p.proftime = 0
				p.profstart = current_systime
				p.profuses = 0
			}
		})
		notify(player, "MUF profiling statistics cleared.")
		return
	case count < 0:
		notify(player, "Count has to be a positive number.")
		return
	case count == 0:
		count = 10
	}

	EachObjectInReverse(func(obj ObjectID, o *Object) {
		if p := o.(Program); IsProgram(obj) && p.code != nil {
			newnode := &profnode{
				prog: obj,
				proftime: p.proftime,
				comptime: current_systime - p.profstart,
				usecount: p.profuses,
			}
			if newnode.comptime > 0 {
				newnode.pcnt = 100.0 * newnode.proftime / newnode.comptime
			}
			switch {
			case tops == nil:
				tops = newnode
				nodecount++
			case newnode.pcnt < tops.pcnt:
				if nodecount < count {
					newnode.next = tops
					tops = newnode
					nodecount++
				}
			default:
				if nodecount >= count {
					tops = tops.next
				} else {
					nodecount++
				}
				if !tops {
					tops = newnode
				} else if newnode.pcnt < tops.pcnt {
					newnode.next = tops
					tops = newnode
				} else {
					for curr = tops; curr.next != nil; curr = curr.next {
						if newnode.pcnt < curr.next.pcnt {
							break
						}
					}
					newnode.next = curr.next
					curr.next = newnode
				}
			}
		}
	})
	notify(player, "     %CPU   TotalTime  UseCount  Program");
	for tops != nil {
		curr = tops
		notify(player, fmt.Sprintf("%10.3f %10.3f %9ld %s", curr->pcnt, curr->proftime, curr->usecount, unparse_object(player, curr.prog)))
		tops = tops.next
	}
	buf = fmt.Sprintf("Profile Length (sec): %5ld  %%idle: %5.2f%%  Total Cycles: %5lu",
			(current_systime-sel_prof_start_time),
			float64(sel_prof_idle) / (float64(current_systime - sel_prof_start_time) + 0.01),
			sel_prof_idle_use
	)
	notify(player,buf);
	notify(player, "*Done*");
}


func do_mpi_topprofs(player ObjectID, arg1 string) {
	struct profnode {
		struct profnode *next;
		ObjectID  prog;
		double proftime;
		double pcnt;
		long   comptime;
		long   usecount;
	} *tops = NULL

	struct profnode *curr = NULL;
	int nodecount = 0;
	ObjectID i = NOTHING;
	int count = atoi(arg1);
	current_systime := time.Now()

	if !Wizard(DB.FetchPlayer(player).Owner) {
		notify(player, "Permission denied. (MPI statistics are wizard-only)")
		return
	}
	if arg1 == "reset" {
		EachObjectInReverse(func(o *Object) {
			if o.MPIUses {
				o.MPIUses = 0
				o.time.Duration.tv_usec = 0
				o.time.Duration.tv_sec = 0
			}
		})
		mpi_prof_start_time = current_systime
		notify(player, "MPI profiling statistics cleared.")
		return
	}
	if (count < 0) {
		notify(player, "Count has to be a positive number.");
		return;
	} else if (count == 0) {
		count = 10;
	}

	EachObjectInReverse(func(obj ObjectID, o *Object) {
		if o.MPIUses {
			newnode := &profnode{
				prog: obj,
				proftime: o.time.Duration.tv_sec,
				comptime: current_systime - mpi_prof_start_time,
				usecount: o.MPIUses,
			}
			proftime += (o.time.Duration.tv_usec / 1000000.0)
			if newnode.comptime > 0 {
				newnode.pcnt = 100.0 * newnode.proftime / newnode.comptime
			} else {
				newnode.pcnt =  0.0
			}
			if !tops {
				tops = newnode
				nodecount++
			} else if newnode.pcnt < tops.pcnt {
				if nodecount < count {
					newnode.next = tops
					tops = newnode
					nodecount++
				}
			} else {
				if nodecount >= count {
					tops = tops.next
				} else {
					nodecount++
				}
				if !tops {
					tops = newnode
				} else if newnode.pcnt < tops.pcnt {
					newnode.next = tops
					tops = newnode
				} else {
					for curr = tops; curr.next != nil; curr = curr.next {
						if newnode.pcnt < curr.next.pcnt {
							break
						}
					}
					newnode.next = curr.next
					curr.next = newnode
				}
			}
		}
	})
	notify(player, "     %CPU   TotalTime  UseCount  Object")
	for tops != nil {
		curr = tops
		notify(player, fmt.Sprintf("%10.3f %10.3f %9ld %s", curr.pcnt, curr.proftime, curr.Uses, unparse_object(player, curr.prog)))
		tops = tops.next
	}
	notify(player, fmt.Sprintf("Profile Length (sec): %5ld  %%idle: %5.2f%%  Total Cycles: %5lu",
			current_systime - sel_prof_start_time,
			float64(sel_prof_idle) / (float64((current_systime - sel_prof_start_time)) + 0.01),
			sel_prof_idle_use,
	))
	notify(player, "*Done*")
}

type profnode struct {
	next *profnode
	prog ObjectID
	proftime float64
	pcnt float64
	comptime time.Duration
	usecount int
	is_mpi bool
}

func do_all_topprofs(player ObjectID, arg1 string) {
	var curr, tops *profnode
	var buf string
	var nodecount int

	current_systime := time.Now()
	switch {
	case !Wizard(DB.FetchPlayer(player).Owner):
		notify(player, "Permission denied. (server profiling statistics are wizard-only)");
	case arg1 == "reset":
		EachObjectInReverse(func(obj ObjectID, o *Object) {
			if o.MPIUses {
				o.MPIUses = 0
				o.time.Duration.tv_usec = 0
				o.time.Duration.tv_sec = 0
			}
			if p := o.(Program); IsProgram(obj) {
				p.proftime = 0
				p.profstart = current_systime
				p.profuses = 0
			}
		})
		sel_prof_idle = 0
		sel_prof_start_time = current_systime
		sel_prof_idle_use = 0
		mpi_prof_start_time = current_systime
		notify(player, "All profiling statistics cleared.")
	default:
		if count := strconv.Atoi(arg1); count < 0 {
			notify(player, "Count has to be a positive number.")
		} else {
			if count == 0 {
				count = 10
			}
			EachObjectInReverse(func(obj ObjectID, o *Object) {
				if o.MPIUses {
					newnode := &profnode{
						prog: obj,
						proftime: o.time.Duration.tv_sec + (o.time.Duration.tv_usec / 1000000.0),
						comptime: current_systime.Sub(mpi_prof_start_time),
						usecount: o.MPIUses,
					}
					if newnode.comptime > 0 {
						newnode.pcnt = 100.0 * newnode.proftime / newnode.comptime
					} else {
						newnode.pcnt =  0.0
					}
					switch {
					case tops == nil:
						tops = newnode
						nodecount++
					case newnode.pcnt < tops.pcnt:
						if nodecount < count {
							newnode.next = tops
							tops = newnode
							nodecount++
						}
					default:
						if nodecount >= count {
							curr = tops
							tops = tops.next
						} else {
							nodecount++
						}
						switch {
						case tops == nil:
							tops = newnode
						case newnode.pcnt < tops.pcnt:
							newnode.next = tops
							tops = newnode
						default:
							for curr = tops; curr.next != nil && newnode.pcnt < curr.next.pcnt; curr = curr.next {}
							newnode.next = curr.next
							curr.next = newnode
						}
					}
				}
				if p := o.(Program); p.code != nil {
					newnode := &profnode{
						prog: i,
						proftime: p.proftime,
						comptime: current_systime.Sub(p.profstart),
						usecount: p.profuses,
						is_mpi: true,
					}
					if newnode.comptime > 0 {
						newnode.pcnt = 100.0 * newnode.proftime / newnode.comptime
					} else {
						newnode.pcnt =  0.0
					}
					switch {
					case tops == nil:
						tops = newnode
						nodecount++
					case newnode.pcnt < tops.pcnt:
						if nodecount < count {
							newnode.next = tops
							tops = newnode
							nodecount++
						}
					default:
						if nodecount >= count {
							curr = tops
							tops = tops.next
						} else {
							nodecount++
						}
						switch {
						case tops == nil:
							tops = newnode
						case newnode.pcnt < tops.pcnt:
							newnode.next = tops
							tops = newnode
						default:
							for curr = tops; cur.next != nil && curr.next.pcnt <= newnode.pcnt; curr = curr.next {}
							newnode.next = curr.next
							curr.next = newnode
						}
					}
				}
			})
			notify(player, "     %CPU   TotalTime  UseCount  Type  Object")
			for ; tops != nil; tops = tops.next {
				if curr := tops; curr.is_mpi {
					notify(player, fmt.Sprintf("%10.3f %10.3f %9ld  MPI   %s", curr.pcnt, curr.proftime, curr.Uses, unparse_object(player, curr.prog)))
				} else {
					notify(player, fmt.Sprintf("%10.3f %10.3f %9ld  MUF   %s", curr.pcnt, curr.proftime, curr.Uses, unparse_object(player, curr.prog)))
				}
			}
			notify(player,
				fmt.Sprintf("Profile Length (sec): %5ld  %%idle: %5.2f%%  Total Cycles: %5lu",
					current_systime.Sub(sel_prof_start_time),
					float64(sel_prof_idle) / (float64(current_systime.Sub(sel_prof_start_time)) + 0.01),
					sel_prof_idle_use
				)
			)
			notify(player, "*Done*")		
		}
	}
}