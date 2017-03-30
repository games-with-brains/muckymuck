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

func escape_html(buf string, buflen int, in string) string {
	char* out = buf;
	while (*in) {
		if (*in == '<') {
			strcpyn(out, buflen, "&lt;");
			out += len(out);
		} else if (*in == '>') {
			strcpyn(out, buflen, "&gt;");
			out += len(out);
		} else if (*in == '&') {
			strcpyn(out, buflen, "&amp;");
			out += len(out);
		} else if (*in == '"') {
			strcpyn(out, buflen, "&quot;");
			out += len(out);
		} else {
			*out++ = *in;
		}
		in++;
	}
	*out++ = '\0';
	return buf;
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

			fprintf(hf, HTML_SECTION, escape_html(buf2, sizeof(buf2), sectptr), escape_html(buf3, sizeof(buf3), sectname));
			fprintf(f, "~\n~\n%s\n%s\n\n", currsect, sectname);
			fprintf(hf, HTML_SECTIDX_BEGIN);
			for (ptr = topichead; ptr; ptr = ptr->next) {
				if (!strcasecmp(currsect, ptr->section)) {
					ptr->printed++;
					cnt++;
					hcol++;
					if (hcol > cols) {
						fprintf(hf, HTML_SECTIDX_NEWROW);
						hcol = 1;
					}
					escape_html(buf3, sizeof(buf3), ptr->topic);
					fprintf(hf, HTML_SECTIDX_ENTRY, (100 / cols), buf3, buf3);
					if (cnt == cols) {
						buf2 = fmt.Sprintf("%-.*s", width - 1, ptr->topic);
					} else {
						buf2 = fmt.Sprintf("%-*.*s", width, width - 1, ptr->topic);
					}
					strcat(buf, buf2);
					if (cnt >= cols) {
						fprintf(f, "%s\n", buf);
						buf[0] = '\0';
						cnt = 0;
					}
				}
			}
			fprintf(hf, HTML_SECTIDX_END);
			if (cnt)
				fprintf(f, "%s\n", buf);
			fprintf(f, "\n");
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
	char buf[256];
	char buf2[256];
	char buf3[256];
	char buf4[256];
	int cnt;
	int width;
	int hcol;
	char *currsect;

	fprintf(f, "~\n");
	fprintf(f, "~%s\n", HRULE_TEXT);
	fprintf(f, "~\n");
	fprintf(f, "CATEGORY|CATEGORIES|TOPICS|SECTIONS\n");
	fprintf(f, "                   List of Topics by Category:\n \n");
	fprintf(f, "You can get more help on the following topics:\n \n");
	fprintf(hf, HTML_SECTLIST_HEAD);
	if (cols < 1) {
		cols = 1;
	}
	width = 78 / cols;
	for (sptr = secthead; sptr; sptr = sptr->next) {
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

		fprintf(hf, HTML_SECTLIST_ENTRY, escape_html(buf3, sizeof(buf3), sectptr), escape_html(buf4, sizeof(buf4), sectname));
		fprintf(f, "  %-40s (%s)\n", sectname, sectptr);
	}
	fprintf(hf, HTML_SECTLIST_FOOT);
	fprintf(f, " \nUse '%s <topicname>' to get more information on a topic.\n", doccmd);
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

	fprintf(hf, HTML_INDEX_BEGIN);
	fprintf(f, "~\n");
	fprintf(f, "~%s\n", HRULE_TEXT);
        fprintf(f, "~\n");
	fprintf(f, "ALPHA|ALPHABETICAL|COMMANDS\n");
        fprintf(f, "                 Alphabetical List of Topics:\n");
        fprintf(f, " \n");
	fprintf(f, "You can get more help on the following topics:\n");
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

		if (cnt > 0) {
			if (!isalpha(alph)) {
				strcpyn(buf, sizeof(buf), "Symbols");
			} else {
				buf[0] = alph;
				buf[1] = '\'';
				buf[2] = 's';
				buf[3] = '\0';
			}
			fprintf(f, "\n%s\n", buf);
			fprintf(hf, HTML_IDXGROUP_BEGIN, buf);
			buf[0] = '\0';
			cnt = 0;
			hcol = 0;
			for (ptr = topichead; ptr; ptr = ptr->next) {
				firstletter = strings.ToUpper(ptr->topic[0]);
				if (firstletter == alph || (!isalpha(alph) && !isalpha(firstletter))) {
					cnt++;
					hcol++;
					if (hcol > cols) {
						fprintf(hf, HTML_IDXGROUP_NEWROW);
						hcol = 1;
					}
					escape_html(buf3, sizeof(buf3), ptr->topic);
					fprintf(hf, HTML_IDXGROUP_ENTRY, /*(100 / cols),*/ buf3, buf3);
					if (cnt == cols) {
						buf2 = fmt.Sprintf("%-.*s", width - 1, ptr->topic);
					} else {
						buf2 = fmt.Sprintf("%-*.*s", width, width - 1, ptr->topic);
					}
					strcat(buf, buf2);
					if (cnt >= cols) {
						fprintf(f, "  %s\n", buf);
						buf[0] = '\0';
						cnt = 0;
					}
				}
			}
			if (cnt) {
				fprintf(f, "  %s\n", buf);
			}
			fprintf(hf, HTML_IDXGROUP_END);
		}
	}
	fprintf(hf, "%s", HTML_INDEX_END);
	fprintf(f, " \nUse '%s <topicname>' to get more information on a topic.\n", doccmd);
}

