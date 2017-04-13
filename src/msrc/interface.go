package main

import "muck"
import "os"

# define NEED_SOCKLEN_T
/* "do not include netinet6/in6.h directly, include netinet/in.h.  see RFC2553" */

type telnet_states_t int
const (
	TELNET_STATE_NORMAL = telnet_states_t(iota)
	TELNET_STATE_IAC
	TELNET_STATE_WILL
	TELNET_STATE_DO
	TELNET_STATE_WONT
	TELNET_STATE_DONT
	TELNET_STATE_SB
	TELOPT_STARTTLS = telnet_states_t(46)
	TELNET_SE = telnet_states_t(240)
	TELNET_NOP = telnet_states_t(241)
	TELNET_DM = telnet_states_t(242)
	TELNET_BRK = telnet_states_t(243)
	TELNET_IP = telnet_states_t(244)
	TELNET_AO = telnet_states_t(245)
	TELNET_AYT = telnet_states_t(246)
	TELNET_EC = telnet_states_t(247)
	TELNET_EL = telnet_states_t(248)
	TELNET_GA = telnet_states_t(249)
	TELNET_SB = telnet_states_t(250)
	TELNET_WILL = telnet_states_t(251)
	TELNET_WONT = telnet_states_t(252)
	TELNET_DO = telnet_states_t(253)
	TELNET_DONT = telnet_states_t(254)
	TELNET_IAC = telnet_states_t(255)
)

var (
	shutdown_flag bool
	restart_flag bool
)

const (
	CONNECTION_FAILED = "Either that player does not exist, or has a different password.\r\n"
	CREATION_FAILED = "Either there is already a player with that name, or that name is illegal.\r\n"
	FLUSHED = "<Output Flushed>\r\n"
	SHUTDOWN_MESSAGE = "\r\nGoing down - Bye\r\n"
)

var resolver_sock [2]int

struct text_block {
	buf string
	nxt *text_block
}

struct text_queue {
	lines int
	head *text_block
	tail *(*text_block)
};

type descriptor_data struct {
	descriptor int
	connected bool
	con_number int
	booted int
	block_writes int
	is_starttls int
	player dbref
	output_prefix string
	output_suffix string
	output_size int
	output text_queue
	input text_queue
	raw_input string
	raw_input_at string
	telnet_enabled bool
	telnet_state telnet_states_t
	telnet_sb_opt int
	short_reads bool
	last_time int
	connected_at int
	last_pinged_at int
	hostname string
	username string
	quota int
	next *descriptor_data
	prev *(*descriptor_data)
	mcpframe McpFrame
}

var descriptor_list *descriptor_data

#define MAX_LISTEN_SOCKS 16

var (
	numports int
	numsocks int
	db_conversion_flag bool
	wizonly_mode bool
	global_resolver_pid pid_t
	global_dumper_pid pid_t
	global_dumpdone bool
	sel_prof_start_time time.Time
	sel_prof_idle time.Duration
	sel_prof_idle_use int
)

var listener_port [MAX_LISTEN_SOCKS]int
var sock [MAX_LISTEN_SOCKS]int
var ndescriptors int

#define socket_write(d, buf, count) write(d->descriptor, buf, count)
#define socket_read(d, buf, count) read(d->descriptor, buf, count)

func show_program_usage(prog string) {
	l := log.New(os.Stderr, "", 0)
	l.Println("Usage: ", prog, " [<options>] [infile [outfile [portnum [portnum ...]]]]")
	l.Println("Arguments:")
	l.Println("        infile           db file loaded at startup.  optional with -dbin.")
	l.Println("        outfile          output db file to save to.  optional with -dbout.")
	l.Println("        portnum          port num to listen for conns on. (16 ports max)")
	l.Println("    Options:")
	l.Println("        -dbin INFILE     use INFILE as the database to load at startup.")
	l.Println("        -dbout OUTFILE   use OUTFILE as the output database to save to.")
	l.Println("        -port NUMBER     sets the port number to listen for connections on.")
	l.Println("        -sport NUMBER    Ignored.  SSL support isn't compiled in.")
	l.Println("        -gamedir PATH    changes directory to PATH before starting up.")
	l.Println("        -convert         load the db, then save and quit.")
	l.Println("        -nosanity        don't do db sanity checks at startup time.")
	l.Println("        -insanity        load db, then enter the interactive sanity editor.")
	l.Println("        -sanfix          attempt to auto-fix a corrupt db after loading.")
	l.Println("        -wizonly         only allow wizards to login.")
	l.Println("        -godpasswd PASS  reset God(#1)'s password to PASS.  Implies -convert")
	l.Println("        -ipv6            enable listening on ipv6 sockets.")
	l.Println("        -version         display this server's version.")
	l.Println("        -help            display this message.")
	l.Fatal()
}

/* NOTE: Will need to think about this more for unicode */
#define isinput( q ) isprint( (q) & 127 )

var descr_lookup_table []*descriptor_data
var current_descr_count int

func main() {
	var ffd *os.File
	var infile_name, outfile_name, num_one_new_passwd string
	var i, val int
	var nomore_options, sanity_skip, sanity_interactive, sanity_autofix bool

	listener_port[0] = TINYPORT
	descr_lookup_table = make([]*descriptor_data, FD_SETSIZE)
    descr_count_table = make([]*descriptor_data, FD_SETSIZE)

	for i := 1; i < len(os.Args) - 1; i++ {
		if !nomore_options && os.Args[i][0] == '-' {
			switch argv[i] {
			case "-convert":
				db_conversion_flag = true
			case "-compress":
				fmt.Print("** -compress no longer does anything\n")
			case "-nosanity":
				sanity_skip = true
			case "-insanity":
				sanity_interactive = true
			case "-wizonly":
				wizonly_mode = true
			case "-sanfix":
				sanity_autofix = true
			case "-version":
				printf("%s\n", VERSION)
				os.Exit(0)
			case "-dbin":
				if i + 1 >= len(os.Args) {
					show_program_usage(os.Args)
				}
				i++
				infile_name = os.Args[i]
			case "-dbout":
				if i + 1 >= len(os.Args) {
					show_program_usage(os.Args)
				}
				i++
				outfile_name = os.Args[i]
			case "-godpasswd":
				if i + 1 >= len(os.Args) {
					show_program_usage(os.Args)
				}
				i++
				num_one_new_passwd = os.Args[i]
				if !ok_password(num_one_new_passwd) {
					fmt.Fprintf(os.Stderr, "Bad -godpasswd password.\n")
					os.Exit(1)
				}
				db_conversion_flag = true
			case "-port":
				if i + 1 >= len(os.Args) {
					show_program_usage(os.Args)
				}
				if numports < MAX_LISTEN_SOCKS {
					i++
					listener_port[numports] = strconv.Atoi(os.Args[i])
					numports++
				}
			case "-sport":
				if i + 1 >= len(os.Args) {
					show_program_usage(os.Args)
				}
				i++
				fmt.Fprint(os.Stderr, "-sport: This server isn't configured to enable SSL.  Sorry.\n")
				os.Exit(1)
			case "-gamedir":
				if i + 1 >= len(os.Args) {
					show_program_usage(os.Args)
				}
				i++
				if chdir(os.Args[i]) {
					perror("cd to gamedir")
					exit(4)
				}
			case "--":
				nomore_options = true
			default:
				show_program_usage(os.Args)
			}
		} else {
			if infile_name == "" {
				infile_name = os.Args[i]
			} else if (!outfile_name) {
				outfile_name = os.Args[i]
			} else {
				val = strconv.Atoi(os.Args[i])
				if val < 1 || val > 65535 {
					show_program_usage(os.Args)
				}
				if numports < MAX_LISTEN_SOCKS {
					listener_port[numports] = val
					numports++
				}
			}
		}
	}
	if numports < 1 {
		numports = 1
	}
	if infile_name = "" || outfile_name = "" {
		show_program_usage(os.Args)
	}

	if !sanity_interactive {
		log_status("INIT: TinyMUCK %s starting.", "version")
#ifdef DETACH
		/* Go into the background unless requested not to */
		if !sanity_interactive && !db_conversion_flag {
			fclose(stdin)
			fclose(stdout)
			fclose(stderr)
			if fork() != 0 {
				_exit(0)
			}
		}
#endif
		/* save the PID for future use */
		if ((ffd = fopen(PID_FILE, "wb")) != NULL) {
			fprintf(ffd, "%d\n", getpid())
			fclose(ffd)
		}
		log_status("%s PID is: %d", argv[0], getpid())
#ifdef DETACH
		if !sanity_interactive && !db_conversion_flag {
			/* Detach from the TTY, log whatever output we have... */
			freopen(LOG_ERR_FILE, "a", stderr)
			setbuf(stderr, NULL)
			freopen(LOG_FILE, "a", stdout)
			setbuf(stdout, NULL)

			/* Disassociate from Process Group */
#  ifdef _POSIX_SOURCE
			setsid()
#  else
#   ifdef SYSV
			setpgrp()			/* The SysV way */
#   else
			setpgid(0, getpid())	/* The POSIX way. */
#   endif						/* SYSV */

#   ifdef  TIOCNOTTY				/* we can force this, POSIX / BSD */
			int fd;
			if fd = open("/dev/tty", O_RDWR) fd >= 0 {
				ioctl(fd, TIOCNOTTY, (char *) 0)	/* lose controll TTY */
				close(fd)
			}
#   endif						/* TIOCNOTTY */
#  endif							/* !_POSIX_SOURCE */
		}
#endif							/* DETACH */
	}

/*
 * You have to change gid first, since setgid() relies on root permissions
 * if you don't call initgroups() -- and since initgroups() isn't standard,
 * I'd rather assume the user knows what he's doing.
*/

	if !sanity_interactive {
		do_setgid(MUD_GID)
		do_setuid(MUD_ID)
	}

	/* Initialize MCP and some packages. */
	mcp_initialize()
	gui_initialize()

    sel_prof_start_time = time(NULL); /* Set useful starting time */
    sel_prof_idle = 0
    sel_prof_idle_use = 0

	if (init_game(infile_name, outfile_name) < 0) {
		fprintf(stderr, "Couldn't load %s!\n", infile_name);
		exit(2);
	}

	if (num_one_new_passwd != NULL) {
		set_password(GOD, num_one_new_passwd);
	}

	if (!sanity_interactive && !db_conversion_flag) {
		set_signals();

		if (!sanity_skip) {
			sanity(AMBIGUOUS);
			if (muck.SanityViolated) {
				wizonly_mode = 1;
				if (sanity_autofix) {
					sanfix(AMBIGUOUS);
				}
			}
		}

		/* go do it */
		shovechars();

		if (restart_flag) {
			close_sockets("\r\nServer restarting.  Try logging back on in a few minutes.\r\n");
		} else {
			close_sockets("\r\nServer shutting down normally.\r\n");
		}

		do_dequeue(-1, (dbref) 1, "all");

	}

	if (sanity_interactive) {
		san_main();
	} else {
		dump_database();
		tune_save_parmsfile();
		fclose(delta_infile);
		fclose(delta_outfile);
		(void) unlink(DELTAFILE_NAME);

		if (restart_flag) {
			char **argslist;
			char numbuf[32];
			int argcnt = numports + 2;
			int argnum = 1;
			argslist = (char**)calloc(argcnt, sizeof(char*));

			for (i = 0; i < numports; i++) {
				int alen = len(numbuf)+1;
				numbuf = fmt.Sprint(listener_port[i])
				argslist[argnum] = (char*)malloc(alen);
				strcpyn(argslist[argnum++], alen, numbuf);
			}

			if (!fork()) {
				argslist[0] = "./restart";
				execv(argslist[0], argslist);

				argslist[0] = "restart";
				execv(argslist[0], argslist);

				fprintf(stderr, "Could not find restart script!\n");
			}
		}
	}

	exit(0);
	return 0;
}

