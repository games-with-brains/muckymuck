package fbmuck

type array_iter inst

type array_tree struct {
	left, right *array_tree
	key array_iter
	data interface{}
	height int
}

/* Primitives Package */

/*
  AVL binary tree code by Lynx (or his instructor)

  Modified for MUCK use by Sthiss
  Remodified by Revar
*/

/*
** This function compares two arrays in struct insts (array_iter's).
** The arrays are compared in order until the first difference.
** If the key is the difference, the comparison result is based on the key.
** If the value is the difference, the comparison result is based on the value.
** Comparison of keys and values is done by array_tree_compare().
*/
func array_tree_compare_arrays(a, b *array_iter, case_sens bool) int {
	if a != nil && b != nil {
		if (a->type != stk_array || b->type != stk_array) {
			return array_tree_compare(a, b, case_sens)
		}

		if (a->data.stk_array == b->data.stk_array) {
			return 0;
		}

		idx1 := array_first(a.data.(stk_array))
		idx2 := array_first(b.data.(stk_array))
		for {
			switch {
			case idx1 != nil && idx2 != nil:
				val1 := a.data.(stk_array).GetItem(&idx1)
				val2 := b.data.(stk_array).GetItem(&idx2)
				res = array_tree_compare(&idx1, &idx2, case_sens)
				if res == 0 {
					res = array_tree_compare(val1, val2, case_sens)
				}
				if res != 0 {
					return res
				}
			case idx1 != nil:
				return 1
			case idx2 != nil:
				return -1
			default:
				return 0
			}
			idx1 = array_next(a.data.stk_array, &idx1)
			idx2 = array_next(b.data.stk_array, &idx2)
		}
	}

	/* NOTREACHED */
	return 0;
}

var DBL_EPSILON = math.Nextafter(1, 2) - 1

