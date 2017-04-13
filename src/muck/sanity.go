package fbmuck

#define SanFixed(ref, fixed) san_fixed_log((fixed), 1, (ref), -1)
#define SanFixed2(ref, ref2, fixed) san_fixed_log((fixed), 1, (ref), (ref2))
#define SanFixedRef(ref, fixed) san_fixed_log((fixed), 0, (ref), -1)

var SanityViolated bool

var san_linesprinted int
func SanPrint(player dbref, format string, args ...interface{}) {
	switch buf := fmt.Sprintf(format, args...); player {
	case NOTHING:
		fprintf(stdout, "%s\n", buf)
		fflush(stdout)
	case AMBIGUOUS:
		fprintf(stderr, "%s\n", buf)
	default:
		notify_nolisten(player, buf, true)
		if san_linesprinted++ > 100 {
			flush_user_output(player)
			san_linesprinted = 0
		}
	}
}

func sane_dump_object(player dbref, arg string) {
	var d dbref
	int result;

	if player > 0 {
		if player != GOD {
			notify(player, "Permission Denied.")
			return
		}
		result = sscanf(arg, "#%i", &d)
	} else {
		result = sscanf(arg, "%i", &d)
	}

	if !result || !valid_reference(d) {
		SanPrint(player, "Invalid Object.")
		return
	}

	p := db.Fetch(d)
	SanPrint(player, "Object:         %s", unparse_object(GOD, d))
	SanPrint(player, "  Owner:          %s", unparse_object(GOD, p.Owner))
	SanPrint(player, "  Location:       %s", unparse_object(GOD, p.Location))
	SanPrint(player, "  Contents Start: %s", unparse_object(GOD, p.Contents))
	SanPrint(player, "  Exits Start:    %s", unparse_object(GOD, p.Exits))
	SanPrint(player, "  Next:           %s", unparse_object(GOD, p.next))

	switch p := p.(type) {
	case Object:
		SanPrint(player, "  Home:           %s", unparse_object(GOD, p.home))
		SanPrint(player, "  Value:          %d", get_property_value(d, MESGPROP_VALUE))
	case Room:
		SanPrint(player, "  Drop-to:        %s", unparse_object(GOD, p.dbref))
	case Player:
		SanPrint(player, "  Home:           %s", unparse_object(GOD, p.home))
		SanPrint(player, "  Pennies:        %d", get_property_value(d, MESGPROP_VALUE))
		if player < 0 {
			SanPrint(player, "  Password MD5:   %s", p.password)
		}
	case Exit:
		SanPrint(player, "  Links:")
		for _, v := range p.Destinations {
			SanPrint(player, "    %s", unparse_object(GOD, v))
		}
	}
	SanPrint(player, "Referring Objects:")
	EachObject(func(obj dbref, o *Object) {
		if o.Contents == d {
			SanPrint(player, "  By contents field: %s", unparse_object(GOD, obj))
		}
		if o.Exits == d {
			SanPrint(player, "  By exits field:    %s", unparse_object(GOD, obj))
		}
		if o.next == d {
			SanPrint(player, "  By next field:     %s", unparse_object(GOD, obj))
		}
	})
	SanPrint(player, "Done.")
}

func violate(player, i dbref, s string) {
	SanPrint(player, "Object \"%s\" %s!", unparse_object(GOD, i), s)
	SanityViolated = true
}

func valid_obj(obj dbref) (r bool) {
	switch {
	case obj == NOTHING, !valid_reference(obj):
	case IsRoom(obj), IsExit(obj), IsPlayer(obj), IsProgram(obj), IsThing(obj):
		r = true
	}
	return
}

func check_next_chain(player, obj dbref) {
	orig := obj
	for obj != NOTHING && valid_reference(obj) {
		for i := orig; i != NOTHING; i = db.Fetch(i).next {
			if i == db.Fetch(obj).next {
				violate(player, obj, "has a 'next' field that forms an illegal loop in an object chain")
				return
			}
			if i == obj {
				break
			}
		}
		obj = db.Fetch(obj).next
	}
	if !valid_reference(obj) && obj != NOTHING {
		violate(player, obj, "has an invalid object in its 'next' chain");
	}
}

