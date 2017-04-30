package fbmuck

func prim_add(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch tf2 := op[0].(type) {
		case float64:
			switch tf1 := op[1].(type) {
			case float64:
				push(arg, top, tf1 + tf2)
			case int:
				push(arg, top, float64(tf1) + tf2)
			default:
				push(arg, top, math.NaN())
			}
		case int:
			switch tf1 := op[1].(type) {
			case float64:
				push(arg, top, tf1 + tf2)
			case int:
				push(arg, top, tf1 + tf2)
			default:
				push(arg, top, math.NaN())
			}
		default:
			push(arg, top, math.NaN())
		}
	})
}

func prim_subtract(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch tf2 := op[0].(type) {
		case float64:
			switch tf1 := op[1].(type) {
			case float64:
				push(arg, top, tf1 - tf2)
			case int:
				push(arg, top, float64(tf1) - tf2)
			default:
				push(arg, top, math.NaN())
			}
		case int:
			switch tf1 := op[1].(type) {
			case float64:
				push(arg, top, tf1 - tf2)
			case int:
				push(arg, top, tf1 - tf2)
			default:
				push(arg, top, math.NaN())
			}
		default:
			push(arg, top, math.NaN())
		}
	})
}

func prim_multiply(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch tf2 := op[0].(type) {
		case float64:
			switch tf1 := op[1].(type) {
			case float64:
				push(arg, top, tf1 * tf2)
			case int:
				push(arg, top, float64(tf1) * tf2)
			default:
				push(arg, top, math.NaN())
			}
		case int:
			switch tf1 := op[1].(type) {
			case float64:
				push(arg, top, tf1 * tf2)
			case int:
				push(arg, top, tf1 * tf2)
			default:
				push(arg, top, math.NaN())
			}
		default:
			push(arg, top, math.NaN())
		}
	})
}

func prim_divide(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) (r interface{}) {
		defer func() {
			if recover() {
				fr.error.error_flags.div_zero = true
				push(arg, top, math.NaN())
			}
		}()
		switch tf2 := op[0].(type) {
		case float64:
			switch tf1 := op[1].(type) {
			case float64:
				push(arg, top, tf1 / tf2)
			case int:
				push(arg, top, float64(tf1) / tf2)
			default:
				push(arg, top, math.NaN())
			}
		case int:
			switch tf1 := op[1].(type) {
			case float64:
				push(arg, top, tf1 / tf2)
			case int:
				push(arg, top, tf1 / tf2)
			default:
				push(arg, top, math.NaN())
			}
		default:
			push(arg, top, math.NaN())
		}
	})
}

func prim_mod(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, op[1].(int) % op[0].(int))
	})
}

func prim_bitor(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, op[1].(int) | op[0].(int))
	})
}

func prim_bitxor(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, op[1].(int) ^ op[0].(int))
	})
}

func prim_bitand(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, op[1].(int) & op[0].(int))
	})
}

func prim_bitshift(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch x := op[0].(int); {
		case x > 0:
			push(arg, top, op[1].(int) << x)
		case x < 0:
			push(arg, top, op[1].(int) >> -x)
		default:
			push(arg, top, op[1].(int))
		}
	})
}

func apply_logic_primitive(top *int, f func(Array) bool) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, MUFBool(f(op []inst)))
	})
}

func prim_and(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_logic_primitive(top, func(op Array) bool {
		return !ValueIsFalse(op[0]) && !ValueIsFalse(op[1])
	})
}

func prim_or(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_logic_primitive(2, top, func(op Array) bool {
		return !ValueIsFalse(op[0]) || !ValueIsFalse(op[1])
	})
}

func prim_xor(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_logic_primitive(top, func(op Array) (ok bool) {
		if ValueIsFalse(op[0]) {
			ok = !ValueIsFalse(op[1])
		} else {
			ok = ValueIsFalse(op[1])
		}
		return
	})
}

func prim_not(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, MUFBool(ValueIsFalse(op[0])))
	})
}

func comp_t(op *inst) (r bool) {
	switch op.data.(type) {
	case int, float64, ObjectID:
		r = true
	}
	return
}

