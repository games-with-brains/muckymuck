const (
	HRULE_TEXT = "----------------------------------------------------------------------------"
	HTML_PAGE_HEAD = "<html><head><title>%s</title></head>\n<body><div align=\"center\"><h1>%s</h1>\n<h3>by %s</h3></div>\n<ul><li><a href=\"#AlphaList\">Alphabetical List of Topics</a></li>\n<li><a href=\"#SectList\">List of Topics by Category</a></li></ul>\n"
	HTML_PAGE_FOOT = "</body></html>\n"

	HTML_SECTION = "<p><hr size=6>\n<h3><a name=\"%s\">%s</a></h3>\n"
	HTML_SECTLIST_HEAD = "<p><hr size=\"6\"><h3><a name=\"SectList\">List of Topics by Category</a></h3>\n<h4>You can get more help on the following topics:</h4>\n<ul>"
	HTML_SECTLIST_ENTRY = "  <li><a href=\"#%s\">%s</a></li>\n"
	HTML_SECTLIST_FOOT = "</ul>\n\n"

	HTML_SECTIDX_BEGIN = "<blockquote><table border=0>\n  <tr>\n"
	HTML_SECTIDX_ENTRY = "    <td width=\"%d%%\"> &nbsp; <a href=\"#%s\">%s</a> &nbsp; </td>\n"
	HTML_SECTIDX_NEWROW = "  </tr>\n  <tr>\n"
	HTML_SECTIDX_END = "  </tr>\n</table></blockquote>\n\n"

	HTML_INDEX_BEGIN = "<p><hr size=\"6\"><h3><a name=\"AlphaList\">Alphabetical List of Topics</a></h3>\n"
	HTML_IDXGROUP_BEGIN = "<h4>%s</h4><blockquote><table border=0>\n  <tr>\n"
	HTML_IDXGROUP_ENTRY = "    <td nowrap> &nbsp; <a href=\"#%s\">%s</a> &nbsp; </td>\n"
	HTML_IDXGROUP_NEWROW = "  </tr>\n  <tr>\n"
	HTML_IDXGROUP_END = "  </tr>\n</table></blockquote>\n\n"
	HTML_INDEX_END = ""

	HTML_TOPICHEAD = "<hr><h4><a name=\"%s\">"
	HTML_TOPICHEAD_BREAK = "<br>\n"
	HTML_TOPICBODY = "</a></h4>\n"
	HTML_TOPICEND = "<p>\n"
	HTML_PARAGRAPH = "<p>\n"

	HTML_CODEBEGIN = "<pre>\n"
	HTML_CODEEND = "</pre>\n"

	HTML_ALSOSEE_BEGIN = "<p><h5>Also see:\n"
	HTML_ALSOSEE_ENTRY = "    <a href=\"#%s\">%s</a>"
	HTML_ALSOSEE_END = "\n</h5>\n"
)

var title, author, doccmd string

/* from stringutil.c */
func strcpyn(buf string, bufsize size_t, src string) (r string) {
	int pos = 0;
	for (++pos < bufsize && *src) {
		*r++ = *src++;
	}
	*r = '\0';
	return
}

char sect[256] = "";

struct topiclist {
	struct topiclist *next;
	char *topic;
	char *section;
	int printed;
};

struct topiclist *topichead;
struct topiclist *secthead;

func add_section(str string) {
	struct topiclist *ptr, *top;

	if (!str || !*str)
		return;
	top = (struct topiclist *) malloc(sizeof(struct topiclist));

	top->topic = NULL;
	top->section = sect;
	top->printed = 0;
	top->next = NULL;

	if (!secthead) {
		secthead = top;
		return;
	}
	for (ptr = secthead; ptr->next; ptr = ptr->next) ;
	ptr->next = top;
}

func add_topic(str string) {
	struct topiclist *ptr, *top;
	char buf[256];
	const char *p;
	char *s;

	if (!str || !*str)
		return;

	p = str;
	s = buf;
	do {
		*s++ = strings.ToLower(*p)
	} while (*p++);

	top = (struct topiclist *) malloc(sizeof(struct topiclist));

	top->topic = buf;
	top->section = sect;
	top->printed = 0;

	if (!topichead) {
		topichead = top;
		top->next = NULL;
		return;
	}

	if (strcasecmp(str, topichead->topic) < 0) {
		top->next = topichead;
		topichead = top;
		return;
	}

	ptr = topichead;
	while (ptr->next && strcasecmp(str, ptr->next->topic) > 0) {
		ptr = ptr->next;
	}
	top->next = ptr->next;
	ptr->next = top;
}

