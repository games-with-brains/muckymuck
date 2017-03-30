/* $Header: /cvsroot/fbmuck/fbmuck/src/boolexp.c,v 1.12 2006/04/19 02:58:54 premchai21 Exp $ */

#include "copyright.h"
#include "config.h"

#include <ctype.h>
#include <stdio.h>

#include "fbstrings.h"
#include "db.h"
#include "props.h"
#include "match.h"
#include "externs.h"
#include "params.h"
#include "tune.h"
#include "interface.h"

/* Lachesis note on the routines in this package:
 *   eval_booexp does just evaluation.
 *
 *   parse_boolexp makes potentially recursive calls to several different
 *   subroutines ---
 *        parse_boolexp_F
 *            This routine does the leaf level parsing and the NOT.
 *        parse_boolexp_E
 *            This routine does the ORs.
 *        parse_boolexp_T
 *            This routine does the ANDs.
 *
 *   Because property expressions are leaf level expressions, I have only
 *   touched eval_boolexp_F, asking it to call my additional parse_boolprop()
 *   routine.
 */

func copy_bool(struct boolexp *old) (r *boolexp) {
	if old == TRUE_BOOLEXP {
		return TRUE_BOOLEXP
	}
	switch old.type {
	case BOOLEXP_AND:
		r = &boolexp{ type: old.type, sub1: copy_bool(old.sub1), sub2: copy_bool(old.sub2) }
	case BOOLEXP_OR:
		r = &boolexp{ type: old.type, sub1: copy_bool(old.sub1), sub2: copy_bool(old.sub2) }
	case BOOLEXP_NOT:
		r = &boolexp{ type: old.type, sub1: copy_bool(old.sub1) }
	case BOOLEXP_CONST:
		r = &boolexp{ type: old.type, thing: old.thing }
	case BOOLEXP_PROP:
		if old.prop_check == nil {
			r = nil
		} else {
			r = &boolexp{ type: old.type, prop_check: NewPropNode(old.prop_check.key) }
			SetPFlagsRaw(r.prop_check, PropFlagsRaw(old.prop_check))
			r.prop_check.data = old.prop_check.data
		}
	default:
		panic("copy_bool(): Error in boolexp !")
	}
	return
}

func eval_boolexp_rec(int descr, dbref player, struct boolexp *b, dbref thing) (r bool) {
	switch b.(type) {
	case TRUE_BOOLEXP:
		r = true
	case BOOLEXP_AND:
		r = eval_boolexp_rec(descr, player, b.sub1, thing) && eval_boolexp_rec(descr, player, b.sub2, thing)
	case BOOLEXP_OR:
		r = eval_boolexp_rec(descr, player, b.sub1, thing) || eval_boolexp_rec(descr, player, b.sub2, thing)
	case BOOLEXP_NOT:
		r = !eval_boolexp_rec(descr, player, b.sub1, thing)
	case BOOLEXP_CONST:
		if b.thing != NOTHING {
			if _, ok := b.thing.(TYPE_PROGRAM):
				var real_player dbref
				switch player.(type) {
				case TYPE_PLAYER, TYPE_THING:
					real_player = player
				default:
					real_player = db.Fetch(player).owner
				}
				if tmpfr := interp(descr, real_player, db.Fetch(player).location, b.thing, thing, PREEMPT, STD_HARDUID, 0); tmpfr != nil {
					r = interp_loop(real_player, b.thing, tmpfr, false) != nil
				}
			}
			r ||= b.thing == player || b.thing == db.Fetch(player).owner || member(b.thing, db.Fetch(player).contents) || b.thing == db.Fetch(player).location
		}
	case BOOLEXP_PROP:
		if v, ok := b.prop_check.data.(string); ok {
			r = contains_property(descr, player, player, b.prop_check.key, v, 0)
		}
	default:
		panic("eval_boolexp_rec(): bad type !");
	}
	return
}

func eval_boolexp(descr int, player dbref, b *boolexp, thing dbref) int {
	return eval_boolexp_rec(descr, player, copy_bool(b), thing)
}

/* If the parser returns TRUE_BOOLEXP, you lose */
/* TRUE_BOOLEXP cannot be typed in by the user; use @unlock instead */