func find_topics(infile *FILE) int {
	char buf[4096];
	char *s, *p;
	int longest, lng;

	longest = 0;
	while (!feof(infile)) {
		do {
			if (!fgets(buf, sizeof(buf), infile)) {
				*buf = '\0';
				break;
			} else {
				case strings.HasPrefix(buf, "~~section "):
					buf[len(buf) - 1] = '\0';
					strcpyn(sect, sizeof(sect), (buf + 10));
					add_section(sect);
				case strings.HasPrefix(buf, "~~title "):
					buf[len(buf) - 1] = '\0';
					title = buf+8;
				case strings.HasPrefix(buf, "~~author "):
					buf[len(buf) - 1] = '\0';
					author = buf+9;
				case strings.HasPrefix(buf, "~~doccmd "):
					buf[len(buf) - 1] = '\0';
					doccmd = buf+9;
				}
			}
		} while (!feof(infile) &&
				 (*buf != '~' || buf[1] == '@' || buf[1] == '~' || buf[1] == '<' ||
				  buf[1] == '!'));

		do {
			if (!fgets(buf, sizeof(buf), infile)) {
				*buf = '\0';
				break;
			} else {
				switch {
				case strings.HasPrefix(buf, "~~section "):
					buf[len(buf) - 1] = '\0';
					strcpyn(sect, sizeof(sect), (buf + 10));
					add_section(sect);
				case strings.HasPrefix(buf, "~~title "):
					buf[len(buf) - 1] = '\0';
					title = buf+8;
				case strings.HasPrefix(buf, "~~author "):
					buf[len(buf) - 1] = '\0';
					author = buf+9;
				case strings.HasPrefix(buf, "~~doccmd "):
					buf[len(buf) - 1] = '\0';
					doccmd = buf+9;
				}
			}
		} while (*buf == '~' && !feof(infile));

		for (s = p = buf; *s; s++) {
			if (*s == '|' || *s == '\n') {
				*s++ = '\0';
				add_topic(p);
				lng = len(p);
				if (lng > longest)
					longest = lng;
				p = s;
				break;
			}
		}
	}
	return (longest);
}

