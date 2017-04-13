package fbmuck

/* property.c
   A whole new lachesis mod.
   Adds property manipulation routines to TinyMUCK.   */

/* Completely rewritten by darkfox and Foxen, for propdirs and other things */

func set_property_nofetch(player dbref, name string, dat interface{}) {
	if name != "" {
		name = strings.TrimLeft(name, PROPDIR_DELIMITER)
		if db.Fetch(player).flags & LISTENER != 0 && (strings.Prefix(name, "_listen") || strings.Prefix(name, "~listen") || strings.Prefix(name, "~olisten")) {
			db.Fetch(player).flags |= LISTENER
		}
		buf := name
		if n := strings.SplitN(buf, PROP_DELIMITER, 2); len(n) > 0 {
			buf = [0]
		}
		if buf != "" {
			p := db.Fetch(player).properties.propdir_new_elem(name)
			p.clear_propnode()
			SetPFlagsRaw(p, dat.flags)
			switch dat.(type) {
			case string:
				if dat == "" {
					SetPType(p, PROP_DIRTYP)
					p.data = nil
					if p.dir == nil {
						remove_property_nofetch(player, name)
					}
				} else {
					p.data = dat
				}
			case int:
				if p.data = dat; dat == 0 {
					SetPType(p, PROP_DIRTYP)
					if p.dir == nil {
						remove_property_nofetch(player, name)
					}
				}
			case float64:
				if p.data = dat; dat == 0.0 {
					SetPType(p, PROP_DIRTYP)
					if p.dir == nil {
						remove_property_nofetch(player, name)
					}
				}
			case dbref:
				if p.data = dat; dat == NOTHING {
					SetPType(p, PROP_DIRTYP)
					p.data = nil
					if p.dir == nil {
						remove_property_nofetch(player, name)
					}
				}
			case Lock:
				p.data = dat
			case PROP_DIRTYP:
				if p.data = nil; p.dir == nil {
					remove_property_nofetch(player, pname)
				}
			}
		}
	}
}

func set_property(player dbref, name string, dat interface{}) {
	set_property_nofetch(player, name, dat)
	db.Fetch(player).flags |= OBJECT_CHANGED
}

func set_lock_property(player dbref, pname, lock string) {
	if lock == "" {
		set_property(player, pname, UNLOCKED)
	} else {
		set_property(player, pname, ParseLock(-1, 1, lock, 1))
	}
}

/* adds a new property to an object */
func add_prop_nofetch(player dbref, pname, strval string, value int) {
	switch {
	case strval != "":
		set_property_nofetch(player, pname, strval)
	case value > 0:
		set_property_nofetch(player, pname, value)
	default:
		set_property_nofetch(player, pname, &PData{ flags: PROP_DIRTYP })
	}
}

/* adds a new property to an object */
func add_property(player dbref, pname, strval string, value int) {
	add_prop_nofetch(player, pname, strval, value)
	db.Fetch(player).flags |= OBJECT_CHANGED
}

func remove_proplist_item(player dbref, p *Plist, all bool) {
	if p != nil {
		if ptr := p.key; !Prop_System(ptr) {
			if !all {
				if ptr == PROP_HIDDEN || ptr == PROP_SEEONLY || ptr[0] == '_' && ptr[1] == '\0' || PropFlags(p) & PROP_SYSPERMS {
					return
				}
			}
			remove_property(player, ptr)
		}
	}
}

/* removes property list --- if it's not there then ignore */
func remove_property_list(player dbref, all int) {
	if l := db.Fetch(player).properties; l != nil {
		p := l.first_node()
		for p != nil {
			n := l.next_node(p.key)
			remove_proplist_item(player, p, all)
			l = db.Fetch(player).properties
			p = n
		}
	}
	db.Fetch(player).flags |= OBJECT_CHANGED
}

/* removes property --- if it's not there then ignore */
func remove_property_nofetch(player dbref, name string) {
	db.Fetch(player).properties = db.Fetch(player).properties.propdir_delete_elem(name)
	db.Fetch(player).flags |= OBJECT_CHANGED
}

func remove_property(player dbref, pname string) {
	remove_property_nofetch(player, pname)
}

