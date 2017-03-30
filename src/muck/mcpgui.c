package fbmuck

type DlogValue struct {
	name string
	value []string
	next *DlogValue
}

func (d *DlogValue) Push(s string) *DlogValue {
	return &DlogValue{ name: s, next: d }
}

func (d *DlogValue) Find(name string) (r *DlogValue) {
	for r = d.values; r != nil && r.name != name; r = r.next {}
	return
}

func (d *DlogValue) FindWithin(name string, limit int) (r *DlogValue) {
	for r = d.values; limit > 0 && r != nil && r.name != name; r = r.next {
		limit--
	}
	if r.name != name {
		r = nil
	}
	return	
}

type DlogData struct {
	id string
	descr int
	dismissed bool
	values *DlogValue
	callback Gui_CB
	error_cb GuiErr_CB
	context interface{}
	next *DlogData
	prev *(*DlogData)
}

var dialog_list *DlogData
var dialog_last_accessed *DlogData


void gui_pkg_callback(McpFrame * mfr, McpMesg * msg, McpVer ver, void *context);


func gui_initialize() {
	McpVer minver = { 1, 0 };  /* { major, minor } */
	McpVer maxver = { 1, 3 };  /* { major, minor } */
	mcp_package_register(GUI_PACKAGE, minver, maxver, gui_pkg_callback, NULL, NULL)
}

func gui_dlog_find(id string) (r *DlogData) {
	r = dialog_last_accessed
	if r == nil || r.id == id {
		for r = dialog_list; r != nil; r = r.next {
			if r.id == dlogid {
				dialog_last_accessed = r
				break
			}
		}
	}
	return
}

func gui_dlog_get_context(id string) (r interface{}) {
	if ptr := gui_dlog_find(id); ptr != nil {
		r = ptr.context
	}
	return
}

func gui_dlog_get_descr(id string) (r int) {
	if ptr := gui_dlog_find(id); ptr != nil {
		r = ptr.descr
	} else {
		r = EGUINODLOG
	}
	return
}

func GuiClosed(id string) (r int) {
	if ptr := gui_dlog_find(id); ptr != nil {
		r = ptr.dismissed
	} else {
		r = EGUINODLOG
	}
	return
}

func gui_value_linecount(dlogid, id string) (r int) {
	if d := gui_dlog_find(dlogid); d == nil {
		r = EGUINODLOG
	} else {
		if ptr := d.Find(id); ptr != nil {
			r = ptr.lines
		}
	}
	return
}

func GuiValueFirst(const char *dlogid) (r string) {
	if ddata := gui_dlog_find(dlogid); d != nil && len(d.values) > 0 {
		r = d.values.name
	}
	return
}

func GuiValueNext(dlogid, id string) (r string) {
	if d := gui_dlog_find(dlogid); d != nil {
		if ptr := d.Find(id); ptr != nil && ptr.next != nil {
			r = ptr.next.name
		}
	}
	return
}

func gui_value_get(dlogid, id string, line int) (r string) {
	if d := gui_dlog_find(dlogid); d != nil {
		if ptr := d.values.Find(id); ptr != nil && line > -1 && line < len(ptr.value) {
			r = ptr.value[line]
		}
	}
	return
}

func gui_value_set_local(dlogid, id string, value []string) {
	if d := gui_dlog_find(dlogid); d != nil {
		if ptr := d.FindWithin(name, 256); ptr != nil {
			ptr.value = make([]string, len(value))
			copy(ptr.value, value)
		} else {
			d.values = d.values.Push(id)
			copy(d.values.value, value)
		}
	}
}