func process_lines(infile, outfile, htmlfile *FILE, cols int) {
	FILE *docsfile;
	char *sectptr;
	char buf[4096];
	char buf2[4096];
	char buf3[4096];
	char buf4[4096];
	int nukenext = 0;
	int topichead = 0;
	int codeblock = 0;
	char *ptr;
	char *ptr2;
	char *ptr3;

	docsfile = stdout;
	escape_html(buf, sizeof(buf), title);
	escape_html(buf2, sizeof(buf2), author);
	fprintf(htmlfile, HTML_PAGE_HEAD, buf, buf, buf2);

	fprintf(outfile, "%*s%s\n", (int)(36-(len(title)/2)), "", title);
	fprintf(outfile, "%*sby %s\n\n", (int)(36-((len(author)+3)/2)), "", author);
	fprintf(outfile, "You may get a listing of topics that you can get help on, either sorted\n");
	fprintf(outfile, "Alphabetically or sorted by Category.  To get these lists, type:\n");
	fprintf(outfile, "        %s alpha        or\n", doccmd);
	fprintf(outfile, "        %s category\n\n", doccmd);

	while (!feof(infile)) {
		if (!fgets(buf, sizeof(buf), infile)) {
			break;
		}
		if (buf[0] == '~') {
			switch buf[1] {
			case '~':
				switch {
				case strings.HasPrefix(buf, "~~file "):
					fclose(docsfile);
					buf[len(buf) - 1] = '\0';
					if docsfile = fopen(buf + 7, "wb"); docsfile == nil {
						fprintf(stderr, "Error: can't write to %s", buf + 7)
						exit(1)
					}
					fprintf(docsfile,  "%*s%s\n", (int)(36-(len(title)/2)), "", title);
					fprintf(docsfile,  "%*sby %s\n\n", (int)(36-((len(author)+3)/2)), "", author);
				case strings.HasPrefix(buf, "~~section "):
					buf[len(buf) - 1] = '\0';
					sectptr = strchr(buf + 10, '|');
					if (sectptr) {
						*sectptr = '\0';
					}
					fprintf(outfile, "~\n~\n~%s\n", HRULE_TEXT);
					fprintf(docsfile, "\n\n%s\n", HRULE_TEXT);
					fprintf(docsfile, "%*s\n", (int)(38 + len(buf + 10) / 2), (buf + 10));
					print_section_topics(outfile, htmlfile, (buf + 10));
					fprintf(outfile, "~%s\n~\n~\n", HRULE_TEXT);
					fprintf(docsfile, "%s\n\n\n", HRULE_TEXT);
				case strings.HasPrefix(buf, "~~alsosee "):
					buf[len(buf) - 1] = '\0';
					fprintf(htmlfile, HTML_ALSOSEE_BEGIN);
					fprintf(outfile, "Also see: ");
					ptr = strings.TrimLeft(buf[10:], unicode.IsSpace)
					while (ptr && *ptr) {
						ptr2 = ptr
						ptr = strchr(ptr, ',')
						if ptr != "" {
							*ptr++ = '\0';
							ptr = strings.TrimLeft(ptr, unicode.IsSpace)
						}
						if (ptr2 > buf + 10) {
							if (!ptr || !*ptr) {
								fprintf(htmlfile, " and\n");
								fprintf(outfile, " and ");
							} else {
								fprintf(htmlfile, ",\n");
								fprintf(outfile, ", ");
							}
						}
						escape_html(buf3, sizeof(buf3), ptr2);
						strcpyn(buf4, sizeof(buf4), buf3);
						for (ptr3 = buf4; *ptr3; ptr3++) {
							*ptr3 = strings.ToLower(*ptr3)
						}
						fprintf(htmlfile, HTML_ALSOSEE_ENTRY, buf4, buf3);
						fprintf(outfile, "%s", ptr2);
					}
					fprintf(htmlfile, HTML_ALSOSEE_END);
					fprintf(outfile, "\n");
				case buf == "~~code\n":
					fprintf(htmlfile, HTML_CODEBEGIN);
					codeblock = 1;
				case buf == "~~endcode\n":
					fprintf(htmlfile, HTML_CODEEND);
					codeblock = 0;
				case buf == "~~sectlist\n":
					print_sections(outfile, htmlfile, cols);
				case buf == "~~secttopics\n":
					/* print_all_section_topics(outfile, htmlfile); */
				case buf == "~~index\n":
					print_topics(outfile, htmlfile);
				}
			case '!':
				fprintf(outfile, "%s", buf + 2);
			case '@':
				escape_html(buf3, sizeof(buf3), buf + 2);
				fprintf(htmlfile, "%s", buf3);
			case '<':
				fprintf(outfile, "%s", buf + 2);
				fprintf(docsfile, "%s", buf + 2);
			case '#':
				fprintf(outfile, "~%s", buf + 2);
				fprintf(docsfile, "%s", buf + 2);
			default:
				if (!nukenext) {
					fprintf(htmlfile, HTML_TOPICEND)
				}
				nukenext = 1
				fprintf(outfile, "%s", buf)
				fprintf(docsfile, "%s", buf + 1)
				escape_html(buf3, sizeof(buf3), buf + 1)
				fprintf(htmlfile, "%s", buf3)
			}
		} else if (nukenext) {
			nukenext = 0;
			topichead = 1;
			fprintf(outfile, "%s", buf);
			for (ptr = buf; *ptr && *ptr != '|' && *ptr != '\n'; ptr++) {
				*ptr = strings.ToLower(*ptr)
			}
			*ptr = '\0';
			escape_html(buf3, sizeof(buf3), buf);
			fprintf(htmlfile, HTML_TOPICHEAD, buf3);
		} else if (buf[0] == ' ') {
			nukenext = 0;
			if (topichead) {
				topichead = 0;
				fprintf(htmlfile, HTML_TOPICBODY);
			} else if (!codeblock) {
				fprintf(htmlfile, HTML_PARAGRAPH);
			}
			fprintf(outfile, "%s", buf);
			fprintf(docsfile, "%s", buf);
			escape_html(buf3, sizeof(buf3), buf);
			fprintf(htmlfile, "%s", buf3);
		} else {
			fprintf(outfile, "%s", buf);
			fprintf(docsfile, "%s", buf);
			escape_html(buf3, sizeof(buf3), buf);
			fprintf(htmlfile, "%s", buf3);
			if (topichead) {
				fprintf(htmlfile, HTML_TOPICHEAD_BREAK);
			}
		}
	}
	fprintf(htmlfile, HTML_PAGE_FOOT);
	fclose(docsfile);
}


