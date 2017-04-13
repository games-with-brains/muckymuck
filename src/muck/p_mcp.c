package fbmuck

/* This allows using GUI if the player owns the MUF, which drops the
   requirement from an M3. This is not .top because .top will usually 
   be $lib/gui. This could be expanded to check the entire MUF chain. */
func apply_mcp_primitive(p dbref, fr *frame, mlev, top *int, n int, f func(Array)) {
	apply_primitive(n, top, func(op Array) {
		if mlev < tp_mcp_muf_mlev && p != db.Fetch(fr.caller.st[1]).Owner {
			panic("Permission denied!!!")
		}
		f(op)
	})
}

func muf_mcp_callback(mfr *McpFrame, mesg *McpMesg, version McpVer, context interface{}) {
	descr := mfr.descriptor.descriptor
	user := mfr.descriptor.player
	var ptr *mcp_binding
	for ptr = db.Fetch(context.(dbref)).(Program).mcp_binding; ptr != nil && (ptr.pkgname != mesg.pkgname || ptr.msgname != mesg.mesgname); ptr = ptr.next {}
	if ptr != nil {
		var argarr *stk_array
		if tmpfr := interp(descr, user, db.Fetch(user).Location, obj, -1, PREEMPT, STD_REGUID, 0); tmpfr != nil {
			tmpfr.argument.top--
			args := make(Dictionary)
			for arg := mesg.args; arg != nil; arg = arg.next {
				switch {
				case arg.value == nil:
					args[arg.name] = ""
				case arg.value.next == nil:
					args[arg.name] = arg.value.value
				default:
					a := make(Array, 0, mcp_mesg_arg_linecount(mesg, arg.name))
					for p := arg.value; p != nil; p = p.next {
						a = append(a, p.value)
					}
					args[arg.name] = a
				}
			}
			push(tmpfr.argument.st, &(tmpfr.argument.top), descr)
			push(tmpfr.argument.st, &(tmpfr.argument.top), args)
			tmpfr.pc = ptr.addr
			interp_loop(user, obj, tmpfr, false)
		}
	}
}

func muf_mcp_event_callback(mfr *McpFrame, mesg *McpMesg, version McpVer, context interface{}) {
	if destfr := timequeue_pid_frame(context.(int)); destfr != nil {
		args := make(Dictionary)
		for arg := mesg.args; arg != nil; arg = arg.next {
			if arg.value == nil {
				args[name] = nil
			} else {
				var count int
				a := make(Array, mcp_mesg_arg_linecount(mesg, arg.name))
				for p := arg.value; p != nil; p = p.next {
					a[count] = p.value
					count++
				}
				args[arg.name] = a
			}
		}

		c := Dictionary{
			"descr": mfr.descriptor.descriptor,
			"package": mesg.pkgname,
			"message": mesg.msgname,
			"args": args,
		}

		if mesg.msgname != "" {
			muf_event_add(destfr, fmt.Sprintf("MCP.%.128s-%.128s", mesg.pkgname, mesg.msgname), c, 0)
		} else {
			muf_event_add(destfr, fmt.Sprintf("MCP.%.128s", mesg.pkgname), c, 0)
		}
	}
}

func stuff_dict_in_mesg(arr map[string] interface{}, msg *McpMesg) (r int) {
	for argname, argval := range arr {
		switch argval := argval.(type) {
		case Array:
			for subname, subval := range argval {
				mcp_mesg_arg_remove(msg, argname)
				switch v := subval.(type) {
				case string:
					mcp_mesg_arg_append(msg, argname, v)
				case int:
					mcp_mesg_arg_append(msg, argname, fmt.Sprint(v))
				case dbref:
					mcp_mesg_arg_append(msg, argname, fmt.Sprintf("#%d", v))
				case float64:
					mcp_mesg_arg_append(msg, argname, fmt.Sprintf("%.15g", v))
				default:
					r = -3
					break
				}
			}
		case Dictionary:
			for subname, subval := range argval {
				mcp_mesg_arg_remove(msg, argname)
				switch v := subval.(type) {
				case string:
					mcp_mesg_arg_append(msg, argname, v)
				case int:
					mcp_mesg_arg_append(msg, argname, fmt.Sprint(v))
				case dbref:
					mcp_mesg_arg_append(msg, argname, fmt.Sprintf("#%d", v))
				case float64:
					mcp_mesg_arg_append(msg, argname, fmt.Sprintf("%.15g", v))
				default:
					r = -3
					break
				}
			}
		case string:
			mcp_mesg_arg_remove(msg, argname)
			mcp_mesg_arg_append(msg, argname, argval)
		case int:
			mcp_mesg_arg_remove(msg, argname)
			mcp_mesg_arg_append(msg, argname, fmt.Sprint(argval))
		case dbref:
			mcp_mesg_arg_remove(msg, argname)
			mcp_mesg_arg_append(msg, argname, fmt.Sprintf("#%d", argval))
		case float64:
			mcp_mesg_arg_remove(msg, argname)
			mcp_mesg_arg_append(msg, argname, fmt.Sprintf("%.15g", argval))
		default:
			r = -4
		}
	}
	return
}

