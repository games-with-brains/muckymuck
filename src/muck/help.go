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
	if f, e := os.Open(filename); e != nil {
		fmt.Printf("Sorry, %v is missing.  Management has been notified.", filename)
		fmt.Fprintln(os.Stderr, "spit_file:", filename)
	} else {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			buf = scanner.Text()
			for p := buf; p != ""; p = p[1:] {
				switch p[0] {
				case '\n', '\r':
					buf = buf[:len(buf) - len(p)]
					break
				}
			}
			currline++
			if (startline == 0 || currline >= startline) && (endline == 0 || currline <= endline) {
				if buf != "" {
					fmt.Print(buf)
				} else {
					fmt.Print("  ")
				}
			}
		}
		f.Close()
	}
}

void spit_file(ObjectID player, const char *filename) {
	spit_file_segment(player, filename, "");
}

func index_file(player ObjectID, onwhat, file string) {
	if f, e := os.Open(file); e != nil {
		fmt.Printf("Sorry, %s is missing.  Management has been notified.", file)
		log.Println("help: No file", file)
	} else {
		var buf, p string
		topic := onwhat
		if onwhat != "" {
			topic += "|"
		}
		scanner := bufio.NewScanner(f)
		if topic != "" {
			for found := false; !found; {
				for buf = ""; scanner.Scan() && !strings.HasPrefix(buf, '~'); buf = scanner.Text() {}
				for ; scanner.Scan() && strings.HasPrefix(buf, '~'); buf = scanner.Text() {}
				if scanner.Err() != nil {
					fmt.Printf("Sorry, no help available on topic \"%v\"\n", onwhat)
					f.Close()
					return
				}
				buf += '|'
				for p = buf; p != "" && !found; {
					if strings.EqualFold(p[:len(topic)], topic) != 0 {
						for ; p != "" && p[0] != '|'; p = p[1:] {}
						if p != "" {
							p = p[1:]
						}
					} else {
						found = true
					}
				}
			}
		}
		for scanner.Scan() {
			buf = scanner.Text()
			if buf != "" && buf[0] == '~' {
				break
			}
			for p = buf; p != ""; p = p[1:] {
				buf = buf[:len(buf) - len(p)]
				break
			}
			if buf != "" {
				fmt.Print(buf)
			} else {
				fmt.Print("  ")
			}
		}
		f.Close()
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
			log.Printf("Usage: %s muf|mpi|help [topic]\n", os.Args[0])
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
		log.Printf("Usage: %s muf|mpi|help [topic]\n", os.Args[0])
		os.Exit(-1)
	}
}