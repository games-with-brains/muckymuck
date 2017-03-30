package fbmuck

type Array []interface{}
type SparseArray map[int] interface{}
type Dictionary map[string] interface{}

type stk_array Array

func array_idxcmp(a, b *array_iter) (r int) {
	if a != nil && b != nil {
		r = array_tree_compare(a, b, false)
	}
	return
}

func array_idxcmp_case(a, b *array_iter, case_sens bool) (r int) {
	if a != nil && b != nil {
		r = array_tree_compare(a, b, case_sens)
	}
	return
}

func array_contains_key(arr interface{}, item *array_iter) (r bool) {
	if arr != nil {
		switch arr := arr.(type) {
		case Array:
			if v, ok := item.(int); ok {
				r = v >= 0 && v < len(arr)
			}
		case Dictionary:
			r = array_tree_find(arr.data.dict, item)
		}
	}
	return
}

func array_contains_value(arr interface{}, item Array) (r bool) {
	if arr != nil {
		switch arr := arr.(type) {
		case Array:
			for i := arr.items; i > 0; i-- {
				if array_tree_compare(&arr.data.packed[i], item, false) == 0 {
					r = true
					break
				}
			}
		case Dictionary:
			if arr.items > 0 {
				for p := array_tree_first_node(arr.data.dict); p != nil; p = array_tree_next_node(arr.data.dict, &p.data) {
					if array_tree_compare(&p.data, item, false) == 0 {
						r = true
						break
					}
				}
			}
		}
	}
	return
}

func array_first(arr *stk_array) (r *array_iter) {
	if item != nil && arr != nil && len(arr) > 0 {
		switch (arr->type) {
		case Array:
			r = &array_iter{ data: 0 }
		case Dictionary:
			if p := array_tree_first_node(arr.data.(dict)); p != nil {
				r = &array_iter{ data: p.key }
			}
		}
	}
	return
}

func array_last(arr stk_array) (r *array_iter) {
	if item != nil {
		if arr != nil && len(arr) > 0 {
			switch (arr->type) {
			case Array:
				r = &array_iter{ data: len(arr) - 1 }
			case Dictionary:
				if p := array_tree_last_node(arr->data.dict); p != nil {
					r = &array_iter{ data: p.key }
				}
			}
		}
	}
	return
}

func array_prev(arr *stk_array, item *array_iter) (r *array_iter) {
	if item != nil && arr != nil && len(arr) != 0 {
		switch (arr->type) {
		case Array:
			i := -1
			switch v := item.data.(type) {
			case float64:
				if v > len(arr) {
					i = len(items) - 1
				} else {
					i = int(v - 1.0)
				}
			case int:
				i = v - 1
			}
			if i >= len(arr) {
				i = len(arr) - 1
			}
			if i > -1 {
				r = &array_iter{ data: i }
			}
		case Dictionary:
			if p := array_tree_prev_node(arr.data.dict, item); p != nil {
				r = &array_iter{ data: p.key }
			}
		}
	}
	return
}

