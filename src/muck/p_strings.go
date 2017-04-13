/* Primitives Package */

var temp1, temp2, temp3 inst
var tmp, result int
var ref dbref
var pname string

/* FMTTOKEN defines the start of a variable formatting string insertion */
#define FMTTOKEN '%'

// FIXME: rewrite fmtstring and fmtstrings to make better use of fmt.Sprintf

func prim_fmtstring(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		int slen, scnt, tstop, tlen, tnum, i;
		int slrj, spad1, spad2, slen1, slen2, temp;
		char sfmt[255], hold[256];
		char *ptr, *begptr;

		var sstr, tbuf string
		buf = ""
		if sstr, ok := op[0].(string); !ok {
			/* We now have the non-null format string, parse it */
			result = 0					/* End of current string */
			tmp = 0						/* Number of props to search for/found */
			slen = len(sstr)
			scnt = 0
			tstop = 0
			for scnt = 0; scnt < slen; {
				if sstr[scnt] == FMTTOKEN {
					if sstr[scnt + 1] == FMTTOKEN {
						buf += FMTTOKEN
						result++
						scnt += 2
					} else {
						scnt++
						switch sstr[scnt] {
						case '-':
							slrj = 1
							scnt++
						case '|':
							slrj = 2
							scnt++
						default:
							slrj = 0
						}

						switch sstr[scnt] {}
						case '+':
							spad1 = 1
							scnt++
						case ' ':
							spad1 = 2
							scnt++
						default:
							spad1 = 0
						}

						if sstr[scnt] == '0' {
							scnt++
							spad2 = 1
						} else {
							spad2 = 0
						}

						slen1 = strconv.Atoi(&sstr[scnt])
						if sstr[scnt] >= '0' && sstr[scnt] <= '9' {
							for sstr[scnt] >= '0' && sstr[scnt] <= '9' {
								scnt++
							}
						} else {
							if sstr[scnt] == '*' {
								scnt++
								checkop(1, top)
								slen1 = POP().data.(int)
							} else {
								slen1 = 0
							}
						}
						if sstr[scnt] == '.' {
							scnt++
							slen2 = strconv.Atoi(&sstr[scnt])
							if sstr[scnt] >= '0' && sstr[scnt] <= '9' {
								for sstr[scnt] >= '0' && sstr[scnt] <= '9' {
									scnt++
								}
							} else {
								if sstr[scnt] == '*' {
									scnt++
									checkop(1, top)
									if slen2 = POP().data.(int); slen2 < 0 {
										panic("Dynamic precision value must be a positive integer.")
									}
								} else {
									panic("Invalid format string.")
								}
							}
						} else {
							slen2 = -1
						}

						checkop(1, top)
						op := POP().data
						sfmt = "\%"
						if slrj == 1 {
							sfmt += "-"
						}
						switch spad1 {
						case 0:
						case 1:
							sfmt += "+"
						default:
							sfmt += " "
							}
						}
						if spad2 != 0 {
							sfmt += "0"
						}
						if slen1 != 0 {
							sfmt += fmt.Sprint(slen1)
						}
						if slen2 != -1 {
							sfmt += fmt.Sprintf(".%d", slen2)
						}

						if sstr[scnt] == '~' {
							switch op.(type) {
							case dbref:
								sstr[scnt] = 'D'
							case float64:
								sstr[scnt] = 'g'
							case int:
								sstr[scnt] = 'i'
							case Lock:
								sstr[scnt] = 'l'
							case string:
								sstr[scnt] = 's'
							default:
								sstr[scnt] = '?'
							}
						}
						switch sstr[scnt] {
						case 'i':
							sfmt += "d"
							tbuf = fmt.Sprintf(sfmt, op.(int))
							tlen = len(tbuf)
							if slrj == 2 {
								tnum = 0
								for tbuf[tnum] == ' ' && tnum < tlen {
									tnum++
								}
								if tnum > 0 && tnum < tlen {
									temp = tnum / 2
									for i := tnum; i < tlen; i++ {
										tbuf[i - temp] = tbuf[i]
									}
									for i := tlen - temp; i < tlen; i++ {
										tbuf[i] = ' '
									}
								}
							}
						case 'S', 's':
							sfmt += "s"
							v := op.(string)
							tbuf = fmt.Sprintf(sfmt, v)
							tlen = len(tbuf)
							if slrj == 2 {
								tnum = 0
								for tbuf[tnum] == ' ' && tnum < tlen {
									tnum++
								}
								if tnum > 0 && tnum < tlen {
									temp = tnum / 2
									for i := tnum; i < tlen; i++ {
										tbuf[i - temp] = tbuf[i];
									}
									for i := tlen - temp; i < tlen; i++ {
										tbuf[i] = ' '
									}
								}
							}
						case '?':
							sfmt += "s"
							switch op.(type) {
							case dbref:
								hold = "OBJECT"
							case float64:
								hold = "FLOAT"
							case int:
								hold = "INTEGER"
							case Lock:
								hold = "LOCK"
							case string:
								hold = "STRING"
							case PROG_VAR:
								hold = "VARIABLE"
							case PROG_LVAR:
								hold = "LOCAL-VARIABLE"
							case PROG_SVAR:
								hold = "SCOPED-VARIABLE"
							case Address:
								hold = "ADDRESS"
							case Array, Dictionary:
								hold = "ARRAY"
							case MUFProc:
								hold = "FUNCTION-NAME"
							case PROG_IF:
								hold = "IF-STATEMENT"
							case PROG_EXEC:
								hold = "EXECUTE"
								break;
							case PROG_JMP:
								hold = "JUMP"
							case PROG_PRIMITIVE:
								hold = "PRIMITIVE"
							default:
								hold = "UNKNOWN"
							}
							tbuf = fmt.Sprintf(sfmt, hold)
							tlen = len(tbuf)
							if slrj == 2 {
								tnum = 0
								for tbuf[tnum] == ' ' && tnum < tlen {
									tnum++
								}
								if tnum > 0 && tnum < tlen {
									temp = tnum / 2
									for i := tnum; i < tlen; i++ {
										tbuf[i - temp] = tbuf[i]
									}
									for i := tlen - temp; i < tlen; i++ {
										tbuf[i] = ' '
									}
								}
							}
						case 'd':
							sfmt += "s"
							obj := op.(dbref)
							hold = fmt.Sprintf("#%d", obj)
							tbuf = fmt.Sprintf(sfmt, hold)
							tlen = len(tbuf)
							if slrj == 2 {
								tnum = 0
								for tbuf[tnum] == ' ' && tnum < tlen {
									tnum
								}
								if tnum > 0 && tnum < tlen {
									temp = tnum / 2
									for i := tnum; i < tlen; i++ {
										tbuf[i - temp] = tbuf[i]
									}
									for i := tlen - temp; i < tlen; i++ {
										tbuf[i] = ' '
									}
								}
							}
						case 'D':
							sfmt += "s"
							ref := valid_remote_object(player, mlev, op)
							if db.Fetch(ref).name {
								hold = db.Fetch(ref).name
							} else {
								hold = ""
							}
							tbuf = fmt.Sprintf(sfmt, hold)
							tlen = len(tbuf)
							if slrj == 2 {
								tnum = 0
								for tbuf[tnum] == ' ' && tnum < tlen {
									tnum++
								}
								if tnum > 0 && tnum < tlen {
									temp = tnum / 2
									for i := tnum; i < tlen; i++ {
										tbuf[i - temp] = tbuf[i]
									}
									for i := tlen - temp; i < tlen; i++ {
										tbuf[i] = ' '
									}
								}
							}
						case 'l':
							sfmt += "s"
							lock := op.(Lock)
							hold = lock.Unparse(ProgUID, true)
							tbuf = fmt.Sprintf(sfmt, hold)
							tlen = len(tbuf)
							if slrj == 2 {
								tnum = 0
								for tbuf[tnum] == ' ' && tnum < tlen {
									tnum++
								}
								if tnum > 0 && tnum < tlen {
									temp = tnum / 2
									for i := tnum; i < tlen; i++ {
										tbuf[i - temp] = tbuf[i]
									}
									for i := tlen - temp; i < tlen; i++ {
										tbuf[i] = ' '
									}
								}
							}
						case 'f', 'e', 'g':
							sfmt += "l"
							hold = fmt.Sprint(sstr[scnt])
							sfmt += hold
							tbuf = fmt.Sprintf(sfmt, op.(float))
							tlen = len(tbuf)
							if slrj == 2 {
								tnum = 0
								for tbuf[tnum] == ' ' && tnum < tlen {
									tnum++
								}
								if tnum > 0 && tnum < tlen {
									temp = tnum / 2
									for i := tnum; i < tlen; i++ {
										tbuf[i - temp] = tbuf[i]
									}
									for i := tlen - temp; i < tlen; i++ {
										tbuf[i] = ' '
									}
								}
							}
						default:
							panic("Invalid format string.")
						}
						buf += tbuf
						result += tlen
						scnt++
						tstop += len(tbuf)
					}
				} else {
					if sstr[scnt] == '\\' && sstr[scnt + 1] == 't' {
						if tstop % 8 == 0) {
							buf += ' '
							result++
							tstop++
						}
						for tstop % 8 != 0 {
							buf += ' '
							result++
							tstop++
						}
						scnt += 2
						tstop = 0
					} else {
						if sstr[scnt] == '\r' {
							tstop = 0
							scnt++
							buf += '\r'
							result++
						} else {
							buf += sstr[scnt]
							scnt++
							result++
							tstop++
						}
					}
				}
			}
		}
		checkop(0, top)
		push(arg, top, buf)
	})
}

