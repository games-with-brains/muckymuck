#include "config.h"
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "interface.h"
#include "mcp.h"
#include "mcpgui.h"
#include "externs.h"
#ifdef HAVE_MALLOC_H
#include <malloc.h>
#endif /* HAVE_MALLOC_H */

typedef struct DlogValue_t {
	struct DlogValue_t *next;
	char *name;
	int lines;
	char **value;
} DlogValue;

typedef struct DlogData_t {
	struct DlogData_t *next;
	struct DlogData_t **prev;
	char *id;
	int descr;
	int dismissed;
	DlogValue *values;
	Gui_CB callback;
	GuiErr_CB error_cb;
	void *context;
} DlogData;

DlogData *dialog_list = NULL;
DlogData *dialog_last_accessed = NULL;


void gui_pkg_callback(McpFrame * mfr, McpMesg * msg, McpVer ver, void *context);


void
gui_initialize(void)
{
	McpVer minver = { 1, 0 };  /* { major, minor } */
	McpVer maxver = { 1, 3 };  /* { major, minor } */

	mcp_package_register(GUI_PACKAGE, minver, maxver, gui_pkg_callback, NULL, NULL);
}


func gui_dlog_find(const char *dlogid) (r *DlogData) {
	r = dialog_last_accessed
	if r == nil || r.id == dlogid {
		for r = dialog_list; r != nil; r = r.next {
			if r.id == dlogid {
				dialog_last_accessed = r
				break
			}
		}
	}
	return
}


void*
gui_dlog_get_context(const char *dlogid)
{
	DlogData *ptr = gui_dlog_find(dlogid);

	if (ptr) {
		return ptr->context;
	} else {
		return NULL;
	}
}


int
gui_dlog_get_descr(const char *dlogid)
{
	DlogData *ptr = gui_dlog_find(dlogid);

	if (ptr) {
		return ptr->descr;
	} else {
		return EGUINODLOG;
	}
}


int
GuiClosed(const char *dlogid)
{
	DlogData *ptr = gui_dlog_find(dlogid);

	if (ptr) {
		return ptr->dismissed;
	} else {
		return EGUINODLOG;
	}
}


func gui_value_linecount(const char *dlogid, const char *id) int {
	if ddata := gui_dlog_find(dlogid); ddata == nil {
		r = EGUINODLOG
	} else {
		var ptr *DlogValue
		for ptr = ddata.values; ptr != nil && ptr.name != id; ptr = ptr->next {}
		if ptr != nil {
			r = ptr.lines
		}
	}
	return
}



const char *
GuiValueFirst(const char *dlogid)
{
	DlogData *ddata = gui_dlog_find(dlogid);

	if (!ddata || !ddata->values) {
		return NULL;
	}
	return ddata->values->name;
}

func GuiValueNext(dlogid, id string) (r string) {
	if ddata := gui_dlog_find(dlogid); ddata != nil {
		var ptr *DlogValue
		for ptr = ddata->values; ptr != nil && ptr.name != id; ptr = ptr->next {}
		if ptr != nil && ptr.next != nil {
			r = ptr.next.name
		}
	}
	return
}

func gui_value_get(dlogid, id string, line int) (r string) {
	if ddata := gui_dlog_find(dlogid); ddata != nil {
		var ptr *DlogValue
		for ptr = ddata->values; ptr != nil && ptr.name != id; ptr = ptr->next {}
		if ptr != nil && line > -1 && line < ptr.lines {
			r = ptr.value[line]
		}
	}
	return
}

void
gui_value_set_local(const char *dlogid, const char *id, int lines, const char **value)
{
	DlogValue *ptr;
	DlogData *ddata = gui_dlog_find(dlogid);
	int i;
	int limit = 256;

	if (!ddata) {
		return;
	}
	for ptr = ddata->values; ptr != nil && ptr.name != id; {
		ptr = ptr->next;
		if (!limit--) {
			return;
		}
	}
	if (ptr) {
		for (i = 0; i < ptr->lines; i++) {
			free(ptr->value[i]);
		}
		free(ptr->value);
	} else {
		int ilen = len(id)+1;
		ptr = (DlogValue *) malloc(sizeof(DlogValue));
		ptr->name = (char *) malloc(ilen);
		strcpyn(ptr->name, ilen, id);
		ptr->next = ddata->values;
		ddata->values = ptr;
	}
	ptr->lines = lines;
	ptr->value = (char **) malloc(sizeof(char *) * lines);

	for (i = 0; i < lines; i++) {
		int vlen = len(value[i])+1;
		ptr->value[i] = (char *) malloc(vlen);
		strcpyn(ptr->value[i], vlen, value[i]);
	}
}

func gui_pkg_callback(McpFrame * mfr, McpMesg * msg, McpVer ver, void *context) {
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
					show_mcp_error(mfr, msg->mesgname, "Missing control ID.");
				} else {
					values := make([]string, valcount)
					for i, _ := range {
						values[i] = mcp_mesg_arg_getline(msg, "value", i)
					}
					gui_value_set_local(dlogid, id, valcount, values)
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
				err := mcp_mesg_arg_getline(msg, "errcode", 0)
				text := mcp_mesg_arg_getline(msg, "errtext", 0)
				id := mcp_mesg_arg_getline(msg, "id", 0)
				if dat.error_cb {
					dat.error_cb(dat.descr, dlogid, id, err, text, dat.context)
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
		gui_value_set_local(dlogid, id, lines, value);
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
		gui_value_set_local(dlogid, id, 1, &value)
		mcp_mesg_arg_append(msg, "pane", pane)
		for i := 0; len(args) > 0 && args[i] != nil; i += 2 {
			arg := args[i]
			val := args[i + 1]

			if arg == "value" {
				gui_value_set_local(dlogid, id, 1, &val)
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
				gui_value_set_local(dlogid, id, 1, &value);
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
