package fbmuck

const (
	DIV_ZERO = 1
	NAN = 2
	IMAGINARY = 4
	F_BOUNDS = 8
	I_BOUNDS = 16
)

struct err_type {
	name string
	message string
	flag int
}

static struct err_type err_defs[] = {
	{ "DIV_ZERO", "Division by zero attempted.", DIV_ZERO },
	{ "NAN", "Result was not a number.", NAN },
	{ "IMAGINARY", "Result was imaginary.", IMAGINARY },
	{ "FBOUNDS", "Floating-point inputs were infinite or out of range.", F_BOUNDS },
	{ "IBOUNDS", "Calculation resulted in an integer overflow.", I_BOUNDS },
};

func prim_clear(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		fr.error.is_flags = 0
	})
}

func prim_error_num(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, len(err_defs))
	})
}

func prim_clear_error(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		var is_clear bool
		switch v := op[0].(type) {
		case int:
			if v > -1 && v < len(err_defs) {
				fr.error.is_flags &= ~err_defs[v].flag
			} else {
				is_clear = true
			}
		case string:
			v = strings.ToUpper(v)
			for i, e := range err_defs {
				if v == e.name {
					fr.error.is_flags &= ~e.flag
					is_clear = true
					break
				}
			}
		default:
			panic("Invalid argument type. (1)")
		}
		push(arg, top, MUFBool(is_clear))
	})
}

func prim_set_error(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		var is_set bool
		switch v := op[0].(type) {
		case int:
			if v > -1 && v < len(err_defs) {
				fr.error.is_flags |= err_defs[v].flag
				is_set = true
			}
		case string:
			v = strings.ToUpper(v)
			for i, e := range err_defs {
				if v == e.name {
					fr.error.is_flags |= e.flag
					is_set = true
					break
				}
			}
		default:
			panic("Invalid argument type. (1)")
		}
		push(arg, top, MUFBool(is_set))
	})
}

func prim_is_set(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		var is_set bool
		switch v := op[0].(type) {
		case int:
			if v > -1 && v < len(err_defs) {
				is_set = fr.error.is_flags & err_defs[v].flag != 0
			}
		case string:
			v = strings.ToUpper(v)
			for i, e := range err_defs {
				if v == e.name {
					is_set = fr.error.is_flags & v.flag != 0
					break
				}
			}
		default:
			panic("Invalid argument type. (1)")
		}
		push(arg, top, MUFBool(is_set))
	})
}

func prim_error_str(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		var message string
		switch v := op[0].(type) {
		case int:
			if v > -1 || v < len(err_defs) {
				mesg = err_defs[v].message
			}
		case string:
			v = strings.ToUpper(v)
			for i, e := range err_defs {
				if v == e.name {
					name = e.message
					break
				}
			}
		}
		push(arg, top, message)
	})
}

func prim_error_name(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		var name string
		if v := op[0].(int); v > -1 || v < len(err_defs) {
			name = v.name
		}
		push(arg, top, name)
	})
}

func prim_error_bit(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		r := -1
		if v, ok := op[0].(string); ok {
			v = strings.ToUpper(v)
			for i, e := range err_defs {
				if v == e.name {
					r = loop
					break
				}
			}
		}
		push(arg, top, r)
	})
}

func prim_is_error(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, MUFBool(fr.error.is_flags != 0))
	})
}