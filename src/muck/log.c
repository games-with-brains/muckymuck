
package fbmuck

func vlog2file(prepend_time int, filename, format string, args ...interface{}) {
	fp *FILE
	time_t lt;
	char buf[40];
	lt = time(NULL);
	*buf = '\0';

	if ((fp = fopen(filename, "ab")) == NULL) {
		fprintf(stderr, "Unable to open %s!\n", filename);
		if (prepend_time)
			fprintf(stderr, "%.16s: ", ctime(&lt));
		vfprintf(stderr, format, args);
	} else {
		if (prepend_time) {
			buf = format_time("%c", localtime(&lt))
			fprintf(fp, "%.32s: ", buf);
		}
		
		vfprintf(fp, format, args);
		fprintf(fp, "\n");

		fclose(fp);
	}
}

void
log2file(char *filename, char *format, ...)
{
	va_list args;
	va_start(args, format);
	vlog2file(0, filename, format, args);
	va_end(args);
}

#define log_function(FILENAME) \
{ \
	va_list args; \
	va_start(args, format); \
	vlog2file(1, FILENAME, format, args); \
	va_end(args); \
}

void
log_sanity(char *format, ...) log_function(LOG_SANITY)

void
log_status(char *format, ...) log_function(LOG_STATUS)

void
log_muf(char *format, ...) log_function(LOG_MUF)

void
log_gripe(char *format, ...) log_function(LOG_GRIPE)

void
log_command(char *format, ...) log_function(COMMAND_LOG)

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

func log_user(player, program dbref, logmessage string) {
	log2file(USER_LOG, "%s", strip_evil_characters(fmt.Sprintf("%s(#%d) [%s(#%d)] at %.32s: %s", db.Fetch(player).name, player, db.Fetch(program).name, program, time.Now(), logmessage)))
}

func notify_fmt(player dbref, format string, args ...interface{}) {
	notify(player, fmt.Sprintf(format, args...))
}