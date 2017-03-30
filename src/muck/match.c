char match_cmdname[BUFFER_LEN];	/* triggering command */
char match_args[BUFFER_LEN];	/* remaining text */

func NewMatch(descr int, player dbref, name string, type int) *match_data {
	return &match_data{
		exact_match: NOTHING,
		last_match: NOTHING,
		match_who: player,
		match_from: player,
		match_descr: descr,
		match_name: name,
		preferred_type: type,
		partial_exits: (TYPE_EXIT == type),
	}
}

func NewMatchCheckKeys(descr int, player dbref, name string, type int) (r *match_data) {
	r = NewMatch(descr, player, name, type)
	r.check_keys = true
	return
}

func NewMatchRemote(descr int, player, what dbref, name string, type int) (r *match_data) {
	r = NewMatch(descr, player, name, type)
	r.match_from = what
	return
}

static dbref
choose_thing(int descr, dbref thing1, dbref thing2, struct match_data *md)
{
	int has1;
	int has2;
	int preferred = md->preferred_type;

	if (thing1 == NOTHING) {
		return thing2;
	} else if (thing2 == NOTHING) {
		return thing1;
	}
	if (preferred != NOTYPE) {
		if (Typeof(thing1) == preferred) {
			if (Typeof(thing2) != preferred) {
				return thing1;
			}
		} else if (Typeof(thing2) == preferred) {
			return thing2;
		}
	}
	if (md->check_keys) {
		has1 = could_doit(descr, md->match_who, thing1);
		has2 = could_doit(descr, md->match_who, thing2);

		if (has1 && !has2) {
			return thing1;
		} else if (has2 && !has1) {
			return thing2;
		}
		/* else fall through */
	}
	return (RANDOM() % 2 ? thing1 : thing2);
}

void
match_player(struct match_data *md)
{
	dbref match;
	const char *p;

	if (*(md->match_name) == LOOKUP_TOKEN && payfor(db.Fetch(md.match_from).owner, tp_lookup_cost)) {
		p = strings.TrimLeftFunc(md.match_name[1:], func(r rune) bool {
			return !unicode.IsSpace(r)
		})
		if match = lookup_player(p); match != NOTHING {
			md->exact_match = match;
		}
	}
}

/* returns dbref if registered object found for name, else NOTHING */
func find_registered_obj(player dbref, name string) (r dbref) {
	r = NOTHING
	if name[0] == REGISTERED_TOKEN {
		if p := strings.TrimSpace(name[1:]); p != "" {
			buf := fmt.Sprintf("_reg/%s", p)
			if _, p := envprop(player, buf); p != nil {
				switch v := p.data.(type) {
				case string:
					if v[0] == NUMBER_TOKEN {
						v = v[1:]
					}
					if unicode.IsNumber(v[0]) {
						if match = dbref(strconv.Atoi(v)); match >= 0 && match < db_top {
							return match
						}
					}
				case dbref:
					if match = v; match >= 0 && match < db_top {
						return match
					}
				case int:
					if match = dbref(v); match > 0 && match < db_top {
						return match
					}
				}
			}
		}
	}
	return
}

func match_registered(md *match_data) {
	if match := find_registered_obj(md.match_from, md.match_name); match != NOTHING {
		md.exact_match = match
	}
}

/* returns nnn if name = #nnn, else NOTHING */
static dbref
absolute_name(struct match_data *md)
{
	dbref match;

	if (*(md->match_name) == NUMBER_TOKEN) {
		match = parse_dbref((md->match_name) + 1);
		if (match < 0 || match >= db_top) {
			return NOTHING;
		} else {
			return match;
		}
	} else {
		return NOTHING;
	}
}

void
match_absolute(struct match_data *md)
{
	dbref match;

	if ((match = absolute_name(md)) != NOTHING) {
		md->exact_match = match;
	}
}

func match_me(md *match_data) {
	if md.match_name == "me" {
		md->exact_match = md->match_who;
	}
}

func match_here(md *match_data) {
	if md.match_name == "here" && db.Fetch(md.match_who).location != NOTHING {
		md.exact_match = db.Fetch(md.match_who).location
	}
}

func match_home(md *match_data) {
	if md.match_name == "home" {
		md.exact_match = HOME
	}
}

func match_list(first dbref, md *match_data) {
	dbref absolute;

	absolute = absolute_name(md);
	if (!controls(db.Fetch(md.match_from).owner, absolute))
		absolute = NOTHING;

	for ; first != NOTHING; first = db.Fetch(first).next {
		if (first == absolute) {
			md->exact_match = first;
			return;
		} else if db.Fetch(first).name == md->match_name {
			/* if there are multiple exact matches, randomly choose one */
			md->exact_match = choose_thing(md->match_descr, md->exact_match, first, md);
		} else if (string_match(db.Fetch(first).name, md->match_name)) {
			md->last_match = first;
			(md->match_count)++;
		}
	}
}

