package fbmuck

func prim_array_make(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		items := op[0].(int)
		if items < 0 {
			panic("Item count must be non-negative.")
		}
		checkop(items, top)
		nu := make(Array, items)
		for ; items > 0; items-- {
			nu = append(nu, POP())
		}
		push(arg, top, nu)
	})
}

func prim_array_make_dict(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		items := op[0].(int)
		if items < 0 {
			panic("Item count must be positive.")
		}
		checkop(2 * items, top)
		nu := make(Dictionary)
		for ; items > 0; items-- {
			v := POP()
			k := POP()
			switch k.(type) {
			case int:
				nu[fmt.Sprint(k)] = v
			case string:
				nu[k] = v
			default:
				panic("Keys must be integers or strings.")
			}
		}
		push(arg, top, nu)
	})
}

func prim_array_explode(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		for arr := op[0].(type) {
		case Array:
			CHECKOFLOW((2 * len(arr)) + 1)
			for i, v := range arr {
				push(arg, top, i)
				push(arg, top, v)
			}
			push(arg, top, len(arr))
		case Dictionary:
			CHECKOFLOW((2 * len(arr)) + 1)
			for k, v := range arr {
				push(arg, top, k)
				push(arg, top, v)
			}
			push(arg, top, len(arr))
		default:
			panic(op[0])
		}
	})
}

func prim_array_vals(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		switch arr := op[0].(type) {
		case Array:
			CHECKOFLOW(len(arr) + 1)
			for _, v := range arr {
				push(arg, top, v)
			}
			push(arg, top, len(arr))
		case Dictionary:
			CHECKOFLOW(len(arr) + 1)
			for _, v := range arr {
				push(arg, top, v)
			}
			push(arg, top, len(arr))
		default:
			panic(op[0])
		}
	})
}

func prim_array_keys(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		arr := op[0].(stk_array)
		result = arr.Len()
		CHECKOFLOW((result + 1))
		for temp1 = array_first(arr); temp1 != nil; temp1 = array_next(arr, temp1) {
			push(arg, top, temp1)
		}
		push(arg, top, result)
	})
}

func prim_array_count(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		result = POP().data.(stk_array).Len()
		push(arg, top, result)
	})
}

func prim_array_first(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if temp1 = array_first(op[0].(stk_array)); temp1 != nil {
			push(arg, top, temp1)
			push(arg, top, 1)
		} else {
			push(arg, top, 0)
			push(arg, top, 0)
		}
	})
}

func prim_array_last(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if temp1 := array_last(op[0].(stk_array)); temp1 != nil {
			push(arg, top, temp1)
			push(arg, top, 1)
		} else {
			push(arg, top, 0)
			push(arg, top, 0)
		}
	})
}

func prim_array_prev(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch oper1.(type) {
		case int, string:
		default:
			panic("Argument not an integer or string. (2)")
		}
		temp1 := array_prev(op[0].(stk_array), op[1].(array_iter))
		if temp1 != nil;  {
			push(arg, top, temp1)
			push(arg, top, 1)
		} else {
			push(arg, top, 0)
		}
	})
}

func prim_array_next(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch oper1.(type) {
		case int, string:
		default:
			panic("Argument not an integer or string. (2)")
		}
		temp1 := op[1]
		result = array_next(op[0].(stk_array), &temp1)
		if result {
			push(arg, top, temp1)
			push(arg, top, 1)
		} else {
			push(arg, top, 0)
		}
	})
}

func prim_array_getitem(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch arr := op[0].(type) {
		case Array:
			push(arg, top, arr[op[1].(int)])
		case Dictionary:
			push(arg, top, arr[fmt.Sprint(op[1])])
		default:
			panic(op[0])
		}
	})
}

func prim_array_setitem(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		switch op[2].(type) {
		case int, string:
		default:
			panic("Argument not an integer or string. (3)")
		}

		if result = array_setitem(&op[1].(stk_array), op[2], op[0]); result < 0 {
			panic("Index out of array bounds. (3)")
		}
		push(arg, top, op[1])
	})
}