func queue_msg(d *descriptor_data, msg string) {
	mcp_frame_output_inband(&d.mcpframe, msg)
}

int notify_nolisten_level = 0;

func notify_nolisten(player dbref, msg string, isprivate bool) (r int) {
	char buf[BUFFER_LEN + 2];
	char buf2[BUFFER_LEN + 2];
	char *ptr1;
	const char *ptr2;
	dbref ref;
    int di;
    int* darr;
    int dcount;

	firstpass := true
	for ptr2 := msg; ptr2 != ""; {
		for ptr2 != "" && *ptr2 != '\r' {
			buf += ptr2[0]
			ptr2 = ptr2[1:]
		}
		buf += '\r\n'
		if ptr2[0] == '\r' {
			ptr2 = ptr2[1:]
		}

		for _, v := range get_player_descrs(player) {
            queue_msg(lookup_descriptor(v), buf)
            if firstpass {
				r++
			}
        }

		if tp_zombies {
			if TYPEOF(player) == TYPE_THING && db.Fetch(player).flags & ZOMBIE != 0 && db.Fetch(db.Fetch(player).Owner).flags & ZOMBIE == 0 && (db.Fetch(player).flags & DARK == 0 || Wizard(db.Fetch(player).Owner)) {
				ref = db.Fetch(player).Location
				if Wizard(db.Fetch(player).Owner) || ref == NOTHING || TYPEOF(ref) != TYPE_ROOM || db.Fetch(ref).flags & ZOMBIE == 0 {
					if isprivate || db.Fetch(player).Location != db.Fetch(db.Fetch(player).Owner).Location {
						ch := match_args[0]
						match_args[0] = ""
						var prefix string
						if notify_nolisten_level <= 0 {
							notify_nolisten_level++
							prefix = do_parse_prop(-1, player, player, MESGPROP_PECHO, "(@Pecho)", MPI_ISPRIVATE)
							notify_nolisten_level--
						}
						match_args[0] = ch
						if prefix == "" {
							buf2 = fmt.Sprint(db.Fetch(player).name, "> ", buf)
						} else {
							buf2 = fmt.Sprint(prefix. " ", buf)
						}

						for _, v := range get_player_descrs(db.Fetch(player).Owner) {
                            queue_msg(lookup_descriptor(v), buf2)
                            if firstpass {
								r++
							}
                        }
					}
				}
			}
		}
		firstpass = false
	}
	return
}

func notify_filtered(from, player dbref, msg string, isprivate bool) (r int) {
	if msg != "" && !ignore_is_ignoring(player, from) {
		r = notify_nolisten(player, msg, isprivate)
	}
	return
}

func notify_from_echo(from, player dbref, msg string, isprivate bool) int {
	ptr := msg
	if tp_listeners {
		if tp_listeners_obj || TYPEOF(player) == TYPE_ROOM {
			listenqueue(-1, from, db.Fetch(from).Location, player, player, NOTHING, "_listen", ptr, tp_listen_mlev, 1, 0)
			listenqueue(-1, from, db.Fetch(from).Location, player, player, NOTHING, "~listen", ptr, tp_listen_mlev, 1, 1)
			listenqueue(-1, from, db.Fetch(from).Location, player, player, NOTHING, "~olisten", ptr, tp_listen_mlev, 0, 1)
		}
	}

	if TYPEOF(player) == TYPE_THING && db.Fetch(player).flags & VEHICLE == 0 && (db.Fetch(player).flags & DARK == 0 || Wizard(db.Fetch(player).Owner)) {
		ref := db.Fetch(player).Location
		if Wizard(db.Fetch(player).Owner) || ref == NOTHING || TYPEOF(ref) != TYPE_ROOM || db.Fetch(ref).flags & VEHICLE == 0 {
			if !isprivate && db.Fetch(from).Location == db.Fetch(player).Location {
				ch := match_args[0]
				match_args[0] = '\0'
				prefix := do_parse_prop(-1, from, player, MESGPROP_OECHO, "(@Oecho)", MPI_ISPRIVATE)
				match_args[0] = ch

				if prefix == "" {
					prefix = "Outside>"
				}
				buf := fmt.Sprint(prefix, " ", msg)
				for ref = db.Fetch(player).Contents; ref != NOTHING; ref = db.Fetch(ref).next {
					notify_filtered(from, ref, buf, isprivate);
				}
			}
		}
	}
	return notify_filtered(from, player, msg, isprivate)
}

func notify_from(from, player dbref, msg string) int {
	return notify_from_echo(from, player, msg, 1)
}

func notify(player dbref, msg string) int {
	return notify_from_echo(player, player, msg, 1)
}

func update_quotas(last, current time.Duration) time.Duration {
	if nslices := (current - last) / (tp_command_time_msec * time.Millisecond); nslices > 0 {
		for d := descriptor_list; d != nil; d = d.next {
			var cmds_per_time int
			if d.connected && db.Fetch(d.player).flags & INTERACTIVE != 0 {
				cmds_per_time = tp_commands_per_time * 8
			} else {										
				cmds_per_time = tp_commands_per_time
			}
			d.quota += cmds_per_time * nslices
			if d.quota > tp_command_burst_size {
				d.quota = tp_command_burst_size
			}
		}
	}
	return last + (nslices * tp_command_time_msec * time.Millisecond)
}

/*
 * long max_open_files()
 *
 * This returns the max number of files you may have open
 * as a long, and if it can use setrlimit() to increase it,
 * it will do so.
 *
 * Becuse there is no way to just "know" if get/setrlimit is
 * around, since its defs are in <sys/resource.h>, you need to
 * define USE_RLIMIT in config.h to attempt it.
 *
 * Otherwise it trys to use sysconf() (POSIX.1) or getdtablesize()
 * to get what is avalible to you.
 */
#ifdef HAVE_RESOURCE_H
# include <sys/resource.h>
#endif

