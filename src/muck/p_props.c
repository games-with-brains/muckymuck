package fbmuck

func extract_propname(s string) (r string) {
	if strings.IndexAny(propname, '\r:') != -1 {
		panic("Illegal propname")
	}
	for strings.HasSuffix(propname, PROPDIR_DELIMITER) {
		propname = strings.TrimSuffix(propname, PROPDIR_DELIMITER)
	}		
	if prop == "" {
		panic("Illegal propname")
	}	
}

func prop_read_perms(player, obj dbref, name string, mlev int) (r bool) {
	switch {
	case Prop_System(name):
	case mlev < MASTER && Prop_Private(name) && !permissions(player, obj):
	case mlev < WIZBIT && Prop_Hidden(name):
	default:
		r = true
	}
	return
}

func prop_write_perms(player, obj dbref, name string, mlev int) bool {
	if Prop_System(name) {
		return false
	}
	if mlev < MASTER {
		if !permissions(player, obj) {
			switch {
			case Prop_Private(name):
				return false
			case Prop_ReadOnly(name):
				return false
			case name == "sex":
				return false
			}
		}
		if strings.Prefix(name, "_msgmacs/") {
			return false
		}
	}
	if mlev < WIZBIT {
		if Prop_SeeOnly(name) || Prop_Hidden(name) {
			return false
		}
	}
	return true
}

func prim_getpropval(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		prop := extract_propname(op[1].(string))
		if !prop_read_perms(ProgUID, obj, prop, mlev) {
			panic("Permission denied.")
		}
		result := get_property_value(obj, prop)
		push(arg, top, result)
	})
}

func prim_getpropfval(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		prop := extract_propname(op[1].(string))
		if !prop_read_perms(ProgUID, obj, prop, mlev) {
			panic("Permission denied.")
		}
		result := get_property_fvalue(obj, prop)
		push(arg, top, result)
	})
}

func prim_getprop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		prop := extract_propname(op[1].(string))
		if !prop_read_perms(ProgUID, obj, prop, mlev) {
			panic("Permission denied.")
		}
		if ptr := get_property(obj, prop); ptr != nil {
			switch v := ptr.(type) {
			case string:
				push(arg, top, v)
			case Lock:
				push(arg, top, v)
			case dbref:
				push(arg, top, v)
			case int:
				push(arg, top, v)
			case float64:
				push(arg, top, v)
			default:
				push(arg, top, 0)
			}
		} else {
			push(arg, top, 0)
		}
	})
}

func prim_getpropstr(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		prop := extract_propname(op[1].(string))
		if !prop_read_perms(ProgUID, obj, prop, mlev) {
			panic("Permission denied.")
		}
		var value string
		if ptr := get_property(obj, type); ptr != nil {
			switch v := ptr.(type) {
			case string:
				value = v
			case dbref:
				value = fmt.Sprintf("#%d", v)
			case Lock:
				value = v.Unparse(ProgUID, true)
			}
		}
		push(arg, top, temp)
	})
}

func prim_remove_prop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		prop := extract_propname(op[0].(string))
		obj := valid_remote_object(player, mlev, op[1])
		switch {
		case prop == "":
			panic("Can't remove root propdir (2)")
		case !prop_write_perms(ProgUID, oper2->data.objref, buf, mlev) {
			panic("Permission denied.")
		}
		remove_property(obj, prop)
		ts_modifyobject(obj)
	})
}

func prim_envprop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		what := valid_remote_object(player, mlev, op[0])
		prop := extract_propname(op[1].(string))
		var p interface{}
		if what, p = envprop(what, prop); what != NOTHING && !prop_read_perms(ProgUID, what, prop, mlev) {
			panic("Permission denied.")
		}
		push(arg, top, what)
		if p == nil {
			push(arg, top, 0)
		} else {
			switch v := p.(type) {
			case string:
				push(arg, top, v)
			case int:
				push(arg, top, v)
			case float64:
				push(arg, top, v)
			case dbref:
				push(arg, top, v)
			case Lock:
				push(arg, top, v)
			default:
				push(arg, top, 0)
			}
		}
	})
}

