package fbmuck

/* these arrays MUST agree with what's in inst.h */
const char *base_inst[] = {
	"JMP", "READ", "SLEEP", "CALL", "EXECUTE", "EXIT", "EVENT_WAITFOR", "CATCH", "CATCH_DETAILED",
	PRIMS_CONNECTS_NAMES,
	PRIMS_DB_NAMES,
	PRIMS_MATH_NAMES,
	PRIMS_MISC_NAMES,
	PRIMS_PROPS_NAMES,
	PRIMS_STACK_NAMES,
	PRIMS_STRINGS_NAMES,
	PRIMS_ARRAY_NAMES,
	PRIMS_FLOAT_NAMES,
	PRIMS_ERROR_NAMES,
	PRIMS_MCP_NAMES,
	PRIMS_REGEX_NAMES,
	PRIMS_INTERNAL_NAMES
};

/* converts an instruction into a printable string, stores the string in
   buffer and returns a pointer to it.
 */
func insttotext(struct frame *fr, int lev, struct inst *theinst, ObjectID program, int expandarrs) (buffer string) {
	const char* ptr;
	char buf2[BUFFER_LEN];
	struct inst temp1;
	struct inst *oper2;
	int firstflag = 1;
	int arrcount = 0;
	
	assert(buflen > 0);
	
	switch op := theinst.data.(type) {
	case PROG_PRIMITIVE:
		if op >= BASE_MIN && op <= BASE_MAX {
			buffer = base_inst[op - BASE_MIN]
		} else {
			buffer = "???"
		}
	case string:
		buffer = fmt.Sprint("\"", op, "\"")
	case Mark:
		buffer = "MARK"
	case Array:
		switch {
		case op == nil:
			buffer = "0{}"
		case tp_expanded_debug && expandarrs:
			s := make([]string, len(op))
			for i, v := range op {
				s[i] = fmt.Sprint(i, ":", insttotext(fr, lev, v, program, 0))
			}
			buffer = fmt.Sprintf("%d{%v}", len(s), strings.Join(s, " "))
		default:
			buffer = fmt.Sprintf("%d{...}", len(op));
		}
	case Dictionary:
		switch {
		case op == nil:
			buffer = "0{}"
		case tp_expanded_debug && expandarrs:
			s := make([]string, 0, len(op))
			for k, v := range op {
				s = append(s, fmt.Sprint(k, ":", insttotext(fr, lev, v, program, 0)))
			}
			buffer = fmt.Sprintf("%d{%v}", len(s), strings.Join(s, " "))
		default:
			buffer = fmt.Sprintf("%d{...}", len(op));
		}
	case int:
		buffer = fmt.Sprint(theinst.data.(int))
		length = len(buffer)
	case float64:
		buffer = fmt.Sprintf("%.16g", theinst->data.fnumber);
		length = len(buffer)
		if (!strchr(buffer, '.') && !strchr(buffer, 'n') && !strchr(buffer, 'e')) {
			strcatn(buffer, buflen, ".0");
		}
	case Address:
		if op.data.(type) == MUFProc && op.data.data.(MUFProc) != nil {
			if op.progref != program {
				buffer = fmt.Sprintf("'#%d'%s", op.progref, op.data.data.(MUFProc).name)
			} else {
				buffer = fmt.Sprintf("'%s", op.data.data.(MUFProc).name)
			}
			length = len(buffer)
		} else {
			if op.progref != program {
				buffer = fmt.Sprintf("'#%d'line%d?", op.progref, op.data.line)
			} else {
				buffer = fmt.Sprintf("'line%d?", op.data.line)
			}
			length = len(buffer)
		}
	case PROG_TRY:
		buffer = fmt.Sprintf("TRY->line%d", op.call.line)
		length = len(buffer)
	case PROG_IF:
		buffer = fmt.Sprintf("IF->line%d", op.call.line)
		length = len(buffer)
	case PROG_EXEC:
		if v, ok := op.call.data.(MUFProc); ok {
			buffer = fmt.Sprintf("EXEC->%s", v.name)
		} else {
			buffer = fmt.Sprintf("EXEC->line%d", op.call.line)
		}
		length = len(buffer)
	case PROG_JMP:
		if v, ok := op.call.data.(MUFProc); ok {
			buffer = fmt.Sprintf("JMP->%s", v.name)
		} else {
			buffer = fmt.Sprintf("JMP->line%d", op.call.line)
		}
		length = len(buffer)
	case ObjectID:
		buffer = fmt.Sprintf("#%d", op)
		length = len(buffer)
	case PROG_VAR:
		buffer = fmt.Sprintf("V%d", op)
		length = len(buffer)
	case PROG_SVAR:
		if fr != nil {
			buffer = fmt.Sprintf("SV%d:%s", op, scopedvar_getname(fr, lev, op))
		} else {
			buffer = fmt.Sprintf("SV%d", op)
		}
		length = len(buffer)
	case PROG_SVAR_AT, PROG_SVAR_AT_CLEAR:
		if fr != nil {
			buffer = fmt.Sprintf("SV%d:%s @", op, scopedvar_getname(fr, lev, op))
		} else {
			buffer = fmt.Sprintf("SV%d @", op)
		}
		length = len(buffer)
	case PROG_SVAR_BANG:
		if fr != nil {
			buffer = fmt.Sprintf("SV%d:%s !", op, scopedvar_getname(fr, lev, op))
		} else {
			buffer = fmt.Sprintf("SV%d !", op)
		}
		length = len(buffer)
	case PROG_LVAR:
		buffer = fmt.Sprintf("LV%d", op)
		length = len(buffer)
	case PROG_LVAR_AT, PROG_LVAR_AT_CLEAR:
		buffer = fmt.Sprintf("LV%d @", op)
		length = len(buffer)
	case PROG_LVAR_BANG:
		buffer = fmt.Sprintf("LV%d !", op)
		length = len(buffer)
	case MUFProc:
		if op.args == 1 {
			buffer = fmt.Sprintf("INIT FUNC: %s (%d arg)", op.name, op.args)
		} else {
			buffer = fmt.Sprintf("INIT FUNC: %s (%d args)", op.name, op.args)
		}
		length = len(buffer)
	case Lock:
		buffer = fmt.Sprintf("[%s]", op.Unparse(0, false))
	default:
		buffer = "?"
	}
	return
}

