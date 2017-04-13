package fbmuck

func unparse_flag(thing dbref, flag int, f string) (r string) {
	if db.Fetch(thing).flags & flag != 0 {
		r = f
	}
	return
}

func unparse_flags(thing dbref) (r string) {
	switch thing.(type) {
	case TYPE_ROOM:
		r = "R"
	case TYPE_EXIT:
		r = "E"
	case TYPE_PLAYER:
		r = "P"
	case TYPE_PROGRAM:
		r = "F"
	}

	if db.Fetch(thing).flags & ~TYPE_MASK != 0 {
		r += unparse_flag(thing, WIZARD, "W")
		r += unparse_flag(thing, LINK_OK, "L")
		r += unparse_flag(thing, KILL_OK, "K")
		r += unparse_flag(thing, DARK, "D")
		r += unparse_flag(thing, STICKY, "S")
		r += unparse_flag(thing, QUELL, "Q")
		r += unparse_flag(thing, BUILDER, "B")
		r += unparse_flag(thing, CHOWN_OK, "C")
		r += unparse_flag(thing, JUMP_OK, "J")
		r += unparse_flag(thing, HAVEN, "H")
		r += unparse_flag(thing, ABODE, "A")
		r += unparse_flag(thing, VEHICLE, "V")
		r += unparse_flag(thing, XFORCIBLE, "X")
		r += unparse_flag(thing, ZOMBIE, "Z")
		if tp_enable_match_yield {
			r += unparse_flag(thing, YIELD, "Y")
			r += unparse_flag(thing, OVERT, "O")
		}
		if MLevRaw(thing) != NON_MUCKER {
			r = fmt.Sprintf("%vM%v", r, MLevRaw(thing))
		}
	}
	return
}

func unparse_object(player, loc dbref) (r string) {
	if player != NOTHING {
		if _, ok := player.(TYPE_PLAYER); !ok {
			player = db.Fetch(player).Owner
		}
	}
	switch loc {
	case NOTHING:
		r = "*NOTHING*"
	case AMBIGUOUS:
		r = "*AMBIGUOUS*"
	case HOME:
		r = "*HOME*"
	case !valid_reference(loc):
		r = "*INVALID*"
	default:
		if player == NOTHING || (db.Fetch(player).flags & STICKY == 0 && (can_link_to(player, NOTYPE, loc) || (Typeof(loc) != TYPE_PLAYER && (controls_link(player, loc) || db.Fetch(loc).flags & CHOWN_OK != 0)))) {
			r = fmt.Sprintf("%s(#%d%s)", db.Fetch(loc).name, loc, unparse_flags(loc))
		} else {
			r = db.Fetch(loc).name
		}
	}
	return
}