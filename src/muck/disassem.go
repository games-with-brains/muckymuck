package fbmuck

func disassemble(player, program ObjectID) {
	curr := DB.Fetch(program).(Program).code
	codestart := curr
	if len(curr) == 0 {
		notify(player, "Nothing to disassemble!")
		return
	}
	var buf string
	for i, v := range DB.Fetch(program).(Program).code {
		switch op := v.data.(type) {
		case PROG_PRIMITIVE:
			if op >= BASE_MIN && op <= BASE_MAX {
				buf = fmt.Sprintf("%d: (line %d) PRIMITIVE: %s", i, v.line, base_inst[op - BASE_MIN])
			} else {
				buf = fmt.Sprintf("%d: (line %d) PRIMITIVE: %d", i, v.line, op)
			}
		case Mark:
			buf = fmt.Sprintf("%d: (line %d) MARK", i, v.line)
		case string:
			buf = fmt.Sprintf("%d: (line %d) STRING: \"%s\"", i, v.line, op)
		case Array:
			buf = fmt.Sprintf("%d: (line %d) ARRAY: %d items", i, v.line, len(op))
		case Dictionary:
			buf = fmt.Sprintf("%d: (line %d) ARRAY: %d items", i, v.line, len(op))
		case MUFProc:
			buf = fmt.Sprintf("%d: (line %d) FUNCTION: %s, VARS: %d, ARGS: %d", i, v.line, op.name, op.vars, op.args)
		case Lock:
			buf = fmt.Sprintf("%d: (line %d) LOCK: [%s]", i, v.line, op.Unparse(0, false))
		case int:
			buf = fmt.Sprintf("%d: (line %d) INTEGER: %d", i, v.line, op)
		case float64:
			buf = fmt.Sprintf("%d: (line %d) FLOAT: %.17g", i, v.line, op)
		case Address:
			buf = fmt.Sprintf("%d: (line %d) ADDRESS: %d", i, v.line, op - codestart)
		case PROG_TRY:
			buf = fmt.Sprintf("%d: (line %d) TRY: %d", i, v.line, op.call - codestart)
		case PROG_IF:
			buf = fmt.Sprintf("%d: (line %d) IF: %d", i, v.line, op.call - codestart)
		case PROG_JMP:
			buf = fmt.Sprintf("%d: (line %d) JMP: %d", i, v.line, op.call - codestart)
		case PROG_EXEC:
			buf = fmt.Sprintf("%d: (line %d) EXEC: %d", i, v.line, op.call - codestart)
		case ObjectID:
			buf = fmt.Sprintf("%d: (line %d) OBJECT REF: %d", i, v.line, op)
		case PROG_VAR:
			buf = fmt.Sprintf("%d: (line %d) VARIABLE: %d", i, v.line, op)
		case PROG_SVAR:
			buf = fmt.Sprintf("%d: (line %d) SCOPEDVAR: %d (%s)", i, v.line, op, scopedvar_getname_byinst(curr, op))
		case PROG_SVAR_AT:
			buf = fmt.Sprintf("%d: (line %d) FETCH SCOPEDVAR: %d (%s)", i, v.line, op, scopedvar_getname_byinst(curr, op))
		case PROG_SVAR_AT_CLEAR:
			buf = fmt.Sprintf("%d: (line %d) FETCH SCOPEDVAR (clear optim): %d (%s)", i, v.line, op, scopedvar_getname_byinst(curr, op))
		case PROG_SVAR_BANG:
			buf = fmt.Sprintf("%d: (line %d) SET SCOPEDVAR: %d (%s)", i, v.line, op, scopedvar_getname_byinst(curr, op))
		case PROG_LVAR:
			buf = fmt.Sprintf("%d: (line %d) LOCALVAR: %d", i, v.line, op)
		case PROG_LVAR_AT:
			buf = fmt.Sprintf("%d: (line %d) FETCH LOCALVAR: %d", i, v.line, op)
		case PROG_LVAR_AT_CLEAR:
			buf = fmt.Sprintf("%d: (line %d) FETCH LOCALVAR (clear optim): %d", i, v.line, op)
		case PROG_LVAR_BANG:
			buf = fmt.Sprintf("%d: (line %d) SET LOCALVAR: %d", i, v.line, op)
		default:
			buf = fmt.Sprintf("%d: (line ?) UNKNOWN INST", i)
		}
		notify(player, buf)
		curr = curr[1:]
	}
}