#if defined(RLIMIT_NOFILE) || defined(RLIMIT_OFILE)
# define USE_RLIMIT
#endif

long
max_open_files(void)
{
#if defined(_SC_OPEN_MAX) && !defined(USE_RLIMIT)	/* Use POSIX.1 method, sysconf() */
/*
 * POSIX.1 code.
 */
	return sysconf(_SC_OPEN_MAX);
#else							/* !POSIX */
# if defined(USE_RLIMIT) && (defined(RLIMIT_NOFILE) || defined(RLIMIT_OFILE))
#  ifndef RLIMIT_NOFILE
#   define RLIMIT_NOFILE RLIMIT_OFILE	/* We Be BSD! */
#  endif						/* !RLIMIT_NOFILE */
/*
 * get/setrlimit() code.
 */
	struct rlimit file_limit;

	getrlimit(RLIMIT_NOFILE, &file_limit);	/* Whats the limit? */

	if (file_limit.rlim_cur < file_limit.rlim_max) {	/* if not at max... */
		file_limit.rlim_cur = file_limit.rlim_max;	/* ...set to max. */
		setrlimit(RLIMIT_NOFILE, &file_limit);

		getrlimit(RLIMIT_NOFILE, &file_limit);	/* See what we got. */
	}

	return (long) file_limit.rlim_cur;
# else							/* !RLIMIT */
/*
 * Don't know what else to do, try getdtablesize().
 * email other bright ideas to me. :) (whitefire)
 */
	return (long) getdtablesize();
# endif							/* !RLIMIT */
#endif							/* !POSIX */
}

func queue_immediate(d *descriptor_data, msg string) int {
	var quote_len int
	if d.mcpframe.enabled && !strings.HasPrefix(msg, MCP_MESG_PREFIX) && !strings.HasPrefix(msg, MCP_QUOTE_PREFIX) {
		quote_len = len(MCP_QUOTE_PREFIX)
		socket_write(d, MCP_QUOTE_PREFIX, quote_len)
	}
	return socket_write(d, msg, len(msg)) + quote_len
}

func goodbye_user(d *descriptor_data) {
	queue_immediate(d, "\r\n")
	queue_immediate(d, tp_leave_mesg)
	queue_immediate(d, "\r\n\r\n")
}

func idleboot_user(d *descriptor_data) {
	queue_immediate(d, "\r\n")
	queue_immediate(d, tp_idle_mesg)
	queue_immediate(d, "\r\n\r\n")
	d.booted = 1
}

static int con_players_max = 0;	/* one of Cynbe's good ideas. */
static int con_players_curr = 0;	/* for playermax checks. */

func shovechars() {
	fd_set input_set, output_set;
	long tmptq;
	int maxd = 0, cnt;
	struct descriptor_data *d, *dnext;
	struct descriptor_data *newd;
	int avail_descriptors;
	int i;

	for i := 0; i < numports; i++ {
		sock[i] = make_socket(listener_port[i])
		maxd = sock[i] + 1
		numsocks++
	}
	last_slice := time.Now()
	avail_descriptors = max_open_files() - 5;
	now := time.Now()

/* And here, we do the actual player-interaction loop */

	for !shutdown_flag {
		current_time := time.Now()
		last_slice = update_quotas(last_slice, current_time)
		next_muckevent()
		process_commands()
		muf_event_process()
		for d = descriptor_list; d; d = dnext {
			dnext = d.next
			if d.booted != 0 {
				process_output(d)
				if d.booted == 2 {
					goodbye_user(d)
				}
				d.booted = 0
				process_output(d)
				shutdownsock(d)
			}
		}
		if global_dumpdone {
			if tp_dumpdone_warning {
				wall_and_flush(tp_dumpdone_mesg)
			}
			global_dumpdone = 0
		}

		if shutdown_flag {
			break
		}
		timeout := 10 * time.Second
		next_slice := last_slice.Add(tp_command_time_msec * time.Microsecond)
		slice_timeout := next_slice - current_time

		FD_ZERO(&input_set)
		FD_ZERO(&output_set)
		if ndescriptors < avail_descriptors {
			for i := 0; i < numsocks; i++ {
				FD_SET(sock[i], &input_set)
			}
		}
		for d := descriptor_list; d != nil; d = d.next {
			if d.input.lines > 100 {
				timeout = slice_timeout
			} else {
				FD_SET(d.descriptor, &input_set)
			}
			if (d.output.head && !d.block_writes) {
				FD_SET(d.descriptor, &output_set)
			}
		}

		tmptq = next_muckevent_time()
		if tmptq >= 0 && timeout > tmptq {
			timeout = tmptq + tp_pause_min
		}
		sel_in := time.Now()
		if select(maxd, &input_set, &output_set, (fd_set *) 0, &timeout) < 0 {
			if errno != EINTR {
				perror("select")
				return
			}
		} else {
			sel_out := time.Now()
			sel_prof_idle += sel_out - sel_in
			sel_prof_idle_use++
			now = time.Now()
			for i := 0; i < numsocks; i++ {
				if (FD_ISSET(sock[i], &input_set)) {
					if (!(newd = new_connection(listener_port[i], sock[i]))) {
						if (errno && errno != EINTR && errno != EMFILE && errno != ENFILE) {
							perror("new_connection");
							/* return; */
						}
					} else {
						if (newd->descriptor >= maxd)
							maxd = newd->descriptor + 1;
					}
				}
			}
			for (cnt = 0, d = descriptor_list; d; d = dnext) {
				dnext = d->next;
				if (FD_ISSET(d->descriptor, &input_set)) {
					if (!process_input(d)) {
						d->booted = 1
					}
				}
				if (FD_ISSET(d->descriptor, &output_set)) {
					if (!process_output(d)) {
						d->booted = 1
					}
				}
				if (d->connected) {
					cnt++;
					if (tp_idleboot && ((now - d->last_time) > tp_maxidle) &&
						!Wizard(d->player)) {
						idleboot_user(d);
					}
				} else {
					/* Hardcode 300 secs -- 5 mins -- at the login screen */
					if ((now - d->connected_at) > 300) {
						log_status("connection screen: connection timeout 300 secs");
						d->booted = 1
					}
				}
				if ( d->connected && tp_idle_ping_enable && (tp_idle_ping_time > 0) && ((now - d->last_pinged_at) > tp_idle_ping_time) ) {
					const char *tmpptr = get_property_class( d->player, "_/sys/no_idle_ping" );
					if( !tmpptr && !send_keepalive(d)) {
						d->booted = 1
					}
				}
			}
			if (cnt > con_players_max) {
				add_property((dbref) 0, "_sys/max_connects", NULL, cnt);
				con_players_max = cnt;
			}
			con_players_curr = cnt;
		}
	}

	/* End of the player processing loop */

	now = time.Now()
	add_property(0, "_sys/lastdumptime", nil, (int) now)
	add_property(0, "_sys/shutdowntime", nil, (int) now)
}


func wall_and_flush(msg string) {
	if msg != "" {
		buf := msg + "\r\n"
		var dnext *descriptor_data
		for d := descriptor_list; d; d = dnext {
			dnext = d.next
			queue_msg(d, buf)
			if !process_output(d) {
				d.booted = 1
			}
		}
	}
}


func flush_user_output(dbref player) {
	for _, v := range get_player_descrs(db.Fetch(player).Owner) {
		if d := lookup_descriptor(v); d != nil && !process_output(d) {
            d.booted = 1
        }
    }
}

func wall_wizards(msg string) {
	buf := msg + "\r\n"
	var dnext *descriptor_data
	for d := descriptor_list; d != nil; d = dnext {
		dnext = d.next
		if d.connected && Wizard(d.player) {
			queue_msg(d, buf)
			if !process_output(d) {
				d.booted = 1
			}
		}
	}
}

func new_connection(port, sock int) (r *descriptor_data) {
	// FIXME: tcp/ip connection setup - with or without TLS
	var addr sockaddr_in

	addr_len := (socklen_t)sizeof(addr);
	if newsock := accept(sock, (struct sockaddr *) &addr, &addr_len); newsock > -1 {
		hostname := addrout(port, addr.sin_addr.s_addr, addr.sin_port)
		log_status("ACCEPT: %s on descriptor %d", hostname, newsock)
		ndescriptors++
		log_status("CONCOUNT: There are now %d open connections.", ndescriptors)
		r = initializesock(newsock, hostname)
	}
	return
}

/*  addrout -- Translate address 'a' from addr struct to text.		*/