func find_orphan_objects(player dbref) {
	SanPrint(player, "Searching for orphan objects...");
	EachObject(func(o *Object) {
		o.flags &= ~SANEBIT
	})

	db.Fetch(GLOBAL_ENVIRONMENT).flags |= SANEBIT

	EachObject(func(obj dbref, o *Object) {
		if o.Exits != NOTHING {
			if p := db.Fetch(o.Exits); p.flags & SANEBIT != 0 {
				violate(player, o.Exits, "is referred to by more than one object's Next, Contents, or Exits field")
			} else {
				p.flags |= SANEBIT
			}
		}
		if o.Contents != NOTHING {
			if p := db.Fetch(o.Contents); p.flags & SANEBIT != 0 {
				violate(player, o.Contents, "is referred to by more than one object's Next, Contents, or Exits field")
			} else {
				p.flags |= SANEBIT
			}
		}
		if o.next != NOTHING {
			if p := db.Fetch(o.next); p.flags & SANEBIT != 0 {
				violate(player, o.next, "is referred to by more than one object's Next, Contents, or Exits field")
			} else {
				p.flags |= SANEBIT
			}
		}
		if o.flags & SANEBIT == 0 {
			violate(player, obj, "appears to be an orphan object, that is not referred to by any other object")
		}
		o.flags &= ~SANEBIT
	})
}

func check_room(player, obj dbref) {
	switch i := db.Fetch(obj).(dbref); {
	case !valid_reference(i) && v != NOTHING && i != HOME:
		violate(player, obj, "has its dropto set to an invalid object")
	case i >= 0 && TYPEOF(i) != TYPE_THING && TYPEOF(i) != TYPE_ROOM:
		violate(player, obj, "has its dropto set to a non-room, non-thing object")
	}
}

func check_exit(player, obj dbref) {
	for i, v := range db.Fetch(obj).(Exit).Destinations {
		if !valid_reference(v) && v != NOTHING && v != HOME {
			violate(player, obj, "has an invalid object as one of its link destinations")
		}
	}
}

func check_player(player, obj dbref) {
	i := db.FetchPlayer(obj).home
	switch {
	case !valid_obj(i):
		violate(player, obj, "has its home set to an invalid object")
	case i >= 0 && !IsRoom(i):
		violate(player, obj, "has its home set to a non-room object")
	}
}

func check_program(player, obj dbref) {
}

func check_contents_list(dbref player, dbref obj) {
	if TYPEOF(obj) != TYPE_PROGRAM && TYPEOF(obj) != TYPE_EXIT {
		var i dbref
		var limit int
		for i, limit = db.Fetch(obj).Contents, db_top - 1; valid_obj(i) && limit && db.Fetch(i).Location == obj && TYPEOF(i) != TYPE_EXIT; i = db.Fetch(i).next {
			limit--
		}
		if i != NOTHING {
			switch {
			case limit > 0:
				check_next_chain(player, db.Fetch(obj).Contents)
				violate(player, obj, "is the containing object, and has the loop in its contents chain")
			case !valid_obj(i):
				violate(player, obj, "has an invalid object in its contents list")
			default:
				if TYPEOF(i) == TYPE_EXIT {
					violate(player, obj, "has an exit in its contents list (it shoudln't)")
				}
				if db.Fetch(i).Location != obj {
					violate(player, obj, "has an object in its contents lists that thinks it is located elsewhere")
				}
			}
		}
	} else {
		if db.Fetch(obj).Contents != NOTHING {
			if TYPEOF(obj) == TYPE_EXIT {
				violate(player, obj, "is an exit/action whose contents aren't #-1")
			} else {
				violate(player, obj, "is a program whose contents aren't #-1")
			}
		}
	}
}

func check_exits_list(player, obj dbref) {
	if TYPEOF(obj) != TYPE_PROGRAM && TYPEOF(obj) != TYPE_EXIT {
		var i dbref
		var limit int
		for i, limit = db.Fetch(obj).Exits, db_top - 1; valid_obj(i) && limit && db.Fetch(i).Location == obj && TYPEOF(i) == TYPE_EXIT; i = db.Fetch(i).next) {
			limit--
		}
		if i != NOTHING {
			switch {
			case limit > 0:
				check_next_chain(player, db.Fetch(obj).Contents)
				violate(player, obj, "is the containing object, and has the loop in its exits chain")
			case !valid_obj(i):
				violate(player, obj, "has an invalid object in it's exits list")
			default:
				if TYPEOF(i) != TYPE_EXIT {
					violate(player, obj, "has a non-exit in it's exits list")
				}
				if db.Fetch(i).Location != obj {
					violate(player, obj, "has an exit in its exits lists that thinks it is located elsewhere")
				}
			}
		}
	} else {
		if db.Fetch(obj).Exits != NOTHING {
			if TYPEOF(obj) == TYPE_EXIT {
				violate(player, obj, "is an exit/action whose exits list isn't #-1");
			} else {
				violate(player, obj, "is a program whose exits list isn't #-1");
			}
		}
	}
}

