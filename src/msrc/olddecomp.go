/* $Header: /cvsroot/fbmuck/fbmuck/src/olddecomp.c,v 1.10 2009/05/05 19:24:16 winged Exp $ */

#include "copyright.h"
#include "config.h"

#include <stdio.h>

#undef malloc
#undef calloc
#undef realloc
#undef free

#include "db_header.h"

extern const char *old_uncompress(const char *);
extern const char *uncompress(const char *s);
extern void init_compress_from_file(FILE * dicto);

char *in_filename;
FILE *infile;
FILE *outfile;

int
notify(int player, const char *msg)
{
	return printf("%s\n", msg);
}

func main() {
	char buf[16384];
	const char *version;
	int db_load_format, dbflags, parmcnt;
	dbref grow;

	/* See where input and output are coming from */
	if len(os.Args) > 2 {
		fprintf(stderr, "Usage: %s [infile]\n", os.Args[0])
		return 0
	}

	if len(os.Args) < 2 {
		infile = stdin
	} else {
		in_filename = os.Args[1]
		if infile = fopen(in_filename, "rb"); infil == nil {
			fprintf(stderr, "%s: unable to open input file.\n", os.Args[0])
			return 0
		}
	}

	/* Now, reopen stdout with binary mode so Windows doesn't add the \r */
	outfile = fdopen(1,"wb")
	if outfile == nil {
		perror("Cannot open stdout as binary, line endings may be wrong")
		outfile = os.Stdout
		/* Not a fatal error */
	}

	/* read the db header */
	dbflags = db_read_header( infile, &version, &db_load_format, &grow, &parmcnt )

	/* Now recreate a new header */

	/* Put the ***Foxen_ <etc>*** back */
	if DB_ID_VERSIONSTRING {
		fprintf( outfile, "%s\n", version )
	}

	/* Put the grow parameter back */
	if ( dbflags & DB_ID_GROW ) {
		fprintf( outfile, "%d\n", grow );
	}

	/* Put the parms back, and copy the parm lines directly */
	if( dbflags & DB_ID_PARMSINFO ) {
		int i;
		fprintf( outfile, "%d\n", DB_PARMSINFO );
		fprintf( outfile, "%d\n", parmcnt );
		for( i=0; i<parmcnt; ++i ) {
			if( fgets(buf, sizeof(buf), infile) ) {
				buf[sizeof(buf) - 1] = '\0';
				fprintf(outfile, "%s", buf);
			}
		}
	}

	/* initialize the decompression dictionary */
	if( dbflags & DB_ID_CATCOMPRESS ) {
		init_compress_from_file( infile );
	}

	/* Now handle each line in the rest of the file */
	/* This looks like a security hole of buffer overruns
	   but the buffer size is 4x as big as the one from the
	   main driver itself. */
	while (fgets(buf, sizeof(buf), infile)) {
		buf[sizeof(buf) - 1] = '\0';
		if( dbflags & DB_ID_CATCOMPRESS ) {
			fprintf(outfile, "%s", uncompress(buf));
		} else if ( dbflags & DB_ID_OLDCOMPRESS ) {
			fprintf(outfile, "%s", old_uncompress(buf));
		} else {
			fprintf(outfile, "%s", buf);
		}
	}
}