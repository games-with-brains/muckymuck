package fbmuck

import "github.com/feyeleanor/slices"

struct tuning_parameter {
	group string
	security int
	label string
}

var (
	tp_dumpwarn_mesg = DUMPWARN_MESG
	tp_deltawarn_mesg = DELTAWARN_MESG
	tp_dumpdeltas_mesg = DUMPDELTAS_MESG
	tp_dumping_mesg = DUMPING_MESG
	tp_dumpdone_mesg = DUMPDONE_MESG
	tp_penny = PENNY
	tp_pennies = PENNIES
	tp_cpenny = CPENNY
	tp_cpennies = CPENNIES
	tp_muckname = MUCKNAME
	tp_huh_mesg = HUH_MESSAGE
	tp_leave_mesg = LEAVE_MESSAGE
	tp_idle_mesg = IDLEBOOT_MESSAGE
	tp_register_mesg = REG_MSG
	tp_playermax_warnmesg = PLAYERMAX_WARNMESG
	tp_playermax_bootmesg = PLAYERMAX_BOOTMESG
	tp_autolook_cmd = AUTOLOOK_CMD
	tp_proplist_counter_fmt = PROPLIST_COUNTER_FORMAT
	tp_proplist_entry_fmt = PROPLIST_ENTRY_FORMAT
	tp_ssl_keyfile_passwd = SSL_KEYFILE_PASSWD
	tp_pcreate_flags = PCREATE_FLAGS
	tp_reserved_names = RESERVED_NAMES
	tp_reserved_player_names = RESERVED_PLAYER_NAMES
)

struct tuning_str_entry {
	tuning_parameter
	str *string
	isnullable bool
	isdefault bool
}

var tune_str_table = map[string] *tune_str_entry{
	"autolook_cmd": 			&tune_str_entry{ "Commands", &tp_autolook_cmd, 0, "Room entry look command", 0, 1 },
	"penny":					&tune_str_entry{ "Currency", &tp_penny, 0, "Currency name", 0, 1 },
	"pennies":					&tune_str_entry{ "Currency", &tp_pennies, 0, "Currency name, plural", 0, 1 },
	"cpenny":					&tune_str_entry{ "Currency", &tp_cpenny, 0, "Currency name, capitalized", 0, 1 },
	"cpennies":					&tune_str_entry{ "Currency", &tp_cpennies, 0, "Currency name, capitalized, plural", 0, 1 },
	"dumpwarn_mesg":			&tune_str_entry{ "DB Dumps", &tp_dumpwarn_mesg, 0, "Full dump warning mesg", 1, 1 },
	"deltawarn_mesg":			&tune_str_entry{ "DB Dumps", &tp_deltawarn_mesg, 0, "Delta dump warning mesg", 1, 1 },
	"dumping_mesg":				&tune_str_entry{ "DB Dumps", &tp_dumping_mesg, 0, "Full dump start mesg", 1, 1 },
	"dumpdeltas_mesg":			&tune_str_entry{ "DB Dumps", &tp_dumpdeltas_mesg, 0, "Delta dump start mesg", 1, 1 },
	"dumpdone_mesg":			&tune_str_entry{ "DB Dumps", &tp_dumpdone_mesg, 0, "Dump completion message", 1, 1 },
	"idle_boot_mesg":			&tune_str_entry{ "Idle Boot", &tp_idle_mesg, 0, "Boot message for idling out", 0, 1 },
	"playermax_warnmesg":		&tune_str_entry{ "Player Max", &tp_playermax_warnmesg, 0, "Max. players login warning", 0, 1 },
	"playermax_bootmesg":		&tune_str_entry{ "Player Max", &tp_playermax_bootmesg, 0, "Max. players boot message", 0, 1 },
	"proplist_counter_fmt":		&tune_str_entry{ "Properties", &tp_proplist_counter_fmt, 0, "Proplist counter name format", 0, 1 },
	"proplist_entry_fmt":		&tune_str_entry{ "Properties", &tp_proplist_entry_fmt, 0, "Proplist entry name format", 0, 1 },
	"register_mesg":			&tune_str_entry{ "Registration", &tp_register_mesg, 0, "Login registration mesg", 0, 1 },
	"muckname":					&tune_str_entry{ "Misc", &tp_muckname, 0, "Muck name", 0, 1 },
	"leave_mesg":				&tune_str_entry{ "Misc", &tp_leave_mesg, 0, "Logoff message", 0, 1 },
	"huh_mesg":					&tune_str_entry{ "Misc", &tp_huh_mesg, 0, "Command unrecognized warning", 0, 1 },
	"ssl_keyfile_passwd":		&tune_str_entry{ "SSL", &tp_ssl_keyfile_passwd, MLEV_GOD, "Password for SSL keyfile", 1, 1 },
	"pcreate_flags":			&tune_str_entry{ "Database", &tp_pcreate_flags, 0, "Initial Player Flags", 1, 1 },
	"reserved_names":			&tune_str_entry{ "Database", &tp_reserved_names, 0, "Reserved names smatch", 1, 1 },
	"reserved_player_names":	&tune_str_entry{ "Database", &tp_reserved_player_names, 0, "Reserved player names smatch", 1, 1 },
}

