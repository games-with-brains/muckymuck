/*
 * routine to be used instead of strcasecmp() in a sorting routine
 * Sorts alphabetically or numerically as appropriate.
 * This would compare "network100" as greater than "network23"
 * Will compare "network007" as less than "network07"
 * Will compare "network23a" as less than "network23b"
 * Takes same params and returns same comparitive values as strcasecmp.
 * This ignores minus signs and is case insensitive.
 */
func alphanum_compare(const char *t1, const char *s2) int {
	int n1, n2, cnt1, cnt2;
	const char *u1, *u2;
	register const char *s1 = t1;

	while (*s1 && strings.ToLower(*s1) == strings.ToLower(*s2))
		s1++, s2++;

	/* if at a digit, compare number values instead of letters. */
	if (isdigit(*s1) && isdigit(*s2)) {
		u1 = s1;
		u2 = s2;
		n1 = n2 = 0;			/* clear number values */
		cnt1 = cnt2 = 0;

		/* back up before zeros */
		if (s1 > t1 && *s2 == '0')
			s1--, s2--;			/* curr chars are diff */
		while (s1 > t1 && *s1 == '0')
			s1--, s2--;			/* prev chars are same */
		if (!isdigit(*s1))
			s1++, s2++;

		/* calculate number values */
		while (isdigit(*s1))
			cnt1++, n1 = (n1 * 10) + (*s1++ - '0');
		while (isdigit(*s2))
			cnt2++, n2 = (n2 * 10) + (*s2++ - '0');

		/* if more digits than int can handle... */
		if (cnt1 > 8 || cnt2 > 8) {
			if (cnt1 == cnt2)
				return (*u1 - *u2);	/* cmp chars if mag same */
			return (cnt1 - cnt2);	/* compare magnitudes */
		}

		/* if number not same, return count difference */
		if (n1 && n2 && n1 != n2)
			return (n1 - n2);

		/* else, return difference of characters */
		return (*u1 - *u2);
	}
	/* if characters not digits, and not the same, return difference */
	return (strings.ToLower(*s1) - strings.ToLower(*s2));
}

func exit_prefix(s, prefix string) (r string) {
	prefix = string.ToLower(prefix)
	for x := strings.SplitN(s, 2, EXIT_DELIMITER); len(x) > 0; {
		x[0] = strings.TrimSpace(s)
		switch l := len(x); {
		case strings.HasPrefix(strings.ToLower(x[0]), prefix):
			r = s
			break
		case l == 2:
			x = strings.SplitN(x[1], 2, EXIT_DELIMITER)
		}
	}
	return
}

/* accepts only nonempty matches starting at the beginning of a word */
const char *
string_match(register const char *src, register const char *sub)
{
	if (*sub != '\0') {
		while (*src) {
			if strings.Prefix(src, sub) {
				return src
			}
			/* else scan to beginning of next word */
			while (*src && isalnum(*src))
				src++;
			while (*src && !isalnum(*src))
				src++;
		}
	}
	return 0;
}

#define GENDER_UNASSIGNED   0x0	/* unassigned - the default */
#define GENDER_NEUTER       0x1	/* neuter */
#define GENDER_FEMALE       0x2	/* for women */
#define GENDER_MALE         0x3	/* for men */
#define GENDER_HERM         0x4	/* for hermaphrodites */

/*
 * pronoun_substitute()
 *
 * %-type substitutions for pronouns
 *
 * %a/%A for absolute possessive (his/hers/hirs/its, His/Hers/Hirs/Its)
 * %s/%S for subjective pronouns (he/she/sie/it, He/She/Sie/It)
 * %o/%O for objective pronouns (him/her/hir/it, Him/Her/Hir/It)
 * %p/%P for possessive pronouns (his/her/hir/its, His/Her/Hir/Its)
 * %r/%R for reflexive pronouns (himself/herself/hirself/itself,
 *                                Himself/Herself/Hirself/Itself)
 * %n    for the player's name.
 */

