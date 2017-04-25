/* mcp.c: MUD Client Protocol.
   Part of the FuzzBall distribution. */

#define MCP_MESG_PREFIX		"#$#"
#define MCP_QUOTE_PREFIX	"#$\""
#define MCP_ARG_EMPTY		"\"\""
#define MCP_INIT_PKG		"mcp"
#define MCP_DATATAG			"_data-tag"
#define MCP_INIT_MESG		"mcp "
#define MCP_NEGOTIATE_PKG	"mcp-negotiate"

McpPkg *mcp_PackageList = NULL;

/* Used internally to escape and quote argument values. */
func msgarg_escape(in string) (buf string) {
	buf += '"'
	for in != "" {
		switch in[0] {
		case '"', '\\':
			buf += '\\'
		}
		buf += in[0]
	}
	buf += '"';
	return
}

/*****************************************************************/
/***                          ************************************/
/*** MCP PACKAGE REGISTRATION ************************************/
/***                          ************************************/
/*****************************************************************/

func mcp_package_register(pkgname string, minver, maxver McpVer, callback McpPkg_CB, context interface{}) {
	p := &McpPkg{ pkgname: pkgname, minver: minver, maxver: maxver, callback: callback, context: context }
	mcp_package_deregister(pkgname)
	p.next = mcp_PackageList
	mcp_PackageList = p
	mcp_frame_package_renegotiate(pkgname)
}

/*****************************************************************
 *
 * void mcp_package_deregister(
 *              const char* pkgname,
 *          );
 *
 *
 *****************************************************************/

func mcp_package_deregister(pkgname string) {
	ptr := mcp_PackageList
	for ; ptr != nil && strings.EqualFold(ptr.pkgname, pkgname); ptr = mcp_PackageList {
		mcp_PackageList = ptr.next
		if ptr.cleanup {
			ptr.cleanup(ptr.context)
		}
		if ptr.pkgname != "" {
			free(ptr.pkgname)
		}
		free(ptr)
	}

	prev := mcp_PackageList
	if ptr != nil {
		ptr = ptr.next
	}

	for ptr != nil {
		if strings.EqualFold(pkgname, ptr.pkgname) {
			prev.next = ptr.next
			if ptr.cleanup {
				ptr.cleanup(ptr.context)
			}
			if ptr.pkgname {
				free(ptr.pkgname)
			}
			free(ptr)
			ptr = prev.next
		} else {
			prev = ptr
			ptr = ptr.next
		}
	}
	mcp_frame_package_renegotiate(pkgname)
}

//	Initializes MCP globally at startup time.
func mcp_initialize() {
	oneoh := &McpVer{ 1, 0 }
	twooh := &McpVer{ 2, 0 }
	mcp_package_register(MCP_NEGOTIATE_PKG, oneoh, twooh, mcp_negotiate_handler, nil)
	mcp_package_register("org-fuzzball-help", oneoh, oneoh, mcppkg_help_request, nil)
	mcp_package_register("org-fuzzball-notify", oneoh, oneoh, mcppkg_simpleedit, nil)
	mcp_package_register("org-fuzzball-simpleedit", oneoh, oneoh, mcppkg_simpleedit, nil)
	mcp_package_register("dns-org-mud-moo-simpleedit", oneoh, oneoh, mcppkg_simpleedit, nil)
}

//	Starts MCP negotiations, if any are to be had.
func mcp_negotiation_start(mfr *McpFrame) {
	mfr.enabled = true
	reply := &McpMesg{ package: MCP_INIT_PKG }
	mcp_mesg_arg_append(&reply, "version", "2.1")
	mcp_mesg_arg_append(&reply, "to", "2.1")
	mcp_frame_output_mesg(mfr, &reply)
	mfr.enabled = false
}

/*****************************************************************/
/***                       ***************************************/
/*** MCP CONNECTION FRAMES ***************************************/
/***                       ***************************************/
/*****************************************************************/

type McpFrameList struct {
	mfr *McpFrame
	next *McpFrameList
}
var mcp_frame_list *McpFrameList

