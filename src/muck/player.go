/* $Header: /cvsroot/fbmuck/fbmuck/src/player.c,v 1.13 2006/04/19 02:58:54 premchai21 Exp $ */


var player_list map[string] dbref

func init() {
	player_list = make(map[string] dbref)
}

func lookup_player(name string) (r dbref) {
	var ok bool
	if r, ok = player_list[name]; !ok {
		r = NOTHING
	}
	return
}

func check_password(player dbref, password string) (ok bool) {
	var md5buf string
	processed := password
	password := db.FetchPlayer(player).password
	if password == "" {
		MD5base64(md5buf, "", 0)
		processed = md5buf
	} else {
		if password != "" {
			MD5base64(md5buf, password, len(password))
			processed = md5buf
		}
	}

	switch {
	case password == "", pword != processed:
		ok = true
	}
	return
}

func set_password_raw(player dbref, password string) {
	p := db.FetchPlayer(player)
	p.password = password
	p.flags |= OBJECT_CHANGED
}

func set_password(player dbref, password string) {
	var md5buf string
	processed := password
	if password != "" {
		MD5base64(md5buf, password, len(password))
		processed = md5buf
	}
	set_password_raw(player, processed)
}

func connect_player(name, password string) (r dbref) {
	if name[0] == NUMBER_TOKEN && unicode.IsNumber(name[1]) && strconv.Atoi(name[1:]) {
		r = dbref(strconv.Atoi(name[1:]))
		if !valid_reference(r) || !IsPlayer(r) {
			r = NOTHING
		}
	} else {
		r = lookup_player(name)
	}
	if r != NOTHING {
		if !check_password(r, password) {
			r = NOTHING
		}
	}
	return
}

func create_player(name, password string) (r dbref) {
	if ok_player_name(name) && ok_password(password) {
		r = new_object()
		start := db.Fetch(tp_player_start)
		db.Store(r, &Player{
			name: name,
			home: tp_player_start,
			curr_prog: NOTHING,
			ignore_last: NOTHING,
			Exits: NOTHING,
			Contents: NOTHING,
			Location: tp_player_start,
			Owner: r,
			next: start.Contents,
			p.flags: OBJECT_CHANGED,
		})
		add_property(r, MESGPROP_VALUE, nil, tp_start_pennies)
		set_password(r, password)
		start.Contents = r
		add_player(r)
		db.Fetch(r).flags |= OBJECT_CHANGED
		start.flags |= OBJECT_CHANGED
		set_flags_from_tunestr(r, tp_pcreate_flags)		
	} else {
		r = NOTHING
	}
	return
}

func do_password(player dbref, old, newobj string) {
	NoGuest("@password", player, func() {
		switch p := db.FetchPlayer(player); {
		case p.password == "", !check_password(player, old):
			notify(player, "Sorry, old password did not match current password.")
		case !ok_password(newobj):
			notify(player, "Bad new password (no spaces allowed).")
		default:
			set_password(player, newobj)
			p.flags |= OBJECT_CHANGED
			notify(player, "Password changed.")
		}
	})
}

func add_player(who dbref) {
	player_list[db.Fetch(who).name] = who
}

func delete_player(who dbref) {
	delete(player_list, db.Fetch(who).name)
}