func prim_array_fmtstrings(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		var oper3 inst

		int slen, scnt, tstop, tlen, tnum, i;
		int slrj, spad1, spad2, slen1, slen2, temp;
		char sfmt[255], hold[256];
		char *ptr, *begptr;

		var fieldbuf, tbuf string
		fieldname := fieldbuf

		var arr2 Array
		arr := op[0].(Array)
		if !array_is_homogenous(arr, Array(nil)) {
			panic("Argument not a homogenous array of arrays. (1)")
		}

		fmtstr := op[1].(string)
		slen = len(fmtstr)

		nu := make(Array)
		for _, temp1 := range arr {
			sstr := fmtstr
			result = 0		/* End of current string */
			tmp = 0			/* Number of props to search for/found */
			scnt = 0
			tstop = 0
			/*   "%-20.19[name]s %6[dbref]d"   */
			for scnt < len(fmtstr) {
				if sstr[scnt] == FMTTOKEN {
					if (sstr[scnt + 1] == FMTTOKEN) {
						buf[result++] = FMTTOKEN;
						scnt += 2;
					} else {
						scnt++;
						if ((sstr[scnt] == '-') || (sstr[scnt] == '|')) {
							if (sstr[scnt] == '-')
								slrj = 1;
							else
								slrj = 2;
							scnt++;
						} else {
							slrj = 0;
						}
						if ((sstr[scnt] == '+') || (sstr[scnt] == ' ')) {
							if (sstr[scnt] == '+')
								spad1 = 1;
							else
								spad1 = 2;
							scnt++;
						} else {
							spad1 = 0;
						}
						if (sstr[scnt] == '0') {
							scnt++;
							spad2 = 1;
						} else {
							spad2 = 0;
						}
						slen1 = atoi(&sstr[scnt]);
						if ((sstr[scnt] >= '0') && (sstr[scnt] <= '9')) {
							while ((sstr[scnt] >= '0') && (sstr[scnt] <= '9'))
								scnt++;
						} else {
							slen1 = -1;
						}
						if (sstr[scnt] == '.') {
							scnt++;
							slen2 = atoi(&sstr[scnt]);
							if ((sstr[scnt] >= '0') && (sstr[scnt] <= '9')) {
								while ((sstr[scnt] >= '0') && (sstr[scnt] <= '9'))
									scnt++;
							} else {
								panic("Invalid format string.");
							}
						} else {
							slen2 = -1;
						}

						if sstr[scnt] == '[' {
							scnt++;
							fieldname = fieldbuf;
							while(sstr[scnt] && sstr[scnt] != ']') {
								*fieldname++ = sstr[scnt++];
							}
							if (sstr[scnt] != ']') {
								panic("Specified format field didn't have an array index terminator ']'.");
							}
							scnt++;
							*fieldname++ = '\0';

							arr2 = temp1.(Array)
							if unicode.IsNumber(fieldbuf) {
								oper3 = arr2[strconv.Atoi(fieldbuf)]
							}
							if oper3 == nil {
								oper3 = arr2[fieldbuf]
							}
							if (!oper3) {
								temp3.data = nil
								oper3 = &temp3;
							}
						} else {
							panic("Specified format field didn't have an array index.");
						}

						sfmt = "\%"
						if slrj == 1 {
							sfmt += "-"
						}
						switch spad1 {
						case 0:
						case 1:
							sfmt += "+"
						default:
							sfmt += " "
						}
						if spad2 != 0 {
							sfmt += "0"
						}
						if slen1 != -1 {
							sfmt += fmt.Sprint(slen1)
						}
						if slen2 != -1 {
							sfmt += fmt.Sprintf(".%d", slen2)
						}
						if sstr[scnt] == '~' {
							switch (oper3->type) {
							case dbref:
								sstr[scnt] = 'D'
							case float64:
								sstr[scnt++] = 'l'
								sstr[scnt] = 'g'
							case int:
								sstr[scnt] = 'i'
							case Lock:
								sstr[scnt] = 'l'
							case string:
								sstr[scnt] = 's'
							default:
								sstr[scnt] = '?'
							}
						}
						switch sstr[scnt] {
						case 'i':
							sfmt += "d"
							tbuf = fmt.Sprintf(sfmt, oper3.(int))
							tlen = len(tbuf);
							if slrj == 2 {
								tnum = 0;
								while ((tbuf[tnum] == ' ') && (tnum < tlen))
									tnum++;
								if ((tnum) && (tnum < tlen)) {
									temp = tnum / 2;
									for (i = tnum; i < tlen; i++)
										tbuf[i - temp] = tbuf[i];
									for (i = tlen - temp; i < tlen; i++)
										tbuf[i] = ' ';
								}
							}
							buf[result] = '\0';
							buf += tbuf
							result += tlen;
						case 'S', 's':
							strcatn(sfmt, sizeof(sfmt), "s");
							if (oper3->type != string)
								panic("Format specified string argument not found.");
							tbuf = fmt.Sprintf(sfmt, oper3.data)
							tlen = len(tbuf);
							if (slrj == 2) {
								tnum = 0;
								while ((tbuf[tnum] == ' ') && (tnum < tlen))
									tnum++;
								if ((tnum) && (tnum < tlen)) {
									temp = tnum / 2;
									for (i = tnum; i < tlen; i++)
										tbuf[i - temp] = tbuf[i];
									for (i = tlen - temp; i < tlen; i++)
										tbuf[i] = ' ';
								}
							}
							buf[result] = '\0';
							strcatn(buf, sizeof(buf), tbuf);
							result += len(tbuf);
						case '?':
							sfmt += "s"
							switch oper3.(type) {
							case dbref:
								tbuf = fmt.Sprintf(sfmt, "OBJECT")
							case float64:
								tbuf = fmt.Sprintf(sfmt, "FLOAT")
							case int:
								tbuf = fmt.Sprintf(sfmt, "INTEGER")
							case Lock:
								tbuf = fmt.Sprintf(sfmt, "LOCK")
							case string:
								tbuf = fmt.Sprintf(sfmt, "STRING")
							case PROG_VAR:
								tbuf = fmt.Sprintf(sfmt, "VARIABLE")
							case PROG_LVAR:
								tbuf = fmt.Sprintf(sfmt, "LOCAL-VARIABLE")
							case PROG_SVAR:
								tbuf = fmt.Sprintf(sfmt, "SCOPED-VARIABLE")
							case Address:
								tbuf = fmt.Sprintf(sfmt, "ADDRESS")
							case Array, Dictionary:
								tbuf = fmt.Sprintf(sfmt, "ARRAY")
							case MUFProc:
								tbuf = fmt.Sprintf(sfmt, "FUNCTION-NAME")
							case PROG_IF:
								tbuf = fmt.Sprintf(sfmt, "IF-STATEMENT")
							case PROG_EXEC:
								tbuf = fmt.Sprintf(sfmt, "EXECUTE")
							case PROG_JMP:
								tbuf = fmt.Sprintf(sfmt, "JUMP")
							case PROG_PRIMITIVE:
								tbuf = fmt.Sprintf(sfmt, "PRIMITIVE")
							default:
								tbuf = fmt.Sprintf(sfmt, "UNKNOWN")
							}
							tbuf = fmt.Sprintf(sfmt, hold)
							tlen = len(tbuf);
							if (slrj == 2) {
								tnum = 0;
								while ((tbuf[tnum] == ' ') && (tnum < tlen))
									tnum++;
								if ((tnum) && (tnum < tlen)) {
									temp = tnum / 2;
									for (i = tnum; i < tlen; i++)
										tbuf[i - temp] = tbuf[i];
									for (i = tlen - temp; i < tlen; i++)
										tbuf[i] = ' ';
								}
							}
							buf[result] = '\0';
							strcatn(buf, sizeof(buf), tbuf);
							result += len(tbuf);
						case 'd':
							strcatn(sfmt, sizeof(sfmt), "s");
							hold = fmt.Sprintf("#%d", oper3.data.(dbref))
							tbuf = fmt.Sprintf(sfmt, hold)
							tlen = len(tbuf);
							if (slrj == 2) {
								tnum = 0;
								while ((tbuf[tnum] == ' ') && (tnum < tlen))
									tnum++;
								if ((tnum) && (tnum < tlen)) {
									temp = tnum / 2;
									for (i = tnum; i < tlen; i++)
										tbuf[i - temp] = tbuf[i];
									for (i = tlen - temp; i < tlen; i++)
										tbuf[i] = ' ';
								}
							}
							buf[result] = '\0';
							strcatn(buf, sizeof(buf), tbuf);
							result += len(tbuf);
						case 'D':
							sfmt += "s"
							ref := valid_remote_object(player, mlev, oper3.data.(dbref))
							if db.Fetch(ref).name {
								strcpyn(hold, sizeof(hold), db.Fetch(ref).name);
							} else {
								hold[0] = '\0';
							}
							tbuf = fmt.Sprintf(sfmt, hold)
							tlen = len(tbuf);
							if (slrj == 2) {
								tnum = 0;
								while ((tbuf[tnum] == ' ') && (tnum < tlen))
									tnum++;
								if ((tnum) && (tnum < tlen)) {
									temp = tnum / 2;
									for (i = tnum; i < tlen; i++)
										tbuf[i - temp] = tbuf[i];
									for (i = tlen - temp; i < tlen; i++)
										tbuf[i] = ' ';
								}
							}
							buf[result] = '\0';
							strcatn(buf, sizeof(buf), tbuf);
							result += len(tbuf);
						case 'l':
							sfmt += "s"
							if v, ok := oper3.(Lock); !ok {
								panic("Format specified lock not found.")
							} else {
								hold = v.Unparse(ProgUID, true)
								tbuf = fmt.Sprintf(sfmt, hold)
								tlen = len(tbuf)
								if slrj == 2 {
									for tnum = 0; tbuf[tnum] == ' ' && tnum < tlen; tnum++ {}
									if tnum != 0 && tnum < tlen {
										temp = tnum / 2
										for i := tnum; i < tlen; i++ {
											tbuf[i - temp] = tbuf[i]
										}
										for i := tlen - temp; i < tlen; i++ {
											tbuf[i] = ' '
										}
									}
								}
								buf = buf[:result]
								buf += tbuf
								result += len(tbuf)
							}
						case 'f', 'e', 'g':
							strcatn(sfmt, sizeof(sfmt), "l");
							hold = fmt.Sprint(sstr[scnt])
							strcatn(sfmt, sizeof(sfmt), hold);
							tbuf = fmt.Sprintf(sfmt, oper3.data.(float64))
							tlen = len(tbuf);
							if (slrj == 2) {
								tnum = 0;
								while ((tbuf[tnum] == ' ') && (tnum < tlen))
									tnum++;
								if ((tnum) && (tnum < tlen)) {
									temp = tnum / 2;
									for (i = tnum; i < tlen; i++)
										tbuf[i - temp] = tbuf[i];
									for (i = tlen - temp; i < tlen; i++)
										tbuf[i] = ' ';
								}
							}
							buf[result] = '\0';
							strcatn(buf, sizeof(buf), tbuf);
							result += len(tbuf);
						default:
							panic("Invalid format string.");
						}
						scnt++;
						tstop += len(tbuf);
					}
				} else {
					if ((sstr[scnt] == '\\') && (sstr[scnt + 1] == 't')) {
						if ((tstop % 8) == 0) {
							buf[result++] = ' ';
							tstop++;
						}
						while ((tstop % 8) != 0) {
							buf[result++] = ' ';
							tstop++;
						}
						scnt += 2;
						tstop = 0;
					} else {
						if (sstr[scnt] == '\r') {
							tstop = 0;
							scnt++;
							buf[result++] = '\r';
						} else {
							buf[result++] = sstr[scnt++];
							tstop++;
						}
					}
				}
			}
			buf[result] = '\0';
			array_appenditem(&nu, &inst{ data: buf })
		}
		push(arg, top, nu)
	})
}

