package fbmuck

/*
 * This file includes the logic needed to parse the start of a database file and
 * determine whether it is old or new format, has old or new compression, etc.
 *
 * Needed because olddecomp.c suddenly gets a lot more complicated when we
 * have to handle all the modern formats.
 *
 * It contains the minimum amount of smarts needed to read this without
 * having to link in everything else.
 *
*/

func do_peek(f *FILE) int {
	int peekch;

	ungetc((peekch = getc(f)), f);

	return (peekch);
}

func getref(f *os.File) (r ObjectID) {
	//	Compiled in with or without timestamps, Sep 1, 1990 by Fuzzy, added to Muck by Kinomon.  Thanks Kino!
	switch peekch := do_peek(f); peekch {
	case NUMBER_TOKEN, LOOKUP_TOKEN:
	default:
		scanner := bufio.NewScanner(f)
		scanner.Scan()
		r = atol(scanner.Text())
	}
	return
}

static char xyzzybuf[BUFFER_LEN];
func getstring_noalloc(f *os.File) string {
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		xyzzybuf = scanner.Text()
		return xyzzybuf
	}
	if i := strings.Index(xyzzybuf, '\n'); i != -1 {
		xyzzybuf = xyzzybuf[:i]
	}
	return xyzzybuf
}

func db_read_header(f *FILE, version *string, load_format *int, grow *ObjectID, parmcnt *int) (r int) {
	var grow_and_dbflags bool
	const char *special;
	char c;

	/* null out the putputs */
	*version = NULL;
	*load_format = 0;
	*grow = 0;
	*parmcnt = 0;

	/* if the db doesn't start with a * it is incredibly ancient and has no header */
	/* this routine can deal with - just return */	
	c = getc( f );
	ungetc( c, f );
	if c != '*' {
		r |= DB_ID_OLDCOMPRESS; /* could be? */
		return
	} else {
		/* read the first line to id it */
		special = getstring_noalloc( f );

		/* save whatever the version string was */
		/* NOTE: This only works because we only do getstring_noalloc once */
		r |= DB_ID_VERSIONSTRING;
		*version = special;

		switch special {
		case "***TinyMUCK DUMP Format***":
			*load_format = 1
			r |= DB_ID_OLDCOMPRESS
		case "***Lachesis TinyMUCK DUMP Format***", "***WhiteFire TinyMUCK DUMP Format***":
			*load_format = 2
			r |= DB_ID_OLDCOMPRESS
		case "***Mage TinyMUCK DUMP Format***":
			*load_format = 3
			r |= DB_ID_OLDCOMPRESS
		case "***Foxen TinyMUCK DUMP Format***":
			*load_format = 4
			r |= DB_ID_OLDCOMPRESS
		case "***Foxen2 TinyMUCK DUMP Format***":
			*load_format = 5
			r |= DB_ID_OLDCOMPRESS
		case "***Foxen3 TinyMUCK DUMP Format***":
			*load_format = 6
			r |= DB_ID_OLDCOMPRESS
		case "***Foxen4 TinyMUCK DUMP Format***":
			*load_format = 6
			*grow = getref(f)
			r |= DB_ID_GROW
			r |= DB_ID_OLDCOMPRESS
		case "***Foxen5 TinyMUCK DUMP Format***":
			*load_format = 7
			grow_and_dbflags = true
		case "***Foxen6 TinyMUCK DUMP Format***":
			*load_format = 8
			grow_and_dbflags = true
		case "***Foxen7 TinyMUCK DUMP Format***":
			*load_format = 9
			grow_and_dbflags = true
		case "***Foxen8 TinyMUCK DUMP Format***":
			*load_format = 10
			grow_and_dbflags = true
		case "***Foxen9 TinyMUCK DUMP Format***":
			*load_format = 11
			grow_and_dbflags = true
		case "****Foxen Deltas Dump Extention***":
			*load_format = 4
			r |= DB_ID_DELTAS
		case "****Foxen2 Deltas Dump Extention***":
			*load_format = 5
			r |= DB_ID_DELTAS
		case "****Foxen4 Deltas Dump Extention***":
			*load_format = 6
			r |= DB_ID_DELTAS
		case "****Foxen5 Deltas Dump Extention***":
			*load_format = 7
			r |= DB_ID_DELTAS
		case "****Foxen6 Deltas Dump Extention***":
			*load_format = 8
			r |= DB_ID_DELTAS
		case "****Foxen7 Deltas Dump Extention***":
			*load_format = 9
			r |= DB_ID_DELTAS
		case "****Foxen8 Deltas Dump Extention***":
			*load_format = 10
			r |= DB_ID_DELTAS
		}

		/* All recent versions could have these */
		if grow_and_dbflags {
			*grow = getref(f)
			r |= DB_ID_GROW

			dbflags := getref(f)
			if dbflags & DB_PARMSINFO != 0 {
				*parmcnt = getref(f)
				r |= DB_ID_PARMSINFO
			}
			if dbflags & DB_COMPRESSED != 0 {
				r |= DB_ID_CATCOMPRESS
			}
		}
	}
	return
}