var (
	tp_dump_interval = DUMP_INTERVAL
	tp_dump_warntime = DUMP_WARNTIME
	tp_monolithic_interval = MONOLITHIC_INTERVAL
	tp_clean_interval = CLEAN_INTERVAL
	tp_aging_time = AGING_TIME
	tp_maxidle = MAXIDLE
	tp_idle_ping_time = IDLE_PING_TIME
)

struct tune_time_entry {
	tuning_parameter
	tim *int
}

var tune_time_table = map[string] *tune_time_entry {
	"aging_time":			&tune_time_entry{ "Database", &tp_aging_time, 0, "When to considered an object old and unused" },
	"dump_interval":		&tune_time_entry{ "DB Dumps", &tp_dump_interval, 0, "Interval between delta dumps" },
	"dump_warntime":		&tune_time_entry{ "DB Dumps", &tp_dump_warntime, 0, "Interval between warning and dump" },
	"monolithic_interval":	&tune_time_entry{ "DB Dumps", &tp_monolithic_interval, 0, "Interval between full dumps" },
	"maxidle":				&tune_time_entry{ "Idle Boot", &tp_maxidle, 0, "Maximum idle time before booting" },
	"idle_ping_time":		&tune_time_entry{ "Idle Boot", &tp_idle_ping_time, 0, "Server side keepalive time in seconds" },
	"clean_interval":		&tune_time_entry{ "Tuning", &tp_clean_interval, 0, "Interval between memory cleanups." },
}

var (
	tp_max_object_endowment = MAX_OBJECT_ENDOWMENT
	tp_object_cost = OBJECT_COST
	tp_exit_cost = EXIT_COST
	tp_link_cost = LINK_COST
	tp_room_cost = ROOM_COST
	tp_lookup_cost = LOOKUP_COST
	tp_max_pennies = MAX_PENNIES
	tp_penny_rate = PENNY_RATE
	tp_start_pennies = START_PENNIES
	tp_kill_base_cost = KILL_BASE_COST
	tp_kill_min_cost = KILL_MIN_COST
	tp_kill_bonus = KILL_BONUS
	tp_command_burst_size = COMMAND_BURST_SIZE
	tp_commands_per_time = COMMANDS_PER_TIME
	tp_command_time_msec = COMMAND_TIME_MSEC
	tp_max_output = MAX_OUTPUT
	tp_max_delta_objs = MAX_DELTA_OBJS
	tp_max_force_level = MAX_FORCE_LEVEL
	tp_max_process_limit = MAX_PROCESS_LIMIT
	tp_max_plyr_processes = MAX_PLYR_PROCESSES
	tp_max_instr_count = MAX_INSTR_COUNT
	tp_max_ml4_preempt_count = MAX_ML4_PREEMPT_COUNT
	tp_instr_slice = INSTR_SLICE
	tp_mpi_max_commands = MPI_MAX_COMMANDS
	tp_pause_min = PAUSE_MIN
	tp_listen_mlev = LISTEN_MLEV
	tp_playermax_limit = PLAYERMAX_LIMIT
	tp_process_timer_limit = PROCESS_TIMER_LIMIT
	tp_cmd_log_threshold_msec = CMD_LOG_THRESHOLD_MSEC
	tp_userlog_mlev = USERLOG_MLEV
	tp_mcp_muf_mlev = MCP_MUF_MLEV
	tp_movepennies_muf_mlev = MOVEPENNIES_MUF_MLEV
	tp_addpennies_muf_mlev = ADDPENNIES_MUF_MLEV
	tp_pennies_muf_mlev = PENNIES_MUF_MLEV
)

struct tune_val_entry {
	tuning_parameter
	val *int
}