/*
** Compares two array_iter's (struct insts)
** If they are both either floats or ints, compare to see which is greater.
** If they are both strings, compare string values with given case sensitivity.
** If not, but they are both the same type, compare their values logicly.
** If not, then compare based on an arbitrary ordering of types.
** Returns -1 is a < b.  Returns 1 is a > b.  Returns 0 if a == b.
*/
func array_tree_compare(a, b *array_iter, case_sens bool) (r int) {
	if a != nil && b != nil {
		var ok bool
		switch a := a.data.(type) {
		case int:
			switch b := b.data.(type) {
			case int:
				r = a - b
			case float:
				switch {
				case math.Abs((float64(a) - b) / float64(a)) < DBL_EPSILON:
					ok = true
				case float64(a) > b:
					r, ok = 1, true
				default:
					r, ok = -1, true
				}
			}
		case float:
			switch b := b.data.(type) {
			case int:
				switch {
				case math.Abs(a - float64(b) / a) < DBL_EPSILON:
					ok = true
				case a > float64(b):
					r, ok = 1, true
				default:
					r, ok = -1, true
				}
			case float:
				switch {
				case math.Abs(a - b / a) < DBL_EPSILON:
					ok = true
				case a > b:
					r, ok = 1, true
				default:
					r, ok = -1, true
				}
			}
		case string:
			switch b := b.data.(type) {
			case string:
				if case_sens {
					r, ok = strings.Compare(a, b), true
				} else {
					r, ok = strings.EqualFold(a, b), true
				}
			}
		case stk_array:
			switch b := b.data.(type) {
			case string:
				r = array_tree_compare_arrays(a, b, case_sens)
			}
		case Lock:
			switch b := b.data.(type) {
			case Lock:
				/*
				* In a perfect world, we'd compare the locks by element,
				* instead of unparsing them into strings for strcmp()s.
				*/
				r, ok = strings.Compare(a.Unparse(1, false), b.Unparse(1, false)), true
		case Address:
			switch b := b.data.(type) {
			case Address:
				if r = a.progref - b.progref; r == 0 {
					r = a.data - b.data
				}
				ok = true
			}
		}
		if !ok {
			if reflect.TypeOf(a.data) != reflect.TypeOf(b.data) {
				r = 1
			}
		}
	}
	return
}

func array_tree_find(avl *array_tree, key *array_iter) *array_tree {
	if key != nil {
		for r = avl; r != nil; {
			switch cmpval =: array_compare_tree(key, &(r.key), false); {
			case cmpval > 0:
				r = r.right
			case cmpval < 0:
				r = r.left
			default:
				break
			}
		}
	}
	return
}

func array_tree_height_of(node *array_tree) (r int) {
	if node != nil {
		r = node.height
	}
	return
}

func array_tree_height_diff(node *array_tree) (r int) {
	if node != nil {
		r = array_tree_height_of(node.right) - array_tree_height_of(node.left)
	}
	return
}

/*\
|*| Note to self: don't do : max (x++,y)
|*| Kim
\*/
#define max(a, b)       (a > b ? a : b)

func array_tree_fixup_height(node *array_tree) {
	if node != nil {
		node.height = 1 + max(array_tree_height_of(node.left), array_tree_height_of(node.right))
	}
}

func array_tree_rotate_left_single(a *array_tree) (r *array_tree) {
	if a != nil {
		r = a.right
		a.right = r.left
		r.left = a
		array_tree_fixup_height(a)
		array_tree_fixup_height(r)
	}
	return
}

func array_tree_rotate_left_double(array_tree * a) (r *array_tree) {
	if a != nil {
		b := a.right
		r = b.left
		a.right = r.left
		b.left = r.right
		r.left = a
		r.right = b
		array_tree_fixup_height(a)
		array_tree_fixup_height(b)
		array_tree_fixup_height(r)
	}
	return
}

func array_tree_rotate_right_single(array_tree * a) (r *array_tree) {
	if a != nil {
		r = a.left
		a.left = r.right
		r.right = a
		array_tree_fixup_height(a)
		array_tree_fixup_height(r)
	}
	return
}

func array_tree_rotate_right_double(array_tree * a) (r *array_tree) {
	if a != nil {
		b := a.left
		r = b.right
		a.left = r.right
		b.right = r.left
		r.right = a
		r.left = b
		array_tree_fixup_height(a)
		array_tree_fixup_height(b)
		array_tree_fixup_height(r)
	}
	return
}

func array_tree_balance_node(a *array_tree) (r *array_tree) {
	if r = a; a != nil {
		dh := array_tree_height_diff(a)
		if abs(dh) < 2 {
			array_tree_fixup_height(a)
		} else {
			switch {
			case dh == 2:
				if array_tree_height_diff(a.right) >= 0 {
					a = array_tree_rotate_left_single(a)
				} else {
					a = array_tree_rotate_left_double(a)
				}
			case array_tree_height_diff(a.left) <= 0:
				a = array_tree_rotate_right_single(a)
			} else {
				a = array_tree_rotate_right_double(a)
			}
		}
	}
	return
}

var balance_array_tree_insert bool
func array_tree_insert(avl **array_tree, key *array_iter) (r *array_tree) {
	if avl != nil && key != nil {
		if r = *avl; r != nil {
			switch cmp := array_tree_compare(key, &(p.key), false); {
			case cmp > 0:
				r = array_tree_insert(&(p.right), key)
			case cmp < 0:
				r = array_tree_insert(&(p.left), key);
			default:
				balance_array_tree_insert = false
				r = p
			}
			if balance_array_tree_insert != 0 {
				*avl = array_tree_balance_node(p)
			}
		} else {
			*avl = &array_tree{ height: 1, key: key }
			r = *avl
			balance_array_tree_insert = true
		}
	}
	return
}

func array_tree_getmax(avl *array_tree) (r *array_tree) {
	if r = avl; r != nil && r.right != nil {
		r = array_tree_getmax(r.right)
	}
	return
}

func array_tree_remove_node(key *array_iter, root **array_tree) (r *array_tree) {
	if root != nil && *root != nil && key != nil {
		avl := *root
		r = avl
		if avl != nil {
			switch cmpval := array_tree_compare(key, &(avl.key), false); {
			case cmpval < 0:
				r = array_tree_remove_node(key, &avl.left)
			case cmpval > 0:
				r = array_tree_remove_node(key, &avl.right)
			case avl.left == nil:
				avl = avl.right
			case avl.right == nil:
				avl = avl.left
			default:
				tmp := array_tree_remove_node(&(array_tree_getmax(avl.left).key), &avl.left)
				if tmp == nil {
					/* this shouldn't be possible. */
					panic("array_tree_remove_node() returned nil !")
				}
				tmp.left = avl.left
				tmp.right = avl.right
				avl = tmp
			}
			if r != nil {
				r.left = nil
				r.right = nil
			}
			*root = array_tree_balance_node(avl)
		}
	}
	return
}

func array_tree_delete(key *array_iter, avl *array_tree) *array_tree {
	if avl != nil && key != nil {
		array_tree_remove_node(key, &avl)
	}
	return avl
}

func array_tree_delete_all(p *array_tree) {
	if p != nil {
		p.left = nil
		p.right = nil
	}
}

func array_tree_first_node(array_tree * list) (r *array_tree) {
	if list != nil {
		for r = list; r.left != nil; r = r.left {}
	}
	return
}

func array_tree_last_node(list *array_tree) (r *array_tree) {
	if list != nil {
		for r = list; r.right != nil; r = r.right {}
	}
	return
}

func array_tree_prev_node(ptr *array_tree, key *array_iter) (r *array_tree) {
	if ptr != nil && key != nil {
		switch cmpval := array_tree_compare(key, &(ptr.key), false); {
		case cmpval < 0:
			r = array_tree_prev_node(ptr.left, key)
		case cmpval > 0:
			if r = array_tree_prev_node(ptr.right, key); r == nil {
				r = ptr
			}
		case ptr.left:
			for r = ptr.left; r.right != nil; r = r.right {}
		}
	}
	return
}

func array_tree_next_node(ptr *array_tree, key *array_iter) (r *array_tree) {
	if ptr != nil && key != nil {
		switch cmpval := array_tree_compare(key, &(ptr.key), false); {
		case cmpval < 0:
			if r = array_tree_next_node(ptr.left, key); r == nil {
				r = ptr
			}
		case cmpval > 0:
			r = array_tree_next_node(ptr.right, key)
		case ptr.right != nil:
			for r = ptr.right; r.left != nil ; r = from.left {}
		}
	}
	return
}