package fbmuck

/* propdirs.c -- handles propdirs property creation, deletion, and finding.  */
/* WARNING: These routines all hack up the path string passed to them. */

/*
 * returns pointer to the new property node.  Returns a pointer to an
 *   existing elem of the given name, if one already exists.  Returns
 *   NULL if the name given is bad.
 * root is the pointer to the root propdir node.  This is updated in this
 *   routine to point to the new root node of the structure.
 * path is the name of the property to insert
 */
func (p *Plist) IsPropDir() bool {
	return p != nil && p.dir != nil
}

func (root *Plist) propdir_new_elem(path string) (p *Plist) {
	path = strings.TrimLeft(path, PROPDIR_DELIMITER)
	if len(path) != 0 {
		n := strchr(path, PROPDIR_DELIMITER)
		while (n && *n == PROPDIR_DELIMITER)
			*(n++) = '\0';

		if (n && *n) {
			/* just another propdir in the path */
			p = root.new_prop(path).dir.propdir_new_elem(n)
		} else {
			/* aha, we are finally to the property itself. */
			p = root.new_prop(path)
		}
	}
	return
}

/* returns pointer to the updated propdir structure's root node */
/* root is the pointer to the root propdir node */
/* path is the name of the property to delete */
func (root *Plist) propdir_delete_elem(path string) (p *Plist) {
	if root != nil {
		path = strings.TrimLeft(path, PROPDIR_DELIMITER)
		if len(path) == 0 {
			p = root
		} else {
			n := strchr(path, PROPDIR_DELIMITER)
			while (n && *n == PROPDIR_DELIMITER) {
				*(n++) = '\0'
			}

			if (n && *n) {
				/* just another propdir in the path */
				p = root.locate_prop(path)
				if p.IsPropDir() {
					/* yup, found the propdir */
					if p.dir = p.dir.propdir_delete_elem(n); p.dir == nil {
						if _, ok := p.data.(PROP_DIRTYP); ok {
							root = root.delete_prop(p.key)
						}
					}
				}
			} else {
				/* aha, we are finally to the property itself. */
				p = root.locate_prop(path)
				if p.IsPropDir() {
					p.dir = nil
				}
				root.delete_prop(path)
			}
			p = root
		}
	}
}

/* returns pointer to given property */
/* root is the pointer to the root propdir node */
/* path is the name of the property to find */
func (root *Plist) propdir_get_elem(path string) (p *Plist) {
	if root != nil {
		path = strings.TrimLeft(path, PROPDIR_DELIMITER)
		if len(path) > 0 {
			n := strchr(path, PROPDIR_DELIMITER);
			for n != nil && *n == PROPDIR_DELIMITER {
				*(n++) = '\0';
			}
			if n != nil && *n != 0 {
				/* just another propdir in the path */
				p = root.locate_prop(path)
				if p.IsPropDir() {
					return p.dir.propdir_get_elem(n)
				}
			} else {
				/* aha, we are finally to the property subname itself. */
				if p = root.locate_prop(path); p != nil {
					return p
				}
			}
		}
	}
	return
}

/* returns pointer to first property in the given propdir */
/* root is the pointer to the root propdir node */
/* path is the name of the propdir to find the first node of */
func (root *Plist) propdir_first_elem(path string) (r *Plist) {
	if path = strings.TrimLeft(path, PROPDIR_DELIMITER); path == "" {
		r = root.first_node()
	} else {
		if r = root.propdir_get_elem(path); r != nil && r.dir != nil {
			r = r.dir.first_node()
		}
	}
	return
}

/* returns pointer to next property after the given one in the propdir */
/* root is the pointer to the root propdir node */
/* path is the name of the property to find the next node after */
/* Note: Finds the next alphabetical property, regardless of the existence of the original property given. */
func (root *Plist) propdir_next_elem(path string) (p *Plist) {
	if root != nil {
		if path = strings.TrimLeft(path, PROPDIR_DELIMITER); path != "" {
			n := strchr(path, PROPDIR_DELIMITER)
			while (n && *n == PROPDIR_DELIMITER)
				*(n++) = '\0';

			if n != "" {
				p = root.locate_prop(path)
				if p.IsPropDir() {
					p = p.dir.propdir_next_elem(n)
				}
			} else {
				p = root.next_node(path)
			}
		}
	}
	return
}

func propdir_name(name string) (r string) {
	for name != "" {
		name = strings.TrimLeft(name, PROPDIR_DELIMITER)
		r += PROPDIR_DELIMITER
		if n := strings.Split(name, PROPDIR_DELIMITER); len(n) > 0 {
			r += n[0]
		}
	}
	return r[:strings.LastIndex(r, PROPDIR_DELIMITER)]
}