/* F -> (E); F -> !F; F -> object identifier */
func parse_boolexp_F(descr int, parsebuf string, player dbref, dbloadp int) (r *boolexp) {
	char msg[BUFFER_LEN];

	parsebuf = strings.TrimSpace(parsebuf)
	switch parsebuf[0] {
	case '(':
		r = parse_boolexp_E(descr, parsebuf[1:], player, dbloadp)
		parsebuf = strings.TrimSpace(parsebuf)
		if r == TRUE_BOOLEXP || parsebuf[0] != ')' {
			r = TRUE_BOOLEXP
		}
	case NOT_TOKEN:
		if r = &boolexp{ type: BOOLEXP_NOT, sub1 = parse_boolexp_F(descr, parsebuf[1:], player, dbloadp) }; r.sub1 == TRUE_BOOLEXP {
			r = TRUE_BOOLEXP
		}
	default:
		/* must have hit an object ref */
		/* load the name into our buffer */
		var buf string
		for parsebuf != "" && parsebuf[0] != AND_TOKEN && parsebuf[0] != OR_TOKEN && parsebuf[0] != ')' {
			buf += parsebuf[0]
			parsebuf = parsebuf[1:]
		}
		/* strip trailing whitespace */
		buf = strings.TrimRightFunc(buf, unicode.IsSpace)
		if strings.Index(buf, PROP_DELIMITER) {
			r = parse_boolprop(buf)
		} else {
			r = &boolexp{ type: BOOLEXP_CONST }
			if !dbloadp {
				md := NewMatch(descr, player, buf, TYPE_THING)
				match_neighbor(&md)
				match_possession(&md)
				match_me(&md)
				match_here(&md)
				match_absolute(&md)
				match_registered(&md)
				match_player(&md)
				r.thing = match_result(&md)

				if r.thing == NOTHING {
					msg = fmt.Sprintf("I don't see %s here.", buf)
					notify(player, msg)
					r = TRUE_BOOLEXP
				} else if r.thing == AMBIGUOUS {
					msg = fmt.Sprintf("I don't know which %s you mean!", buf)
					notify(player, msg)
					r = TRUE_BOOLEXP
				}
			} else {
				if buf[0] != NUMBER_TOKEN || !unicode.IsNumber(buf[1]) {
					r = TRUE_BOOLEXP
				}
				r.thing = (dbref) strconv.Atoi(buf + 1)
				if r.thing < 0 || r.thing >= db_top {
					r = TRUE_BOOLEXP
			}
		}
	}
	return
}

/* T -> F; T -> F & T */
func parse_boolexp_T(descr int, parsebuf string, player dbref, dbloadp int) (r *boolexp) {
	if r = parse_boolexp_F(descr, parsebuf, player, dbloadp); r != TRUE_BOOLEXP {
		parsebuf = strings.TrimSpace(parsebuf)
		if **parsebuf == AND_TOKEN {
			if r = &boolexp{ type: BOOLEXP_AND, sub1: r, sub2: parse_boolexp_T(descr, parsebuf[1:], player, dbloadp); r.sub2 == TRUE_BOOLEXP {
				r = TRUE_BOOLEXP
			}
		}
	}
}

/* E -> T; E -> T | E */
func parse_boolexp_E(descr int, parsebuf string, player dbref, dbloadp int) (r *boolexp) {
	if r = parse_boolexp_T(descr, parsebuf, player, dbloadp); r != TRUE_BOOLEXP {
		if parsebuf = strings.TrimSpace(parsebuf); parsebuf[0] == OR_TOKEN {
			if r = &boolexp{ type: BOOLEXP_OR, sub1: r, sub2: parse_boolexp_E(descr, parsebuf[1:], player, dbloadp) }; r.sub2 == TRUE_BOOLEXP {
				r = TRUE_BOOLEXP
			}
		}
	}
	return
}

func parse_boolexp(descr int, player dbref, buf string, dbloadp int) *boolexp {
	return parse_boolexp_E(descr, &buf, player, dbloadp)
}

/* parse a property expression
   If this gets changed, please also remember to modify set.c       */
func parse_boolprop(buf string) (r *boolexp) {
	char *type = buf
	char *strval = (char *) strchr(type, PROP_DELIMITER);
	char *x;
	char *temp;

	x = type;
	r = &boolexp{ type: BOOLEXP_PROP, thing: NOTHING }
	type = strings.TrimSpace(type)
	if *type == PROP_DELIMITER {
		/* Oops!  Clean up and return a TRUE */
		return TRUE_BOOLEXP
	}
	strval++;
	while (unicode.IsSpace(*strval) && *strval)
		strval++;
	if (!*strval) {
		/* Oops!  CLEAN UP AND RETURN A TRUE */
		return TRUE_BOOLEXP
	}
	/* get rid of trailing spaces */
	for (temp = strval; !unicode.IsSpace(*temp) && *temp; temp++) ;
	*temp = '\0';

	p := NewPropNode(type)
	r.prop_check = p
	p.data = strval
	return
}

func negate_boolexp(struct boolexp *b) (r *boolexp) {
	/* Obscure fact: !NOTHING == NOTHING in old-format databases! */
	if b == TRUE_BOOLEXP {
		r = TRUE_BOOLEXP
	} else {
		r = &boolexp{ type: BOOLEXP_NOT, sub1: b }
	}
	return
}

func getboolexp1(f *FILE) (b *boolexp) {
	char buf[BUFFER_LEN];		/* holds string for reading in property */
	int i;						/* index into buf */

	c := getc(f)
	switch (c) {
	case EOF:
		panic("getboolexp1(): unexpected EOF in boolexp !");

	case '\n':
		ungetc(c, f);
		return TRUE_BOOLEXP;

	case '(':
		b = new(boolexp)
		if c = getc(f); c == '!' {
			b.type = BOOLEXP_NOT
			b.sub1 = getboolexp1(f)
			if getc(f) != ')' {
				goto error
			}
			return b
		} else {
			ungetc(c, f)
			b.sub1 = getboolexp1(f)
			switch c = getc(f); c {
			case AND_TOKEN:
				b.type = BOOLEXP_AND
			case OR_TOKEN:
				b.type = BOOLEXP_OR
			default:
				goto error;
			}
			b.sub2 = getboolexp1(f)
			if getc(f) != ')' {
				goto error
			}
			return b
		}

	case '-':
		/* obsolete NOTHING key */
		/* eat it */
		while ((c = getc(f)) != '\n') {
			if (c == EOF) {
				panic("getboolexp1(): unexpected EOF in boolexp !");
			}
		}
		ungetc(c, f);
		return TRUE_BOOLEXP;

	case '[':
		/* property type */
		b = &boolexp{ type: BOOLEXP_PROP }
		i = 0;
		while ((c = getc(f)) != PROP_DELIMITER && i < BUFFER_LEN) {
			buf[i] = c;
			i++;
		}
		if (i >= BUFFER_LEN && c != PROP_DELIMITER)
			goto error;
		buf[i] = '\0';

		b->prop_check = NewPropNode(buf)
		p := b->prop_check

		i = 0;
		while ((c = getc(f)) != ']') {
			if (c == '\\')
				c = getc(f);
			buf[i] = c;
			i++;
		}
		buf[i] = '\0';
		if (i >= BUFFER_LEN && c != ']')
			goto error;
		if !unicode.IsNumber(buf) {
			p.data = buf
		} else {
			p.data = strconv.Atol(buf)
		}
		return b;

	default:
		/* better be a dbref */
		ungetc(c, f);
		b = &boolexp{ type: BOOLEXP_CONST }

		/* NOTE possibly non-portable code */
		/* Will need to be changed if putref/getref change */
		while (isdigit(c = getc(f))) {
			b->thing = b->thing * 10 + c - '0';
		}
		ungetc(c, f);
		return b;
	}

  error:
	panic("getboolexp1(): error in boolexp !"); /* bomb out */
	return NULL;
}

func getboolexp(f *FILE) (b *boolexp) {
	b = getboolexp1(f)
	if getc(f) != '\n' {
		panic("getboolexp(): parse error !")
	}
	return
}