func prim_array_appenditem(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, append(op[1].(Array), op[0]))
	})
}

func prim_array_insertitem(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		switch op[2].(type) {
		case int, string:
		default:
			panic("Argument not an integer or string. (3)")
		}
		arr := op[1].(stk_array)
		if result = array_insertitem(&arr, op[2], op[0]); result < 0 {
			panic("Index out of array bounds. (3)")
		}
		push(arg, top, arr)
	})
}

func prim_array_getrange(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		switch op[2].(type) {
		case int, string:
		default:
			panic("Argument not an integer or string. (3)")
		}
		switch op[1].(type) {
		case int, string:
		default:
			panic("Argument not an integer or string. (2)")
		}

		nu := array_getrange(op[0].(stk_array), op[1], op[2])
		if nu == nil {
			nu = make(stk_array)
		}
		push(arg, top, nu)
	})
}

func prim_array_setrange(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		switch op[1].(type) {
		case int, string:
		default:
			panic("Argument not an integer or string. (2)")
		}
		if result = array_setrange(op[0].(stk_array), op[1], op[2].(stk_array)); result < 0 {
			panic("Index out of array bounds. (2)")
		}
		push(arg, top, op[0].(stk_array))
	})
}

func prim_array_insertrange(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		switch op[1].(type) {
		case int, string:
		default:
			panic("Argument not an integer or string. (2)")
		}
		if result = array_insertrange(&op[0].(stk_array), op[1], op[2].(stk_array)); result < 0 {
			panic("Index out of array bounds. (2)")
		}
		push(arg, top, op[0].(stk_array))
	})
}

func prim_array_delitem(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch op[0].(type) {
		case int, string:
		default:
			panic("Argument not an integer or string. (2)")
		}
		if result = array_delitem(&op[1].(stk_array), op[0]); result < 0 {
			abort_interp("Bad array index specified.")
		}
		push(arg, top, op[1].(stk_array))
	})
}

func prim_array_delrange(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		switch op[2].(type) {
		case int, string:
		default:
			panic("Argument not an integer or string. (3)")
		}
		switch op[1].(type) {
		case int, string:
		default:
			panic("Argument not an integer or string. (2)")
		}
		if result = array_delrange(&op[0].(stk_array), op[1], op[2]); result < 0 {
			panic("Bad array range specified.")
		}
		nu := op[0].(stk_array)
		push(arg, top, nu)
	})
}

func prim_array_cut(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		var nu1, nu2 stk_array
		arr := op[0].(stk_array)
		if temps := array_first(arr); temps != nil {
			switch tempc := op[1].(type) {
			case int:
				if tempc = array_prev(arr, tempc); tempc != nil {
					nu1 = array_getrange(arr, &temps, &tempc)
				}

				if tempe := array_last(arr); tempe != nil {
					nu2 = array_getrange(arr, oper2, &tempe)
				}
			case string:
				if tempc = array_prev(arr, tempc); tempc != nil {
					nu1 = array_getrange(arr, &temps, &tempc)
				}

				if tempe := array_last(arr); tempe != nil {
					nu2 = array_getrange(arr, oper2, &tempe)
				}
			}
		}

		if nu1 == nil {
			nu1 = make(stk_array)
		}
		if nu2 == nil {
			nu2 = make(stk_array)
		}
		push(arg, top, nu1)
		push(arg, top, nu2)
	})
}

func prim_array_n_union(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		items := op[0].(int)
		if items < 0 {
			panic("Item count must be positive.")
		}
		checkop(items, top)
		var new_union Array
		if items > 0 {
			new_mash := make(Dictionary)
			for ; items > 0; items-- {
				array_mash(POP().(interface{}), &new_mash, 1)
			}
			new_union = array_demote_only(new_mash, 1)
		}
		push(arg, top, new_union)
	})
}