func match_possession(md *match_data) {
	match_list(db.Fetch(md.match_from).contents, md)
}

func match_neighbor(md  *match_data) {
	if loc := db.Fetch(md.match_from).location); loc != NOTHING {
		match_list(db.Fetch(loc).contents, md)
	}
}

/*
 * match_exits matches a list of exits, starting with 'first'.
 * It will match exits of players, rooms, or things.
 */
func match_exits(first dbref, md *match_data) {
	dbref exit, absolute;
	const char *exitname, *p;
	int i, exitprog, lev, partial;

	if first == NOTHING || db.Fetch(md.match_from).location == NOTHING {
		return
	}
	absolute = absolute_name(md);	/* parse #nnn entries */
	if (!controls(db.Fetch(md.match_from).owner, absolute))
		absolute = NOTHING;

	for exit := first; exit != NOTHING; exit = db.Fetch(exit).next {
		if (exit == absolute) {
			md->exact_match = exit;
			continue;
		}
		exitprog = false
		switch {
		case db.Fetch(exit).flags & HAVEN != 0:
			exitprog = true
		case db.Fetch(exit).sp.exit.dest != nil:
			for _, v := range db.Fetch(exit).sp.exit.dest {
				if TYPEOF(v) == TYPE_PROGRAM {
					exitprog = true
					break
				}
			}
		}
		partial = tp_enable_prefix && exitprog && md.partial_exits && (db.Fetch(exit).flags & XFORCIBLE) && db.Fetch(db.Fetch(exit).owner).flags & WIZARD != 0
		for exitname = db.Fetch(exit).name; exitname != ""; {
			int notnull = 0;
			for (p = md.match_name;	p != "" && strings.ToLower(p) == strings.ToLower(exitname) && exitname != EXIT_DELIMITER; p++, exitname++) {
				if !unicode.IsSpace(p[0]) {
					notnull = 1;
				}
			}
			/* did we get a match on this alias? */
			if ((partial && notnull) || ((*p == '\0') || (*p == ' ' && exitprog))) {
				/* make sure there's nothing afterwards */
				exitname = strings.TrimLeftFunc(exitname, func(r rune) bool {
					return !unicode.IsSpace(r)
				})
				lev = PLevel(exit);
				if (tp_compatible_priorities && (lev == 1) && (db.Fetch(exit).location == NOTHING || TYPEOF(db.Fetch(exit).location) != TYPE_THING || controls(db.Fetch(exit).owner, db.Fetch(md.match_from).location))) {
					lev = 2
				}
				if (*exitname == '\0' || *exitname == EXIT_DELIMITER) {
					/* we got a match on this alias */
					if (lev >= md->match_level) {
						if (len(md->match_name) - len(p) > md->longest_match) {
							if (lev > md->match_level) {
								md->match_level = lev;
								md->block_equals = 0;
							}
							md->exact_match = exit;
							md->longest_match = len(md->match_name) - len(p);
							if ((*p == ' ') || (partial && notnull)) {
								strcpyn(match_args, sizeof(match_args), (partial && notnull)? p : (p + 1));
								{
									char *pp;
									int ip;

									for (ip = 0, pp = (char *) md->match_name;
										 *pp && (pp != p); pp++)
										match_cmdname[ip++] = *pp;
									match_cmdname[ip] = '\0';
								}
							} else {
								*match_args = '\0';
								strcpyn(match_cmdname, sizeof(match_cmdname), (char *) md->match_name);
							}
						} else if ((len(md->match_name) - len(p) ==
									md->longest_match) && !((lev == md->match_level) &&
															(md->block_equals))) {
							if (lev > md->match_level) {
								md->exact_match = exit;
								md->match_level = lev;
								md->block_equals = 0;
							} else {
								md->exact_match =
										choose_thing(md->match_descr, md->exact_match, exit,
													 md);
							}
							if (md->exact_match == exit) {
								if ((*p == ' ') || (partial && notnull)) {
									strcpyn(match_args, sizeof(match_args), (partial && notnull) ? p : (p + 1));
									{
										char *pp;
										int ip;

										for (ip = 0, pp = (char *) md->match_name;
											 *pp && (pp != p); pp++)
											match_cmdname[ip++] = *pp;
										match_cmdname[ip] = '\0';
									}
								} else {
									*match_args = '\0';
									strcpyn(match_cmdname, sizeof(match_cmdname), (char *) md->match_name);
								}
							}
						}
					}
					goto next_exit;
				}
			}
			/* we didn't get it, go on to next alias */
			while (*exitname && *exitname++ != EXIT_DELIMITER) ;
			exitname = strings.TrimLeftFunc(exitname, unicode.IsSpace)
		}						/* end of while alias string matches */
	  next_exit:
		;
	}
}

/*
 * match_invobj_actions
 * matches actions attached to objects in inventory
 */
