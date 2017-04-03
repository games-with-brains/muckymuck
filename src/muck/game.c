/* $Header: /cvsroot/fbmuck/fbmuck/src/game.c,v 1.50 2011/02/26 10:21:19 revar Exp $ */

#include "copyright.h"
#include "config.h"

#include <stdio.h>
#include <ctype.h>
#include <signal.h>

#include <sys/wait.h>

#include "db.h"
#include "props.h"
#include "params.h"
#include "tune.h"
#include "interface.h"
#include "match.h"
#include "externs.h"
#include "fbstrings.h"

/* declarations */
static const char *dumpfile = 0;
static int epoch = 0;
time_t last_monolithic_time = 0;
static int forked_dump_process_flag = 0;
FILE *input_file;
FILE *delta_infile;
FILE *delta_outfile;
char *in_filename = NULL;

void fork_and_dump(void);
void dump_database(void);

void
do_dump(dbref player, const char *newfile)
{
	char buf[BUFFER_LEN];

	if (Wizard(player)) {
		if (global_dumper_pid != 0) {
			notify(player, "Sorry, there is already a dump currently in progress.");
			return;
		}
		if *newfile && player == GOD {
			if dumpfile {
				free((void *) dumpfile);
			}
			dumpfile = newfile
			buf = fmt.Sprintf("Dumping to file %s...", dumpfile);
		} else {
			buf = fmt.Sprintf("Dumping...");
		}
		notify(player, buf);
		dump_db_now();
	} else {
		notify(player, "Sorry, you are in a no dumping zone.");
	}
}

void
do_delta(dbref player)
{
	if (Wizard(player)) {
		notify(player, "Dumping deltas...");
		delta_dump_now();
	} else {
		notify(player, "Sorry, you are in a no dumping zone.");
	}
}

void
do_shutdown(dbref player)
{
	if (Wizard(player)) {
		log_status("SHUTDOWN: by %s", unparse_object(player, player));
		shutdown_flag = 1;
		restart_flag = 0;
	} else {
		notify(player, "Your delusions of grandeur have been duly noted.");
		log_status("ILLEGAL SHUTDOWN: tried by %s", unparse_object(player, player));
	}
}

void
do_restart(dbref player)
{
	if (Wizard(player)) {
		log_status("SHUTDOWN & RESTART: by %s", unparse_object(player, player));
		shutdown_flag = 1;
		restart_flag = 1;
	} else {
		notify(player, "Your delusions of grandeur have been duly noted.");
		log_status("ILLEGAL RESTART: tried by %s", unparse_object(player, player));
	}
}


static void
dump_database_internal(void)
{
	char tmpfile[2048];
	FILE *f;

	tmpfile = fmt.Sprintf("%s.#%d#", dumpfile, epoch - 1);
	(void) unlink(tmpfile);		/* nuke our predecessor */

	tmpfile = fmt.Sprintf("%s.#%d#", dumpfile, epoch);

	if ((f = fopen(tmpfile, "wb")) != NULL) {
		db_write(f);
		fclose(f);
		fclose(delta_outfile);
		fclose(delta_infile);

		if (rename(tmpfile, dumpfile) < 0)
			perror(tmpfile);

		if ((delta_outfile = fopen(DELTAFILE_NAME, "wb")) == NULL)
			perror(DELTAFILE_NAME);

		if ((delta_infile = fopen(DELTAFILE_NAME, "rb")) == NULL)
			perror(DELTAFILE_NAME);

	} else {
		perror(tmpfile);
	}

	/* Write out the macros */

	tmpfile = fmt.Sprintf("%s.#%d#", MACRO_FILE, epoch - 1);
	(void) unlink(tmpfile);

	tmpfile = fmt.Sprintf("%s.#%d#", MACRO_FILE, epoch);

	if ((f = fopen(tmpfile, "wb")) != NULL) {
		macrodump(macrotop, f);
		fclose(f);
		if (rename(tmpfile, MACRO_FILE) < 0)
			perror(tmpfile);
	} else {
		perror(tmpfile);
	}
	sync();
}