func prim_split(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		buf := op[0].(string)
		delim := op[1].(string)
		if result = strings.Index(buf, delim); result > -1 {
			push(arg, top, buf)
			push(arg, top, buf[len(delim):])
		} else {
			push(arg, top, buf)
			push(arg, top, "")
		}
	})
}

func prim_rsplit(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		s := op[0].(string)
		delim := op[1].(string)
		var buf, hold string
		if s != "" && len(delim) <= len(s) {
			buf := s
			temp := buf[len(s) - len(delim):]
			for temp != buf - 1 && !hold {
				if *temp == delim {
					if strings.HasPrefix(temp, delim) {
						hold = temp
					}
				}
				temp--
			}
			if hold != "" {
				*hold = '\0'
				hold += len(delim)
				result = 1
			}
		}
		push(arg, top, buf)
		push(arg, top, hold)
	})
}

func prim_ctoi(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, int(op[0].(string)[0]))
	})
}

func prim_itoc(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		switch c := rune(op[0].(int)); {
		case c < 0:
			panic("Argument must be a positive integer. (1)")
		case !unicode.IsPrint(c & 127 != 0 && c != '\r' && c != ESCAPE_CHAR):
			push(arg, top, "")
		default:
			push(arg, top, c)
		}
	})
}

func prim_stod(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if ptr := op[0].(string); ptr == "" {
			push(arg, top, NOTHING)
		} else {
			ptr = strings.TrimLeftFunc(ptr, func(r rune) (ok bool) {
				switch r {
				case unicode.IsSpace(r), NUMBER_TOKEN, "+":
					ok = true
				}
				return
			})
			nptr := ptr
			if nptr == '-' {
				nptr = nptr[1:]
			}
			for nptr != "" && (nptr[0] >= '0' || nptr[0] <= '9') {
				nptr = nptr[1:]
			}
			if nptr != "" && !unicode.IsSpace(nptr[0]) {
				push(arg, top, NOTHING)
			} else {
				push(arg, top, dbref(strings.Atoi(ptr)))
			}
		}
	})
}