func check_object(player, obj dbref) {
	if !db.Fetch(obj).name {
		violate(player, obj, "doesn't have a name")
	}

	switch {
	case !valid_obj(db.Fetch(obj).Owner):
		violate(player, obj, "has an invalid object as its owner.")
	case TYPEOF(db.Fetch(obj).Owner) != TYPE_PLAYER:
		violate(player, obj, "has a non-player object as its owner.")
	}

	//	check location 
	if !valid_obj(db.Fetch(obj).Location) && !(obj == GLOBAL_ENVIRONMENT && db.Fetch(obj).Location == NOTHING) {
		violate(player, obj, "has an invalid object as it's location")
	}

	if db.Fetch(obj).Location != NOTHING && (TYPEOF(db.Fetch(obj).Location) == TYPE_EXIT || TYPEOF(db.Fetch(obj).Location) == TYPE_PROGRAM) {
		violate(player, obj, "thinks it is located in a non-container object")
	}

	check_contents_list(player, obj)
	check_exits_list(player, obj)

	switch TYPEOF(obj) {
	case TYPE_ROOM:
		check_room(player, obj)
	case TYPE_PLAYER:
		check_player(player, obj)
	case TYPE_EXIT:
		check_exit(player, obj)
	case TYPE_PROGRAM:
		check_program(player, obj)
	default:
		violate(player, obj, "has an unknown object type, and its flags may also be corrupt")
	}
}

func sanity(dbref player) {
	if player > NOTHING && player != GOD {
		notify(player, "Permission Denied.")
	} else {
		SanityViolated = false
		increp := 10000
		EachObject(func(obj dbref) {
			if obj % increp == 0 {
				j := obj + increp - 1
				if j >= db_top {
					j = db_top - 1
				}
				SanPrint(player, "Checking objects %d to %d...", obj, j)
				if player >= 0 {
					flush_user_output(player)
				}
			}
			check_object(player, obj)
		})
		find_orphan_objects(player)
		SanPrint(player, "Done.")
	}
}

func san_fixed_log(format string, unparse bool, ref1, ref2 dbref) {
	if unparse {
		var buf1, buf2 string
		if ref1 >= 0 {
			buf1 = unparse_object(GOD, ref1)
		}
		if ref2 >= 0 {
			buf2 = unparse_object(GOD, ref2)
		}
		log2file("logs/sanfixed", format, buf1, buf2)
	} else {
		log2file("logs/sanfixed", format, ref1, ref2)
	}
}

func cut_all_chains(obj dbref) {
	if db.Fetch(obj).Contents != NOTHING {
		SanFixed(obj, "Cleared contents of %s")
		db.Fetch(obj).Contents = NOTHING
		db.Fetch(obj).flags |= OBJECT_CHANGED
	}
	if db.Fetch(obj).Exits != NOTHING {
		SanFixed(obj, "Cleared exits of %s")
		db.Fetch(obj).Exits = NOTHING
		db.Fetch(obj).flags |= OBJECT_CHANGED
	}
}

func cut_bad_contents(obj dbref) {
	prev := NOTHING
	for loop := db.Fetch(obj).Contents; loop != NOTHING; loop = db.Fetch(loop).next {
		if !valid_obj(loop) || db.Fetch(loop).flags & SANEBIT || TYPEOF(loop) == TYPE_EXIT || db.Fetch(loop).Location != obj || loop == obj {
			switch {
			case !valid_obj(loop):
				SanFixed(obj, "Contents chain for %s cut at invalid dbref")
			case TYPEOF(loop) == TYPE_EXIT:
				SanFixed2(obj, loop, "Contents chain for %s cut at exit %s")
			case loop == obj:
				SanFixed(obj, "Contents chain for %s cut at self-reference")
			case db.Fetch(loop).Location != obj:
				SanFixed2(obj, loop, "Contents chain for %s cut at misplaced object %s")
			case db.Fetch(loop).flags & SANEBIT:
				SanFixed2(obj, loop, "Contents chain for %s cut at already chained object %s")
			default:
				SanFixed2(obj, loop, "Contents chain for %s cut at %s")
			}
			if prev != NOTHING {
				db.Fetch(prev).next = NOTHING
				db.Fetch(prev).flags |= OBJECT_CHANGED
			} else {
				db.Fetch(obj).Contents = NOTHING
				db.Fetch(obj).flags |= OBJECT_CHANGED
			}
			return
		}
		db.Fetch(loop).flags |= SANEBIT
		prev = loop
	}
}

