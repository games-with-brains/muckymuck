package fbmuck

func prim_add(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch tf2 := op[0].data.(type) {
		case float64:
			switch tf1 := op[1].data.(type) {
			case float64:
				push(arg, top, tf1 + tf2)
			case int:
				push(arg, top, float64(tf1) + tf2)
			default:
				push(arg, top, math.NaN())
			}
		case int:
			switch tf1 := op[1].data.(type) {
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

func prim_subtract(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch tf2 := op[0].data.(type) {
		case float64:
			switch tf1 := op[1].data.(type) {
			case float64:
				push(arg, top, tf1 - tf2)
			case int:
				push(arg, top, float64(tf1) - tf2)
			default:
				push(arg, top, math.NaN())
			}
		case int:
			switch tf1 := op[1].data.(type) {
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

func prim_multiply(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch tf2 := op[0].data.(type) {
		case float64:
			switch tf1 := op[1].data.(type) {
			case float64:
				push(arg, top, tf1 * tf2)
			case int:
				push(arg, top, float64(tf1) * tf2)
			default:
				push(arg, top, math.NaN())
			}
		case int:
			switch tf1 := op[1].data.(type) {
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

func prim_divide(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) (r interface{}) {
		defer func() {
			if recover() {
				fr.error.error_flags.div_zero = true
				push(arg, top, math.NaN())
			}
		}()
		switch tf2 := op[0].data.(type) {
		case float64:
			switch tf1 := op[1].data.(type) {
			case float64:
				push(arg, top, tf1 / tf2)
			case int:
				push(arg, top, float64(tf1) / tf2)
			default:
				push(arg, top, math.NaN())
			}
		case int:
			switch tf1 := op[1].data.(type) {
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

func prim_mod(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, op[1].data.(int) % op[0].data.(int))
	})
}

func prim_bitor(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, op[1].data.(int) | op[0].data.(int))
	})
}

func prim_bitxor(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, op[1].data.(int) ^ op[0].data.(int))
	})
}

func prim_bitand(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, op[1].data.(int) & op[0].data.(int))
	})
}

func prim_bitshift(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch x := op[0].dat.(int); {
		case x > 0:
			push(arg, top, op[1].data.(int) << x)
		case x < 0:
			push(arg, top, op[1].data.(int) >> -x)
		default:
			push(arg, top, op[1].data.(int))
		}
	})
}

func apply_logic_primitive(top *int, f func(Array) bool) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, MUFBool(f(op []inst)))
	})
}

func prim_and(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_logic_primitive(top, func(op Array) bool {
		return !ValueIsFalse(op[0]) && !ValueIsFalse(op[1])
	})
}

func prim_or(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_logic_primitive(2, top, func(op Array) bool {
		return !ValueIsFalse(op[0]) || !ValueIsFalse(op[1])
	})
}

func prim_xor(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_logic_primitive(top, func(op Array) (ok bool) {
		if ValueIsFalse(op[0]) {
			ok = !ValueIsFalse(op[1])
		} else {
			ok = ValueIsFalse(op[1])
		}
		return
	})
}

func prim_not(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, MUFBool(ValueIsFalse(op[0])))
	})
}

