int *nexted;

func readd_contents(obj dbref) {
	var what dbref
	where := db.Fetch(obj).location
	switch db.Fetch(obj).flags & TYPE_MASK {
	case TYPE_ROOM, TYPE_THING, TYPE_PROGRAM, TYPE_PLAYER:
		if db.Fetch(where).contents == NOTHING {
			db.Fetch(where).contents = obj
			return
		}
		what = db.Fetch(where).contents
		for db.Fetch(what).next != NOTHING {
			what = db.Fetch(what).next
		}
		db.Fetch(what).next = obj
	case TYPE_EXIT:
		switch db.Fetch(where).flags & TYPE_MASK {
		case TYPE_ROOM, TYPE_THING, TYPE_PLAYER:
			if db.Fetch(where).exits == NOTHING {
				db.Fetch(where).exits = obj
				return
			}
			what = db.Fetch(where).exits
		}
		for db.Fetch(what).next != NOTHING {
			what = db.Fetch(what).next
		}
		db.Fetch(what).next = obj
	}
}

func check_contents(obj dbref) {
	switch db.Fetch(obj).flags & TYPE_MASK {
	case TYPE_PROGRAM, TYPE_EXIT:
	default:
		if db.Fetch(obj).contents != NOTHING {
			for db.Fetch(obj).contents != NOTHING && db.Fetch(db.Fetch(obj).contents).location != obj {
				lastwhere := db.Fetch(obj).contents
				db.Fetch(obj).contents = db.Fetch(lastwhere).next
				db.Fetch(lastwhere).next = NOTHING
				readd_contents(lastwhere)
			}
			if where := db.Fetch(obj).contents; where != NOTHING {
				for db.Fetch(where).next != NOTHING {
					if db.Fetch(db.Fetch(where).next).location != obj {
						lastwhere := db.Fetch(where).next
						db.Fetch(where).next = db.Fetch(lastwhere).next
						db.Fetch(lastwhere).next = NOTHING
						readd_contents(lastwhere)
					} else {
						where = db.Fetch(where).next
					}
				}
			}
		}
		if db.Fetch(obj).exits != NOTHING {
			for db.Fetch(obj).exits != NOTHING && db.Fetch(db.Fetch(obj).exits).location != obj {
				lastwhere := db.Fetch(obj).exits
				db.Fetch(obj).exits = db.Fetch(lastwhere).next
				db.Fetch(lastwhere).next = NOTHING
				readd_contents(lastwhere)
			}
			if where := db.Fetch(obj).exits; where != NOTHING {
				for db.Fetch(where).next != NOTHING {
					if db.Fetch(db.Fetch(where).next).location != obj {
						lastwhere := db.Fetch(where).next
						db.Fetch(where).next = db.Fetch(lastwhere).next
						db.Fetch(lastwhere).next = NOTHING
						readd_contents(lastwhere)
					} else {
						where = db.Fetch(where).next
					}
				}
			}
		}
	}
}

func check_common(obj dbref) {
	/* check name */
	if !db.Fetch(obj).name {
		buf = fmt.Sprintf("Unknown%d", obj)
		db.Fetch(obj).name = buf
	}

	if db.Fetch(obj).location >= db_top {
		db.Fetch(obj).location = tp_player_start
	}

	if db.Fetch(obj).contents < db_top {
		nexted[db.Fetch(obj).contents] = obj
	} else {
		db.Fetch(obj).contents = NOTHING
	}

	if db.Fetch(obj).next < db_top {
		nexted[db.Fetch(obj).next] = obj
	} else {
		db.Fetch(obj).next = NOTHING
	}
}

func check_room(obj dbref) {
	if db.Fetch(obj).sp.(dbref) >= db_top || (db.Fetch(db.Fetch(obj).sp.(dbref)).flags & TYPE_MASK != TYPE_ROOM && db.Fetch(obj).sp != NOTHING && db.Fetch(obj).sp != HOME) {
		db.Fetch(obj).sp = NOTHING
	}

	if db.Fetch(obj).exits < db_top {
		nexted[db.Fetch(obj).exits] = obj
	} else {
		db.Fetch(obj).exits = NOTHING
	}

	if db.Fetch(obj).owner >= db_top || (db.Fetch(db.Fetch(obj).owner).flags & TYPE_MASK != TYPE_PLAYER) {
		db.Fetch(obj).owner = GOD
	}
}

