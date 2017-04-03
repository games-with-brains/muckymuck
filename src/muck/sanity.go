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

	if (!result || d < 0 || d >= db_top) {
		SanPrint(player, "Invalid Object.")
		return
	}

	SanPrint(player, "Object:         %s", unparse_object(GOD, d))
	SanPrint(player, "  Owner:          %s", unparse_object(GOD, db.Fetch(d).owner))
	SanPrint(player, "  Location:       %s", unparse_object(GOD, db.Fetch(d).location))
	SanPrint(player, "  Contents Start: %s", unparse_object(GOD, db.Fetch(d).contents))
	SanPrint(player, "  Exits Start:    %s", unparse_object(GOD, db.Fetch(d).exits))
	SanPrint(player, "  Next:           %s", unparse_object(GOD, db.Fetch(d).next))

	switch TYPEOF(d) {
	case TYPE_THING:
		SanPrint(player, "  Home:           %s", unparse_object(GOD, db.Fetch(d).sp.(player_specific).home))
		SanPrint(player, "  Value:          %d", get_property_value(d, MESGPROP_VALUE))
	case TYPE_ROOM:
		SanPrint(player, "  Drop-to:        %s", unparse_object(GOD, db.Fetch(d).sp.(dbref)))
	case TYPE_PLAYER:
		SanPrint(player, "  Home:           %s", unparse_object(GOD, db.Fetch(d).sp.(player_specific).home))
		SanPrint(player, "  Pennies:        %d", get_property_value(d, MESGPROP_VALUE))
		if player < 0 {
			SanPrint(player, "  Password MD5:   %s", db.Fetch(d).sp.(player_specific).password)
		}
	case TYPE_EXIT:
		SanPrint(player, "  Links:")
		for _, v := range db.Fetch(d).sp.exit.dest {
			SanPrint(player, "    %s", unparse_object(GOD, v))
		}
	}
	SanPrint(player, "Referring Objects:")
	for i := 0; i < db_top; i++ {
		if db.Fetch(i).contents == d {
			SanPrint(player, "  By contents field: %s", unparse_object(GOD, i))
		}
		if db.Fetch(i).exits == d {
			SanPrint(player, "  By exits field:    %s", unparse_object(GOD, i))
		}
		if db.Fetch(i).next == d {
			SanPrint(player, "  By next field:     %s", unparse_object(GOD, i))
		}
	}
	SanPrint(player, "Done.")
}

func violate(player, i dbref, s string) {
	SanPrint(player, "Object \"%s\" %s!", unparse_object(GOD, i), s)
	SanityViolated = true
}

func valid_ref(obj dbref) (r bool) {
	switch {
	case obj == NOTHING:
		r = true
	case obj < 0:
	case obj >= db_top:
	default:
		r = true
	}
	return
}

func valid_obj(obj dbref) (r bool) {
	switch {
	case obj == NOTHING:
	case !valid_ref(obj):
	default:
		switch TYPEOF(obj) {
		case TYPE_ROOM, TYPE_EXIT, TYPE_PLAYER, TYPE_PROGRAM, TYPE_THING:
			r = true
		}
	}
	return
}