func comp_t(op *inst) (r bool) {
	switch op.data.(type) {
	case int, float64, dbref:
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

func prim_lessthan(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_comparison_primitive(top, func(op Array) (ok bool) {
		switch l := op[0].data.(type) {
		case float64:
			switch r := op[1].data.(type) {
			case float64:
				ok = r < l
			case int:
				ok = r < l
			case dbref:
				ok = r < l
			}
		case int:
			switch r := op[1].data.(type) {
			case float64:
				ok = r < l
			case int:
				ok = r < l
			case dbref:
				ok = r < l
			}
		}
		return
	})
}

func prim_greathan(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_comparison_primitive(top, func(op Array) (ok bool) {
		switch l := op[0].data.(type) {
		case float64:
			switch r := op[1].data.(type) {
			case float64:
				ok = r > l
			case int:
				ok = r > l
			case objref:
				ok = r > l
			}
		case int:
			switch r := op[1].data.(type) {
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

func prim_equal(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_comparison_primitive(top, func(op Array) (ok bool) {
		switch l := op[0].data.(type) {
		case float64:
			switch r := op[1].data.(type) {
			case float64:
				ok = r == l
			case int:
				ok = r == l
			case objref:
				ok = r == l
			}
		case int:
			switch r := op[1].data.(type) {
			case float64:
				ok = r == l
			case int:
				ok = r == l
			case objref:
				ok = r == l
			}
		}
		return
	})
}

func prim_lesseq(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_comparison_primitive(top, func(op Array) (ok bool) {
		switch l := op[0].data.(type) {
		case float64:
			switch r := op[1].data.(type) {
			case float64:
				ok = r <= l
			case int:
				ok = r <= l
			case objref:
				ok = r <= l
			}
		case int:
			switch r := op[1].data.(type) {
			case float64:
				ok = r <= l
			case int:
				ok = r <= l
			case objref:
				ok = r <= l
			}
		}
		return
	})
}

func prim_greateq(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_comparison_primitive(top, func(op Array) (ok bool) {
		switch l := op[0].data.(type) {
		case float64:
			switch r := op[1].data.(type) {
			case float64:
				ok = r >= l
			case int:
				ok = r >= l
			case objref:
				ok = r >= l
			}
		case int:
			switch r := op[1].data.(type) {
			case float64:
				ok = r >= l
			case int:
				ok = r >= l
			case dbref:
				ok = r >= l
			}
		}
		return
	})
}

func prim_random(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	CHECKOFLOW(1)
	push(arg, top, RANDOM())
}

func prim_srand(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		if fr.rndbuf == nil {
			fr.rndbuf = init_seed(nil)
		}
		push(arg, top, int(rnd(fr.rndbuf)))
	})
}

func prim_getseed(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		if fr.rndbuf == nil {
			push(arg, top, "")
		} else {
			buf2 := make([]byte, 16)
			copy(buf2, fr.rndbuf)
			buf := make([]byte, 32)
			for loop := 0; loop < 16; loop++ {
				buf[loop * 2] = (buf2[loop] & 0x0F) + 65
				buf[(loop * 2) + 1] = ((buf2[loop] & 0xF0) >> 4) + 65
			}
			push(arg, top, buf)
		}
	})
}

func prim_setseed(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if data, ok := op[0].(string); ok {
			if fr.rndbuf != nil {
				delete_seed(fr.rndbuf)
				fr.rndbuf = nil
			}
			if len(data) == 0 {
				fr.rndbuf = init_seed(nil)
			} else {
				holdbuf := make([]byte, 32)
				if slen := len(data); slen < 32 {
					for sloop := 0; sloop < 32; sloop++ {
						holdbuf[sloop] = data[sloop % slen]
					}
				} else {
					copy(holdbuf, data)
				}

				buf := make([]byte, 16)
				for sloop := 0; sloop < 16; sloop++ {
					buf[sloop] = ((holdbuf[sloop * 2] - 65) & 0x0F) | (((holdbuf[(sloop * 2) + 1] - 65) & 0x0F) << 4)
				}
				fr.rndbuf = init_seed(buf)
			}
		} else {
			panic("Invalid argument type.")
		}
	})
}

func prim_int(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		switch data := op[0].(type) {
		case dbref:
			push(arg, top, int(data))
		case PROG_VAR:
			push(arg, top, int(data))
		case PROG_LVAR:
			push(arg, top, int(data))
		case float64:
			push(arg, top, int(data))
		default:
			panic("Invalid argument type.")
		}
	})
}

func prim_plusplus(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
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
		case dbref:
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
		case dbref:
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

func prim_minusminus(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
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
		case dbref:
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
		case dbref:
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

func prim_abs(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if result = op[0].(int); result < 0 {
			result = -result
		}
		push(arg, top, result)
	})
}

func prim_sign(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
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