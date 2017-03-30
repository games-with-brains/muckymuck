#define PRIMS_DB_FUNCS1 prim_addpennies, prim_moveto, prim_pennies,      \
    prim_dbcomp, prim_dbref, prim_contents, prim_exits, prim_next,       \
    prim_name, prim_setname, prim_match, prim_rmatch, prim_copyobj,      \
    prim_set, prim_mlevel, prim_flagp, prim_playerp, prim_thingp,        \
    prim_roomp, prim_programp, prim_exitp, prim_okp, prim_location,      \
    prim_owner, prim_getlink, prim_setlink, prim_setown, prim_newobject, \
    prim_newroom, prim_newexit, prim_lockedp, prim_recycle,              \
    prim_setlockstr, prim_getlockstr, prim_part_pmatch, prim_controls,   \
    prim_checkpassword, prim_nextowned, prim_getlinks,                   \
    prim_pmatch, prim_movepennies, prim_findnext, prim_nextentrance,     \
    prim_newplayer, prim_copyplayer, prim_instances,        \
    prim_compiledp, prim_newprogram, prim_contents_array,                \
    prim_exits_array, prim_getlinks_array, prim_entrances_array,         \
    prim_compile, prim_uncompile, prim_newpassword, prim_getpids,        \
    prim_program_getlines, prim_getpidinfo, prim_program_setlines,	 \
    prim_setlinks_array

#define PRIMS_DB_NAMES1 "ADDPENNIES", "MOVETO", "PENNIES", \
    "DBCMP", "DBREF", "CONTENTS", "EXITS", "NEXT",         \
    "NAME", "SETNAME", "MATCH", "RMATCH", "COPYOBJ",       \
    "SET", "MLEVEL", "FLAG?", "PLAYER?", "THING?",         \
    "ROOM?", "PROGRAM?", "EXIT?", "OK?", "LOCATION",       \
    "OWNER", "GETLINK", "SETLINK", "SETOWN", "NEWOBJECT",  \
    "NEWROOM", "NEWEXIT", "LOCKED?", "RECYCLE",            \
    "SETLOCKSTR", "GETLOCKSTR", "PART_PMATCH", "CONTROLS", \
    "CHECKPASSWORD", "NEXTOWNED", "GETLINKS",              \
    "PMATCH", "MOVEPENNIES", "FINDNEXT", "NEXTENTRANCE",   \
    "NEWPLAYER", "COPYPLAYER", "INSTANCES",      \
    "COMPILED?", "NEWPROGRAM", "CONTENTS_ARRAY",           \
    "EXITS_ARRAY", "GETLINKS_ARRAY", "ENTRANCES_ARRAY",    \
    "COMPILE", "UNCOMPILE", "NEWPASSWORD", "GETPIDS",      \
    "PROGRAM_GETLINES", "GETPIDINFO", "PROGRAM_SETLINES",  \
    "SETLINKS_ARRAY"

#define PRIMS_DB_CNT1 61


#ifdef SCARY_MUF_PRIMS

 /* These add dangerous, but possibly useful prims. */
# define PRIMS_DB_FUNCS PRIMS_DB_FUNCS1, prim_toadplayer
# define PRIMS_DB_NAMES PRIMS_DB_NAMES1, "TOADPLAYER"
# define PRIMS_DB_CNT (PRIMS_DB_CNT1 + 1)

#else
# define PRIMS_DB_FUNCS PRIMS_DB_FUNCS1
# define PRIMS_DB_NAMES PRIMS_DB_NAMES1
# define PRIMS_DB_CNT PRIMS_DB_CNT1
#endif