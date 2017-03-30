package fbmuck

func prim_pop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		POP()
	})
}

func prim_dup(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	checkop_readonly(1, top)
	CHECKOFLOW(1)
	push(arg, top, arg[*top - 1].Dup())
}

func prim_popn(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if count, ok := op[0].(int); !ok {
			panic("Operand not an integer.")
		} else {
			if count < 0 {
				panic("Operand is negative.")
			}
			for i := count; i > 0; i-- {
				checkop(1, top)
				POP()
			}
		}
	})
}

func prim_dupn(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if result, ok := op[0].(int); !ok {
			panic("Operand is not an integer.")
		} else {
			if result < 0 {
				panic("Operand is negative.")
			}
			checkop(result, top)
			CHECKOFLOW(result)
			for i := result; i > 0; i-- {
				push(arg, top, arg[*top - result].Dup())
			}
		}
	})
}

func prim_ldup(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	checkop_readonly(1, top)
	if result, ok := arg[*top - 1].data.(int); !ok {
		panic("Operand is not an integer.")
	} else {
		if result < 0 {
			panic("Operand is negative.")
		}
		result++
		checkop_readonly(result, top)
		CHECKOFLOW(result)
		for i := result; i > 0; i-- {
			push(arg, top, arg[*top - result].Dup())
		}
	}
}

func prim_at(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		switch data := (*op[0]).data.(type) {
		case PROG_LVAR:
			if data < 0 || data >= MAX_VAR {
				panic("Variable number out of range.")
			}
			push(arg, top, localvars_get(fr, program).lvars[data].Dup())
		case PROG_VAR:
			if data < 0 || data >= MAX_VAR {
				panic("Variable number out of range.")
			}
			push(arg, top, fr.variables[data].Dup())
		case PROG_SVAR:		/* SCOPEDVAR */
			if data < 0 || data >= MAX_VAR {
				panic("Variable number out of range.")
			}
			if tmp := scopedvar_get(fr, 0, data); tmp == nil {
				panic("Scoped variable number out of range.")
			} else {
				push(arg, top, tmp.Dup())
			}
		default:
			panic("Non-variable argument.")
		}
	})
}

func prim_bang(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch data := op[1].(type) {
		case PROG_VAR:
			if data < 0 || data >= MAX_VAR {
				panic("Variable number out of range. (2)")
			}
			fr.variables[data] = oper2.Dup()
		case PROG_LVAR:
			if data < 0 || data >= MAX_VAR {
				panic("Variable number out of range. (2)")
			}
			localvars_get(fr, program).lvars[data] = op[0].Dup()
		case PROG_SVAR:
			if data < 0 || data >= MAX_VAR {
				panic("Variable number out of range. (2)")
			}
			if tmp := scopedvar_get(fr, 0, data); tmp == nil {
				panic("Scoped variable number out of range.")
			} else {
				tmp = op[0].Dup()
			}
		default:
			panic("Non-variable argument (2)")
		}
	})
}

func prim_var(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, op[0].(int))
	})
}

func prim_localvar(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, op[0].(int))
	})
}

func prim_swap(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, op[1])
		push(arg, top, op[0])
	})
}

func prim_over(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	checkop_readonly(2, top)
	CHECKOFLOW(1)
	push(arg, top, arg[*top - 2].Dup())
}

func prim_pick(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		depth := *op[0].(int)
		if depth <= 0 {
			panic("Operand not a positive integer.")
		}
		checkop_readonly(depth, top)
		push(arg, top, arg[*top - depth].Dup())
	})
}

func prim_put(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		depth := op[1].(int)
		if depth <= 0 {
			panic("Operand not a positive integer.")
		}
		checkop(depth, top)
		arg[*top - tmp] = op[0].Dup()
	})
}

func prim_rot(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		push(arg, top, op[1])
		push(arg, top, op[2])
		push(arg, top, op[0])
	})
}