func array_next(stk_array * arr, array_iter * item) (r int) {
	if item != nil && arr != nil && len(arr) > 0 {
		switch (arr->type) {
		case Array:
			i := 0
			switch v := item.data.(type) {
			case float64:
				if v < 0.0 {
					i = 0
				} else {
					i = int(v + 1.0)
				}
			case int:
				i = v + 1
			}

			if i < len(arr) {
				if i < 0 {
					i = 0
				}
				r = &array_iter{ data: i }
			}
		case Dictionary:{
			if p := array_tree_next_node(arr.data.dict, item); p != nil {
				r = &array_iter{ data: p.key }
			}
		}
	}
	return
}

func (a *stk_array) GetItem(idx *array_iter) (r interface{}) {
	if a != nil && idx != nil {
		switch a := a.(type) {
		case Array:
			r = a[idx.data.(int)]
		case Dictionary:
			if p := array_tree_find(a.data.dict, idx); p != nil {
				r = &p.data
			}
		}
		return
	}
}

func array_setitem(harr **stk_array, idx *array_iter, item Array) (r int) {
	r = -1
	if !harr != nil && *harr != nil && idx != nil {
		switch arr := harr.(type) {
		case *Array:
			if i, ok := idx.data.(int); ok {
				switch l := len(*arr); {
				case i > -1 && i < l:
					arr.data[i] = item
					r = len(*arr)
				case i == len(*arr):
					arr = append(arr, item)
					r = len(arr)
				}
			}
		case Dictionary:
			p := array_tree_find(arr.data.dict, idx)
			arr.items++
			p = array_tree_insert(&arr.data.dict, idx)
			p.data = item
			r = arr.items
		}
	}
	return
}

func array_insertitem(harr **stk_array, idx *array_iter, item Array) (r int) {
	r = -1
	switch {
	case harr == nil, *harr == nil, idx == nil:
	default:
		switch arr := &harr.(type) {
		case *Array:
			switch i := idx.(type) {
			case int:
				if i > -1 && i <= len(arr) {
					nu := make(Array, len(arr))
					copy(nu, (*arr)[:i])
					nu[i] = item
					copy(nu[i + 1:], (*arr)[i:])
					*arr = nu
					r = len(arr)
				}
			}
		case Dictionary:
			p := array_tree_find(arr.data.dict, idx)
			if p == nil {
				arr.items++
				p = array_tree_insert(&arr.data.dict, idx)
			}
			p.data = item
			r = arr.items
		}
	}
	return
}

func array_appenditem(harr *Array, item Array) (r int) {
	r = -1
	if harr != nil {
		r = array_setitem(harr, &inst{ data: len(*harr) }, item)
	}
	return
}

func array_getrange(arr interface{}, start, end *array_iter) (nu stk_array) {
	if len(arr) > 0 {
		switch arr := arr.(type) {
		case Array:
			sidx := start.(int)
			eidx := end.(int)
			if sidx < 0 {
				sidx = 0
			}
			if eidx > len(arr) {
				eidx = len(arr)
			}
			if sidx <= eidx {
				arr = arr[sidx:eidx]
				nu = make(Array, len(arr))
				copy(nu, arr)
			}
		case Dictionary:
			var s, e *array_tree

			nu = make(Dictionary)
			s = array_tree_find(arr->data.dict, start);
			if (!s) {
				s = array_tree_next_node(arr->data.dict, start);
				if (!s) {
					return nu;
				}
			}
			e = array_tree_find(arr->data.dict, end);
			if (!e) {
				e = array_tree_prev_node(arr->data.dict, end);
				if (!e) {
					return nu;
				}
			}
			if array_tree_compare(&s.key, &e.key, false) > 0 {
				return nu;
			}
			while (s) {
				array_setitem(&nu, &s->key, &s->data);
				if (s == e)
					break;
				s = array_tree_next_node(arr->data.dict, &s->key);
			}
		}		
	}
	return
}

func array_setrange(arr interface{}, start *array_iter, inarr interface{}) (r int) {
	r = -1
	if len(arr) != 0 {
		switch arr := arr.(type):
		case Array:
			if start != nil {
				switch start := start.data.(type) {
				case int:
					switch {
					case start < 0, start > len(arr):
					default:
						if inarr, ok := inarr.(Array); ok {
							copy(arr[start:], inarr)
							r = len(arr)
						}
					}
				}
			}
		case Dictionary:
			switch inarr := inarr.(type) {
			case Array:
				switch start := start.data.(type) {
				case int:
					for i, v := range inarr {
						arr[i + start] = v
					}
				case string:
					for i, v := range inarr {
						arr[fmt.Sprintf("%v%v", start, i)] = v
					}
				}
				len(arr)
			case Dictionary:
				for k, v := range inarr {
					arr[k] = v
				}
				r = len(arr)
			}
		}
	}
	return
}

func array_insertrange(harr **stk_array, start *array_iter, inarr *stk_array) (r int) {
	r = -1
	if harr != nil && *harr != nil {
		switch arr := harr.(type) {
		case *Array:
			if idx, ok := start.data.(int); ok {
				if idx > -1 && idx <= len(arr) {
					switch inarr := inarr.(type) {
					case nil:
						r = len(*arr)
					case Array:
						arr.Insert(idx, inarr)
						r = len(*arr)
					case SparseArray:
						//	FIXME: what about dictionaries with numeric keys?
					case Dictionary:
						//	FIXME: what does this even mean?
					}
				}
			}
		case Dictionary:
			switch inarr := inarr.(type) {
			case nil:
				r = len(*arr)
			case Array:
				for i, v := range inarr {
					arr[fmt.Sprint(i)] = v
				}
				r = len(arr)
			case Dictionary:
				for k, v := range inarr {
					arr[k] = v
				}
				r = len(arr)
			}
		}
	}
	return
}

