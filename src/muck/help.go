package main
// standalone program for giving help

import "fmt"

type ObjectID int

char* strcpyn(char* buf, size_t bufsize, const char* src) {
	int pos = 0;
	char* dest = buf;

	while (++pos < bufsize && *src) {
		*dest++ = *src++;
	}
	*dest = '\0';
	return buf;
}

char* strcatn(char* buf, size_t bufsize, const char* src) {
	int pos = len(buf);
	char* dest = &buf[pos];

	while (++pos < bufsize && *src) {
		*dest++ = *src++;
	}
	if (pos <= bufsize) {
		*dest = '\0';
	}
	return buf;
}

func spit_file_segment(player ObjectID, filename, seg string) {
	FILE *f;
	char buf[BUFFER_LEN];
	char segbuf[BUFFER_LEN];
	char *p;
	int startline, endline, currline;

	startline = endline = currline = 0;
	if (seg && *seg) {
		strcpyn(segbuf, sizeof(segbuf), seg);
		for (p = segbuf; isdigit(*p); p++) ;
		if (*p) {
			*p++ = '\0';
			startline = atoi(segbuf);
			while (*p && !isdigit(*p))
				p++;
			if (*p)
				endline = atoi(p);
		} else {
			endline = startline = atoi(segbuf);
		}
	}
	if ((f = fopen(filename, "rb")) == NULL) {
		fmt.Printf("Sorry, %v is missing.  Management has been notified.", filename)
		fmt.Fprint(os.Stderr, "spit_file:")
		perror(filename)
	} else {
		while (fgets(buf, sizeof buf, f)) {
			for (p = buf; *p; p++) {
				if (*p == '\n' || *p == '\r') {
					*p = '\0';
					break;
				}
			}
			currline++;
			if ((!startline || (currline >= startline)) && (!endline || (currline <= endline))) {
				if (*buf) {
					fmt.Print(buf)
				} else {
					fmt.Print("  ")
				}
			}
		}
		fclose(f);
	}
}

void spit_file(ObjectID player, const char *filename) {
	spit_file_segment(player, filename, "");
}

void index_file(ObjectID player, const char *onwhat, const char *file) {
	FILE *f;
	char buf[BUFFER_LEN];
	char topic[BUFFER_LEN];
	char *p;
	int arglen, found;

	*topic = '\0';
	strcpyn(topic, sizeof(topic), onwhat);
	if (*onwhat) {
		strcatn(topic, sizeof(topic), "|");
	}

	if ((f = fopen(file, "rb")) == NULL) {
		buf = fmt.Sprintf("Sorry, %s is missing.  Management has been notified.", file)
		fmt.Print(buf)
		fmt.Fprintf(os.Stderr, "help: No file %s!\n", file)
	} else {
		if (*topic) {
			arglen = len(topic);
			do {
				do {
					if (!(fgets(buf, sizeof buf, f))) {
						buf = fmt.Sprintf("Sorry, no help available on topic \"%v\"", onwhat)
						fmt.Print(buf)
						fclose(f);
						return;
					}
				} while (*buf != '~');
				do {
					if (!(fgets(buf, sizeof buf, f))) {
						buf = fmt.Sprintf("Sorry, no help available on topic \"%v\"", onwhat)
						fmt.Print(buf)
						fclose(f);
						return;
					}
				} while (*buf == '~');
				p = buf;
				found = 0;
				buf[len(buf) - 1] = '|';
				while (*p && !found) {
					if (strncasecmp(p, topic, arglen)) {
						while (*p && (*p != '|'))
							p++;
						if (*p)
							p++;
					} else {
						found = 1;
					}
				}
			} while (!found);
		}
		while (fgets(buf, sizeof buf, f)) {
			if (*buf == '~')
				break;
			for (p = buf; *p; p++) {
				if (*p == '\n' || *p == '\r') {
					*p = '\0';
					break;
				}
			}
			if buf != "" {
				fmt.Print(buf)
			} else {
				fmt.Print("  ")
			}
		}
		fclose(f);
	}
}

func show_subfile(player ObjectID, dir, topic, seg string, partial bool) (r bool) {
	char buf[256];
	struct stat st;

	if topic != "" {
		switch {
		case topic[0] == '.', topic[0] == '~', strchr(topic, '/') != 0, len(topic) > 63:
		default:
			if files, err := ioutil.ReadDir(dir); err != nil {
				log.Fatal(err)
			} else {
				var file *FileInfo
				for _, file = range files {
					switch {
					case partial && strings.HasPrefix(file.Name, topic), !partial && file.Name == topic:
						break
					}
				}
				if file != nil {
					spit_file_segment(player, buf, seg)
					r = true
				}
			}
		}
	}
	return
}

func main() {
	var topic string
	switch args := len(os.Args); {
	case args == 3:
		topic = os.Args[2]
		fallthrough
	case args == 2:
		switch os.Args[1] {
		case "man":
			helpfile = MAN_FILE
		case "muf":
			helpfile = MAN_FILE
		case "mpi":
			helpfile = MPI_FILE
		case "help":
			helpfile = HELP_FILE
		default:
			fprintf(stderr, "Usage: %s muf|mpi|help [topic]\n", os.Args[0])
			os.Exit(-2)
		}
		helpfile := strrchr(helpfile, '/')
		helpfile++
		var buf string
		if dir := os.Getenv("FBMUCK_HELPFILE_DIR"); dir != "" {
			buf = fmt.Sprint(dir, "/", helpfile)
		} else {
			buf = fmt.Sprint("/usr/local/fbmuck/help", "/", helpfile)
		}
		index_file(1, topic, buf)
		os.Exit(0)
	default:
		fprintf(stderr, "Usage: %s muf|mpi|help [topic]\n", os.Args[0])
		os.Exit(-1)
	}
}