func check_thing(obj dbref) {
	if db.Fetch(obj).sp.thing.home >= db_top || ((db.Fetch(db.Fetch(obj).sp.thing.home).flags & TYPE_MASK != TYPE_ROOM) && (db.Fetch(db.Fetch(obj).sp.thing.home).flags & TYPE_MASK != TYPE_PLAYER)) {
		db.Fetch(obj).sp.thing.home = tp_player_start
	}

	if db.Fetch(obj).exits < db_top {
		nexted[db.Fetch(obj).exits] = obj
	} else {
		db.Fetch(obj).exits = NOTHING
	}

	if db.Fetch(obj).owner >= db_top || (db.Fetch(db.Fetch(obj).owner).flags & TYPE_MASK != TYPE_PLAYER) {
		db.Fetch(obj).owner = GOD
	}
}

func check_exit(ref dbref) {
	obj := db.Fetch(ref)
	for i, v := range obj.sp.exit.dest {
		if v >= db_top {
			obj.sp.exit.dest[i] = NOTHING
		}
	}
	if obj.owner >= db_top || (db.Fetch(obj.owner).flags & TYPE_MASK != TYPE_PLAYER) {
		obj.owner = GOD
	}
}

func check_player(ref dbref) {
	obj := db.Fetch(ref)
	player := obj.sp.(player_specific)
	if player.home >= db_top || (db.Fetch(player.home).flags & TYPE_MASK != TYPE_ROOM) {
		player.home = tp_player_start
	}

	if obj.exits < db_top {
		nexted[obj.exits] = obj
	} else {
		obj.exits = NOTHING
	}
}

func check_program(obj dbref) {
	if db.Fetch(obj).owner >= db_top || (db.Fetch(db.Fetch(obj).owner).flags & TYPE_MASK != TYPE_PLAYER) {
		db.Fetch(obj).owner = GOD
	}
}

FILE *input_file;
FILE *delta_infile = NULL;
FILE *delta_outfile = NULL;
FILE *output_file;

char *in_filename;
char *out_filename;

func main() {
	switch {
	case len(os.Args) != 3:
		fprintf(stderr, "Usage: %v infile outfile\n", argv[0])
		return
	case os.Args[1] == os.Args[2]:
		fprintf(stderr, "%v: input and output files can't have same name.\n", os.Args[0])
		return
	default:
		in_filename := os.Args[1]
		if input_file := fopen(in_filename, "rb"); input_file == nil {
			fprintf(stderr, "%v: unable to open input file.\n", os.Args[0])
			return
		} else {
			out_filename := os.Args[2]
			if output_file := fopen(out_filename, "wb"); output_file == nil {
				fprintf(stderr, "%v: unable to write to output file.\n", os.Args[0])
				return
			} else {
				db_free()
				db_read(input_file)

				nexted := malloc((db_top + 1) * sizeof(int))
				for i := 0; i < db_top; i++ {
					nexted[i] = NOTHING
				}

				for i := 0; i < db_top; i++ {
					check_common(i)
					switch db.Fetch(i).flags & TYPE_MASK {
					case TYPE_ROOM:
						check_room(i)
					case TYPE_THING:
						check_thing(i)
					case TYPE_EXIT:
						check_exit(i)
					case TYPE_PLAYER:
						check_player(i)
					case TYPE_PROGRAM:
						check_program(i)
					default:
						db.Fetch(i).flags &= ~TYPE_MASK
					}
				}
				for i := 0; i < db_top; i++ {
					if i != 0 && nexted[i] == NOTHING {
						readd_contents(i)
					}
				}
				for i := 0; i < db_top; i++ {
					check_contents(i)
				}
				db_write(output_file)
			}
		}
	}
}

/* dummy compiler */
func do_compile(int descr, dbref p, dbref pr, int force_err_disp) {}

func new_macro(name, definition string, player dbref) *macrotable {
	return nil
}

func log_status(format, p1, p2, p3, p4, p5, p6, p7, p8 string) {}

func clear_players() {}

func add_event(descr int, player, loc, trig dbref, dtime int, program dbref, fr *frame, strdata string) {}

func init_primitives() {}

func add_player(who dbref) {}

func do_parse_mesg(descr int, player, what dbref, inbuf, abuf, outbuf string, outbuflen int, mesgtyp int) string {
	return ""
}