var tune_val_table[] = map[string] *tune_val_entry {
	"max_object_endowment":		&tune_val_entry{ "Costs", &tp_max_object_endowment, 0, "Max value of object" },
	"object_cost":				&tune_val_entry{ "Costs", &tp_object_cost, 0, "Cost to create thing" },
	"exit_cost":				&tune_val_entry{ "Costs", &tp_exit_cost, 0, "Cost to create exit" },
	"link_cost":				&tune_val_entry{ "Costs", &tp_link_cost, 0, "Cost to link exit" },
	"room_cost":				&tune_val_entry{ "Costs", &tp_room_cost, 0, "Cost to create room" },
	"lookup_cost":				&tune_val_entry{ "Costs", &tp_lookup_cost, 0, "Cost to lookup playername" },
	"max_pennies":				&tune_val_entry{ "Currency", &tp_max_pennies, 0, "Player currency cap" },
	"penny_rate":				&tune_val_entry{ "Currency", &tp_penny_rate, 0, "Moves between finding currency, avg" },
	"start_pennies": 			&tune_val_entry{ "Currency", &tp_start_pennies, 0, "Player starting currency count" },
	"kill_base_cost":			&tune_val_entry{ "Killing", &tp_kill_base_cost, 0, "Cost to guarentee kill" },
	"kill_min_cost":			&tune_val_entry{ "Killing", &tp_kill_min_cost, 0, "Min cost to kill" },
	"kill_bonus":				&tune_val_entry{ "Killing", &tp_kill_bonus, 0, "Bonus to killed player" },
	"kill_bonus":				&tune_val_entry{ "Listeners", &tp_listen_mlev, 0, "Mucker Level required for Listener progs" },
	"cmd_log_threshold_msec":	&tune_val_entry{ "Logging", &tp_cmd_log_threshold_msec, 0, "Log commands that take longer than X millisecs" },
	"max_force_level":			&tune_val_entry{ "Misc", &tp_max_force_level, MLEV_GOD, "Maximum number of forces processed within a command" },
	"max_process_limit":		&tune_val_entry{ "MUF", &tp_max_process_limit, 0, "Max concurrent processes on system" },
	"max_plyr_processes": 		&tune_val_entry{ "MUF", &tp_max_plyr_processes, 0, "Max concurrent processes per player" },
	"max_instr_count":			&tune_val_entry{ "MUF", &tp_max_instr_count, 0, "Max MUF instruction run length for ML1" },
	"max_ml4_preempt_count":	&tune_val_entry{ "MUF", &tp_max_ml4_preempt_count, 0, "Max MUF preempt instruction run length for ML4, (0 = no limit)" },
	"instr_slice": 				&tune_val_entry{ "MUF", &tp_instr_slice, 0, "Instructions run per timeslice" },
	"process_timer_limit":		&tune_val_entry{ "MUF", &tp_process_timer_limit, 0, "Max timers per process" },
	"mcp_muf_mlev":				&tune_val_entry{ "MUF", &tp_mcp_muf_mlev, 0, "Mucker Level required to use MCP" },
	"userlog_mlev":				&tune_val_entry{ "MUF", &tp_userlog_mlev, 0, "Mucker Level required to write to userlog" },
	"movepennies_muf_mlev":		&tune_val_entry{ "MUF", &tp_movepennies_muf_mlev, 0, "Mucker Level required to move pennies non-destructively" },
	"addpennies_muf_mlev":		&tune_val_entry{ "MUF", &tp_addpennies_muf_mlev, 0, "Mucker Level required to create/destroy pennies" },
	"pennies_muf_mlev":			&tune_val_entry{ "MUF", &tp_pennies_muf_mlev, 0, "Mucker Level required to read the value of pennies, settings above 1 disable {money}" },
	"mpi_max_commands":			&tune_val_entry{ "MPI", &tp_mpi_max_commands, 0, "Max MPI instruction run length" },
	"playermax_limit":			&tune_val_entry{ "Player Max", &tp_playermax_limit, 0, "Max player connections allowed" },
	"command_burst_size": 		&tune_val_entry{ "Spam Limits", &tp_command_burst_size, 0, "Commands before limiter engages" },
	"commands_per_time":		&tune_val_entry{ "Spam Limits", &tp_commands_per_time, 0, "Commands allowed per time period" },
	"command_time_msec": 		&tune_val_entry{ "Spam Limits", &tp_command_time_msec, 0, "Millisecs per spam limiter time period" },
	"max_output":				&tune_val_entry{ "Spam Limits", &tp_max_output, 0, "Max output buffer size" },
	"pause_min":				&tune_val_entry{ "Tuning", &tp_pause_min, 0, "Min ms to pause between MUF timeslices" },
	"max_delta_objs":			&tune_val_entry{ "Tuning", &tp_max_delta_objs, 0, "Percentage changed objects to force full dump" },
}

var (
	tp_player_start dbref = PLAYER_START
	tp_default_room_parent dbref = GLOBAL_ENVIRONMENT
)

struct tune_ref_entry {
	tuning_parameter
	typ int
	ref *dbref
}

var tune_ref_table = map[string] *tune_ref_entry {
	"default_room_parent":	&tune_ref_entry{ "Database", TYPE_ROOM, &tp_default_room_parent, 0, "Place to parent new rooms to" },
	"player_start":			&tune_ref_entry{ "Database", TYPE_ROOM, &tp_player_start, 0, "Place where new players start" },
}