/* produce one line summary of current state.  Note that sp is the next
 *    space on the stack -- 0..sp-1 is the current contents. */

#define DEBUG_DEPTH 8 /* how far to give a stack list, at most */

debug_inst(fr *frame, lev int, pc *inst, pid int, stack *inst, buffer string, buflen, sp int, program ObjectID) string {
	char* bend;
	char* bstart;
	char* ptr;
	int length;

	char buf2[BUFFER_LEN];
	/* To hold Debug> ... at the beginning */
	char buf3[64];
	int count;

	buffer[buflen - 1] = '\0';

	buf3 = fmt.Sprintf("Debug> Pid %d: #%d %d (", pid, program, pc.line)
	length = len(buf3)
	if length == -1 {
		length = sizeof(buf3) - 1;
	}
	bstart = buffer + length		/* start far enough away so we can fit Debug> #xxx xxx ( thingy. */
	length = buflen - length - 1	/* - 1 for the '\0' */
	bend = buffer + (buflen - 1)	/* - 1 for the '\0' */

	/* + 10 because we must at least be able to store " ... ) ..." after that. */
	if (bstart + 10 > bend) {	/* we have no room. Eeek! */
	    /*					123456789012345678 */
	    memcpy((void*)buffer, (const void*)"Need buffer space!", (buflen - 1 > 18) ? 18 : buflen - 1 );
	    return buffer;
	}

	ptr = insttotext(fr, lev, pc, program, 1);
	if (*ptr) {
	    length -= prepend_string(&bend, bstart, ptr);
	} else {
		strcpyn(buffer, buflen, buf3);
		strcatn(buffer, buflen, " ... ) ...");
		return buffer;
	}
	
	length -= prepend_string(&bend, bstart, ") ");
	
	if count = sp - 1; count >= 0 {
	    for {
			if count && length <= 5 {
				length -= prepend_string(&bend, bstart, "...")
				break
			}
			ptr = insttotext(fr, lev, stack + count, program, 1)
			if ptr != "" {
				length -= prepend_string(&bend, bstart, ptr)
			} else {
				length -= prepend_string(&bend, bstart, "...")
				break	/* done because we couldn't display all that */
			}
			if count > 0 && count > sp - 8 {
				length -= prepend_string(&bend, bstart, ", ")
			} else {
				if count != 0 {
					length -= prepend_string(&bend, bstart, "..., ")
				}
				break	/* all done! */
			}
			count--
	    }
	}

	/* we don't use bstart, because it's compensated for the length of this. */
	prepend_string(&bend, buffer, buf3)
	/* and return the pointer to the beginning of our backwards grown string. */
	return bend
}