//	Cleans up an McpFrame for a closing connection.
//	You MUST call this when you are done using an McpFrame.
func mcp_frame_clear(mfr *McpFrame) {
	mfr.authkey = ""
	for p := mfr.packages; p != nil; p = mfr.packages {
		mfr.packages = p.next
		p.pkgname = nil
	}
	for p := mfr.messages; p != nil; p = mfr.messages {
		mfr.messages = p.next
	}
	for mfrl := mcp_frame_list; mfrl != nil && mfrl.mfr == mfr; mfrl = mcp_frame_list {
		mcp_frame_list = mfrl.next
	}
	if mcp_frame_list != nil {
		prev := mcp_frame_list
		for mfrl := prev.next; mfrl != nil; {
			if mfrl.mfr == mfr {
				prev.next = mfrl.next
				mfrl = prev.next;
			} else {
				prev = mfrl
				mfrl = mfrl.next
			}
		}
	}
}

/*****************************************************************
 *
 * void mcp_frame_package_renegotiate(
 *              McpFrame* mfr,
 *              char* package
 *          );
 *
 *   Removes a package from the list of supported packages
 *   for all McpFrames, and initiates renegotiation of that
 *   package.
 *
 *****************************************************************/

func mcp_frame_package_renegotiate(const char* package) {
	p := mcp_PackageList
	for ; p != nil && !strings.EqualFold(p.pkgname, package); p = p.next {}

	cando := &McpMesg{ package: MCP_NEGOTIATE_PKG, mesgname: "can" }
	if p == nil {
		mcp_mesg_arg_append(&cando, "package", package)
		mcp_mesg_arg_append(&cando, "min-version", "0.0")
		mcp_mesg_arg_append(&cando, "max-version", "0.0")
	} else {
		mcp_mesg_arg_append(&cando, "package", p.pkgname)
		mcp_mesg_arg_append(&cando, "min-version", fmt.Sprintf("%d.%d", p.minver.major, p.minver.minor))
		mcp_mesg_arg_append(&cando, "max-version", fmt.Sprintf("%d.%d", p.maxver.major, p.maxver.minor))
	}

	for mfrl := mcp_frame_list; mfrl != nil; mfrl = mfrl->next {
		if mfr := mfrl.mfr; mfr.enabled {
			if mcp_version_compare(mfr.version, McpVer{}) > 0 {
				mcp_frame_package_remove(mfr, package)
				mcp_frame_output_mesg(mfr, &cando)
			}
		}
	}
}

/*****************************************************************
 *
 * void mcp_frame_package_add(
 *              McpFrame* mfr,
 *              const char* package,
 *              McpVer minver,
 *              McpVer maxver
 *          );
 *
 *   Attempt to register a package for this connection.
 *   Returns EMCP_SUCCESS if the package was deemed supported.
 *   Returns EMCP_NOMCP if MCP is not supported on this connection.
 *   Returns EMCP_NOPACKAGE if the package versions didn't overlap.
 *
 *****************************************************************/

func mcp_frame_package_add(mfr *McpFrame, package string, minver, maxver McpVer) int {
	if !mfr.enabled {
		return EMCP_NOMCP
	}

	var ptr *McpPkg
	for ptr = mcp_PackageList; ptr != nil && !strings.EqualFold(ptr.pkgname, package); ptr = ptr.next {}
	if ptr == nil {
		return EMCP_NOPACKAGE
	}

	mcp_frame_package_remove(mfr, package)
	selver := mcp_version_select(ptr->minver, ptr->maxver, minver, maxver);
	nu := &McpPkg{ pkgname: ptr.pkgname, minver: selver, maxver: selver, callback: ptr.callback, context: ptr.context }
	if mfr.packages == nil {
		mfr.packages = nu
	} else {
		lastpkg := mfr.packages
		for ; lastpkg.next != nil; lastpkg = lastpkg.next {}
		lastpkg.next = nu
	}
	return EMCP_SUCCESS
}

/*****************************************************************
 *
 * void mcp_frame_package_remove(
 *              McpFrame* mfr,
 *              char* package
 *          );
 *
 *   Removes a package from the list of supported packages
 *   for this McpFrame.
 *
 *****************************************************************/

