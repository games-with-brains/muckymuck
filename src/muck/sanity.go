package fbmuck

#define SanFixed(ref, fixed) san_fixed_log((fixed), 1, (ref), -1)
#define SanFixed2(ref, ref2, fixed) san_fixed_log((fixed), 1, (ref), (ref2))
#define SanFixedRef(ref, fixed) san_fixed_log((fixed), 0, (ref), -1)

var SanityViolated bool

var san_linesprinted int
func SanPrint(player ObjectID, format string, args ...interface{}) {
	switch m := fmt.Sprintf(format, args...); player {
	case NOTHING:
		fmt.Println(os.Stdout, m)
		os.Stdout.Sync()
	case AMBIGUOUS:
		log.Println(m)
	default:
		notify_nolisten(player, buf, true)
		if san_linesprinted++ > 100 {
			flush_user_output(player)
			san_linesprinted = 0
		}
	}
}

func sane_dump_object(player ObjectID, arg string) {
	var d ObjectID
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

	if !result || !d.IsValid() {
		SanPrint(player, "Invalid Object.")
		return
	}

	p := DB.Fetch(d)
	SanPrint(player, "Object:         %s", unparse_object(GOD, d))
	SanPrint(player, "  Owner:          %s", unparse_object(GOD, p.Owner))
	SanPrint(player, "  Location:       %s", unparse_object(GOD, p.Location))
	SanPrint(player, "  Contents Start: %s", unparse_object(GOD, p.Contents))
	SanPrint(player, "  Exits Start:    %s", unparse_object(GOD, p.Exits))
	SanPrint(player, "  Next:           %s", unparse_object(GOD, p.next))

	switch p := p.(type) {
	case Object:
		SanPrint(player, "  Home:           %s", unparse_object(GOD, p.Home))
		SanPrint(player, "  Value:          %d", get_property_value(d, MESGPROP_VALUE))
	case Room:
		SanPrint(player, "  Drop-to:        %s", unparse_object(GOD, p.ObjectID))
	case Player:
		SanPrint(player, "  Home:           %s", unparse_object(GOD, p.Home))
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
	EachObject(func(obj ObjectID, o *Object) {
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

func violate(player, i ObjectID, s string) {
	SanPrint(player, "Object \"%s\" %s!", unparse_object(GOD, i), s)
	SanityViolated = true
}

func valid_obj(obj ObjectID) (r bool) {
	switch {
	case obj == NOTHING, !obj.IsValid():
	case IsRoom(obj), IsExit(obj), IsPlayer(obj), IsProgram(obj), IsThing(obj):
		r = true
	}
	return
}

func check_next_chain(player, obj ObjectID) {
	orig := obj
	for obj != NOTHING && obj.IsValid() {
		for i := orig; i != NOTHING; i = DB.Fetch(i).next {
			if i == DB.Fetch(obj).next {
				violate(player, obj, "has a 'next' field that forms an illegal loop in an object chain")
				return
			}
			if i == obj {
				break
			}
		}
		obj = DB.Fetch(obj).next
	}
	if !obj.IsValid() && obj != NOTHING {
		violate(player, obj, "has an invalid object in its 'next' chain");
	}
}

func find_orphan_objects(player ObjectID) {
	SanPrint(player, "Searching for orphan objects...");
	EachObject(func(o *Object) {
		o.flags &= ~SANEBIT
	})

	DB.Fetch(GLOBAL_ENVIRONMENT).flags |= SANEBIT

	EachObject(func(obj ObjectID, o *Object) {
		if o.Exits != NOTHING {
			if p := DB.Fetch(o.Exits); p.flags & SANEBIT != 0 {
				violate(player, o.Exits, "is referred to by more than one object's Next, Contents, or Exits field")
			} else {
				p.flags |= SANEBIT
			}
		}
		if o.Contents != NOTHING {
			if p := DB.Fetch(o.Contents); p.flags & SANEBIT != 0 {
				violate(player, o.Contents, "is referred to by more than one object's Next, Contents, or Exits field")
			} else {
				p.flags |= SANEBIT
			}
		}
		if o.next != NOTHING {
			if p := DB.Fetch(o.next); p.flags & SANEBIT != 0 {
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

func check_room(player, obj ObjectID) {
	switch i := DB.Fetch(obj).(ObjectID); {
	case !i.IsValid() && v != NOTHING && i != HOME:
		violate(player, obj, "has its dropto set to an invalid object")
	case i >= 0 && !IsThing(i) && !IsRoom(i):
		violate(player, obj, "has its dropto set to a non-room, non-thing object")
	}
}

func check_exit(player, obj ObjectID) {
	for i, v := range DB.Fetch(obj).(Exit).Destinations {
		if !v.IsValid() && v != NOTHING && v != HOME {
			violate(player, obj, "has an invalid object as one of its link destinations")
		}
	}
}

func check_player(player, obj ObjectID) {
	i := DB.FetchPlayer(obj).Home
	switch {
	case !valid_obj(i):
		violate(player, obj, "has its home set to an invalid object")
	case i >= 0 && !IsRoom(i):
		violate(player, obj, "has its home set to a non-room object")
	}
}

func check_program(player, obj ObjectID) {
}

func check_contents_list(ObjectID player, ObjectID obj) {
	if TYPEOF(obj) != TYPE_PROGRAM && TYPEOF(obj) != TYPE_EXIT {
		var i ObjectID
		var limit int
		for i, limit = DB.Fetch(obj).Contents, db_top - 1; valid_obj(i) && limit && DB.Fetch(i).Location == obj && TYPEOF(i) != TYPE_EXIT; i = DB.Fetch(i).next {
			limit--
		}
		if i != NOTHING {
			switch {
			case limit > 0:
				check_next_chain(player, DB.Fetch(obj).Contents)
				violate(player, obj, "is the containing object, and has the loop in its contents chain")
			case !valid_obj(i):
				violate(player, obj, "has an invalid object in its contents list")
			default:
				if TYPEOF(i) == TYPE_EXIT {
					violate(player, obj, "has an exit in its contents list (it shoudln't)")
				}
				if DB.Fetch(i).Location != obj {
					violate(player, obj, "has an object in its contents lists that thinks it is located elsewhere")
				}
			}
		}
	} else {
		if DB.Fetch(obj).Contents != NOTHING {
			if TYPEOF(obj) == TYPE_EXIT {
				violate(player, obj, "is an exit/action whose contents aren't #-1")
			} else {
				violate(player, obj, "is a program whose contents aren't #-1")
			}
		}
	}
}

func check_exits_list(player, obj ObjectID) {
	if TYPEOF(obj) != TYPE_PROGRAM && TYPEOF(obj) != TYPE_EXIT {
		var i ObjectID
		var limit int
		for i, limit = DB.Fetch(obj).Exits, db_top - 1; valid_obj(i) && limit && DB.Fetch(i).Location == obj && TYPEOF(i) == TYPE_EXIT; i = DB.Fetch(i).next) {
			limit--
		}
		if i != NOTHING {
			switch {
			case limit > 0:
				check_next_chain(player, DB.Fetch(obj).Contents)
				violate(player, obj, "is the containing object, and has the loop in its exits chain")
			case !valid_obj(i):
				violate(player, obj, "has an invalid object in it's exits list")
			default:
				if TYPEOF(i) != TYPE_EXIT {
					violate(player, obj, "has a non-exit in it's exits list")
				}
				if DB.Fetch(i).Location != obj {
					violate(player, obj, "has an exit in its exits lists that thinks it is located elsewhere")
				}
			}
		}
	} else {
		if DB.Fetch(obj).Exits != NOTHING {
			if TYPEOF(obj) == TYPE_EXIT {
				violate(player, obj, "is an exit/action whose exits list isn't #-1");
			} else {
				violate(player, obj, "is a program whose exits list isn't #-1");
			}
		}
	}
}

func check_object(player, obj ObjectID) {
	if !DB.Fetch(obj).name {
		violate(player, obj, "doesn't have a name")
	}

	switch {
	case !valid_obj(DB.Fetch(obj).Owner):
		violate(player, obj, "has an invalid object as its owner.")
	case TYPEOF(DB.Fetch(obj).Owner) != TYPE_PLAYER:
		violate(player, obj, "has a non-player object as its owner.")
	}

	//	check location 
	if !valid_obj(DB.Fetch(obj).Location) && !(obj == GLOBAL_ENVIRONMENT && DB.Fetch(obj).Location == NOTHING) {
		violate(player, obj, "has an invalid object as it's location")
	}

	if DB.Fetch(obj).Location != NOTHING && (TYPEOF(DB.Fetch(obj).Location) == TYPE_EXIT || TYPEOF(DB.Fetch(obj).Location) == TYPE_PROGRAM) {
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

func sanity(ObjectID player) {
	if player > NOTHING && player != GOD {
		notify(player, "Permission Denied.")
	} else {
		SanityViolated = false
		increp := 10000
		EachObject(func(obj ObjectID) {
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

func san_fixed_log(format string, unparse bool, ref1, ref2 ObjectID) {
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

func cut_all_chains(obj ObjectID) {
	if DB.Fetch(obj).Contents != NOTHING {
		SanFixed(obj, "Cleared contents of %s")
		DB.Fetch(obj).Contents = NOTHING
		DB.Fetch(obj).Touch()
	}
	if DB.Fetch(obj).Exits != NOTHING {
		SanFixed(obj, "Cleared exits of %s")
		DB.Fetch(obj).Exits = NOTHING
		DB.Fetch(obj).Touch()
	}
}

func cut_bad_contents(obj ObjectID) {
	prev := NOTHING
	for loop := DB.Fetch(obj).Contents; loop != NOTHING; loop = DB.Fetch(loop).next {
		if !valid_obj(loop) || DB.Fetch(loop).flags & SANEBIT || TYPEOF(loop) == TYPE_EXIT || DB.Fetch(loop).Location != obj || loop == obj {
			switch {
			case !valid_obj(loop):
				SanFixed(obj, "Contents chain for %s cut at invalid ObjectID")
			case TYPEOF(loop) == TYPE_EXIT:
				SanFixed2(obj, loop, "Contents chain for %s cut at exit %s")
			case loop == obj:
				SanFixed(obj, "Contents chain for %s cut at self-reference")
			case DB.Fetch(loop).Location != obj:
				SanFixed2(obj, loop, "Contents chain for %s cut at misplaced object %s")
			case DB.Fetch(loop).flags & SANEBIT:
				SanFixed2(obj, loop, "Contents chain for %s cut at already chained object %s")
			default:
				SanFixed2(obj, loop, "Contents chain for %s cut at %s")
			}
			if prev != NOTHING {
				DB.Fetch(prev).next = NOTHING
				DB.Fetch(prev).Touch()
			} else {
				DB.Fetch(obj).Contents = NOTHING
				DB.Fetch(obj).Touch()
			}
			return
		}
		DB.Fetch(loop).flags |= SANEBIT
		prev = loop
	}
}

func cut_bad_exits(obj ObjectID) {
	prev := NOTHING;
	for loop := DB.Fetch(obj).Exits; loop != NOTHING; loop = DB.Fetch(loop).next {
		if !valid_obj(loop) || DB.Fetch(loop).flags & SANEBIT || TYPEOF(loop) != TYPE_EXIT || DB.Fetch(loop).Location != obj {
			switch {
			case !valid_obj(loop):
				SanFixed(obj, "Exits chain for %s cut at invalid ObjectID")
			case TYPEOF(loop) != TYPE_EXIT:
				SanFixed2(obj, loop, "Exits chain for %s cut at non-exit %s")
			case DB.Fetch(loop).Location != obj:
				SanFixed2(obj, loop, "Exits chain for %s cut at misplaced exit %s")
			case DB.Fetch(loop).flags & SANEBIT:
				SanFixed2(obj, loop, "Exits chain for %s cut at already chained exit %s")
			default:
				SanFixed2(obj, loop, "Exits chain for %s cut at %s")
			}
			if prev != NOTHING {
				DB.Fetch(prev).next = NOTHING
				DB.Fetch(prev).Touch()
			} else {
				DB.Fetch(obj).Exits = NOTHING
				DB.Fetch(obj).Touch()
			}
			return
		}
		DB.Fetch(loop).flags |= SANEBIT
		prev = loop
	}
}

func hacksaw_bad_chains() {
	EachObject(func(obj ObjectID) {
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

func create_lostandfound(player, room *ObjectID) {
	player_name = "lost+found"
	*room = new_object()
	DB.Fetch(*room).NowCalled("lost+found")
	DB.Fetch(*room).MoveTo(GLOBAL_ENVIRONMENT)
	DB.Fetch(*room).Exits = NOTHING
	DB.Fetch(*room).sp = NOTHING
	DB.Fetch(*room).flags = TYPE_ROOM | SANEBIT

	DB.Fetch(*room).next = DB.Fetch(GLOBAL_ENVIRONMENT.Contents)
	DB.Fetch(*room).Touch()
	DB.Fetch(GLOBAL_ENVIRONMENT.Contents) = *room

	SanFixed(*room, "Using %s to resolve unknown location")
	for i := 1; lookup_player(player_name) != NOTHING; i++ {
		player_name = fmt.Sprintf("lost+found%d", i)
	}
	*player = new_object()
	DB.Fetch(*player).NowCalled(player_name)
	DB.Fetch(*player).MoveTo(*room)
	DB.Fetch(*player).flags = TYPE_PLAYER | SANEBIT
	DB.Fetch(*player).GiveTo(*player)
	*player = &Player{ home: *room, exits: NOTHING, curr_prog: NOTHING }
	add_property(*player, MESGPROP_VALUE, NULL, tp_start_pennies)
	rpass := rand_password()
	set_password(*player, rpass)
	DB.Fetch(*player).next = DB.Fetch(*room).Contents
	DB.Fetch(*player).Touch()
	DB.Fetch(*room).Contents = *player
	DB.Fetch(*player).Touch()
	add_player(*player)
	log2file("logs/sanfixed", "Using %s (with password %s) to resolve unknown owner", unparse_object(GOD, *player), rpass)
	DB.Fetch(*room).GiveTo(*player)
	DB.Fetch(*room).Touch()
	DB.Fetch(*player).Touch()
	DB.Fetch(GLOBAL_ENVIRONMENT).Touch()
}

func fix_room(obj ObjectID) {
	p := DB.Fetch(obj)
	switch i := p.(ObjectID); {
	case !i.IsValid() && i != NOTHING && i != HOME:
		SanFixed(obj, "Removing invalid drop-to from %s")
		p.sp = NOTHING
		p.Touch()
	case i >= 0 && !IsThing(i) && !IsRoom(i):
		SanFixed2(obj, i, "Removing drop-to on %s to %s")
		p.sp = NOTHING
		p.Touch()
	}
}

func fix_thing(obj ObjectID) {
	p := DB.Fetch(obj).(Object)
	if i := p.Home; !valid_obj(i) || (!IsRoom(i) && !IsThing(i) && !IsPlayer(i)) {
		SanFixed2(obj, p.Owner, "Setting the home on %s to %s, it's owner")
		p.LiveAt(p.Owner)
		p.Touch()
	}
}

func fix_exit(obj ObjectID) {
	dest := DB.Fetch(obj).(Exit).Destinations
	l := len(dest)
	for i := 0; i < l; {
		if o := valid_obj_or_home(dest[i], false); o == NOTHING {
			SanFixed(obj, "Removing invalid destination from %s")
			DB.Fetch(obj).Touch()
			for j := i; j < l; j++ {
				dest[j:] = dest[j + 1:]
			}
			l--
		} else {
			i++
		}
	}
	if len(dest) > l * 1.25 {
		d := make([]ObjectID, l, l * 1.25)
		copy(d, dest)
		DB.Fetch(obj).(Exit).Destinations = d
	} else {
		for ol := len(dest); ol > l; ol-- {
			dest[ol] = nil
		}
		dest = dest[:l]
	}
}

func fix_player(obj ObjectID) {
	p := DB.FetchPlayer(obj)
	if i := p.Home; !valid_obj(i) || IsRoom(i) {
		SanFixed2(obj, tp_player_start, "Setting the home on %s to %s")
		p.LiveAt(tp_player_start)
		p.Touch()
	}
}

func find_misplaced_objects() {
	player := NOTHING
	var room ObjectID
	EachObject(func(obj ObjectID, o *Object) {
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
					o.NowCalled(name)
					add_player(obj)
				} else {
					o.NowCalled("Unnamed")
				}
				SanFixed(obj, "Gave a name to %s")
				o.Touch()
			}
			if !valid_obj(o.Owner) || !IsPlayer(o.Owner) {
				if player == NOTHING {
					create_lostandfound(&player, &room)
				}
				SanFixed2(obj, player, "Set owner of %s to %s")
				o.GiveTo(player)
				o.Touch()
			}
		
			if obj != GLOBAL_ENVIRONMENT && !valid_obj(o.Location) || IsExit(o.Location) || IsProgram(o.Location) || (IsPlayer(obj) && IsPlayer(o.Location)) {
				if IsPlayer(obj) {
					if valid_obj(o.Location) && IsPlayer(o.Location) {
						loc := o.Location
						if loc.Contents == obj {
							loc.Contents = o.next
							loc.Touch
						} else {
							for contents := loc.Contents; contents != NOTHING; contents = contents.next {
								if contents.next == obj {
									contents.next = o.next
									contents.Touch()
									break
								}
							}
						}
					}
					o.MoveTo(tp_player_start)
				} else {
					if player == NOTHING {
						create_lostandfound(&player, &room)
					}
					o.MoveTo(room)
				}
				if IsExit(obj) {
					o.next = DB.Fetch(o.Location).Exits
					DB.Fetch(o.Location).Exits = obj
				} else {
					o.next = DB.Fetch(o.Location).Contents
					DB.Fetch(o.Location).Contents = obj
				}
				DB.Fetch(o.Location).Touch()
				o.flags |= SANEBIT
				o.Touch()
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
	EachObject(func(obj ObjectID, o *Object) (done bool) {
		if o.flags & SANEBIT == 0 {
			o.Touch()
			switch TYPEOF(loop) {
			case IsRoom(obj), IsThing(obj), IsPlayer(obj), IsProgram(obj):
				o.next = DB.Fetch(o.Location).Contents
				DB.Fetch(o.Location).Contents = obj
				SanFixed2(loop, o.Location, "Orphaned object %s added to contents of %s")
				done = true
			case IsExit(obj):
				o.next = DB.Fetch(o.Location).Exits
				DB.Fetch(o.Location).Exits = loop
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
	if DB.Fetch(GLOBAL_ENVIRONMENT).next != NOTHING {
		SanFixed(GLOBAL_ENVIRONMENT, "Removed the global environment %s from a chain")
		DB.Fetch(GLOBAL_ENVIRONMENT).next = NOTHING
		DB.Fetch(GLOBAL_ENVIRONMENT).Touch()
	}
	if DB.Fetch(GLOBAL_ENVIRONMENT).Location != NOTHING {
		SanFixed2(GLOBAL_ENVIRONMENT, DB.Fetch(GLOBAL_ENVIRONMENT).Location, "Removed the global environment %s from %s")
		DB.Fetch(GLOBAL_ENVIRONMENT).MoveTo(NOTHING)
		DB.Fetch(GLOBAL_ENVIRONMENT).Touch()
	}
}

func sanfix(player ObjectID) {
	if player > NOTHING && player != GOD {
		notify(player, "Yeah right!  With a psyche like yours, you think theres any hope of getting your sanity fixed?")
		return
	}

	SanityViolated = false
	EachObject(func(o *Object) {
		obj.flags &= ~SANEBIT
	})
	DB.Fetch(GLOBAL_ENVIRONMENT).flags |= SANEBIT

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
		log.Print("Database repair complete, ")
		if !SanityViolated {
			log.Println("please re-run sanity check.")
		} else {}
			log.Println("however the database is still corrupt.\n Please re-run sanity check for details and fix it by hand.")
		}
		log.Println("For details of repairs made, check logs/sanfixed.")
	}
	if SanityViolated {
		log2file("logs/sanfixed", "WARNING: The database is still corrupted, please repair by hand")
	}
}

char cbuf[1000];
var buf2 string

func sanechange(player ObjectID, command string) {
	var field, which, value string
	var results int
	var d, v ObjectID
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
	if !d.IsValid() {
		SanPrint(player, "## %d is an invalid ObjectID.", d)
		return
	}

	buf2 = ""
	switch field {
	case "next":
		p := DB.Fetch(d)
		buf2 = unparse_object(GOD, p.next))
		p.next = v
		p.Touch()
		SanPrint(player, "## Setting #%d's next field to %s", d, unparse_object(GOD, v))
	case "exits":
		p := DB.Fetch(d)
		buf2 = unparse_object(GOD, p.Exits)
		p.Exits = v
		p.Touch()
		SanPrint(player, "## Setting #%d's Exits list start to %s", d, unparse_object(GOD, v))
	case "contents":
		p := DB.Fetch(d)
		buf2 = unparse_object(GOD, p.Contents)
		p.Contents = v
		p.Touch()
		SanPrint(player, "## Setting #%d's Contents list start to %s", d, unparse_object(GOD, v))
	case "location":
		p := DB.Fetch(d)
		buf2 = unparse_object(GOD, d.Location)
		d.MoveTo(v)
		d.Touch()
		SanPrint(player, "## Setting #%d's location to %s", d, unparse_object(GOD, v))
	case "owner":
		p := DB.Fetch(d)
		buf2 = unparse_object(GOD, p.Owner)
		p.GiveTo(v)
		p.Touch()
		SanPrint(player, "## Setting #%d's owner to %s", d, unparse_object(GOD, v))
	case "home":
		var ip *int
		p := DB.Fetch(d)
		switch p := p.(type) {
		case Player, Object:
			ip = &(p.Home)
		default:
			fmt.Printf("%s has no home to set.\n", unparse_object(GOD, d))
			return
		}
		buf2 = unparse_object(GOD, *ip)
		*ip = v
		p.Touch()
		printf("Setting home to: %s\n", unparse_object(GOD, v))
	default:
		if player > NOTHING {
			notify(player, "@sanchange <ObjectID> <field> <object>")
		} else {
			SanPrint(player, "change command help:")
			SanPrint(player, "c <ObjectID> <field> <object>")
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
		case ObjectID:
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

func extract_props_rec(f *FILE, obj ObjectID, dir string, p *Plist) {
	if p != nil {
		extract_props_rec(f, obj, dir, p.left)
		extract_prop(f, dir, p)
		if p.dir != nil {
			extract_props_rec(f, obj, fmt.Sprint(dir, p.key, PROPDIR_DELIMITER), p.dir)
		}
		extract_props_rec(f, obj, dir, p.right())
	}
}

func extract_props(f *os.File, obj ObjectID) {
	extract_props_rec(f, obj, "/", DB.Fetch(obj).Properties())
}

func extract_program(f *os.File, obj ObjectID) {
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

func extract_object(f *FILE, d ObjectID) {
	p := DB.Fetch(d)
	fmt.Fprintf(f, "  #%d\n", d)
	fmt.Fprintf(f, "  Object:         %s\n", unparse_object(GOD, d))
	fmt.Fprintf(f, "  Owner:          %s\n", unparse_object(GOD, p.Owner))
	fmt.Fprintf(f, "  Location:       %s\n", unparse_object(GOD, p.Location))
	fmt.Fprintf(f, "  Contents Start: %s\n", unparse_object(GOD, p.Contents))
	fmt.Fprintf(f, "  Exits Start:    %s\n", unparse_object(GOD, p.Exits))
	fmt.Fprintf(f, "  Next:           %s\n", unparse_object(GOD, p.next))

	switch TYPEOF(d) {
	case Object:
		fmt.Fprintf(f, "  Home:           %s\n", unparse_object(GOD, p.Home))
		fmt.Fprintf(f, "  Value:          %d\n", get_property_value(d, MESGPROP_VALUE))
	case Room:
		fmt.Fprintf(f, "  Drop-to:        %s\n", unparse_object(GOD, p.ObjectID))
	case Player:
		fmt.Fprintf(f, "  Home:           %s\n", unparse_object(GOD, p.Home))
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

	if p.Properties() != nil {
		fmt.Fprintf(f, "  Properties:\n")
		extract_props(f, d)
	} else {
		fmt.Fprintf(f, "  No properties\n")
	}
	fmt.Fprintf(f, "\n")
}

func extract() {
	var filename string
	var player ObjectID
	i := sscanf(cbuf, "%*s %d %s", &d, filename)
	if !valid_obj(player) {
		fmt.Printf("%d is an invalid ObjectID.\n", player)
	} else {
		if i == 2 {
			if f, e := os.OpenFile(filename, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0755); e != nil {
				fmt.Printf("Writing to file %s\n", filename)
				EachObject(func(obj ObjectID) {
					if DB.Fetch(obj).Owner == player {
						extract_object(f, obj)
					}
				})
				f.Close()
			} else {
				fmt.Println("Could not open file", filename)
				return
			}
		} else {
			EachObject(func(obj ObjectID) {
				if DB.Fetch(obj).Owner == player {
					extract_object(os.Stdout, obj)
				}
			})
		}
		printf("\nDone.\n")
	}
}

func extract_single() {
	var filename string
	var player ObjectID
	i := sscanf(cbuf, "%*s %d %s", &player, &filename)
	if !valid_obj(player) {
		fmt.Printf("%d is an invalid ObjectID.\n", player)
	} else {
		if i == 2 {
			if f, e := os.OpenFile(filename, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0755); e != nil {
				fmt.Println("Writing to file", filename)
				extract_object(f, player)
				f.Close()
			} else {
				fmt.Println("Could not open file", filename)
				return
			}
		} else {
			extract_object(os.Stdout, player)
		}
		fmt.Printf("\nDone.\n")
	}
}

func hack_it_up() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\nCommand: (? for help)")
		if scanner.Scan() {
			switch cbuf := scanner.Text(); strings.ToLower(cbuf[0]) {
			case 'q':
				break
			case 's':
				fmt.Println("Running Sanity...")
				sanity(NOTHING)
			case 'f':
				fmt.Println("Running Sanfix...")
				sanfix(NOTHING)
			case 'p':
				if i := strings.IndexFunc(cbuf, unicode.IsSpace); i != -1 {
					cbuf = cbuf[i:]
				}
				if len(cbuf) > 0 {
					cbuf = cbuf[1:]
				}
				sane_dump_object(NOTHING, cbuf)
			case 'w':
				if sscanf(cbuf, "%*s %s", buf2); buf2 != "" {
					fmt.Printf("Writing database to %s...\n", buf2)
				} else {
					fmt.Println("Writing database...")
				}
				do_dump(GOD, buf2)
				fmt.Println("Done.")
			case 'c':
				if i := strings.IndexFunc(cbuf, unicode.IsSpace); i != -1 {
					cbuf = cbuf[i:]
				}
				if len(cbuf) > 0 {
					cbuf = cbuf[1:]
				}
				sanechange(NOTHING, cbuf)
			case 'x':
				extract()
			case 'y':
				extract_single()
			case 'h', '?':
				fmt.Println()
				fmt.Println("s                           Run Sanity checks on database")
				fmt.Println("f                           Automatically fix the database")
				fmt.Println("p <ObjectID>                   Print an object")
				fmt.Println("q                           Quit")
				fmt.Println("w <file>                    Write database to file.")
				fmt.Println("c <ObjectID> <field> <value>   Change a field on an object.")
				fmt.Println("                              (\"c ? ?\" for list)")
				fmt.Println("x <ObjectID> [<filename>]      Extract all objects belonging to <ObjectID>")
				fmt.Println("y <ObjectID> [<filename>]      Extract the single object <ObjectID>")
				fmt.Println("?                           Help! (Displays this screen.")
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Quitting.\n")
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