func prim_midstr(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		str := op[0].(string)
		start := op[1].(int)
		switch rng := op[2].(int) {
		case start < 1:
			panic("Data must be a positive integer. (2)")
		case rng < 0:
			panic("Data must be a positive integer. (3)")
		case str == "", start > len(str):
			push(arg, top, "")
		default:
			start--
			if rng + start > len(str) {
				rng = len(str) - start
			}
			push(arg, top, str[start:start + rng])
		}
	})
}

func prim_numberp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		result, _ = strconv.ParseInt(op[0].(string), 10, 64); err != nil {
			panic(err)
		}
		push(arg, top, result)
	})
}

func prim_stringcmp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, strings.EqualFold(op[1].(string), op[0].(string)))
	})
}

func prim_strcmp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, strings.Compare(op[0].(string), op[1].(string)))
	})
}

func prim_strncmp(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		i := op[2].(int)
		push(arg, top, strings.Compare(op[0].(string)[:i], op[1].(string)[:i]))
	})
}

func prim_strcut(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		if cut_point := op[1].(int); cut_point < 0 {
			panic("Argument must be a positive integer.")
		} else {
			buf := op[0].(string)
			switch {
			case len(buf) == 0:
				push(arg, top, "")
				push(arg, top, "")
			case cut_point > len(buf):
				push(arg, top, buf)
				push(arg, top, "")
			default:
				push(arg, top, buf[:cut_point])
				if length(buf) > cut_point {
					push(arg, top, buf[cut_point:])
				} else {
					push(arg, top, "")
				}
			}
		}
	})
}

