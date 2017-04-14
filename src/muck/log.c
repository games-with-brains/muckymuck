package fbmuck

func log2file(filename string, format string, v ...interface{}) {
	if fp, err := fopen(filename, "ab"); err != nil {
		fmt.Fprintf(stderr, "Unable to open %s: %v\n", filename, err)
		fmt.Fprintf(stderr, format, v...)
	} else {
		fmt.Fprintf(fp, format, v...)
		fmt.Fprintln(fp)
		fp.Close()
	}
}

func vlog2file(filename, format string, v ...interface{}) {
	if fp, err := fopen(filename, "ab"); err != nil {
		fmt.Fprintf(stderr, "Unable to open %s: %v\n", filename, err)
		fmt.Fprintf(stderr, "%.16s: ", time.Now())
		fmt.Fprintf(stderr, format, v...)
	} else {
		fmt.Fprintf(fp, "%.32s: ", time.Now())
		fmt.Fprintf(fp, format, v...)
		fmt.Fprintln(fp)
		fp.Close()
	}
}

func log_sanity(format string, v ...interface{}) {
	vlog2file(LOG_SANITY, format, v...)
}

func log_status(format string, v ...interface{}) {
	vlog2file(LOG_STATUS, format, v...)
}

func log_muf(format string, v ...interface{}) {
	vlog2file(LOG_MUF, format, v...)
}

func log_gripe(format string, v ...interface{}) {
	vlog2file(LOG_GRIPE, format, v...)
}

func log_command(format string, v ...interface{}) {
	vlog2file(COMMAND_LOG, format, v...)
}

func strip_evil_characters(s string) (r string) {
 	for ; s != ""; s = s[1:] {
		switch c := s[0] & 127 {
 		case c == 0x1b:
 			r += '['
		case !unicode.IsPrint(c):
 			r += '_'
		default:
			r += c
		}
 	}
  	return
}

func log_user(player, program ObjectID, logmessage string) {
	log2file(USER_LOG, "%s", strip_evil_characters(fmt.Sprintf("%s(#%d) [%s(#%d)] at %.32s: %s", DB.Fetch(player).name, player, DB.Fetch(program).name, program, time.Now(), logmessage)))
}

func notify_fmt(player ObjectID, format string, args ...interface{}) {
	notify(player, fmt.Sprintf(format, args...))
}