func addrout(int lport, long a, unsigned short prt) string {
	static char buf[128];
	struct in_addr addr;

	memset(&addr, 0, sizeof(addr))
	memcpy(&addr.s_addr, &a, sizeof(struct in_addr));

	prt = ntohs(prt);

	if (tp_hostnames) {
		/* One day the nameserver Qwest uses decided to start */
		/* doing halfminute lags, locking up the entire muck  */
		/* that long on every connect.  This is intended to   */
		/* prevent that, reduces average lag due to nameserver */
		/* to 1 sec/call, simply by not calling nameserver if */
		/* it's in a slow mood *grin*. If the nameserver lags */
		/* consistently, a hostname cache ala OJ's tinymuck2.3 */
		/* would make more sense:                             */
		static int secs_lost = 0;

		if (secs_lost) {
			secs_lost--;
		} else {
			time_t gethost_start = time(NULL);

			struct hostent *he = gethostbyaddr(((char *) &addr),
											   sizeof(addr), AF_INET);
			time_t gethost_stop = time(NULL);
			time_t lag = gethost_stop - gethost_start;

			if (lag > 10) {
				secs_lost = lag;

#if MIN_SECS_TO_LOG
				if (lag >= CFG_MIN_SECS_TO_LOG) {
					log_status("GETHOSTBYNAME-RAN: secs %3d", lag);
				}
#endif

			}
			if (he) {
				buf = fmt.Sprintf("%s(%u)", he->h_name, prt);
				return buf;
			}
		}
	}

	a = ntohl(a);
	buf = fmt.Sprintf("%ld.%ld.%ld.%ld(%u)", (a >> 24) & 0xff, (a >> 16) & 0xff, (a >> 8) & 0xff, a & 0xff, prt);
	return buf;
}

func shutdownsock(d *descriptor_data) {
	if d != nil {
		if d.connected {
			log_status("DISCONNECT: descriptor %d player %s(%d) from %s(%s)", d.descriptor, db.Fetch(d.player).name, d.player, d.hostname, d.username)
			announce_disconnect(d)
		} else {
			log_status("DISCONNECT: descriptor %d from %s(%s) never connected.", d.descriptor, d.hostname, d.username)
		}
		d.output_prefix = ""
		d.output_suffix = ""
		shutdown(d.descriptor, 2)
		close(d.descriptor)
		descr_lookup_table[d.descriptor] = nil
		freeqs(d)
		*d.prev = d.next
		if d.next != nil {
			d.next.prev = d.prev
		}
		d.hostname = ""
		d.username = ""
		mcp_frame_clear(&d.mcpframe)
		free(d)
		ndescriptors--
		log_status("CONCOUNT: There are now %d open connections.", ndescriptors)
	}
}

func FlushText(mfr *McpFrame) {
	if d := mfr.descriptor; d != nil && !process_output(d) {
		d.booted = 1
	}
}

func initializesock(s int, hostname string) (d *descriptor_data) {
	d = &descriptor_data{
		descriptor: s,
		player: -1,
		telnet_state: TELNET_STATE_NORMAL,
		quota: tp_command_burst_size,
	}
	make_nonblocking(s)
	d.output.tail = &d.output.head
	d.input.tail = &d.input.head
	d.mcpframe = &McpFrame{ descriptor: d }
	mcp_frame_list = &McpFrameList{ mfr: d.mcpframe, next: mcp_frame_list }
	buf := hostname
	ptr := strchr(buf, ')')
	if ptr {
		*ptr = '\0'
	}
	ptr = strchr(buf, '(')
	*ptr++ = '\0'
	d.hostname = buf
	d.username = ptr
	if descriptor_list != nil {
		descriptor_list.prev = &d.next
	}
	d.next = descriptor_list
	d.prev = &descriptor_list
	descriptor_list = d
	descr_lookup_table[d.descriptor] = d
	mcp_negotiation_start(&d.mcpframe)
	welcome_user(d)
	return
}

func make_socket(int port) int {
	int s;
	int opt;
	struct sockaddr_in server;

	s := socket(AF_INET, SOCK_STREAM, 0)
	if s < 0 {
		perror("creating stream socket");
		exit(3);
	}

	opt = 1;
	if (setsockopt(s, SOL_SOCKET, SO_REUSEADDR, (char *) &opt, sizeof(opt)) < 0) {
		perror("setsockopt");
		exit(1);
	}

	opt = 1;
	if (setsockopt(s, SOL_SOCKET, SO_KEEPALIVE, (char *) &opt, sizeof(opt)) < 0) {
		perror("setsockopt");
		exit(1);
	}

	/*
	opt = 240;
	if (setsockopt(s, SOL_TCP, TCP_KEEPIDLE, (char *) &opt, sizeof(opt)) < 0) {
		perror("setsockopt");
		exit(1);
	}
	*/

	server.sin_family = AF_INET;
	server.sin_addr.s_addr = INADDR_ANY;
	server.sin_port = htons(port);

	if (bind(s, (struct sockaddr *) &server, sizeof(server))) {
		perror("binding stream socket");
		close(s);
		exit(4);
	}
	listen(s, 5);
	return s;
}

func (q *text_queue) Add(b string) {
	if b != "" {
		p := &text_block{ buf: b }
		*q.tail = p
		q.tail = &p.nxt
		q.lines++
	}
}

func (q *text_queue) Flush(desired int) (actual int) {
	desired += len(FLUSHED)
	for p = q.head; n > 0 && p != nil; p = q.head {
		desired -= len(p.buf)
		actual += len(p.buf)
		q.head = p.nxt
		q.lines--
	}
	p := &text_block{ buf: FLUSHED, nxt: q.head }
	q.head = p
	q.lines++
	if p.nxt == nil {
		q.tail = &p.nxt
	}
	actual -= len(p.buf)
	return
}

func (d *descriptor_data) QueueWrite(b string) int {
	if space := tp_max_output - d.output_size - n; space < 0 {
		d.output_size -= d.output.Flush(-space)
	}
	d.output.Add(b)
	d.output_size += len(b)
	return len(b)
}

func send_keepalive(struct descriptor_data *d) int {
	int cnt;
	unsigned char telnet_nop[] = {
		TELNET_IAC, TELNET_NOP, '\0'
	};

	/* drastic, but this may give us crash test data */
	if (!d || !d->descriptor) {
		panic("send_keepalive(): bad descriptor or connect struct !");
	}

	if (d->telnet_enabled) {
		cnt = socket_write(d, telnet_nop, 2);
	} else {
		cnt = socket_write(d, "", 0);
	}
	/* We expect a 0 return */
	if (cnt < 0) {
		if (errno == EWOULDBLOCK)
			return 1;
		if (errno == 0)
			return 1;
		log_status("keepalive socket write descr=%i, errno=%i", d->descriptor, errno);
		return 0;
	}
	return 1;
}

func process_output(d *descriptor_data) bool {
	switch {
	case d == nil, d.descriptor == nil:
		panic("process_output(): bad descriptor or connect struct !")
	case d.output.lines == 0, d.block_writes:
	default:
		qp := &d.output.head
		for cur = *qp; cur != 0; cur = *qp {
			cnt := socket_write(d, cur.start, cur.nchars)
			if cnt <= 0 {
				return errno == EWOULDBLOCK
			}
			d.output_size -= cnt
			if cnt == cur.nchars {
				d.output.lines--
				if !cur.nxt {
					d.output.tail = qp
					d.output.lines = 0
				}
				*qp = cur.nxt
			} else {
				cur.nchars -= cnt
				cur.start += cnt
				break
			}
		}
	}
	return true
}

# if !defined(O_NONBLOCK)	/* POSIX ME HARDER */
#  ifdef FNDELAY					/* SUN OS */
#   define O_NONBLOCK FNDELAY
#  else
#   ifdef O_NDELAY				/* SyseVil */
#    define O_NONBLOCK O_NDELAY
#   endif
#  endif
# endif

func make_nonblocking(s int) {
	if fcntl(s, F_SETFL, O_NONBLOCK) == -1 {
		perror("make_nonblocking: fcntl")
		panic("O_NONBLOCK fcntl failed")
	}
}

func freeqs(d *descriptor_data) {
	cur := d.output.head
	for cur != nil {
		cur = cur.next
	}
	d.output.lines = 0
	d.output.head = 0
	d.output.tail = &d.output.head

	cur = d.input.head
	d.input.head = nil
	for cur != nil {
		cur = cur.nxt
	}
	d.input.lines = 0
	d.input.head = 0
	d.input.tail = &d.input.head

	d.raw_input = ""
	d.raw_input_at = ""
}