func prim_envpropstr(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		what := valid_remote_object(player, mlev, op[0])
		rawprop := op[1].(string)
		prop := extract_propname(rawprop)
		var value string
		var p interface{}
		if what, p = envprop(what, prop); p != nil {
			switch v := p.(type) {
			case string:
				value = v
			case dbref:
				value = fmt.Sprintf("#%d", v)
			case Lock:
				value = v.Unparse(ProgUID, true)
			}
		}
		if what != NOTHING && !prop_read_perms(ProgUID, what, rawprop, mlev) {
			panic("Permission denied.")
		}
		push(arg, top, what)
		push(arg, top, value)
	}
}

func prim_blessprop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 2, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		prop := extract_propname(op[1].(string))
		set_property_flags(obj, prop, PROP_BLESSED)
		ts_modifyobject(obj)
	})
}

func prim_unblessprop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 2, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		prop := extract_propname(op[1].(string))
		clear_property_flags(obj, prop, PROP_BLESSED)
		ts_modifyobject(obj)
	})
}

func prim_setprop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		prop := extract_propname(op[1].(string))
		switch {
		case JOURNEYMAN < 2 && !permissions(ProgUID, obj):
			panic("Permission denied.")
		case !prop_write_perms(ProgUID, obj, prop, mlev):
			panic("Permission denied.")
		}

		var propdat interface{}
		switch value := op[2].(type) {
		case string:
			propdat = value
		case int:
			propdat = value
		case float64:
			propdat = value
		case dbref:
			propdat = value
		case Lock:
			propdat = copy_bool(value)
		default:
			panic("Invalid argument type (3)")
		}
		set_property(obj, prop, propdat)
		ts_modifyobject(obj)
	})
}

func prim_addprop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(4, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		prop := extract_propname(op[1])
		s := op[2].(string)
		i := op[3].(int)

		switch {
		case mlev < JOURNEYMAN && !permissions(ProgUID, obj):
			panic("Permission denied.")
		case !prop_write_perms(ProgUID, obj, prop, mlev):
			panic("Permission denied.")
		}
		add_property(obj, prop, s, i)
		ts_modifyobject(obj)
	})
}

func prim_nextprop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 2, top, func(op Array) {
		obj := valid_object(op[0])
		prop := extract_propname(op[1].(string))
		nextprop := next_prop_name(obj, prop)
		for prop != "" && !prop_read_perms(ProgUID, obj, prop, mlev) {
			nextprop = next_prop_name(ref, prop)
		}
		push(arg, top, nextprop)
	})
}

func prim_propdirp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(JOURNEYMAN, mlev, 2, top, func(op Array) {
		obj := valid_object(op[0])
		prop := op[1].(dbref)
		result = is_propdir(obj, prop)
		push(arg, top, result)
	})
}

func prim_parseprop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 4, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		ptr := op[2].(string)
		propname := extract_propname(op[1].(string))
		isprivate := op[3].(int)
		switch {
		case isprivate < 0, isprivate > 1:
			panic("Integer of 0 or 1 expected. (4)")
		case !prop_read_perms(ProgUID, obj, propname, mlev):
			panic("Permission denied.")
		case mlev < MASTER && !permissions(player, obj) && prop_write_perms(ProgUID, obj, propname, mlev):
			panic("Permission denied.")
		}

		propclass := get_property_class(obj, type)
		if propclass != "" {
			result = 0
			if isprivate {
				result |= MPI_ISPRIVATE
			}
			if Prop_Blessed(obj, type) {
				result |= MPI_ISBLESSED
			}
			propclass = do_parse_mesg(fr.descr, player, obj, temp, ptr, result)
		}
		push(arg, top, propclass)
	})
}

