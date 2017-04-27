type Match struct {
	exact ObjectID					/* holds result of exact match */
	check_keys bool				/* if non-zero, check for keys */
	last ObjectID					/* holds result of last match */
	count int				/* holds total number of inexact matches */
	who ObjectID					/* player used for me, here, and messages */
	from ObjectID					/* object which is being matched around */
	descr int					/* descriptor initiating the match */
	name string					/* name to match */
	preferred_type int			/* preferred type */
	longest int			/* longest matched string */
	level int				/* the highest priority level so far */
	block_equals int			/* block matching of same name exits */
	partial_exits int			/* if non-zero, allow exits to match partially */
}

#define NOMATCH_MESSAGE "I don't see that here."
#define AMBIGUOUS_MESSAGE "I don't know which one you mean!"

char match_cmdname[BUFFER_LEN];	/* triggering command */
char match_args[BUFFER_LEN];	/* remaining text */

func IsExit(v interface{}) (ok bool) {
	_, ok = v.(TYPE_EXIT)
	return
}

func IsThing(v interface{}) (ok bool) {
	_, ok = v.(TYPE_THING)
	return	
}

func IsPlayer(v interface{}) (ok bool) {
	_, ok = v.(TYPE_PLAYER)
	return	
}

func IsRoom(v interface{}) (ok bool) {
	_, ok = v.(TYPE_ROOM)
	return
}

func IsProgram(v interface{}) (ok bool) {
	_, ok = v.(TYPE_PROGRAM)
	return
}

func NewMatch(descr int, player ObjectID, name string, type_checker func(interface{}) bool) *Match {
	return &Match{
		exact: NOTHING,
		last: NOTHING,
		who: player,
		from: player,
		descr: descr,
		name: name,
		preferred_type: type_checker,
		partial_exits: type_checker == IsExit,
	}
}

func NewMatchCheckKeys(descr int, player ObjectID, name string, type_checker func(interface{}) bool) (r *Match) {
	r = NewMatch(descr, player, name, datatype)
	r.check_keys = true
	return
}

func NewMatchRemote(descr int, player, what ObjectID, name string, type_checker func(interface{}) bool) (r *Match) {
	r = NewMatch(descr, player, name, datatype)
	r.from = what
	return
}

func (m *Match) EitherOf(thing1, thing2 ObjectID) (r ObjectID) {
	switch {
	case thing1 == NOTHING:
		r = thing2
	case thing2 == NOTHING:
		r = thing1
	case preferred != nil && m.preferred_type(thing1) && !m.preferred_type(thing2):
		r = thing1
	case preferred != nil && m.preferred_type(thing2):
		r = thing2
	case m.check_keys:
		has1 := could_doit(m.descr, m.who, thing1)
		has2 := could_doit(m.descr, m.who, thing2)

		switch {
		case has1 && !has2:
			r = thing1
		case has2 && !has1:
			r = thing2
		}
		if RANDOM() % 2 == 0 {
			r = thing1
		} else {
			r = thing2
		}
	default:
		if RANDOM() % 2 == 0 {
			r = thing1
		} else {
			r = thing2
		}
	}
	return
}

func (m *Match) MatchPlayer() *Match {
	if m.name[0] == LOOKUP_TOKEN && payfor(DB.Fetch(m.from).Owner, tp_lookup_cost) {
		var p string
		if i := strings.IndexFunc(n.name[1:], unicode.IsSpace); i != -1 {
			p = n.name[i + 1:]
		}
		if match := lookup_player(p); match != NOTHING {
			m.exact = match
		}
	}
	return m
}

/* returns ObjectID if registered object found for name, else NOTHING */
func find_registered_obj(player ObjectID, name string) (r ObjectID) {
	if r = NOTHING; name[0] == REGISTERED_TOKEN {
		if p := strings.TrimSpace(name[1:]); p != "" {
			if _, p := envprop(player, fmt.Sprint("_reg/", p)); p != nil {
				switch v := p.data.(type) {
				case string:
					if v[0] == NUMBER_TOKEN {
						v = v[1:]
					}
					if unicode.IsNumber(v[0]) {
						if v := ObjectID(strconv.Atoi(v)); v.IsValid() {
							return v
						}
					}
				case ObjectID:
					if v.IsValid() {
						return v
					}
				case int:
					if v := ObjectID(v); v.IsValid() {
						return v
					}
				}
			}
		}
	}
	return
}

func (m *Match) MatchRegistered() *Match {
	if match := find_registered_obj(m.from, m.name); match != NOTHING {
		m.exact = match
	}
	return m
}

/* returns nnn if name = #nnn, else NOTHING */
func (m *Match) AbsoluteName() (r ObjectID) {
	if m.name[0] == NUMBER_TOKEN {
		r = parse_ObjectID(m.name[1:])
		if !match.IsValid() {
			r = NOTHING
		}
	} else {
		r = NOTHING
	}
	return
}

func (m *Match) MatchAbsolute() *Match {
	if match := m.AbsoluteName(); match != NOTHING {
		m.exact = match
	}
	return m
}