func gui_pkg_callback(mfr *McpFrame, msg *McpMesg, ver McpVer, context interface{}) {
	id := mcp_mesg_arg_getline(msg, "id", 0)
	if dlogid := mcp_mesg_arg_getline(msg, "dlogid", 0); dlogid == "" {
		show_mcp_error(mfr, msg.mesgname, "Missing dialog ID.")
	} else {
		if dat := gui_dlog_find(dlogid); dat == nil {
			show_mcp_error(mfr, msg.mesgname, "Invalid dialog ID.")
		} else {
			switch msg.mesgname {
			case "ctrl-value":
				valcount := mcp_mesg_arg_linecount(msg, "value")
				if id == "" {
					show_mcp_error(mfr, msg.mesgname, "Missing control ID.");
				} else {
					values := make([]string, valcount)
					for i, _ := range {
						values[i] = mcp_mesg_arg_getline(msg, "value", i)
					}
					gui_value_set_local(dlogid, id, values)
				}
			case "ctrl-event":
				if id == "" {
					show_mcp_error(mfr, msg.mesgname, "Missing control ID.")
				} else {
					evt := mcp_mesg_arg_getline(msg, "event", 0)
					if evt == "" {
						evt = "buttonpress"
					}
					did_dismiss := true
					switch dismissed := mcp_mesg_arg_getline(msg, "dismissed", 0); dismissed {
					case "false", "0":
						did_dismiss = false
					}
					if did_dismiss {
						dat.dismissed = true
					}
					if dat.callback {
						dat.callback(dat.descr, dlogid, id, evt, msg, did_dismiss, dat.context)
					}
				}
			case "error":
				if dat.error_cb {
					dat.error_cb(
						dat.descr,
						dlogid,
						mcp_mesg_arg_getline(msg, "id", 0),
						mcp_mesg_arg_getline(msg, "errcode", 0),
						mcp_mesg_arg_getline(msg, "errtext", 0),
						dat.context,
					)
				}
			}
		}
	}
}

func gui_dlog_alloc(int descr, Gui_CB callback, GuiErr_CB error_cb, void *context) string {
	for {
		tmpid = fmt.Sprintf("%08lX", RANDOM())
		if !gui_dlog_find(tmpid) {
			break
		}
	}
	ptr := &DlogData{
		id: tmpid,
		descr: descr,
		callback: callback,
		error_cb: error_cb,
		context: context,
		prev: &dialog_list,
		next: dialog_list,
	}
	if dialog_list != nil {
		dialog_list.prev = ptr.next
	}
	dialog_list = ptr
	dialog_last_accessed = ptr
	return ptr.id
}

func GuiFree(id string) (r int) {
	if ptr := gui_dlog_find(id); ptr == nil {
		r = EGUINODLOG
	} else {
		*(ptr.prev) = ptr.next
		if ptr.next {
			ptr.next.prev = ptr.prev
		}
		if dialog_last_accessed == ptr {
			dialog_last_accessed = nil
		}
	}
	return
}

func gui_dlog_closeall_descr(descr int) bool {
	for ptr := dialog_list; ptr != nil; ptr = ptr.next {
		for ; ptr != nil && ptr.descr != descr; ptr = ptr.next {}
		if ptr != nil && ptr.callback != nil {
			msg := &McpMesg{ package: GUI_PACKAGE, mesgname: "ctrl-event" }
			mcp_mesg_arg_append(msg, "dlogid", ptr.id)
			mcp_mesg_arg_append(msg, "id", "_closed")
			mcp_mesg_arg_append(msg, "dismissed", "1")
			mcp_mesg_arg_append(msg, "event", "buttonpress")
			ptr->callback(ptr.descr, ptr.id, "_closed", "buttonpress", msg, 1, ptr.context)
		}
	}
	return
}

func GuiVersion(int descr) (r McpVer) {
	if mfr := descr_mcpframe(descr); mfr != nil {
		r = mcp_frame_package_supported(mfr, GUI_PACKAGE)
	}
	return
}

func GuiSupported(descr int) (r bool) {
	if mfr := descr_mcpframe(descr); mfr != nil {
		supp := mcp_frame_package_supported(mfr, GUI_PACKAGE)
		r = supp.minor != 0 || supp.major != 0
	}
	return
}