func apply_comparison_primitive(top *int, f func(op Array) bool) {
	apply_primitive(2, top, func(op Array) {
		if comp_t(op[0]) && comp_t(op[1]) {
			push(arg, top, MUFBool(f(op)))
		} else {
			panic("Invalid argument type.")
		}
	}
}

func prim_lessthan(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_comparison_primitive(top, func(op Array) (ok bool) {
		switch l := op[0].(type) {
		case float64:
			switch r := op[1].(type) {
			case float64:
				ok = r < l
			case int:
				ok = r < l
			case ObjectID:
				ok = r < l
			}
		case int:
			switch r := op[1].(type) {
			case float64:
				ok = r < l
			case int:
				ok = r < l
			case ObjectID:
				ok = r < l
			}
		}
		return
	})
}

func prim_greathan(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_comparison_primitive(top, func(op Array) (ok bool) {
		switch l := op[0].(type) {
		case float64:
			switch r := op[1].(type) {
			case float64:
				ok = r > l
			case int:
				ok = r > l
			case objref:
				ok = r > l
			}
		case int:
			switch r := op[1].(type) {
			case float64:
				ok = r > l
			case int:
				ok = r > l
			case objref:
				ok = r > l
			}
		}
		return
	})
}

func prim_equal(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_comparison_primitive(top, func(op Array) (ok bool) {
		switch l := op[0].(type) {
		case float64:
			switch r := op[1].(type) {
			case float64:
				ok = r == l
			case int:
				ok = r == l
			case objref:
				ok = r == l
			default:
				panic(r)
			}
		case int:
			switch r := op[1].(type) {
			case float64:
				ok = r == l
			case int:
				ok = r == l
			case objref:
				ok = r == l
			default:
				panic(r)
			}
		}
		return
	})
}

func prim_lesseq(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_comparison_primitive(top, func(op Array) (ok bool) {
		switch l := op[0].(type) {
		case float64:
			switch r := op[1].(type) {
			case float64:
				ok = r <= l
			case int:
				ok = r <= l
			case objref:
				ok = r <= l
			default:
				panic(r)
			}
		case int:
			switch r := op[1].(type) {
			case float64:
				ok = r <= l
			case int:
				ok = r <= l
			case objref:
				ok = r <= l
			default:
				panic(r)
			}
		}
		return
	})
}

func prim_greateq(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_comparison_primitive(top, func(op Array) (ok bool) {
		switch l := op[0].(type) {
		case float64:
			switch r := op[1].(type) {
			case float64:
				ok = r >= l
			case int:
				ok = r >= l
			case objref:
				ok = r >= l
			default:
				panic(r)
			}
		case int:
			switch r := op[1].(type) {
			case float64:
				ok = r >= l
			case int:
				ok = r >= l
			case ObjectID:
				ok = r >= l
			default:
				panic(r)
			}
		default:
			panic(l)
		}
		return
	})
}

func prim_random(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	CHECKOFLOW(1)
	push(arg, top, rand.Int())
}

func prim_srand(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		if fr.seed == nil {
			fr.SetSeed(nil)
		}
		push(arg, top, fr.Rand.Int())
	})
}

func prim_getseed(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		a := make([]byte, 0, 32)
		for i, v := range fr.GetSeed() {
			a = append(a, (v & 0x0F) + 65)
			a = append(((v & 0xF0) >> 4) + 65)
		}
		push(arg, top, a)
	})
}

func prim_setseed(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		fr.SetSeed(op[0])
	})
}

func prim_int(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		switch data := op[0].(type) {
		case ObjectID:
			push(arg, top, int(data))
		case PROG_VAR:
			push(arg, top, int(data))
		case PROG_LVAR:
			push(arg, top, int(data))
		case float64:
			push(arg, top, int(data))
		default:
			panic(data)
		}
	})
}

func prim_plusplus(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		temp1 := *op[0]
		var temp2 inst
		switch idx := op[0].(type) {
		case PROG_VAR:
			temp2 = fr.variables[idx]
		case PROG_SVAR:
			tmp := scopedvar_get(fr, 0, idx)
			tmp = temp2.Dup()
		case PROG_LVAR:
			temp2 = localvars_get(fr, program).lvars[idx].Dup()
		case int:
			idx++
			push(arg, top, idx)
			return
		case ObjectID:
			idx++
			push(arg, top, idx)
			return
		case float64:
			idx++
			push(arg, top, idx)
			return
		default:
			panic("Invalid datatype.")
		}

		switch data := temp2.data.(type) {
		case int:
			data++
		case ObjectID:
			data++
		case float64:
			data++
		default:
			panic("Invalid datatype in variable.")
		}

		switch vnum := temp1.data(type) {
		case PROG_VAR:
			fr.variables[vnum] = temp2.Dup()
		case PROG_SVAR:
			if tmp2 := scopedvar_get(fr, 0, vnum); tmp2 == nil {
				panic("Scoped variable number out of range.")
			} else {
				tmp2 = temp2.Dup()
			}
		case PROG_LVAR:
			localvars_get(fr, program).lvars[vnum] = temp2.Dup()
		}
	})
}

func prim_minusminus(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		var tmp *inst
		var temp2 *inst
		var vnum int

		temp1 := *op[0]
		switch data := op[0].(type) {
		case PROG_VAR:
			temp2 = fr.variables[data].Dup()
		case PROG_SVAR:
			tmp = scopedvar_get(fr, 0, temp1.data.(int))
			temp2 = tmp.Dup()
		case PROG_LVAR:
			tmp2 := localvars_get(fr, program)
			temp2 = tmp2.lvars[data].Dup()
		case int:
			data--
			push(arg, top, data)
			return
		case ObjectID:
			data--
			push(arg, top, data)
			return
		case float64:
			data--
			push(arg, top, data)
			return
		default:
			panic("Invalid datatype.")
		}

		switch data := temp2.data.(type) {
		case int:
			data--
		case ObjectID:
			data--
		case float64:
			data--
		default:
			panic("Invalid datatype in variable.")
		}

		switch vnum := temp1.data.(type) {
		case PROG_VAR:
			fr.variables[vnum] = temp2.Dup()
		case PROG_SVAR:
			if tmp2 := scopedvar_get(fr, 0, vnum); !tmp2 {
				panic("Scoped variable number out of range.")
			} else {
				tmp2 = temp2.Dup()
			}
		case PROG_LVAR:
			localvars_get(fr, program).lvars[vnum] = temp2.Dup()
		}
	})
}

func prim_abs(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if result = op[0].(int); result < 0 {
			result = -result
		}
		push(arg, top, result)
	})
}

func prim_sign(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		switch v := op[0].(int) {
		case v > 0:
			push(arg, top, 1)
		case v < 0:
			push(arg, top, -1)
		default:
			push(arg, top, 0)
		}
	})
}