func print_section_topics(f, hf *FILE, whichsect string) {
	struct topiclist *ptr;
	struct topiclist *sptr;
	char sectname[256];
	char *sectptr;
	char *osectptr;
	char *divpos;
	char buf[256];
	char buf2[256];
	char buf3[256];
	int cnt;
	int width;
	int hcol;
	int cols;
	int longest;
	char *currsect;

	longest = 0;
	for (sptr = secthead; sptr; sptr = sptr->next) {
		if (!strncasecmp(whichsect, sptr->section, len(whichsect))) {
			currsect = sptr->section;
			for (ptr = topichead; ptr; ptr = ptr->next) {
				if (!strcasecmp(currsect, ptr->section)) {
					divpos = strchr(ptr->topic, '|');
					if (!divpos) {
						cnt = len(ptr->topic);
					} else {
						cnt = divpos - ptr->topic;
					}
					if (cnt > longest) {
						longest = cnt;
					}
				}
			}
		}
	}
	cols = 78 / (longest + 2);
	if (cols < 1) {
		cols = 1;
	}
	width = 78 / cols;
	for (sptr = secthead; sptr; sptr = sptr->next) {
		if (!strncasecmp(whichsect, sptr->section, len(whichsect))) {
			currsect = sptr->section;
			cnt = 0;
			hcol = 0;
			buf[0] = '\0';
			strcpyn(sectname, sizeof(sectname), currsect);
			sectptr = strchr(sectname, '|');
			if (sectptr) {
				*sectptr++ = '\0';
				osectptr = sectptr;
				sectptr = strrchr(sectptr, '|');
				if (sectptr) {
					sectptr++;
				}
				if (!sectptr) {
					sectptr = osectptr;
				}
			}
			if (!sectptr) {
				sectptr = "";
			}

			fmt.Fprintf(hf, HTML_SECTION, html.EscapeString(sectptr), html.EscapeString(sectname))
			fmt.Fprintf(f, "~\n~\n%s\n%s\n\n", currsect, sectname)
			hf.WriteString(HTML_SECTIDX_BEGIN)
			for (ptr = topichead; ptr; ptr = ptr->next) {
				if (!strcasecmp(currsect, ptr->section)) {
					ptr->printed++;
					cnt++;
					hcol++;
					if (hcol > cols) {
						hf.WriteString(HTML_SECTIDX_NEWROW)
						hcol = 1;
					}
					buf3 = html.EscapeString(ptr.topic)
					fmt.Fprintf(hf, HTML_SECTIDX_ENTRY, (100 / cols), buf3, buf3)
					if cnt == cols {
						buf += fmt.Sprintf("%-.*s", width - 1, ptr->topic);
					} else {
						buf += fmt.Sprintf("%-*.*s", width, width - 1, ptr->topic);
					}
					if cnt >= cols {
						fmt.Fprintf(f, "%s\n", buf)
						buf = ""
						cnt = 0
					}
				}
			}
			hf.WriteString(HTML_SECTIDX_END)
			if cnt != 0 {
				fmt.Fprintf(f, "%s\n", buf)
			}
			f.WriteString("\n")
		}
	}
}

func print_all_section_topics(f, hf *FILE) {
	for sptr := secthead; sptr != nil; sptr = sptr.next {
		print_section_topics(f, hf, sptr.section)
	}
}