func GuiSimple(descr int, title string, callback Gui_CB, error_cb GuiErr_CB, context interface{}) (r string) {
	if mfr := descr_mcpframe(descr); mfr != nil {
		if GuiSupported(descr) {
			id := gui_dlog_alloc(descr, callback, error_cb, context)
			msg := &McpMesg{ package: GUI_PACKAGE, mesgname: "dlog-create" }
			mcp_mesg_arg_append(msg, "dlogid", id)
			mcp_mesg_arg_append(msg, "type", "simple")
			mcp_mesg_arg_append(msg, "title", title)
			mcp_frame_output_mesg(mfr, msg)
			r = id
		}
	}
	return
}

func GuiTabbed(descr int, title string, pagecount int, pagenames, pageids []string, callback Gui_CB, error_cb GuiErr_CB, context interface{}) (r string) {
	if mfr := descr_mcpframe(descr); mfr != nil {
		if GuiSupported(descr) {
			id := gui_dlog_alloc(descr, callback, error_cb, context)
			msg := &McpMesg{ package: GUI_PACKAGE, mesgname: "dlog-create" }
			mcp_mesg_arg_append(msg, "dlogid", id)
			mcp_mesg_arg_append(msg, "type", "tabbed")
			mcp_mesg_arg_append(msg, "title", title)
			for i := 0; i < pagecount; i++ {
				mcp_mesg_arg_append(msg, "panes", pageids[i])
				mcp_mesg_arg_append(msg, "names", pagenames[i])
			}
			mcp_frame_output_mesg(mfr, msg)
			r = id
		}
	}
	return
}

func GuiHelper(descr int, title string, pagecount int, pagenames, pageids []string, callback Gui_CB, error_cb GuiErr_CB, context interface{}) (r string) {
	if mfr := descr_mcpframe(descr); mfr != nil {
		if GuiSupported(descr) {
			id := gui_dlog_alloc(descr, callback, error_cb, context)
			msg := &McpMesg{ package: GUI_PACKAGE, mesgname: "dlog-create" }
			mcp_mesg_arg_append(msg, "dlogid", id)
			mcp_mesg_arg_append(msg, "type", "helper")
			mcp_mesg_arg_append(msg, "title", title)
			for i := 0; i < pagecount; i++ {
				mcp_mesg_arg_append(msg, "panes", pageids[i])
				mcp_mesg_arg_append(msg, "names", pagenames[i])
			}
			mcp_frame_output_mesg(mfr, msg)
			r = id
		}
	}
	return
}


int
GuiShow(const char *id)
{
	McpMesg msg;
	McpFrame *mfr;
	int descr = gui_dlog_get_descr(id);

	mfr = descr_mcpframe(descr);
	if (!mfr) {
		return EGUINODLOG;
	}
	if (GuiSupported(descr)) {
		if (!GuiClosed(id)) {
			msg = &McpMesg{ package: GUI_PACKAGE, mesgname: "dlog-show" }
			mcp_mesg_arg_append(&msg, "dlogid", id);
			mcp_frame_output_mesg(mfr, &msg);
		}
		return 0;
	}
	return EGUINOSUPPORT;
}



int
GuiClose(const char *id)
{
	McpMesg msg;
	McpFrame *mfr;
	int descr = gui_dlog_get_descr(id);

	mfr = descr_mcpframe(descr);
	if (!mfr) {
		return EGUINODLOG;
	}
	if (GuiSupported(descr)) {
		if (!GuiClosed(id)) {
			msg = &McpMesg{ package: GUI_PACKAGE, mesgname: "dlog-close" }
			mcp_mesg_arg_append(&msg, "dlogid", id);
			mcp_frame_output_mesg(mfr, &msg);
		}
		return 0;
	}
	return EGUINOSUPPORT;
}