func cut_bad_exits(obj dbref) {
	prev := NOTHING;
	for loop := db.Fetch(obj).Exits; loop != NOTHING; loop = db.Fetch(loop).next {
		if !valid_obj(loop) || db.Fetch(loop).flags & SANEBIT || TYPEOF(loop) != TYPE_EXIT || db.Fetch(loop).Location != obj {
			switch {
			case !valid_obj(loop):
				SanFixed(obj, "Exits chain for %s cut at invalid dbref")
			case TYPEOF(loop) != TYPE_EXIT:
				SanFixed2(obj, loop, "Exits chain for %s cut at non-exit %s")
			case db.Fetch(loop).Location != obj:
				SanFixed2(obj, loop, "Exits chain for %s cut at misplaced exit %s")
			case db.Fetch(loop).flags & SANEBIT:
				SanFixed2(obj, loop, "Exits chain for %s cut at already chained exit %s")
			default:
				SanFixed2(obj, loop, "Exits chain for %s cut at %s")
			}
			if prev != NOTHING {
				db.Fetch(prev).next = NOTHING
				db.Fetch(prev).flags |= OBJECT_CHANGED
			} else {
				db.Fetch(obj).Exits = NOTHING
				db.Fetch(obj).flags |= OBJECT_CHANGED
			}
			return
		}
		db.Fetch(loop).flags |= SANEBIT
		prev = loop
	}
}

func hacksaw_bad_chains() {
	EachObject(func(obj dbref) {
		if !IsRoom(obj) && !IsThing(obj) && !IsPlayer(obj) {
			cut_all_chains(obj)
		} else {
			cut_bad_contents(obj)
			cut_bad_exits(obj)
		}
	})
}

var pwdchars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";
func rand_password() (password string) {
	charslen := len(pwdchars)
	for loop := 0; loop < 16; loop++ {
		password = append(password, pwdchars[(RANDOM() >> 8) % charslen])
	}
	return
}

func create_lostandfound(player, room *dbref) {
	player_name = "lost+found"
	*room = new_object()
	db.Fetch(*room).name = "lost+found"
	db.Fetch(*room).Location = GLOBAL_ENVIRONMENT
	db.Fetch(*room).Exits = NOTHING
	db.Fetch(*room).sp = NOTHING
	db.Fetch(*room).flags = TYPE_ROOM | SANEBIT

	db.Fetch(*room).next = db.Fetch(GLOBAL_ENVIRONMENT.Contents)
	db.Fetch(*room).flags |= OBJECT_CHANGED
	db.Fetch(GLOBAL_ENVIRONMENT.Contents) = *room

	SanFixed(*room, "Using %s to resolve unknown location")
	for i := 1; lookup_player(player_name) != NOTHING; i++ {
		player_name = fmt.Sprintf("lost+found%d", i)
	}
	*player = new_object()
	db.Fetch(*player).name = player_name
	db.Fetch(*player).Location = *room
	db.Fetch(*player).flags = TYPE_PLAYER | SANEBIT
	db.Fetch(*player).Owner = *player
	*player = &Player{ home: *room, exits: NOTHING, curr_prog: NOTHING }
	add_property(*player, MESGPROP_VALUE, NULL, tp_start_pennies)
	rpass := rand_password()
	set_password(*player, rpass)
	db.Fetch(*player).next = db.Fetch(*room).Contents
	db.Fetch(*player).flags |= OBJECT_CHANGED
	db.Fetch(*room).Contents = *player
	db.Fetch(*player).flags |= OBJECT_CHANGED
	add_player(*player)
	log2file("logs/sanfixed", "Using %s (with password %s) to resolve unknown owner", unparse_object(GOD, *player), rpass)
	db.Fetch(*room).Owner = *player
	db.Fetch(*room).flags |= OBJECT_CHANGED
	db.Fetch(*player).flags |= OBJECT_CHANGED
	db.Fetch(GLOBAL_ENVIRONMENT).flags |= OBJECT_CHANGED
}

func fix_room(obj dbref) {
	p := db.Fetch(obj)
	switch i := p.(dbref); {
	case !valid_reference(i) && i != NOTHING && i != HOME:
		SanFixed(obj, "Removing invalid drop-to from %s")
		p.sp = NOTHING
		p.flags |= OBJECT_CHANGED
	case i >= 0 && !IsThing(i) && !IsRoom(i):
		SanFixed2(obj, i, "Removing drop-to on %s to %s")
		p.sp = NOTHING
		p.flags |= OBJECT_CHANGED
	}
}

func fix_thing(obj dbref) {
	p := db.Fetch(obj).(Object)
	if i := p.home; !valid_obj(i) || (!IsRoom(i) && !IsThing(i) && !IsPlayer(i)) {
		SanFixed2(obj, p.Owner, "Setting the home on %s to %s, it's owner")
		p.home = p.Owner
		p.flags |= OBJECT_CHANGED
	}
}

