package fbmuck

func prim_inf(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, math.Inf(1))
	})
}

func prim_ceil(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if v := op[0].(float64); math.IsInf(v, 0) {
			fr.error.error_flags.f_bounds = true
			push(arg, top, v)
		} else {
			push(arg, top, ceil(v))
		}
	})
}

func prim_floor(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if v := op[0].(float64); math.IsInf(v, 0) {
			fr.error.error_flags.f_bounds = true
			fresult = v
		} else {
			fresult = floor(v);
		}
	})
}

func prim_sqrt(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		switch op[0].(float64); {
		case math.IsInf(v, 0):
			fr.error.error_flags.f_bounds = true
			push(arg, top, v)
		case v < 0.0:
			fr.error.error_flags.imaginary = true
			push(arg, top, 0.0)
		default:
			push(arg, top, sqrt(v))
		}
	})
}

func prim_pi(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, math.Pi)
	})
}

func prim_epsilon(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, math.Nextafter(1, 2) - 1)
	})
}

func prim_round(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		if v := op[0].(float64); math.IsInf(v, 0) {
			fr.error.error_flags.f_bounds = true
			push(arg, top, 0.0)
		} else {
			temp := pow(10.0, op[1].(int))
			var r float64
			switch tnum := modf(temp * v, &r); {
			case tnum >= 0.5:
				r = r + 1.0
			case tnum <= -0.5:
				r = r - 1.0
			}
			push(arg, top, r / temp)
		}
	})
}

func prim_sin(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, math.Sin(op[0].(float64)))
	})
}

func prim_cos(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, math.Cos(op[0].(float64)))
	})
}

func prim_tan(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, math.Tan(op[0].(float64)))
	})
}

func prim_asin(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, math.Asin(op[0].(float64)))
	})
}

func prim_acos(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, math.Acos(op[0].(float64)))
	})
}

func prim_atan(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, math.Atan(op[0].(float64)))
	})
}

func prim_atan2(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, math.Atan2(op[0].(float64), op[1].(float64)))
	})
}

func prim_dist3d(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		x := op[0].(float64)
		y := op[1].(float64)
		z := op[2].(float64)
		push(arg, top, math.Sqrt((x * x) + (y * y) + (z * z)))
	})
}

func prim_diff3(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(6, top, func(op Array) {
		push(arg, top, op[3].(float64) - op[0].(float64))
		push(arg, top, op[4].(float64) - op[1].(float64))
		push(arg, top, op[5].(float64) - op[2].(float64))
	})
}

func prim_xyz_to_polar(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		x := op[0].(float64)
		y := op[1].(float64)
		z := op[2].(float64)
		switch {
		case math.IsInf(x, 0), math.IsInf(y, 0), math.IsInf(z, 0):
			fr.error.error_flags.nan = true
			push(arg, top, 0.0)
			push(arg, top, 0.0)
			push(arg, top, 0.0)
		default:
			if dist := math.Sqrt((x * x) + (y * y) + (z * z)); dist > 0.0 {
				push(arg, top, dist)
				push(arg, top, atan2(y, x))
				push(arg, top, acos(z / dist))
			} else {
				push(arg, top, dist)
				push(arg, top, 0.0)
				push(arg, top, 0.0)
			}
		}
	})
}

func prim_polar_to_xyz(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(3, top, func(op Array) {
		dist := op[0].(float64)
		theta := op[1].(float64)
		phi := op[2].(float64)
		switch {
		case math.IsInf(dist, 0), math.IsInf(theta, 0), math.IsInf(phi, 0):
			fr.error.error_flags.nan = true
			push(arg, top, 0.0)
			push(arg, top, 0.0)
			push(arg, top, 0.0)
		default:
			push(arg, top, dist * cos(theta) * sin(phi))
			push(arg, top, dist * sin(theta) * sin(phi))
			push(arg, top, dist * cos(phi))
		}
	})
}

func prim_exp(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, math.Exp(op[0].(float64)))
	})
}

func prim_log(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, math.Log(op[0].(float64)))
	})
}

func prim_log10(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, math.Log10(op[0].(float64)))
	})
}

func prim_fabs(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, math.Abs(op[0].(float64)))
	})
}

func prim_float(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		push(arg, top, float64(op[0].(int)))
	})
}

func prim_pow(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, math.Pow(op[0].(float64), op[1].(float64)))
	})
}

func prim_frand(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(0, top, func(op Array) {
		CHECKOFLOW(1)
		push(arg, top, rand.Float64())
	})
}

/* We use these two statics to prevent lost work. */
var gaussian_r float64
var gaussian_second_call bool

func prim_gaussian(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		var r float64
		/* This is a Box-Muller polar conversion.
		 * Taken in part from code as demonstrated by Everett F. Carter, Jr.
		 * This code is not copyrighted. */
		if gaussian_second_call  {
			/* We should have a correlated value to use from the previous call, still. */
			r = gaussian_r
			gaussian_second_call = false
		} else {
			var radius, srca, srcb float64
			for radius = 1.0; radius >= 1.0; radius = srca * srca + srcb * srcb {
				srca = 2.0 * rand.Float64() - 1.0
				srcb = 2.0 * rand.Float64() - 1.0
			}

			radius = sqrt( (-2.0 * log(radius) ) / radius );
			r = srca * radius
			gaussian_r = srcb * radius
			gaussian_second_call = true
		}
		push(arg, top, op[1].(float64) + r * op[0].(float64))
	})
}

func prim_fmod(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(2, top, func(op Array) {
		push(arg, top, math.Mod(op[0].(float64), op[1].(float64)))
	})
}

func prim_modf(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		d, r := math.Modf(op[0].(float64))
		CHECKOFLOW(2)
		push(arg, top, d)
		push(arg, top, r)
	})
}

func prim_strtof(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		if v, e := strconv.ParseFloat(op[0].string, 64); e == nil {
			fr.error.error_flags.nan = true
				push(arg, top, 0.0)
		} else {
			push(arg, top, v)
		}
	})
}

func prim_ftostr(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		switch v := op[0].(type) {
		case int:
			push(arg, top, fmt.Sprintf("%#.15g", float64(v))
		case float64:
			push(arg, top, fmt.Sprintf("%#.15g", v)
		default:
			panic("Non-float argument. (1)");
		})
	})
}

func prim_ftostrc(player, program ObjectID, mlev int, pc, arg *inst, top *int, fr *frame) {
	apply_primitive(1, top, func(op Array) {
		switch v := op[0].(type) {
		case int:
			buf := fmt.Sprintf("%.15g", float64(v))
			if !strchr(buf, '.') && !strchr(buf, 'e') && !strchr(buf, 'n') {
				buf += ".0"
			}
			push(arg, top, buf)
		case float64:
			buf := fmt.Sprintf("%.15g", v)
			if !strchr(buf, '.') && !strchr(buf, 'e') && !strchr(buf, 'n') {
				buf += ".0"
			}
			push(arg, top, buf)
		default:
			panic("Non-float argument. (1)");
		}
	})
}