func get_property(player dbref, name string) (p *Plist) {
	return db.Fetch(player).properties.propdir_get_elem(name)
}

/* checks if object has property, returning true if it or any of its contents has the property stated */
func contains_property(descr int, player, what dbref, pname, strval string, value int) (r bool) {
	if has_property(descr, player, what, pname, strval, value) {
		r = true
	} else {
		for things := db.Fetch(what).Contents; things != NOTHING; things = db.Fetch(things).next {
			if contains_property(descr, player, things, pname, strval, value) {
				r = true
			}
		}
		if r == 0 && tp_lock_envcheck {
			for things := getparent(what); things != NOTHING; things = getparent(things) {
				if has_property(descr, player, things, pname, strval, value) {
					r = true
					break
				}
			}
		}
	}
	return
}

var has_prop_recursion_limit = 2
/* checks if object has property, returns true if it has the property */
func has_property(descr int, player, what dbref, name, strval string, value int) (r bool) {
	if p := get_property(what, pname); p != nil {
		switch v := p.data.(type) {
		case string:
			if has_prop_recursion_limit--; has_prop_recursion_limit > 0 {
				if PropFlags(p) & PROP_BLESSED != 0 {
					v = do_parse_mesg(descr, player, what, v, "(Lock)", MPI_ISPRIVATE | MPI_ISLOCK | MPI_ISBLESSED)
				} else {
					v = do_parse_mesg(descr, player, what, v, "(Lock)", MPI_ISPRIVATE | MPI_ISLOCK)
				}
			}
			has_prop_recursion_limit++
			r = !smatch(strval, v)
		case int:
			r = value == v
		case float64:
			r = value == int(v)
		}
	}
	return
}

/* return string value of property */
func get_property_class(player dbref, name string) (r string) {
	if p := get_property(player, pname); p {
		if v, ok := p.data.(string); ok {
			r = v
		}
	}
	return
}

/* return value of property */
func get_property_value(player dbref, name string) (r int) {
	if p := get_property(player, name); p != nil {
		if v, ok := p.data.(int); ok {
			r = v
		}
	}
}

/* return float value of a property */
func get_property_fvalue(player dbref, name string) (r float) {
	if p := get_property(player, name); p != nil {
		if ok, v := p.data.(float64); ok {
			r = v
		}
	}
	return
}

func get_property_dbref(player dbref, name string) (r dbref) {
	r = NOTHING
	if p := get_property(player, name); p != nil {
		if v, ok := p.data.(dbref); ok {
			r = v
		}
	}
	return
}

func get_property_lock(player dbref, name string) (r Lock) {
	if p := get_property(player, pname); p == nil {
		r = UNLOCKED
	} else {
		if v, ok := p.data.(Lock); ok {
			r = v
		} else {
			r = UNLOCKED
		}
	}
	return
}

/* return flags of property */
func get_property_flags(player dbref, name string) (r int) {
	if p := get_property(player, name); p != nil {
		r = PropFlags(p)
	}
	return
}

/* return flags of property */
func clear_property_flags(player dbref, name string, flags int) {
	flags &= ~PROP_TYPMASK
	if p := get_property(player, name); p != nil {
		SetPFlags(p, (PropFlags(p) & ~flags))
	}
}

/* return flags of property */
func set_property_flags(player dbref, name string, flags int) {
	flags &= ~PROP_TYPMASK
	if p := get_property(player, name); p != nil {
		SetPFlags(p, PropFlags(p) | flags)
	}
}

func copy_prop(old dbref) *Plist {
	return db.Fetch(old).properties.copy_proplist(old)
}

/* Return a pointer to the first property in a propdir and duplicates the
   property name into 'name'.  Returns NULL if the property list is empty
   or does not exist. */
func (list *Plist) first_prop_nofetch(player dbref, dir, name string) (n string, p *Plist) {
	if dir = strings.TrimLeft(dir, PROPDIR_DELIMITER); dir = "" {
		*list = db.Fetch(player).properties
		if p = *list.first_node(); p != nil {
			n = p.key
		}
	} else {
		buf := dir
		p = db.Fetch(player).properties.propdir_get_elem(buf)
		*list = p
		if p != nil {
			*list = p.dir
			if p = *list.first_node(); p != nil {
				name = p.key
			}
		}
	}
	return
}