int
GuiSetVal(const char *dlogid, const char *id, int lines, const char **value)
{
	McpMesg msg;
	McpFrame *mfr;
	int i;
	int descr = gui_dlog_get_descr(dlogid);

	mfr = descr_mcpframe(descr);
	if (!mfr) {
		return EGUINODLOG;
	}
	if (GuiSupported(descr)) {
		msg = &McpMesg{ package: GUI_PACKAGE, mesgname: "ctrl-value" }
		mcp_mesg_arg_append(&msg, "dlogid", dlogid);
		mcp_mesg_arg_append(&msg, "id", id);
		for (i = 0; i < lines; i++) {
			mcp_mesg_arg_append(&msg, "value", value[i]);
		}
		mcp_frame_output_mesg(mfr, &msg);
		gui_value_set_local(dlogid, id, value)
		return 0;
	}
	return EGUINOSUPPORT;
}



int
GuiListInsert(const char *dlogid, const char *id, int after, int lines, const char **value)
{
	McpMesg msg;
	McpFrame *mfr;
	char numbuf[32];
	int i;
	int descr = gui_dlog_get_descr(dlogid);

	mfr = descr_mcpframe(descr);
	if (!mfr) {
		return EGUINODLOG;
	}
	if (GuiSupported(descr)) {
		msg = &McpMesg{ package: GUI_PACKAGE, mesgname: cmdname "list-insert" }
		mcp_mesg_arg_append(&msg, "dlogid", dlogid);
		mcp_mesg_arg_append(&msg, "id", id);
		if after > 0 {
			mcp_mesg_arg_append(&msg, "after", fmt.Sprint(after))
		}
		for (i = 0; i < lines; i++) {
			mcp_mesg_arg_append(&msg, "values", value[i]);
		}
		mcp_frame_output_mesg(mfr, &msg);
		return 0;
	}
	return EGUINOSUPPORT;
}



int
GuiListDel(const char *dlogid, const char *id, int from, int to)
{
	descr := gui_dlog_get_descr(dlogid)
	mfr := descr_mcpframe(descr)
	if !mfr {
		return EGUINODLOG
	}
	if GuiSupported(descr) {
		msg := &McpMesg{ package: GUI_PACKAGE, mesgname: "list-delete" }
		mcp_mesg_arg_append(&msg, "dlogid", dlogid)
		mcp_mesg_arg_append(&msg, "id", id)
		if from == GUI_LIST_END {
			mcp_mesg_arg_append(&msg, "from", "end")
		} else {
			mcp_mesg_arg_append(&msg, "from", fmt.Sprint(from))
		}
		if to == GUI_LIST_END {
			mcp_mesg_arg_append(&msg, "to", "end")
		} else {
			mcp_mesg_arg_append(&msg, "to", fmt.Sprint(to))
		}
		mcp_frame_output_mesg(mfr, &msg)
		return 0
	}
	return EGUINOSUPPORT
}



int
GuiMenuItem(const char *dlogid, const char *id, const char *type, const char *name, const char **args)
{
	McpMesg msg;
	McpFrame *mfr;
	int i;
	int descr = gui_dlog_get_descr(dlogid);

	mfr = descr_mcpframe(descr);
	if (!mfr) {
		return EGUINODLOG;
	}
	if (GuiSupported(descr)) {
		msg = &McpMesg{ package: GUI_PACKAGE, mesgname: "menu-item" }
		mcp_mesg_arg_append(&msg, "dlogid", dlogid);
		mcp_mesg_arg_append(&msg, "id", id);
		mcp_mesg_arg_append(&msg, "name", name);
		mcp_mesg_arg_append(&msg, "type", name);
		i = 0;
		while (args && args[i]) {
			const char *arg = args[i];
			const char *val = args[i + 1];

			mcp_mesg_arg_append(&msg, arg, val);
			i += 2;
		}
		mcp_frame_output_mesg(mfr, &msg);
		return 0;
	}
	return EGUINOSUPPORT;
}