func prim_array_n_intersection(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		items := op[0].(int)
		if items < 0 {
			panic("Item count must be positive.")
		}
		checkop(items, top)
		var nu Array
		if items > 0 {
			new_mash := make(Dictionary)
			for ; items > 0; items-- {
				temp_mash := make(Dictionary)
				array_mash(POP().data.(Array), &temp_mash, 1)
				nu = array_demote_only(temp_mash, 1)
				push(arg, top, nu)
				array_mash(POP().data.(stk_array), &new_mash, 1)
			}
			nu = array_demote_only(new_mash, result)
		}
		push(arg, top, new_union)
	})
}

func prim_array_n_difference(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		items := op[0].(int)
		if items < 0 {
			panic("Item count must be positive.")
		}
		checkop(items, top)
		var new_union Array
		if items > 0 {
			new_mash := make(Dictionary)
			array_mash(POP().(Array), &new_mash, 1)
			for ; items > 0; items-- {
				array_mash(POP().(Array), &new_mash, -1)
			}
			new_union = array_demote_only(new_mash, 1)
		}
		push(arg, top, new_union)
	})
}

func prim_array_notify(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		strarr := op[0].(Array)
		if !array_is_homogenous(strarr, "") {
			panic("Argument not an array of strings. (1)")
		}
		refarr := op[1].(Array)
		if !array_is_homogenous(refarr, dbref(0)) {
			panic("Argument not an array of dbrefs. (2)")
		}
		for i, k := range strarr {
			data := k.(string)
			if tp_force_mlev1_name_notify && mlev < JOURNEYMAN {
				data = prefix_message(data, db.Fetch(player).name)
			}
			for _, v := range refarr {
				obj := v.(dbref)
				notify_listeners(player, program, obj, db.Fetch(obj).location, data, 1)
			}
		}
	})
}

func prim_array_reverse(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		arr := op[0].(Array)
		nu := copy(make(Array, len(arr)), arr)
		nu.Reverse()
		push(arg, top, nu)
	})
}

func sortcomp_shuffle(x, y interface{}) int {
	return (((RANDOM() >> 8) % 5) - 2)
}

/* Sort types:
 * 1: case, ascending
 * 2: nocase, ascending
 * 3: case, descending
 * 4: nocase, descending
 * 5: randomize
 */
func prim_array_sort(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		arr := op[0].(Array)
		sort_type := op[1].(int)

		temp1.data = 0
		count := len(arr)
		nu := make(Array, count)
		tmparr := make(Array, count)
		copy(tmparr, nu)

		var comparator func(interface{}, interface{}) int
		if sort_type & SORTTYPE_SHUFFLE != 0 {
			comparator = sortcomp_shuffle
		} else {
			var idx *inst
			caseinsens := sort_type & SORTTYPE_CASEINSENS != 0
			descending := sort_type & SORTTYPE_DESCENDING != 0
			var a, b inst
			comparator = func(x, y interface{}) int {
				if descending {
					a = y.(inst)
					b = x.(inst)
				} else {
					a = x.(inst)
					b = y.(inst)
				}
				if idx != nil {
					/* This should only be set if comparators are all arrays. */
					a = a.data.(Array)[idx]
					b = b.data.(Array)[idx]
					switch {
					case a == nil && b == nil:
						return 0
					case a == nil:
						return -1
					case b == nil {
						return 1
					}
				}
			}
		}

		qsort(tmparr, count, sizeof(struct inst*), comparator);

		for i := 0; i < count; i++ {
			temp1.data = i
			array_setitem(&nu, &temp1, tmparr[i])
		}
		push(arg, top, nu)
	})
}


/* Sort types:
 * 1: case, ascending
 * 2: nocase, ascending
 * 3: case, descending
 * 4: nocase, descending
 * 5: randomize
 */