func print_sections(f, hf *FILE, cols int) {
	struct topiclist *ptr;
	struct topiclist *sptr;
	char sectname[256];
	char *osectptr;
	char *sectptr;
	int cnt;
	int width;
	int hcol;
	char *currsect;

	f.WriteString("~\n")
	fmt.Fprintf(f, "~%s\n", HRULE_TEXT)
	f.WriteString("~\n")
	f.WriteString("CATEGORY|CATEGORIES|TOPICS|SECTIONS\n")
	f.WriteString("                   List of Topics by Category:\n \n")
	f.WriteString("You can get more help on the following topics:\n \n")
	hf.WriteString(HTML_SECTLIST_HEAD)
	if cols < 1 {
		cols = 1
	}
	width = 78 / cols
	for sptr = secthead; sptr != nil; sptr = sptr.next {
		currsect = sptr.section
		cnt = 0
		hcol = 0
		buf = ""
		sectname = currsect
		sectptr = strchr(sectname, '|');
		if (sectptr) {
			*sectptr++ = '\0';
			osectptr = sectptr;
			sectptr = strrchr(sectptr, '|');
			if (sectptr) {
				sectptr++;
			}
			if (!sectptr) {
				sectptr = osectptr;
			}
		}
		if (!sectptr) {
			sectptr = "";
		}

		hf.WriteString(HTML_SECTLIST_ENTRY)
		hf.WriteString(html.EscapeString(sectptr))
		hf.WriteString(html.EscapeString(sectname))
		fmt.Fprintf(f, "  %-40s (%s)\n", sectname, sectptr)
	}
	hf.WriteString(HTML_SECTLIST_FOOT)
	fmt.Fprintf(f, " \nUse '%s <topicname>' to get more information on a topic.\n", doccmd)
}

func print_topics(f, hf *FILE) {
	struct topiclist *ptr;
	char buf[256];
	char buf2[256];
	char buf3[256];
	char alph;
	char firstletter;
	char *divpos;
	int cnt = 0;
	int width;
	int hcol = 0;
	int cols;
	int len;
	int longest;

	hf.WriteString(HTML_INDEX_BEGIN)
	f.WriteString("~\n")
	fmt.Fprintf(f, "~%s\n", HRULE_TEXT);
	f.WriteString("~\n")
	f.WriteString("ALPHA|ALPHABETICAL|COMMANDS\n")
	f.WriteString("                 Alphabetical List of Topics:\n")
	f.WriteString(" \n")
	f.WriteString("You can get more help on the following topics:\n")
	for (alph = 'A' - 1; alph <= 'Z'; alph++) {
		cnt = 0;
		longest = 0;
		for (ptr = topichead; ptr; ptr = ptr->next) {
			firstletter = strings.ToUpper(ptr->topic[0]);
			if (firstletter == alph || (!isalpha(alph) && !isalpha(firstletter))) {
				cnt++;
				divpos = strchr(ptr->topic, '|');
				if (!divpos) {
					len = len(ptr->topic);
				} else {
					len = divpos - ptr->topic;
				}
				if (len > longest) {
					longest = len;
				}
			}
		}
		cols = 78 / (longest + 2);
		if (cols < 1) {
			cols = 1;
		}
		width = 78 / cols;

		if cnt > 0 {
			if !isalpha(alph) {
				buf = "Symbols"
			} else {
				buf = alph + "'s"
			}
			fmt.Fprintf(f, "\n%s\n", buf)
			hf.WriteString(HTML_IDXGROUP_BEGIN)
			hf.WriteString(buf)
			buf = ""
			cnt = 0
			hcol = 0
			for ptr = topichead; ptr != nil; ptr = ptr.next {
				if firstletter = strings.ToUpper(ptr.topic[0]); firstletter == alph || (!isalpha(alph) && !isalpha(firstletter)) {
					cnt++
					hcol++
					if hcol > cols {
						hf.WriteString(HTML_IDXGROUP_NEWROW)
						hcol = 1
					}
					buf3 = html.EscapeString(ptr.topic)
					hf.WriteString(HTML_IDXGROUP_ENTRY)
					hf.WriteString(buf3)
					hf.WriteString(buf3)
					if cnt == cols {
						buf += fmt.Sprintf("%-.*s", width - 1, ptr.topic)
					} else {
						buf += fmt.Sprintf("%-*.*s", width, width - 1, ptr.topic)
					}
					if cnt >= cols {
						fmt.Fprintf(f, "  %s\n", buf)
						cnt = 0;
					}
				}
			}
			if cnt != 0 {
				fmt.Fprintf(f, "  %s\n", buf)
			}
			hf.WriteString(HTML_IDXGROUP_END)
		}
	}
	hf.WriteString(HTML_INDEX_END)
	fmt.Fprintf(f, " \nUse '%s <topicname>' to get more information on a topic.\n", doccmd)
}