func fix_exit(obj dbref) {
	dest := db.Fetch(obj).(Exit).Destinations
	l := len(dest)
	for i := 0; i < l; {
		if o := valid_obj_or_home(dest[i], false); o == NOTHING {
			SanFixed(obj, "Removing invalid destination from %s")
			db.Fetch(obj).flags |= OBJECT_CHANGED
			for j := i; j < l; j++ {
				dest[j:] = dest[j + 1:]
			}
			l--
		} else {
			i++
		}
	}
	if len(dest) > l * 1.25 {
		d := make([]dbref, l, l * 1.25)
		copy(d, dest)
		db.Fetch(obj).(Exit).Destinations = d
	} else {
		for ol := len(dest); ol > l; ol-- {
			dest[ol] = nil
		}
		dest = dest[:l]
	}
}

func fix_player(obj dbref) {
	p := db.FetchPlayer(obj)
	if i := p.home; !valid_obj(i) || IsRoom(i) {
		SanFixed2(obj, tp_player_start, "Setting the home on %s to %s")
		p.home = tp_player_start
		p.flags |= OBJECT_CHANGED
	}
}

func find_misplaced_objects() {
	player := NOTHING
	var room dbref
	EachObject(func(obj dbref, o *Object) {
		if !IsRoom(obj) && !IsThing(obj) && !IsPlayer(obj) && !IsExit(obj) && !IsProgram(obj) {
			SanFixedRef(loop, "Object #%d is of unknown type")
			SanityViolated = true
		} else {
			if o.name == "" {
				if IsPlayer(obj) {
					name := "Unnamed"
					for i := 1; lookup_player(name) != NOTHING; i++ {
						name = fmt.Sprintf("Unnamed%d", i)
					}
					o.name = name
					add_player(obj)
				} else {
					o.name = "Unnamed"
				}
				SanFixed(obj, "Gave a name to %s")
				o.flags |= OBJECT_CHANGED
			}
			if !valid_obj(o.Owner) || !IsPlayer(o.Owner) {
				if player == NOTHING {
					create_lostandfound(&player, &room)
				}
				SanFixed2(obj, player, "Set owner of %s to %s")
				o.Owner = player
				o.flags |= OBJECT_CHANGED
			}
		
			if obj != GLOBAL_ENVIRONMENT && !valid_obj(o.Location) || IsExit(o.Location) || IsProgram(o.Location) || (IsPlayer(obj) && IsPlayer(o.Location)) {
				if IsPlayer(obj) {
					if valid_obj(o.Location) && IsPlayer(o.Location) {
						loc := o.Location
						if loc.Contents == obj {
							loc.Contents = o.next
							loc.flags |= OBJECT_CHANGED
						} else {
							for contents := loc.Contents; contents != NOTHING; contents = contents.next {
								if contents.next == obj {
									contents.next = o.next
									contents.flags |= OBJECT_CHANGED
									break
								}
							}
						}
					}
					o.Location = tp_player_start
				} else {
					if player == NOTHING {
						create_lostandfound(&player, &room)
					}
					o.Location = room
				}
				o.flags |= OBJECT_CHANGED
				db.Fetch(o.Location).flags |= OBJECT_CHANGED
				if IsExit(obj) {
					o.next = db.Fetch(o.Location).Exits
					o.flags |= OBJECT_CHANGED
					db.Fetch(o.Location).Exits = obj
				} else {
					o.next = db.Fetch(o.Location).Contents
					o.flags |= OBJECT_CHANGED
					db.Fetch(o.Location).Contents = obj
				}
				o.flags |= SANEBIT
				SanFixed2(obj, o.Location, "Set location of %s to %s")
			}
			switch {
			case IsRoom(obj):
				fix_room(obj)
			case IsThing(obj):
				fix_thing(obj)
			case IsPlayer(obj):
				fix_player(obj)
			case IsExit(obj):
				fix_exit(obj)
			}
		}
	})
}

func adopt_orphans() {
	EachObject(func(obj dbref, o *Object) (done bool) {
		if o.flags & SANEBIT == 0 {
			o.flags |= OBJECT_CHANGED
			switch TYPEOF(loop) {
			case IsRoom(obj), IsThing(obj), IsPlayer(obj), IsProgram(obj):
				o.next = db.Fetch(o.Location).Contents
				db.Fetch(o.Location).Contents = obj
				SanFixed2(loop, o.Location, "Orphaned object %s added to contents of %s")
				done = true
			case IsExit(obj):
				o.next = db.Fetch(o.Location).Exits
				db.Fetch(o.Location).Exits = loop
				SanFixed2(loop, o.Location, "Orphaned exit %s added to exits of %s")
				done = true
			default:
				SanityViolated = true
				done = true
			}
		}
		return
	})
}