func prim_mcp_register(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 3, func(op Array) {
		pkgname := op[0].(string);
		vermin := McpVer{
			major = int(op[1].(float64))
			minor = int((op[1].(float64) * 1000) % 1000)
		}
		vermax := McpVer{
			major = int(op[2].(float64))
			minor = int((op[2].(float64) * 1000) % 1000)
		}
		mcp_package_register(pkgname, vermin, vermax, muf_mcp_callback, program)
	})
}

func prim_mcp_register_event(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 3, func(op Array) {
		pkgname := op[0].(string)
		vermin := McpVer{
			major : int(op[1].(float64))
			minor : int((op[1].(float64) * 1000) % 1000)
		}
		vermax := McpVer{
			major: int(op[2].(float64))
			minor: int((op[2].(float64) * 1000) % 1000)
		}
		mcp_package_register(pkgname, vermin, vermax, muf_mcp_event_callback, fr.pid)
	})
}

func prim_mcp_supports(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 2, func(op Array) {
		descr := op[0].(int)
		pkgname := op[1].(string)
		if mfr = descr_mcpframe(descr); mfr != nil {
			ver := mcp_frame_package_supported(mfr, pkgname)
			push(arg, top, ver.major + (ver.minor / 1000.0))
		} else {
			push(arg, top, 0.0)
		}
	})
}

func prim_mcp_bind(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 3, func(op Array) {
		pkgname := op[0].(string)
		msgname := op[1].(string)
		address := op[2].(Address)
		switch {
		case program != address.progref:
			panic("Destination address outside current program. (3)")
		case !valid_reference(address.progref), !IsProgram(address.progref):
			panic("Invalid address. (3)")
		}

		p := db.Fetch(program)
		ptr := p.(Program).mcp_binding
		for ; ptr != nil && (ptr.pkgname == pkgname || ptr.msgname == msgname); ptr = ptr.next {}
		if ptr == nil {
			p.(Program).mcp_binding = &mcp_binding{ pkgname: pkgname, msgname: msgname, next: p.(Program).mcp_binding }
		}
		ptr.addr = address.addr.data
	})
}

func prim_mcp_send(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 4, func(op Array) {
		pkgname := op[0].(string)
		msgname := op[1].(string)
		arr := op[2].(stk_array)
		descr := op[3].(int)
		if mfr := descr_mcpframe(descr); mfr != nil {
			ver := mcp_frame_package_supported(mfr, pkgname)
			if ver.minor == 0 && ver.major == 0 {
				panic("MCP package not supported by that descriptor.")
			}
			msg := &McpMesg{ package: pkgname, mesgname: msgname }
			switch stuff_dict_in_mesg(arr, msg) {
			case -1:
				panic("Args dictionary can only have string keys. (4)")
			case -2:
				panic("Args dictionary cannot have a null string key. (4)")
			case -3:
				panic("Unsupported value type in list value. (4)")
			case -4:
				panic("Unsupported value type in args dictionary. (4)")
			}
			mcp_frame_output_mesg(mfr, msg)
		}
	})
}

