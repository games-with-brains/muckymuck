// prim_pop	/*     ? --              */
// prim_dup	/*     ? -- ? ?          */
// prim_at	/*   V|v -- ?            */
// prim_bang	/* ? V|v --              */
// prim_var	/*     i -- V            */
// prim_localvar	/*     i -- v            */
// prim_swap	/* ?1 ?2 -- ?2 ?1        */
// prim_over	/* ?1 ?1 -- ?1 ?2 ?1     */
// prim_pick	/*    ?n ... ?1 i -- ?n ... ?1 ?n */
// prim_put	/* ?n ... ?1 ?x i -- ?x ... ?1    */
// prim_rot	/*       ?1 ?2 ?3 -- ?2 ?3 ?1     */
// prim_rotate	/* Rotates top N stack elements */
// prim_dbtop	/*       -- i            */
// prim_depth	/*       -- i            */
// prim_version	/*       -- s            */
// prim_prog	/*       -- d            */
// prim_trig	/*       -- d            */
// prim_caller	/*       -- d            */
// prim_intp	/*     ? -- i            */
// prim_arrayp	/*     ? -- i            */
// prim_dictionaryp	/*     ? -- i            */
// prim_floatp	/*     ? -- i            */
// prim_stringp	/*     ? -- i            */
// prim_dbrefp	/*     ? -- i            */
// prim_addressp	/*     ? -- i            */
// prim_lockp	/*     ? -- i            */
// prim_checkargs	/*     s --              */
// prim_mode	/*       -- i            */
// prim_setmode	/*     i --              */
// prim_interp	/* d d s -- ?            */
// prim_reverse	/* ?n ... ?1 i -- ?1 ... ?n    */
// prim_lreverse	/* ?n ... ?1 i -- ?1 ... ?n i  */
// prim_dupn
// prim_ldup	/*   {?} -- {?} {?}      */
// prim_popn	/*   {?} --              */
// prim_for	/* i i i --              */
// prim_foreach	/*     i --              */

// prim_foriter	/*       -- i  or  @ ?   */
// prim_forpop	/*       --              */
// prim_mark	/*       -- m            */
// prim_findmark	/* m ?n ... ?1 -- ?n ... ?1 i    */
// prim_trypop	/* -- */


#define PRIMS_STACK_FUNCS prim_pop, prim_dup, prim_at, prim_bang, prim_var,  \
    prim_localvar, prim_swap, prim_over, prim_pick, prim_put, prim_rot,      \
    prim_rotate, prim_dbtop, prim_depth, prim_version, prim_prog, prim_trig, \
    prim_caller, prim_intp, prim_stringp, prim_dbrefp, prim_addressp,        \
    prim_lockp, prim_checkargs, prim_mode, prim_setmode, prim_interp,        \
    prim_for, prim_foreach, prim_floatp, prim_reverse, prim_popn, prim_dupn, \
    prim_ldup, prim_lreverse, prim_arrayp, prim_dictionaryp, prim_mark,      \
    prim_findmark

#define PRIMS_STACK_NAMES "POP", "DUP", "@", "!", "VARIABLE", \
    "LOCALVAR", "SWAP", "OVER", "PICK", "PUT", "ROT",         \
    "ROTATE", "DBTOP", "DEPTH", "VERSION", "PROG", "TRIG",    \
    "CALLER", "INT?", "STRING?", "DBREF?", "ADDRESS?",        \
    "LOCK?", "CHECKARGS", "MODE", "SETMODE", "INTERP",        \
    " FOR", " FOREACH", "FLOAT?", "REVERSE", "POPN", "DUPN",  \
    "LDUP", "LREVERSE", "ARRAY?", "DICTIONARY?", "{",         \
    "}"

#define PRIMS_STACK_CNT 39

#define PRIMS_INTERNAL_FUNCS prim_foriter, prim_forpop, prim_trypop

#define PRIMS_INTERNAL_NAMES " FORITER", " FORPOP", " TRYPOP"

#define PRIMS_INTERNAL_CNT 3