int
GuiMenuCmd(const char *dlogid, const char *id, const char *name)
{
	return GuiMenuItem(dlogid, id, "command", name, NULL);
}



int
GuiMenuCheckBtn(const char *dlogid, const char *id, const char *name, const char **args)
{
	return GuiMenuItem(dlogid, id, "checkbutton", name, args);
}



void
gui_ctrl_process_layout(McpMesg * msg, int layout)
{
	char buf[32];

	buf[0] = '\0';
	if ((layout & GUI_N))
		strcatn(buf, sizeof(buf), "n");

	if ((layout & GUI_S))
		strcatn(buf, sizeof(buf), "s");

	if ((layout & GUI_E))
		strcatn(buf, sizeof(buf), "e");

	if ((layout & GUI_W))
		strcatn(buf, sizeof(buf), "w");

	if buf != "" {
		mcp_mesg_arg_append(msg, "sticky", buf)
	}
	if buf = fmt.Sprint(GET_COLSKIP(layout)); buf != "0" {
		mcp_mesg_arg_append(msg, "colskip", buf)
	}
	if buf = fmt.Sprint(GET_COLSPAN(layout)); buf != "1" {
		mcp_mesg_arg_append(msg, "colspan", buf)
	}
	if buf = fmt.Sprint(GET_ROWSPAN(layout)); buf != "1" {
		mcp_mesg_arg_append(msg, "rowspan", buf)
	}
	if buf = fmt.Sprint(GET_LEFTPAD(layout)); buf != "0" {
		mcp_mesg_arg_append(msg, "leftpad", buf)
	}
	if buf = fmt.Sprint(GET_TOPPAD(layout)); buf != "0" {
		mcp_mesg_arg_append(msg, "toppad", buf)
	}


	if ((layout & GUI_NONL))
		mcp_mesg_arg_append(msg, "newline", "0");

	if ((layout & GUI_HEXP))
		mcp_mesg_arg_append(msg, "hweight", "1");

	if ((layout & GUI_VEXP))
		mcp_mesg_arg_append(msg, "vweight", "1");

	if ((layout & GUI_REPORT))
		mcp_mesg_arg_append(msg, "report", "1");

	if ((layout & GUI_REQUIRED))
		mcp_mesg_arg_append(msg, "required", "1");
}

func gui_ctrl_make_v(dlogid, type, pane, id, text, value string, layout int, args []string) (r int) {
	descr := gui_dlog_get_descr(dlogid)
	switch mfr := descr_mcpframe(descr); {
	case mfr == nil:
		r = EGUINODLOG
	case GuiSupported(descr):
		msg := &McpMesg{ package: GUI_PACKAGE, mesgname: fmt.Sprintf("ctrl-%.55s", type) }
		gui_ctrl_process_layout(msg, layout)
		mcp_mesg_arg_append(msg, "dlogid", dlogid)
		mcp_mesg_arg_append(msg, "id", id)
		mcp_mesg_arg_append(msg, "text", text)
		mcp_mesg_arg_append(msg, "value", value)
		gui_value_set_local(dlogid, id, &value)
		mcp_mesg_arg_append(msg, "pane", pane)
		for i := 0; len(args) > 0 && args[i] != nil; i += 2 {
			arg := args[i]
			val := args[i + 1]

			if arg == "value" {
				gui_value_set_local(dlogid, id, &val)
			}
			mcp_mesg_arg_append(msg, arg, val)
			i += 2
		}
		mcp_frame_output_mesg(mfr, msg)
		r = 0
	default:
		r = EGUINOSUPPORT
	}
	return
}