func mcp_frame_package_remove(mfr *McpFrame, package string) {
	for mfr.packages != nil && strings.EqualFold(mfr.packages.pkgname, package) {
		tmp := mfr.packages
		mfr.packages = tmp.next
		if tmp.pkgname != "" {
			free(tmp.pkgname)
		}
		free(tmp)
	}

	for prev := mfr.packages; prev != nil && prev.next != nil; {
		if strings.EqualFold(prev.next.pkgname, package) {
			tmp := prev.next
			prev.next = tmp.next
			if tmp.pkgname != "" {
				free(tmp.pkgname)
			}
			free(tmp)
		} else {
			prev = prev.next
		}
	}
}

/*****************************************************************
 *
 * McpVer mcp_frame_package_supported(
 *              McpFrame* mfr,
 *              char* package
 *          );
 *
 *   Returns the supported version of the given package.
 *   Returns {0,0} if the package is not supported.
 *
 *****************************************************************/

func mcp_frame_package_supported(mfr *McpFrame, package string) (r McpVer) {
	if mfr.enabled {
		var ptr *McpPkg
		for ptr = mfr.packages; ptr != nil && !strings.EqualFold(ptr.pkgname, package); ptr = ptr.next {}
		if ptr != nil {
			r = ptr.minver
		}
	}
	return
}

/*****************************************************************
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
 *****************************************************************/

func mcp_frame_package_docallback(mfr *McpFrame, msg *McpMesg) int {
	if strings.EqualFold(msg.package, MCP_INIT_PKG) {
		mcp_basic_handler(mfr, msg, nil)
	} else {
		if !mfr.enabled {
			return = EMCP_NOMCP
		}

		var ptr *McpPkg
		for ptr = mfr.packages; ptr != nil && !strings.EqualFold(ptr.pkgname, msg.package); ptr = ptr.next {}
		if ptr == nil {
			if strings.EqualFold(msg.package, MCP_NEGOTIATE_PKG) {
				mcp_negotiate_handler(mfr, msg, &McpVer{ 2, 0 }, nil)
			} else {
				return EMCP_NOPACKAGE
			}
		} else {
			ptr.callback(mfr, msg, ptr.maxver, ptr.context)
		}
	}
	return EMCP_SUCCESS
}

/*****************************************************************
 *
 *   Check a line of input for MCP commands.
 *   outbuf will contain the in-band data on return, if any.
 *
 *****************************************************************/

func mcp_frame_process_input(mfr *McpFrame, linein string) (buf string, inband bool) {
	switch {
	case strings.EqualFold(linein[:3], MCP_MESG_PREFIX[:3]):
		/* treat it as an out-of-band message, and parse it. */
		if mfr.enabled || strings.EqualFold(MCP_INIT_MESG[:3], (linein[3:])[:4]) {
			if !mcp_internal_parse(mfr, linein[3:]) {
				buf, inband = linein, true
			}
		} else {
			buf, inband = linein, true
		}
	case mfr.enabled && strings.EqualFold(linein[:3], MCP_QUOTE_PREFIX[:3]):
		/* It's quoted in-band data.  Strip the quoting. */
		buf, inband = linein[3:], true
	default:
		/* It's in-band data.  Return it raw. */
		buf, inband = linein, true
	}
	return
}

func mcp_frame_output_inband(mfr *McpFrame, lineout string) {
	if mfr.enabled && (strings.HasPrefix(lineout, MCP_MESG_PREFIX) || strings.HasPrefix(lineout, MCP_QUOTE_PREFIX)) {
		mfr.descriptor.QueueWrite(MCP_QUOTE_PREFIX)
	}
	mfr.descriptor.QueueWrite(lineout)
}

/*****************************************************************
 *
 * int mcp_frame_output_mesg(
 *             McpFrame* mfr,
 *             McpMesg* msg
 *         );
 *
 *   Sends an MCP message to the given connection.
 *   Returns EMCP_SUCCESS if successful.
 *   Returns EMCP_NOMCP if MCP isn't supported on this connection.
 *   Returns EMCP_NOPACKAGE if this connection doesn't support the needed package.
 *
 *****************************************************************/