func prim_array_sort_indexed(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		arr := op[0].(Array)
		if !array_is_homogenous(arr, Array(nil)) {
			panic("Argument must be a list array of arrays. (1)")
		}

		sort_type := op[1].(int)
		switch op[2].(type) {
		case int, float64:
		default:
			panic("Index argument not an integer or string. (3)")
		}

		temp1.data = 0
		count := len(arr)
		nu := make(stk_array, count)

		tmparr := make(Array, len(arr))
		copy(tmparr, arr)

		var comparator func(interface{}, interface{}) int
		if sort_type & SORTTYPE_SHUFFLE != 0 {
			comparator = sortcomp_shuffle
		} else {
			idx := op[2]
			caseinsens := sort_type & SORTTYPE_CASEINSENS != 0
			descending := sort_type & SORTTYPE_DESCENDING != 0
			var a, b inst
			comparator = func(x, y interface{}) int {
				if descending {
					a = y.(inst)
					b = x.(inst)
				} else {
					a = x.(inst)
					b = y.(inst)
				}
				if idx != nil {
					/* This should only be set if comparators are all arrays. */
					a = a.data.(Array)[idx]
					b = b.data.(Array)[idx]
					switch {
					case a == nil && b == nil:
						return 0
					case a == nil:
						return -1
					case b == nil {
						return 1
					}
				}
				return (array_idxcmp_case(a, b, !caseinsens))
			}
		}

		qsort(tmparr, count, sizeof(struct inst*), comparator)
		/* WORK: if we go multithreaded, the mutex should be released here. */
		/*       Share this mutex with ARRAY_SORT. */

		for i, v := range tmparr {
			array_setitem(&nu, &array_iter{ data: i }, v)
		}
		push(arg, top, nu)
	})
}

func prim_array_get_propdirs(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 2, top, func(op Array) {
		ref := valid_object(op[0])
		dir := op[1].(string)
		if dir == "" {
			dir = "/"
		}
		if l := len(dir) - 1; l > 0 && dir[l] == PROPDIR_DELIMITER {
			dir = dir[:l]
		}
		nu := make(Array)
		var pptr *Plist
		var propname string
		for propadr := pptr.first_prop(ref, dir, propname); propadr != nil; propadr, propname = propadr.next_prop(pptr) {
			if prop_read_perms(ProgUID, ref, fmt.Sprint(dir, PROPDIR_DELIMITER, propname), mlev) {
				if p := get_property(ref, buf); p.IsPropDir() {
					nu = append(nu, propname)
				}
			}
		}
		push(arg, top, nu)
	})
}

func prim_array_get_propvals(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 2, top, func(op Array) {
		ref := valid_object(op[0])
		dir := op[1].(string)
		if dir == "" {
			dir = "/"
		}
		nu := make(Dictionary)
		var pptr *Plist
		var propname string
		for propadr := pptr.first_prop(ref, dir, propname); propadr != nil; propadr, propname = propadr.next_prop(pptr) {
			buf := fmt.Sprint(dir, PROPDIR_DELIMITER, propname)
			if prop_read_perms(ProgUID, ref, buf, mlev) {
				if prptr := get_property(ref, buf); prptr != nil {
					switch v := prptr.data.(type) {
					case string:
						nu[propname] = v
					case *boolexp:			//	FIXME: lock
						if v != TRUE_BOOLEXP {
							nu[propname] = copy_bool(v)
						} else {
							nu[propname] = TRUE_BOOLEXP
						}
					case dbref:
						nu[propname] = v
					case int:
						nu[propname] = v
					case float64:
						nu[propname] = v
					}
				}
			}
		}
		push(arg, top, nu)
	})
}