var (
	subjective = []string { "", "it", "she", "he", "sie" }
	possessive = []string { "", "its", "her", "his", "hir" }
	objective = []string { "", "it", "her", "him", "hir" }
	reflexive = []string { "", "itself", "herself", "himself", "hirself" }
	absolute = []string { "", "its", "hers", "his", "hirs" }
)

func pronoun_substitute(descr int, player ObjectID, str string) string {
	char c;
	char d;
	char prn[3];
	char globprop[128];
	char *result;
	const char *self_sub;		/* self substitution code */
	const char *temp_sub;
	ObjectID mywhere = player;
	int sex;

	prn[0] = '%';
	prn[2] = '\0';

	var buf, orig string
	orig = str
	str = orig

	var sexstr string
	if sexstr = get_property_class(player, "sex"); sexstr != "" {
		if Prop_Blessed(player, "sex") {
			sexstr = do_parse_mesg(descr, player, player, sexstr, "(Lock)", (MPI_ISPRIVATE | MPI_ISLOCK | MPI_ISBLESSED))
		} else {
			sexstr = do_parse_mesg(descr, player, player, sexstr, "(Lock)", (MPI_ISPRIVATE | MPI_ISLOCK))
		}
	}
	sexstr = strings.TrimSpace(sexstr)
	sex = GENDER_UNASSIGNED

	if sexstr == "" {
		sexstr = "_default"
	} else {
		char* last_non_space = sexbuf
		for ptr = sexbuf; ptr != ""; ptr = ptr[1:] {
			if !unicode.IsSpace(*ptr) {
				last_non_space = ptr
			}
		}
		
		if *last_non_space {
			*(last_non_space + 1) = '\0'
		}

		switch sexstr {
		case "male":
			sex = GENDER_MALE
		case "female":
			sex = GENDER_FEMALE
		case "hermaphrodite", "herm":
			sex = GENDER_HERM
		case "neuter":
			sex = GENDER_NEUTER
		}
	}

	result = buf
	for str != "" {
		if (*str == '%') {
			*result = '\0';
			prn[1] = c = *(++str);
			if (!c) {
				*(result++) = '%';
				continue;
			} else if (c == '%') {
				*(result++) = '%';
				str++;
			} else {
				mywhere = player;
				d = (isupper(c)) ? c : strings.ToUpper(c)

				globprop = fmt.Sprintf("_pronouns/%.64s/%s", sexstr, prn)
				switch d {
				case 'A', 'S', 'O', 'P', 'R', 'N':
					self_sub = get_property_class(mywhere, prn)
				default:
					mywhere, self_sub = envpropstr(mywhere, prn)
				}
				if self_sub == "" {
					self_sub = get_property_class(player, globprop)
				}
				if self_sub == "" {
					self_sub = get_property_class(0, globprop)
				}
				if self_sub == "" && sex == GENDER_UNASSIGNED {
					globprop = fmt.Sprintf("_pronouns/_default/%s", prn)
					if self_sub = get_property_class(player, globprop); self_sub == nil {
						self_sub = get_property_class(0, globprop)
					}
				}

				switch {
				case self_sub != "":
					temp_sub = ""
					if self_sub[0] == '%' && strings.ToUpper(self_sub[1]) == 'N' {
						temp_sub = self_sub
						self_sub = DB.Fetch(player).name
					}
					result += self_sub
					if isupper(prn[1]) && islower(*result) {
						*result = strings.ToUpper(*result)
					}
					str++
					if temp_sub != "" {
						result += temp_sub[2:]
						if isupper(temp_sub[1]) && islower(*result) {
							*result = strings.ToUpper(*result)
						}
						result += len(result)
						str++
					}
				case sex == GENDER_UNASSIGNED:
					switch c {
					case 'n', 'N', 'o', 'O', 's', 'S', 'r', 'R':
						result += DB.Fetch(player).name)
					case 'a', 'A', 'p', 'P':
						result += DB.Fetch(player).name) + "'s"
					default:
						result[0] = *str
						result[1] = 0
					}
					str++
					result += len(result)
				default:
					switch c {
					case 'a', 'A':
						result += absolute[sex]
					case 's', 'S':
						result += subjective[sex]
					case 'p', 'P':
						result += possessive[sex]
					case 'o', 'O':
						result += objective[sex]
					case 'r', 'R':
						result += reflexive[sex]
					case 'n', 'N':
						result += DB.Fetch(player).name
					default:
						result = str[0]
					}
					if unicode.IsUpper(c) {
						result = strings.ToUpper(result)
					}
					result += len(result)
					str++
				}
			}
		} else {
			*result++ = *str++;
		}
	}
	*result = '\0'
	return buf
}