func mcp_frame_output_mesg(McpFrame * mfr, McpMesg * msg) int {
	char *p;

	if !mfr.enabled && !strings.EqualFold(msg.package, MCP_INIT_PKG) {
		return EMCP_NOMCP
	}

	var mesgname string
	if msg.mesgname != "" {
		mesgname = fmt.Sprintf("%s-%s", msg.package, msg.mesgname)
	} else {
		mesgname = fmt.Sprint(msg.package)
	}

	outbuf := MCP_MESG_PREFIX + mesgname
	if !strings.EqualFold(mesgname, MCP_INIT_PKG) {
		nullver := new(McpVer)
		outbuf += " " + mfr.authkey
		if !strings.EqualFold(msg.package, MCP_NEGOTIATE_PKG) {
			ver := mcp_frame_package_supported(mfr, msg.package)
			if mcp_version_compare(ver, nullver) == 0 {
				return EMCP_NOPACKAGE
			}
		}
	}

	/* If the argument lines contain newlines, split them into separate lines. */
	for anarg := msg.args; anarg != nil; anarg = anarg.next {
		if anarg.value != nil {
			for ap := anarg.value; ap != nil; ap = ap.next {
				count := 0
				for p = ap.value; p != ""; count++ {
					switch p[0] {
					case '\n', '\r':
						nu := &McpArgPart{ next: ap.next, value: p[1:] }
						ap.value = ap.value[:count]
						ap.next = nu
						ap = nu
						p = nu.value
					} else {
						p = p[1:]
					}
				}
			}
		}
	}

	/* Build the initial message string */
	var mlineflag bool
	for anarg := msg.args; anarg != nil; anarg = anarg.next {
		switch {
		case anarg.value == nil:
			anarg.was_shown = true
			outbuf += fmt.Sprintf(" %s: %s", anarg.name, MCP_ARG_EMPTY)
		case anarg.value.next != "":
			/* Value is multi-line.  Send on separate line(s). */
			mlineflag = true
			anarg.was_shown = 0
			outbuf += fmt.Sprintf(" %s*: %s", anarg.name, MCP_ARG_EMPTY)
		default:
			anarg.was_shown = true
			outbuf += fmt.Sprintf(" %s: %s", anarg.name, msgarg_escape(anarg.value.value))
		}
	}

	/* If the message is multi-line, make sure it has a _data-tag field. */
	var datatag string
	if mlineflag {
		datatag = fmt.Sprintf("%.8lX", (unsigned long)(RANDOM() ^ RANDOM()))
		outbuf += fmt.Sprintf(" %s: %s", MCP_DATATAG, datatag)
	}

	/* Send the initial line. */
	mfr.descriptor.QueueWrite(outbuf)
	mfr.descriptor.QueueWrite("\r\n")

	if mlineflag {
		/* Start sending arguments whose values weren't already sent. */
		/* This is usually just multi-line argument values. */
		flushcount := 8
		for anarg := msg.args; anarg != nil; anarg = anarg.next {
			if !anarg.was_shown {
				for ap := anarg.value; ap != nil; ap = ap.next {
					mfr.descriptor.QueueWrite(fmt.Sprintf("%s* %s %s: %s", MCP_MESG_PREFIX, datatag, anarg.name, ap.value))
					mfr.descriptor.QueueWrite("\r\n")
					flushcount--
					if flushcount == 0 {
						FlushText(mfr)
						flushcount = 8
					}
				}
			}
		}

		/* Let the other side know we're done sending multi-line arg vals. */
		mfr.descriptor.QueueWrite(fmt.Sprintf("%s: %s", MCP_MESG_PREFIX, datatag))
		mfr.descriptor.QueueWrite("\r\n")
	}
	return EMCP_SUCCESS
}

/*****************************************************************/
/***                  ********************************************/
/*** ARGUMENT METHODS ********************************************/
/***                  ********************************************/
/*****************************************************************/



/*****************************************************************
 *
 * int mcp_mesg_arg_linecount(
 *         McpMesg* msg,
 *         const char* name
 *     );
 *
 *   Returns the count of the number of lines in the given arg of
 *   the given message.
 *
 *****************************************************************/

func mcp_mesg_arg_linecount(McpMesg * msg, const char *name) (r int) {
	ptr := msg.args
	for ; ptr != nil && !strings.EqualFold(ptr.name, name); ptr = ptr.next {}
	if ptr != nil {
		ptr2 := ptr.value
		for ptr2 != nil {
			ptr2 = ptr2.next
			r++
		}
	}
	return
}