func match_invobj_actions(md *match_data) {
	for thing := db.Fetch(md.match_from).contents; thing != NOTHING; thing = db.Fetch(thing).next {
		if TYPEOF(thing) == TYPE_THING && db.Fetch(thing).exits != NOTHING {
			match_exits(db.Fetch(thing).exits, md)
		}
	}
}

/*
 * match_roomobj_actions
 * matches actions attached to objects in the room
 */
func match_roomobj_actions(md *match_data) {
	if loc := db.Fetch(md.match_from).location) != NOTHING {
		for thing := db.Fetch(loc).contents; thing != NOTHING; thing = db.Fetch(thing).next {
			if TYPEOF(thing) == TYPE_THING && db.Fetch(thing).exits != NOTHING {
				match_exits(db.Fetch(thing).exits, md)
			}
		}
	}
}

/*
 * match_player_actions
 * matches actions attached to player
 */
func match_player_actions(md *match_data) {
	switch TYPEOF(md.match_from) {
	case TYPE_PLAYER:, TYPE_ROOM, TYPE_THING:
		match_exits(db.Fetch(md.match_from).exits, md)
	}
}

/*
 * match_room_exits
 * Matches exits and actions attached to player's current room.
 * Formerly 'match_exit'.
 */
func match_room_exits(loc dbref, md *match_data) {
	switch TYPEOF(loc) {
	case TYPE_PLAYER, TYPE_ROOM, TYPE_THING:
		match_exits(db.Fetch(loc).exits, md)
	}
}

/*
 * match_all_exits
 * Matches actions on player, objects in room, objects in inventory,
 * and room actions/exits (in reverse order of priority order).
 */
func match_all_exits(md *match_data) {
	dbref loc;
	int limit = 88;
	int blocking = 0;

	var match_args, match_cmdname string
	if loc = db.Fetch(md.match_from).location; loc != NOTHING) {
		match_room_exits(loc, md)
	}

	if md.exact_match != NOTHING {
		md.block_equals = true
	}
	match_invobj_actions(md)

	if md.exact_match != NOTHING {
		md.block_equals = true
	}
	match_roomobj_actions(md)

	if md.exact_match != NOTHING {
		md.block_equals = true
	}
	match_player_actions(md)

	if loc != NOTHING {
		/* if player is in a vehicle, use environment of vehicle's home */
		if Typeof(loc) == TYPE_THING {
			if loc = db.Fetch(loc).sp.(player_specific).home; loc == NOTHING {
				return
			}
			if md->exact_match != NOTHING {
				md->block_equals = true
			}
			match_room_exits(loc, md)
		}

        /* Walk the environment chain to #0, or until depth chain limit
           has been hit, looking for a match. */
        for loc = db.Fetch(loc).location; loc != NOTHING; loc = db.Fetch(loc).location {
			/* If we're blocking (because of a yield), only match a room if
			   and only if it has overt set on it. */
			if (blocking && db.Fetch(loc).flags & OVERT != 0) || !blocking {
				if md.exact_match != NOTHING {
					md->block_equals = true
				}
				match_room_exits(loc, md);
			}
			if !limit-- {
				break
			}
			/* Does this room have env-chain exit blocking enabled? */
			if !blocking && tp_enable_match_yield && db.Fetch(loc).flags & YIELD != 0 {
				blocking = true
			}
        }
}

void
match_everything(struct match_data *md)
{
	match_all_exits(md);
	match_neighbor(md);
	match_possession(md);
	match_me(md);
	match_here(md);
	match_registered(md);
	if Wizard(db.Fetch(md.match_from).owner) || Wizard(md.match_who) {
		match_absolute(md);
		match_player(md);
	}
}

dbref
match_result(struct match_data *md)
{
	if (md->exact_match != NOTHING) {
		return (md->exact_match);
	} else {
		switch (md->match_count) {
		case 0:
			return NOTHING;
		case 1:
			return (md->last_match);
		default:
			return AMBIGUOUS;
		}
	}
}

/* use this if you don't care about ambiguity */
dbref
last_match_result(struct match_data * md)
{
	if (md->exact_match != NOTHING) {
		return (md->exact_match);
	} else {
		return (md->last_match);
	}
}

dbref
noisy_match_result(struct match_data * md)
{
	dbref match;

	switch (match = match_result(md)) {
	case NOTHING:
		notify(md->match_who, NOMATCH_MESSAGE);
		return NOTHING;
	case AMBIGUOUS:
		notify(md->match_who, AMBIGUOUS_MESSAGE);
		return NOTHING;
	default:
		return match;
	}
}

func match_rmatch(arg1 dbref, md *match_data) {
	if arg1 != NOTHING {
		switch TYPEOF(arg1) {
		case TYPE_PLAYER, TYPE_ROOM, TYPE_THING:
			match_list(db.Fetch(arg1).contents, md)
			match_exits(db.Fetch(arg1).exits, md)
		}
	}
}