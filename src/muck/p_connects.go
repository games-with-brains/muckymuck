package fbmuck

func prim_awakep(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		switch ref := op[0].(ObjectID).ValidObject(); {
		case Typeof(ref) == TYPE_THING && DB.Fetch(ref).flags & ZOMBIE != 0:
			ref = DB.Fetch(ref).Owner
		case Typeof(ref) != TYPE_PLAYER:
			panic("invalid argument.")
		}
		push(arg, top, online(ref))
	})
}

func prim_online(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 0, top, func(op Array) {
		mycount := current_descr_count
		CHECKOFLOW(mycount + 1)
		for i := mycount; i > 0; i-- {
			push(arg, top, pObjectID(i))
		}
		push(arg, top, mycount)
	})
}

func prim_online_array(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 0, top, func(op Array) {
		CHECKOFLOW(1)
		nu := make(Array, current_descr_count)
		for i, v := range nu {
			nu[i] = pObjectID(i + 1)
		}
		push(arg, top, nu)
	})
}

func prim_concount(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, current_descr_count)
	})
}

func prim_descr(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, fr.descr)
	})
}

func prim_conObjectID(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		switch i := op[0].(int); {
		case i < 1, i > current_descr_count:
			panic("Invalid connection number. (1)")
		default:
			push(arg, top, pObjectID(i))
		}
	})
}

func prim_descr_ObjectID(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		if d := lookup_descriptor(op[0].(int)); d != nil {
			push(arg, top, d.player)
		} else {
			push(arg, top, NOTHING)
		}
	})
}

func prim_conidle(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		switch i := op[0].(int); {
		case i < 1, i > current_descr_count:
			panic("Invalid connection number. (1)")
		default:
			push(arg, top, pidle(i))
		}
	})
}

func prim_descr_idle(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		if d := lookup_descriptor(op[0].(int)); d != nil {
			push(arg, top, time.Now() - d.last_time)
		} else {
			panic("Invalid descriptor number. (1)")
		}
	})
}

func prim_descr_least_idle(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		if i := pdescr(least_idle_player_descr(op[0].(ObjectID).ValidObject())); i == 0 {
			panic("Invalid descriptor number. (1)")
		} else {
			push(arg, top, i)
		}
	})
}

func prim_descr_most_idle(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		if i := pdescr(most_idle_player_descr(op[0].(ObjectID).ValidObject())); i == 0 {
			panic("Invalid descriptor number. (1)")
		} else {
			push(arg, top, i)
		}
	})
}

func prim_contime(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		switch i := op[0].(int); {
		case i < 1, i > current_descr_count:
			panic("Invalid connection number. (1)")
		default:
			push(arg, top, pontime(i))
		}
	})
}

func prim_descr_time(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		now := time.Now()
		if d := lookup_descriptor(op[0].(int)); d != nil {
			push(arg, top, now - d.connected_at)
		} else {
			panic("Invalid descriptor number. (1)")
		}
	})
}

func prim_conhost(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 1, top, func(op Array) {
		switch i := op[0].(int); {
		case i < 1, i > current_descr_count:
			panic("Invalid connection number. (1)")
		default:
			if d := descrdata_by_count(i); d != nil {
				push(arg, top, d.hostname)
			} else {
				push(arg, top, "")
			}

		}
	})
}

func prim_descr_host(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 1, top, func(op Array) {
		if d := lookup_descriptor(op[0].(int)); d != nil {
			push(arg, top, d.hostname)
		} else {
			panic("Invalid descriptor number. (1)")
		}
	})
}

func prim_conuser(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 1, top, func(op Array) {
		switch i := op[0].(int); {
		case i < 1, i > current_descr_count:
			panic("Invalid connection number. (1)")
		default:
			if d := descrdata_by_count(i); d != nil {
				push(arg, top, d.username)
			} else {
				push(arg, top, "")
			}
		}
	})
}

func prim_descr_user(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 1, top, func(op Array) {
		if d := lookup_descriptor(op[0].(int)); d != nil {
			push(arg, top, d.username)
		} else {
			panic("Invalid descriptor number. (1)")
		}
	})
}

func prim_conboot(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 1, top, func(op Array) {
		switch i := op[0].(int); {
		case i < 1, i > current_descr_count:
			panic("Invalid connection number. (1)")
		default:
			pboot(i)
		}
	})
}

func prim_descr_boot(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 1, top, func(op Array) {
	    if d := lookup_descriptor(op[0].(int)); d != nil {
			process_output(d)
			d.booted = 1
			/* shutdownsock(d) */
	    } else {
			panic("Invalid descriptor number. (1)")
		}
	})
}