//	Gets the value of a named argument in the given message.
func mcp_mesg_arg_getline(msg *McpMesg, argname string, linenum int) (r string) {
	ptr := msg.args
	for ; ptr != nil && !strings.EqualFold(ptr.name, argname); ptr = ptr.next {}
	if ptr != nil {
		ptr2 := ptr.value
		for linenum > 0 && ptr2 != nil {
			ptr2 = ptr2.next
			linenum--
		}
		if ptr2 != nil {
			r = ptr2.value
		}
	}
	return
}

//	Appends to the list value of the named arg in the given mesg.
//	If that named argument doesn't exist yet, it will be created.
//	This is used to construct arguments that have lists as values.
//	Returns the success state of the call.  EMCP_SUCCESS if the
//	call was successful.  EMCP_ARGCOUNT if this would make too
//	many arguments in the message.  EMCP_ARGLENGTH is this would
//	cause an argument to exceed the max allowed number of lines.
func mcp_mesg_arg_append(msg *McpMesg, argname, argval string) int {
	if len(argname) > MAX_MCP_ARGNAME_LEN) {
		return EMCP_ARGNAMELEN
	}
	if len(argval) + msg.bytes > MAX_MCP_MESG_SIZE {
		return EMCP_MESGSIZE;
	}
	ptr := msg.args
	for ; ptr != nil && !strings.EqualFold(ptr.name, argname); ptr = ptr.next {}
	if ptr == nil {
		if len(argname) + len(argval) + msg.bytes > MAX_MCP_MESG_SIZE {
			return EMCP_MESGSIZE
		}
		ptr = &McpArg{ name: argname }
		if msg.args == nil {
			msg.args = ptr
		} else {
			limit := MAX_MCP_MESG_ARGS
			lastarg := msg.args
			for ; lastarg.next != nil; lastarg = lastarg.next {
				if limit <= 0 {
					return EMCP_ARGCOUNT
				}
				limit--
			}
			lastarg.next = ptr
		}
		msg.bytes += sizeof(McpArg) + namelen + 1
	}

	if len(argval) > 0 {
		nu := &McpArgPart{ value: argval }
		if ptr.last == nil {
			ptr.last = nu
			ptr.value = ptr.last
		} else {
			ptr.last.next = nu
			ptr.last = ptr.last.next
		}
		msg.bytes += sizeof(McpArgPart) + vallen + 1
	}
	ptr.was_shown = false
	return EMCP_SUCCESS
}

/*****************************************************************
 *
 * void mcp_mesg_arg_remove(
 *         McpMesg* msg,
 *         const char* argname
 *     );
 *
 *   Removes the named argument from the given message.
 *
 *****************************************************************/

func mcp_mesg_arg_remove(msg *McpMesg, argname string) {
	var ptr *McpArg
	for ptr = msg.args; ptr != nil && strings.EqualFold(ptr.name, argname); ptr = msg.args {
		msg.args = ptr.next
		msg.bytes -= sizeof(McpArg)
		if ptr.name {
			ptr.NowCalled("")
			msg.bytes -= len(ptr.name) + 1
		}
		for ptr.value != nil {
			ptr2 := ptr.value
			ptr.value = ptr.value.next
			msg.bytes -= sizeof(McpArgPart)
			if ptr2.value != nil {
				msg.bytes -= len(ptr2.value) + 1
				ptr2.value = nil
			}
		}
	}

	prev := msg.args
	if ptr != nil {
		ptr = ptr.next
	}

	for ptr != nil {
		if strings.EqualFold(argname, ptr.name) {
			prev.next = ptr.next
			msg.bytes -= sizeof(McpArg)
			if ptr.name != "" {
				msg.bytes -= len(ptr.name) + 1
				ptr.NowCalled("")
			}
			for ptr.value != nil {
				ptr2 := ptr.value
				ptr.value = ptr.value.next
				msg.bytes -= sizeof(McpArgPart)
				if ptr2.value != nil {
					msg.bytes -= len(ptr2.value) + 1
					ptr2.value = ""
				}
			}
			ptr = prev.next
		} else {
			prev = ptr
			ptr = ptr.next
		}
	}
}

/*****************************************************************/
/***                 *********************************************/
/*** VERSION METHODS *********************************************/
/***                 *********************************************/
/*****************************************************************/




