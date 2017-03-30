#define DB_VERSION_STRING "***Foxen9 TinyMUCK DUMP Format***"

/* return masks from db_identify */
#define DB_ID_VERSIONSTRING	0x00000001 /* has a returned **version */
#define DB_ID_DELTAS		0x00000002 /* doing a delta file */

#define DB_ID_GROW 		0x00000010 /* grow parameter will be set */
#define DB_ID_PARMSINFO		0x00000020 /* parmcnt set, need to do a tune_load_parms_from_file */
#define DB_ID_OLDCOMPRESS	0x00000040 /* whether it COULD be */
#define DB_ID_CATCOMPRESS	0x00000080 /* whether it COULD be */