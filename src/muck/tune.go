package fbmuck

import "github.com/feyeleanor/slices"

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
	tp_dump_interval = DUMP_INTERVAL
	tp_dump_warntime = DUMP_WARNTIME
	tp_monolithic_interval = MONOLITHIC_INTERVAL
	tp_clean_interval = CLEAN_INTERVAL
	tp_aging_time = AGING_TIME
	tp_maxidle = MAXIDLE
	tp_idle_ping_time = IDLE_PING_TIME
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
	tp_player_start ObjectID = PLAYER_START
	tp_default_room_parent ObjectID = GLOBAL_ENVIRONMENT
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

type tuning_entry struct {
	group string
	label string
	MLevel
	variable interface{}
	isnullable bool
	isdefault bool
}

type TuningTable map[string] *tuning_entry

func (t TuningTable) GetAs(security int, name string) (r string) {
	if t, ok := t[name]; ok {
		if t.MLevel <= security {
			r = fmt.Sprintf("%v", t.variable)
		}
	}
	return
}

func (t TuningTable) SetAs(security MLevel, name, val string) (r int) {
	r = TUNESET_UNKNOWN
	if t, ok := t[name]; ok {
		if t.MLevel > security {
			r = TUNESET_DENIED
		} else {
			switch v := t.variable.(type) {
			case *string:
				if !t.isnullable && val == "":
					r = TUNESET_BADVAL
				default:
					if val[0] == '-' {
						val = val[1:]
					}
					*v = val
					t.isdefault = false
					r = TUNESET_SUCCESS
				}
			case *time.Time:
				var days, hrs, mins, secs, result int
				switch val[len(val) - 1] {
				case 's', 'S':
					if !unicode.IsNumber(al) {
						r = TUNESET_SYNTAX
					} else {
						secs = strconv.Atoi(val)
					}
				case 'm', 'M':
					if !unicode.IsNumber(val) {
						return TUNESET_SYNTAX
					} else {
						mins = strconv.Atoi(val)
					}
				case 'h', 'H':
					if !unicode.IsNumber(val) {
						r = TUNESET_SYNTAX
					} else {
						hrs = strconv.Atoi(val)
					}
				case 'd', 'D':
					if !unicode.IsNumber(val) {
						r = TUNESET_SYNTAX
					} else {
						days = strconv.Atoi(val)
					}
				default:
					if result = sscanf(val, "%dd %2d:%2d:%2d", &days, &hrs, &mins, &secs); result != 4 {
						r = TUNESET_SYNTAX
					}
				}
				*v = (days * 86400) + (hrs * 3600) + (mins * 60) + secs
					r = TUNESET_SUCCESS
				}
			case *int:
				if !unicode.IsNumber(val[0]) {
					r = TUNESET_SYNTAX
				 } else {
					*v = strconv.Atoi(val)
					r = TUNESET_SUCCESS
				}
			case *ObjectID:
				switch {
				case val[0] != NUMBER_TOKEN:
					r = TUNESET_SYNTAX
				case !unicode.IsNumber(val[1]):
					r = TUNESET_SYNTAX
				default:
					if obj := strconv.Atoi(val[1:]); !obj.IsValid() {
						r = TUNESET_SYNTAX
					} else {
						r = TUNESET_SUCCESS
					}
				}
			case *bool:
				switch val {
				case 'y', 'Y':
					*v = true
					r = TUNESET_SUCCESS
				case 'n', 'N':
					*v = false
					r = TUNESET_SUCCESS
				default:
					r = TUNESET_SYNTAX
				}
			}
		}
	}
	return
}


type TuningParamType int
const (
	COMMANDS = iota
	COSTS
	CURRENCY
	DARK
	DB_DUMPS
	DATABASE
	IDLE_BOOT
	KILLING
	LISTENERS
	LOGGING
	MISC
	MPI
	MUF
	PROPERTIES
	PLAYER_MAX
	REGISTRATION
	SPAM_LIMITS
	SSL
	TUNING
)

