/* Editor routines --- Also contains routines to handle input */

/* This routine determines if a player is editing or running an interactive
   command.  It does it by checking the frame pointer field of the player ---
   if the program counter is NULL, then the player is not running anything
   The reason we don't just check the pointer but check the pc too is because
   I plan to leave the frame always on to save the time required allocating
   space each time a program is run.
   */

func get_program_line(program dbref, i int) (r *line) {
	for r = db.Fetch(program).(Program).first; i > 0 && r != nil; i-- {
		r = r.next
	}	
}

func interactive(descr int, player dbref, command string) {
	if db.Fetch(player).flags & READMODE != 0 {
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

func new_macro(name, definition string, player dbref) *macrotable {
	return &macrotable{ name: strings.ToLower(name), definition: definition, implementor: player }
}

func grow_macro_tree(struct macrotable *node, struct macrotable *newmacro) (r bool) {
	switch value := strings.Compare(newmacro.name, node.name); {
	case value < 0:
		if node.left {
			r = grow_macro_tree(node.left, newmacro)
		} else {
			node.left = newmacro
			r = true
		}
	case node.right != nil:
		r = grow_macro_tree(node.right, newmacro)
	default:
		node.right = newmacro
		r = true
	}
	return
}

func insert_macro(macroname, macrodef string, player dbref, node **macrotable) (r int) {
	if newmacro := new_macro(macroname, macrodef, player); *node == nil {
		*node = newmacro
		r = 1
	} else {
		r = grow_macro_tree((*node), newmacro)
	}
	return
}

func do_list_tree(node *macrotable, first, last string, length int, player dbref) {
	static char buf[BUFFER_LEN];
	if node != nil {
		if strings.Compare(node.name[:len(first)], first) >= 0 {
			do_list_tree(node.left, first, last, length, player)
		}
		if strings.Compare(node.name[:len(first)], first) >= 0 && strings.Compare(node.name[:len(last)], last) <= 0 {
			if length > 0 {
				notify(player, fmt.Sprintf("%-16s %-16s  %s", node.name, db.Fetch(node.implementor).name, node.definition))
				buf = ""
			} else {
				blen := len(buf)
				buf[blen:] = fmt.Sprintf("%-16s", node.name)
				buf[sizeof(buf) - 1] = '\0'
				if len(buf) > 70 {
					notify(player, buf)
					buf = ""
				}
			}
		}
		if strings.Compare(last, node.name[:len(last)]) >= 0 {
			do_list_tree(node.right, first, last, length, player)
		}
		if node == macrotop && !length {
			notify(player, buf)
			buf = ""
		}
	}
}

func list_macros(words []string, k int, player dbref, length int) {
	if k == 0 {
		do_list_tree(macrotop, "\001", "\377", length, player)
	} else {
		k--
		do_list_tree(macrotop, words[0], words[k], length, player)
	}
	notify(player, "End of list.")
}

func erase_node(oldnode, node *macrotable, killname string, mtop *macrotable) (r bool) {
	switch {
	case node == nil:
	case strings.Compare(killname, node.name) < 0:
		r = erase_node(node, node.left, killname, mtop)
	case strings.Compare(killname, node.name) > 0:
		r = erase_node(node, node.right, killname, mtop)
	default:
		if node == oldnode.left {
			oldnode.left = node.left
			if node.right != nil {
				grow_macro_tree(mtop, node.right)
			}
		} else {
			oldnode.right = node.right
			if node.left {
				grow_macro_tree(mtop, node.left)
			}
		}
		free((void *) node)
		r= true
	}
	return
}

func kill_macro(const char *macroname, dbref player, struct macrotable **mtop) (r bool) {
	switch {
	case *mtop != nil:
	case macroname == (*mtop).name:
		macrotemp := *mtop
		var leftwards bool
		if (*mtop).left {
			leftwards = true
		}

		if leftwards {
			*mtop = (*mtop).left
			if *mtop != nil && macrotemp.right != nil {
				grow_macro_tree((*mtop), macrotemp.right)
			}
		} else {
			*mtop = (*mtop).right
			if *mtop != nil && macrotemp.left != nil {
				grow_macro_tree((*mtop), macrotemp.right)
			}
		}
		*macrotemp = nil
		r = true
	case erase_node(*mtop, *mtop, macroname, *mtop)):
		r = true
	}
	return
}

/* The editor itself --- this gets called each time every time to
 * parse a command.
 */

func editor(descr int, player dbref, command string) {
	if db.Fetch(player).(Player).insert_mode {
		insert(player, command)	/* insert it! */
	} else {
		program := db.Fetch(player).(Player).curr_prog
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
				case insert_macro(words[1], words[2], player, &macrotop):
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
					if kill_macro(words[0], player, &macrotop) {
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
func do_insert(player, program dbref, line int) {
	db.Fetch(player).(Player).insert_mode = true
	db.Fetch(program).(Program).curr_line = line - 1
}

func delete_program_line(program dbref, curr *line) *line {
	if curr.prev != nil {
		curr.prev.next = curr.next
	} else {
		db.Fetch(program).(Program).first = curr.next
	}
	if curr.next != nil {
		curr.next.prev = curr.prev
	}
	return curr.next
}

func delete_program_lines(player dbref, curr *line, count int) {
	if curr != nil {
		db.Fetch(program).(Program).curr_line = curr
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
func do_delete(player, program dbref, args []int) {
	switch len(args) {
	case 0:
		delete_program_lines(player, db.Fetch(program).(Program).curr_line, 1)
	case 1:
		delete_program_lines(player, get_program_line(program, args[0] - 1), 1)
	case 2:
		delete_program_lines(player, get_program_line(program, args[0] - 1), args[1] - args[0])
	default:
		notify_nolisten(player, "Too many arguments!", true)
	}
}

/* quit from edit mode.  Put player back into the regular game mode */
func do_quit(player, program dbref) {
	log_status("PROGRAM SAVED: %s by %s(%d)", unparse_object(player, program), db.Fetch(player).name, player)
	write_program(db.Fetch(program).(Program).first, program)

	if tp_log_programs {
		log_program_text(db.Fetch(program).(Program).first, player, program)
	}

	db.Fetch(program).(Program).first = nil
	db.Fetch(program).flags &= ~INTERNAL
	db.Fetch(player).flags &= ~INTERACTIVE
	db.Fetch(player).(Player).curr_prog = NOTHING
	db.Fetch(player).flags |= OBJECT_CHANGED
	db.Fetch(program).flags |= OBJECT_CHANGED
}

func MatchAndList(descr int, player dbref, name, linespec string) {
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
	case !controls(player, thing) && db.Fetch(thing).flags & VEHICLE == 0:
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
		tmpline := db.Fetch(thing).(Program).first
		db.Fetch(thing).(Program).first = read_program(thing)
		do_list(player, thing, ranges)
		db.Fetch(thing).(Program).first = tmpline
	}
	return
}

func print_program_line(player dbref, n int, curr *line) {
	if db.Fetch(player).flags & INTERNAL == 1 {
		notify_nolisten(player, fmt.Sprintf("%3d: %s", n, curr.this_line), true)
	} else {
		notify_nolisten(player, fmt.Sprint(curr.this_line), true)
	}
}

func print_program_lines(player dbref, i int, curr *line, count int) {
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
func do_list(player, program dbref, args []int) {
	switch len(args) {
	case 0:
		print_program_lines(player, db.Fetch(program).(Program).curr_line, 1)
	case 1:
		print_program_lines(player, get_program_line(program, args[0] - 1), 1)
	case 2:
		print_program_lines(player, get_program_line(program, args[0] - 1), args[1] - args[0])
	default:
		notify_nolisten(player, "Too many arguments!", true)
	}
}

func val_and_head(player dbref, header int) {
	switch program := dbref(header); {
	case !valid_reference(program), !IsProgram(program):
		notify(player, "That isn't a program.")
	case !(controls(player, program) || Linkable(program)):
		notify(player, "That's not a public program.")
	default:
		do_list_header(player, program)
	}
}

func do_list_header(player, program dbref) {
	for curr := read_program(program); curr != nil && curr.this_line[0] == '('; curr = curr.next {
		notify(player, curr.this_line)
	}
	notify(player, "Done.")
}

func list_publics(descr int, player dbref, args []int) {
	var program dbref
	switch len(args) {
	case 0:
		program = db.Fetch(player).(Player).curr_prog
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
		if db.Fetch(program).(Program).code == nil {
			if program == db.Fetch(player).(Player).curr_prog {
				do_compile(descr, db.Fetch(program).Owner, program, 0)
			} else {
				tmpline := db.Fetch(program).(Program).first
				db.Fetch(program).(Program).first = read_program(program)
				do_compile(descr, db.Fetch(program).Owner, program, 0)
				db.Fetch(program).(Program).first = tmpline
			}
			if db.Fetch(program).(Program).code == nil {
				notify(player, "Program not compilable.")
				return
			}
		}
		do_list_publics(player, program)
	}
}

func do_list_publics(player, program dbref) {
	notify(player, "PUBLIC funtions:")
	for ptr := db.Fetch(program).(Program).PublicAPI; ptr != nil; ptr = ptr.next {
		notify(player, ptr.subname)
	}
}

func toggle_numbers(player dbref, args []int) {
	switch {
	case len(args):
		if args[0] == 0 {
			db.Fetch(player).flags &= ~INTERNAL
			notify(player, "Line numbers off.")
		} else {
			db.Fetch(player).flags |= INTERNAL
			notify(player, "Line numbers on.")
		}
	case db.Fetch(player).flags & INTERNAL != 0:
		db.Fetch(player).flags &= ~INTERNAL
		notify(player, "Line numbers off.")
	default:
		db.Fetch(player).flags |= INTERNAL
		notify(player, "Line numbers on.")
	}
}

/* insert this line into program */
func insert(player dbref, line string) {
	program := db.Fetch(player).(Player).curr_prog
	if line == EXIT_INSERT {
		db.Fetch(player).(Player).insert_mode = false
		notify_nolisten(player, "Exiting insert mode.", true)
		return
	}
	i := db.Fetch(program).(Program).curr_line - 1
	curr := get_program_line(program, i)

	new_line := get_new_line();	/* initialize line */
	if line == "" {
		new_line.this_line = " "
	} else {
		new_line.this_line = line
	}
	switch {
	case !db.Fetch(program).(Program).first:	/* nothing --- insert in front */
		db.Fetch(program).(Program).first = new_line
		db.Fetch(program).(Program).curr_line = 2
	case curr == nil:				/* insert at the end */
		i = 1
		for curr = db.Fetch(program).(Program).first; curr.next; curr = curr.next {
			i++				/* count lines */
		}
		db.Fetch(program).(Program).curr_line = i + 2
		new_line.prev = curr
		curr.next = new_line
	case !db.Fetch(program).(Program).curr_line:	/* insert at the beginning */
		db.Fetch(program).(Program).curr_line = 1
		new_line.next = db.Fetch(program).(Program).first
		db.Fetch(program).(Program).first = new_line
	default:
		/* inserting in the middle */
		db.Fetch(program).(Program).curr_line = db.Fetch(program).(Program).curr_line + 1
		new_line.prev = curr
		new_line.next = curr.next
		if new_line.next != nil {
			new_line.next.prev = new_line
		}
		curr.next = new_line
	}
}