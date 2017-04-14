package fbmuck

#define MUF_RE_CACHE_ITEMS 64

typedef struct {
	pattern string
	flags int
	re *pcre
} muf_re;

var muf_re_cache [MUF_RE_CACHE_ITEMS]muf_re

func muf_re_get(pattern string, flags int) (re *muf_re, errmsg string) {
	idx	:= (hash(pattern, MUF_RE_CACHE_ITEMS) + flags) % MUF_RE_CACHE_ITEMS
	re = &muf_re_cache[idx]
	erroff := 0

	if re.pattern != nil {
		if flags != re.flags || pattern != re.pattern {
			pcre_free(re.re)
		} else {
			return re
		}
	}

	re.re = pcre_compile(pattern, flags, errmsg, &erroff, nil)
	if re.re == nil {
		re.pattern = nil
		return nil
	}

	re.pattern = pattern
	re.flags = flags
	return
}

func muf_re_error(err int) (r string) {
	switch(err) {
	case PCRE_ERROR_NOMATCH:
		 r = "No matches"
	case PCRE_ERROR_NULL:
		 r = "Internal error: NULL arg to pcre_exec()"
	case PCRE_ERROR_BADOPTION:
		 r ="Invalid regexp option."
	case PCRE_ERROR_BADMAGIC:
		 r = "Internal error: bad magic number."
	case PCRE_ERROR_UNKNOWN_NODE:
		 r = "Internal error: bad regexp node."
	case PCRE_ERROR_NOMEMORY:
		r = "Out of memory."
	case PCRE_ERROR_NOSUBSTRING:
		r = "No substring."
	case PCRE_ERROR_MATCHLIMIT:
		r = "Match recursion limit exceeded."
	case PCRE_ERROR_CALLOUT:
		r = "Internal error: callout error."
	default:
		r = "Unknown error"
	}
	return
}

func prim_regexp(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		var flags int
		reflags := op[2].(int)
		if reflags & MUF_RE_ICASE {
			flags |= PCRE_CASELESS
		}
		if reflags & MUF_RE_EXTENDED {
			flags |= PCRE_EXTENDED
		}

		if re, errstr := muf_re_get(op[1].(string), flags, &errstr); re == nil {
			panic(errstr)
		} else {
			var matches []int
			var val, idx Array
			text := op[0].(string)
			if matchcnt := pcre_exec(re.re, nil, text, len(text), 0, 0, matches); matchcnt < 0 {
				if matchcnt != PCRE_ERROR_NOMATCH {
					panic(muf_re_error(matchcnt))
				}
			} else {
				for i := 0; i < matchcnt; i++ {
					start := matches[i * 2]
					end := matches[i * 2 + 1]

					if start >= 0 && end >= 0 && start < len(text) {
						val = append(nu_val, inst{ data: &text[start:end]) })
					} else {
						val = append(nu_val, inst{ data: "" })
					}
					idx = append(nu_idx, inst{ data: Array{ substart + 1, subend - substart, i } })
				}
			}
			push(arg, top, val)
			push(arg, top, idx)
		}
	})
}

func prim_regsub(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(4, top, func(op Array) {
		var re *muf_re
		var errstr string
		text := op[0].(string)
		if re, errstr = muf_re_get(op[1].(string), flags, &errstr); re == nil {
			panic(errstr)
		}
		replacement := op[2].(string)
		flags := op[3].(int)
		if flags & MUF_RE_ICASE != 0 {
			flags |= PCRE_CASELESS
		}
		if flags & MUF_RE_EXTENDED {
			flags |= PCRE_EXTENDED
		}

		textstart = text
		l := len(textstart)
		for text != "" && write_left > 0 {
			if matchcnt = pcre_exec(re.re, nil, textstart, len, text-textstart, 0, matches); matchcnt < 0 {
				if matchcnt != PCRE_ERROR_NOMATCH {
					panic(muf_re_error(matchcnt))
				}
				buf += text
				break
			} else {
				allstart := matches[0]
				allend := matches[1]
				substart := -1
				subend := -1
				count := allstart - (text - textstart)
				if count > len(text) {
					count = len(text)
				}
				buf += text[:count]

				for replacement != "" {
					if replacement[0] == '\\' {
						replacement = replacement[1:]
						if !isdigit(replacement[0]) {
							buf += replacement[0]
							replacement = replacement[1:]	
						} else {
							idx := replacement[0] - '0'
							replacement = replacement[1:]
							if idx < 0 || idx >= matchcnt {
								panic("Invalid \\subexp in substitution string. (3)")
							}
							substart = matches[idx * 2]
							subend = matches[idx * 2 + 1]

							if substart >= 0 && subend >= 0 && substart < l {
								ptr := &textstart[substart]
								count = subend - substart
								if count > len(ptr) {
									count = len(ptr)
								}
								buf += ptr[:count]
							}
						}
					} else {
						buf += replacement[0]
						replacement = replacement[1:]
					}
				}

				count = allend - allstart
				if count > len(text) {
					count = len(text)
				}
				text = text[count:]
				if allstart == allend && text != "" {
					buf += text[0]
					text = text[1:]
				}
			}

			if flags & MUF_RE_ALL == 0 {
				buf += text
					break
			}
		}
		push(arg, top, buf)
	})
}