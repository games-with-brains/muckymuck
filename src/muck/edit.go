/* Editor routines --- Also contains routines to handle input */

/* This routine determines if a player is editing or running an interactive
   command.  It does it by checking the frame pointer field of the player ---
   if the program counter is NULL, then the player is not running anything
   The reason we don't just check the pointer but check the pc too is because
   I plan to leave the frame always on to save the time required allocating
   space each time a program is run.
   */

func get_program_line(program ObjectID, i int) (r *line) {
	for r = DB.Fetch(program).(Program).first; i > 0 && r != nil; i-- {
		r = r.next
	}	
}

func interactive(descr int, player ObjectID, command string) {
	if DB.Fetch(player).flags & READMODE != 0 {
		/*
		 * process command, push onto stack, and return control to forth
		 * program
		 */
		handle_read_event(descr, player, command);
	} else {
		editor(descr, player, command)
	}
}

func macro_expansion(node *macrotable, match string) (r string) {
	if node != nil {
		switch value := strings.Compare(match, node.name); value {
		case -1:
			r = macro_expansion(node.left, match)
		case 0:
			r = node.definition
		case 1:
			r = macro_expansion(node.right, match)
		}
	}
	return
}

/* The editor itself --- this gets called each time every time to
 * parse a command.
 */

func editor(descr int, player ObjectID, command string) {
	if p := DB.FetchPlayer(player); p.insert_mode {
		insert(player, command)	/* insert it! */
	} else {
		program := p.curr_prog
		var words []string
		var args []int
		var i int
		for ; command != ""; i++ {
			command = strings.TrimLeftFunc(command, unicode.IsSpace)
			items := strings.SplitN(command, " ", 2)
			words = append(words, items[0])
			if i == 1 && words[0] == "def" {
				switch words[2] = strings.TrimLeftFunc(command, unicode.IsSpace); {
				case words[2] == "":
					notify(player, "Invalid definition syntax.")
				case insert_macro(words[1], words[2], player, Macros):
					notify(player, "Entry created.")
				default:
					notify(player, "That macro already exists!")
				}
				return
			}
			if args[i] = strconv.Atoi(items[0]); args[i] < 0 {
				notify(player, "Negative arguments not allowed!")
				return
			}
		}
		for i--; i >= 0 && words[i] == 0; i-- {}
		if i > -1 {
			switch words[i][0] {
			case KILL_COMMAND:
				if !Wizard(player) {
					notify(player, "I'm sorry Dave, but I can't let you do that.")
				} else {
					if kill_macro(words[0], player, Macros) {
						notify(player, "Macro entry deleted.")
					} else {
						notify(player, "Macro to delete not found.")
					}
				}
			case SHOW_COMMAND:
				list_macros(words, i, player, 1)
			case SHORTSHOW_COMMAND:
				list_macros(words, i, player, 0)
			case INSERT_COMMAND:
				do_insert(player, program, args[0])
				notify(player, "Entering insert mode.")
			case DELETE_COMMAND:
				do_delete(player, program, args)
			case QUIT_EDIT_COMMAND:
				do_quit(player, program)
				notify(player, "Editor exited.")
			case COMPILE_COMMAND:
				/* FIXME: compile code belongs in compile.c, not in the editor */
				do_compile(descr, player, program, 1)
				notify(player, "Compiler done.")
			case LIST_COMMAND:
				do_list(player, program, args)
			case EDITOR_HELP_COMMAND:
				spit_file(player, EDITOR_HELP_FILE)
			case VIEW_COMMAND:
				val_and_head(player, args[0])
			case UNASSEMBLE_COMMAND:
				disassemble(player, program)
			case NUMBER_COMMAND:
				toggle_numbers(player, args)
			case PUBLICS_COMMAND:
				list_publics(descr, player, args)
			default:
				notify(player, "Illegal editor command.")
			}
		}		
	}
}

/* puts program into insert mode */
func do_insert(player, program ObjectID, line int) {
	DB.FetchPlayer(player).insert_mode = true
	DB.Fetch(program).(Program).curr_line = line - 1
}

func delete_program_line(program ObjectID, curr *line) *line {
	if curr.prev != nil {
		curr.prev.next = curr.next
	} else {
		DB.Fetch(program).(Program).first = curr.next
	}
	if curr.next != nil {
		curr.next.prev = curr.prev
	}
	return curr.next
}