func RestrictedTuningParameter(level int, t TuningParamType, variable interface{}, label string) (r *tuning_entry) {
	switch t {
	case CHARSET:
		r = &tuning_entry{ "Charset", param, level, label, false, true }
	case COMMANDS:
		r = &tuning_entry{ "Commands", param, level, label, false, true }
	case COSTS:
		r = &tuning_entry{ "Costs", param, level, label, false, true }
	case CURRENCY:
		r = &tuning_entry{ "Currency", param, level, label, false, true }
	case DARK:
		r = &tuning_entry{ "Dark", param, level, label, false, true }
	case DB_DUMPS:
		r = &tuning_entry{ "DB Dumps", param, level, label, true, true }
	case DATABASE:
		r = &tuning_entry{ "Database", param, level, label, true, true }
	case IDLE_BOOT:
		r = &tuning_entry{ "Idle Boot", param, level, label, false, true }
	case KILLING:
		r = &tuning_entry{ "Killing", param, level, label, false, true }
	case LISTENERS:
		r = &tuning_entry{ "Listeners", param, level, label, false, true }
	case LOGGING:
		r = &tuning_entry{ "Logging", param, level, label, false, true }
	case MISC:
		r = &tuning_entry{ "Misc", param, level, label, false, true }
	case MOVEMENT:
		r = &tuning_entry{ "Movement", param, level, label, false, true }
	case MPI:
		r = &tuning_entry{ "MPI", param, level, label, false, true }
	case MUF:
		r = &tuning_entry{ "MUF", param, level, label, false, true }
	case PROPERTIES:
		r = &tuning_entry{ "Properties", param, level, label, false, true }
	case PLAYER_MAX:
		r = &tuning_entry{ "Player Max", param, level, label, false, true }
	case REGISTRATION:
		r = &tuning_entry{ "Registration", param, level, label, false, true }
	case SPAM_LIMITS:
		r = &tuning_entry{ "Spam Limits", param, level, label, false, true }
	case SSL:
		r = &tuning_entry{ "SSL", param, MLEV_GOD, label, true, true }
	case TUNING:
		r = &tuning_entry{ "Tuning", param, level, label, false, true },
	case WHO:
		r = &tuning_entry{ "WHO", param, level, label, false, true },
	}
	return
}

func TuningParameter(t TuningParamType, variable interface{}, label string) *tuning_entry {
	return RestrictedTuningParameter(0, t, variable, label)
}