/*****************************************************************
 *
 * int mcp_version_compare(McpVer v1, McpVer v2);
 *
 *   Compares two McpVer structs.
 *   Results are similar to strcmp():
 *     Returns negative if v1 <  v2
 *     Returns 0 (zero) if v1 == v2
 *     Returns positive if v1 >  v2
 *
 *****************************************************************/

func mcp_version_compare(McpVer v1, McpVer v2) (r int) {
	if v1.major != v2.major {
		r = (v1.major - v2.major)
	} else {
		r = v1.minor - v2.minor
	}
	return
}

/*****************************************************************
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

McpVer
mcp_version_select(McpVer min1, McpVer max1, McpVer min2, McpVer max2)
{
	McpVer result = { 0, 0 };

	if (mcp_version_compare(min1, max1) > 0) {
		return result;
	}
	if (mcp_version_compare(min2, max2) > 0) {
		return result;
	}
	if (mcp_version_compare(min1, max2) > 0) {
		return result;
	}
	if (mcp_version_compare(min2, max1) > 0) {
		return result;
	}
	if (mcp_version_compare(max1, max2) > 0) {
		return max2;
	} else {
		return max1;
	}
}

/*****************************************************************/
/***                       ***************************************/
/***  MCP PACKAGE HANDLER  ***************************************/
/***                       ***************************************/
/*****************************************************************/

func mcp_basic_handler(mfr *McpFrame, mesg *McpMesg, dummy interface{}) {
	myminver := McpVer{ 2, 1 }
	mymaxver := McpVer{ 2, 1 }
	minver := McpVer{ 0, 0 }
	maxver := McpVer{ 0, 0 }
	nullver := McpVer{ 0, 0 }

	if mesg.mesgname == "" {
		auth := mcp_mesg_arg_getline(mesg, "authentication-key", 0);
		if auth != "" {
			mfr.authkey = auth
		} else {
			reply := &McpMesg{ package: MCP_INIT_PKG }
			mcp_mesg_arg_append(&reply, "version", "2.1")
			mcp_mesg_arg_append(&reply, "to", "2.1")
			authval := fmt.Sprintf("%.8lX", (unsigned long)(RANDOM() ^ RANDOM()))
			mcp_mesg_arg_append(&reply, "authentication-key", authval)
			mfr.authkey = authval
			mcp_frame_output_mesg(mfr, &reply)
		}

		if ptr := mcp_mesg_arg_getline(mesg, "version", 0); ptr != "" {
			for ; len(ptr) > 0 && isdigit(ptr[0]); ptr = ptr[1:] {
				minver.major = (minver.major * 10) + (ptr[0] - '0')
			}
			if ptr[0] == '.' {
				for ptr = ptr[1:]; len(ptr) > 0 && isdigit(*ptr); ptr = ptr[1:] {
					minver.minor = (minver.minor * 10) + (ptr[0] - '0')
				}
				if ptr = mcp_mesg_arg_getline(mesg, "to", 0); ptr == "" {
					maxver = minver
				} else {
					for ; len(ptr) > 0 && isdigit(ptr[0]); ptr = ptr[1:] {
						maxver.major = (maxver.major * 10) + (ptr[0] - '0')
					}
					if ptr[0] != '.' {
						return
					}
					for ptr = ptr[1:]; len(ptr) > 0 && isdigit(ptr[0]); ptr = ptr[1:] {
						maxver.minor = (maxver.minor * 10) + (ptr[0] - '0')
					}
				}

				mfr.version = mcp_version_select(myminver, mymaxver, minver, maxver);
				if mcp_version_compare(mfr.version, nullver) {
					char verbuf[32];
					McpPkg *p = mcp_PackageList;

					mfr.enabled = true
					for p := mcp_PackageList; p != nil; p = p.next {
						if !strings.EqualFold(p.pkgname, MCP_INIT_PKG) {
							cando := &McpMesg{ package: MCP_NEGOTIATE_PKG, mesgname: "can" }
							mcp_mesg_arg_append(&cando, "package", p.pkgname)
							mcp_mesg_arg_append(&cando, "min-version", fmt.Sprintf("%d.%d", p.minver.major, p.minver.minor))
							mcp_mesg_arg_append(&cando, "max-version", fmt.Sprintf("%d.%d", p.maxver.major, p.maxver.minor))
							mcp_frame_output_mesg(mfr, &cando)
						}
					}
					cando = &McpMesg{ package: MCP_NEGOTIATE_PKG, mesgname: "end" }
					mcp_frame_output_mesg(mfr, &cando);
				}
			}
		}
	}
}