func delete_program_lines(player ObjectID, curr *line, count int) {
	if curr != nil {
		DB.Fetch(program).(Program).curr_line = curr
		n := count
		for n > 0 && curr != nil {
			curr = delete_program_line(program, curr)
			n--
		}
		notify(player, fmt.Sprintf("%d lines deleted", count - n)
	} else
		notify(player, "No line to delete!")
	}
}

/* deletes line n if one argument,
   lines arg1 -- arg2 if two arguments
   current line if no argument */
func do_delete(player, program ObjectID, args []int) {
	switch len(args) {
	case 0:
		delete_program_lines(player, DB.Fetch(program).(Program).curr_line, 1)
	case 1:
		delete_program_lines(player, get_program_line(program, args[0] - 1), 1)
	case 2:
		delete_program_lines(player, get_program_line(program, args[0] - 1), args[1] - args[0])
	default:
		notify_nolisten(player, "Too many arguments!", true)
	}
}

/* quit from edit mode.  Put player back into the regular game mode */
func do_quit(player, program ObjectID) {
	log_status("PROGRAM SAVED: %s by %s(%d)", unparse_object(player, program), DB.Fetch(player).name, player)
	write_program(DB.Fetch(program).(Program).first, program)

	if tp_log_programs {
		log_program_text(DB.Fetch(program).(Program).first, player, program)
	}

	DB.Fetch(program).(Program).first = nil
	DB.Fetch(program).flags &= ~INTERNAL
	DB.Fetch(player).flags &= ~INTERACTIVE
	DB.FetchPlayer(player).curr_prog = NOTHING
	DB.Fetch(player).flags |= OBJECT_CHANGED
	DB.Fetch(program).flags |= OBJECT_CHANGED
}

func MatchAndList(descr int, player ObjectID, name, linespec string) {
	thing := NewMatch(descr, player, name, IsProgram).
		MatchNeighbor().
		MatchPossession().
		MatchRegistered().
		MatchAbsolute().
		NoisyMatchResult()
	switch {
	case thing == NOTHING:
	case Typeof(thing) != TYPE_PROGRAM:
		notify(player, "You can't list anything but a program.")
	case !controls(player, thing) && DB.Fetch(thing).flags & VEHICLE == 0:
		notify(player, "Permission denied. (You don't control the program, and it's not set Viewable)")
	default:
		var ranges []int
		if linespec == "" {
			ranges = append(ranges, 1)
			ranges = append(ranges, -1)
		} else {
			items := strings.SplitN(strings.TrimLeftFunc(linespec, unicode.IsSpace), 2)
			if linespec = items[0]; isdigit(linespec[0]) {
				ranges = append(ranges, strconv.Atoi(linespec))
			} else {
				ranges = append(ranges, 1)
			}
			if linespec = strings.TrimFunc(items[1], unicode.IsSpace); linespec != "" {
				ranges = append(ranges, strconv.Atoi(linespec))
			} else {
				ranges = append(ranges, -1)
			}
		}
		tmpline := DB.Fetch(thing).(Program).first
		DB.Fetch(thing).(Program).first = read_program(thing)
		do_list(player, thing, ranges)
		DB.Fetch(thing).(Program).first = tmpline
	}
	return
}

func print_program_line(player ObjectID, n int, curr *line) {
	if DB.Fetch(player).flags & INTERNAL == 1 {
		notify_nolisten(player, fmt.Sprintf("%3d: %s", n, curr.this_line), true)
	} else {
		notify_nolisten(player, fmt.Sprint(curr.this_line), true)
	}
}

func print_program_lines(player ObjectID, i int, curr *line, count int) {
	if curr != nil {
		switch {
		case count == -1:
			var n int
			for ; curr != nil; curr = curr.next {
				print_program_line(player, n + i, curr)
				n++
			}
			notify_nolisten(player, fmt.Sprintf("%v lines displayed.", i), true)			
		 case count > -1:
			for i := count; curr != nil && i > 0; curr = curr.next {
				print_program_line(player, count, curr)
				i--
			}
			notify_nolisten(player, fmt.Sprintf("%v line displayed.", count - i), true)
		default:
			notify_nolisten(player, "No lines to display.", true)
		}
	} else {
		notify_nolisten(player, "No lines to display.", true)
	}
}

/* list --- if no argument, redisplay the current line
   if 1 argument, display that line
   if 2 arguments, display all in between   */