func prim_rotate(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		tmp = op[0].(int);	/* Depth on stack */
		checkop(abs(tmp), top)
		switch {
		case tmp > 0:
			temp2 = arg[*top - tmp]
			for ; tmp > 0; tmp-- {
				arg[*top - tmp] = arg[*top - tmp + 1]
			}
			arg[*top - 1] = temp2
		case tmp < 0:
			temp2 = arg[*top - 1]
			for tmp = -1; tmp > oper1.data.(int); tmp-- {
				arg[*top + tmp] = arg[*top + tmp - 1]
			}
			arg[*top + tmp] = temp2
		}
	})
}

func prim_dbtop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		ref = (dbref) db_top
		CHECKOFLOW(1)
		push(arg, top, ref)
	})
}

func prim_depth(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, *top)
	})
}

func prim_version(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, VERSION)
	})
}

func prim_prog(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, program)
	})
}

func prim_trig(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, fr.trig)
	})
}

func prim_caller(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, fr.caller.st[fr.caller.top - 1])
	})
}

func prim_intp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		_, ok := op[0].(int)
		push(arg, top, MUFBool(ok))
	})
}

func prim_floatp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		_, ok := op[0].(float64)
		push(arg, top, MUFBool(ok))
	})
}

func prim_arrayp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		_, ok := op[0].(Array)
		push(arg, top, MUFBool(ok))
	})
}

func prim_dictionaryp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		_, ok := op[0].(Dictionary)
		push(arg, top, MUFBool(ok))
	})
}

func prim_stringp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		_, ok := op[0].(string)
		push(arg, top, MUFBool(ok))
	})
}

func prim_dbrefp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		_, ok := op[0].(dbref)
		push(arg, top, MUFBool(ok))
	})
}

func prim_addressp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		_, ok := op[0].(Address)
		push(arg, top, MUFBool(ok))
	})
}

func prim_lockp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		_, ok := op[0].(*boolexp)
		push(arg, top, MUFBool(ok))
	})
}

func abort_checkargs(stackpos int, msg string) (r string) {
	if *top == stackpos + 1 {
		r = fmt.Sprintf("%s (top)", msg)
	} else {
		abort_interp(fmt.Sprintf("%s (top-%d)", msg, ((*top) - stackpos - 1)))
	}
}

type RangeStackFrame struct {
	is_a_repeat bool
	pos int
	count int
	next *RangeStackFrame
}