func prim_array_get_proplist(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		ref := valid_object(op[0])
		dir := op[1].(string)
		if dir == "" {
			dir = "/"
		}
		propname := fmt.Sprintf("%s#", dir)
		maxcount := get_property_value(ref, propname)
		if maxcount == 0 {
			strval := get_property_class(ref, propname)
			if strval != "" && unicode.IsNumber(strval) {
				maxcount = strconv.Atoi(strval)
			}
			if maxcount == 0 {
				propname = fmt.Sprintf("%s%c#", dir, PROPDIR_DELIMITER)
				if maxcount = get_property_value(ref, propname); maxcount == 0 {
					strval = get_property_class(ref, propname)
					if strval != "" && unicode.IsNumber(strval) {
						maxcount = strconv.Atoi(strval)
					}
				}
			}
		}

		nu := make(Array)
		for count := 1; maxcount > 0; count++ {
			propname = fmt.Sprintf("%s#%c%d", dir, PROPDIR_DELIMITER, count)
			prptr := get_property(ref, propname)
			if prptr == nil {
				propname = fmt.Sprintf("%s%c%d", dir, PROPDIR_DELIMITER, count)
				if prptr = get_property(ref, propname); prptr == nil {
					propname = fmt.Sprintf("%s%d", dir, count)
					prptr = get_property(ref, propname)
				}
			}
			if maxcount > 1023 {
				maxcount = 1023
			}
			switch {
			case maxcount != 0 && count > maxcount, maxcount == 0 && prptr == nil:
				break
			}
			if prop_read_perms(ProgUID, ref, propname, mlev) {
				if prptr == nil {
					nu = append(nu, 0)
				} else {
					switch v := prptr.data.(type) {
					case string:
						nu = append(nu, v)
					case *boolexp:			//	FIXME: lock
						if v.lock != TRUE_BOOLEXP {
							nu = append(nu, copy_bool(v))
						} else {
							nu = append(nu, v)
						}
					case dbref:
						nu = append(nu, v)
					case int:
						nu = append(nu, v)
					case float64:
						nu = append(nu, v)
					default:
						nu = append(nu, nil)
					}
				}
			} else {
				break
			}
		}
		push(arg, top, nu)
	})
}

func prim_array_put_propvals(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		obj := valid_object(op[0])
		dir := op[1].(string)
		switch case arr := op[2].(type) {
		case Array:
			for i, v := range arr {
				propname := fmt.Sprint(dir, PROPDIR_DELIMITER, v)
				if !prop_write_perms(ProgUID, ref, propname, mlev) {
					panic("Permission denied while trying to set protected property.")
				}
				switch v.(type) {
				case string, int, float64, dbref:
					set_property(ref, propname, v)
				case *boolexp:		//	FIXME: lock
					set_property(ref, propname, copy_bool(v.(*boolexp)))
				}
			}
		case Dictionary:
			for k, v := range arr {
				propname := fmt.Sprint(dir, PROPDIR_DELIMITER, k)

				if !prop_write_perms(ProgUID, ref, propname, mlev) {
					panic("Permission denied while trying to set protected property.")
				}

				switch v.(type) {
				case string, int, float64, dbref:
					set_property(ref, propname, v)
				case *boolexp:		//	FIXME: lock
					set_property(ref, propname, copy_bool(v.(*boolexp)))
				}
			}
		case map[int] interface{}:
			for k, v := range arr {
				propname := fmt.Sprintf("%s%c%d", dir, PROPDIR_DELIMITER, k)

				if !prop_write_perms(ProgUID, ref, propname, mlev) {
					panic("Permission denied while trying to set protected property.")
				}

				switch v.(type) {
				case string, int, float64, dbref:
					set_property(ref, propname, v)
				case *boolexp:		//	FIXME: lock
					set_property(ref, propname, copy_bool(v.(*boolexp)))
				}
			}
		case map[float64] interface{}:
			for k, v := range arr {
				propname := fmt.Sprintf("%s%c%.15g", dir, PROPDIR_DELIMITER, k)
				if !strings.ContainsAny(propname, '.ne') {
					propname += ".0"
				}

				if !prop_write_perms(ProgUID, ref, propname, mlev) {
					panic("Permission denied while trying to set protected property.")
				}

				switch v.(type) {
				case string, int, float64, dbref:
					set_property(ref, propname, v)
				case *boolexp:		//	FIXME: lock
					set_property(ref, propname, copy_bool(v.(*boolexp)))
				}
			}
		default:
			panic(op[2])
		}
	})
}