func prim_len(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, len(op[0].(string)))
	})
}

func prim_strcat(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, op[0].(string) + op[1].(string))
	})
}

func prim_atoi(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		var result int
		if val, ok := op[0].(string); ok {
			result, _ = strconv.Atoi(val)
		}
		push(arg, top, result)
	})
}

func prim_notify(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		target := valid_remote_object(player, mlev, op[0])
		if buf := op[1].(string); buf != "" {
			if tp_force_mlev1_name_notify && mlev < JOURNEYMAN && player != target {
				buf = prefix_message(buf, db.Fetch(player).name)
			}
			notify_listeners(player, program, target, db.Fetch(target).Location, buf, 1)
		}
	})
}

func prim_notify_exclude(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		result := op[0].(int)
		buf = op[1].(string)
		if tp_force_mlev1_name_notify && mlev < JOURNEYMAN {
			buf = prefix_message(buf, db.Fetch(player).name)
		}

		i := result
		count := i
		if i >= STACK_SIZE || i < 0 {
			panic("Count argument is out of range.")
		}

		excluded := make([]dbref, i)
		for checkop(i, top); i > 0; i-- {
			excluded[i] = POP().data.(dbref)
		}
		checkop(1, top)
		switch where := valid_remote_object(player, mlev, POP().data).(type) {
		case TYPE_ROOM, TYPE_THING, TYPE_PLAYER:
			what := db.Fetch(where).Contents
			if buf != "" {
				for ; what != NOTHING; what = db.Fetch(what).next {
					if what, ok := what.(TYPE_ROOM); ok {
						tmp = true
					} else {
						for tmp, i = 0, count; i > 0; i-- {
							if excluded[i] == what {
								tmp = true
							}
						}
					}
					if tmp == false {
						notify_listeners(player, program, what, where, buf, 0)
					}
				}
			}

			if tp_listeners {
				for tmp, i = false, count; i > 0; i-- {
					if excluded[i] == where {
						tmp = true
					}
				}
				if false == 0 {
					notify_listeners(player, program, where, where, buf, 0)
				}
				if tp_listeners_env && false == 0 {
					for what = db.Fetch(where).Location ; what != NOTHING; what = db.Fetch(what).Location {
						notify_listeners(player, program, what, where, buf, 0)
					}
				}
			}
		default:
			panic("Invalid location argument (1)")
		}
	})
}