func clean_global_environment() {
	if db.Fetch(GLOBAL_ENVIRONMENT).next != NOTHING {
		SanFixed(GLOBAL_ENVIRONMENT, "Removed the global environment %s from a chain")
		db.Fetch(GLOBAL_ENVIRONMENT).next = NOTHING
		db.Fetch(GLOBAL_ENVIRONMENT).flags |= OBJECT_CHANGED
	}
	if db.Fetch(GLOBAL_ENVIRONMENT).Location != NOTHING {
		SanFixed2(GLOBAL_ENVIRONMENT, db.Fetch(GLOBAL_ENVIRONMENT).Location, "Removed the global environment %s from %s")
		db.Fetch(GLOBAL_ENVIRONMENT).Location = NOTHING
		db.Fetch(GLOBAL_ENVIRONMENT).flags |= OBJECT_CHANGED
	}
}

func sanfix(player dbref) {
	if player > NOTHING && player != GOD {
		notify(player, "Yeah right!  With a psyche like yours, you think theres any hope of getting your sanity fixed?")
		return
	}

	SanityViolated = false
	EachObject(func(o *Object) {
		obj.flags &= ~SANEBIT
	})
	db.Fetch(GLOBAL_ENVIRONMENT).flags |= SANEBIT

	if !valid_obj(tp_player_start) || TYPEOF(tp_player_start) != TYPE_ROOM {
		SanFixed(GLOBAL_ENVIRONMENT, "Reset invalid player_start to %s")
		tp_player_start = GLOBAL_ENVIRONMENT
	}

	hacksaw_bad_chains()
	find_misplaced_objects()
	adopt_orphans()
	clean_global_environment()

	EachObject(func(o *Object) {
		o.flags &= ~SANEBIT
	})

	if player > NOTHING {
		if !SanityViolated {
			notify_nolisten(player, "Database repair complete, please re-run @sanity.  For details of repairs, check logs/sanfixed.", true)
		} else {
			notify_nolisten(player, "Database repair complete, however the database is still corrupt.  Please re-run @sanity.", true)
		}
	} else {
		fprintf(stderr, "Database repair complete, ")
		if !SanityViolated {
			fprintf(stderr, "please re-run sanity check.\n")
		} else {}
			fprintf(stderr, "however the database is still corrupt.\n Please re-run sanity check for details and fix it by hand.\n")
		}
		fprintf(stderr, "For details of repairs made, check logs/sanfixed.\n")
	}
	if SanityViolated {
		log2file("logs/sanfixed", "WARNING: The database is still corrupted, please repair by hand")
	}
}

char cbuf[1000];
var buf2 string

func sanechange(player dbref, command string) {
	var field, which, value string
	var results int
	var d, v dbref
	if player > NOTHING {
		if player != GOD {
			notify(player, "Only GOD may alter the basic structure of the universe!")
			return
		}
		results = sscanf(command, "%s %s %s", which, field, value)
		sscanf(which, "#%d", &d)
		sscanf(value, "#%d", &v)
	} else {
		results = sscanf(command, "%s %s %s", which, field, value)
		sscanf(which, "%d", &d)
		sscanf(value, "%d", &v)
	}
	if results != 3 {
		d = v = 0
		field = "help"
	}
	if !valid_reference(d) {
		SanPrint(player, "## %d is an invalid dbref.", d)
		return
	}

	buf2 = ""
	switch field {
	case "next":
		p := db.Fetch(d)
		buf2 = unparse_object(GOD, p.next))
		p.next = v
		p.flags |= OBJECT_CHANGED
		SanPrint(player, "## Setting #%d's next field to %s", d, unparse_object(GOD, v))
	case "exits":
		p := db.Fetch(d)
		buf2 = unparse_object(GOD, p.Exits)
		p.Exits = v
		p.flags |= OBJECT_CHANGED
		SanPrint(player, "## Setting #%d's Exits list start to %s", d, unparse_object(GOD, v))
	case "contents":
		p := db.Fetch(d)
		buf2 = unparse_object(GOD, p.Contents)
		p.Contents = v
		p.flags |= OBJECT_CHANGED
		SanPrint(player, "## Setting #%d's Contents list start to %s", d, unparse_object(GOD, v))
	case "location":
		p := db.Fetch(d)
		buf2 = unparse_object(GOD, d.Location)
		d.Location = v
		d.flags |= OBJECT_CHANGED
		SanPrint(player, "## Setting #%d's location to %s", d, unparse_object(GOD, v))
	case "owner":
		p := db.Fetch(d)
		buf2 = unparse_object(GOD, p.Owner)
		p.Owner = v
		p.flags |= OBJECT_CHANGED
		SanPrint(player, "## Setting #%d's owner to %s", d, unparse_object(GOD, v))
	case "home":
		var ip *int
		p := db.Fetch(d)
		switch p := p.(type) {
		case Player, Object:
			ip = &(p.home)
		default:
			fmt.Printf("%s has no home to set.\n", unparse_object(GOD, d))
			return
		}
		buf2 = unparse_object(GOD, *ip)
		*ip = v
		p.flags |= OBJECT_CHANGED
		printf("Setting home to: %s\n", unparse_object(GOD, v))
	default:
		if player > NOTHING {
			notify(player, "@sanchange <dbref> <field> <object>")
		} else {
			SanPrint(player, "change command help:")
			SanPrint(player, "c <dbref> <field> <object>")
		}
		SanPrint(player, "Fields are:     exits       Start of Exits list.")
		SanPrint(player, "                contents    Start of Contents list.")
		SanPrint(player, "                next        Next object in list.")
		SanPrint(player, "                location    Object's Location.")
		SanPrint(player, "                home        Object's Home.")
		SanPrint(player, "                owner       Object's Owner.")
		return
	}

	if buf2 != "" {
		SanPrint(player, "## Old value was %s", buf2)
	}
}