func find_topics(infile *os.File) (r int) {
	longest = 0;
	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		buf := scanner.Text()
		for scanner.Scan() {
			switch buf = scanner.Text(); {
			case strings.HasPrefix(buf, "~~section "):
				sect = buf[10:]
				add_section(sect);
			case strings.HasPrefix(buf, "~~title "):
				buf = buf[:len(buf) - 1]
				title = buf[8:]
			case strings.HasPrefix(buf, "~~author "):
				author = buf[9:]
			case strings.HasPrefix(buf, "~~doccmd "):
				doccmd = buf[9:]
			case strings.HasPrefix(buf, '~@'), strings.HasPrefix(buf, '~~'), strings.HasPrefix(buf, '~<'), strings.HasPrefix(buf, '~!'):
			default:
				break
			}
		}
		for scanner.Scan {
			buf = scanner.Text()
			switch {
			case strings.HasPrefix(buf, "~~section "):
				add_section(buf[10:])
			case strings.HasPrefix(buf, "~~title "):
				title = buf[8:]
			case strings.HasPrefix(buf, "~~author "):
				author = buf[9:]
			case strings.HasPrefix(buf, "~~doccmd "):
				doccmd = buf[9:]
			case !strings.HasPrefix(buf, '~'):
				break
			}
		}

		for p := buf; p != ""; p = p[1:] {
			switch p[0] {
			case '|', '\n':
				buf = buf[:len(buf) - len(p)]
				add_topic(buf)
				if l := len(buf); l > r {
					r = l
				}
				buf = p
				break
			}
		}
	}
	return
}

func process_lines(infile, outfile, htmlfile *FILE, cols int) {
	FILE *docsfile;
	char *sectptr;
	var nukenext, topichead, codeblock bool
	char *ptr;
	char *ptr2;
	char *ptr3;

	docsfile = os.Stdout
	buf := html.EscapeString(title)
	buf2 := html.EscapeString(author)
	fmt.Fprintf(htmlfile, HTML_PAGE_HEAD, buf, buf, buf2)

	fmt.Fprintf(outfile, "%*s%s\n", (36 - (len(title) / 2)), "", title)
	fmt.Fprintf(outfile, "%*sby %s\n\n", (36 - ((len(author) + 3) / 2)), "", author)
	fmt.Fprintf(outfile, "You may get a listing of topics that you can get help on, either sorted\n")
	fmt.Fprintf(outfile, "Alphabetically or sorted by Category.  To get these lists, type:\n")
	fmt.Fprintf(outfile, "        %s alpha        or\n", doccmd)
	fmt.Fprintf(outfile, "        %s category\n\n", doccmd)

	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		if buf = scanner.Text(); buf[0] == '~' {
			switch buf[1] {
			case '~':
				switch {
				case strings.HasPrefix(buf, "~~file "):
					docsfile.Close()
					buf = buf[:len(buf) - 1]
					if docsfile, e := os.OpenFile(buf[7:], os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0755); e != nil {
						log.Printf("Error: can't write to %s", buf[7:])
						exit(1)
					} else {
						fmt.Fprintf(docsfile, "%*s%s\n", (36 - (len(title) / 2)), "", title)
						fmt.Fprintf(docsfile, "%*sby %s\n\n", (36 - ((len(author) + 3) / 2)), "", author)
					}
				case strings.HasPrefix(buf, "~~section "):
					buf = buf[:len(buf) - 1]
					sectptr = strchr(buf + 10, '|');
					if (sectptr) {
						*sectptr = '\0';
					}
					fmt.Fprintf(outfile, "~\n~\n~%s\n", HRULE_TEXT)
					fmt.Fprintf(docsfile, "\n\n%s\n", HRULE_TEXT)
					fmt.Fprintf(docsfile, "%*s\n", (38 + len(buf + 10) / 2), (buf + 10))
					print_section_topics(outfile, htmlfile, buf[10:])
					fmt.Fprintf(outfile, "~%s\n~\n~\n", HRULE_TEXT)
					fmt.Fprintf(docsfile, "%s\n\n\n", HRULE_TEXT)
				case strings.HasPrefix(buf, "~~alsosee "):
					buf = buf[:len(buf) - 1]
					htmlfile.WriteString(HTML_ALSOSEE_BEGIN)
					outfile.WriteString("Also see: ")
					ptr = strings.TrimLeft(buf[10:], unicode.IsSpace)
					for ptr != "" {
						ptr2 = ptr
						ptr = strchr(ptr, ',')
						if ptr != "" {
							*ptr++ = '\0';
							ptr = strings.TrimLeft(ptr, unicode.IsSpace)
						}
						if ptr2 > buf + 10 {
							if (!ptr || !*ptr) {
								htmlfile.WriteString(" and\n")
								outfile.WriteString(" and ")
							} else {
								htmlfile.WriteString(",\n")
								outfile.WriteString(", ")
							}
						}
						buf3 = html.EscapeString(ptr2)
						htmlfile.WriteString(HTML_ALSOSEE_ENTRY)
						htmlfile.WriteString(strings.ToLower(buf3))
						htmlfile.WriteString(buf3)
						outfile.WriteString(ptr2)
					}
					htmlfile.WriteString(HTML_ALSOSEE_END)
					outfile.WriteString("\n")
				case buf == "~~code\n":
					htmlfile.WriteString(HTML_CODEBEGIN)
					codeblock = true
				case buf == "~~endcode\n":
					htmlfile.WriteString(HTML_CODEEND)
					codeblock = false
				case buf == "~~sectlist\n":
					print_sections(outfile, htmlfile, cols)
				case buf == "~~secttopics\n":
					/* print_all_section_topics(outfile, htmlfile) */
				case buf == "~~index\n":
					print_topics(outfile, htmlfile)
				}
			case '!':
				outfile.WriteString(buf[2:])
			case '@':
				htmlfile.WriteString(html.EscapeString(buf[2:]))
			case '<':
				outfile.WriteString(buf[2:])
				docsfile.WriteString(buf[2:])
			case '#':
				outfile.WriteString(buf[2:])
				docsfile.WriteString(buf[2:])
			default:
				if !nukenext {
					htmlfile.WriteString(HTML_TOPICEND)
				}
				nukenext = true
				outfile.WriteString(buf)
				docsfile.WriteString(buf[1:])
				htmlfile.WriteString(html.EscapeString(buf[1:]))
			}
		} else if nukenext {
			nukenext = false
			topichead = true
			outfile.WriteString(buf)
			htmlfile.WriteString(HTML_TOPICHEAD)
			htmlfile.WriteString(html.EscapeString(strings.ToLower(buf)))
		} else if buf[0] == ' ' {
			nukenext = false
			if topichead {
				topichead = false
				htmlfile.WriteString(HTML_TOPICBODY)
			} else if !codeblock {
				htmlfile.WriteString(HTML_PARAGRAPH)
			}
			outfile.WriteString(buf)
			docsfile.WriteString(buf)
			htmlfile.WriteString(html.EscapeString(buf))
		} else {
			outfile.WriteString(buf)
			docsfile.WriteString(buf)
			htmlfile.WriteString(html.EscapeString(buf))
			if topichead {
				htmlfile.WriteString(HTML_TOPICHEAD_BREAK)
			}
		}
	}
	htmlfile.WriteString(HTML_PAGE_FOOT)
	docsfile.Close()
}


