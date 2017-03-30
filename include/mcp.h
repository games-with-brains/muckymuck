/* the type used to specify the connection */
typedef void *connection_t;

#define MCP_MESG_PREFIX     "#$#"
#define MCP_QUOTE_PREFIX    "#$\""

#define EMCP_SUCCESS		 0	/* successful result */
#define EMCP_NOMCP			-1	/* MCP isn't supported on this connection. */
#define EMCP_NOPACKAGE		-2	/* Package isn't supported for this connection. */
#define EMCP_ARGCOUNT		-3	/* Too many arguments in mesg. */
#define EMCP_ARGNAMELEN		-5	/* Arg name is too long. */
#define EMCP_MESGSIZE		-6	/* Message is too large. */

#define MAX_MCP_ARGNAME_LEN    30 /* max length of argument name. */
#define MAX_MCP_MESG_ARGS      30 /* max number of args per mesg. */
#define MAX_MCP_MESG_SIZE  262144 /* max mesg size in bytes. */

/* This is a convenient struct for dealing with MCP versions. */
type McpVer struct {
	major uint8;		/* major version number */
	minor uint8		/* minor version number */
}

/* This is one line of a multi-line argument value. */
type McpArgPart struct {
	value string
	next *McpArgPart
}

/* This is one argument of a message. */
type McpArg struct {
	name string
	value *McpArgPart
	last *McpArgPart
	was_shown bool
	next *McpArg
}

/* This is an MCP message. */
type McpMesg struct {
	package string
	mesgname string
	datatag string
	args *McpArg
	incomplete int
	bytes int
	next *McpMesg
}


struct McpFrame_T;
typedef void (*McpPkg_CB) (struct McpFrame_T * mfr,

						   McpMesg * mesg, McpVer version, void *context);

typedef void (*ContextCleanup_CB) (void *context);



/* This is used to keep track of registered packages. */
typedef struct McpPkg_T {
	char *pkgname;				/* Name of the package */
	McpVer minver;				/* min supported version number */
	McpVer maxver;				/* max supported version number */
	McpPkg_CB callback;			/* function to call with mesgs */
	void *context;				/* user defined callback context */
	ContextCleanup_CB cleanup;  /* callback to use to free context */
	struct McpPkg_T *next;
} McpPkg;



/* This keeps connection specific data for MCP. */
typedef struct McpFrame_T {
	void *descriptor;			/* The descriptor to send output to */
	unsigned int enabled;		/* Flag denoting if MCP is enabled. */
	char *authkey;				/* Authorization key. */
	McpVer version;				/* Supported MCP version number. */
	McpPkg *packages;			/* Pkgs supported on this connection. */
	McpMesg *messages;			/* Partial messages, under construction. */
} McpFrame;




/*****************************************************************
 *
 * void mcp_package_deregister(
 *              const char* pkgname,
 *          );
 *
 *
 *****************************************************************
 *
 * void mcp_frame_package_add(
 *              McpFrame* mfr,
 *              char* package,
 *              McpVer minver,
 *              McpVer maxver
 *          );
 *
 *   Attempt to register a package for this connection.
 *   Returns EMCP_SUCCESS if the package was deemed supported.
 *   Returns EMCP_NOMCP if MCP is not supported on this connection.
 *   Returns EMCP_NOPACKAGE if the package versions didn't overlap.
 *
 *****************************************************************
 *
 * void mcp_frame_package_remove(
 *              McpFrame* mfr,
 *              char* package,
 *              McpVer minver,
 *              McpVer maxver
 *          );
 *
 *   Deregisters a package for a given frame.
 *
 *****************************************************************
 *
 * void mcp_frame_package_supported(
 *              McpFrame* mfr,
 *              char* package
 *          );
 *
 *   Returns the supported version of the given package.
 *   Returns {0,0} if the package is not supported.
 *
 *****************************************************************
 *
 * void mcp_frame_package_docallback(
 *              McpFrame* mfr,
 *              McpMesg* msg
 *          );
 *
 *   Executes the callback function for the given message.
 *   Returns EMCP_SUCCESS if the call completed successfully.
 *   Returns EMCP_NOMCP if MCP is not supported for that connection.
 *   Returns EMCP_NOPACKAGE if the package is not supported.
 *
 *****************************************************************
 *
 * int mcp_frame_process_input(
 *           McpFrame* mfr,
 *           const char* linein,
 *           char *outbuf,
 *           int bufsize
 *      );
 *
 *   Check a line of input for MCP commands.
 *   Returns 0 if the line was an out-of-band MCP message.
 *   Returns 1 if the line was in-band data.
 *     outbuf will contain the in-band data on return, if any.
 *
 *****************************************************************
 *
 * int mcp_frame_output_mesg(
 *             McpFrame* mfr,
 *             McpMesg* msg
 *         );
 *
 *   Sends an MCP message to the given connection.
 *   Returns EMCP_SUCCESS if successful.
 *   Returns EMCP_NOMCP if MCP isn't supported on this connection.
 *   Returns EMCP_NOPACKAGE if this connection doesn't support the package.
 *
 *****************************************************************
 *
 * int mcp_mesg_arg_linecount(
 *         McpMesg* msg,
 *         const char* name
 *     );
 *
 *   Returns the count of the number of lines in the given arg of
 *   the given message.
 *
 *****************************************************************
 *
 * void mcp_mesg_arg_remove(
 *         McpMesg* msg,
 *         const char* argname
 *     );
 *
 *   Removes the named argument from the given message.
 *
 *****************************************************************
 *
 * int mcp_version_compare(McpVer v1, McpVer v2);
 *
 *   Compares two McpVer structs.
 *   Results are similar to strcmp():
 *     Returns negative if v1 <  v2
 *     Returns 0 (zero) if v1 == v2
 *     Returns positive if v1 >  v2
 *
 *****************************************************************
 *
 * McpVer mcp_version_select(
 *                McpVer min1,
 *                McpVer max1,
 *                McpVer min2,
 *                McpVer max2
 *            );
 *
 *   Given the min and max package versions supported by a client
 *     and server, this will return the highest version that is
 *     supported by both.
 *   Returns a McpVer of {0, 0} if there is no version overlap.
 *
 *****************************************************************/