func extract_prop(f *FILE, dir string, p *Plist) {
	if _, ok := p.data.(PROP_DIRTYP); !ok {
		buf := dir[1:] + p.key + PROP_DELIMITER + intostr(PropFlagsRaw(p)) + PROP_DELIMITER
		switch v := p.data.(type) {
		case int:
			if v != 0 {
				buf += intostr(v)
			}
		case float64:
			if v != 0 {
				buf += fmt.Sprintf("%.17g", v)
			}
		case dbref:
			if v != NOTHING {
				buf += intostr(v)
			}
		case string:
			if v != "" {
				buf += v
			}
		case Lock:		// FIXME: lock
			if !v.IsTrue() {
				buf += v.Unparse(1, false)
			}
		}
		if buf != "" {
			if _, err := f.WriteString(buf + '\n'); err != nil {
				fmt.Fprintf(os.Stderr, "extract_prop(): %v failed write!\n", err)
				abort()
			}
		}
	}
}

func extract_props_rec(f *FILE, obj dbref, dir string, p *Plist) {
	if p != nil {
		extract_props_rec(f, obj, dir, p.left)
		extract_prop(f, dir, p)
		if p.dir != nil {
			extract_props_rec(f, obj, fmt.Sprint(dir, p.key, PROPDIR_DELIMITER), p.dir)
		}
		extract_props_rec(f, obj, dir, p.right())
	}
}

func extract_props(f *os.File, obj dbref) {
	extract_props_rec(f, obj, "/", db.Fetch(obj).properties)
}