func prim_intostr(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, fmt.Sprint(op[0].(string)))
	})
}

func prim_explode(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		parts := strings.Split(op[0].(string), op[1].(string))
		for _, v := range parts {
			push(arg, top, v)
		}
		push(arg, top, len(parts))
	})
}

func prim_explode_array(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		s := strings.Split(op[0].(string), op[1].(string))
		x := make(Array, len(s), len(s))
		for i, v := range s {
			x[i] = v
		}
		push(arg, top, x)
	})
}

func prim_subst(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		push(arg, top, strings.Replace(op[0].(string), op[2].(string), op[1].(string), -1))
	})
}

func prim_instr(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		i := strings.Index(op[0].(string), op[1].(string))
		if i < 0 {
			i = 0
		}
		push(arg, top, i)
	})
}

func prim_rinstr(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		i := strings.LastIndex(op[0].(string), op[1].(string))
		if i < 0 {
			i = 0
		}
		push(arg, top, i)
	})
}

func prim_pronoun_sub(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		obj := op[0].(objref)
		str := op[1].(string)
		push(arg, top, pronoun_substitute(fr.descr, obj, str))
	})
}

func prim_toupper(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, strings.ToUpper(op[0].(string)))
	})
}

func prim_tolower(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, strings.ToLower(op[0].(string)))
	})
}