func do_list(player, program ObjectID, args []int) {
	switch len(args) {
	case 0:
		print_program_lines(player, DB.Fetch(program).(Program).curr_line, 1)
	case 1:
		print_program_lines(player, get_program_line(program, args[0] - 1), 1)
	case 2:
		print_program_lines(player, get_program_line(program, args[0] - 1), args[1] - args[0])
	default:
		notify_nolisten(player, "Too many arguments!", true)
	}
}

func val_and_head(player ObjectID, header int) {
	switch program := ObjectID(header); {
	case !program.IsValid(), !IsProgram(program):
		notify(player, "That isn't a program.")
	case !(controls(player, program) || Linkable(program)):
		notify(player, "That's not a public program.")
	default:
		do_list_header(player, program)
	}
}

func do_list_header(player, program ObjectID) {
	for curr := read_program(program); curr != nil && curr.this_line[0] == '('; curr = curr.next {
		notify(player, curr.this_line)
	}
	notify(player, "Done.")
}

func list_publics(descr int, player ObjectID, args []int) {
	var program ObjectID
	switch len(args) {
	case 0:
		program = DB.FetchPlayer(player).curr_prog
	case 1:
		program = args[0]
	default:
		notify(player, "I don't understand which program you want to list PUBLIC functions for.")
		return
	}

	switch {
	case Typeof(program) != TYPE_PROGRAM:
		notify(player, "That isn't a program.")
	case !controls(player, program) && !Linkable(program):
		notify(player, "That's not a public program.")
	default:
		if DB.Fetch(program).(Program).code == nil {
			if program == DB.FetchPlayer(player).curr_prog {
				do_compile(descr, DB.Fetch(program).Owner, program, 0)
			} else {
				tmpline := DB.Fetch(program).(Program).first
				DB.Fetch(program).(Program).first = read_program(program)
				do_compile(descr, DB.Fetch(program).Owner, program, 0)
				DB.Fetch(program).(Program).first = tmpline
			}
			if DB.Fetch(program).(Program).code == nil {
				notify(player, "Program not compilable.")
				return
			}
		}
		do_list_publics(player, program)
	}
}

func do_list_publics(player, program ObjectID) {
	notify(player, "PUBLIC funtions:")
	for ptr := DB.Fetch(program).(Program).PublicAPI; ptr != nil; ptr = ptr.next {
		notify(player, ptr.subname)
	}
}

func toggle_numbers(player ObjectID, args []int) {
	switch {
	case len(args):
		if args[0] == 0 {
			DB.Fetch(player).flags &= ~INTERNAL
			notify(player, "Line numbers off.")
		} else {
			DB.Fetch(player).flags |= INTERNAL
			notify(player, "Line numbers on.")
		}
	case DB.Fetch(player).flags & INTERNAL != 0:
		DB.Fetch(player).flags &= ~INTERNAL
		notify(player, "Line numbers off.")
	default:
		DB.Fetch(player).flags |= INTERNAL
		notify(player, "Line numbers on.")
	}
}

/* insert this line into program */
func insert(player ObjectID, line string) {
	program := DB.FetchPlayer(player).curr_prog
	if line == EXIT_INSERT {
		DB.FetchPlayer(player).insert_mode = false
		notify_nolisten(player, "Exiting insert mode.", true)
		return
	}
	i := DB.Fetch(program).(Program).curr_line - 1
	curr := get_program_line(program, i)

	new_line := get_new_line();	/* initialize line */
	if line == "" {
		new_line.this_line = " "
	} else {
		new_line.this_line = line
	}
	switch {
	case !DB.Fetch(program).(Program).first:	/* nothing --- insert in front */
		DB.Fetch(program).(Program).first = new_line
		DB.Fetch(program).(Program).curr_line = 2
	case curr == nil:				/* insert at the end */
		i = 1
		for curr = DB.Fetch(program).(Program).first; curr.next; curr = curr.next {
			i++				/* count lines */
		}
		DB.Fetch(program).(Program).curr_line = i + 2
		new_line.prev = curr
		curr.next = new_line
	case !DB.Fetch(program).(Program).curr_line:	/* insert at the beginning */
		DB.Fetch(program).(Program).curr_line = 1
		new_line.next = DB.Fetch(program).(Program).first
		DB.Fetch(program).(Program).first = new_line
	default:
		/* inserting in the middle */
		DB.Fetch(program).(Program).curr_line = DB.Fetch(program).(Program).curr_line + 1
		new_line.prev = curr
		new_line.next = curr.next
		if new_line.next != nil {
			new_line.next.prev = new_line
		}
		curr.next = new_line
	}
}