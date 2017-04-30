/*
 * config.h
 *
 * Tunable parameters -- Edit to you heart's content 
 *
 * Parameters that control system behavior, and tell the system
 * what resources are available (most of this is now done by
 * the configure script).
 *
 * Most of the goodies that used to be here are now in @tune.
 */

#ifndef _CONFIG_H
#define _CONFIG_H

#include "copyright.h"
#include "autoconf.h"

/************************************************************************
   Administrative Options 

   Various things that affect how the muck operates and what privs
 are compiled in.
 ************************************************************************/

/* Detaches the process as a daemon so that it don't cause problems
 * keeping a terminal line open and such. Logs normal output to a file
 * and writes out a netmuck.pid file 
 */
#define DETACH

/*
 * Port where tinymuck lives -- Note: If you use a port lower than
 * 1024, then the program *must* run suid to root!
 * Port 4201 is a port with historical significance, as the original
 * TinyMUD Classic ran on that port.  It was the office number of James
 * Aspnes, who wrote TinyMUD from which TinyMUCK eventually was derived.
 */
#define TINYPORT 4201			/* Port that players connect to */

/*
 * Some systems can hang for up to 30 seconds while trying to resolve
 * hostnames.  Define this to use a non-blocking second process to resolve
 * hostnames for you.  NOTE:  You need to compile the 'fb-resolver' program
 * (make resolver) and put it in the directory that the fbmuck program is
 * run from.
 */
#define SPAWN_HOST_RESOLVER

/*
 * There's a set of MUF prims that are considered dangerous.
 * Currently these include only:
 *     TOADPLAYER which can @toad players.
 * These require wizbits, but if someone finds a new way to hack the
 * system to get wizbit perms, these can cause damage that may require
 * a reversion to a previous database to fix.
 * Define this, to make these MUF prims available.
 */
#undef SCARY_MUF_PRIMS

/*
 * This is a fairly interesting one -- if there's no DISKBASING, and thus
 * the saves are fork()ed off into the background, then then the child
 * may cause I/O contention with the parent (the interactive, player-
 * connected process).  If this occurs, you can set this to a number
 * greater than 0 to make it be slightly nicer to the rest of the
 * system.  (Usually, setting it to 1 is all that's needed.)
 */
#define NICEVAL 1

/*
 * This allows to use the SIGUSR2 signal as a request for an immediate clean
 * shutdown of the MUCK. Such a signal may be used, for instance, when the UPS
 * is going to fail and shuts down the system. In this case, sending SIGUSR2
 * to the MUCK server process prints a special message on the logged users'
 * screen ("Emergency signal received ! (power failure ?) The database will be
 * saved."), and saves cleanly the database before shutting down the MUCK.
 */
#define SIGEMERG

/************************************************************************
   Game Options

   These are the ones players will notice. 
 ************************************************************************/

/* Make the `examine' command display full names for types and flags */
#define VERBOSE_EXAMINE

//	FIX: PLAYER_NAME_LIMIT abolished - needs to be removed from display formatting code
#define PLAYER_NAME_LIMIT 16

/************************************************************************
   Various Messages 
 
   Printed from the server at times, esp. during login.
 ************************************************************************/

/*
 * Welcome message if you don't have a welcome.txt
 */
#define DEFAULT_WELCOME_MESSAGE "Welcome to TinyMUCK.\r\nTo connect to your existing character, enter \"connect name password\"\r\nTo create a new character, enter \"create name password\"\r\nIMPORTANT! Use the news command to get up-to-date news on program changes.\r\n\r\nYou can disconnect using the QUIT command, which must be capitalized as shown.\r\n\r\nAbusing or harrassing other players will result in expellation.\r\nUse the WHO command to find out who is currently active.\r\n"

/*
 * Error messeges spewed by the help system.
 */
#define NO_NEWS_MSG "That topic does not exist.  Type 'news topics' to list the news topics available."
#define NO_HELP_MSG "That topic does not exist.  Type 'help index' to list the help topics available."
#define NO_MAN_MSG "That topic does not exist.  Type 'man' to list the MUF topics available."
#define NO_INFO_MSG "That file does not exist.  Type 'info' to get a list of the info files available."

/************************************************************************
   File locations
 
   Where the system looks for its datafiles.
 ************************************************************************/

#define SSL_CERT_FILE		"data/server.pem"
#define SSL_KEY_FILE		"data/server.pem"

#define WELC_FILE "data/welcome.txt"	/* For the opening screen      */
#define MOTD_FILE "data/motd.txt"	/* For the message of the day  */

#define HELP_FILE "data/help.txt"	/* For the 'help' command      */
#define HELP_DIR  "data/help"	/* For 'help' subtopic files   */
#define NEWS_FILE "data/news.txt"	/* For the 'news' command      */
#define NEWS_DIR  "data/news"	/* For 'news' subtopic files   */
#define MAN_FILE  "data/man.txt"	/* For the 'man' command       */
#define MAN_DIR   "data/man"	/* For 'man' subtopic files    */
#define MPI_FILE  "data/mpihelp.txt"	/* For the 'mpi' command       */
#define MPI_DIR   "data/mpihelp"	/* For 'mpi' subtopic files    */
#define INFO_DIR  "data/info/"
#define EDITOR_HELP_FILE "data/edit-help.txt"	/* editor help file   */

#define DELTAFILE_NAME "data/deltas-file"	/* The file for deltas */
#define PARMFILE_NAME "data/parmfile.cfg"	/* The file for config parms */
#define WORDLIST_FILE "data/wordlist.txt"	/* File for compression dict. */

#define LOG_CMD_TIMES "logs/cmd-times"	/* Command times Log */
#define LOG_GRIPE   "logs/gripes"	/* Gripes Log */
#define LOG_STATUS  "logs/status"	/* System errors and stats */
#define LOG_SANITY  "logs/sanity"	/* Database corruption and errors */
#define LOG_MUF     "logs/muf-errors"	/* Muf compiler errors and warnings. */
#define COMMAND_LOG "logs/commands"	/* Player commands */
#define PROGRAM_LOG "logs/programs"	/* text of changed programs */
#define USER_LOG    "logs/user"		/* log of player/program-init msgs. */

#define MACRO_FILE  "muf/macros"

/*
 * File to write the server pid to.  This is generally "netmuck.pid"
 * for historical reasons.
 */
#define PID_FILE    "netmuck.pid"

#define RESOLVER_PID_FILE "hostfind.pid"	/* Write the resolver pid to ... */

#ifdef DETACH
# define LOG_FILE     "logs/fbmuck"     /* Log stdout to ... */
# define LOG_ERR_FILE "logs/fbmuck.err" /* Log stderr to ... */
#endif							/* DETACH */

#define TRUE  1
#define FALSE 0

#endif /* _CONFIG_H */