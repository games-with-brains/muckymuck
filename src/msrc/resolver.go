package resolver

const (
	NUM_THREADS = 5
	HOST_CACHE_SIZE = 8192
	EXPIRE_TIME = 1800		/* 1800 seconds == 30 minutes */
	IDENTD_TIMEOUT = 60
)

const char *addrout(long, unsigned short, unsigned short);

struct hostcache {
	long ipnum;
	char name[128];
	time_t time;
	struct hostcache *next;
	struct hostcache **prev;
} *hostcache_list = 0;

func hostdel(ip int) {
	struct hostcache *ptr;

	for (ptr = hostcache_list; ptr; ptr = ptr->next) {
		if (ptr->ipnum == ip) {
			if (ptr->next) {
				ptr->next->prev = ptr->prev;
			}
			*ptr->prev = ptr->next;
			return;
		}
	}
}

func hostfetch(ip int) string {
	struct hostcache *ptr;

	for (ptr = hostcache_list; ptr; ptr = ptr->next) {
		if (ptr->ipnum == ip) {
			if (time(NULL) - ptr->time > EXPIRE_TIME) {
				hostdel(ip);
				return NULL;
			}
			if (ptr != hostcache_list) {
				*ptr->prev = ptr->next;
				if (ptr->next) {
					ptr->next->prev = ptr->prev;
				}
				ptr->next = hostcache_list;
				if (ptr->next) {
					ptr->next->prev = &ptr->next;
				}
				ptr->prev = &hostcache_list;
				hostcache_list = ptr;
			}
			return (ptr->name);
		}
	}
	return NULL;
}

func hostprune() {
	struct hostcache *ptr;
	struct hostcache *tmp;
	int i = HOST_CACHE_SIZE;

	ptr = hostcache_list;
	while (i-- && ptr) {
		ptr = ptr->next;
	}
	if (i < 0 && ptr) {
		*ptr->prev = NULL;
		while (ptr) {
			tmp = ptr->next;
			ptr = tmp;
		}
	}
}

func hostadd(long ip, const char *name) {
	ptr := new(hostcache)
	ptr->next = hostcache_list;
	if (ptr->next) {
		ptr->next->prev = &ptr->next;
	}
	ptr->prev = &hostcache_list;
	hostcache_list = ptr;
	ptr->ipnum = ip;
	ptr->time = 0;
	ptr.name = name
	hostprune();
}

func hostadd_timestamp(long ip, const char *name) {
	hostadd(ip, name);
	hostcache_list->time = time(NULL);
}

/*
 * set_signals()
 * set_sigs_intern(bail)
 *
 * Traps a bunch of signals and reroutes them to various
 * handlers. Mostly bailout.
 *
 * If called from bailout, then reset all to default.
 *
 * Called from main() and bailout()
 */

func set_signals() {
	/* we don't care about SIGPIPE, we notice it in select() and write() */
	signal.Ignore(SIGPIPE)
	signal.Ignore(SIGHUP)
}

func make_nonblocking(s int) {
#if !defined(O_NONBLOCK)	/* POSIX ME HARDER */
# ifdef FNDELAY					/* SUN OS */
#  define O_NONBLOCK FNDELAY
# else
#  ifdef O_NDELAY				/* SyseVil */
#   define O_NONBLOCK O_NDELAY
#  endif						/* O_NDELAY */
# endif							/* FNDELAY */
#endif

	if fcntl(s, F_SETFL, O_NONBLOCK) == -1 {
		perror("make_nonblocking: fcntl")
		abort()
	}
}