/* first_prop() returns a pointer to the first property.
 * player    dbref of object that the properties are on.
 * dir       pointer to string name of the propdir
 * list      pointer to a proplist pointer.  Returns the root node.
 * name      printer to a string.  Returns the name of the first node.
 */
func (list *Plist) first_prop(player dbref, dir string, name string) (p *Plist) {
	_, p = list.first_prop_nofetch(player, dir, name)
	return
}

/* next_prop() returns a pointer to the next property node.
 * list    Pointer to the root node of the list.
 * name    Pointer to a string.  Returns the name of the next property.
 */
func (prop *Plist) next_prop(list *Plist) (p *Plist, name string) {
	if p = prop; p != nil {
		if p = list.next_node(p.key); p != nil {
			name = p.key
		}
	}
	return
}

/* next_prop_name() returns a ptr to the string name of the next property.
 * player   object the properties are on.
 * outbuf   pointer to buffer to return the next prop's name in.
 * name     pointer to the name of the previous property.
 *
 * Returns null if propdir doesn't exist, or if no more properties in list.
 * Call with name set to "" to get the first property of the root propdir.
 */
func next_prop_name(player dbref, name string) (r string) {
	switch buf := name; {
	case len(name) == 0, name[len(name) - 1] == PROPDIR_DELIMITER:
		if p := propdir_first_elem(db.Fetch(player).properties, name); p != nil {
			r = name + p.key
		}
	} else {
		if p := db.Fetch(player).properties.propdir_next_elem(buf); p != nil {
			ptr := strrchr(name, PROPDIR_DELIMITER)
			if ptr == nil {
				ptr = name
			}
			r = ptr + PROPDIR_DELIMITER + p.key
		}
	}
	return
}

func is_propdir(player dbref, name string) bool {
	return db.Fetch(player).properties.propdir_get_elem(name).IsPropDir()
}

func envprop(where dbref, propname string) (obj dbref, p *Plist) {
	for obj = where; obj != NOTHING && p == nil; obj = getparent(obj) {
		p = get_property(obj, propname)
	}
	return
}

func envpropstr(where dbref, propname string) (obj dbref, r string) {
	var p interface{}
	obj, p = envprop(where, propname)
	if v, ok := p.(string); ok {
		r = v
	}
	return
}

func displayprop(player, obj dbref, name string) (r string) {
	int pdflag;
	if p := get_property(obj, name); p == nil {
		r = fmt.Sprint("%v: No such property.", name)
	} else {
		blesschar := "-"
		if PropFlags(p) & PROP_BLESSED != 0 {
			blesschar = 'B'
		}

		var mybuf string
		if p.dir != nil {
			mybuf = fmt.Sprint(name, PROPDIR_DELIMITER)
		} else {
			mybuf = name
		}

		switch v := p.data.(type) {
		case string:
			r = fmt.Sprintf("%c str %s:%v", blesschar, mybuf, v)
		case dbref:
			r = fmt.Sprintf("%c ref %s:%s", blesschar, mybuf, unparse_object(player, v))
		case int:
			r = fmt.Sprintf("%c int %s:%d", blesschar, mybuf, v)
		case float64:
			r = fmt.Sprintf("%c flt %s:%.17g", blesschar, mybuf, v)
		case Lock:		//	FIXME: lock
			r = fmt.Sprintf("%c lok %s:%s", blesschar, mybuf, v.Unparse(player, true))
		case PROP_DIRTYP:
			r = fmt.Sprintf("%c dir %s:(no value)", blesschar, mybuf)
		}
	}
	return
}

extern short db_conversion_flag;

func corrupt_property_warning(strfmt string, args ...interface{}) {
	wall_wizards("## WARNING! A corrupt property was found while trying to read it from disk.")
	wall_wizards("##   This property has been skipped and will not be loaded.  See the sanity")
	wall_wizards("##   logfile for technical details.")
	log_sanity(strfmt, args...)
}