var Tuneables = TuningTable {
	"7bit_thing_names":			RestrictedTuningParameter(MLEV_GOD, CHARSET, &tp_7bit_thing_names, "Thing names may contain only 7-bit characters"),
	"7bit_other_names":			RestrictedTuningParameter(MLEV_GOD, CHARSET, &tp_7bit_other_names, "Exit/room/muf names may contain only 7-bit characters"),

	"autolook_cmd": 			TuningParameter(COMMANDS, &tp_autolook_cmd, "Room entry look command"),
	"enable_home":				RestrictedTuningParameter(MLEV_GOD, COMMANDS, &tp_allow_home, "Enable 'home' command"),
	"enable_prefix":			RestrictedTuningParameter(MLEV_GOD, COMMANDS, &tp_enable_prefix, "Enable prefix actions"),
	"enable_match_yield":		RestrictedTuningParameter(MLEV_GOD, COMMANDS, &tp_enable_match_yield, "Enable yield/overt flags on rooms and things"),
	"verbose_clone":			RestrictedTuningParameter(MLEV_GOD, COMMANDS, &tp_verbose_clone, "Verbose @clone command"),
	"recognize_null_command":	RestrictedTuningParameter(MLEV_GOD, COMMANDS, &tp_recognize_null_command, "Recognize null command"),

	"max_object_endowment":		TuningParameter(COSTS, &tp_max_object_endowment, "Max value of object"),
	"object_cost":				TuningParameter(COSTS, &tp_object_cost, "Cost to create thing"),
	"exit_cost":				TuningParameter(COSTS, &tp_exit_cost, "Cost to create exit"),
	"link_cost":				TuningParameter(COSTS, &tp_link_cost, "Cost to link exit"),
	"room_cost":				TuningParameter(COSTS, &tp_room_cost, "Cost to create room"),
	"lookup_cost":				TuningParameter(COSTS, &tp_lookup_cost, "Cost to lookup playername"),

	"penny":					TuningParameter(CURRENCY, &tp_penny, "Currency name"),
	"pennies":					TuningParameter(CURRENCY, &tp_pennies, "Currency name, plural"),
	"cpenny":					TuningParameter(CURRENCY, &tp_cpenny, "Currency name, capitalized"),
	"cpennies":					TuningParameter(CURRENCY, &tp_cpennies, "Currency name, capitalized, plural"),
	"max_pennies":				TuningParameter(CURRENCY, &tp_max_pennies, "Player currency cap"),
	"penny_rate":				TuningParameter(CURRENCY, &tp_penny_rate, "Moves between finding currency, avg"),
	"start_pennies": 			TuningParameter(CURRENCY, &tp_start_pennies, "Player starting currency count"),

	"exit_darking":				TuningParameter(DARK, &tp_exit_darking, "Allow setting exits dark"),
	"thing_darking":			TuningParameter(DARK, &tp_thing_darking, "Allow setting things dark"),
	"dark_sleepers":			TuningParameter(DARK, &tp_dark_sleepers, "Make sleeping players dark"),
	"who_hides_dark":			RestrictedTuningParameter(MLEV_GOD, DARK, &tp_who_hides_dark, "Hide dark players from WHO list"),

	"realms_control":			TuningParameter(DATABASE, &tp_realms_control, "Enable Realms control"),
	"compatible_priorities":	TuningParameter(DATABASE, &tp_compatible_priorities, "Use legacy exit priority levels on things"),
	"pcreate_flags":			TuningParameter(DATABASE, &tp_pcreate_flags, "Initial Player Flags"),
	"reserved_names":			TuningParameter(DATABASE, &tp_reserved_names, "Reserved names smatch"),
	"reserved_player_names":	TuningParameter(DATABASE, &tp_reserved_player_names, "Reserved player names smatch"),
	"aging_time":				TuningParameter(DATABASE, &tp_aging_time, "When to considered an object old and unused"),
	"default_room_parent":		TuningParameter(DATABASE, &tp_default_room_parent, "Place to parent new rooms to"),
	"player_start":				TuningParameter(DATABASE, &tp_player_start, "Place where new players start"),

	"dumpwarn_mesg":			TuningParameter(DB_DUMPS, &tp_dumpwarn_mesg, "Full dump warning mesg"),
	"deltawarn_mesg":			TuningParameter(DB_DUMPS, &tp_deltawarn_mesg, "Delta dump warning mesg"),
	"dumping_mesg":				TuningParameter(DB_DUMPS, &tp_dumping_mesg, "Full dump start mesg"),
	"dumpdeltas_mesg":			TuningParameter(DB_DUMPS, &tp_dumpdeltas_mesg, "Delta dump start mesg"),
	"dumpdone_mesg":			TuningParameter(DB_DUMPS, &tp_dumpdone_mesg, "Dump completion message"),
	"dump_interval":			TuningParameter(DB_DUMPS, &tp_dump_interval, "Interval between delta dumps"),
	"dump_warntime":			TuningParameter(DB_DUMPS, &tp_dump_warntime, "Interval between warning and dump"),
	"monolithic_interval":		TuningParameter(DB_DUMPS, &tp_monolithic_interval, "Interval between full dumps"),
	"dbdump_warning":			TuningParameter(DB_DUMPS, &tp_dbdump_warning, "Enable warning messages for full DB dumps"),
	"deltadump_warning":		TuningParameter(DB_DUMPS, &tp_deltadump_warning, "Enable warning messages for delta DB dumps"),
	"dumpdone_warning":			TuningParameter(DB_DUMPS, &tp_dumpdone_warning, "Enable notification of DB dump completion"),

	"idle_boot_mesg":			TuningParameter(IDLE_BOOT, &tp_idle_mesg, "Boot message for idling out"),
	"maxidle":					TuningParameter(IDLE_BOOT, &tp_maxidle, "Maximum idle time before booting"),
	"idle_ping_time":			TuningParameter(IDLE_BOOT, &tp_idle_ping_time, "Server side keepalive time in seconds"),
	"idleboot":					TuningParameter(IDLE_BOOT, &tp_idleboot, "Enable booting of idle players"),
	"idle_ping_enable":			TuningParameter(IDLE_BOOT, &tp_idle_ping_enable, "Enable server side keepalive"),

	"restrict_kill":			TuningParameter(KILLING, &tp_restrict_kill, "Restrict kill command to players set Kill_OK"),
	"kill_base_cost":			TuningParameter(KILLING, &tp_kill_base_cost, "Cost to guarentee kill"),
	"kill_min_cost":			TuningParameter(KILLING, &tp_kill_min_cost, "Min cost to kill"),
	"kill_bonus":				TuningParameter(KILLING, &tp_kill_bonus, "Bonus to killed player"),

	"listen_mlev":				TuningParameter(LISTENERS, &tp_listen_mlev, "Mucker Level required for Listener progs"),
	"allow_listeners":			TuningParameter(LISTENERS, &tp_listeners, "Enable programs to listen to player output"),
	"allow_listeners_obj":		TuningParameter(LISTENERS, &tp_listeners_obj, "Allow listeners on things"),
	"allow_listeners_env":		TuningParameter(LISTENERS, &tp_listeners_env, "Allow listeners down environment"),

	"cmd_log_threshold_msec":	TuningParameter(LOGGING, &tp_cmd_log_threshold_msec, "Log commands that take longer than X millisecs"),
	"log_commands":				RestrictedTuningParameter(MLEV_GOD, LOGGING, &tp_log_commands, "Enable logging of player commands"),
	"log_failed_commands":		RestrictedTuningParameter(MLEV_GOD, LOGGING, &tp_log_failed_commands, "Enable logging of unrecognized commands"),
	"log_interactive":			RestrictedTuningParameter(MLEV_GOD, LOGGING, &tp_log_interactive, "Enable logging of text sent to MUF"),
	"log_programs":				RestrictedTuningParameter(MLEV_GOD, LOGGING, &tp_log_programs, "Log programs every time they are saved"),

	"max_force_level":			RestrictedTuningParameter(MLEV_GOD, MISC, &tp_max_force_level, "Maximum number of forces processed within a command"),
	"muckname":					TuningParameter(MISC, &tp_muckname, "Muck name"),
	"leave_mesg":				TuningParameter(MISC, &tp_leave_mesg, "Logoff message"),
	"huh_mesg":					TuningParameter(MISC, &tp_huh_mesg, "Command unrecognized warning"),
	"allow_zombies":			TuningParameter(MISC, &tp_zombies, "Enable Zombie things to relay what they hear"),
	"wiz_vehicles":				TuningParameter(MISC, &tp_wiz_vehicles, "Only let Wizards set vehicle bits"),
	"ignore_support":			RestrictedTuningParameter(MLEV_MASTER, MISC, &tp_ignore_support, "Enable support for @ignoring players"),
	"ignore_bidirectional":		RestrictedTuningParameter(MLEV_MASTER, MISC, &tp_ignore_bidirectional, "Enable bidirectional @ignore"),
	"m3_huh":					RestrictedTuningParameter(MLEV_MASTER, MISC, &tp_m3_huh, "Enable huh? to call an exit named \"huh?\" and set M3, with full command string"),

	"teleport_to_player":		TuningParameter(MOVEMENT, &tp_teleport_to_player, "Allow teleporting to a player"),
	"secure_teleport":			TuningParameter(MOVEMENT, &tp_secure_teleport, "Restrict actions to Jump_OK or controlled rooms"),
	"secure_thing_movement":	RestrictedTuningParameter(MLEV_GOD, MOVEMENT, &tp_thing_movement, "Moving things act like player"),

	"mpi_max_commands":			TuningParameter(MPI, &tp_mpi_max_commands, "Max MPI instruction run length"),
	"do_mpi_parsing":			TuningParameter(MPI, &tp_do_mpi_parsing, "Enable parsing of mesgs for MPI"),
	"lazy_mpi_istype_perm":		TuningParameter(MPI, &tp_lazy_mpi_istype_perm, "Enable looser legacy perms for MPI {istype}"),

	"max_process_limit":		TuningParameter(MUF, &tp_max_process_limit, "Max concurrent processes on system" },
	"max_plyr_processes": 		TuningParameter(MUF, &tp_max_plyr_processes, "Max concurrent processes per player" },
	"max_instr_count":			TuningParameter(MUF, &tp_max_instr_count, "Max MUF instruction run length for ML1" },
	"max_ml4_preempt_count":	TuningParameter(MUF, &tp_max_ml4_preempt_count, "Max MUF preempt instruction run length for ML4, (0 = no limit)" },
	"instr_slice": 				TuningParameter(MUF, &tp_instr_slice, "Instructions run per timeslice" },
	"process_timer_limit":		TuningParameter(MUF, &tp_process_timer_limit, "Max timers per process" },
	"mcp_muf_mlev":				TuningParameter(MUF, &tp_mcp_muf_mlev, "Mucker Level required to use MCP" },
	"userlog_mlev":				TuningParameter(MUF, &tp_userlog_mlev, "Mucker Level required to write to userlog" },
	"movepennies_muf_mlev":		TuningParameter(MUF, &tp_movepennies_muf_mlev, "Mucker Level required to move pennies non-destructively" },
	"addpennies_muf_mlev":		TuningParameter(MUF, &tp_addpennies_muf_mlev, "Mucker Level required to create/destroy pennies" },
	"pennies_muf_mlev":			TuningParameter(MUF, &tp_pennies_muf_mlev, "Mucker Level required to read the value of pennies, settings above 1 disable {money}" },
	"optimize_muf":				TuningParameter(MUF, &tp_optimize_muf, "Enable MUF bytecode optimizer"),
	"expanded_debug_trace":		TuningParameter(MUF, &tp_expanded_debug, "MUF debug trace shows array contents"),
	"force_mlev1_name_notify":	TuningParameter(MUF, &tp_force_mlev1_name_notify, "MUF notify prepends username at ML1"),
	"muf_comments_strict":		TuningParameter(MUF, &tp_muf_comments_strict, "MUF comments are strict and not recursive"),

	"playermax_warnmesg":		TuningParameter(PLAYER_MAX, &tp_playermax_warnmesg, "Max. players login warning"),
	"playermax_bootmesg":		TuningParameter(PLAYER_MAX, &tp_playermax_bootmesg, "Max. players boot message"),
	"playermax_limit":			TuningParameter(PLAYER_MAX, &tp_playermax_limit, "Max player connections allowed"),
	"playermax":				TuningParameter(PLAYER_MAX, &tp_playermax, "Limit number of concurrent players allowed"),

	"proplist_counter_fmt":		TuningParameter(PROPERTIES, &tp_proplist_counter_fmt, "Proplist counter name format"),
	"proplist_entry_fmt":		TuningParameter(PROPERTIES, &tp_proplist_entry_fmt, "Proplist entry name format"),
	"look_propqueues":			TuningParameter(PROPERTIES, &tp_look_propqueues, "When a player looks, trigger _look/ propqueues"},
	"lock_envcheck":			TuningParameter(PROPERTIES, &tp_lock_envcheck, "Locks check environment for properties"},
	"proplist_int_counter":		TuningParameter(PROPERTIES, &tp_proplist_int_counter, "Proplist counter uses an integer property"},

	"register_mesg":			TuningParameter(REGISTRATION, &tp_register_mesg, "Login registration mesg"),
	"registration":				TuningParameter(REGISTRATION, &tp_registration, "Require new players to register manually"),

	"command_burst_size": 		TuningParameter(SPAM_LIMITS, &tp_command_burst_size, "Commands before limiter engages"),
	"commands_per_time":		TuningParameter(SPAM_LIMITS, &tp_commands_per_time, "Commands allowed per time period"),
	"command_time_msec": 		TuningParameter(SPAM_LIMITS, &tp_command_time_msec, "Millisecs per spam limiter time period"),
	"max_output":				TuningParameter(SPAM_LIMITS, &tp_max_output, "Max output buffer size"),

	"ssl_keyfile_passwd":		RestrictedTuningParameter(MLEV_GOD, SSL, &tp_ssl_keyfile_passwd, "Password for SSL keyfile"),
	"starttls_allow":			RestrictedTuningParameter(MLEV_MASTER, SSL, &tp_starttls_allow, "Enable TELNET STARTTLS encryption on plaintext port"),

	"clean_interval":			TuningParameter(TUNING, &tp_clean_interval, "Interval between memory cleanups."),
	"pause_min":				TuningParameter(TUNING, &tp_pause_min, "Min ms to pause between MUF timeslices"),
	"max_delta_objs":			TuningParameter(TUNING, &tp_max_delta_objs, "Percentage changed objects to force full dump"),
	"periodic_program_purge":	TuningParameter(TUNING, &tp_periodic_program_purge, "Periodically free unused MUF programs"),


	"use_hostnames":			TuningParameter(WHO, &tp_hostnames, "Resolve IP addresses into hostnames"),
	"secure_who":				TuningParameter(WHO, &tp_secure_who, "Disallow WHO command from login screen and programs"),
	"who_doing":				TuningParameter(WHO, &tp_who_doing, "Show '_/do' property value in WHO lists"),
}

