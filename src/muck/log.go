package fbmuck

import(
	"log"
	"os"
)

func log2file(filename string, format string, v ...interface{}) {
	if f, e := os.OpenFile(filename, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0755); e == nil {
		fmt.Fprintf(f, format, v...)
		fmt.Fprintln(f)
		f.Close()
	} else {
		log.Printf("Unable to open %s: %v\n", filename, err)
		log.Printf(format, v...)
	}
}

func vlog2file(filename, format string, v ...interface{}) {
	if f, e := os.OpenFile(filename, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0755); e == nil {
		fmt.Fprintf(f, "%.32s: ", time.Now())
		fmt.Fprintf(f, format, v...)
		fmt.Fprintln(f)
		f.Close()
	} else {
		log.Printf("Unable to open %s: %v\n", filename, err)
		log.Printf("%.16s: ", time.Now())
		log.Printf(format, v...)
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