func fbgui_muf_event_cb(descr int, dlogid, id, event string, msg *McpMesg, did_dismiss bool, context interface{}) {
	values := make(Dictionary)
	for name := GuiValueFirst(dlogid); name != ""; name = GuiValueNext(dlogid, name) {
		lines := make(Array, gui_value_linecount(dlogid, name))
		for i, _ := range lines {
			lines[i] = gui_value_get(dlogid, name, i)
		}
		values[name] = lines
	}

	args := make(Array, mcp_mesg_arg_linecount(msg, "data"))
	for i, _ := range args {
		args[i] = mcp_mesg_arg_getline(msg, "data", i)
	}
	muf_event_add(context.(*frame), fmt.Sprintf("GUI.%s", dlogid), Dictionary{
		"dismissed": did_dismiss,
		"descr": descr,
		"dlogid": dlogid,
		"id": id,
		"event": event,
		"values": values,
		"data": args,
	}, 0)
}

func fbgui_muf_error_cb(descr int, dlogid, id, errcode, errtext string, context interface{}) {
	muf_event_add(context.(*frame), fmt.Sprintf("GUI.%s", dlogid), Dictionary{
		"descr": descr,
		"dlogid": dlogid,
		"id": id,
		"errcode": errcode,
		"errtext": errtext,
	}, 0)
}

func prim_gui_available(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 1, func(op Array) {
		ver := GuiVersion(op[0].(int))
		push(arg, top, ver.major + (ver.minor / 1000.0))
	})
}

func prim_gui_dlog_create(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 4, func(op Array) {
		mfr := descr_mcpframe(op[0].(int))
		wintype := op[1].(string)
		if wintype == ""{
			wintype = "simple"
		}
		title := op[2].(string)
		arr := op[3].(stk_array)
		switch {
		case mfr == nil:
			panic("Invalid descriptor number. (1)")
		case !GuiSupported(mfr):
			panic("The MCP GUI package is not supported for this connection.")
		}
		dlogid := gui_dlog_alloc(mfr, fbgui_muf_event_cb, fbgui_muf_error_cb, fr)
		msg := &McpMesg{ package: GUI_PACKAGE, mesgname: "dlog-create" }
		mcp_mesg_arg_append(msg, "title", title)
		switch stuff_dict_in_mesg(arr, msg) {
		case -1:
			panic("Args dictionary can only have string keys. (4)")
		case -2:
			panic("Args dictionary cannot have a null string key. (4)")
		case -3:
			panic("Unsupported value type in list value. (4)")
		case -4:
			panic("Unsupported value type in args dictionary. (4)")
		}
		mcp_mesg_arg_remove(msg, "type")
		mcp_mesg_arg_append(msg, "type", wintype)
		mcp_mesg_arg_remove(msg, "dlogid")
		mcp_mesg_arg_append(msg, "dlogid", dlogid)
		mcp_frame_output_mesg(mfr, msg)
		muf_dlog_add(fr, dlogid)
		push(arg, top, dlogid)
	})
}

func prim_gui_dlog_show(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 1, func(op Array) {
		dlogid := op[0].(string)
		switch GuiShow(dlogid) {
		case EGUINOSUPPORT:
			panic("GUI not available.  Internal error.")
		case EGUINODLOG:
			panic("Invalid dialog ID.")
		}
	})
}

func prim_gui_dlog_close(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 1, func(op Array) {
		dlogid := op[0].(string)
		switch GuiClose(dlogid) {
		case EGUINOSUPPORT:
			panic("Internal error: GUI not available.")
		case EGUINODLOG:
			panic("Invalid dialog ID.")
		}
		muf_dlog_remove(fr, dlogid)
	})
}