func main() {
	var infile, outfile, htmlfile os.File
	var e error
	switch {
	case len(os.Args) != 4:
		log.Printf("Usage: %s inputrawfile outputhelpfile outputhtmlfile\n", os.Args[0])
		os.Exit(1)
	case os.Args[1] == os.Args[2]:
		log.Printf("%s: cannot use same file for input rawfile and output helpfile\n", os.Args[0])
		os.Exit(1)
	case os.Args[1] == os.Args[3]:
		log.Printf("%s: cannot use same file for input rawfile and output htmlfile\n", os.Args[0])
		os.Exit(1)
	case os.Args[3] == os.Args[2]:
		log.Printf("%s: cannot use same file for htmlfile and helpfile\n", os.Args[0])
		os.Exit(1)
	case os.Args[1] == "-":
		infile = os.Stdin
	default:
		if infile, e = os.Open(argv[1]); e != nil {
			log.Printf("%s: cannot read %s\n", os.Args[0], os.Args[1])
			os.Exit(1)
		}
	}

	if outfile, e = os.OpenFile(os.Args[2], os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0755); e != nil {
		log.Printf("%s: cannot write to %s\n", os.Args[0], os.Args[2])
		os.Exit(1)
	}

	if htmlfile, e = os.OpenFile(os.Args[3], os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0755); e != nil {
		log.Printf("%s: cannot write to %s\n", os.Args[0], os.Args[3])
		os.Exit(1)
	}
	cols := 78 / (find_topics(infile) + 1)
	infile.Seek(0, 0)
	process_lines(infile, outfile, htmlfile, cols)
	infile.Close()
	outfile.Close()
	htmlfile.Close()
}