func prim_array_put_proplist(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		ref := valid_object(op[0])
		dir := op[1].(string)
		arr := op[2].(Array)

		var propname string
		for fmtin := tp_proplist_counter_fmt; fmtin != ""; fmtin = fmtin[1:] {
			if fmtin[0] == 'P' {
				propname += dir
			} else {
				propname += fmtin[0]
			}
		}

		if !prop_write_perms(ProgUID, ref, propname, mlev)
			panic("Permission denied while trying to set protected property.")
		}

		var propdat interface{}
		if tp_proplist_int_counter {
			propdat = arr.Len()
		} else {
			propdat = fmt.Sprint(arr.Len())
		}
		set_property(ref, propname, &propdat)

		for i, v := range arr {
			for fmtin := tp_proplist_entry_fmt; fmtin != ""; fmtin = fmtin[1:] {
				switch fmtin[0] {
				case 'N':
					propname += fmt.Sprint(i + 1)
				case 'P':
					propname += dir
				default:
					propname += fmtin[0]
				}
			}

			if !prop_write_perms(ProgUID, ref, propname, mlev) {
				panic("Permission denied while trying to set protected property.")
			}

			switch v.(type) {
			case string, int, float64, dbref:
				propdat = v
			case *boolexp:		//	FIXME: lock
				propdat = copy_bool(v.(*boolexp))
			default:
				propdat = nil
			}
			set_property(ref, propname, propdat)
		}

		for count := len(arr); ; count++ {
			for fmtin := tp_proplist_entry_fmt[count:]; fmtin != ""; fmtin = fmtin[1:] {
				switch fmtin[0] {
				case 'N':
					propname += fmt.Sprint(count + 1)
				case 'P':
					propname += dir
				default:
					propname += fmtin[0]
				}
			}
			if get_property(ref, propname) {
				remove_property(ref, propname)
			} else {
				break
			}
		}
	})
}

func prim_array_get_reflist(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		ref := valid_object(op[0])
		dir := op[1].(string)
		if !prop_read_perms(ProgUID, ref, dir, mlev) {
			panic("Permission denied.")
		}
		nu := make(Array)
		if rawstr := get_property_class(ref, dir); rawstr != "" {
			rawstr = strings.TrimSpace(rawstr)
			for count := 0; rawstr != ""; count++ {
				if rawstr[0] == NUMBER_TOKEN {
					rawstr = rawstr[1:]
				}
				if !isdigit(rawstr[0]) && (rawstr != '-') {
					break
				}
				nu = append(nu, strconv.Atoi(rawstr))
				rawstr = strings.TrimLeftFunc(rawstr, func(r rune) bool {
					return !unicode.IsSpace(r)
				})
				rawstr = strings.TrimSpace(rawstr)
			}
		}
		push(arg, top, nu)
	})
}

func prim_array_put_reflist(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		ref := valid_object(op[0])
		dir := op[1].(string)
		arr := op[2].(Array)
		switch {
		case !array_is_homogenous(arr, dbref(0)):
			panic("Argument must be a list array of dbrefs. (3)")
		case !prop_write_perms(ProgUID, ref, dir, mlev):
			panic("Permission denied.")
		}
		var out string
		for _, v := range arr {
			if out != "" {
				out += ' '
			}
			out += fmt.Sprintf("#%d", v.(dbref))
		}
		remove_property(ref, dir)
		set_property(ref, dir, out)
	})
}

func prim_array_findval(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		switch arr := op[0].(type) {
		case Array:
			nu := make(Array)
			for i, v := range arr {
				if array_idxcmp(v, op[1]) == 0 {
					nu = append(nu, i)
				}
			}
			push(arg, top, nu)
		case map(Array) interface{}:
			nu := make(Array)
			for k, v := range arr {
				if array_idxcmp(v, op[1]) == 0 {
					nu = append(nu, k)
				}
			}
			push(arg, top, nu)
		default:
			panic(op[0])
		}
	})
}