func check_next_chain(player, obj dbref) {
	orig := obj
	for obj != NOTHING && valid_ref(obj) {
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
	if !valid_ref(obj) {
		violate(player, obj, "has an invalid object in its 'next' chain");
	}
}

func find_orphan_objects(player dbref) {
	SanPrint(player, "Searching for orphan objects...");
	for i := 0; i < db_top; i++ {
		db.Fetch(i).flags &= ~SANEBIT
	}

	db.Fetch(GLOBAL_ENVIRONMENT).flags |= SANEBIT

	for i := 0; i < db_top; i++ {
		if db.Fetch(i).exits != NOTHING {
			if db.Fetch(db.Fetch(i).exits).flags & SANEBIT != 0 {
				violate(player, db.Fetch(i).exits, "is referred to by more than one object's Next, Contents, or Exits field")
			} else {
				db.Fetch(db.Fetch(i).exits).flags |= SANEBIT
			}
		}
		if db.Fetch(i).contents != NOTHING {
			if db.Fetch(db.Fetch(i).contents).flags & SANEBIT != 0 {
				violate(player, db.Fetch(i).contents, "is referred to by more than one object's Next, Contents, or Exits field")
			} else {
				db.Fetch(db.Fetch(i).contents).flags |= SANEBIT
			}
		}
		if db.Fetch(i).next != NOTHING {
			if db.Fetch(db.Fetch(i).next).flags & SANEBIT != 0 {
				violate(player, db.Fetch(i).next, "is referred to by more than one object's Next, Contents, or Exits field")
			} else {
				db.Fetch(db.Fetch(i).next).flags |= SANEBIT
			}
		}
	}

	for i := 0; i < db_top; i++ {
		if db.Fetch(i).flags & SANEBIT == 0 {
			violate(player, i, "appears to be an orphan object, that is not referred to by any other object")
		}
	}

	for i := 0; i < db_top; i++ {
		db.Fetch(i).flags &= ~SANEBIT
	}
}

func check_room(player, obj dbref) {
	switch i := db.Fetch(obj).sp.(dbref); {
	case !valid_ref(i) && i != HOME:
		violate(player, obj, "has its dropto set to an invalid object")
	case i >= 0 && TYPEOF(i) != TYPE_THING && TYPEOF(i) != TYPE_ROOM:
		violate(player, obj, "has its dropto set to a non-room, non-thing object")
	}
}

func check_thing(player, obj dbref) {
	switch i := db.Fetch(obj).sp.(player_specific).home; {
	case !valid_obj(i):
		violate(player, obj, "has its home set to an invalid object");
	case TYPEOF(i) != TYPE_ROOM && TYPEOF(i) != TYPE_THING && TYPEOF(i) != TYPE_PLAYER:
		violate(player, obj, "has its home set to an object that is not a room, thing, or player")
	}
}

func check_exit(player, obj dbref) {
	for i, v := range db.Fetch(obj).sp.exit.dest {
		if !valid_ref(v) && v != HOME {
			violate(player, obj, "has an invalid object as one of its link destinations")
		}
	}
}

func check_player(player, obj dbref) {
	i := db.Fetch(obj).sp.(player_specific).home
	switch {
	case !valid_obj(i):
		violate(player, obj, "has its home set to an invalid object")
	case i >= 0 && TYPEOF(i) != TYPE_ROOM:
		violate(player, obj, "has its home set to a non-room object")
	}
}

func check_program(player, obj dbref) {
}

func check_contents_list(dbref player, dbref obj) {
	if TYPEOF(obj) != TYPE_PROGRAM && TYPEOF(obj) != TYPE_EXIT {
		var i dbref
		var limit int
		for i, limit = db.Fetch(obj).contents, db_top - 1; valid_obj(i) && limit && db.Fetch(i).location == obj && TYPEOF(i) != TYPE_EXIT; i = db.Fetch(i).next {
			limit--
		}
		if i != NOTHING {
			switch {
			case limit > 0:
				check_next_chain(player, db.Fetch(obj).contents)
				violate(player, obj, "is the containing object, and has the loop in its contents chain")
			case !valid_obj(i):
				violate(player, obj, "has an invalid object in its contents list")
			default:
				if TYPEOF(i) == TYPE_EXIT {
					violate(player, obj, "has an exit in its contents list (it shoudln't)")
				}
				if db.Fetch(i).location != obj {
					violate(player, obj, "has an object in its contents lists that thinks it is located elsewhere")
				}
			}
		}
	} else {
		if db.Fetch(obj).contents != NOTHING {
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
		for i, limit = db.Fetch(obj).exits, db_top - 1; valid_obj(i) && limit && db.Fetch(i).location == obj && TYPEOF(i) == TYPE_EXIT; i = db.Fetch(i).next) {
			limit--
		}
		if i != NOTHING {
			switch {
			case limit > 0:
				check_next_chain(player, db.Fetch(obj).contents)
				violate(player, obj, "is the containing object, and has the loop in its exits chain")
			case !valid_obj(i):
				violate(player, obj, "has an invalid object in it's exits list")
			default:
				if TYPEOF(i) != TYPE_EXIT {
					violate(player, obj, "has a non-exit in it's exits list")
				}
				if db.Fetch(i).location != obj {
					violate(player, obj, "has an exit in its exits lists that thinks it is located elsewhere")
				}
			}
		}
	} else {
		if db.Fetch(obj).exits != NOTHING {
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
	case !valid_obj(db.Fetch(obj).owner):
		violate(player, obj, "has an invalid object as its owner.")
	case TYPEOF(db.Fetch(obj).owner) != TYPE_PLAYER:
		violate(player, obj, "has a non-player object as its owner.")
	}

	//	check location 
	if !valid_obj(db.Fetch(obj).location) && !(obj == GLOBAL_ENVIRONMENT && db.Fetch(obj).location == NOTHING) {
		violate(player, obj, "has an invalid object as it's location")
	}

	if db.Fetch(obj).location != NOTHING && (TYPEOF(db.Fetch(obj).location) == TYPE_EXIT || TYPEOF(db.Fetch(obj).location) == TYPE_PROGRAM) {
		violate(player, obj, "thinks it is located in a non-container object")
	}

	check_contents_list(player, obj)
	check_exits_list(player, obj)

	switch TYPEOF(obj) {
	case TYPE_ROOM:
		check_room(player, obj)
	case TYPE_THING:
		check_thing(player, obj)
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
		for i, j, increp := 0, 0, 10000; i < db_top; i++ {
			if i % increp == 0 {
				j = i + increp - 1
				if j >= db_top {
					j = db_top - 1
				}
				SanPrint(player, "Checking objects %d to %d...", i, j)
				if player >= 0 {
					flush_user_output(player)
				}
			}
			check_object(player, i)
		}
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
	if db.Fetch(obj).contents != NOTHING {
		SanFixed(obj, "Cleared contents of %s")
		db.Fetch(obj).contents = NOTHING
		db.Fetch(obj).flags |= OBJECT_CHANGED
	}
	if db.Fetch(obj).exits != NOTHING {
		SanFixed(obj, "Cleared exits of %s")
		db.Fetch(obj).exits = NOTHING
		db.Fetch(obj).flags |= OBJECT_CHANGED
	}
}

func cut_bad_contents(obj dbref) {
	prev := NOTHING
	for loop := db.Fetch(obj).contents; loop != NOTHING; loop = db.Fetch(loop).next {
		if !valid_obj(loop) || db.Fetch(loop).flags & SANEBIT || TYPEOF(loop) == TYPE_EXIT || db.Fetch(loop).location != obj || loop == obj {
			switch {
			case !valid_obj(loop):
				SanFixed(obj, "Contents chain for %s cut at invalid dbref")
			case TYPEOF(loop) == TYPE_EXIT:
				SanFixed2(obj, loop, "Contents chain for %s cut at exit %s")
			case loop == obj:
				SanFixed(obj, "Contents chain for %s cut at self-reference")
			case db.Fetch(loop).location != obj:
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
				db.Fetch(obj).contents = NOTHING
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
	for loop := db.Fetch(obj).exits; loop != NOTHING; loop = db.Fetch(loop).next {
		if !valid_obj(loop) || db.Fetch(loop).flags & SANEBIT || TYPEOF(loop) != TYPE_EXIT || db.Fetch(loop).location != obj {
			switch {
			case !valid_obj(loop):
				SanFixed(obj, "Exits chain for %s cut at invalid dbref")
			case TYPEOF(loop) != TYPE_EXIT:
				SanFixed2(obj, loop, "Exits chain for %s cut at non-exit %s")
			case db.Fetch(loop).location != obj:
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
				db.Fetch(obj).exits = NOTHING
				db.Fetch(obj).flags |= OBJECT_CHANGED
			}
			return
		}
		db.Fetch(loop).flags |= SANEBIT
		prev = loop
	}
}

func hacksaw_bad_chains() {
	for i := 0; i < db_top; i++ {
		if TYPEOF(i) != TYPE_ROOM && TYPEOF(i) != TYPE_THING && TYPEOF(i) != TYPE_PLAYER {
			cut_all_chains(i)
		} else {
			cut_bad_contents(i)
			cut_bad_exits(i)
		}
	}
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
	db.Fetch(*room).location = GLOBAL_ENVIRONMENT
	db.Fetch(*room).exits = NOTHING
	db.Fetch(*room).sp = NOTHING
	db.Fetch(*room).flags = TYPE_ROOM | SANEBIT

	db.Fetch(*room).next = db.Fetch(GLOBAL_ENVIRONMENT.contents)
	db.Fetch(*room).flags |= OBJECT_CHANGED
	db.Fetch(GLOBAL_ENVIRONMENT.contents) = *room

	SanFixed(*room, "Using %s to resolve unknown location")
	for i := 1; lookup_player(player_name) != NOTHING; i++ {
		player_name = fmt.Sprintf("lost+found%d", i)
	}
	*player = new_object()
	db.Fetch(*player).name = player_name
	db.Fetch(*player).location = *room
	db.Fetch(*player).flags = TYPE_PLAYER | SANEBIT
	db.Fetch(*player).owner = *player
	*player = &player_specific{ home: *room, exits: NOTHING, curr_prog: NOTHING }
	add_property(*player, MESGPROP_VALUE, NULL, tp_start_pennies)
	rpass := rand_password()
	set_password(*player, rpass)
	db.Fetch(*player).next = db.Fetch(*room).contents
	db.Fetch(*player).flags |= OBJECT_CHANGED
	db.Fetch(*room).contents = *player
	db.Fetch(*player).flags |= OBJECT_CHANGED
	add_player(*player)
	log2file("logs/sanfixed", "Using %s (with password %s) to resolve unknown owner", unparse_object(GOD, *player), rpass)
	db.Fetch(*room).owner = *player
	db.Fetch(*room).flags |= OBJECT_CHANGED
	db.Fetch(*player).flags |= OBJECT_CHANGED
	db.Fetch(GLOBAL_ENVIRONMENT).flags |= OBJECT_CHANGED
}

func fix_room(obj dbref) {
	switch i := db.Fetch(obj).sp.(dbref); {
	case !valid_ref(i) && i != HOME:
		SanFixed(obj, "Removing invalid drop-to from %s")
		db.Fetch(obj).sp = NOTHING
		db.Fetch(obj).flags |= OBJECT_CHANGED
	case i >= 0 && TYPEOF(i) != TYPE_THING && TYPEOF(i) != TYPE_ROOM:
		SanFixed2(obj, i, "Removing drop-to on %s to %s")
		db.Fetch(obj).sp = NOTHING
		db.Fetch(obj).flags |= OBJECT_CHANGED
	}
}

func fix_thing(obj dbref) {
	if i := db.Fetch(obj).sp.(player_specific).home; !valid_obj(i) || (TYPEOF(i) != TYPE_ROOM && TYPEOF(i) != TYPE_THING && TYPEOF(i) != TYPE_PLAYER) {
		SanFixed2(obj, db.Fetch(obj).owner, "Setting the home on %s to %s, it's owner")
		db.Fetch(obj).sp.(player_specific).home = db.Fetch(obj).owner
		db.Fetch(obj).flags |= OBJECT_CHANGED
	}
}

func fix_exit(obj dbref) {
	dest := db.Fetch(obj).sp.exit.dest
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
		db.Fetch(obj).sp.exit.dest = d
	} else {
		for ol := len(dest); ol > l; ol-- {
			dest[ol] = nil
		}
		dest = dest[:l]
	}
}

func fix_player(obj dbref) {
	if i := db.Fetch(obj).sp.(player_specific).home; !valid_obj(i) || TYPEOF(i) != TYPE_ROOM {
		SanFixed2(obj, tp_player_start, "Setting the home on %s to %s")
		db.Fetch(obj).sp.(player_specific).home = tp_player_start
		db.Fetch(obj).flags |= OBJECT_CHANGED
	}
}

func find_misplaced_objects() {
	player := NOTHING
	var room dbref

	for loop := 0; loop < db_top; loop++ {
		if TYPEOF(loop) != TYPE_ROOM && TYPEOF(loop) != TYPE_THING && TYPEOF(loop) != TYPE_PLAYER && TYPEOF(loop) != TYPE_EXIT && TYPEOF(loop) != TYPE_PROGRAM {
			SanFixedRef(loop, "Object #%d is of unknown type")
			SanityViolated = true
			continue
		}
		if db.Fetch(loop).name == "" {
			if TYPEOF(loop) == TYPE_PLAYER {
				name := "Unnamed"
				for i := 1; lookup_player(name) != NOTHING; i++ {
					name = fmt.Sprintf("Unnamed%d", i)
				}
				db.Fetch(loop).name = name
				add_player(loop)
			} else {
				db.Fetch(loop).name = "Unnamed"
			}
			SanFixed(loop, "Gave a name to %s")
			db.Fetch(loop).flags |= OBJECT_CHANGED
		}
		if !valid_obj(db.Fetch(loop).owner) || TYPEOF(db.Fetch(loop).owner) != TYPE_PLAYER {
			if player == NOTHING {
				create_lostandfound(&player, &room)
			}
			SanFixed2(loop, player, "Set owner of %s to %s")
			db.Fetch(loop).owner = player
			db.Fetch(loop).flags |= OBJECT_CHANGED
		}
		
		if loop != GLOBAL_ENVIRONMENT && !valid_obj(db.Fetch(loop).location) || TYPEOF(db.Fetch(loop).location) == TYPE_EXIT || TYPEOF(db.Fetch(loop).location) == TYPE_PROGRAM || (TYPEOF(loop) == TYPE_PLAYER && TYPEOF(db.Fetch(loop).location) == TYPE_PLAYER) {
			if TYPEOF(loop) == TYPE_PLAYER {
				if valid_obj(db.Fetch(loop).location) && TYPEOF(db.Fetch(loop).location) == TYPE_PLAYER {
					dbref loop1;

					loop1 = db.Fetch(loop).location
					if db.Fetch(loop1).contents == loop {
						db.Fetch(loop1).contents = db.Fetch(loop).next
						db.Fetch(loop1).flags |= OBJECT_CHANGED
					} else
						for loop1 = db.Fetch(loop1).contents; loop1 != NOTHING; loop1 = db.Fetch(loop1).next {
							if db.Fetch(loop1).next == loop {
								db.Fetch(loop1).next = db.Fetch(loop).next
								db.Fetch(loop1).flags |= OBJECT_CHANGED
								break
							}
						}
				}
				db.Fetch(loop).location = tp_player_start
			} else {
				if player == NOTHING {
					create_lostandfound(&player, &room)
				}
				db.Fetch(loop).location = room
			}
			db.Fetch(loop).flags |= OBJECT_CHANGED
			db.Fetch(db.Fetch(loop).location).flags |= OBJECT_CHANGED
			if TYPEOF(loop) == TYPE_EXIT {
				db.Fetch(loop).next = db.Fetch(db.Fetch(loop).location).exits
				db.Fetch(loop).flags |= OBJECT_CHANGED
				db.Fetch(db.Fetch(loop).location).exits = loop
			} else {
				db.Fetch(loop).next = db.Fetch(db.Fetch(loop).location).contents
				db.Fetch(loop).flags |= OBJECT_CHANGED
				db.Fetch(db.Fetch(loop).location).contents = loop
			}
			db.Fetch(loop).flags |= SANEBIT
			SanFixed2(loop, db.Fetch(loop).location, "Set location of %s to %s")
		}
		switch TYPEOF(loop) {
		case TYPE_ROOM:
			fix_room(loop)
		case TYPE_THING:
			fix_thing(loop)
		case TYPE_PLAYER:
			fix_player(loop)
		case TYPE_EXIT:
			fix_exit(loop)
		}
	}
}

func adopt_orphans() {
	for loop := 0; loop < db_top; loop++ {
		if db.Fetch(loop).flags & SANEBIT == 0 {
			db.Fetch(loop).flags |= OBJECT_CHANGED
			switch TYPEOF(loop) {
			case TYPE_ROOM, TYPE_THING, TYPE_PLAYER, TYPE_PROGRAM:
				db.Fetch(loop).next = db.Fetch(db.Fetch(loop).location).contents
				db.Fetch(db.Fetch(loop).location).contents = loop
				SanFixed2(loop, db.Fetch(loop).location, "Orphaned object %s added to contents of %s")
				break
			case TYPE_EXIT:
				db.Fetch(loop).next = db.Fetch(db.Fetch(loop).location).exits
				db.Fetch(db.Fetch(loop).location).exits = loop
				SanFixed2(loop, db.Fetch(loop).location, "Orphaned exit %s added to exits of %s")
				break
			default:
				SanityViolated = true
				break
			}
		}
	}
}

func clean_global_environment() {
	if db.Fetch(GLOBAL_ENVIRONMENT).next != NOTHING {
		SanFixed(GLOBAL_ENVIRONMENT, "Removed the global environment %s from a chain")
		db.Fetch(GLOBAL_ENVIRONMENT).next = NOTHING
		db.Fetch(GLOBAL_ENVIRONMENT).flags |= OBJECT_CHANGED
	}
	if db.Fetch(GLOBAL_ENVIRONMENT).location != NOTHING {
		SanFixed2(GLOBAL_ENVIRONMENT, db.Fetch(GLOBAL_ENVIRONMENT).location, "Removed the global environment %s from %s")
		db.Fetch(GLOBAL_ENVIRONMENT).location = NOTHING
		db.Fetch(GLOBAL_ENVIRONMENT).flags |= OBJECT_CHANGED
	}
}

func sanfix(player dbref) {
	if player > NOTHING && player != GOD {
		notify(player, "Yeah right!  With a psyche like yours, you think theres any hope of getting your sanity fixed?")
		return
	}

	SanityViolated = false
	for loop := 0; loop < db_top; loop++ {
		db.Fetch(loop).flags &= ~SANEBIT
	}
	db.Fetch(GLOBAL_ENVIRONMENT).flags |= SANEBIT

	if !valid_obj(tp_player_start) || TYPEOF(tp_player_start) != TYPE_ROOM {
		SanFixed(GLOBAL_ENVIRONMENT, "Reset invalid player_start to %s")
		tp_player_start = GLOBAL_ENVIRONMENT
	}

	hacksaw_bad_chains()
	find_misplaced_objects()
	adopt_orphans()
	clean_global_environment()

	for loop := 0; loop < db_top; loop++ {
		db.Fetch(loop).flags &= ~SANEBIT
	}

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
	if !valid_ref(d) || d < 0 {
		SanPrint(player, "## %d is an invalid dbref.", d)
		return
	}

	buf2 = ""
	switch field {
	case "next":
		buf2 = unparse_object(GOD, db.Fetch(d).next))
		db.Fetch(d).next = v
		db.Fetch(d).flags |= OBJECT_CHANGED
		SanPrint(player, "## Setting #%d's next field to %s", d, unparse_object(GOD, v))
	case "exits":
		buf2 = unparse_object(GOD, db.Fetch(d).exits)
		db.Fetch(d).exits = v
		db.Fetch(d).flags |= OBJECT_CHANGED
		SanPrint(player, "## Setting #%d's Exits list start to %s", d, unparse_object(GOD, v))
	case "contents":
		buf2 = unparse_object(GOD, db.Fetch(d).contents)
		db.Fetch(d).contents = v
		db.Fetch(d).flags |= OBJECT_CHANGED
		SanPrint(player, "## Setting #%d's Contents list start to %s", d, unparse_object(GOD, v))
	case "location":
		buf2 = unparse_object(GOD, db.Fetch(d).location)
		db.Fetch(d).location = v
		db.Fetch(d).flags |= OBJECT_CHANGED
		SanPrint(player, "## Setting #%d's location to %s", d, unparse_object(GOD, v))
	case "owner":
		buf2 = unparse_object(GOD, db.Fetch(d).owner)
		db.Fetch(d).owner = v
		db.Fetch(d).flags |= OBJECT_CHANGED
		SanPrint(player, "## Setting #%d's owner to %s", d, unparse_object(GOD, v))
	case "home":
		var ip *int
		switch TYPEOF(d) {
		case TYPE_PLAYER:
			ip = &(db.Fetch(d).sp.(player_specific).home)
		case TYPE_THING:
			ip = &(db.Fetch(d).sp.(player_specific).home)
		default:
			printf("%s has no home to set.\n", unparse_object(GOD, d))
			return
		}
		buf2 = unparse_object(GOD, *ip)
		*ip = v
		db.Fetch(d).flags |= OBJECT_CHANGED
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
	fprintf(f, "  #%d\n", d);
	fprintf(f, "  Object:         %s\n", unparse_object(GOD, d))
	fprintf(f, "  Owner:          %s\n", unparse_object(GOD, db.Fetch(d).owner))
	fprintf(f, "  Location:       %s\n", unparse_object(GOD, db.Fetch(d).location))
	fprintf(f, "  Contents Start: %s\n", unparse_object(GOD, db.Fetch(d).contents))
	fprintf(f, "  Exits Start:    %s\n", unparse_object(GOD, db.Fetch(d).exits))
	fprintf(f, "  Next:           %s\n", unparse_object(GOD, db.Fetch(d).next))

	switch TYPEOF(d) {
	case TYPE_THING:
		fprintf(f, "  Home:           %s\n", unparse_object(GOD, db.Fetch(d).sp.(player_specific).home))
		fprintf(f, "  Value:          %d\n", get_property_value(d, MESGPROP_VALUE))
	case TYPE_ROOM:
		fprintf(f, "  Drop-to:        %s\n", unparse_object(GOD, db.Fetch(d)sp.(dbref)))
	case TYPE_PLAYER:
		fprintf(f, "  Home:           %s\n", unparse_object(GOD, db.Fetch(d).sp.(player_specific).home))
		fprintf(f, "  Pennies:        %d\n", get_property_value(d, MESGPROP_VALUE))
	case TYPE_EXIT:
		fprintf(f, "  Links:         ")
		for _, v := range db.Fetch(d).sp.exit.dest {
			fprintf(f, " %s;", unparse_object(GOD, v))
		}
		fprintf(f, "\n")
	case TYPE_PROGRAM:
		fprintf(f, "  Listing:\n")
		extract_program(f, d)
	}

	if db.Fetch(d).properties {
		fprintf(f, "  Properties:\n")
		extract_props(f, d)
	} else {
		fprintf(f, "  No properties\n")
	}
	fprintf(f, "\n")
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
		for i := 0; i < db_top; i++ {
			if db.Fetch(i).owner == d {
				extract_object(f, i)
			}
		}
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