func db_get_single_prop(f *FILE, obj dbref, pos int, pnode *Plist, pdir string) (r int) {
	r = 1
	var tpos int
	if pos != 0 {
		fseek(f, pos, 0)
	}

	var getprop_buf string
	switch name := fgets(getprop_buf, sizeof(getprop_buf), f); {
	case len(name) == 0:
		corrupt_property_warning("Failed to read property from disk: Failed disk read.  obj = #%d, pos = %ld, pdir = %s", obj, pos, pdir)
		r = -1
	case name[0] == '*' && name != "*End*\n":
		r = 0
	default:
		switch flags := strchr(name, PROP_DELIMITER); {
		case len(flags) == 0:
			corrupt_property_warning("Failed to read property from disk: Corrupt property, flag delimiter not found.  obj = #%d, pos = %ld, pdir = %s, data = %s", obj, pos, pdir, name)
			r = -1
		case !unicode.IsNumber(flags):
			corrupt_property_warning("Failed to read property from disk: Corrupt property flags.  obj = #%d, pos = %ld, pdir = %s, data = %s:%s:%s", obj, pos, pdir, name, flags, value)
			r = -1
		default:
			switch value := strchr(flags, PROP_DELIMITER); {
			case len(value) == 0:
				corrupt_property_warning()
				log_sanity("Failed to read property from disk: Corrupt property, value delimiter not found.  obj = #%d, pos = %ld, pdir = %s, data = %s:%s", obj, pos, pdir, name, flags)
				r = -1
			default:
				p := strchr(value, '\n')
				switch flg := strconv.Atoi(flags); flg & PROP_TYPMASK {
				case PROP_STRTYP:
					if pos != 0 {
						if pnode != nil {
							pnode.data = value
							SetPFlagsRaw(pnode, flg)
						} else {
							set_property_nofetch(obj, name, value)
						}
					} else {
						set_property_nofetch(obj, name, tpos)
					}
				case PROP_LOKTYP:
					if pos != 0 {
						if lock := ParseLock(-1, 1, value, 32767); pnode != nil {
							pnode.data = lock
							SetPFlagsRaw(pnode, flg)
						} else {
							set_property_nofetch(obj, name, lock)
						}
					} else {
						set_property_nofetch(obj, name, tpos)
					}
				case PROP_INTTYP:
					if i, e := strconv.Atoi(value); e == nil {
						set_property_nofetch(obj, name, i)
					} else {
						corrupt_property_warning("Failed to read property from disk: Corrupt integer value.  obj = #%d, pos = %ld, pdir = %s, data = %s:%s:%s", obj, pos, pdir, name, flags, value)
						r = -1
					}
				case PROP_FLTTYP:
					var mydat float64
					if !unicode.IsNumber(value) && ifloat(value) {
						tpnt := value
						var dtemp bool
						switch tpnt[0] {
						case '+':
							tpnt = tpnt[1:]
						case '-'
							dtemp = true
							tpnt = tpnt[1:]
						}
						tpnt = strings.ToUpper(tpnt)
						switch {
						case strings.HasPrefix(tpnt, "INF"):
							if dtemp {
								mydat = math.Inf(-1)
							} else {
								mydat = math.Inf(1)
							}
						case strings.HasPrefix(tpnt, "NAN"):
							/* FIXME: This should be NaN. */
							mydat = math.Inf(1)
						}
					} else {
						sscanf(value, "%lg", &mydat)
					}
					set_property_nofetch(obj, name, mydat)
				case PROP_REFTYP:
					if !unicode.IsNumber(value) {
						corrupt_property_warning("Failed to read property from disk: Corrupt dbref value.  obj = #%d, pos = %ld, pdir = %s, data = %s:%s:%s", obj, pos, pdir, name, flags, value)
						r = -1
					}
				}
			}
		}
	}
}

func db_getprops(f *FILE, obj dbref, dir string) {
	for db_get_single_prop(f, obj, 0, nil, dir) != nil {}
}