var (
	tp_hostnames = HOSTNAMES
	tp_log_commands = LOG_COMMANDS
	tp_log_failed_commands = LOG_FAILED_COMMANDS
	tp_log_programs = LOG_PROGRAMS
	tp_dbdump_warning = DBDUMP_WARNING
	tp_deltadump_warning = DELTADUMP_WARNING
	tp_dumpdone_warning = DUMPDONE_WARNING
	tp_periodic_program_purge = PERIODIC_PROGRAM_PURGE
	tp_secure_who = SECURE_WHO
	tp_who_doing = WHO_DOING
	tp_realms_control = REALMS_CONTROL
	tp_listeners = LISTENERS
	tp_listeners_obj = LISTENERS_OBJ
	tp_listeners_env = LISTENERS_ENV
	tp_zombies = ZOMBIES
	tp_wiz_vehicles = WIZ_VEHICLES
	tp_force_mlev1_name_notify = FORCE_MLEV1_NAME_NOTIFY
	tp_restrict_kill = RESTRICT_KILL
	tp_registration = REGISTRATION
	tp_teleport_to_player = TELEPORT_TO_PLAYER
	tp_secure_teleport = SECURE_TELEPORT
	tp_exit_darking = EXIT_DARKING
	tp_thing_darking = THING_DARKING
	tp_dark_sleepers = DARK_SLEEPERS
	tp_who_hides_dark = WHO_HIDES_DARK
	tp_compatible_priorities = COMPATIBLE_PRIORITIES
	tp_do_mpi_parsing = DO_MPI_PARSING
	tp_look_propqueues = LOOK_PROPQUEUES
	tp_lock_envcheck = LOCK_ENVCHECK
	tp_idleboot = IDLEBOOT
	tp_playermax = PLAYERMAX
	tp_allow_home = ALLOW_HOME
	tp_enable_prefix = ENABLE_PREFIX
	tp_enable_match_yield = ENABLE_MATCH_YIELD
	tp_thing_movement = SECURE_THING_MOVEMENT
	tp_expanded_debug = EXPANDED_DEBUG_TRACE
	tp_proplist_int_counter = PROPLIST_INT_COUNTER
	tp_log_interactive = LOG_INTERACTIVE
	tp_lazy_mpi_istype_perm = LAZY_MPI_ISTYPE_PERM
	tp_optimize_muf = OPTIMIZE_MUF
	tp_ignore_support = IGNORE_SUPPORT
	tp_ignore_bidirectional = IGNORE_BIDIRECTIONAL
	tp_verbose_clone = VERBOSE_CLONE
	tp_muf_comments_strict = MUF_COMMENTS_STRICT
	tp_starttls_allow = STARTTLS_ALLOW
	tp_m3_huh = M3_HUH
	tp_7bit_thing_names = ASCII_THING_NAMES
	tp_7bit_other_names = ASCII_OTHER_NAMES
	tp_idle_ping_enable = IDLE_PING_ENABLE
	tp_recognize_null_command = RECOGNIZE_NULL_COMMAND
)

struct tune_bool_entry {
	tuning_parameter
	boolval *int
}

