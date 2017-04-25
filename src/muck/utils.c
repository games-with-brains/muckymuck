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

#include "DB.h"
#include "tune.h"

/* remove the first occurence of what in list headed by first */
func remove_first(ObjectID first, ObjectID what) ObjectID {
	ObjectID prev;

	/* special case if it's the first one */
	if (first == what) {
		return DB.Fetch(first).next
	} else {
		/* have to find it */

		for prev = first; prev != NOTHING; prev = DB.Fetch(prev).next {
			if DB.Fetch(prev).next == what {
				DB.Fetch(prev).next = DB.Fetch(what).next
				DB.Fetch(prev).Touch()
				return first
			}
		}
		return first
	}
}

func member(thing, list ObjectID) (r bool) {
	for ; !r && list != NOTHING; list = DB.Fetch(list).next {
		r = list == thing || (DB.Fetch(list).Contents && member(thing, DB.Fetch(list).Contents))
	}
	return
}

func reverse(list ObjectID) (newlist ObjectID) {
	for newlist = NOTHING; list != NOTHING; {
		rest := DB.Fetch(list).next
		DB.Fetch(list).next = newlist
		DB.Fetch(list).Touch()
		newlist = list
		DB.Fetch(newlist).Touch()
		list = rest
	}
	return
}