func (t TuningTable) Display(security int, player ObjectID, name string, security int) {
	for k, v := range t {
		switch {
		case v.MLevel > security:
		case name == "", smatch(name, k) == 0:
			notify(player, fmt.Sprintf("%v = %v", k, *(v.variable)))
		}
	}
	notify(player, "*done*")
}

func (t TuningTable) SaveTo(f *FILE) {
	for k, v := range t {
		switch v := v.variable.(type) {
		case *string:
			fmt.Fprintf(f, "%v=%v\n", k, *v)
		case *int:
			fmt.Fprintf(f, "%v=%v\n", k, *v)
		case *bool:
			fmt.Fprintf(f, "%v=%v\n", k, *v)
		case *time.Time:
			fmt.Fprintf(f, "%v=%v\n", k, *v)
		case *ObjectID:
			fmt.Fprintf(f, "%v=%v\n", k, *v)
		}
	}
}

func (t TuningTable) ArrayAs(security int, name string) (r Array) {
	for k, v := range t {
		switch {
		case v.MLevel > security:
		case name == "", smatch(name, k) == 0:
			item := Dictionary{
				"group": v.group,
				"name": v.name,
				"value": *v.variable,
				"mlev": v.MLevel,
				"label": v.label,
			}
			switch v := v.variable.(type) {
			case *bool:
				item["type"] ="boolean"
				if *v {
					item["value"] = 1
				} else {
					item["value"] = 0
				}
			case *time.Time:
				item["type"] = "timespan"
			case *int:
				item["type"] = "integer"
			case *string:
				item["type"] = "string"
			case *ObjectID:
				item["type"] = "ObjectID"
				switch Typeof(*v) {
				case NOTYPE:
					item["objtype"] = "any"
				case Player:
					item["objtype"] = "player"
				case Object:
					item["objtype"] = "thing"
				case Room:
					item["objtype"] = "room"
				case Exit:
					item["objtype"] = "exit"
				case Program:
					item["objtype"] = "program"
				case Lock:
					item["objtype"] = "lock"
				default:
					item["objtype"] = "unknown"
				}
			}
			r = append(r, item)
		}
	}
	return
}