var tune_bool_table = map[string] *tune_bool_entry {
	"enable_home":				&tune_bool_entry{ "Commands", &tp_allow_home, 4, "Enable 'home' command"},
	"enable_prefix":			&tune_bool_entry{ "Commands", &tp_enable_prefix, 4, "Enable prefix actions"},
	"enable_match_yield":		&tune_bool_entry{ "Commands", &tp_enable_match_yield, 4, "Enable yield/overt flags on rooms and things"},
	"verbose_clone":			&tune_bool_entry{ "Commands", &tp_verbose_clone, 4, "Verbose @clone command"},
	"recognize_null_command":	&tune_bool_entry{ "Commands", &tp_recognize_null_command, 4, "Recognize null command"},
	"exit_darking":				&tune_bool_entry{ "Dark", &tp_exit_darking, 0, "Allow setting exits dark"},
	"thing_darking":			&tune_bool_entry{ "Dark", &tp_thing_darking, 0, "Allow setting things dark"},
	"dark_sleepers":			&tune_bool_entry{ "Dark", &tp_dark_sleepers, 0, "Make sleeping players dark"},
	"who_hides_dark":			&tune_bool_entry{ "Dark", &tp_who_hides_dark, 4, "Hide dark players from WHO list"},
	"realms_control":			&tune_bool_entry{ "Database", &tp_realms_control, 0, "Enable Realms control"},
	"compatible_priorities":	&tune_bool_entry{ "Database", &tp_compatible_priorities, 0, "Use legacy exit priority levels on things"},
	"dbdump_warning":			&tune_bool_entry{ "DB Dumps", &tp_dbdump_warning, 0, "Enable warning messages for full DB dumps"},
	"deltadump_warning":		&tune_bool_entry{ "DB Dumps", &tp_deltadump_warning, 0, "Enable warning messages for delta DB dumps"},
	"dumpdone_warning":			&tune_bool_entry{ "DB Dumps", &tp_dumpdone_warning, 0, "Enable notification of DB dump completion"},
	"idleboot":					&tune_bool_entry{ "Idle Boot", &tp_idleboot, 0, "Enable booting of idle players"},
	"idle_ping_enable":			&tune_bool_entry{ "Idle Boot", &tp_idle_ping_enable, 0, "Enable server side keepalive"},
	"restrict_kill":			&tune_bool_entry{ "Killing", &tp_restrict_kill, 0, "Restrict kill command to players set Kill_OK"},
	"allow_listeners":			&tune_bool_entry{ "Listeners", &tp_listeners, 0, "Enable programs to listen to player output"},
	"allow_listeners_obj":		&tune_bool_entry{ "Listeners", &tp_listeners_obj, 0, "Allow listeners on things"},
	"allow_listeners_env":		&tune_bool_entry{ "Listeners", &tp_listeners_env, 0, "Allow listeners down environment"},
	"log_commands":				&tune_bool_entry{ "Logging", &tp_log_commands, 4, "Enable logging of player commands"},
	"log_failed_commands":		&tune_bool_entry{ "Logging", &tp_log_failed_commands, 4, "Enable logging of unrecognized commands"},
	"log_interactive":			&tune_bool_entry{ "Logging", &tp_log_interactive, 4, "Enable logging of text sent to MUF"},
	"log_programs":				&tune_bool_entry{ "Logging", &tp_log_programs, 4, "Log programs every time they are saved"},
	"teleport_to_player":		&tune_bool_entry{ "Movement", &tp_teleport_to_player, 0, "Allow teleporting to a player"},
	"secure_teleport":			&tune_bool_entry{ "Movement", &tp_secure_teleport, 0, "Restrict actions to Jump_OK or controlled rooms"},
	"secure_thing_movement":	&tune_bool_entry{ "Movement", &tp_thing_movement, 4, "Moving things act like player"},
	"do_mpi_parsing":			&tune_bool_entry{ "MPI", &tp_do_mpi_parsing, 0, "Enable parsing of mesgs for MPI"},
	"lazy_mpi_istype_perm":		&tune_bool_entry{ "MPI", &tp_lazy_mpi_istype_perm, 0, "Enable looser legacy perms for MPI {istype}"},
	"optimize_muf":				&tune_bool_entry{ "MUF", &tp_optimize_muf, 0, "Enable MUF bytecode optimizer"},
	"expanded_debug_trace":		&tune_bool_entry{ "MUF", &tp_expanded_debug, 0, "MUF debug trace shows array contents"},
	"force_mlev1_name_notify":	&tune_bool_entry{ "MUF", &tp_force_mlev1_name_notify, 0, "MUF notify prepends username at ML1"},
	"muf_comments_strict":		&tune_bool_entry{ "MUF", &tp_muf_comments_strict, 0, "MUF comments are strict and not recursive"},
	"playermax":				&tune_bool_entry{ "Player Max", &tp_playermax, 0, "Limit number of concurrent players allowed"},
	"look_propqueues":			&tune_bool_entry{ "Properties", &tp_look_propqueues, 0, "When a player looks, trigger _look/ propqueues"},
	"lock_envcheck":			&tune_bool_entry{ "Properties", &tp_lock_envcheck, 0, "Locks check environment for properties"},
	"proplist_int_counter":		&tune_bool_entry{ "Properties", &tp_proplist_int_counter, 0, "Proplist counter uses an integer property"},
	"registration":				&tune_bool_entry{ "Registration", &tp_registration, 0, "Require new players to register manually"},
	"periodic_program_purge":	&tune_bool_entry{ "Tuning", &tp_periodic_program_purge, 0, "Periodically free unused MUF programs"},
	"use_hostnames":			&tune_bool_entry{ "WHO", &tp_hostnames, 0, "Resolve IP addresses into hostnames"},
	"secure_who":				&tune_bool_entry{ "WHO", &tp_secure_who, 0, "Disallow WHO command from login screen and programs"},
	"who_doing":				&tune_bool_entry{ "WHO", &tp_who_doing, 0, "Show '_/do' property value in WHO lists"},
	"allow_zombies":			&tune_bool_entry{ "Misc", &tp_zombies, 0, "Enable Zombie things to relay what they hear"},
	"wiz_vehicles":				&tune_bool_entry{ "Misc", &tp_wiz_vehicles, 0, "Only let Wizards set vehicle bits"},
	"ignore_support":			&tune_bool_entry{ "Misc", &tp_ignore_support, 3, "Enable support for @ignoring players"},
	"ignore_bidirectional":		&tune_bool_entry{ "Misc", &tp_ignore_bidirectional, 3, "Enable bidirectional @ignore"},
	"m3_huh":					&tune_bool_entry{ "Misc", &tp_m3_huh, 3, "Enable huh? to call an exit named \"huh?\" and set M3, with full command string"},
	"starttls_allow":			&tune_bool_entry{ "SSL", &tp_starttls_allow, 3, "Enable TELNET STARTTLS encryption on plaintext port"},
	"7bit_thing_names":			&tune_bool_entry{ "Charset", &tp_7bit_thing_names, 4, "Thing names may contain only 7-bit characters"},
	"7bit_other_names":			&tune_bool_entry{ "Charset", &tp_7bit_other_names, 4, "Exit/room/muf names may contain only 7-bit characters"},
}


