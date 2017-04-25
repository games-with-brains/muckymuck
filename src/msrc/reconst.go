var nexted []int

func readd_contents(obj ObjectID) {
	var what ObjectID
	where := DB.Fetch(obj).Location
	switch {
	case IsRoom(obj), IsThing(obj), IsProgram(obj), IsPlayer(obj):
		if DB.Fetch(obj).Contents == NOTHING {
			DB.Fetch(obj).Contents = obj
			return
		}
		for what = DB.Fetch(where).Contents; DB.Fetch(what).next != NOTHING; what = DB.Fetch(what).next {}
		DB.Fetch(what).next = obj
	case IsExit(obj):
		switch {
		case IsRoom(where), IsThing(where), IsPlayer(where):
			if DB.Fetch(where).Exits == NOTHING {
				DB.Fetch(where).Exits = obj
				return
			}
			what = DB.Fetch(where).Exits
		}
		for DB.Fetch(what).next != NOTHING {
			what = DB.Fetch(what).next
		}
		DB.Fetch(what).next = obj
	}
}

func check_contents(obj ObjectID) {
	switch {
	case IsProgram(obj), IsExit(obj):
	default:
		if DB.Fetch(obj).Contents != NOTHING {
			for DB.Fetch(obj).Contents != NOTHING && DB.Fetch(DB.Fetch(obj).Contents).Location != obj {
				lastwhere := o.Contents
				DB.Fetch(obj).Contents = DB.Fetch(lastwhere).next
				DB.Fetch(lastwhere).next = NOTHING
				readd_contents(lastwhere)
			}
			if where := DB.Fetch(obj).Contents; where != NOTHING {
				for DB.Fetch(where).next != NOTHING {
					if DB.Fetch(DB.Fetch(where).next).Location != obj {
						lastwhere := DB.Fetch(where).next
						DB.Fetch(where).next = DB.Fetch(lastwhere).next
						DB.Fetch(lastwhere).next = NOTHING
						readd_contents(lastwhere)
					} else {
						where = DB.Fetch(where).next
					}
				}
			}
		}
		if DB.Fetch(obj).Exits != NOTHING {
			for DB.Fetch(obj).Exits != NOTHING && DB.Fetch(DB.Fetch(obj).Exits).Location != obj {
				lastwhere := DB.Fetch(obj).Exits
				DB.Fetch(obj).Exits = DB.Fetch(lastwhere).next
				DB.Fetch(lastwhere).next = NOTHING
				readd_contents(lastwhere)
			}
			if where := DB.Fetch(obj).Exits; where != NOTHING {
				for DB.Fetch(where).next != NOTHING {
					if DB.Fetch(DB.Fetch(where).next).Location != obj {
						lastwhere := DB.Fetch(where).next
						DB.Fetch(where).next = DB.Fetch(lastwhere).next
						DB.Fetch(lastwhere).next = NOTHING
						readd_contents(lastwhere)
					} else {
						where = DB.Fetch(where).next
					}
				}
			}
		}
	}
}

func check_common(id ObjectID) {
	o := DB.Fetch(id)
	if !o.name {
		o.NowCalled(fmt.Sprintf("Unknown%d", id))
	}
	if o.Location >= db_top {
		o.MoveTo(tp_player_start)
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

func check_room(obj ObjectID) {
	if DB.Fetch(obj).(ObjectID) >= db_top || (DB.Fetch(DB.Fetch(obj).(ObjectID)).flags & TYPE_MASK != TYPE_ROOM && DB.Fetch(obj).sp != NOTHING && DB.Fetch(obj).sp != HOME) {
		DB.Fetch(obj).sp = NOTHING
	}

	if DB.Fetch(obj).Exits < db_top {
		nexted[DB.Fetch(obj).Exits] = obj
	} else {
		DB.Fetch(obj).Exits = NOTHING
	}

	if DB.Fetch(obj).Owner >= db_top || (DB.Fetch(DB.Fetch(obj).Owner).flags & TYPE_MASK != TYPE_PLAYER) {
		DB.Fetch(obj).GiveTo(GOD)
	}
}

func check_exit(ref ObjectID) {
	obj := DB.Fetch(ref)
	for i, v := range obj.(Exit).Destinations {
		if v >= db_top {
			obj.(Exit).Destinations[i] = NOTHING
		}
	}
	if obj.Owner >= db_top || (DB.Fetch(obj.Owner).flags & TYPE_MASK != TYPE_PLAYER) {
		obj.GiveTo(GOD)
	}
}

func check_player(ref ObjectID) {
	obj := DB.Fetch(ref)
	player := obj.(Player)
	if player.Home >= db_top || (DB.Fetch(player.Home).flags & TYPE_MASK != TYPE_ROOM) {
		player.LiveAt(tp_player_start)
	}

	if obj.Exits < db_top {
		nexted[obj.Exits] = obj
	} else {
		obj.Exits = NOTHING
	}
}

func check_program(obj ObjectID) {
	if DB.Fetch(obj).Owner >= db_top || (DB.Fetch(DB.Fetch(obj).Owner).flags & TYPE_MASK != TYPE_PLAYER) {
		DB.Fetch(obj).GiveTo(GOD)
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
		log.Printf("Usage: %v infile outfile\n", argv[0])
		return
	case os.Args[1] == os.Args[2]:
		log.Printf("%v: input and output files can't have same name.\n", os.Args[0])
		return
	default:
		if f, input_file := os.Open(os.Args[1]); e != nil {
			log.Printf("%v: unable to open input file.\n", os.Args[1])
		} else {
			if output_file, e := os.OpenFile(os.Args[2], os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0755); e != nil {
				log.Printf("%v: unable to write to output file.\n", os.Args[2])
			} else {
				db_free()
				db_read(input_file)

				nexted := make([]int, db_top + 1)
				EachObject(func(obj ObjectID, o *Object) {
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
func do_compile(int descr, ObjectID p, ObjectID pr, int force_err_disp) {}

func add_event(descr int, player, loc, trig ObjectID, dtime int, program ObjectID, fr *frame, strdata string) {}

func init_primitives() {}

func add_player(who ObjectID) {}

func do_parse_mesg(descr int, player, what ObjectID, inbuf, abuf, outbuf string, outbuflen int, mesgtyp int) string {
	return ""
}