package fbmuck

import (
	"bufio"
	"fmt"
	"os"
)

type Macro struct {
	Definition string
	Author ObjectID
}

type MacroTable map[string] *Macro

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
		fmt.Fprintf(f, "%v\n", m.Author)
	}
}

func (m MacroTable) Expand(name string) (r string) {
	if d := m[strings.ToLower(name)]; d != nil {
		r = d.Definition
	}
	return
}

//	Macros can't be redefined
func (m MacroTable) Create(name, def string, player ObjectID) (r bool) {
	name = strings.ToLower(name)
	if _, r = m[name]; !ok {
		m[name] = &Macro{ Definition: def, Author: player }
	}
	return
}

func (m MacroTable) list(words []string, player ObjectID, f func(string, *Macro)) {
	if len(words) > 0 {
		for _, v := range words {
			if d, ok := m[v]; ok {
				f(v, d)
			}
		}
	} else {
		for k, d := range m {
			f(k, d)
		}
	}
	notify(player, "End of list.")
}

func (m MacroTable) Describe(words []string, player ObjectID) {
	m.list(words, player, func(k string, v *Macro) {
		notify(player, fmt.Sprintf("%-16s %-16s  %s", k, DB.Fetch(v.Author).name, v.Definition))
	})
}

func (m MacroTable) List(words []string, player ObjectID) {
	var terms []string
	m.list(words, player, func(k string, v *Macro) {
		terms = append(terms, fmt.Sprintf("%-16s", v))
		if buf := strings.Join(terms, " "); len(buf) > 70 {
			notify(player, buf)
			terms = nil
		}
	})
}

func (m MacroTable) Delete(name string, player ObjectID) (r bool) {
	if _, r = m[name]; r {
		delete(m, name)
	}
	return
}