func process_input(d *descriptor_data) bool {
	char buf[MAX_COMMAND_LEN * 2];
	int maxget = sizeof(buf);
	int got;
	char *p, *pend, *q, *qend;

	if d.short_reads {
	    maxget = 1
	}
	got = socket_read(d, buf, maxget)

	if got <= 0 {
		return false
	}

	if d.raw_input == "" {
		d.raw_input_at = ""
	}
	p = d.raw_input_at
	pend = len(d.raw_input)
	for _, q := range buf {
		switch {
		case q[0] == '\n':
			d.last_time = time(NULL)
			if d.raw_input != "" {
				d.input.Add(d.raw_input)
			}
			d.raw_input = ""
		case d.telnet_state == TELNET_STATE_IAC:
			switch q[0] {
			case TELNET_NOP:
				d.telnet_state = TELNET_STATE_NORMAL
			case TELNET_BRK, TELNET_IP:
				//	BREAK or INTERRUPT
				d.input.Add(BREAK_COMMAND)
				d.telnet_state = TELNET_STATE_NORMAL
			case TELNET_AO:
				/* Abort Output */
				/* could be handy, but for now leave unimplemented */
				d.telnet_state = TELNET_STATE_NORMAL
			case TELNET_AYT:
				/* AYT */
				sendbuf := "[Yes]\r\n"
				socket_write(d, sendbuf, len(sendbuf))
				d.telnet_state = TELNET_STATE_NORMAL
			case TELNET_EC:
				/* Erase character */
				if d.raw_input != "" {
					d.raw_input = d.raw_input[:len(d.raw_input) - 1]
				}
				d.telnet_state = TELNET_STATE_NORMAL
			case TELNET_EL:
				/* Erase line */
				d.raw_input = ""
				d.telnet_state = TELNET_STATE_NORMAL
			case TELNET_GA:
				/* Go Ahead */
				/* treat as a NOP (?) */
				d.telnet_state = TELNET_STATE_NORMAL
			case TELNET_WILL:
				d.telnet_state = TELNET_STATE_WILL
			case TELNET_WONT:
				d.telnet_state = TELNET_STATE_WONT
			case TELNET_DO:
				d.telnet_state = TELNET_STATE_DO;
			case TELNET_DONT:
				d.telnet_state = TELNET_STATE_DONT
			case TELNET_SB:
				/* SB (option subnegotiation) */
				d.telnet_state = TELNET_STATE_SB
			case TELNET_SE:
				/* Go Ahead */
				d.telnet_state = TELNET_STATE_NORMAL
			case TELNET_IAC:
				/* IAC a second time */
				d.telnet_state = TELNET_STATE_NORMAL
			default:
				/* just ignore */
				d.telnet_state = TELNET_STATE_NORMAL
			}
		case d.telnet_state == TELNET_STATE_WILL:
			/* We don't negotiate: send back DONT option */
			sendbuf := TELNET_IAC + TELNET_DONT + q[0]
			socket_write(d, sendbuf, 3)
			d.telnet_state = TELNET_STATE_NORMAL
			d.telnet_enabled = true
		case d.telnet_state == TELNET_STATE_DO:
			/* We don't negotiate: send back WONT option */
			sendbuf := TELNET_IAC + TELNET_WONT + q[0]
			socket_write(d, sendbuf, 3);
			d.telnet_state = TELNET_STATE_NORMAL
			d.telnet_enabled = true
		case d.telnet_state == TELNET_STATE_WONT:
			/* Ignore WONT option. */
			d.telnet_state = TELNET_STATE_NORMAL
			d.telnet_enabled = true
		case d.telnet_state == TELNET_STATE_DONT:
			/* We don't negotiate: send back WONT option */
			sendbuf[0] := TELNET_IAC + TELNET_WONT + q[0]
			socket_write(d, sendbuf, 3)
			d.telnet_state = TELNET_STATE_NORMAL
			d.telnet_enabled = true
		case d.telnet_state == TELNET_STATE_SB:
			d.telnet_sb_opt = q[0]
			/* TODO: Start remembering subnegotiation data. */
			d.telnet_state = TELNET_STATE_NORMAL
		case q[0] == TELNET_IAC:
			/* Got TELNET IAC, store for next byte */	
			d.telnet_state = TELNET_STATE_IAC
		default:
			switch {
			case isprint(q & 127):
				d.raw_input += q[0]
			case q[0] == '\t':
				d.raw_input += ' '
			case q[0] == 8, q[0] == 127 {
				/* if BS or DEL, delete last character */
				if d.raw_input != "" {
					d.raw_input = d.raw_input[:len(d.raw_input) - 1]
				}
			}
			d.telnet_state = TELNET_STATE_NORMAL
		}
	}
	if d.raw_input != "" {
		d.raw_input_at = d.raw_input
	} else {
		d.raw_input = ""
		d.raw_input_at = ""
	}
	return true
}

func process_commands() {
	for {
		nprocessed := 0
		for d := descriptor_list; d != nil; {
			if t := d.input.head; d.quota > 0 && t != nil {
				if d.connected && db.Fetch(d.player).(Player).block != nil && !is_interface_command(t.start)) {
					tmp := t.start
					if strings.HasPrefix(tmp, "#$\"") {
						/* Un-escape MCP escaped lines */
						tmp = tmp[3:]
					}
					/* WORK: send player's foreground/preempt programs an exclusive READ mufevent */
					if !read_event_notify(d.descriptor, d.player, tmp) && tmp == "" {
						/* Didn't send blank line.  Eat it.  */
						nprocessed++
						d.input.head = t.nxt
						d.input.lines--
						if !d.input.head {
							d.input.tail = &d.input.head
							d.input.lines = 0
						}
					}
				} else {
					if !strings.HasPrefix(t.start, "#$#") {
						/* Not an MCP mesg, so count this against quota. */
						d.quota--
					}
					nprocessed++
					if !do_command(d, t.start) {
						d.booted = 2
						/* Disconnect player next pass through main event loop. */
					}
					d.input.head = t.nxt
					d.input.lines--
					if !d.input.head {
						d.input.tail = &d.input.head
						d.input.lines = 0
					}
				}
			}
			d = d.next
		}
		if processed == 0 {
			break
		}
	}
}

func is_interface_command(cmd string) (r bool) {
	tmp := cmd
	if strings.HasPrefix(tmp, "#$\"") {
		/* dequote MCP quoting. */
		tmp = cmd[3:]
	}
	switch {
	case strings.HasPrefix(cmd, "#$#"):
		/* MCP mesg. */
		r = true
	case tmp == BREAK_COMMAND:
		r = true
	case tmp == QUIT_COMMAND:
		r = true
	case strings.HasPrefix(tmp, WHO_COMMAND):
		r = true
	case strings.HasPrefix(tmp, PREFIX_COMMAND):
		r = true
	case strings.HasPrefix(tmp, SUFFIX_COMMAND):
		r = true
	case tp_recognize_null_command && tmp == NULL_COMMAND:
		r = true
	}
	return
}