static const char *
timestr_full(long dtime)
{
	static char buf[32];
	int days, hours, minutes, seconds;

	days = dtime / 86400;
	dtime %= 86400;
	hours = dtime / 3600;
	dtime %= 3600;
	minutes = dtime / 60;
	seconds = dtime % 60;

	buf = fmt.Sprintf("%3dd %2d:%02d:%02d", days, hours, minutes, seconds)

	return buf;
}

func tune_count_parms() int {
	return len(tune_str_table) + len(tune_time_table) + len(tune_val_table) + len(tune_ref_table) + len(tune_bool_table)
}

func tune_display_parms(player dbref, name string, security int) {
	for k, v := range tune_str_table {
		switch {
		case v.security > security:
		case name == "" || !smatch(name, k)
			notify(player, fmt.Sprintf("(str)  %-20s = %.4096s", k, *(v.str)))
		}
	}

	for k, v := range tune_time_table {
		switch {
		case v.security > security:
		case name == "" || !smatch(name, k)
			notify(player, fmt.Sprintf("(time) %-20s = %d", k, timestr_full(*vtim)))
		}
	}

	for k, v := range tune_val_table {
		switch {
		case v.security > security:
		case name == "" || !smatch(name, k)
			notify(player, fmt.Sprintf("(time) %-20s = %d", k, *(tval.val)))
		}
	}

	for k, v := range tune_ref_table {
		switch {
		case v.security > security:
		case name == "" || !smatch(name, k)
			notify(player, fmt.Sprintf("(time) %-20s = %d", k, unparse_object(player, *v.ref)))
		}
	}

	for k, v := range tune_ref_table {
		switch {
		case v.security > security:
		case name == "" || !smatch(name, k)
			if *(tbool.boolval) {
				notify(player, fmt.Sprintf("(bool) %-20s = yes", k))
			} else {
				notify(player, fmt.Sprintf("(bool) %-20s = no", k))
			}
		}
	}
	notify(player, "*done*")
}

func tune_save_parms_to_file(f *FILE) {
	for k, v := range tune_str_table {
		fprintf(f, "%s=%.4096s\n", k, (*v.str))
	}

	for k, v := range tune_time_table {
		fprintf(f, "%s=%s\n", k, timestr_full(*v.tim))
	}

	for k, v := range tune_val_table {
		fprintf(f, "%s=%s\n", k, timestr_full(*v.val))
	}

	for k, v := range tune_ref_table {
		fprintf(f, "%s=#%d\n", k, *(v.ref))
	}

	for k, v := range tune_bool_table {
		if *(v.boolval) {
			fprintf(f, "%s=yes\n", k)
		} else {
			fprintf(f, "%s=no\n", k)
		}
	}
}