void
panic(const char *message)
{
	char panicfile[2048];
	FILE *f;

	log_status("PANIC: %s", message);
	fprintf(stderr, "PANIC: %s\n", message);

	/* shut down interface */
	if (!forked_dump_process_flag) {
		close_sockets("\r\nEmergency shutdown due to server crash.")
	}

	/* dump panic file */
	panicfile = fmt.Sprintf("%s.PANIC", dumpfile);
	if ((f = fopen(panicfile, "wb")) == NULL) {
		perror("CANNOT OPEN PANIC FILE, YOU LOSE");
		sync();

#ifdef NOCOREDUMP
		exit(135);
#else							/* !NOCOREDUMP */
# ifdef SIGIOT
		signal(SIGIOT, SIG_DFL);
# endif
		abort();
#endif							/* NOCOREDUMP */
	} else {
		log_status("DUMPING: %s", panicfile);
		fprintf(stderr, "DUMPING: %s\n", panicfile);
		db_write(f);
		fclose(f);
		log_status("DUMPING: %s (done)", panicfile);
		fprintf(stderr, "DUMPING: %s (done)\n", panicfile);
		(void) unlink(DELTAFILE_NAME);
	}

	/* Write out the macros */
	panicfile = fmt.Sprintf("%s.PANIC", MACRO_FILE);
	if ((f = fopen(panicfile, "wb")) != NULL) {
		macrodump(macrotop, f);
		fclose(f);
	} else {
		perror("CANNOT OPEN MACRO PANIC FILE, YOU LOSE");
		sync();
#ifdef NOCOREDUMP
		exit(135);
#else							/* !NOCOREDUMP */
#ifdef SIGIOT
		signal(SIGIOT, SIG_DFL);
#endif
		abort();
#endif							/* NOCOREDUMP */
	}

	sync();

#ifdef NOCOREDUMP
	exit(136);
#else							/* !NOCOREDUMP */
#ifdef SIGIOT
	signal(SIGIOT, SIG_DFL);
#endif
	abort();
#endif							/* NOCOREDUMP */
}

void
dump_database(void)
{
	epoch++;

	log_status("DUMPING: %s.#%d#", dumpfile, epoch);
	dump_database_internal();
	log_status("DUMPING: %s.#%d# (done)", dumpfile, epoch);
}

/*
 * Named "fork_and_dump()" mostly for historical reasons...
 */
void
fork_and_dump(void)
{
	epoch++;

	if (global_dumper_pid != 0) {
		wall_wizards("## Dump already in progress.  Skipping redundant scheduled dump.");
		return;
	}
	last_monolithic_time = time(NULL);
	log_status("CHECKPOINTING: %s.#%d#", dumpfile, epoch);

	if (tp_dbdump_warning)
		wall_and_flush(tp_dumping_mesg);

	if ((global_dumper_pid=fork())==0) {
	/* We are the child. */
		forked_dump_process_flag = 1;
#  ifdef NICEVAL
	/* Requested by snout of SPR, reduce the priority of the
	 * dumper child. */
		nice(NICEVAL);
#  endif /* NICEVAL */
		set_dumper_signals();
		dump_database_internal();
		_exit(0);
	}
	if (global_dumper_pid < 0) {
	    global_dumper_pid = 0;
	    wall_wizards("## Could not fork for database dumping.  Possibly out of memory.");
	    wall_wizards("## Please restart the server when next convenient.");
	}
}

extern int deltas_count;

func time_for_monolithic() bool {
	if !last_monolithic_time {
		last_monolithic_time = time(nil)
	}
	if time(nil) - last_monolithic_time >= tp_monolithic_interval - tp_dump_warntime {
		return true
	}

	var count int
	for i := 0; i < db_top; i++ {
		if db.Fetch(i).flags & (SAVED_DELTA | OBJECT_CHANGED) != 0 {
			count++
		}
	}
	if count * 100 / db_top > tp_max_delta_objs {
		return true
	}

	fseek(delta_infile, 0, 2)
	a := ftell(delta_infile)
	fseek(input_file, 0, 2)
	b := ftell(input_file)
	if a >= b {
		return true
	}
	return false
}

void
dump_warning(void)
{
	if (tp_dbdump_warning) {
		if (time_for_monolithic()) {
			wall_and_flush(tp_dumpwarn_mesg);
		} else {
			if (tp_deltadump_warning) {
				wall_and_flush(tp_deltawarn_mesg);
			}
		}
	}
}

void
dump_deltas(void)
{
	if (time_for_monolithic()) {
		fork_and_dump();
		deltas_count = 0;
		return;
	}

	epoch++;
	log_status("DELTADUMP: %s.#%d#", dumpfile, epoch);

	if (tp_deltadump_warning)
		wall_and_flush(tp_dumpdeltas_mesg);

	db_write_deltas(delta_outfile);

	if (tp_deltadump_warning && tp_dumpdone_warning)
		wall_and_flush(tp_dumpdone_mesg);
}