func prim_checkargs(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	var range_stack *RangeStackFrame

	var zbuf string

	apply_primitive(1, top, func(op Array) {
		if buf := op[0].(string); buf != "" {
			currpos := len(buf) - 1
			stackpos := *top - 1
			for currpos >= 0 {
				switch {
				case isdigit(buf[currpos]):
					tmp = 1
					result = 0
					for currpos >= 0 && isdigit(buf[currpos]) {
						result = result + (tmp * (buf[currpos] - '0'))
						tmp = tmp * 10
						currpos--
					}
					switch {
					case result == 0:
						panic("Bad multiplier '0' in argument expression.")
					case result >= STACK_SIZE:
						panic("Multiplier too large in argument expression.")
					}
					range_stack = &RangeStackFrame{ is_a_repeat: true, pos: currpos, count: result, next: range_stack })
				case buf[currpos] == '}':
					switch {
					case stackpos < 0:
						abort_checkargs(stackpos, "Stack underflow.")
					}
					if result = arg[stackpos].data.(int); result < 0 {
						abort_checkargs(stackpos, "Range counter should be non-negative.")
					}
					range_stack = &RangeStackFrame{ pos: currpos -1, count: result, next: range_stack })
					currpos--
					if result == 0 {
						for currpos > 0 && buf[currpos] != '{' {
							currpos--
						}
					}
					stackpos--
				case buf[currpos] == '{':
					switch {
					case range_stack == nil:
						panic("Mismatched { in argument expression")
					case range_stack.is_a_repeat:
						panic("Misformed argument expression.")
					}
					if range_stack.count--; range_stack.count > 0:
						currpos = range_stack.pos
					} else {
						range_stack = range_stack.next
						currpos--
						if range_stack != nil && range_stack.is_a_repeat {
							range_stack.count--
							if range_stack.count > 0 {
								currpos = range_stack.pos
							} else {
								range_stack = range_stack.next
							}
						}
					}
				default:
					switch buf[currpos] {
					case 'i':
						if stackpos < 0 {
							abort_checkargs(stackpos, "Stack underflow.")
						}
						if _, ok := arg[stackpos].data.(int) {
							abort_checkargs(stackpos, "Expected an integer.")
						}
					case 'n':
						if stackpos < 0 {
							abort_checkargs(stackpos, "Stack underflow.")
						}
						if _, ok := arg[stackpos].data.(float64) {
							abort_checkargs(stackpos, "Expected a float.")
						}
					case 's', 'S':
						if stackpos < 0 {
							abort_checkargs(stackpos, "Stack underflow.")
						}
						if v, ok := arg[stackpos].data.(string); !ok {
							abort_checkargs(stackpos, "Expected a string.")
						} else {
							if buf[currpos] == 'S' && v == "" {
								abort_checkargs(stackpos, "Expected a non-null string.")
							}
						}
					case 'd', 'p', 'r', 't', 'e', 'f', 'D', 'P', 'R', 'T', 'E', 'F':
						if stackpos < 0 {
							abort_checkargs(stackpos, "Stack underflow.")
						}
						if ref, ok := arg[stackpos].data.(objref); !ok {
							abort_checkargs(stackpos, "Expected a dbref.")
						} else {
							if ref >= db_top || ref < HOME {
								abort_checkargs(stackpos, "Invalid dbref.")
							}
						}
						switch buf[currpos] {
						case 'D':
							if ref < 0 && ref != HOME {
								abort_checkargs(stackpos, "Invalid dbref.")
							}
							fallthrough
						case 'd':
							if ref < HOME {
								abort_checkargs(stackpos, "Invalid dbref.")
							}
						case 'P':
							if ref < 0 {
								abort_checkargs(stackpos, "Expected player dbref.")
							}
							fallthrough
						case 'p':
							switch {
							case ref >= 0 && Typeof(ref) != TYPE_PLAYER:
								abort_checkargs(stackpos, "Expected player dbref.")
							case ref == HOME:
								abort_checkargs(stackpos, "Expected player dbref.")
							}
						case 'R':
							if ref < 0 && ref != HOME {
								abort_checkargs(stackpos, "Expected room dbref.")
							}
							fallthrough
						case 'r':
							if ref >= 0 && Typeof(ref) != TYPE_ROOM {
								abort_checkargs(stackpos, "Expected room dbref.")
							}
						case 'T':
							if ref < 0 {
								abort_checkargs(stackpos, "Expected thing dbref.")
							}
							fallthrough
						case 't':
							switch {
							case ref >= 0 && Typeof(ref) != TYPE_THING:
								abort_checkargs(stackpos, "Expected thing dbref.")
							case ref == HOME:
								abort_checkargs(stackpos, "Expected player dbref.")
							}
						case 'E':
							if ref < 0 {
								abort_checkargs(stackpos, "Expected exit dbref.")
							}
							fallthrough
						case 'e':
							switch {
							case ref >= 0 && Typeof(ref) != TYPE_EXIT:
								abort_checkargs(stackpos, "Expected exit dbref.")
							case ref == HOME:
								abort_checkargs(stackpos, "Expected player dbref.")
							}
						case 'F':
							if ref < 0 {
								abort_checkargs(stackpos, "Expected program dbref.")
							}
							fallthrough
						case 'f':
							switch {
							case ref >= 0 && Typeof(ref) != TYPE_PROGRAM:
								abort_checkargs(stackpos, "Expected program dbref.")
							case ref == HOME:
								abort_checkargs(stackpos, "Expected player dbref.")
							}
						}
					case '?':
						if stackpos < 0 {
							abort_checkargs(stackpos, "Stack underflow.")
						}
					case 'l':
						if stackpos < 0 {
							abort_checkargs(stackpos, "Stack underflow.")
						}
						if _, ok := arg[stackpos].data.(*boolexp); !ok {
							abort_checkargs(stackpos, "Expected a lock boolean expression.")
						}
					case 'v':
						if stackpos < 0 {
							abort_checkargs(stackpos, "Stack underflow.")
						}
						switch arg[stackpos].(type) {
						case PROG_VAR, PROG_LVAR, PROG_SVAR:
						default:
							abort_checkargs(stackpos, "Expected a variable.")
						}
					case 'a':
						if stackpos < 0 {
							abort_checkargs(stackpos, "Stack underflow.")
						}
						if _, ok := arg[stackpos].data.(Address); !ok {
							abort_checkargs(stackpos, "Expected a function address.")
						}
					case ' ':
						/* this is meaningless space.  Ignore it. */
						stackpos++
					default:
						panic("Unkown argument type in expression.")
					}

					currpos--			/* decrement string index */
					stackpos--			/* move on to next stack item down */

					/* are we expecting a repeat of the last argument or range? */
					if range_stack != nil && range_stack.is_a_repeat {
						/* is the repeat done yet? */
						range_stack.count--
						if range_stack != nil {
							/* no, repeat last argument or range */
							currpos = range_stack.pos
						} else {
							/* yes, we're done with this repeat */
							range_stack = range_stack.next
						}
					}
				}
			}							/* while loop */

			if range_stack != nil {
				panic("Badly formed argument expression.")
			}
		}
	})
}