func extract_program(f *os.File, obj dbref) {
	if pf, err := os.Open(fmt.Sprintf("muf/%v.m", obj)); err != nil {
		log.Fatal(err)
		fmt.Fprintf(f, "  (No listing found)\n")
	} else {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		c := 0
		for scanner.Scan() {
			c++
			fmt.Fprintf(f, "%s\n", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		} else {
			fmt.Fprintf(f, "  End of program listing (%v lines)\n", c)
		}
	}
}

func extract_object(f *FILE, d dbref) {
	p := db.Fetch(d)
	fmt.Fprintf(f, "  #%d\n", d)
	fmt.Fprintf(f, "  Object:         %s\n", unparse_object(GOD, d))
	fmt.Fprintf(f, "  Owner:          %s\n", unparse_object(GOD, p.Owner))
	fmt.Fprintf(f, "  Location:       %s\n", unparse_object(GOD, p.Location))
	fmt.Fprintf(f, "  Contents Start: %s\n", unparse_object(GOD, p.Contents))
	fmt.Fprintf(f, "  Exits Start:    %s\n", unparse_object(GOD, p.Exits))
	fmt.Fprintf(f, "  Next:           %s\n", unparse_object(GOD, p.next))

	switch TYPEOF(d) {
	case Object:
		fmt.Fprintf(f, "  Home:           %s\n", unparse_object(GOD, p.home))
		fmt.Fprintf(f, "  Value:          %d\n", get_property_value(d, MESGPROP_VALUE))
	case Room:
		fmt.Fprintf(f, "  Drop-to:        %s\n", unparse_object(GOD, p.dbref))
	case Player:
		fmt.Fprintf(f, "  Home:           %s\n", unparse_object(GOD, p.home))
		fmt.Fprintf(f, "  Pennies:        %d\n", get_property_value(d, MESGPROP_VALUE))
	case Exit:
		fmt.Fprintf(f, "  Links:         ")
		for _, v := range p.Destinations {
			fmt.Fprintf(f, " %s;", unparse_object(GOD, v))
		}
		fmt.Fprintf(f, "\n")
	case Program:
		fmt.Fprintf(f, "  Listing:\n")
		extract_program(f, d)
	}

	if p.properties {
		fmt.Fprintf(f, "  Properties:\n")
		extract_props(f, d)
	} else {
		fmt.Fprintf(f, "  No properties\n")
	}
	fmt.Fprintf(f, "\n")
}

func extract() {
	var filename string
	var d dbref
	i := sscanf(cbuf, "%*s %d %s", &d, filename)
	if !valid_obj(d) {
		printf("%d is an invalid dbref.\n", d)
	} else {
		var f *FILE
		if i == 2 {
			if f = fopen(filename, "wb"); f == nil {
				printf("Could not open file %s\n", filename);
				return
			}
			printf("Writing to file %s\n", filename)
		} else {
			f = stdout
		}
		EachObject(func(obj dbref) {
			if db.Fetch(obj).Owner == d {
				extract_object(f, obj)
			}
		})
		if f != stdout {
			fclose(f)
		}
		printf("\nDone.\n")
	}
}

func extract_single() {
	var filename string
	var d dbref
	i := sscanf(cbuf, "%*s %d %s", &d, filename)
	if !valid_obj(d) {
		printf("%d is an invalid dbref.\n", d);
	} else {
		var f *FILE
		if i == 2 {
			if f = fopen(filename, "wb"); f == nil {
				printf("Could not open file %s\n", filename)
				return;
			}
			printf("Writing to file %s\n", filename)
		} else {
			f = stdout
		}
		extract_object(f, d)
		/* extract only objects owned by this player */
		if f != stdout {
			fclose(f)
		}
		printf("\nDone.\n")
	}
}

func hack_it_up() {
	var ptr string
	do {
		printf("\nCommand: (? for help)\n")
		fgets(cbuf, sizeof(cbuf), stdin)

		switch strings.ToLower(cbuf[0]) {
		case 's':
			printf("Running Sanity...\n")
			sanity(NOTHING);
		case 'f':
			printf("Running Sanfix...\n")
			sanfix(NOTHING)
		case 'p':
			if i := strings.IndexFunc(cbuf, unicode.IsSpace); i != -1 {
				ptr = cbuf[i:]
			}
			if len(ptr) > 0 {
				ptr = ptr[1:]
			}
			sane_dump_object(NOTHING, ptr);
		case 'w':
			if sscanf(cbuf, "%*s %s", buf2); buf2 != "" {
				printf("Writing database to %s...\n", buf2)
			} else {
				printf("Writing database...\n")
			}
			do_dump(GOD, buf2)
			printf("Done.\n")
		case 'c':
			if i := strings.IndexFunc(cbuf, unicode.IsSpace); i != -1 {
				ptr = cbuf[i:]
			}
			if len(ptr) > 0 {
				ptr = ptr[1:]
			}
			sanechange(NOTHING, ptr)
		case 'x':
			extract()
		case 'y':
			extract_single()
		case 'h':
		case '?':
			printf("\n")
			printf("s                           Run Sanity checks on database\n");
			printf("f                           Automatically fix the database\n");
			printf("p <dbref>                   Print an object\n");
			printf("q                           Quit\n");
			printf("w <file>                    Write database to file.\n");
			printf("c <dbref> <field> <value>   Change a field on an object.\n");
			printf("                              (\"c ? ?\" for list)\n");
			printf("x <dbref> [<filename>]      Extract all objects belonging to <dbref>\n");
			printf("y <dbref> [<filename>]      Extract the single object <dbref>\n");
			printf("?                           Help! (Displays this screen.\n");
		}
	} while cbuf[0] != 'q'
	printf("Quitting.\n\n")
}

func san_main() {
	printf("\nEntering the Interactive Sanity DB editor.\n")
	printf("Good luck!\n\n")
	printf("Number of objects in DB is: %d\n", db_top - 1)
	printf("Global Environment is: %s\n", unparse_object(GOD, GLOBAL_ENVIRONMENT))
	printf("God is: %s\n", unparse_object(GOD, GOD))
	printf("\n")
	hack_it_up()
	printf("Exiting sanity editor...\n\n")
}