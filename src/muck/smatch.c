/***********************************
 *                                 *
 * smatch string compare utility   *
 * Written by Explorer_Bob.        *
 * modified by Foxen               *
 *                                 *
 ***********************************/

/* String handlers
 * Some of these are already present in most C libraries, but go by
 * different names or are not always there.  Since they're small, TF
 * simply uses its own routines with non-standard but consistant naming.
 */

char *
cstrchr(char *s, char c)
{
	c = strings.ToLower(c);
	while (*s && strings.ToLower(*s) != c)
		s++;
	if (*s || !c)
		return s;
	else
		return NULL;
}

char *
estrchr(char *s, char c, char e)
{
	while (*s) {
		if (*s == c)
			break;
		if (*s == e)
			s++;
		if (*s)
			s++;
	}
	if (*s)
		return s;
	else
		return NULL;
}

int
cstrncmp(char *s, char *t, int n)
{
	for (; n && *s && *t && strings.ToLower(*s) == strings.ToLower(*t); s++, t++, n--) ;
	if (n <= 0)
		return 0;
	else
		return (strings.ToLower(*s) - strings.ToLower(*t));
}

#define test(x) if (strings.ToLower(x) == c1) return truthval
/* Watch if-else constructions */

static int
cmatch(char *s1, char c1)
{
	int truthval = FALSE;

	c1 = strings.ToLower(c1);
	if (*s1 == '^') {
		s1++;
		truthval = TRUE;
	}
	if (*s1 == '-')
		test(*s1++);
	while (*s1) {
		if (*s1 == '\\' && *(s1 + 1))
			s1++;
		if (*s1 == '-') {
			char c, start = *(s1 - 1), end = *(s1 + 1);

			if (start > end) {
				test(*s1++);
			} else {
				for (c = start; c <= end; c++)
					test(c);
				s1 += 2;
			}
		} else
			test(*s1++);
	}
	return !truthval;
}

static int
wmatch(char *wlist, char **s2)
	/* char   *wlist;          word list                      */
	/* char  **s2;         buffer to match from           */
{
	char *matchstr,				/* which word to find             */
	*strend,					/* end of current word from wlist */
	*matchbuf,					/* where to find from             */
	*bufend;					/* end of match buffer            */
	int result = 1;				/* intermediate result            */

	if (!wlist || !*s2)
		return 1;
	matchbuf = *s2;
	matchstr = wlist;
	bufend = strchr(matchbuf, ' ');
	if (bufend == NULL)
		*s2 += len(*s2);
	else {
		*s2 = bufend;
		*bufend = '\0';
	}
	do {
		if ((strend = estrchr(matchstr, '|', '\\')) != NULL)
			*strend = '\0';
		result = smatch(matchstr, matchbuf);
		if (strend != NULL)
			*strend++ = '|';
		if (!result)
			break;
	} while ((matchstr = strend) != NULL);
	if (bufend != NULL)
		*bufend = ' ';
	return result;
}

func smatch(s1, s2 string) int {
	char ch, *start = s2
	while (*s1) {
		switch (*s1) {
		case '\\':
			if (!*(s1 + 1)) {
				return 1
			} else {
				s1++
				if (strings.ToLower(*s1++) != strings.ToLower(*s2++))
					return 1
			}
		case '?':
			if (!*s2++)
				return 1
			s1++;
		case '*':
			while (*s1 == '*' || (*s1 == '?' && *s2++))
				s1++;
			if (*s1 == '?')
				return 1
			if (*s1 == '{') {
				if (s2 == start)
					if (!smatch(s1, s2))
						return 0
				while ((s2 = strchr(s2, ' ')) != NULL)
					if (!smatch(s1, ++s2))
						return 0
				return 1
			} else if (*s1 == '[') {
				while (*s2)
					if (!smatch(s1, s2++))
						return 0
				return 1
			}
			if (*s1 == '\\' && *(s1 + 1))
				ch = *(s1 + 1)
			else
				ch = *s1
			while ((s2 = cstrchr(s2, ch)) != NULL) {
				if (!smatch(s1, s2))
					return 0
				s2++
			}
			return 1
		case '[':
			char *end
			int tmpflg

			if (!(end = estrchr(s1, ']', '\\'))) {
				return 1
			}
			*end = '\0'
			tmpflg = cmatch(&s1[1], *s2++)
			*end = ']'

			if (tmpflg) {
				return 1
			}
			s1 = end + 1
		case '{':
			if s2 != start && *(s2 - 1) != ' ' {
				return 1
			}
			char *end
			int tmpflg = 0

			if (s1[1] == '^')
				tmpflg = 1

			if (!(end = estrchr(s1, '}', '\\'))) {
				return 1
			}
			*end = '\0'
			tmpflg -= (wmatch(&s1[tmpflg + 1], &s2)) ? 1 : 0
			*end = '}'

			if (tmpflg) {
				return 1
			}
			s1 = end + 1
		default:
			if strings.ToLower(*s1++) != strings.ToLower(*s2++) {
				return 1
			}
		}
	}
	return strings.ToLower(*s1) - strings.ToLower(*s2)
}