func do_command(d *descriptor_data, command string) (r bool) {
	char buf[BUFFER_LEN];
	char cmdbuf[BUFFER_LEN];

	if !mcp_frame_process_input(&d->mcpframe, command, cmdbuf, sizeof(cmdbuf)) {
		d.quota++
		r = true
	} else {
		command = cmdbuf
		if d.connected {
			ts_lastuseobject(d.player)
		}

		switch {
		case command == BREAK_COMMAND:
			if r = d.connected; r {
				if dequeue_prog(d.player, 2) {
					if d.output_prefix {
						queue_msg(d, d.output_prefix)
						d.QueueWrite("\r\n")
					}
					queue_msg(d, "Foreground program aborted.\r\n")
					if db.Fetch(d.player).flags & INTERACTIVE != 0 && db.Fetch(d.player).flags & READMODE != 0 {
						process_command(d.descriptor, d.player, command)
					}
					if d.output_suffix {
						queue_msg(d, d.output_suffix)
						d.QueueWrite("\r\n")
					}
				}
				db.Fetch(d.player).(Player).block = false
			}
		case command == QUIT_COMMAND:
			r = false
		case tp_recognize_null_command && command == NULL_COMMAND:
			r = true
		case strings.HasPrefix(command, WHO_COMMAND), command[0] == OVERIDE_TOKEN && strings.HasPrefix(command[1:], WHO_COMMAND):
			if d.output_prefix != "" {
				queue_msg(d, d.output_prefix)
				d.QueueWrite("\r\n")
			}
			buf = fmt.Sprintf("@%v %v", WHO_COMMAND, command[len(WHO_COMMAND):]
			if !d.connected || db.Fetch(d.player).flags & INTERACTIVE != 0 {
				if tp_secure_who {
					queue_msg(d, "Sorry, WHO is unavailable at this point.\r\n")
				} else {
					dump_users(d, command[len(WHO_COMMAND):])
				}
			} else {
				if (!(TrueWizard(db.Fetch(d.player).Owner) && (*command == OVERIDE_TOKEN))) && can_move(d.descriptor, d.player, buf, 2) {
					do_move(d.descriptor, d.player, buf, 2)
				} else {
					dump_users(d, command + sizeof(WHO_COMMAND) - ((*command == OVERIDE_TOKEN) ? 0 : 1))
				}
			}
			if d.output_suffix != "" {
				queue_msg(d, d.output_suffix)
				d.QueueWrite("\r\n")
			}
			r = true
		case strings.HasPrefix(command, PREFIX_COMMAND):
			d.output_suffix = strings.TrimSpace(command[PREFIX_COMMAND:])
			r = true
		case strings.HasPrefix(command, SUFFIX_COMMAND):
			d.output_suffix = strings.TrimSpace(command[SUFFIX_COMMAND:])
			r = true
		default:
			if d.connected {
				if d.output_prefix != "" {
					queue_msg(d, d.output_prefix)
					d.QueueWrite("\r\n")
				}
				process_command(d.descriptor, d.player, command)
				if d.output_suffix != "" {
					queue_msg(d, d.output_suffix)
					d.QueueWrite("\r\n")
				}
			} else {
				check_connect(d, command)
			}
			r = true
		}
	}
	return
}

func interact_warn(player dbref) {
	switch {
	case db.Fetch(player).flags & INTERACTIVE == 0:
	case db.Fetch(player).flags & READMODE != 0:
		notify(player, "*** You are currently using a program.  Use \"@Q\" to return to a more reasonable state of control. ***")
	case db.Fetch(player).(Player).insert_mode:
		notify(player, "*** You are currently inserting MUF program text.  Use \".\" to return to the editor, then \"quit\" if you wish to return to your regularly scheduled Muck universe. ***")
	default:
		notify(player, "*** You are currently using the MUF program editor. ***")
	}
}

func check_connect(d *descriptor_data, msg string) {
	var command, user, password string
	parse_connect(msg, command, user, password)
	switch {
	case strings.HasPrefix(command, "co"):
		if player := connect_player(user, password); player == NOTHING {
			queue_msg(d, CONNECTION_FAILED)
			log_status("FAILED CONNECT %s on descriptor %d", user, d.descriptor)
		} else {
			if wizonly_mode || (tp_playermax && con_players_curr >= tp_playermax_limit) && !TrueWizard(player) {
				if wizonly_mode {
					queue_msg(d, "Sorry, but the game is in maintenance mode currently, and only wizards are allowed to connect.  Try again later.")
				} else {
					queue_msg(d, tp_playermax_bootmesg)
				}
				d.QueueWrite("\r\n")
				d.booted = 1
			} else {
				log_status("CONNECTED: %s(%d) on descriptor %d", db.Fetch(player).name, player, d.descriptor)
				d.connected = true
				d.connected_at = time(nil)
				d.player = player
				update_desc_count_table()
				remember_player_descr(player, d.descriptor)
				/* cks: someone has to initialize this somewhere. */
				db.Fetch(d.player).(Player).block = false
				spit_file(player, MOTD_FILE)
				announce_connect(d.descriptor, player)
				interact_warn(player)
				if (muck.SanityViolated && Wizard(player)) {
					notify(player, "#########################################################################")
					notify(player, "## WARNING!  The DB appears to be corrupt!  Please repair immediately! ##")
					notify(player, "#########################################################################")
				}
				con_players_curr++
			}
		}
	case strings.HasPrefix(command, "cr"):
		if !tp_registration {
			if wizonly_mode || (tp_playermax && con_players_curr >= tp_playermax_limit) {
				if wizonly_mode {
					queue_msg(d, "Sorry, but the game is in maintenance mode currently, and only wizards are allowed to connect.  Try again later.")
				} else {
					queue_msg(d, tp_playermax_bootmesg)
				}
				d.QueueWrite("\r\n")
				d.booted = 1
			} else {
				if player := create_player(user, password); player == NOTHING {
					queue_msg(d, CREATION_FAILED)
					log_status("FAILED CREATE %s on descriptor %d", user, d.descriptor)
				} else {
					log_status("CREATED %s(%d) on descriptor %d", db.Fetch(player).name, player, d.descriptor)
					d.connected = true
					d.connected_at = time(nil)
					d.player = player
					update_desc_count_table()
					remember_player_descr(player, d.descriptor)
					/* cks: someone has to initialize this somewhere. */
					db.Fetch(d.player).(Player).block = false
					spit_file(player, MOTD_FILE)
					announce_connect(d.descriptor, player)
					con_players_curr++
				}
			}
		} else {
			queue_msg(d, tp_register_mesg)
			d.QueueWrite("\r\n")
			log_status("FAILED CREATE %s on descriptor %d", user, d.descriptor)
		}
	case command == "":
		/* do nothing */
	default:
		welcome_user(d)
	}
}

func parse_connect(msg string) (command, user, pass string) {
	msg = strings.TrimSpace(msg)
	if i := strings.IndexFunc(msg, unicode.IsSpace); i != -1 {
		command = msg[:i]
		msg = strings.TrimSpace(msg[i:])
	}

	if i := strings.IndexFunc(msg, unicode.IsSpace); i != -1 {
		user = msg[:i]
		msg = strings.TrimSpace(msg[i:])
	}

	if i := strings.IndexFunc(msg, unicode.IsSpace); i != -1 {
		pass = msg[:i]
	}
}


func boot_off(player dbref) (r bool) {
	if arr := get_player_descrs(player); arr != nil {
        if last := lookup_descriptor(arr[0]); last != nil {
			process_output(last)
			last.booted = 1
			r = true
		}
	}
	return
}

func boot_player_off(player dbref) {
	for _, v := get_player_descrs(player) {
        if d := lookup_descriptor(v); d != nil {
            d.booted = 1
        }
    }
}

func close_sockets(msg string) {
	var dnext *descriptor_data
	for d := descriptor_list; d != nil; d = dnext {
		dnext = d.next
		if d.connected {
			forget_player_descr(d.player, d.descriptor)
		}
		socket_write(d, msg, len(msg))
		socket_write(d, SHUTDOWN_MESSAGE, len(SHUTDOWN_MESSAGE))
		d.output_prefix = ""
		d.output_suffix = ""
		if shutdown(d.descriptor, 2) < 0 {
			perror("shutdown")
		}
		close(d.descriptor)
		freeqs(d)
		*d.prev = d.next
		if d.next != nil {
			d.next.prev = d.prev
		}
		d.hostname = ""
		d.username = ""
		mcp_frame_clear(&d.mcpframe)
		ndescriptors--
	}
	update_desc_count_table();
	for i := 0; i < numsocks; i++ {
		close(sock[i])
	}
}


func do_armageddon(player dbref, msg string) {
	if !Wizard(player) {
		notify(player, "Sorry, but you don't look like the god of War to me.")
		log_status("ILLEGAL ARMAGEDDON: tried by %s", unparse_object(player, player))
	} else {
		buf := fmt.Sprintf("\r\nImmediate shutdown initiated by %s.\r\n", db.Fetch(player).name)
		buf += msg
		log_status("ARMAGEDDON initiated by %s(%d): %s", db.Fetch(player).name, player, msg)
		fprintf(stderr, "ARMAGEDDON initiated by %s(%d): %s\n", db.Fetch(player).name, player, msg)
		close_sockets(buf)
		exit(1)
	}
}

func dump_users(e *descriptor_data, user string) {
	int players;
	time_t now;
	char buf[2048];
	char pbuf[64];

/* -- Wizard should always override tp_who_doing JES
	if (tp_who_doing) {
		wizard = e->connected && e.player == GOD
	} else {
		wizard = e->connected && Wizard(e->player);
	}
*/

	var wizard bool
	for user != "" && (unicode.IsSpace(user[0]) || user[0] == '*') {
		if tp_who_doing && *user == '*' && e.connected && Wizard(e.player) {
			wizard = true
		}
		user = user[1:]
	}

	if wizard {
		/* S/he is connected and not quelled. Okay; log it. */
		log_command("WIZ: %s(%d) in %s(%d):  %s", db.Fetch(e.player).name, e.player, db.Fetch(db.Fetch(e.player).Location).name, db.Fetch(e.player).Location, "WHO")
	}

	(void) time(&now);
	if wizard {
		queue_msg(e, "Player Name                Location     On For Idle   Host\r\n")
	} else {
		if tp_who_doing {
			queue_msg(e, "Player Name           On For Idle   Doing...\r\n");
		} else {
			queue_msg(e, "Player Name           On For Idle\r\n");
		}
	}

	players = 0
	for d := descriptor_list; d != nil; d = d.next {
		players++
		if d.connected && (!tp_who_hides_dark || wizard || db.Fetch(d.player).flags & DARK == 0) && players != 0 && (user == nil || strings.Prefix(db.Fetch(d.player).name, user)) {
			if wizard {
				/* don't print flags, to save space */
				pbuf = fmt.Sprintf("%s(#%d) [%6d]", db.Fetch(d.player).name, d.player, db.Fetch(d.player).Location)
				if e.player != GOD {
					if db.Fetch(d.player).flags & INTERACTIVE != 0 {
						buf = fmt.Sprintf("%s %10s %4s*%c %s\r\n", pbuf, time_format_1(now - d.connected_at), time_format_2(now - d.last_time), secchar, d.hostname)
					} else {
						buf = fmt.Sprintf("%s %10s %4s %c %s\r\n", pbuf, time_format_1(now - d.connected_at), time_format_2(now - d.last_time), secchar, d.hostname)
					}
				} else {
					if db.Fetch(d.player).flags & INTERACTIVE != 0 {
						buf = fmt.Sprintf("%s %10s %4s*  %s(%s)\r\n", pbuf, time_format_1(now - d.connected_at), time_format_2(now - d.last_time), d.hostname, d.username)
					} else {
						buf = fmt.Sprintf("%s %10s %4s   %s(%s)\r\n", pbuf, time_format_1(now - d.connected_at), time_format_2(now - d.last_time), d.hostname, d.username)
					}
				}
			} else {
				if tp_who_doing {
					if db.Fetch(d.player).flags & INTERACTIVE != 0 {
						buf = fmt.Sprintf("%s %10s %4s*  %s\r\n", db.Fetch(d.player).name, time_format_1(now - d.connected_at), time_format_2(now - d.last_time), get_property_class(d.player, MESGPROP_DOING))
					} else {
						buf = fmt.Sprintf("%s %10s %4s   %s\r\n", db.Fetch(d.player).name, time_format_1(now - d.connected_at), time_format_2(now - d.last_time), get_property_class(d.player, MESGPROP_DOING))
					}
				} else {
					if db.Fetch(d.player).flags & INTERACTIVE != 0 {
						buf = fmt.Sprintf("%s %10s %4s* \r\n", db.Fetch(d.player).name, time_format_1(now - d.connected_at), time_format_2(now - d.last_time))
					} else {
						buf = fmt.Sprintf("%s %10s %4s  \r\n", db.Fetch(d.player).name, time_format_1(now - d.connected_at), time_format_2(now - d.last_time))
					}
				}
			}
			queue_msg(e, buf)
		}
	}
	if players > con_players_max {
		con_players_max = players
	}
	if players == 1 {
		queue_msg(e, fmt.Sprintf("1 player is connected.  (Max was %d)\r\n", con_players_max))
	} else {
		queue_msg(e, fmt.Sprintf("%d players are connected.  (Max was %d)\r\n", players, con_players_max))
	}
}

func time_format_1(dt long) (r string) {
	if delta := gmtime((time_t *) &dt); delta.tm_yday > 0 {
		r = fmt.Sprintf("%dd %02d:%02d", delta.tm_yday, delta.tm_hour, delta.tm_min)
	} else {
		r = fmt.Sprintf("%02d:%02d", delta.tm_hour, delta.tm_min)
	}
	return
}

func time_format_2(dt int) (r string) {
	switch delta := gmtime((time_t *) &dt); {
	case delta.tm_yday > 0:
		r = fmt.Sprintf("%dd", delta.tm_yday)
	case delta.tm_hour > 0:
		r = fmt.Sprintf("%dh", delta.tm_hour)
	case delta.tm_min > 0:
		r = fmt.Sprintf("%dm", delta.tm_min)
	default:
		r = fmt.Sprintf("%ds", delta.tm_sec)
	}
	return
}

func announce_puppets(player dbref, msg, prop string) {
	EachObject(func(what dbref, o *Object) {
		if IsThing(what) && o.flags & ZOMBIE != 0 {
			if o.Owner == player {
				where := o.Location
				if !Dark(where) && !Dark(player) && !Dark(what) {
					msg2 := msg
					if ptr := get_property_class(what, prop); ptr != "" {
						msg2 = ptr
					}
					notify_except(db.Fetch(where).Contents, what, fmt.Sprintf("%.512s %.3000s", o.name, msg2), what)
				}
			}
		}
	})
}

func announce_connect(descr int, player dbref) {
	dbref loc;
	char buf[BUFFER_LEN];
	dbref exit;

	if loc = db.Fetch(player).Location; loc == NOTHING {
		return
	}

	if !Dark(player) && !Dark(loc) {
		buf = fmt.Sprintf("%s has connected.", db.Fetch(player).name)
		notify_except(db.Fetch(loc).Contents, player, buf, player)
	}

	exit = NOTHING;
	if (online(player) == 1) {
		md := NewMatch(descr, player, "connect", IsExit)	/* match for connect */
		md.level = 1
		md.MatchAllExits()
		exit = md.MatchResult()
		if exit == AMBIGUOUS {
			exit = NOTHING
		}
	}

	if exit == NOTHING || db.Fetch(exit).flags & STICKY == 0 {
		if can_move(descr, player, tp_autolook_cmd, 1) {
			do_move(descr, player, tp_autolook_cmd, 1)
		} else {
			do_look_around(descr, player)
		}
	}


	/*
	 * See if there's a connect action.  If so, and the player is the first to
	 * connect, send the player through it.  If the connect action is set
	 * sticky, then suppress the normal look-around.
	 */

	if (exit != NOTHING)
		do_move(descr, player, "connect", 1);

	if (online(player) == 1) {
		announce_puppets(player, "wakes up.", "_/pcon");
	}

	/* queue up all _connect programs referred to by properties */
	envpropqueue(descr, player, db.Fetch(player).Location, NOTHING, player, NOTHING, "_connect", "Connect", 1, 1);
	envpropqueue(descr, player, db.Fetch(player).Location, NOTHING, player, NOTHING, "_oconnect", "Oconnect", 1, 0);
	ts_useobject(player);
	return;
}

func announce_disconnect(d *descriptor_data) {
	player := d.player
	char buf[BUFFER_LEN];
	int dcount;

	if loc := db.Fetch(player).Location; loc != NOTHING {
		if len(get_player_descrs(d.player)) < 2 && dequeue_prog(player, 2) {
			notify(player, "Foreground program aborted.")
		}

		if !Dark(player) && !Dark(loc) {
			notify_except(db.Fetch(loc).Contents, player, fmt.Sprintf("%s has disconnected.", db.Fetch(player).name), player)
		}

		/* trigger local disconnect action */
		if online(player) {
			if can_move(d.descriptor, player, "disconnect", 1) {
				do_move(d.descriptor, player, "disconnect", 1)
			}
			announce_puppets(player, "falls asleep.", "_/pdcon")
		}
		gui_dlog_closeall_descr(d.descriptor)

		d.connected = false
		d.player = NOTHING

	    forget_player_descr(player, d.descriptor)
	    update_desc_count_table()

		/* queue up all _connect programs referred to by properties */
		envpropqueue(d.descriptor, player, db.Fetch(player).Location, NOTHING, player, NOTHING, "_disconnect", "Disconnect", 1, 1)
		envpropqueue(d.descriptor, player, db.Fetch(player).Location, NOTHING, player, NOTHING, "_odisconnect", "Odisconnect", 1, 0)
		ts_lastuseobject(player)
		db.Fetch(player).flags |= OBJECT_CHANGED
	}
}

func do_setuid(name string) {
	var pw *passwd

	if pw = getpwnam(name); pw == nil {
		log_status("can't get pwent for %s", name)
		os.Exit(1)
	}
	if setuid(pw.pw_uid) == -1 {
		log_status("can't setuid(%d): ", pw.pw_uid)
		perror("setuid")
		os.Exit(1)
	}
}

func do_setgid(name string) {
	var gr *group
	if gr = getgrnam(name); gr == nil {
		log_status("can't get grent for group %s", name)
		os.Exit(1)
	}
	if setgid(gr.gr_gid) == -1 {
		log_status("can't setgid(%d): ", gr.gr_gid)
		perror("setgid")
		os.Exit(1)
	}
}

/***** O(1) Connection Optimizations *****/
func update_desc_count_table() {
	var c int
	current_descr_count = 0
	for d := descriptor_list; d != nil; d = d.next {
		if d.connected {
			d.con_number = c + 1
			descr_count_table[c] = d
			c++
			current_descr_count++
		}
	}
}

func descrdata_by_count(c int) (r *descriptor_data) {
	if c--; c > -1 && c < current_descr_count {
		r = descr_count_table[c]
	}
	return
}

func index_descr(index int) (r int) {
	switch {
	case index < 0, index >= FD_SETSIZE, descr_lookup_table[index] == nil:
		r = -1
	default:
		r = descr_lookup_table[index].descriptor
	}
	return
}

func get_player_descrs(player dbref) (r []int) {
	if Typeof(player) == TYPE_PLAYER {
		r = db.Fetch(player).(Player).descrs
	}
	return
}

func remember_player_descr(player dbref, descr int) {
	if Typeof(player) == TYPE_PLAYER {
		p := db.Fetch(player).(Player)
		p.descrs = append(p.descrs, descr)
	}
}

func forget_player_descr(player dbref, descr int) {
	if Typeof(player) == TYPE_PLAYER {
		p := db.Fetch(player).(Player)
		arr := p.descrs
		if len(arr) > 1 {
			var dest int
			for i, v := range arr {
				if v != descr {
					if i != dest {
						arr[dest] = v
					}
					dest++
				}
			}
			arr = arr[:dest]
		}
		p.descrs = arr
	}
}

func lookup_descriptor(c int) (r *descriptor_data) {
	if v > -1 && c < FD_SETSIZE {
		r = descr_lookup_table[c]
	}
	return
}

func online(player dbref) int {
	return db.Fetch(player).(Player).descrs
}

func pidle(c int) (r int) {
	if d := descrdata_by_count(c); d != nil {
		r = time.Now() - d.last_time
	} else {
		r = -1
	}
	return
}

func pdbref(int c) (r dbref) {
	d := descrdata_by_count(c); d != nil {
		r = d.player
	} else {
		r = NOTHING
	}
	return
}

func pontime(c int) (r int) {
	if d := descrdata_by_count(c); d != nil {
		r = time.Now() - d.connected_at
	} else {
		r = -1
	}
	return
}

/*** Foxen ***/
func least_idle_player_descr(who dbref) (r int) {
	var best_time int
	var best_d * descriptor_data
	for _, v := range get_player_descrs(who) {
		if d := lookup_descriptor(v); d != nil && (best_time == 0 || d.last_time > best_time)) {
			best_d = d
			best_time = d.last_time
		}
	}
	if best_d != nil {
		r = best_d.con_number
	}
	return
}