/*****************************************************************/
/***                                 *****************************/
/***  MCP-NEGOTIATE PACKAGE HANDLER  *****************************/
/***                                 *****************************/
/*****************************************************************/

func mcp_negotiate_handler(McpFrame * mfr, McpMesg * mesg, McpVer version, void *dummy) {
	var minver, maxver McpVer
	if mesg.mesgname == "can" {
		if pkg := mcp_mesg_arg_getline(mesg, "package", 0); pkg != nil {
			if ptr := mcp_mesg_arg_getline(mesg, "min-version", 0); ptr != nil {
				for isdigit(*ptr) {
					minver.major = (minver.major * 10) + (*ptr++ - '0')
				}
				if *ptr == '.' {
					ptr++
					for isdigit(*ptr) {
						minver.minor = (minver.minor * 10) + (*ptr++ - '0')
					}

					if ptr = mcp_mesg_arg_getline(mesg, "max-version", 0); ptr == nil {
						maxver = minver
					} else {
						for isdigit(*ptr) {
							maxver.major = (maxver.major * 10) + (*ptr++ - '0')
						}
						if *ptr != '.' {
							return
						}
						ptr++
						for isdigit(*ptr) {
							maxver.minor = (maxver.minor * 10) + (*ptr++ - '0')
						}
					}
					mcp_frame_package_add(mfr, pkg, minver, maxver)
				}
			}
		}
	}
}





/*****************************************************************/
/****************                *********************************/
/**************** INTERNAL STUFF *********************************/
/****************                *********************************/
/*****************************************************************/

func mcp_intern_is_ident(in string) (buf string, ok bool) {
	if unicode.IsAlpha(in[0]) || in[0] != '_' {
		for _, r := range in[1:] {
			if unicode.IsAlpha(r) || r == '_' || unicode.IsDigit(r) || r == "-" {
				buf = append(buf, r)
			}
		}
		ok = true
	}
	return
}

func mcp_intern_is_simplechar(in rune) bool {
	return in != '*' && in != ':' && in != '\\' && in != '"' && in != ' ' && unicode.IsPrint(in)
}

func mcp_intern_is_unquoted(in string) (buf string, ok int) {
	int origbuflen = buflen;

	if mcp_intern_is_simplechar(in[0]) {
		for _, c := range in[1:] {
			if !mcp_intern_is_simplechar(c) {
				break
			}
			buf = append(buf, c)
		}
		ok = true
	}
	return
}

func mcp_intern_is_quoted(in string) (buf string, r bool) {
	old = in
	if in[0] == '""' {
		in = in[1:]
		for _, c := range in {
			switch c {
			case '\\':
				//	escape code - consume next character (???)
				buf = append(buf, c++)
			case '"':
				break
			default:
				buf = append(buf, c++)
			}
		}

		if (**in == '"') {
			(*in)++;
			r = true
		} else {
			*in = old
		}
	}
	return
}

func mcp_intern_is_keyval(msg *McpMesg, in string) (r bool) {
	if unicode.IsSpace(in[0]) {
		in = strings.TrimLeftFunc(in, unicode.IsSpace)
		if keyname, ok := mcp_intern_is_ident(in); ok {
			var deferred bool
			if in[0] == '*' {
				msg.incomplete = true
				deferred = true
				in = in[1:]
			}
			if in[0] == ':' {
				in = in[1:]
				if unicode.IsSpace(in[0]) {
					in = strings.TrimLeftFunc(in, unicode.IsSpace)
					var value string
					if value, ok = mcp_intern_is_unquoted(in); !ok {
						if value, ok = mcp_intern_is_quoted(in); !ok {
							return false
						}
					}

					if deferred {
						mcp_mesg_arg_append(msg, keyname, nil);
					} else {
						mcp_mesg_arg_append(msg, keyname, value);
					}
					r = true
				}
			}
		}
	}
	return
}

