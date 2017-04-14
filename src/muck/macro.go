package fbmuck

import (
	"bufio"
	"fmt"
	"os"
)

type Macro struct {
	Definition string
	Implementor ObjectID
}

type MacroTable map[string] Macro

func LoadMacros(f *FILE) (m MacroTable) {
	var name, def string
	var imp int
	var e error

	m = make(MacroTable)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		name = scanner.Text
		if scanner.Scan() {
			def = scanner.Text()
			if scanner.Scan() {
				if imp, e = strconv.Atoi(scanner.Text()); e == nil {
					m[name] = &Macro{ def, ObjectID(imp) }
				}
			}
		}
	}
	switch err := scanner.Err(); {
	case err != nil:
		fmt.Fprintln(os.Stderr, "reading macro definition:", err)
	case e != nil:
		fmt.Fprintln(os.Stderr, "reading macro implementor:", e)
	}
	return
}

func (m MacroTable) Dump(f *FILE) {
	for k, v := range m {
		fmt.Fprintln(f, k)
		fmt.Fprintln(f, m.Definition)
		fmt.Fprintf(f, "%v\n", m.Implementor)
	}
}

	struct macrotable {
		name string
		definition string
		implementor ObjectID
		left *macrotable
		right *macrotable
	};


func new_macro(name, definition string, player ObjectID) *macrotable {
	return &macrotable{ name: strings.ToLower(name), definition: definition, implementor: player }
}

func grow_macro_tree(struct macrotable *node, struct macrotable *newmacro) (r bool) {
	switch value := strings.Compare(newmacro.name, node.name); {
	case value < 0:
		if node.left != nil {
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

func insert_macro(macroname, macrodef string, player ObjectID, node **macrotable) (r int) {
	if newmacro := new_macro(macroname, macrodef, player); *node == nil {
		*node = newmacro
		r = 1
	} else {
		r = grow_macro_tree((*node), newmacro)
	}
	return
}

func do_list_tree(node *macrotable, first, last string, length int, player ObjectID) {
	static char buf[BUFFER_LEN];
	if node != nil {
		if strings.Compare(node.name[:len(first)], first) >= 0 {
			do_list_tree(node.left, first, last, length, player)
		}
		if strings.Compare(node.name[:len(first)], first) >= 0 && strings.Compare(node.name[:len(last)], last) <= 0 {
			if length > 0 {
				notify(player, fmt.Sprintf("%-16s %-16s  %s", node.name, DB.Fetch(node.implementor).name, node.definition))
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
		if node == Macros && !length {
			notify(player, buf)
			buf = ""
		}
	}
}

func list_macros(words []string, k int, player ObjectID, length int) {
	if k == 0 {
		do_list_tree(Macros, "\001", "\377", length, player)
	} else {
		k--
		do_list_tree(Macros, words[0], words[k], length, player)
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

func kill_macro(const char *macroname, ObjectID player, struct macrotable **mtop) (r bool) {
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