func tune_parms_array(pattern string, mlev int) (r Array) {
	for name, tbool := range tune_bool_table {
		if tbool.security <= mlev {
			if pattern == "" || !smatch(pattern, name) {
				item := Dictionary{
					"type": "boolean",
					"group": tbool.group,
					"name":  name
					"mlev":  tbool.security,
					"label": tbool.label,
				}
				if tbool.boolval {
					item["value"] = 1
				} else {
					item["value"] = 0
				}
				r = append(r, item)
			}
		}
	}

	for name, ttim := range tune_time_table {
		if ttim.security <= mlev {
			if pattern == "" || !smatch(pattern, name) {
				r = append(r, Dictionary{
					"type": "timespan",
					"group": ttim.group,
					"name":  name,
					"value": *(ttim.tim),
					"mlev":  ttim.security,
					"label": ttim.label,
				})
			}
		}
	}

	for name, tval := range tune_val_table {
		if tval.security <= mlev {
			if pattern == "" || !smatch(pattern, name) {
				r = append(r, Dictionary{
					"type": "integer",
					"group": tval.group,
					"name":  name,
					"value": *(tval.val),
					"mlev":  tval.security,
					"label": tval.label,
				})
			}
		}
	}

	for name, tref := range tune_ref_table {
		if tref.security <= mlev {
			if pattern == "" || !smatch(pattern, name) {
				item := Dictionary{
					"type": "dbref",
					"group": tref.group,
					"name":  name,
					"value": *(tref.ref),
					"mlev":  tref.security,
					"label": tref.label,
				}
				switch tref.typ {
				case NOTYPE:
					item["objtype"] = "any"
				case TYPE_PLAYER:
					item["objtype"] = "player"
				case TYPE_THING:
					item["objtype"] = "thing"
				case TYPE_ROOM:
					item["objtype"] = "room"
				case TYPE_EXIT:
					item["objtype"] = "exit"
				case TYPE_PROGRAM:
					item["objtype"] = "program"
				default:
					item["objtype"] = "unknown"
				}
				r = append(r, item)
			}
		}
	}

	for name, tstr := range tune_str_table {
		if tstr.security <= mlev {
			if pattern == "" || !smatch(pattern, name) {
				r = append(r, Dictionary{
					"type": "string",
					"group": tstr.group,
					"name":  tstr.name,
					"value": *(tstr.str),
					"mlev":  tstr.security,
					"label": tstr.label,
				})
			}
		}
	}
	return
}

func tune_save_parmsfile(void) {
	if f := fopen(PARMFILE_NAME, "wb"); f == nil {
		log_status("Couldn't open file %s!", PARMFILE_NAME)
	} else {
		tune_save_parms_to_file(f)
		fclose(f)
	}
}

func tune_get_parmstring(name string, mlev int) string {
	if tstr, ok := tune_str_table[parmname]; ok {
		if tstr.security <= mlev {
			r = *(tstr.str)
		}
		return
	}

	if ttim, ok := tune_time_table[parmname]; ok {
		if ttim.security <= mlev {
			r = fmt.Sprint(*(ttim.tim))
		}
		return
	}

	if tval, ok := tune_val_table[parmname]; ok {
		if ttim.security <= mlev {
			r = fmt.Sprint(*(tval.val))
		}
		return
	}

	if tref, ok := tune_ref_table[parmname]; ok {
		if ttim.security <= mlev {
			r = fmt.Sprint(*(tref.ref))
		}
		return
	}

	if tbool, ok := tune_bool_table[parmname]; ok {
		if ttim.security <= mlev {
			if *(tbool.boolval) {
				r = "yes"
			} else {
				r = "no"
			}
		}
		return
	}
	return
}

func tune_freeparms() {
	if tstr, ok := tune_str_table[parmname]; ok {
		if !tstr.isdefault {
			*(tstr.str) = ""
		}
	}
}

func tune_setparm(parmname, val string, security int) (r int) {
	parmval := val
	r = TUNESET_UNKNOWN

	if tstr, ok := tune_str_table[parmname]; ok {
		switch {
		case tstr.security > security:
			r = TUNESET_DENIED
		case !tstr.isnullable && parmval == "":
			r = TUNESET_BADVAL
		default:
			if parmval[0] == '-' {
				parmval = parmval[1:]
			}
			*(tstr.str) = parmval
			tstr.isdefault = false
			r = TUNESET_SUCCESS
		}
		return
	}

	for name, ttim := range tune_time_table {
		if parmname == name {
			if ttim.security > security {
				r = TUNESET_DENIED
			} else {
				var days, hrs, mins, secs, result int
				char *end;

				end = parmval + len(parmval) - 1;
				switch *end {
				case 's', 'S':
					*end = '\0'
					if !unicode.IsNumber(parmval) {
						r = TUNESET_SYNTAX
					} else {
						secs = strconv.Atoi(parmval)
					}
				case 'm', 'M':
					*end = '\0';
					if !unicode.IsNumber(parmval) {
						return TUNESET_SYNTAX
					} else {
						mins = strconv.Atoi(parmval)
					}
				case 'h', 'H':
					*end = '\0'
					if !unicode.IsNumber(parmval) {
						r = TUNESET_SYNTAX
					} else {
						hrs = strconv.Atoi(parmval)
					}
				case 'd', 'D':
					*end = '\0'
					if !unicode.IsNumber(parmval) {
						r = TUNESET_SYNTAX
					} else {
						days = strconv.Atoi(parmval)
					}
				default:
					if result = sscanf(parmval, "%dd %2d:%2d:%2d", &days, &hrs, &mins, &secs); result != 4 {
						r = TUNESET_SYNTAX
					}
				}
				ttim.tim = (days * 86400) + (3600 * hrs) + (60 * mins) + secs
				r = TUNESET_SUCCESS
			}
			return
		}
	}

	for name, tval := range tune_val_table {
		if parmname == name {
			switch {
			case tval.security > security:
				r = TUNESET_DENIED
			case !unicode.IsNumber(parmval):
				r = TUNESET_SYNTAX
			default:
				tval.val = strconv.Atoi(parmval)
				r = TUNESET_SUCCESS
			}
			return
		}
	}

	for name, tref := range tune_ref_table {
		if parmname == name {
			switch {
			case tref.security > security:
				r = TUNESET_DENIED
			case parmval[0] != NUMBER_TOKEN:
				r = TUNESET_SYNTAX
			case !unicode.IsNumber(parmval[1]):
				r = TUNESET_SYNTAX
			default:
				if obj := strconv.Atoi(parmval[1:]); obj < 0 || obj >= db_top {
					r = TUNESET_SYNTAX
				} else {
					switch tref.(type) {
					case NOTYPE, Typeof(obj):
						r = TUNESET_BADVAL
					default:
						*tref.ref = obj
						r = TUNESET_SUCCESS
					}
				}
			}
			return
		}
	}

	for name, tbool := range tune_bool_table {
		if parmname == name {
			switch {
			case tbool.security > security:
				r = TUNESET_DENIED
			case parmval == 'y', parmval == 'Y':
				tbool.boolval = true
				r = TUNESET_SUCCESS
			case parmval == 'n', parmval == 'N':
				tbool.boolval = false
				r = TUNESET_SUCCESS
			default:
				r = TUNESET_SYNTAX
			}
			return
		}
	}
	return
}