func (t TuningTable) Save(name string) {
	if f, e := os.OpenFile(name, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0755); e == nil {
		t.SaveTo(f)
		f.Close()
	} else {
		log_status("Couldn't open file %s!", PARMFILE_NAME)
	}
}

func (t TuningTable) LoadFrom(f *FILE, player ObjectID, cnt int) {
	var e error
	for scanner := bufio.NewScanner(f); scanner.Scan() && cnt != 0; {
		switch line := scanner.Text(); {
		case line == "":
		case line[0] != "#":
			if terms := strings.SplitN(line, "=", 2); len(terms) == 2 {
				term[0] = strings.TrimSpace(term[0])
				term[1] = strings.TrimSpace(term[1])
				switch result := Tuneables.SetAs(MLEV_GOD, term[0], term[1]); result {
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
		if cnt > 0 {
			cnt--
		}
	}		
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading tuning configuration file:", err)
	}
}

func (t TuningTable) Load(name string, player ObjectID) {
	if f, e := os.Open(name); e == nil {
		defer func() {
			if e := f.Close(); e != nil {
				log.Fatal(e)
			}
		}()
		Tuneables.LoadFrom(f, player, -1)
	} else {
		log_status("Couldn't open file %s!", PARMFILE_NAME)
	}
}

func (t TuningTable) Tune(player ObjectID, name string, val interface{}, has_delimiter bool) {
	if Wizard(player) {
		var security int
		if player == GOD {
			security = MLEV_GOD
		} else {
			security = MLEV_WIZARD
		}
		switch {
		case name != "" && has_delimiter {
			if force_level > 0 {
				notify(player, "You cannot force setting a @tune.")
			} else {
		 		oldvalue := t.GetAs(security, name)
				switch result := t.SetAs(security, name, val); result {
				case TUNESET_SUCCESS:
					log_status("TUNED: %s(%d) tuned %s from '%s' to '%s'", DB.Fetch(player).name, player, name, oldvalue, val)
					notify(player, "Parameter set.")
					Tuneables.Display(player, name, security)
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
		case name == "save":
			t.Save(PARMFILE_NAME)
			notify(player, "Saved parameters to configuration file.")
		case name == "load":
			t.Load(PARMFILE_NAME, player)
			notify(player, "Restored parameters from configuration file.")
		case name != "", val == "":
			t.Display(player, name, security)
		default:
			notify(player, "But what do you want to tune?")
		}
	} else {
		notify(player, "You pull out a harmonica and play a short tune.")
	}
}