func most_idle_player_descr(who dbref) (r int) {
	var best_time int
	var best_d *descriptor_data
	for _, v := range get_player_descrs(who) {
		if d := lookup_descriptor(v); d != nil && (!best_time || d.last_time < best_time) {
			best_d = d
			best_time = d.last_time
		}
	}
	if best_d != nil {
		r = best_d.con_number
	}
	return
}

func pboot(c int) {
	if d := descrdata_by_count(c); d != nil {
		process_output(d)
		d.booted = 1
	}
}

func pdescrboot(c int) int {
    if d := lookup_descriptor(c); d != nil {
		process_output(d)
		d.booted = 1
		/* shutdownsock(d) */
		return 1
    }
	return 0
}


func pnotify(c int, msg string) {
	if d := descrdata_by_count(c); d != nil {
		queue_msg(d, msg)
		d.QueueWrite("\r\n")
	}
}

func pdescr(c int) (r int) {
	if d := descrdata_by_count(c); d != nil {
		r = d.descriptor
	} else {
		r = -1
	}
	return
}

func pdescrcon(c int) (r int) {
    if d := lookup_descriptor(c); d != nil {
		r = d.con_number
	}
	return
}

func dbref_first_descr(c dbref) (r int) {
	if arr := get_player_descrs(c); len(arr) > 0 {
		r = arr[len(arr) - 1]
	} else {
		r = -1
	}
	return
}