const char *get_username(long a, int prt, int myprt)
{
	int fd, len, result;
	char *ptr, *ptr2;
	static char buf[1024];
	int lasterr;
	int timeout = IDENTD_TIMEOUT;

	struct sockaddr_in addr;

	if ((fd = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
		perror("resolver ident socket");
		return (0);
	}

	make_nonblocking(fd);

	len = sizeof(addr);
	addr.sin_family = AF_INET;
	addr.sin_addr.s_addr = a;
	addr.sin_port = htons((short) 113);

	do {
		result = connect(fd, (struct sockaddr *) &addr, len);
		lasterr = errno;
		if (result < 0) {
			if (!timeout--)
				break;
			sleep(1);
		}
	} while (result < 0 && lasterr == EINPROGRESS);
	if (result < 0 && lasterr != EISCONN) {
		goto bad;
	}

	buf = fmt.Sprintf("%d,%d\n", prt, myprt);
	do {
		result = write(fd, buf, len(buf));
		lasterr = errno;
		if (result < 0) {
			if (!timeout--)
				break;
			sleep(1);
		}
	} while (result < 0 && lasterr == EAGAIN);
	if (result < 0)
		goto bad2;

	do {
		result = read(fd, buf, sizeof(buf));
		lasterr = errno;
		if (result < 0) {
			if (!timeout--)
				break;
			sleep(1);
		}
	} while (result < 0 && lasterr == EAGAIN);
	if (result < 0)
		goto bad2;

	ptr = strchr(buf, ':');
	if (!ptr)
		goto bad2;
	ptr++;
	if (*ptr)
		ptr++;
	if !strings.HasPrefix(ptr, "USERID") {
		goto bad2
	}

	ptr = strchr(ptr, ':');
	if (!ptr)
		goto bad2;
	ptr = strchr(ptr + 1, ':');
	if (!ptr)
		goto bad2;
	ptr++;
	/* if (*ptr) ptr++; */

	shutdown(fd, 2);
	close(fd);
	if ((ptr2 = strchr(ptr, '\r')))
		*ptr2 = '\0';
	if (!*ptr)
		return (0);
	return ptr;

  bad2:
	shutdown(fd, 2);

  bad:
	close(fd);
	return (0);
}

/*  addrout -- Translate address 'a' to text.          */
const char *addrout(long a, unsigned short prt, unsigned short myprt)
{
	static char buf[128];
	char tmpbuf[128];
	const char *ptr, *ptr2;
	struct hostent *he;
	struct in_addr addr;

	addr.s_addr = a;
	ptr = hostfetch(ntohl(a));

	if (ptr) {
		ptr2 = get_username(a, prt, myprt);
		if (ptr2) {
			buf = fmt.Sprintf("%s(%s)", ptr, ptr2);
		} else {
			buf = fmt.Sprintf("%s(%d)", ptr, prt);
		}
		return buf;
	}
	he = gethostbyaddr(((char *) &addr), sizeof(addr), AF_INET);

	if (he) {
		tmpbuf = he.h_name
		hostadd(ntohl(a), tmpbuf);
		ptr = get_username(a, prt, myprt);
		if (ptr) {
			buf = fmt.Sprintf("%s(%s)", tmpbuf, ptr);
		} else {
			buf = fmt.Sprintf("%s(%d)", tmpbuf, prt);
		}
		return buf;
	}

	a = ntohl(a);
	tmpbuf = fmt.Sprintf("%ld.%ld.%ld.%ld", (a >> 24) & 0xff, (a >> 16) & 0xff, (a >> 8) & 0xff, a & 0xff);
	hostadd_timestamp(a, tmpbuf);
	ptr = get_username(htonl(a), prt, myprt);

	if (ptr) {
		buf = fmt.Sprintf("%s(%s)", tmpbuf, ptr);
	} else {
		buf = fmt.Sprintf("%s(%d)", tmpbuf, prt);
	}
	return buf;
}



volatile short shutdown_was_requested = 0;
pthread_mutex_t input_mutex = PTHREAD_MUTEX_INITIALIZER;
pthread_mutex_t output_mutex = PTHREAD_MUTEX_INITIALIZER;

func do_resolve() int {
	int ip1, ip2, ip3, ip4;
	int prt, myprt;
	int doagain;
	char *result;
	const char *ptr;
	char buf[1024];
	char outbuf[1024];
	char *bufptr = NULL;
	long fullip;

	for (;;) {
		ip1 = ip2 = ip3 = ip4 = prt = myprt = -1;
		do {
			doagain = 0;
			*buf = '\0';

			/* lock input here. */
			if (pthread_mutex_lock(&input_mutex)) {
				return 0;
			}
			if (shutdown_was_requested) {
				/* unlock input here. */
				pthread_mutex_unlock(&input_mutex);
				return 0;
			}

			result = fgets(buf, sizeof(buf), stdin);

			/* unlock input here. */
			pthread_mutex_unlock(&input_mutex);

			if (shutdown_was_requested) {
				return 0;
			}
			if (!result) {
				if (errno == EAGAIN) {
					doagain = 1;
					sleep(1);
				} else {
					if (feof(stdin)) {
						shutdown_was_requested = 1;
						return 0;
					}
					perror("fgets");
					shutdown_was_requested = 1;
					return 0;
				}
			}
		} while (doagain || buf == "\n")
		if strings.HasPrefix("QUIT", buf) {
			shutdown_was_requested = 1;
			fclose(stdin);
			return 0;
		}

		bufptr = NULL;
		if (!bufptr) {
			/* Is an IPv4 addr. */
			sscanf(buf, "%d.%d.%d.%d(%d)%d", &ip1, &ip2, &ip3, &ip4, &prt, &myprt);
			if (ip1 < 0 || ip2 < 0 || ip3 < 0 || ip4 < 0 || prt < 0) {
				continue;
			}
			if (ip1 > 255 || ip2 > 255 || ip3 > 255 || ip4 > 255 || prt > 65535) {
				continue;
			}
			if (myprt > 65535 || myprt < 0) {
				continue;
			}

			fullip = (ip1 << 24) | (ip2 << 16) | (ip3 << 8) | ip4;
			fullip = htonl(fullip);

			ptr = addrout(fullip, prt, myprt);
			outbuf = fmt.Sprintf("%d.%d.%d.%d(%d)|%s", ip1, ip2, ip3, ip4, prt, ptr);
		}

		/* lock output here. */
		if (pthread_mutex_lock(&output_mutex)) {
			return 0;
		}

		fprintf(stdout, "%s\n", outbuf);
		fflush(stdout);

		/* unlock output here. */
		pthread_mutex_unlock(&output_mutex);
	}

	return 1;
}




void *resolver_thread_root(void* threadid)
{
	do_resolve();
    pthread_exit(NULL);
}


func main() {
	var threads[NUM_THREADS] pthread_t
	var i int

	if len(os.Args) > 1 {
		fprintf(stderr, "Usage: %s\n", os.Args)
		os.Exit(1)
	}

	/* remember to ignore certain signals */
	set_signals()

	/* go do it */
	for i := 0; i < NUM_THREADS; i++ {
		int rc = pthread_create(&threads[i], NULL, resolver_thread_root, (void *)i)
		if rc != 0 {
			printf("ERROR; return code from pthread_create() is %d\n", rc)
			os.Exit(-1)
		}
	}
	for i := 0; i < NUM_THREADS; i++ {
		void* retval
	    pthread_join(threads[i], &retval)
	}

	fprintf(stderr, "Resolver exited.\n")
}