func main() {
	var infile, outfile, htmlfile *FILE

	switch {
	case len(os.Args) != 4:
		fmt.Fprintf(os.Stderr, "Usage: %s inputrawfile outputhelpfile outputhtmlfile\n", os.Args[0])
		os.Exit(1)
	case os.Args[1] == os.Args[2]:
		fmt.Fprintf(os.Stderr, "%s: cannot use same file for input rawfile and output helpfile\n", os.Args[0])
		os.Exit(1)
	case os.Args[1] == os.Args[3]:
		fmt.Fprintf(os.Stderr, "%s: cannot use same file for input rawfile and output htmlfile\n", os.Args[0])
		os.Exit(1)
	case argv[3] == argv[2]:
		fmt.Fprintf(os.Stderr, "%s: cannot use same file for htmlfile and helpfile\n", os.Args[0])
		os.Exit(1)
	case os.Args[1] == "-":
		infile = os.Stdin
	default:
		if infile = fopen(argv[1], "rb"); infile == nil {
			fmt.Fprintf(os.Stderr, "%s: cannot read %s\n", os.Args[0], os.Args[1])
			os.Exit(1)
		}
	}

	if outfile = fopen(argv[2], "wb"); outfile == nil {
		fmt.Fprintf(os.Stderr, "%s: cannot write to %s\n", os.Args[0], os.Args[2])
		os.Exit(1)
	}

	if htmlfile = fopen(argv[3], "wb"); htmlfile == nil {
		fmt.Fprintf(os.Stderr, "%s: cannot write to %s\n", os.Args[0], os.Args[3])
		os.Exit(1)
	}
	cols := 78 / (find_topics(infile) + 1)
	fseek(infile, 0L, 0)
	process_lines(infile, outfile, htmlfile, cols)
	fclose(infile)
	fclose(outfile)
	fclose(htmlfile)
}