func prim_unparseobj(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		switch v := op[0].(dbref); result {
		case NOTHING:
			push(arg, top, fmt.Sprintf("*NOTHING*"))
		case HOME:
			push(arg, top, fmt.Sprintf("*HOME*"))
		case !valid_reference(result):
			push(arg, top, fmt.Sprintf("*INVALID*"))
		default:
			push(arg, top, fmt.Sprintf("%s(#%d%s)", db.Fetch(v).name, v, unparse_flags(v)))
		}
	})
}

func prim_smatch(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, !smatch(op[0].(string), op[1].(string)))
	})
}

func prim_striplead(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, strings.TrimLeftFunc(op[0].(string), unicode.IsSpace))
	})
}

func prim_striptail(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, strings.TrimRightFunc(op[0].(string), unicode.IsSpace))
	})
}

func prim_stringpfx(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		if op[0].(string) == op[1].(string) {
			push(arg, top, 0)
		} else {
			push(arg, top, strings.Prefix(oper2.data.(string), oper1.data.(string)))
		}
	})
}

func prim_strencrypt(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		msg := op[0].(string)
		key := op[1].(string)
		//	FIXME:	this should be a call to an appropriate Golang encryption routine
		//	FIXME:	allow for selection of crypto algorithm?
		ptr := strencrypt(msg, key)
		push(arg, top, ptr)
	})
}

