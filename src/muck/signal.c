package fbmuck

# define our_signal(s,f) signal((s),(f))

func set_dumper_signals() {
	signal.Ignore(signal.SIGPIPE)		//	Ignore Blocked Pipe
	signal.Ignore(signal.SIGHUP)		//	Ignore Terminal Hangup
	signal.Ignore(signal.SIGCHLD)		//	Ignore Child termination
	signal.Ignore(signal.SIGFPE)		//	Ignore FP exceptions
	signal.Ignore(signal.SIGUSR1)		//	Ignore SIGUSR1
	signal.Ignore(signal.SIGUSR2)		//	Ignore SIGUSR2
	signal.Reset(signal.SIGINT)			//	Take Interrupt signal and die!
	signal.Reset(signal.SIGTERM)		//	Take Terminate signal and die!
	signal.Reset(signal.SIGSEGV)		//	Take Segfault and die!
	signal.Reset(signal.SIGTRAP)
	signal.Reset(signal.SIGIOT)
	signal.Reset(signal.SIGEMT)
	signal.Reset(signal.SIGBUS)
	signal.Reset(signal.SIGSYS)
	signal.Ignore(signal.SIGXCPU)		//	CPU usage limit exceeded
	signal.Ignore(signal.SIGXFSZ)		//	Exceeded file size limit
	signal.Reset(signal.SIGVTALRM)
}

/*
 * set_signals()
 * set_sigs_intern(bail)
 *
 * Traps a bunch of signals and reroutes them to various
 * handlers. Mostly bailout.
 *
 * If called from bailout, then reset all to default.
 *
 * Called from main() and bailout()
 */
func set_sigs_intern(bail bool) {
	if bail {
		signal.Reset(signal.SIGPIPE)
		signal.Reset(signal.SIGHUP)
		signal.Reset(signal.SIGINT)
		signal.Reset(signal.SIGTERM)
		signal.Reset(signal.SIGTRAP)
		signal.Reset(signal.SIGIOT)
		signal.Reset(signal.SIGEMT)
		signal.Reset(signal.SIGBUS)
		signal.Reset(signal.SIGSYS)
		signal.Reset(signal.SIGFPE)
		signal.Reset(signal.SIGSEGV)
		signal.Reset(signal.SIGTERM)
		signal.Reset(signal.SIGXCPU)
		signal.Reset(signal.SIGXFSZ)
		signal.Reset(signal.SIGVTALRM)
#ifdef SIGEMERG
		our_signal(SIGUSR2, sig_emerg)
#else
		signal.Reset(signal.SIGUSR2)
#endif
		signal.Reset(signal.SIGUSR1)
	} else {
		our_signal(SIGPIPE, bailout)
		our_signal(SIGHUP, bailout)
		our_signal(SIGINT, bailout)
		our_signal(SIGTERM, bailout)
		our_signal(SIGTRAP, bailout)
		our_signal(SIGIOT, bailout)
		our_signal(SIGEMT, bailout)
		our_signal(SIGBUS, bailout)
		our_signal(SIGSYS, bailout)
		our_signal(SIGFPE, bailout)
		our_signal(SIGSEGV, bailout)
		our_signal(SIGTERM, sig_shutdown)
		our_signal(SIGXCPU, bailout)
		our_signal(SIGXFSZ, bailout)
		our_signal(SIGVTALRM, bailout)
#ifdef SIGEMERG
		our_signal(SIGUSR2, sig_emerg)
#else
		our_signal(SIGUSR2, bailout)
#endif
		our_signal(SIGUSR1, sig_dump_status)
	}
}

func set_signals() {
	set_sigs_intern(false)
}

func bailout(sig int) RETSIGTYPE {
	/* turn off signals */
	set_sigs_intern(true)
	message := fmt.Sprintf("BAILOUT: caught signal %d", sig)
	panic(message)
	exit(7)
	return 0
}

//	Spew WHO to file
func sig_dump_status(i int) RETSIGTYPE {
	dump_status()
	return 0
}

func sig_emerg(i int) RETSIGTYPE {
	wall_and_flush("\nEmergency signal received ! (power failure ?)\nThe database will be saved.\n")
	dump_database()
	shutdown_flag = true
	restart_flag = false
	return 0
}

func sig_shutdown(i int) RETSIGTYPE {
	log_status("SHUTDOWN: via SIGNAL")
	shutdown_flag = true
	restart_flag = false
	return 0
}