func prim_array_compare(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		arr1 := op[0].(stk_array)
		arr2 := op[1].(stk_array)
		switch temp1, temp2 := array_first(arr1), array_first(arr2); {
		case temp1 == nil && temp2 == nil:
			push(arg, top, 0)
		case temp1 == nil:
			push(arg, top, -1)
		case temp2 == nil:
			push(arg, top, 1)
		default:
			var result int
			for ; temp1 != nil && temp2 != nil; temp1, temp2 = array_next(arr1, temp1), array_next(arr2, temp2) {
				if result = array_idxcmp(temp1, temp2); result != 0 {
					break
				}
				if result = array_idxcmp(arr1.GetItem(temp1), arr2.GetItem(temp2)); result != 0 {
					break
				}
			}
			switch {
			case temp1 == nil && temp2 == nil:
				result = 0
			case temp1 == nil:
				result = -1
			case temp2 == nil:
				result = 1
			}
			push(arg, top, result)
		}
	})
}

func prim_array_matchkey(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		arr := op[0].(Dictionary)
		pattern := op[1].(string)
		nu := make(Dictionary)
		for k, v := range arr {
			if k, ok := k; ok {
				if smatch(pattern, k) == 0 {
					nu[k] = v
				}
			}
		}
		push(arg, top, nu)
	})
}

func prim_array_matchval(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		pattern := op[1].(string)
		switch arr := op[0].(type) {
		case Array:
			nu := make(Dictionary)
			for i, v := range arr {
				if v, ok := v.(int); ok {
					if smatch(pattern, v) == 0 {
						nu[fmt.Sprint(i)] = v
					}
				}
			}
			push(arg, top, nu)
		case Dictionary:
			nu := make(Dictionary)
			for k, v := range arr {
				ks := fmt.Sprint(v)
				if !smatch(pattern, ks) {
					nu[k] = ks
				}
			}
			push(arg, top, nu)
		default:
			panic(op[0])
		}
	})
}

func prim_array_extract(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		idxarr := op[1].(Array)
		switch arr := op[0].(type) {
		case Array:
			nu := make(Dictionary)
			for _, v := range idxarr {
				nu[fmt.Sprint(v)] = arr[v.(int)]
			}
			push(arg, top, nu)
		case Dictionary:
			nu := make(Dictionary)
			for _, v := range idxarr {
				k := fmt.Sprint(v)
				nu[k] = arr[k]
			}
			push(arg, top, nu)
		default:
			panic(op[0])
		}
	})
}

func prim_array_excludeval(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		arr := op[0].(Array)
		nu := make(Array)
		for _, v := range arr {
			if array_idxcmp(v, op[1]) {
				nu = append(nu, temp1)
			}
		}
		push(arg, top, nu)
	})
}

void prim_array_join(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		arr := op[0].(Array)
		items := make([]string, len(arr)
		for i, v := range arr {
			switch v := v.(type) {
			case string:
				items[i] = v
			case int:
				items[i] = fmt.Sprint(v)
			case dbref:
				items[i] = fmt.Sprintf("#%d", v)
			case float64:
				text := fmt.Sprintf("%.15g", v)
				if !strings.ContainsAny(text, ".ne") {
					text += ".0"
				}
				items[i] = text
			case *boolexp:
				items[i] = unparse_boolexp(ProgUID, v, true)
			default:
				items[i] = "<UNSUPPORTED>"
			}
		}
		push(arg, top, strings.Join(items, op[1].(string)))
	})
}

func prim_array_interpret(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
    apply_primitive(1, top, func(op Array) {
		var text string
		for i, v := range op[0].(Array) {
			switch v := v.(type) {
			case string:
				text += v
			case int:
				text += fmt.Sprint(data)
			case dbref:
				switch {
				case v == NOTHING:
					text += "*NOTHING*"
				case v == AMBIGUOUS:
					text += "*AMBIGUOUS*"
				case v == HOME:
					text += "*HOME*"
				case v < HOME || v >= db_top:
					text + = "*INVALID*"
				default:
					text += fmt.Sprint(db.Fetch(v).name)
				}
			case float64:
				text += fmt.Sprintf("%.15g", v)
				if !strings.ContainsAny(r, '.ne') {
					text += ".0"
				}
			case *boolexp:
				text += unparse_boolexp(ProgUID, v, true)
			default:
				text += "<UNSUPPORTED>"
			}
		}
		push(arg, top, text)
    })
}