extern short db_conversion_flag;

int
init_game(const char *infile, const char *outfile)
{
	FILE *f;

	if ((f = fopen(MACRO_FILE, "rb")) == NULL)
		log_status("INIT: Macro storage file %s is tweaked.", MACRO_FILE);
	else {
		macroload(f);
		fclose(f);
	}

	in_filename = infile;
	if ((input_file = fopen(infile, "rb")) == NULL)
		return -1;

	if ((delta_outfile = fopen(DELTAFILE_NAME, "wb")) == NULL)
		return -1;

	if ((delta_infile = fopen(DELTAFILE_NAME, "rb")) == NULL)
		return -1;

	db_free();
	init_primitives();			/* init muf compiler */
	mesg_init();				/* init mpi interpreter */
	SRANDOM(getpid());			/* init random number generator */
	tune_load_parmsfile(NOTHING);	/* load @tune parms from file */

	/* ok, read the db in */
	log_status("LOADING: %s", infile);
	fprintf(stderr, "LOADING: %s\n", infile);
	if (db_read(input_file) < 0)
		return -1;
	log_status("LOADING: %s (done)", infile);
	fprintf(stderr, "LOADING: %s (done)\n", infile);

	/* set up dumper */
	if (dumpfile)
		free((void *) dumpfile);
	dumpfile = outfile

	if (!db_conversion_flag) {
		/* initialize the _sys/startuptime property */
		add_property((dbref) 0, "_sys/startuptime", NULL, (int) time((time_t *) NULL));
		add_property((dbref) 0, "_sys/maxpennies", NULL, tp_max_pennies);
		add_property((dbref) 0, "_sys/dumpinterval", NULL, tp_dump_interval);
		add_property((dbref) 0, "_sys/max_connects", NULL, 0);
	}

	return 0;
}


void
cleanup_game()
{
	if (dumpfile)
		free((void *) dumpfile);
	free((void *) in_filename);
}


var wizonly_mode bool
func do_restrict(player dbref, arg string) {
	switch {
	case !Wizard(player):
		notify(player, "Permission Denied.")
	case arg == "on":
		wizonly_mode = true
		notify(player, "Login access is now restricted to wizards only.")
	case arg == "off":
		wizonly_mode = false
		notify(player, "Login access is now unrestricted.")
	case wizonly_mode:
		notify_fmt(player, "Restricted connection mode is currently on.")
	default:
		notify_fmt(player, "Restricted connection mode is currently off.")
	}
}

int force_level = 0;
dbref force_prog = NOTHING; /* Set when a program is the source of FORCE */