func (m *Match) MatchMe() *Match {
	if m.name == "me" {
		m.exact = m.who
	}
	return m
}

func (m *Match) MatchHere() *Match {
	if m.name == "here" && DB.Fetch(m.who).Location != NOTHING {
		m.exact = DB.Fetch(m.who).Location
	}
	return m
}

func (m *Match) MatchHome() *Match {
	if m.name == "home" {
		m.exact = HOME
	}
	return m
}

func (m *Match) MatchList(first ObjectID) *Match {
	absolute := m.AbsoluteName()
	if !controls(DB.Fetch(m.from).Owner, absolute) {
		absolute = NOTHING
	}
	for ; first != NOTHING; first = DB.Fetch(first).next {
		switch {
		case first == absolute:
			m.exact = first
			break
		case DB.Fetch(first).name == m.name:
			/* if there are multiple exact matches, randomly choose one */
			m.exact = md.EitherOf(m.exact, first)
		case string_match(DB.Fetch(first).name, m.name):
			m.last = first
			m.count++
		}
	}
	return m
}

func (m *Match) MatchPossession() *Match {
	m.MatchList(DB.Fetch(m.from).Contents)
	return m
}

func (m *Match) MatchNeighbor() *Match {
	if loc := DB.Fetch(m.from).Location); loc != NOTHING {
		m.MatchList(DB.Fetch(loc).Contents)
	}
	return m
}

//	MatchExits matches a list of exits, starting with 'first'.
//	It will match exits of players, rooms, or things.
func (m *Match) MatchExits(first ObjectID) *Match {
	if first != NOTHING && DB.Fetch(m.from).Location != NOTHING {
		absolute := m.AbsoluteName()
		if !controls(DB.Fetch(m.from).Owner, absolute) {
			absolute = NOTHING
		}
		for exitid := first; exitid != NOTHING; {
			if exitid == absolute {
				m.exact = exitid
			} else {
				exit := DB.Fetch(exitid)
				var exitprog bool
				switch {
				case exit.IsFlagged(HAVEN):
					exitprog = true
				case exit.(Exit).Destinations != nil:
					for _, v := range exit.(Exit).Destinations {
						if IsProgram(v) {
							exitprog = true
							break
						}
					}
				}
				partial := tp_enable_prefix && exitprog && md.partial_exits && exit.IsFlagged(XFORCIBLE) && DB.Fetch(exit.Owner).IsFlagged(WIZARD)
				for exitname := strings.TrimSpace(exit.name); exitname != ""; exitname = strings.TrimSpace(exitname) {
					var notnull bool
					var p string
					for p = m.name; p != "" && strings.ToLower(p[0]) == strings.ToLower(exitname[0]) && exitname != EXIT_DELIMITER; p = p[1:] {
						if !unicode.IsSpace(p[0]) {
							notnull = true
						}
						exitname = exitname[1:]
					}
					/* did we get a match on this alias? */
					if (partial && notnull) || p == "" || (p[0] == " " && exitprog) {
						/* make sure there's nothing afterwards */
						if i := strings.IndexFunc(exitname, unicode.IsSpace); i != -1 {
							exitname = exitname[:i]
						}
						lev := PLevel(exitid)
						if tp_compatible_priorities && lev == 1 && (exit.Location == NOTHING || TYPEOF(exit.Location) != TYPE_THING || controls(exit.Owner, DB.Fetch(m.from).Location)) {
							lev = 2
						}
						if exitname == "" || exitname[0] == EXIT_DELIMITER {
							/* we got a match on this alias */
							switch {
							case lev < m.level:
							case len(m.name) - len(p) > m.longest:
								if lev > m.level {
									m.level = lev
									m.block_equals = 0
								}
								m.exact = exitid
								m.longest = len(m.name) - len(p)
								if p[0] == ' ' || (partial && notnull) {
									if partial && notnull {
										match_args = p
									} else {
										match_args = p[1:]
									}
									var ip int
									for pp := m.name; pp != "" && pp[0] != p; pp = pp[1:] {
										ip++
									}
									match_cmdname = pp[:ip]
								} else {
									match_args = ""
									match_cmdname = m.name
								}
							case (len(m.name) - len(p) == m.longest) && !(lev == m.level && m.block_equals):
								if lev > m.level {
									m.exact = exitid
									m.level = lev
									m.block_equals = 0
								} else {
									m.exact = md.EitherOf(md.exact, exitid)
								}
								if m.exact == exitid {
									if p[0] == " " || (partial && notnull) {
										if partial && notnull {
											match_args = p
										} else {
											match_args = p[1:]
										}
										var ip int
										for pp := m.name; pp != "" && pp[0] != p; pp = pp[1:] {
											ip++
										}
										match_cmdname = pp[:ip]
									} else {
										match_args = ""
										match_cmdname = m.name
									}
								}
							}
							goto next_exit
						}
					}
					/* we didn't get it, go on to next alias */
					for ; exitname != "" && exitname[0] != EXIT_DELIMITER ; exitname = exitname[1:] {}
				}
next_exit:
			}
			exitid = DB.Fetch(exit).next
		}
	}
	return md
}