func mcp_intern_is_mesg_start(mfr *McpFrame, in string) bool {
	char *subname = NULL;
	McpMesg *newmsg = NULL;
	McpPkg *pkg = NULL;
	int longlen = 0;

	var mesgname string
	var ok bool
	if mesgname, ok = mcp_intern_is_ident(&in, mesgname, sizeof(mesgname)); !ok {
		return false
	}
	if !strings.EqualFold(mesgname, MCP_INIT_PKG) {
		if !unicode.IsSpace(in[0]) {
			return false
		}
		in = strings.TrimLeftFunc(in, unicode.IsSpace)
		var authkey string
		if authkey, ok = mcp_intern_is_unquoted(&in, authkey, sizeof(authkey)); !ok {
			return false
		}
		if authkey != mfr.authkey {
			return false
		}
	}

	if !strings.EqualFold(mesgname[:3], MCP_INIT_PKG[:3]) {
		for pkg = mfr.packages; pkg != nil; pkg = pkg.next {
			i := len(pkg.pkgname)
			if strings.EqualFold(pkg.pkgname[:i], mesgname[:i]) {
				if len(mesgname) == i || mesgname[i] == '-' {
					if i > longlen {
						longlen = i
					}
				}
			}
		}
	}
	if (!longlen) {
		int neglen = len(MCP_NEGOTIATE_PKG);

		switch {
		case strings.EqualFold(mesgname[:neglen], MCP_NEGOTIATE_PKG[:neglen]):
			longlen = neglen
		case strings.EqualFold(mesgname, MCP_INIT_PKG):
			longlen = len(mesgname)
		default:
			return false
		}
	}
	subname = mesgname + longlen;
	if (*subname) {
		*subname++ = '\0';
	}

	newmsg = &McpMesg{ package: mesgname, mesgname: subname }
	while (*in) {
		if !mcp_intern_is_keyval(newmsg, &in) {
			free(newmsg);
			return false
		}
	}

	/* Okay, we've recieved a valid message. */
	if (newmsg->incomplete) {
		/* It's incomplete.  Remember it to finish later. */
		const char *msgdt = mcp_mesg_arg_getline(newmsg, MCP_DATATAG, 0);

		newmsg->datatag = msgdt;
		mcp_mesg_arg_remove(newmsg, MCP_DATATAG);
		newmsg->next = mfr->messages;
		mfr->messages = newmsg;
	} else {
		/* It's complete.  Execute the callback function for this package. */
		mcp_frame_package_docallback(mfr, newmsg);
		free(newmsg);
	}
	return true
}

func mcp_intern_is_mesg_cont(mfr *McpFrame, in string) (r bool) {
	if in[0] == '*' {
		in = in[1:]
		if unicode.IsSpace(in[0]) {
			in = strings.TrimLeftFunc(in, unicode.IsSpace)
			if datatag, ok := mcp_intern_is_unquoted(in); ok {
				if unicode.IsSpace(in[0]) {
					in = strings.TrimLeftFunc(in, unicode.IsSpace)
					if keyname, ok := mcp_intern_is_ident(in); ok {
						if in[0] == ':' {
							in = in[1:]
							if unicode.IsSpace(in[0]) {
								in = in[1:]
								for ptr := mfr.messages; ptr != nil; ptr = ptr.next {
									if datatag == ptr.datatag {
										mcp_mesg_arg_append(ptr, keyname, in)
										r = true
										break
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return
}

func mcp_intern_is_mesg_end(mfr *McpFrame, in string) (r bool) {
	if in[0] == ':' {
		in = in[1:]
		if unicode.IsSpace(in[0]) {
			in = strings.TrimLeftFunc(in, unicode.IsSpace)
			if datatag, ok := mcp_intern_is_unquoted(in); !ok {
				if len(in) == 0 {
					var ptr *McpMesg
					prev := &(mfr.messages)
					for ptr = mfr.messages; ptr != nil; ptr = ptr.next {
						if datatag == ptr.datatag {
							*prev = ptr.next
							break
						}
						prev = &ptr.next
					}
					if ptr != nil {
						ptr.incomplete = false
						mcp_frame_package_docallback(mfr, ptr)
						free(ptr)
						r = true
					}
				}
			}
		}
	}
	return
}

func mcp_internal_parse(mfr *McpFrame, in string) bool {
	return mcp_intern_is_mesg_cont(mfr, in) || mcp_intern_is_mesg_end(mfr, in) || mcp_intern_is_mesg_start(mfr, in)
}