func prim_gui_ctrl_create(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 4, func(op Array) {
		dlogid := op[0].(string)
		ctrltype := op[1].(string)
		ctrlid := op[2].(string)
		arr := op[3].(stk_array)
		descr := gui_dlog_get_descr(dlogid)
		mfr := descr_mcpframe(descr)
		switch {
		case mfr == nil:
			panic("No such dialog currently exists. (1)")
		case !GuiSupported(descr):
			panic("Internal error: The given dialog's descriptor doesn't support the GUI package. (1)")
		}

		msg := &McpMesg{ package: GUI_PACKAGE, mesgname: fmt.Sprintf("ctrl-%.55s", ctrltype) }
		switch result = stuff_dict_in_mesg(arr, msg); result {
		case -1:
			panic("Args dictionary can only have string keys. (4)")
		case -2:
			panic("Args dictionary cannot have a null string key. (4)")
		case -3:
			panic("Unsupported value type in list value. (4)")
		case -4:
			panic("Unsupported value type in args dictionary. (4)")
		}
	
		vallines := mcp_mesg_arg_linecount(msg, "value")
		valname := mcp_mesg_arg_getline(msg, "valname", 0)
		if valname == "" {
			valname = ctrlid
		}
		if valname != "" && vallines > 0 {
			vallist := make([]string, vallines)
			for i := 0; i < vallines; i++ {
				vallist[i] = mcp_mesg_arg_getline(msg, "value", i)
			}
			gui_value_set_local(dlogid, valname, vallist)
		}
		mcp_mesg_arg_remove(msg, "dlogid")
		mcp_mesg_arg_append(msg, "dlogid", dlogid)
		mcp_mesg_arg_remove(msg, "id")
		mcp_mesg_arg_append(msg, "id", ctrlid)
		mcp_frame_output_mesg(mfr, msg)
	})
}

func prim_gui_ctrl_command(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 4, func(op Array) {
		dlogid := op[0].(string)
		ctrlid := op[1].(string)
		ctrlcmd := op[2].(string)
		arr := op[3].(stk_array)
		descr := gui_dlog_get_descr(dlogid)
		mfr := descr_mcpframe(descr)
		switch {
		case mfr == nil:
			panic("No such dialog currently exists. (1)")
		case !GuiSupported(descr):
			panic("Internal error: The given dialog's descriptor doesn't support the GUI package. (1)")
		}

		msg := &McpMesg{ package: GUI_PACKAGE, mesgname: "ctrl-command" }
		switch result = stuff_dict_in_mesg(arr, msg); result {
		case -1:
			panic("Args dictionary can only have string keys. (4)")
		case -2:
			panic("Args dictionary cannot have a null string key. (4)")
		case -3:
			panic("Unsupported value type in list value. (4)")
		case -4:
			panic("Unsupported value type in args dictionary. (4)")
		}
		mcp_mesg_arg_remove(msg, "dlogid")
		mcp_mesg_arg_append(msg, "dlogid", dlogid)
		mcp_mesg_arg_remove(msg, "id")
		mcp_mesg_arg_append(msg, "id", ctrlid)
		mcp_mesg_arg_remove(msg, "command")
		mcp_mesg_arg_append(msg, "command", ctrlcmd)
		mcp_frame_output_mesg(mfr, msg)
	})
}

func prim_gui_value_set(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 3, func(op Array) {
		dlogid := op[0].(string)
		ctrlid := op[1].(string)
		var valarray []string
		switch value := op[2].(type) {
		case string:
			valarray = []string{ value }
		case Array:
			valarray = make([]string, len(value))
			for i, v := range value {
				switch v := v.(type) {
				case string:
					value = v
				case int:
					value = fmt.Sprint(v)
				case dbref:
					value = fmt.Sprintf("#%d", v)
				case float64:
					value = fmt.Sprintf("%.15g", v)
				default:
					panic("Unsupported value type in list value. (3)")
				}
				valarray[i] = value
			}
		default:
			panic("String or string list control value expected. (3)")
		}
		GuiSetVal(dlogid, ctrlid, len(valarray), valarray)
	})
}

func prim_gui_values_get(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 1, func(op Array) {
		dlogid := op[0].(string)
		nu := make(Dictionary)
		for name := GuiValueFirst(dlogid); name != ""; name = GuiValueNext(dlogid, name) {
			lines := make(Array, gui_value_linecount(dlogid, name))
			for i, _ := range lines {
				lines[i] = gui_value_get(dlogid, name, i)
			}
			nu[name] = lines
		}
		push(arg, top, nu)
	})
}

func prim_gui_value_get(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_mcp_primitive(player, mlev, fr, top, 2, func(op Array) {
		dlogid := op[0].(string)
		ctrlid := op[1].(string)
		nu := make(stk_array, gui_value_linecount(dlogid, ctrlid))
		for i, _ := range nu {
			nu[i] = &inst{ data: gui_value_get(dlogid, ctrlid, i) }
		}
		push(arg, top, nu)
	})
}