//	matches actions attached to objects in inventory
func (m *Match) MatchInvobjActions() *Match {
	for thing := DB.Fetch(m.from).Contents; thing != NOTHING; thing = DB.Fetch(thing).next {
		if IsThing(thing) && DB.Fetch(thing).Exits != NOTHING {
			m.MatchExits(DB.Fetch(thing).Exits)
		}
	}
	return m
}

//	matches actions attached to objects in the room
func (m *Match) MatchRoomobjActions() *Match {
	if loc := DB.Fetch(m.from).Location) != NOTHING {
		for thing := DB.Fetch(loc).Contents; thing != NOTHING; thing = DB.Fetch(thing).next {
			if IsThing(thing) && DB.Fetch(thing).Exits != NOTHING {
				m.MatchExits(DB.Fetch(thing).Exits)
			}
		}
	}
	return m
}

//	matches actions attached to player
func (m *Match) MatchPlayerActions() *Match {
	switch m.from.(type) {
	case TYPE_PLAYER:, TYPE_ROOM, TYPE_THING:
		m.MatchExits(DB.Fetch(m.from).Exits)
	}
	return m
}

//	Matches exits and actions attached to player's current room.
//	Formerly 'match_exit'.
func (m *Match) MatchRoomExits(loc ObjectID) *Match {
	switch loc.(type) {
	case TYPE_PLAYER, TYPE_ROOM, TYPE_THING:
		m.MatchExits(DB.Fetch(loc).Exits)
	}
	return m
}

/*
 * MatchAllExits
 * Matches actions on player, objects in room, objects in inventory,
 * and room actions/exits (in reverse order of priority order).
 */
func (m *Match) MatchAllExits() *Match {
	ObjectID loc;
	int limit = 88;
	int blocking = 0;

	var match_args, match_cmdname string
	if loc = DB.Fetch(m.from).Location; loc != NOTHING) {
		md.MatchRoomExits(loc)
	}

	if m.exact != NOTHING {
		m.block_equals = true
	}
	m.MatchInvobjActions()

	if m.exact != NOTHING {
		m.block_equals = true
	}
	m.MatchRoomobjActions()

	if m.exact != NOTHING {
		m.block_equals = true
	}
	m.MatchPlayerActions()

	if loc != NOTHING {
		/* if player is in a vehicle, use environment of vehicle's home */
		if loc, ok := loc.(Object); ok {
			if loc = DB.FetchPlayer(loc).Home; loc == NOTHING {
				return
			}
			if m.exact != NOTHING {
				m.block_equals = true
			}
			m.MatchRoomExits(loc)
		}

        /* Walk the environment chain to #0, or until depth chain limit
           has been hit, looking for a match. */
        for loc = DB.Fetch(loc).Location; loc != NOTHING; loc = DB.Fetch(loc).Location {
			/* If we're blocking (because of a yield), only match a room if
			   and only if it has overt set on it. */
			if (blocking && DB.Fetch(loc).IsFlagged(OVERT)) || !blocking {
				if m.exact != NOTHING {
					m.block_equals = true
				}
				m.MatchRoomExits(loc)
			}
			if !limit-- {
				break
			}
			/* Does this room have env-chain exit blocking enabled? */
			if !blocking && tp_enable_match_yield && DB.Fetch(loc).IsFlagged(YIELD) {
				blocking = true
			}
        }
	}
	return m
}

func (m *Match) MatchEverything() *Match {
	m.MatchAllExits().
		MatchNeighbor().
		MatchPossession().
		MatchMe().
		MatchHere().
		MatchRegistered()
	if Wizard(DB.Fetch(m.from).Owner) || Wizard(m.who) {
		m.MatchAbsolute().MatchPlayer()
	}
	return m
}

func (m *Match) MatchResult() (r ObjectID) {
	if md.exact != NOTHING {
		r = m.exact
	} else {
		switch m.count {
		case 0:
			r = NOTHING
		case 1:
			r = m.last
		default:
			r = AMBIGUOUS
		}
	}
	return
}

/* use this if you don't care about ambiguity */
func (m *Match) LastMatchResult() (r ObjectID) {
	if m.exact != NOTHING {
		r = m.exact
	} else {
		r = m.last
	}
	return
}

func (m *Match) NoisyMatchResult() (r ObjectID) {
	switch r = m.MatchResult(); {
	case NOTHING:
		notify(m.who, NOMATCH_MESSAGE)
		r = NOTHING
	case AMBIGUOUS:
		notify(m.who, AMBIGUOUS_MESSAGE)
		r = NOTHING
	default:
		r = match
	}
	return
}

func (m *Match) RMatch(v ObjectID) {
	if v != NOTHING {
		switch v.(type) {
		case TYPE_PLAYER, TYPE_ROOM, TYPE_THING:
			md.MatchList(DB.Fetch(v).Contents)
			md.MatchExits(DB.Fetch(v).Exits)
		}
	}
}