func prim_connotify(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 2, top, func(op Array) {
		switch i := op[0].(int); {
		case i < 1, i > current_descr_count:
			panic("Invalid connection number. (1)")
		default:
			pnotify(result, op[1].(string))
		}
	})
}

func prim_descr_notify(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 2, top, func(op Array) {
		if d := lookup_descriptor(op[0].(int)); d != nil {
			queue_msg(d, op[1].(string))
			d.QueueWrite("\r\n")
		} else {
			panic("Invalid descriptor number. (1)")
		}
	})
}

func prim_condescr(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		switch i := op[0].(int); {
		case i < 1, i > current_descr_count:
			panic("Invalid connection number. (1)")
		default:
			push(arg, top, pdescr(i))
		}
	})
}

func prim_descrcon(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		push(arg, top, pdescrcon(op[0].(int)))
	})
}

func prim_nextdescr(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		var d *descriptor_data
	    if d = lookup_descriptor(op[0].(int)); d != nil {
			for d = d.next; d != nil && !d.connected; d = d.next {}
		}
		if d != nil {
			push(arg, top, d.descriptor)
		} else {
			push(arg, top, 0)
		}
	})
}

func prim_descriptors(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		if op[0] == NOTHING {
			mycount := current_descr_count
			CHECKOFLOW(mycount + 1)
			for i := mycount; i != 0; i-- {
				push(arg, top, pdescr(i))
			}
			push(arg, top, mycount)
		} else {
			ref := op[0].(ObjectID).ValidPlayer()
			arr := get_player_descrs(ref)
			mycount := len(arr)
			CHECKOFLOW(mycount + 1)
			for _, v := range arr {
            	push(arg, top, v)
        	}
			push(arg, top, mycount)
		}
	})
}

func prim_descr_array(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		if op[0] == NOTHING {
			nu := make(Array, current_descr_count)
			for i, _ := range nu {
				nu[i] = pdescr(i + 1)
    		}
			push(arg, top, nu)
		} else {
			ref := op[0].(ObjectID).ValidPlayer()
			arr := get_player_descrs(ref)
			nu := make(Array, len(arr))
			l := len(arr) - 1
			for i, v := range arr {
				nu[l - i] = v
        	}
			push(arg, top, nu)
		}
	})
}

func prim_descr_setuser(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(WIZBIT, mlev, 3, top, func(op Array) {
		if op[1] != NOTHING {
			descr := op[0].(int)
			ref := op[1].(ObjectID).ValidPlayer(func(obj ObjectID) {
				if !check_password(obj, op[2].(string)) {
					panic("Incorrect password.")
				} else {
					log_status("DESCR_SETUSER: %s(%d) to %s(%d) on descriptor %d", DB.Fetch(player).name, player, DB.Fetch(ref).name, ref, descr)
				}
			})
		    if d := lookup_descriptor(descr); d != nil && d.connected {
				announce_disconnect(d)
				if who != NOTHING {
					d.player = ref
					d.connected = true
					update_desc_count_table()
		            remember_player_descr(who, d.descriptor)
					announce_connect(d.descriptor, ref)
				}
				push(arg, top, 1)
			} else {
				push(arg, top, 0)
			}
		}
	})
}

func prim_descrflush(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		if op[0].(int) != -1 {
			if d := lookup_descriptor(c); d != nil {
				if !process_output(d) {
					d.booted = 1
				}
				r++
			}
		} else {
			for d := descriptor_list; d != nil; d = d.next {
				if !process_output(d) {
					d.booted = 1
				}
				r++
			}
		}
	})
}

func prim_firstdescr(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		var status int
		if op[0] == NOTHING {
			if d := descrdata_by_count(1); d != nil {
				status = d.descriptor
		} else {
			op[0].(ObjectID).ValidPlayer(func(obj ObjectID) {
				if online(obj) {
					arr := get_player_descrs(obj)
					status = index_descr(arr[len(arr) - 1])
				}
			})
		}
		push(arg, top, status)
	})
}

func prim_lastdescr(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		var status int
		if op[0] == NOTHING {
			d := descrdata_by_count(current_descr_count); d != nil {
				status = d.descriptor
			}
		} else {
			op[0].(ObjectID).ValidPlayer(func(obj ObjectID) {
				status = index_descr(get_player_descrs(obj)[0])
			})
		}
		push(arg, top, status)
	})
}

func prim_descr_securep(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		push(arg, top, MUFBool(false))
	})
}

func prim_descr_bufsize(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_restricted_primitive(MASTER, mlev, 1, top, func(op Array) {
		if d := lookup_descriptor(op[0].(int)); d != nil {
			push(arg, top, tp_max_output - d.output_size)
		} else {
			panic("Invalid descriptor number. (1)")
		}
	})
}