func tune_load_parms_from_file(f *FILE, player dbref, cnt int) {
	for result := 0; !feof(f) && (cnt < 0 || cnt != 0); cnt-- {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			switch line := scanner.Text(); {
			case line == "":
			case line[0] != "#":
				if terms := strings.SplitN(line, "=", 2); len(terms) == 2 {
					term[0] = strings.TrimSpace(term[0])
					term[1] = strings.TrimSpace(term[1])
					switch result = tune_setparm(term[0], term[1], MLEV_GOD); result {
					case TUNESET_SUCCESS:
						line += ": Parameter set."
					case TUNESET_UNKNOWN:
						line += ": Unknown parameter."
					case TUNESET_SYNTAX:
						line += ": Bad parameter syntax."
					case TUNESET_BADVAL:
						line += ": Bad parameter value."
					case TUNESET_DENIED:
						line += ": Permission denied."
					}
					if result != 0 && player != NOTHING {
						notify(player, line)
					}
				}
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading tuning configuration file:", err)
		}
	}
}

func tune_load_parmsfile(player dbref) {
	f := fopen(PARMFILE_NAME, "rb")
	if (!f) {
		log_status("Couldn't open file %s!", PARMFILE_NAME)
		return
	}
	tune_load_parms_from_file(f, player, -1)
	fclose(f)
}

func do_tune(player dbref, parmname, parmval string, full_command_has_delimiter int) {
	if Wizard(player) {
		var security int
		if player == GOD {
			security = MLEV_GOD
		} else {
			security = MLEV_WIZARD
		}
		switch {
		case parmname != "" && full_command_has_delimiter {
			if force_level {
				notify(player, "You cannot force setting a @tune.")
				return;
			} else {
		 		oldvalue := tune_get_parmstring(parmname, security)
				switch result := tune_setparm(parmname, parmval, security); result {
				case TUNESET_SUCCESS:
					log_status("TUNED: %s(%d) tuned %s from '%s' to '%s'", db.Fetch(player).name, player, parmname, oldvalue, parmval)
					notify(player, "Parameter set.")
					tune_display_parms(player, parmname, security)
				case TUNESET_UNKNOWN:
					notify(player, "Unknown parameter.")
				case TUNESET_SYNTAX:
					notify(player, "Bad parameter syntax.")
				case TUNESET_BADVAL:
					notify(player, "Bad parameter value.")
				case TUNESET_DENIED:
					notify(player, "Permission denied.")
				}
			}
		case parmname != "":
			switch parname {
			case "save":
				tune_save_parmsfile()
				notify(player, "Saved parameters to configuration file.")
			case "load":
				tune_load_parmsfile(player)
				notify(player, "Restored parameters from configuration file.")
			default:
				tune_display_parms(player, parmname, security);
			}
		case parmval == "":
			tune_display_parms(player, parmname, security)
		default:
			notify(player, "But what do you want to tune?")
		}
	} else {
		notify(player, "You pull out a harmonica and play a short tune.")
	}
}