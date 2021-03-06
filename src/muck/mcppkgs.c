package fbmuck

func show_mcp_error(McpFrame * mfr, char *topic, char *text) {
	McpMesg msg;
	McpVer supp = mcp_frame_package_supported(mfr, "org-fuzzball-notify");

	if (supp.minor != 0 || supp.major != 0) {
		msg = &McpMesg{ package: "org-fuzzball-notify", mesgname: "error" }
		mcp_mesg_arg_append(&msg, "text", text);
		mcp_mesg_arg_append(&msg, "topic", topic);
		mcp_frame_output_mesg(mfr, &msg);
	} else {
		notify(mfr.descriptor.player, text)
	}
}

/*
 * reference is in the format objnum.category.misc where objnum is the
 * object reference, and category can be one of the following:
 *    prop     to set a property named by misc.
 *    proplist to store a string proplist named by misc.
 *    prog     to set the program text of the given object.  Ignores misc.
 *    sysparm  to set an @tune value.  Ignores objnum.
 *    user     to return data to a muf program.
 *
 * If the category is prop, then it accepts the following types:
 *    string        to set the property to a string value.
 *    string-list   to set the property to a multi-line string value.
 *    integer       to set the property to an integer value.
 *
 * Any other values are ignored.
 */
func mcppkg_simpleedit(McpFrame * mfr, McpMesg * msg, McpVer ver, void *context) {
	if msg.mesgname == "set" {
		ObjectID obj = NOTHING;
		char category[BUFFER_LEN];
		char *ptr;
		char buf[BUFFER_LEN];
		char *content;
		int line;

		reference := mcp_mesg_arg_getline(msg, "reference", 0)
		valtype := mcp_mesg_arg_getline(msg, "type", 0)
		lines := mcp_mesg_arg_linecount(msg, "content")
		player := mfr.descriptor.player
		descr := mfr.descriptor.descriptor

		/* extract object number.  -1 for none.  */
		if isdigit(*reference) {
			obj = 0;
			for isdigit(*reference) {
				obj = (10 * obj) + (*reference++ - '0')
				if obj >= 100000000 {
					show_mcp_error(mfr, "simpleedit-set", "Bad reference object.")
					return
				}
			}
		}
		if *reference != '.' {
			show_mcp_error(mfr, "simpleedit-set", "Bad reference value.")
			return
		}
		reference++

		/* extract category string */
		ptr = category
		for *reference && *reference != '.' {
			*ptr++ = *reference++
		}
		*ptr = '\0'
		if *reference != '.' {
			show_mcp_error(mfr, "simpleedit-set", "Bad reference value.")
			return
		}
		reference++

		/* the rest is category specific data. */
		switch category {
		case "prop":
			switch {
			case !obj.IsValid():
				show_mcp_error(mfr, "simpleedit-set", "Bad reference object.")
			case !controls(player, obj):
				show_mcp_error(mfr, "simpleedit-set", "Permission denied.")
			default:
				for ptr = reference; *ptr; ptr++ {
					if *ptr == ':' {
						show_mcp_error(mfr, "simpleedit-set", "Bad property name.")
						return
					}
				}
				if Prop_System(reference) || (!Wizard(player) && (Prop_SeeOnly(reference) || Prop_Hidden(reference))) {
					show_mcp_error(mfr, "simpleedit-set", "Permission denied.")
				} else {
					switch valtype {
					case "string-list", "string":
						int left = BUFFER_LEN - 1
						buf = ""
						for line = 0; line < lines; line++ {
							content = mcp_mesg_arg_getline(msg, "content", line)
							if line > 0 {
								if left > 0 {
									buf += "\r"
									left--
								} else {
									break
								}
							}
							if l := len(content); l > left - 2 {
								buf += content[:left]
								left = 0
								break
							} else {
								buf += content
								left -= l
							}
						}
						add_property(obj, reference, buf, 0)
					case "integer":
						if lines == 1 {
							content = mcp_mesg_arg_getline(msg, "content", 0)
							add_property(obj, reference, nil, strconv.Atoi(content))
						} else {
							show_mcp_error(mfr, "simpleedit-set", "Bad integer value.")
						}
					}
				}
			}
		case "proplist":
			switch {
			case obj.IsValid():
				show_mcp_error(mfr, "simpleedit-set", "Bad reference object.")
			case !controls(player, obj):
				show_mcp_error(mfr, "simpleedit-set", "Permission denied.")
			default:
				for ptr = reference; *ptr; ptr++ {
					if *ptr == ':' {
						show_mcp_error(mfr, "simpleedit-set", "Bad property name.")
						return
					}
				}
				if Prop_System(reference) || (!Wizard(player) && (Prop_SeeOnly(reference) || Prop_Hidden(reference))) {
					show_mcp_error(mfr, "simpleedit-set", "Permission denied.")
				} else {
					switch valtype {
					case "string-list":
						if lines == 0 {
							buf = fmt.Sprintf("%s#", reference)
							remove_property(obj, buf)
						} else {
							buf = fmt.Sprintf("%s#", reference)
							remove_property(obj, buf)
							add_property(obj, buf, "", lines)
							for line = 0; line < lines; line++ {
								content = mcp_mesg_arg_getline(msg, "content", line)
								if !content || !*content {
									content = " "
								}
								buf = fmt.Sprintf("%s#/%d", reference, line + 1)
								add_property(obj, buf, content, 0);
							}
						}
					case "string", "integer":
						show_mcp_error(mfr, "simpleedit-set", "Bad value type for proplist.")
						return
					}
				}
			}
		case "prog":
			if obj, ok := obj.(TYPE_PROGRAM); !ok {
				show_mcp_error(mfr, "simpleedit-set", "Permission denied.")
			} else {
				switch {
				case !obj.IsValid():
					show_mcp_error(mfr, "simpleedit-set", "Bad reference object.")
				case !controls(player, obj):
					show_mcp_error(mfr, "simpleedit-set", "Permission denied.")
				case !Mucker(player):
					show_mcp_error(mfr, "simpleedit-set", "Permission denied.")
				case DB.Fetch(obj).IsFlagged(INTERNAL):
					show_mcp_error(mfr, "simpleedit-set", "Sorry, this program is currently being edited.  Try again later.")
				default:
					tmpline := DB.Fetch(obj).(Program).first
					DB.Fetch(obj).(Program).first = nil

					var curr *line
					for line := 0; line < lines; line++ {
						new_line := new(line)
						if content := mcp_mesg_arg_getline(msg, "content", line); content == "" {
							new_line.this_line = " "
						} else {
							new_line.this_line = content
						}
						if line == 0 {
							DB.Fetch(obj).(Program).first = new_line
						} else {
							curr.next = new_line
						}
						curr = new_line
					}
					log_status("PROGRAM SAVED: %s by %s(%d)", unparse_object(player, obj), DB.Fetch(player).name, player)
					write_program(DB.Fetch(obj).(Program).first, obj)
					if tp_log_programs {
						log_program_text(DB.Fetch(obj).(Program).first, player, obj)
					}
					do_compile(descr, player, obj, 1)
					DB.Fetch(obj).(Program).first = tmpline
					DB.Fetch(player).Touch()
					DB.Fetch(obj).Touch()
				}
			}
		case "sysparm":
			switch {
			case !Wizard(player):
				show_mcp_error(mfr, "simpleedit-set", "Permission denied.")
			case lines != 1:
				show_mcp_error(mfr, "simpleedit-set", "Bad @tune value.")
			default:
				content = mcp_mesg_arg_getline(msg, "content", 0)
				if player == GOD {
					MLEV_GOD
					Tuneables.SetAs(MLEV_GOD, reference, content)
				} else {
					Tuneables.SetAs(MLEV_WIZARD, reference, content)
				}
			}
		case "user":
		default:
			show_mcp_error(mfr, "simpleedit-set", "Unknown reference category.")
		}
	}
}