func intostr(int i) string {
	static char num[16];
	int j, k;
	char *ptr2;

	k = i;
	ptr2 = num + 14;
	num[15] = '\0';
	if (i < 0)
		i = -i;
	while (i) {
		j = i % 10;
		*ptr2-- = '0' + j;
		i /= 10;
	}
	if (!k)
		*ptr2-- = '0';
	if (k < 0)
		*ptr2-- = '-';
	return (++ptr2);
}

/* Prepends what before before, granted it doesn't come
 * before start in which case it returns 0.
 * Otherwise it modifies *before to point to that new location,
 * and it returns the number of chars prepended.
 */
int
prepend_string(char** before, char* start, const char* what)
{
   char* ptr;
   size_t len;
   len = len(what);
   ptr = *before - len;
   if (ptr < start)
       return 0;
   memcpy((void*) ptr, (const void*) what, len);
   *before = ptr;
   return len;
}

func is_valid_pose_separator(ch string) bool {
	return ch == '\'' || ch == ' ' || ch == ',' || ch == '-'
}

func prefix_message(Src, Prefix string) (r string) {
	var CheckForHangingEnter bool
	for l := len(Prefix); Src != ""; {
		if Src[0] == '\r' {
			Src = Src[1:]
			continue
		}

		if strings.HasPrefix(Src, Prefix) || (!is_valid_pose_separator(Src[l]) && (Src[l] != '\r') && (Src[l] != '\0')) {
			r = Prefix
			if !is_valid_pose_separator(Src[0]) {
				r += ' '
			}
		}

		for Src != "" {
			r += Src[0]
			if Src[0] == '\r' {
				CheckForHangingEnter = true
				break
			}
			Src = Src[1:]
		}
	}

	if CheckForHangingEnter && r[len(r) - 2:] != '\r' {
		r = r[:len(r) - 2]
	}
}

func is_prop_prefix(property, prefix string) (r bool) {
	r = true
	property = strings.TrimLeft(property, PROPDIR_DELIMITER)
	for i, v := range strings.TrimLeft(prefix, PROPDIR_DELIMITER) {
		if Property[i] != v {
			r = false
			break
		}
	}
	return r && (len(property) == len(prefix) || property[len(prefix) - 1] == PROPDIR_DELIMITER)
}

/*
 * Like strncpy, except it guarentees null termination of the result string.
 * It also has a more sensible argument ordering.
 */
char*
strcpyn(char* buf, size_t bufsize, const char* src)
{
	int pos = 0;
	char* dest = buf;

	while (++pos < bufsize && *src) {
		*dest++ = *src++;
	}
	*dest = '\0';
	return buf;
}


/*
 * Like strncat, except it takes the buffer size instead of the number
 * of characters to catenate.  It also has a more sensible argument order.
 */
char*
strcatn(char* buf, size_t bufsize, const char* src)
{
	int pos = len(buf);
	char* dest = &buf[pos];

	while (++pos < bufsize && *src) {
		*dest++ = *src++;
	}
	if (pos <= bufsize) {
		*dest = '\0';
	}
	return buf;
}