func gui_ctrl_make_l(const char *dlogid, const char *type, const char *pane, const char *id, const char *text, const char *value, int layout, ...) int {
	va_list ap;
	McpMesg msg;
	McpFrame *mfr;
	int descr;

	va_start(ap, layout);

	descr = gui_dlog_get_descr(dlogid);
	mfr = descr_mcpframe(descr);
	if (!mfr) {
		va_end(ap);
		return EGUINODLOG;
	}
	if (GuiSupported(descr)) {
		cmdname := fmt.Sprintf("ctrl-%.55s", type)
		msg = &McpMesg{ package: GUI_PACKAGE, mesgname: cmdname }
		gui_ctrl_process_layout(&msg, layout);
		mcp_mesg_arg_append(&msg, "dlogid", dlogid);
		mcp_mesg_arg_append(&msg, "id", id);
		if (text)
			mcp_mesg_arg_append(&msg, "text", text);
		if (value) {
			mcp_mesg_arg_append(&msg, "value", value);
			if (id) {
				gui_value_set_local(dlogid, id, &value);
			}
		}
		if (pane)
			mcp_mesg_arg_append(&msg, "pane", pane);
		while (1) {
			const char *val;
			const char *arg;

			arg = va_arg(ap, const char *);

			if (!arg)
				break;
			val = va_arg(ap, const char *);

			mcp_mesg_arg_append(&msg, arg, val);
		}
		mcp_frame_output_mesg(mfr, &msg);
		va_end(ap);
		return 0;
	}
	va_end(ap);
	return EGUINOSUPPORT;
}

func GuiEdit(const char *dlogid, const char *pane, const char *id, const char *text, const char *value, int width, int layout) int {
	buf := fmt.Sprint(width)
	return gui_ctrl_make_l(dlogid, "edit", pane, id, text, value, layout, "width", buf, NULL);
}

func GuiText(const char *dlogid, const char *pane, const char *id, const char *value, int width, int layout) int {
	widthbuf := fmt.Sprint(width)
	return gui_ctrl_make_l(dlogid, "text", pane, id, nil, value, layout, "width", widthbuf, nil)
}

func GuiSpinner(const char *dlogid, const char *pane, const char *id, const char *text, int value, int width, int min, int max, int layout) int {
	widthbuf := fmt.Sprint(width)
	valbuf := fmt.Sprint(value)
	minbuf := fmt.Sprint(min)
	maxbuf := fmt.Sprint(max)
	return gui_ctrl_make_l(dlogid, "spinner", pane, id, text, valbuf, layout, "width", widthbuf, "min", minbuf, "max", maxbuf, nil)
}

func GuiCombo(const char *dlogid, const char *pane, const char *id, const char *text, const char *value, int width, int editable, int layout) int {
	buf := fmt.Sprintf(buf, sizeof(buf), "%d", width);
	return gui_ctrl_make_l(dlogid, "combobox", pane, id, text, value, layout, "width", buf, "editable", editable ? "1" : "0", nil)
}

func GuiMulti(const char *dlogid, const char *pane, const char *id, const char *value, int width, int height, int fixed, int layout) int {
	widthbuf := fmt.Sprint(width)
	heightbuf := fmt.Sprint(height)
	return gui_ctrl_make_l(dlogid, "multiedit", pane, id, nil, value, layout, "width", widthbuf, "height", heightbuf, "font", (fixed ? "fixed" : "variable"), nil)
}

func GuiHRule(const char *dlogid, const char *pane, const char *id, int height, int layout) int {
	heightbuf := fmt.Sprint(height)
	return gui_ctrl_make_l(dlogid, "hrule", pane, id, nil, nil, layout, "height", heightbuf, nil)
}

func GuiVRule(const char *dlogid, const char *pane, const char *id, int thickness, int layout) int {
	widthbuf := fmt.Sprintf(widthbuf, sizeof(widthbuf), "%d", thickness)
	return gui_ctrl_make_l(dlogid, "vrule", pane, id, nil, nil, layout, "width", widthbuf, nil)
}

func GuiFrame(const char *dlogid, const char *pane, const char *id, int layout) int {
	return gui_ctrl_make_l(dlogid, "frame", pane, id, nil, nil, layout, nil)
}