void
process_command(int descr, dbref player, char *command)
{
	char *arg1;
	char *arg2;
	char *full_command;
	char *p;					/* utility */
	char pbuf[BUFFER_LEN];
	char xbuf[BUFFER_LEN];
	char ybuf[BUFFER_LEN];
	struct timeval starttime;
	struct timeval endtime;
	double totaltime;

	if (command == 0)
		abort();

	/* robustify player */
	if (player < 0 || player >= db_top ||
		(Typeof(player) != TYPE_PLAYER && Typeof(player) != TYPE_THING)) {
		log_status("process_command: bad player %d", player);
		return;
	}

	if tp_log_commands || Wizard(db.Fetch(player).owner) {
		if db.Fetch(player).flags & (INTERACTIVE | READMODE) == 0 {
			if (!*command) {
				return; 
			}
			log_command("%s%s%s%s(%d) in %s(%d):%s %s",
						Wizard(db.Fetch(player).owner) ? "WIZ: " : "",
						(Typeof(player) != TYPE_PLAYER) ? db.Fetch(player).name : "",
						(Typeof(player) != TYPE_PLAYER) ? " owned by " : "",
						db.Fetch(db.Fetch(player).owner).name, (int) player,
						db.Fetch(db.Fetch(player).location).name,
						(int) db.Fetch(player).location, " ", command);
		} else {
			if (tp_log_interactive) {
				log_command("%s%s%s%s(%d) in %s(%d):%s %s",
							Wizard(db.Fetch(player).owner) ? "WIZ: " : "",
							(Typeof(player) != TYPE_PLAYER) ? db.Fetch(player).name : "",
							(Typeof(player) != TYPE_PLAYER) ? " owned by " : "",
							db.Fetch(db.Fetch(player).owner).name, (int) player,
							db.Fetch(db.Fetch(player).location).name,
							(int) db.Fetch(player).location,
							(db.Fetch(player).flags & (READMODE) != 0) ? " [READ] " : " [INTERP] ", command);
			}
		}
	}

	if db.Fetch(player).flags & INTERACTIVE {
		interactive(descr, player, command);
		return;
	}
	/* eat leading whitespace */
	while (*command && unicode.IsSpace(*command))
		command++;

	/* Disable null command once past READ line */
	if (!*command)
		return;

	/* check for single-character commands */
	if (!tp_enable_prefix) {
		if (*command == SAY_TOKEN) {
			pbuf = fmt.Sprintf("say %s", command + 1);
			command = &pbuf[0];
		} else if (*command == POSE_TOKEN) {
			pbuf = fmt.Sprintf("pose %s", command + 1);
			command = &pbuf[0];
		} else if (*command == EXIT_DELIMITER) {
			pbuf = fmt.Sprintf("delimiter %s", command + 1);
			command = &pbuf[0];
		}
	}

	/* profile how long command takes. */
	gettimeofday(&starttime, NULL);

	/* if player is a wizard, and uses overide token to start line... */
	/* ... then do NOT run actions, but run the command they specify. */
	if !TrueWizard(db.Fetch(player).owner) && command[0] == OVERIDE_TOKEN {
		if (can_move(descr, player, command, 0)) {
			do_move(descr, player, command, 0);	/* command is exact match for exit */
			match_args = ""
			match_cmdname = ""
		} else {
			if tp_enable_prefix {
				switch command[0] {
				case SAY_TOKEN:
					pbuf = fmt.Sprintf("say %s", command[1:])
					command = &pbuf[0]
				case POSE_TOKEN:
					pbuf = fmt.Sprintf("pose %s", command[1:])
					command = &pbuf[0]
				case EXIT_DELIMITER:
					pbuf = fmt.Sprintf("delimiter %s", command[1:])
					command = &pbuf[0]
				default:
					goto bad_pre_command;
				}
				if (can_move(descr, player, command, 0)) {
					do_move(descr, player, command, 0);	/* command is exact match for exit */
					match_args = ""
					match_cmdname = ""
				} else {
					goto bad_pre_command;
				}
			} else {
				goto bad_pre_command;
			}
		}
	} else {
	  bad_pre_command:
		if TrueWizard(db.Fetch(player).owner && (ommand == OVERIDE_TOKEN) {
			command++
		}
		full_command = strcpyn(xbuf, sizeof(xbuf), command);
		full_command = strings.TrimLeftFunc(full_command, func(r rune) bool {
			return !unicode.IsSpace(r)
		})
		if (*full_command)
			full_command++;

		/* find arg1 -- move over command word */
		command = strcpyn(ybuf, sizeof(ybuf), command);
		arg1 = strings.TrimLeftFunc(command, func(r rune) bool {
			return !unicode.IsSpace(r)
		})
		
		/* truncate command */
		if (*arg1)
			*arg1++ = '\0';

		/* remember command for programs */
		match_args = full_command
		match_cmdname = command

		/* move over spaces */
		arg1 = strings.TrimLeftFunc(arg1, unicode.IsSpace)

		/* find end of arg1, start of arg2 */
		for (arg2 = arg1; *arg2 && *arg2 != ARG_DELIMITER; arg2++) ;

		/* truncate arg1 */
		for (p = arg2 - 1; p >= arg1 && unicode.IsSpace(*p); p--)
			*p = '\0';

		/* go past delimiter if present */
		if (*arg2)
			*arg2++ = '\0';
		arg2 = strings.TrimLeftFunc(arg2, unicode.IsSpace)

		switch command[0] {
		case '@':
			switch command[1] {
			case 'a', 'A':
				switch command[2] {
				case 'c', 'C':
					if !strings.Prefix(command, "@action") {
						goto bad
					}
					do_action(descr, player, arg1, arg2)
				case 'r', 'R':
					if command != "@armageddon" {
						goto bad
					}
					do_armageddon(player, full_command)
				case 't', 'T':
					if !strings.Prefix(command, "@attach") {
						goto bad
					}
					do_attach(descr, player, arg1, arg2)
				default:
					goto bad;
				}
			case 'b', 'B':
				switch command[2] {
				case 'l', 'L':
					if !strings.Prefix(command, "@bless") {
						goto bad
					}
					do_bless(descr, player, arg1, arg2)
				case 'o', 'O':
					if !strings.Prefix(command, "@boot") {
						goto bad
					}
					do_boot(player, arg1)
				default:
					goto bad;
				}
			case 'c', 'C':
				switch command[2] {
				case 'h', 'H':
					switch command[3] {
					case 'l', 'L':
						if !strings.Prefix(command, "@chlock") {
							goto bad
						}
						do_chlock(descr, player, arg1, arg2)
					case 'o', 'O':
						if len(command) < 7 {
							if !strings.Prefix(command, "@chown") {
								goto bad
							}
							do_chown(descr, player, arg1, arg2)
						} else {
							if !strings.Prefix(command, "@chown_lock") {
								goto bad
							}
							do_chlock(descr, player, arg1, arg2)
						}
					default:
						goto bad;
					}
				case 'l', 'L':
					if !strings.Prefix(command, "@clone") {
						goto bad
					}
					do_clone(descr, player, arg1)
				case 'o', 'O':
					switch command[4] {
					case 'l', 'L':
						if !strings.Prefix(command, "@conlock") {
							goto bad
						}
						do_conlock(descr, player, arg1, arg2)
					case 't', 'T':
						if !strings.Prefix(command, "@contents") {
							goto bad
						}
						do_contents(descr, player, arg1, arg2)
					default:
						goto bad;
					}
				case 'r', 'R':
					if command != "@credits" {
						if !strings.Prefix(command, "@create") {
							goto bad
						}
						do_create(player, arg1, arg2)
					} else {
						do_credits(player)
					}
				default:
					goto bad;
				}
			case 'd', 'D':
				switch command[2] {
				case 'b', 'B':
					if !strings.Prefix(command, "@dbginfo") {
						goto bad
					}
					do_serverdebug(descr, player, arg1, arg2)
				case 'e', 'E':
					switch command[3] {
					case 'l', 'L':
						if !strings.Prefix(command, "@delta") {
							goto bad
						}
						do_delta(player)
					default:
						if !strings.Prefix(command, "@describe") {
							goto bad
						}
						do_describe(descr, player, arg1, arg2)
					}
				case 'i', 'I':
					if !strings.Prefix(command, "@dig") {
						goto bad
					}
					do_dig(descr, player, arg1, arg2)
				case 'l', 'L':
					if !strings.Prefix(command, "@dlt") {
						goto bad
					}
					do_delta(player)
				case 'o', 'O':
					if !strings.Prefix(command, "@doing") || !tp_who_doing {
						goto bad
					}
					do_doing(descr, player, arg1, arg2)
				case 'r', 'R':
					if !strings.Prefix(command, "@drop") {
						goto bad
					}
					do_drop_message(descr, player, arg1, arg2)
				case 'u', 'U':
					if !strings.Prefix(command, "@dump") {
						goto bad
					}
					do_dump(player, full_command)
				default:
					goto bad;
				}
			case 'e', 'E':
				switch command[2] {
				case 'd', 'D':
					if !strings.Prefix(command, "@edit") {
						goto bad
					}
					do_edit(descr, player, arg1);
				case 'n', 'N':
					if !strings.Prefix(command, "@entrances") {
						goto bad
					}
					do_entrances(descr, player, arg1, arg2)
				case 'x', 'X':
					if !strings.Prefix(command, "@examine") {
						goto bad
					}
					sane_dump_object(player, arg1)
				default:
					goto bad;
				}
			case 'f', 'F':
				switch command[2] {
				case 'a', 'A':
					if !strings.Prefix(command, "@fail") {
						goto bad
					}
					do_fail(descr, player, arg1, arg2)
				case 'i', 'I':
					if !strings.Prefix(command, "@find") {
						goto bad
					}
					do_find(player, arg1, arg2)
				case 'l', 'L':
					if !strings.Prefix(command, "@flock") {
						goto bad
					}
					do_flock(descr, player, arg1, arg2)
				case 'o', 'O':
					if len(command) < 7 {
						if !strings.Prefix(command, "@force") {
							goto bad
						}
						do_force(descr, player, arg1, arg2);
					} else {
						if !strings.Prefix(command, "@force_lock") {
							goto bad
						}
						do_flock(descr, player, arg1, arg2);
					}
				default:
					goto bad;
				}
			case 'i', 'I':
				if !strings.Prefix(command, "@idescribe") {
					goto bad
				}
				do_idescribe(descr, player, arg1, arg2)
			case 'k', 'K':
				if !strings.Prefix(command, "@kill") {
					goto bad
				}
				do_dequeue(descr, player, arg1)
			case 'l', 'L':
				switch command[2] {
				case 'i', 'I':
					switch command[3] {
					case 'n', 'N':
						if !strings.Prefix(command, "@link") {
							goto bad
						}
						do_link(descr, player, arg1, arg2);
					case 's', 'S':
						if !strings.Prefix(command, "@list") {
							goto bad
						}
						MatchAndList(descr, player, arg1, arg2);
					default:
						goto bad;
					}
				case 'o', 'O':
					if !strings.Prefix(command, "@lock") {
						goto bad
					}
					do_lock(descr, player, arg1, arg2);
				default:
					goto bad;
				}
			case 'm', 'M':
				switch command[2] {
				case 'c', 'C':
					if strings.Prefix(command, "@mcpedit") {
						do_mcpedit(descr, player, arg1)
					} else {
						if !strings.Prefix(command, "@mcpprogram") {
							goto bad
						}
						do_mcpprogram(descr, player, arg1)
					}
				case 'p', 'P':
					if !strings.Prefix(command, "@mpitops") {
						goto bad
					}
			        do_mpi_topprofs(player, arg1);
			    case 'u', 'U':
					if !strings.Prefix(command, "@muftops") {
						goto bad
					}
			        do_muf_topprofs(player, arg1);
				default:
					goto bad;
				}
				break;
			case 'n', 'N':
				switch command[2] {
				case 'a', 'A':
					if !strings.Prefix(command, "@name") {
						goto bad
					}
					do_name(descr, player, arg1, arg2);
				case 'e', 'E':
					if command != "@newpassword" {
						goto bad
					}
					do_newpassword(player, arg1, arg2);
				default:
					goto bad;
				}
			case 'o', 'O':
				switch command[2] {
				case 'd', 'D':
					if !strings.Prefix(command, "@odrop") {
						goto bad
					}
					do_odrop(descr, player, arg1, arg2);
				case 'e', 'E':
					if !strings.Prefix(command, "@oecho") {
						goto bad
					}
					do_oecho(descr, player, arg1, arg2);
				case 'f', 'F':
					if !strings.Prefix(command, "@ofail") {
						goto bad
					}
					do_ofail(descr, player, arg1, arg2);
				case 'p', 'P':
					if !strings.Prefix(command, "@open") {
						goto bad
					}
					do_open(descr, player, arg1, arg2);
				case 's', 'S':
					if !strings.Prefix(command, "@success") {
						goto bad
					}
					do_osuccess(descr, player, arg1, arg2);
				case 'w', 'W':
					if !strings.Prefix(command, "@owned") {
						goto bad
					}
					do_owned(player, arg1, arg2);
				default:
					goto bad;
				}
			case 'p', 'P':
				switch command[2] {
				case 'a', 'A':
					if !strings.Prefix(command, "@password") {
						goto bad
					}
					do_password(player, arg1, arg2)
				case 'c', 'C':
					if !strings.Prefix(command, "@pcreate") {
						goto bad
					}
					do_pcreate(player, arg1, arg2)
				case 'e', 'E':
					if !strings.Prefix(command, "@pecho") {
						goto bad
					}
					do_pecho(descr, player, arg1, arg2)
				case 'r', 'R':
					if strings.Prefix(command, "@program") {
						if !strings.Prefix(command, "@program") {
							goto bad
						}
						do_prog(descr, player, arg1)
					} else {
						if !strings.Prefix(command, "@propset") {
							goto bad
						}
						do_propset(descr, player, arg1, arg2)
					}
				case 's', 'S':
					if !strings.Prefix(command, "@ps") {
						goto bad
					}
					list_events(player)
				default:
					goto bad
				}
			case 'r', 'R':
				switch command[3] {
				case 'c', 'C':
					if !strings.Prefix(command, "@recycle") {
						goto bad
					}
					do_recycle(descr, player, arg1)
				case 'l', 'L':
					if !strings.Prefix(command, "@relink") {
						goto bad
					}
					do_relink(descr, player, arg1, arg2);
				case 's', 'S':
					switch command {
					case "@restart":
						do_restart(player)
					case "@restrict":
						do_restrict(player, arg1)
					default:
						goto bad
					}
				default:
					goto bad
				}
			case 's', 'S':
				switch command[2] {
				case 'a', 'A':
					switch command {
					case "@sanity":
						sanity(player)
					case "@sanchange":
						sanechange(player, full_command)
					case "@sanfix":
						sanfix(player)
					default:
						goto bad
					}
				case 'e', 'E':
					if !strings.Prefix(command, "@set") {
						goto bad
					}
					do_set(descr, player, arg1, arg2);
				case 'h', 'H':
					switch command {
					case "@shutdown":
						do_shutdown(player)
					default:
						goto bad
					}
				case 't', 'T':
					if !strings.Prefix(command, "@stats") {
						goto bad
					}
					do_stats(player, arg1);
				case 'u', "U":
					if !strings.Prefix(command, "@success") {
						goto bad
					}
					do_success(descr, player, arg1, arg2)
				case 'w', 'W':
					if !strings.Prefix(command, "@sweep") {
						goto bad
					}
					do_sweep(descr, player, arg1)
				default:
					goto bad;
				}
			case 't', 'T':
				switch command[2] {
				case 'e', 'E':
					if !strings.Prefix(command, "@teleport") {
						goto bad
					}
					do_teleport(descr, player, arg1, arg2);
				case 'o', 'O':
					switch command {
					case "@toad":
						do_toad(descr, player, arg1, arg2);
					case "@tops":
						do_all_topprofs(player, arg1);
					default:
						goto bad;
					}
				case 'r', 'R':
					if !strings.Prefix(command, "@trace") {
						goto bad
					}
					do_trace(descr, player, arg1, atoi(arg2));
				case 'u', 'U':
					if !strings.Prefix(command, "@tune") {
						goto bad
					}
					do_tune(player, arg1, arg2, !!strchr(full_command, ARG_DELIMITER));
				default:
					goto bad;
				}
			case 'u', 'U':
				switch command[2] {
				case 'N', 'n':
					switch {
					case strings.Prefix(command, "@unb"):
						if !strings.Prefix(command, "@unbless") {
							goto bad
						}
						do_unbless(descr, player, arg1, arg2)
					case strings.Prefix(command, "@unli"):
						if !strings.Prefix(command, "@unlink") {
							goto bad
						}
						do_unlink(descr, player, arg1)
					case strings.Prefix(command, "@unlo"):
						if !strings.Prefix(command, "@unlock") {
							goto bad
						}
						do_unlock(descr, player, arg1)
					case strings.Prefix(command, "@uncom"):
						if !strings.Prefix(command, "@uncompile") {
							goto bad
						}
						do_uncompile(player)
					default:
						goto bad
					}
				default:
					goto bad;
				}
			case 'v', 'V':
				if !strings.Prefix(command, "@version") {
					goto bad
				}
				do_version(player);
			case 'w', 'W':
				if command != "@wall" {
					goto bad
				}
				do_wall(player, full_command);
			default:
				goto bad;
			}
		case 'd', 'D':
			switch command[1] {
			case 'i', 'I':
				if !strings.Prefix(command, "disembark") {
					goto bad
				}
				do_leave(descr, player);
			case 'r', 'R':
				if !strings.Prefix(command, "drop") {
					goto bad
				}
				do_drop(descr, player, arg1, arg2);
			default:
				goto bad;
			}
		case 'e', 'E':
			if !strings.Prefix(command, "examine") {
				goto bad
			}
			do_examine(descr, player, arg1, arg2);
		case 'g', 'G':
			switch command[1] {
			case 'e', 'E':
				if !strings.Prefix(command, "get") {
					goto bad
				}
				do_get(descr, player, arg1, arg2);
			case 'i', 'I':
				if !strings.Prefix(command, "give") {
					goto bad
				}
				do_give(descr, player, arg1, atoi(arg2));
			case 'o', 'O':
				if !strings.Prefix(command, "goto") {
					goto bad
				}
				do_move(descr, player, arg1, 0);
			case 'r', 'R':
				if command != "gripe" {
					goto bad;
				}
				do_gripe(player, full_command);
			default:
				goto bad;
			}
		case 'h', 'H':
			if !strings.Prefix(command, "help") {
				goto bad
			}
			do_help(player, arg1, arg2);
		case 'i', 'I':
			if command != "info" {
				if !strings.Prefix(command, "inventory") {
					goto bad
				}
				do_inventory(player);
			} else {
				do_info(player, arg1, arg2);
			}
		case 'k', 'K':
			if !strings.Prefix(command, "kill") {
				goto bad
			}
			do_kill(descr, player, arg1, atoi(arg2));
		case 'l', 'L':
			if strings.Prefix(command, "look") {
				do_look_at(descr, player, arg1, arg2);
			} else {
				if !strings.Prefix(command, "leave") {
					goto bad
				}
				do_leave(descr, player);
			}
		case 'm', 'M':
			switch {
			case strings.Prefix(command, "move"):
				do_move(descr, player, arg1, 0)
			case command == "motd":
				do_motd(player, full_command)
			case command == "mpi":
				do_mpihelp(player, arg1, arg2)
			default:
				if command != "man" {
					goto bad
				}
				do_man(player, (!*arg1 && !*arg2 && arg1 != arg2) ? "=" : arg1, arg2)
			}
		case 'n', 'N':
			if !strings.Prefix(command, "news") {
				goto bad
			}
			do_news(player, arg1, arg2);
		case 'p', 'P':
			switch command[1] {
			case 'a', 'A':
				if !strings.Prefix(command, "page") {
					goto bad
				}
				do_page(player, arg1, arg2);
			case 'o', 'O':
				if !strings.Prefix(command, "pose") {
					goto bad
				}
				do_pose(player, full_command);
			case 'u', 'U':
				if !strings.Prefix(command, "put") {
					goto bad
				}
				do_drop(descr, player, arg1, arg2);
			default:
				goto bad;
			}
			break;
		case 'r', 'R':
			switch command[1] {
			case 'e', 'E':
				/* undocumented alias for look */
				if !strings.Prefix(command, "read") {
					goto bad
				}
				do_look_at(descr, player, arg1, arg2);
			case 'o', 'O':
				if !strings.Prefix(command, "rob") {
					goto bad
				}
				do_rob(descr, player, arg1);
			default:
				goto bad;
			}
			break;
		case 's', 'S':
			switch command[1] {
			case 'a', 'A':
				if !strings.Prefix(command, "say") {
					goto bad
				}
				do_say(player, full_command);
			case 'c', 'C':
				if !strings.Prefix(command, "score") {
					goto bad
				}
				do_score(player);
			default:
				goto bad;
			}
		case 't', 'T':
			switch command[1] {
			case 'a', 'A':
				if !strings.Prefix(command, "take") {
					goto bad
				}
				do_get(descr, player, arg1, arg2);
			case 'h', 'H':
				if !strings.Prefix(command, "throw") {
					goto bad
				}
				do_drop(descr, player, arg1, arg2);
			default:
				goto bad;
			}
			break;
		case 'w', 'W':
			if !strings.Prefix(command, "whisper") {
				goto bad
			}
			do_whisper(descr, player, arg1, arg2)
		default:
		  bad:
			if (tp_m3_huh != 0)
			{
				hbuf := fmt.Sprintf("HUH? %s", command);
				if(can_move(descr, player, hbuf, 3)) {
					do_move(descr, player, hbuf, 3);
					match_args = ""
					match_cmdname = ""
					break;
				}
			}	
			notify(player, tp_huh_mesg);
			if tp_log_failed_commands && !controls(player, db.Fetch(player).location) {
				log_status("HUH from %s(%d) in %s(%d)[%s]: %s %s", db.Fetch(player).name, player, db.Fetch(db.Fetch(player).location).name, db.Fetch(player).location, db.Fetch(db.Fetch(db.Fetch(player).location).owner).name, command, full_command);
			}
		}
	}

	/* calculate time command took. */
	gettimeofday(&endtime, NULL);
	if (starttime.tv_usec > endtime.tv_usec) {
		endtime.tv_usec += 1000000;
		endtime.tv_sec -= 1;
	}
	endtime.tv_usec -= starttime.tv_usec;
	endtime.tv_sec -= starttime.tv_sec;

	totaltime = endtime.tv_sec + (endtime.tv_usec * 1.0e-6);
	if (totaltime > (tp_cmd_log_threshold_msec / 1000.0)) {
		log2file(LOG_CMD_TIMES, "%6.3fs, %.16s: %s%s%s%s(%d) in %s(%d):%s %s",
					totaltime, ctime((time_t *)&starttime.tv_sec),
					Wizard(db.Fetch(player).owner) ? "WIZ: " : "",
					(Typeof(player) != TYPE_PLAYER) ? db.Fetch(player).name : "",
					(Typeof(player) != TYPE_PLAYER) ? " owned by " : "",
					db.Fetch(db.Fetch(player).owner).name, (int) player,
					db.Fetch(db.Fetch(player).location).name,
					(int) db.Fetch(player).location, " ", command);
	}
}