func descr_mcpframe(c int) (r *McpFrame) {
	if d := lookup_descriptor(c); d != nil {
		r = &d.mcpframe
	}
	return
}

func partial_pmatch(name string) (last dbref) {
	last = NOTHING
	for d := descriptor_list; d != nil; d = d.next {
		if d.connected && last != d.player && strings.Prefix(db.Fetch(d.player).name, name)) {
			if last != NOTHING {
				last = AMBIGUOUS
				break
			}
			last = d.player
		}
	}
	return
}

func welcome_user(d *descriptor_data) {
	if f := fopen(WELC_FILE, "rb"); f == nil {
		queue_msg(d, DEFAULT_WELCOME_MESSAGE)
		perror("spit_file: welcome.txt")
	} else {
		var buf string
		for fgets(buf, sizeof(buf) - 3, f) {
			ptr := strchr(buf, '\n')
			if (ptr && ptr > buf && *(ptr - 1) != '\r') {
				*ptr++ = '\r';
				*ptr++ = '\n';
				*ptr++ = '\0';
			}
			queue_msg(d, buf)
		}
		fclose(f)
	}
	switch {
	case wizonly_mode:
		queue_msg(d, "## The game is currently in maintenance mode, and only wizards will be able to connect.\r\n")
	case tp_playermax && con_players_curr >= tp_playermax_limit && tp_playermax_warnmesg != "":
		queue_msg(d, tp_playermax_warnmesg)
		d.QueueWrite("\r\n")
	}
}

func dump_status() {
	var now time_t
	time(&now)
	log_status("STATUS REPORT:");
	for d := descriptor_list; d; d = d.next {
		var buf string
		if d.connected {
			buf = fmt.Sprintf("PLAYING descriptor %d player %s(%d) from host %s(%s), %s.\n", d.descriptor, db.Fetch(d.player).name, d.player, d.hostname, d.username, d.last_time ? "idle %d seconds" : "never used")
		} else {
			buf = fmt.Sprintf("CONNECTING descriptor %d from host %s(%s), %s.\n", d.descriptor, d.hostname, d.username, d.last_time ? "idle %d seconds" : "never used")
		}
		log_status(buf, now - d.last_time)
	}
}

/* Ignore support -- Could do with moving into its own file */

func ignore_is_ignoring_sub(player, who dbref) bool {
	switch {
	case !tp_ignore_support, !valid_reference(player), !valid_reference(who):
		return false
	}

	player = db.Fetch(player).Owner
	who = db.Fetch(who).Owner

	/* You can't ignore yourself, or an unquelled wizard, */
	/* and unquelled wizards can ignore no one. */
	switch {
	case player == who, Wizard(player), Wizard(who):
		return false
	case db.Fetch(player).(Player).ignore_last == AMBIGUOUS:
		return false
	/* Ignore the last player ignored without bothering to look them up */
	case db.Fetch(player).(Player).ignore_last == who:
		return true
	case db.Fetch(player).(Player).ignore_cache == nil && !ignore_prime_cache(player):
		return false
	}

	top := 0
	bottom := len(db.Fetch(player).(Player).ignore_cache) - 1
	list := db.Fetch(player).(Player).ignore_cache

	for top < bottom {
		middle := top + (bottom - top) / 2
		switch {
		case list[middle] == who:
			break
		case list[middle] < who:
			top = middle + 1
		default:
			bottom = middle
		}
	}
	if top >= bottom {
		return false
	}
	db.Fetch(player).(Player).ignore_last = who
	return true
}

int ignore_is_ignoring(dbref Player, dbref Who)
{
	return ignore_is_ignoring_sub(Player, Who) || (tp_ignore_bidirectional && ignore_is_ignoring_sub(Who, Player));
}

static int ignore_dbref_compare(const void* Lhs, const void* Rhs)
{
	return *(dbref*)Lhs - *(dbref*)Rhs;
}

func ignore_prime_cache(player dbref) bool {
	switch {
	case !tp_ignore_support, !valid_reference(player), !IsPlayer(player):
		return false
	}

	txt := strings.TrimLeftFunc(get_property_class(player, IGNORE_PROP), unicode.IsSpace)
	if txt == "" {
		db.Fetch(player).(Player).ignore_last = AMBIGUOUS
		return false
	}

	var list []dbref
	for ptr := txt; ptr != ""; ptr = strings.TrimLeftFunc(ptr, unicode.IsSpace) {
		if ptr[0] == NUMBER_TOKEN {
			ptr = ptr[1:]
		}
		if isdigit(ptr[0]) {
			list = append(list, strconv.Atoi(ptr))
		} else {
			list = append(list, NOTHING)
		}
		ptr = strings.TrimLeftFunc(ptr, func(r rune) bool {
			return !unicode.IsSpace(r)
		}
		if i := strings.IndexFunc(ptr, unicode.IsSpace); i != -1 {
			ptr = ptr[i:]
		} else {
			ptr = ""
		}
	}

	qsort(list, len(list), sizeof(dbref), ignore_dbref_compare)
	db.Fetch(player).(Player).ignore_cache = list
	return true
}

func ignore_flush_cache(player dbref) {
	if valid_reference(player) && IsPlayer(player) {
		db.Fetch(player).(Player).ignore_cache = nil
		db.Fetch(player).(Player).ignore_last = NOTHING
	}
}

func ignore_flush_all_cache() {
	/* Don't touch the database if it's not been loaded yet... */
	if db != nil {
		EachObject(func(obj dbref, o object) {
			if IsPlayer(obj) {
				p := o.(Player)
				p.ignore_cache = nil
				p.ignore_last = NOTHING
			}
		})
	}
}

func ignore_add_player(player, who dbref) {
	switch {
	case !tp_ignore_support, !valid_reference(player), !valid_reference(who):
	default:
		reflist_add(db.Fetch(player).Owner, IGNORE_PROP, db.Fetch(who).Owner)
		ignore_flush_cache(db.Fetch(player).Owner)
	}
}

func ignore_remove_player(player, who dbref) {
	switch {
	case !tp_ignore_support, !valid_reference(player), !valid_reference(who):
	default:
		reflist_del(db.Fetch(player).Owner, IGNORE_PROP, db.Fetch(who).Owner)
		ignore_flush_cache(db.Fetch(player).Owner)
	}
}

func ignore_remove_from_all_players(player dbref) {
	if tp_ignore_support {
		EachObject(func(obj dbref) {
			if IsPlayer(obj) {
				reflist_del(obj, IGNORE_PROP, Player)
			}
		})
		ignore_flush_all_cache()
	}
}