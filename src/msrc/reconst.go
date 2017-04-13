var nexted []int

func readd_contents(obj dbref) {
	var what dbref
	where := db.Fetch(obj).Location
	switch {
	case IsRoom(obj), IsThing(obj), IsProgram(obj), IsPlayer(obj):
		if db.Fetch(obj).Contents == NOTHING {
			db.Fetch(obj).Contents = obj
			return
		}
		for what = db.Fetch(where).Contents; db.Fetch(what).next != NOTHING; what = db.Fetch(what).next {}
		db.Fetch(what).next = obj
	case IsExit(obj):
		switch {
		case IsRoom(where), IsThing(where), IsPlayer(where):
			if db.Fetch(where).Exits == NOTHING {
				db.Fetch(where).Exits = obj
				return
			}
			what = db.Fetch(where).Exits
		}
		for db.Fetch(what).next != NOTHING {
			what = db.Fetch(what).next
		}
		db.Fetch(what).next = obj
	}
}

func check_contents(obj dbref) {
	switch {
	case IsProgram(obj), IsExit(obj):
	default:
		if db.Fetch(obj).Contents != NOTHING {
			for db.Fetch(obj).Contents != NOTHING && db.Fetch(db.Fetch(obj).Contents).Location != obj {
				lastwhere := o.Contents
				db.Fetch(obj).Contents = db.Fetch(lastwhere).next
				db.Fetch(lastwhere).next = NOTHING
				readd_contents(lastwhere)
			}
			if where := db.Fetch(obj).Contents; where != NOTHING {
				for db.Fetch(where).next != NOTHING {
					if db.Fetch(db.Fetch(where).next).Location != obj {
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
		if db.Fetch(obj).Exits != NOTHING {
			for db.Fetch(obj).Exits != NOTHING && db.Fetch(db.Fetch(obj).Exits).Location != obj {
				lastwhere := db.Fetch(obj).Exits
				db.Fetch(obj).Exits = db.Fetch(lastwhere).next
				db.Fetch(lastwhere).next = NOTHING
				readd_contents(lastwhere)
			}
			if where := db.Fetch(obj).Exits; where != NOTHING {
				for db.Fetch(where).next != NOTHING {
					if db.Fetch(db.Fetch(where).next).Location != obj {
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

func check_common(id dbref) {
	o := db.Fetch(id)
	if !o.name {
		o.name = fmt.Sprintf("Unknown%d", id)
	}
	if o.Location >= db_top {
		o.Location = tp_player_start
	}
	if o.Contents < db_top {
		nexted[o.Contents] = id
	} else {
		o.Contents = NOTHING
	}
	if o.next < db_top {
		nexted[o.next] = obj
	} else {
		o.next = NOTHING
	}
}

func check_room(obj dbref) {
	if db.Fetch(obj).(dbref) >= db_top || (db.Fetch(db.Fetch(obj).(dbref)).flags & TYPE_MASK != TYPE_ROOM && db.Fetch(obj).sp != NOTHING && db.Fetch(obj).sp != HOME) {
		db.Fetch(obj).sp = NOTHING
	}

	if db.Fetch(obj).Exits < db_top {
		nexted[db.Fetch(obj).Exits] = obj
	} else {
		db.Fetch(obj).Exits = NOTHING
	}

	if db.Fetch(obj).Owner >= db_top || (db.Fetch(db.Fetch(obj).Owner).flags & TYPE_MASK != TYPE_PLAYER) {
		db.Fetch(obj).Owner = GOD
	}
}

func check_exit(ref dbref) {
	obj := db.Fetch(ref)
	for i, v := range obj.(Exit).Destinations {
		if v >= db_top {
			obj.(Exit).Destinations[i] = NOTHING
		}
	}
	if obj.Owner >= db_top || (db.Fetch(obj.Owner).flags & TYPE_MASK != TYPE_PLAYER) {
		obj.Owner = GOD
	}
}

func check_player(ref dbref) {
	obj := db.Fetch(ref)
	player := obj.(Player)
	if player.home >= db_top || (db.Fetch(player.home).flags & TYPE_MASK != TYPE_ROOM) {
		player.home = tp_player_start
	}

	if obj.Exits < db_top {
		nexted[obj.Exits] = obj
	} else {
		obj.Exits = NOTHING
	}
}

func check_program(obj dbref) {
	if db.Fetch(obj).Owner >= db_top || (db.Fetch(db.Fetch(obj).Owner).flags & TYPE_MASK != TYPE_PLAYER) {
		db.Fetch(obj).Owner = GOD
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
		} else {
			out_filename := os.Args[2]
			if output_file := fopen(out_filename, "wb"); output_file == nil {
				fprintf(stderr, "%v: unable to write to output file.\n", os.Args[0])
				return
			} else {
				db_free()
				db_read(input_file)

				nexted := make([]int, db_top + 1)
				EachObject(func(obj dbref, o *Object) {
					nexted[obj] = NOTHING
					check_common(obj)
					switch {
					case IsRoom(obj):
						check_room(obj)
					case IsExit(obj):
						check_exit(obj)
					case IsPlayer(obj):
						check_player(obj)
					case IsProgram(obj):
						check_program(obj)
					default:
						o.flags &= ~TYPE_MASK
					}
					if obj != 0 && nexted[obj] == NOTHING {
						readd_contents(obj)
					}
					check_contents(obj)
				})
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

func add_event(descr int, player, loc, trig dbref, dtime int, program dbref, fr *frame, strdata string) {}

func init_primitives() {}

func add_player(who dbref) {}

func do_parse_mesg(descr int, player, what dbref, inbuf, abuf, outbuf string, outbuflen int, mesgtyp int) string {
	return ""
}