func GuiGroupBox(const char *dlogid, const char *pane, const char *id, const char *text, int collapsible, int collapsed, int layout) int {
	return gui_ctrl_make_l(dlogid, "frame", pane, id, text, nil, layout, "visible", "1", "collapsible", collapsible ? "1" : "0", "collapsed", collapsed ? "1" : "0", nil)
}

func GuiButton(const char *dlogid, const char *pane, const char *id, const char *text, int width, int dismiss, int layout) int {
	widthbuf := fmt.Sprint(width)
	return gui_ctrl_make_l(dlogid, "button", pane, id, text, nil, layout, "width", widthbuf, "dismiss", dismiss ? "1" : "0", nil)
}


/******************************************************************************************/
/* MUF gui functions
 **********************/

func muf_dlog_add(fr *frame, dlogid string) {
	fr.dlogids = &dlogidlist{ dlogid: dlogid, next: fr.dlogids }
}

func muf_dlog_remove(fr *frame, dlogid string) {
	for prev := &fr.dlogids; *prev != nil; {
		if (*prev).dlogid == dlogid {
			*prev = (*prev).next
		} else {
			prev = &((*prev).next)
		}
	}
	GuiFree(dlogid)
}

func muf_dlog_purge(fr *frame) {
	for fr.dlogids != nil {
		GuiClose(fr.dlogids.dlogid)
		GuiFree(fr.dlogids.dlogid)
		fr.dlogids = fr.dlogids.next
	}
}

/***************************************************************************
 ***************************************************************************
 ***************************************************************************/

func post_dlog_cb(descr int, dlogid, id, event string, msg *McpMesg, did_dismiss bool, context interface{}) {
	switch id {
	case "post":
		subject := gui_value_get(dlogid, "subj", 0)
		keywords := gui_value_get(dlogid, "keywd", 0)
		pnotify(pdescrcon(descr), fmt.Sprintf("Subject: %s", subject))
		pnotify(pdescrcon(descr), fmt.Sprintf("Keywords: %s", keywords))
	case "cancel":
		pnotify(pdescrcon(descr), "Posting cancelled.")
	default:
		pnotify(pdescrcon(descr), "Invalid event!")
	}
	if did_dismiss {
		GuiFree(dlogid)
	}
}

void
do_post_dlog(int descr, const char *text)
{
	const char *keywords[] = { "Misc.", "Wedding", "Party", "Toading", "New MUCK" };
	const char *dlg = GuiSimple(descr, "A demonstration dialog", post_dlog_cb, NULL, NULL);

	GuiEdit(dlg, NULL, "subj", "Subject", text, 60, GUI_EW);
	GuiCombo(dlg, NULL, "keywd", "Keywords", "Misc.", 60, 1, GUI_EW | GUI_HEXP);
	GuiListInsert(dlg, "keywd", GUI_LIST_END, 5, keywords);

	GuiMulti(dlg, NULL, "body", NULL, 80, 12, 1, GUI_NSEW | GUI_VEXP | COLSPAN(2));
	GuiHRule(dlg, NULL, NULL, 2, COLSPAN(2));
	GuiFrame(dlg, NULL, "bfr", GUI_EW | COLSPAN(2) | TOPPAD(0));
	GuiFrame(dlg, "bfr", NULL, GUI_EW | GUI_HEXP | GUI_NONL);
	GuiVRule(dlg, NULL, NULL, 2, GUI_NONL);
	GuiButton(dlg, NULL, "post", "Post", 8, 1, GUI_E | GUI_NONL);
	GuiButton(dlg, NULL, "cancel", "Cancel", 8, 1, GUI_E | TOPPAD(0));

	GuiShow(dlg);
}
static const char *mcpgui_c_version = "$RCSfile: mcpgui.c,v $ $Revision: 1.28 $";
const char *get_mcpgui_c_version(void) { return mcpgui_c_version; }
