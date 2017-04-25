package OldDecomp

var in_filename string
var infile, outfile *FILE

func notify(player int, msg string) int {
	return fmt.Println(msg)
}

func main() {
	char buf[16384]
	var version string
	var db_load_format, dbflags, parmcnt int
	var grow ObjectID

	/* See where input and output are coming from */
	if len(os.Args) > 2 {
		log.Println("Usage: ", os.Args[0], " [infile]")
		return 0
	}

	var e error
	if len(os.Args) < 2 {
		infile = os.Stdin
	} else {
		if infile, e = os.Open(os.Args[1]); e != nil {
			log.Printf("%s: unable to open input file.\n", os.Args[0])
			return 0
		}
	}

	/* read the db header */
	dbflags = db_read_header(infile, &version, &db_load_format, &grow, &parmcnt)

	/* Now recreate a new header */

	/* Put the ***Foxen_ <etc>*** back */
	if DB_ID_VERSIONSTRING {
		fmt.Fprintf(outfile, "%s\n", version)
	}

	/* Put the grow parameter back */
	if dbflags & DB_ID_GROW != 0 {
		fmt.Fprintf(outfile, "%d\n", grow)
	}

	/* Put the parms back, and copy the parm lines directly */
	if dbflags & DB_ID_PARMSINFO != 0 {
		fmt.Fprintf(outfile, "%d\n", DB_PARMSINFO)
		fmt.Fprintf(outfile, "%d\n", parmcnt)
		for i := 0; i < parmcnt; i++ {
			if fgets(buf, sizeof(buf), infile) {
				buf[sizeof(buf) - 1] = '\0'
				fmt.Fprint(outfile, buf)
			}
		}
	}

	/* initialize the decompression dictionary */
	if dbflags & DB_ID_CATCOMPRESS != 0 {
		init_compress_from_file(infile)
	}

	/* Now handle each line in the rest of the file */
	/* This looks like a security hole of buffer overruns
	   but the buffer size is 4x as big as the one from the
	   main driver itself. */
	for fgets(buf, sizeof(buf), infile) {
		buf[sizeof(buf) - 1] = '\0'
		switch {
		case dbflags & DB_ID_CATCOMPRESS != 0:
			fmt.Fprint(outfile, uncompress(buf))
		case dbflags & DB_ID_OLDCOMPRESS != 0:
			fmf.Fprint(outfile, old_uncompress(buf))
		default:
			fmt.Fprint(outfile, buf)
		}
	}
}