func prim_array_get_ignorelist(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		ref := db.Fetch(valid_object(op[0])).owner
		nu := make(Array)
		if tp_ignore_support {
			if rawstr := get_property_class(ref, IGNORE_PROP); rawstr != "" {
				rawstr = strings.TrimSpace(rawstr)
				for count := 0; rawstr != ""; count++ {
					if rawstr[0] == NUMBER_TOKEN {
						rawstr = rawstr[1:]
					}
					if !isdigit(rawstr[0]) {
						break
					}
					result = strconv.Atoi(rawstr)
					rawstr = strings.TrimLeftFunc(buf, func(r rune) bool {
						return !unicode.IsSpace(r)
					})
					rawstr = strings.TrimSpace(rawstr)
					nu = append(nu, result)
				}
			}
		}
		push(arg, top, nu)
	})
}

func prim_array_nested_get(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		arr := op[0].(stk_array)
		idxarr := op[1].(Array)
		idxcnt := idxarr.Len()
		for i := 0; dat && i < idxcnt; i++ {
			idx := idxarr[i]
			switch idx.data.(type) {
			case int, string:
			default:
				panic("Argument not an array of indexes. (2)")
			}
			arr = arr.GetItem(idx).data.(stk_array)
		}
		push(arg, top, inst{ data: arr })
	})
}

func prim_array_nested_set(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		arr := op[1].(stk_array)
		idxarr := op[2].(Array)
		nest := make(Array, len(idxarr))
		if len(nest) == 0 {
			push(arg, top, op[1])
		} else {
			var temp inst
			var idx *inst
			for i := 0; i < idxcnt; i++ {
				nest[i] = arr.Dup()
				idx = idxarr[i]
				switch idx.(type) {
				case int, string:
				default:
					panic("Argument not an array of indexes. (3)")
				}
				if i < idxcnt - 1 {
					if arr = nest[i].(stk_array).GetItem(idx); arr == nil {
						switch idx := idx.(type) {
						case int:
							if idx == 0 {
								temp = make(Array, 1)
							} else {
								temp = make(Dictionary)
							}
						default:
							temp = make(Dictionary)
						}
						arr = temp
					}
				}
			}

			array_setitem(&nest[idxcnt - 1].data.(stk_array), idx, op[0])
			for i := idxcnt - 1; i > 0; i-- {
				array_setitem(&nest[i], idxarr[i], &nest[i + 1])
			}
			push(arg, top, nest[0].Dup())
		}
	})
}

func prim_array_nested_del(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		arr := op[0].(stk_array)
		idxarr := op[1].(Array)
		nest := make(Array, len(idxarr))
		if len(nest) == 0 {
			push(arg, top, arr)
		} else {
			var doneearly bool
			for i, v := range idxarr {
				nest[i] = arr.Dup()
				idx := idxarr[i]
				switch idx.(type) {
				case int, string:
				default:
					panic("Argument not an array of indexes. (2)")
				}
				if i < len(idxarr) - 1 {
					if arr = nest[i].(stk_array).GetItem(idx); arr != nil {
						arr.(stk_array)
					} else {
						doneearly = true
						break
					}
				}
			}
			if !doneearly {
				array_delitem(&nest[len(idxarr - 1)].(stk_array), idx)
				for i := len(idxarr) - 1; i > 0; i-- {
					array_setitem(&nest[i], idxarr[i], &nest[i + 1])
				}
			}
			push(arg, top, nest[0])
		}
	})
}

func prim_array_filter_flags(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
    apply_primitive(2, top, func(op Array) {
		objs := op[0].(Array)
		if !array_is_homogenous(objs, dbref(0)) {
			panic("Argument not an array of dbrefs. (1)")
		}
		flags := op[1].(string)
		nw := make(Array)
		_, check := init_checkflags(player, flags)
		for _, in := range objs {
			if valid_object(in, false) && checkflags(in.(dbref), check) {
				nw = append(nw, in)
			}
		}
		push(arg, top, nw)
    })
}