func prim_mode(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		result = fr.multitask
		CHECKOFLOW(1)
		push(arg, top, result)
	})
}

func prim_mark(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, Mark{})
	})
}

func prim_findmark(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		depth := 1
		height := *top - 1
		for height >= 0 {
			if _, ok := arg[height].(Mark); ok {
				break
			} else {
				height--
				depth++
			}
		}
		count := depth - 1
		if height < 0
			panic("No matching mark on stack!")
		}
		if depth > 1 {
			temp2 = arg[*top - depth]
			for ; depth > 1; depth-- {
				arg[*top - depth] = arg[*top - depth + 1]
			}
			arg[*top - 1] = temp2
		}
		POP()
		push(arg, top, count)
	})
}

func prim_setmode(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	spply_primitive(1, top, func() {
		switch op[0].(int) {
		case BACKGROUND:
			fr.been_background = 1
			fr.writeonly = 1
		case FOREGROUND:
			if fr.been_background {
				panic("Cannot FOREGROUND a BACKGROUNDed program.")
			}
		case PREEMPT:
		default:
			panic("Invalid mode.")
		}
		fr.multitask = result
	})
}

func prim_interp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		program := valid_object(op[0])
		trigger := valid_remote_object(player, mlev, op[1])
		match_args = op[2].(string)

		switch {
		case Typeof(program) != TYPE_PROGRAM:
			panic("Bad program reference. (1)")
		case mlev < MASTER && !permissions(ProgUID, trigger):
			panic("Permission denied.")
		case fr.level > 8:
			panic("Interp call loops not allowed.")
		}

		buf := match_args
		var rv *inst
		if tmpfr := interp(fr.descr, player, db.Fetch(player).location, program, trigger, PREEMPT, STD_HARDUID, 0); tmpfr != nil {
			rv = interp_loop(player, oper1->data.objref, tmpfr, true)
		}
		match_args = buf

		if rv != nil {
			switch rv := rv.(type) {
			case int:
				push(arg, top, rv)
			case string:
				push(arg, top, rv)
			default:
				push(arg, top, "")
			}
		} else {
			push(arg, top, "")
		}
	})
}

