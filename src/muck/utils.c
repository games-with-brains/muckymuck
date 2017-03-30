/* $Header: /cvsroot/fbmuck/fbmuck/src/utils.c,v 1.3 2005/07/04 12:04:24 winged Exp $ */

/*
 * $Log: utils.c,v $
 * Revision 1.3  2005/07/04 12:04:24  winged
 * Initial revisions for everything.
 *
 * Revision 1.2  2000/03/29 12:21:02  revar
 * Reformatted all code into consistent format.
 * 	Tabs are 4 spaces.
 * 	Indents are one tab.
 * 	Braces are generally K&R style.
 * Added ARRAY_DIFF, ARRAY_INTERSECT and ARRAY_UNION to man.txt.
 * Rewrote restart script as a bourne shell script.
 *
 * Revision 1.1.1.1  1999/12/16 03:23:29  revar
 * Initial Sourceforge checkin, fb6.00a29
 *
 * Revision 1.1.1.1  1999/12/12 07:27:44  foxen
 * Initial FB6 CVS checkin.
 *
 * Revision 1.1  1996/06/12 03:07:13  foxen
 * Initial revision
 *
 * Revision 5.3  1994/03/14  12:20:58  foxen
 * Fb5.20 release checkpoint.
 *
 * Revision 5.2  1994/01/18  20:52:20  foxen
 * Version 5.15 release.
 *
 * Revision 5.1  1993/12/17  00:07:33  foxen
 * initial revision.
 *
 * Revision 1.3  90/09/16  04:43:15  rearl
 * Preparation code added for disk-based MUCK.
 *
 * Revision 1.2  90/08/11  04:12:19  rearl
 * *** empty log message ***
 *
 * Revision 1.1  90/07/19  23:04:18  casie
 * Initial revision
 *
 *
 */

#include "copyright.h"
#include "config.h"

#include "db.h"
#include "tune.h"

/* remove the first occurence of what in list headed by first */
func remove_first(dbref first, dbref what) dbref {
	dbref prev;

	/* special case if it's the first one */
	if (first == what) {
		return db.Fetch(first).next
	} else {
		/* have to find it */

		for prev = first; prev != NOTHING; prev = db.Fetch(prev).next {
			if db.Fetch(prev).next == what {
				db.Fetch(prev).next = db.Fetch(what).next
				db.Fetch(prev).flags |= OBJECT_CHANGED
				return first
			}
		}
		return first
	}
}

func member(thing, list dbref) (r bool) {
	for ; !r && list != NOTHING; list = db.Fetch(list).next {
		r = list == thing || (db.Fetch(list).contents && member(thing, db.Fetch(list).contents))
	}
	return
}

func reverse(list dbref) (newlist dbref) {
	for newlist = NOTHING; list != NOTHING; {
		rest := db.Fetch(list).next
		db.Fetch(list).next = newlist
		db.Fetch(list).flags |= OBJECT_CHANGED
		newlist = list
		db.Fetch(newlist).flags |= OBJECT_CHANGED
		list = rest
	}
	return
}