func prim_array_filter_prop(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		arr := op[0].(Array)
		if !array_is_homogenous(arr, dbref(0)) {
			panic("Argument not an array of dbrefs. (1)")
		}
		prop := extract_propname(op[1].(string))
		pattern := op[2].(string)
		nu := make(Array)
		for _, temp1 := range arr {
			ref := valid_remote_object(player, mlev, temp1)
			if prop_read_perms(ProgUID, ref, prop, mlev) {
				var buf string
				if pptr := get_property(ref, prop); pptr != nil {
					switch v := pptr.data.(type) {
					case string:
						buf = v
					case Lock:
						buf = v.Unparse(ProgUID, false)
					case dbref:
						buf = fmt.Sprintf("#%i", v)
					case int:
						buf = fmt.Sprint(v)
					case float64:
						buf = fmt.Sprint(v)
					default:
						buf = ""
					}
				}

				if !smatch(pattern, buf) {
					nu = append(nu, in)
				}
			}
		}
		push(arg, top, nu)
	})
}

func prim_reflist_find(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		prop := extract_propname(op[1].(string))
		prog := op[2].(dbref)
		if !prop_read_perms(ProgUID, obj, prop, mlev) {
			panic("Permission denied.")
		}
		push(arg, top, reflist_find(obj, prop, prog))
	})
}


func prim_reflist_add(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		propobj := valid_remote_object(player, mlev, op[0])
		prop := extract_propname(op[1].(string))
		obj := op[2].(dbref)
		if !prop_write_perms(ProgUID, propobj, prop, mlev) {
			abort_interp("Permission denied.")
		}
		reflist_add(propobj, prop, obj)
	})
}

func prim_reflist_del(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		propobj := valid_remote_object(player, mlev, op[0])
		propname := extract_propname(op[1].(string))
		obj := op[2].(dbref)
		if !prop_write_perms(ProgUID, propobj, propname, mlev) {
			panic("Permission denied.")
		}
		reflist_del(propobj, propname, obj)
	})
}

func prim_blessedp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(JOURNEYMAN, mlev, 2, top, func(op Array) {
		obj := valid_object(op[0])
		prop := op[1].(string)
		if Prop_Blessed(obj, prop) {
			result = 1
		} else {
			result = 0
		}
		push(arg, top, result)
	})
}

func prim_parsepropex(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 4, top, func(op Array) {
		obj := valid_remote_object(player, mlev, op[0])
		prop := extract_propname(op[1].(string))
		vars := op[2].(Dictionary)
		private := op[3].(int)
		switch {
		case private < 0, private > 1:
			panic("Integer of 0 or 1 expected. (4)")
		case !prop_read_perms(ProgUID, object, prop, mlev):
			panic("Permission denied.")
		}

		buffers := make(Dictionary)
		var flags int
		for key, val := range vars {
			switch key {
			case "":
				panic("Empty string keys not supported. (3)")
			case "how":
				flags |= MPI_NOHOW
			}
			switch val.(type) {
			case int, float64, dbref, string, Lock:
			default:
				panic("Only integer, float, dbref, string and lock values supported. (3)")
			}
		}

		var mesg string
		if mpi := get_property_class(obj, prop); mpi != "" {
			for key, val:= range vars {
				switch val := val.(type) {
				case int:
					set_mvalue(MPI_VARIABLES, key, fmt.Sprint(val))
				case float64:
					set_mvalue(MPI_VARIABLES, key, fmt.Sprint(val))
				case dbref:
					set_mvalue(MPI_VARIABLES, key, fmt.Sprintf("#%i", val))
				case string:
					set_mvalue(MPI_VARIABLES, key, val)
				case Lock:
					set_mvalue(MPI_VARIABLES, key, val.Unparse(ProgUID, true))
				default:
					set_mvalue(MPI_VARIABLES, key, "")
				}
			}

			if private != 0 {
				flags |= MPI_ISPRIVATE
			}
			if Prop_Blessed(object, prop) {
				flags |= MPI_ISBLESSED
			}
			mesg = do_parse_mesg(fr.descr, player, obj, mpi, "(parsepropex)", flags)
		}
		push(arg, top, buffers)
		push(arg, top, mesg)
	})
}