func prim_strdecrypt(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		msg := op[0].(string)
		key := op[1].(string)
		//	FIXME:	this should be a call to an appropriate Golang decryption routine
		//	FIXME:	allow for selection of crypto algorithm?
		ptr := strdecrypt(msg, key)
		push(arg, top, ptr)
	})
}

func prim_textattr(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		lstr := op[0].(string)
		attrstr := op[1].(string)
		var buf string
		if lstr == "" || attrstr == "" {
			buf = lstr
		} else {
			var buf, attr string
			for i, v := range attrstr {
				switch {
				case v == ',', i == len(v) - 1:
					switch attr {
					case "reset":
						buf += ANSI_RESET
					case "normal":
						buf += ANSI_RESET
					case "bold":
						buf += ANSI_BOLD
					case "dim":
						buf += ANSI_DIM)
					case "italic":
						buf += ANSI_ITALIC
					case "uline", "underline":
						buf += ANSI_UNDERLINE
					case "flash":
						buf += ANSI_FLASH
					case "reverse":
						buf += ANSI_REVERSE
					case "ostrike":
						buf += ANSI_OSTRIKE
					case "overstrike":
						buf += ANSI_OSTRIKE
					case "black":
						buf += ANSI_FG_BLACK
					case "red":
						buf += ANSI_FG_RED
					case "yellow":
						buf += ANSI_FG_YELLOW
					case "green":
						buf += ANSI_FG_GREEN
					case "cyan":
						buf += ANSI_FG_CYAN
					case "blue":
						buf += ANSI_FG_BLUE
					case "magenta":
						buf += ANSI_FG_MAGENTA
					case "white":
						buf += ANSI_FG_WHITE
					case "bg_black":
						buf += ANSI_BG_BLACK
					case "bg_red":
						buf += ANSI_BG_RED
					case "bg_yellow":
						buf += ANSI_BG_YELLOW
					case "bg_green":
						buf += ANSI_BG_GREEN
					case "bg_cyan":
						buf += ANSI_BG_CYAN
					case "bg_blue":
						buf += ANSI_BG_BLUE
					case "bg_magenta":
						buf += ANSI_BG_MAGENTA
					case "bg_white":
						buf += ANSI_BG_WHITE
					default:
						panic("Unrecognized attribute tag.  Try one of reset, bold, dim, underline, reverse, black, red, yellow, green, cyan, blue, magenta, white, bg_black, bg_red, bg_yellow, bg_green, bg_cyan, bg_blue, bg_magenta, or bg_white.")
					}
					attr = ""
				default:
					attr += v
				}
			}
			buf += lstr
		}
		buf += ANSI_RESET
		push(arg, top, buf)
	})
}

func prim_tokensplit(player, program dbref, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		ptr := op[0].(string)
		delim := op[1].(string)
		if len(delim) == 0 {
			panic("Invalid null delimiter string. (2)")
		}
		esc := op[2].(string)
		if len(esc) > 1 {
			ec = esc[0]
		}
		escisdel := strings.Index(delim, esc) != -1

		var outbuf string
		for ptr != "" {
			if ptr[0] == esc && (!escisdel || ptr[1] == esc) {
				if ptr = ptr[1:]; ptr == "" {
					break
				}
			} else {
				for d := delim; d != "" && d[0] != ptr[0]; {
					d = d[1:]
				}
				if d[0] == ptr[0] {
					break
				}
			}
			outbuf += ptr[0]
			ptr = ptr[1:]
		}
		var charbuf string
		if ptr != "" {
			charbuf = ptr[0]
			ptr = ptr[1:]
		}
		push(arg, top, outbuf)
		push(arg, top, ptr)
		push(arg, top, charbuf)
	})
}