func (p *Plist) db_putprop(f *os.File, dir string) {
	var buf string
	switch v := p.data.(type) {
	case PROP_DIRTYP:
	case int:
		if v != 0 {
			buf = intostr(v)
		}
	case float64:
		if v != 0 {
			buf = fmt.Sprintf("%.17g", v)
		}
	case dbref:
		if v != NOTHING {
			buf = intostr(v)
		}
	case string:
		if v != "" {
			buf = v
		}
	case Lock:
		if !v.IsTrue() {
			buf = v.Unparse(1, false)
		}
	}
	if buf != "" {
		if _, err := f.WriteString(fmt.Sprintf("%v%v%v%v%v%v\n", dir[1:], p.key, PROP_DELIMITER, PropFlagsRaw(p), PROP_DELIMITER, buf)); err != nil {
			log_sanity("Failed to write out property %v.db_putprop(%v, %v)", p, f, dir)
			abort()
		}
	}
}

func (p *Plist) db_dump_props_rec(obj dbref, f *FILE, dir string) (r int) {
	if p != nil {
		r = p.left.db_dump_props_rec(obj, f, dir)
		p.db_putprop(f, dir)
		if p.dir != nil {
			r += p.dir.db_dump_props_rec(obj, f, dir + p.key + PROPDIR_DELIMITER)
		}
		r += p.right.db_dump_props_rec(obj, f, dir)
	}
	return
}

func db_dump_props(f *FILE, obj dbref) {
	db.Fetch(obj).properties.db_dump_props_rec(obj, f, "/")
}

func reflist_add(dbref obj, const char* propname, dbref toadd) {
	if ptr := get_property(obj, propname); ptr != nil {
		switch temp := ptr.data.(type) {
		case string:
			var count, charcount int
			var pat, outbuf string

			list := temp
			buf := fmt.Sprint(toadd)
			for _, v := range temp {
				if v == NUMBER_TOKEN {
					pat = buf
					count++
					charcount = temp - list
				} else if (pat) {
					if (!*pat) {
						if (!*temp || *temp == ' ') {
							break
						}
						pat = ""
					} else if (*pat != *temp) {
						pat = ""
					} else {
						pat++
					}
				}
				temp++
			}
			if (pat && !*pat) {
				if (charcount > 0) {
					outbuf = list
				}
				outbuf += temp
			} else {
				outbuf = list
			}
			outbuf += fmt.Sprintf(" #%d", toadd)
			temp = strings.TrimLeftFunc(outbuf, unicode.IsSpace)
			add_property(obj, propname, temp, 0)
		case dbref:
			if temp != toadd {
				add_property(obj, propname, fmt.Sprintf("#%d #%d", temp, toadd), 0)
			}
		default:
			add_property(obj, propname, fmt.Sprintf("#%d", toadd), 0)
		}
	} else {
		add_property(obj, propname, fmt.Sprintf("#%d", toadd), 0)
	}
}

func reflist_del(dbref obj, const char* propname, dbref todel) {
	if ptr := get_property(obj, propname); ptr != nil {
		switch temp := ptr.data.(type) {
		case string:
			var pat string
			var count, charcount

			list := temp
			buf := fmt.Sprint(todel)
			for _, v := range temp {
				if v == NUMBER_TOKEN {
					pat = buf
					count++
					charcount = temp - list
				} else if (pat) {
					if (!*pat) {
						if (!*temp || *temp == ' ') {
							break
						}
						pat = ""
					} else if (*pat != *temp) {
						pat = ""
					} else {
						pat++
					}
				}
			}
			if pat != "" {
				var outbuf string
				if charcount > 0 {
					outbuf = list
				}
				outbuf += temp
				temp = strings.TrimLeftFunc(outbuf, unicode.IsSpace)
				add_property(obj, propname, temp, 0)
			}
		case dbref:
			if temp == todel {
				add_property(obj, propname, "", 0)
			}
		}
	}
}

func reflist_find(obj dbref, propname string, tofind dbref) (r int) {
	if ptr := get_property(obj, propname); ptr != nil {
		switch temp := ptr.data.(type) {
		case string:
			buf := fmt.Sprint(tofind)
			var count int
			var pat string
			for _, v := range temp {
				switch {
				case v == NUMBER_TOKEN:
					pat += v
					count++
				case pat == "":
					if temp == "" || temp[0] == ' ' {
						break
					}
					pat = ""
				case pat[0] != temp[0]:
					pat = ""
				default:
					pat = pat[1:]
				}
				temp = temp[1:]
			}
			if pat != "" {
				r = count
			}
		case dbref:
			if temp == tofind {
				r = 1
			}
		}
	}
	return
}