func prim_for(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		if fr.forstack.top >= STACK_SIZE {
			panic("Too many nested FOR loops.")
		}
		fr.forstack.top++
		fr.forstack.st = push_for(fr.forstack.st)
		fr.forstack.st.cur = op[0].(int)
		fr.forstack.st.end = op[1].(int)
		fr.forstack.st.step = op[2].(int)
		fr.forstack.st.didfirst = 0

		if fr.trys.st {
			fr.trys.st.for_count++
		}
	})
}

func prim_foreach(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if fr.forstack.top >= STACK_SIZE {
			panic("Too many nested FOR loops.")
		}
		arr := op[0].(stk_array)
		fr.forstack.top++
		fr.forstack.st = push_for(fr.forstack.st)

		fr.forstack.st.cur.line = 0
		fr.forstack.st.cur.data = 0

		if fr.trys.st != nil {
			fr.trys.st.for_count++
		}

		fr.forstack.st.end = arr.Dup()
		fr.forstack.st.step = 0
		fr.forstack.st.didfirst = 0
	})
}

func prim_foriter(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		if !fr.forstack.st {
			panic("Internal error; FOR stack underflow.")
		}

		if arr, ok := fr.forstack.st.end.(stk_array); ok {
			if fr.forstack.st.didfirst {
				fr.forstack.st.cur = array_next(arr, fr.forstack.st.cur)
				result = fr.forstack.st.cur != nil
			} else {
				fr.forstack.st.cur = array_first(arr)
				result = arr != nil
				fr.forstack.st.didfirst = 1
			}
			if result {
				if val := arr.GetItem(&fr.forstack.st.cur); val != nil {
					CHECKOFLOW(2)
					push(arg, top, &fr.forstack.st.cur)	/* push key onto stack */
					push(arg, top, val)	/* push value onto stack */
					tmp = 1		/* tell following IF to NOT branch out of loop */
				} else {
					tmp = 0		/* tell following IF to branch out of loop */
				}
			} else {
				fr.forstack.st.cur.line = 0
				fr.forstack.st.cur.data = 0
				tmp = 0			/* tell following IF to branch out of loop */
			}
		} else {
			cur := fr.forstack.st.cur.data.(int)
			end := fr.forstack.st.end.data.(int)

			if tmp = fr.forstack.st.step > 0; tmp {
				tmp = !(cur > end)
			} else {
				tmp = !(cur < end)
			}
			if tmp {
				CHECKOFLOW(1)
				result = cur
				fr.forstack.st.cur.data.(int) += fr.forstack.st.step
				push(arg, top, result)
			}
		}
		CHECKOFLOW(1)
		push(arg, top, tmp)
	})
}

func prim_forpop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		if fr.forstack.top == 0 {
			panic("Internal error; FOR stack underflow.")
		}
		if fr.trys.st != nil {
			fr.trys.st.for_count--
		}
		fr.forstack.top--
		fr.forstack.st = pop_for(fr.forstack.st)
	})
}

func prim_trypop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		if !fr.trys.top {
			panic("Internal error; TRY stack underflow.")
		}
		fr.trys.top--
		fr.trys.st = pop_try(fr.trys.st)
	})
}

func prim_reverse(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		depth := op[0].(int)
		if depth < 0 {
			panic("Argument must be positive.")
		}
		checkop(depth, top)
		if depth > 0 {
			for i := 0; i < depth / 2; i++ {
				temp2 = arg[*top - (depth - i)]
				arg[*top - (depth - i)] = arg[*top - (i + 1)]
				arg[*top - (i + 1)] = temp2
			}
		}
	})
}

func prim_lreverse(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		depth := op[0].(int)
		if depth < 0 {
			panic("Argument must be positive.")
		}
		checkop(depth, top)
		if depth > 0 {
			for i := 0; i < depth / 2; i++ {
				temp2 = arg[*top - (depth - i)]
				arg[*top - (depth - i)] = arg[*top - (i + 1)]
				arg[*top - (i + 1)] = temp2
			}
		}
		push(arg, top, tmp)
	})
}