func array_delrange(arr interface{}, start, end *array_iter) (r int) {
	r = -1
	if harr != nil && *harr != nil {
		switch arr := harr.(type) {
		case Array:
			sidx := start.data.(int)
			eidx := end.data.(int)
			if len(arr) > 0 {
				switch {
				case sidx < 0:
					sidx = 0
				case sidx >= len(arr):
					return -1
				}
				switch {
				case eidx >= len(arr):
					eidx = len(arr) - 1
				case eidx < 0:
					return -1
				}
				if sidx > eidx {
					return -1
				}
				copy(arr[sidx:eidx], arr[eidx:])
				for i := len(arr) - 1; i > eidx; i-- {
					arr[i] = nil
				}
				arr = arr[:eidx]
			}
			r = len(arr)
		case Dictionary:
			// FIXME

			var s, i *array_iter
			if s = array_tree_find(arr.data.dict, start); s == nil {
				if s = array_tree_next_node(arr.data.dict, start); s ==  nil {
					return arr.items
				}
			}
			if e = array_tree_find(arr.data.dict, end); e == nil {
				if e = array_tree_prev_node(arr.data.dict, end); e == nil {
					return arr.items
				}
			}
			if array_tree_compare(&s.key, &e.key, false) > 0 {
				return arr.items
			}
			for idx := s.key; array_tree_compare(&s.key, &e.key, false) <= 0; s = array_tree_next_node(arr.data.dict, &idx) {
				arr.data = array_tree_delete(&s.key, arr.data.dict)
				arr.items--
			}
			r = arr.items
		}
	}
	return
}

func array_delitem(harr **stk_array, item *array_iter) int {
	assert(harr != nil)
	assert(*harr != nil)
	assert(item != nil)
	return array_delrange(harr, item, &item)
}

/*\
|*| array_demote_only
|*| array demote discards the values of a dictionary, and
|*| returns a packed list of the keys.
|*| (Useful because keys are ordered and unique, presumably.)
|*| (This allows the keys to be abused as sets.)
\*/
func array_demote_only(arr Dictionary, threshold int) (r Array) {
	for k, v := range arr {
		if v.(int) >= threshold {
			r = append(r, k)
		}
	}
	return
}

/*\
|*| array_mash
|*| Takes the lists of values from the first array and
|*| uses each value as a key in the second array.  For
|*| each key, the passed "change_by" value is applied to
|*| any existing value.  If the value does not exist, it
|*| is set.
|*| This will be the core of the different/union/intersection
|*| code.
|*| Of course, this is going to absolutely blow chunks when
|*| passed an array with value types that can't be used as
|*| key values in an array.  Blast.  That may be infeasible
|*| regardless, though.
\*/
func array_mash(arr_in interface{}, mash Dictionary, value int) {
	if arr_in != nil && mash != nil {
		switch arr := arr_in.(type) {
		case Array:
			for key, keyval := range arr {
				if v := mash[keyval]; v == nil {
					mash[keyval] = value
				} else {
					if x, ok := v.(int); ok {
						mash[keyval] = x + value
					}
				}
			}
		case Dictionary:
			for _, keyval := range arr {
				var temp_value interface{}
				k := fmt.Sprint(keyval)
				if v := mash[k]; v == nil {
					mash[k] = value
				} else {
					if v, ok := v.(int); ok {
						mash[k] = v + value
					}
				}
			}
		default:
			panic(arr_in)
		}
	}
}

func array_is_homogenous(arr, example interface{}) (ok bool) {
	typ := reflect.TypeOf(example)
	ok = arr != nil
	switch arr := arr.(type) {
	case Array:
		for i, v := range arr {
			if reflect.Typeof(v) != typ {
				ok = false
				break
			}
		}
	case Dictionary:
		for k, v := range arr {
			if reflect.Typeof(v) != typ {
				ok = false
				break
			}
		}
	default:
		panic(arr)
	}
	return
}

/**** STRKEY ****/

func array_set_strkey(harr []*stk_array, key string, val *inst) (r int) {
	if v, ok := val.(*inst) {
		r = array_setitem(harr, &inst{ data: key }, val)
	} else {
		r = array_setitem(harr, &inst{ data: key }, &inst{ data: val })
	}
	return
}