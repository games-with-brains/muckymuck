package main

const (
	WORD_HASH_SIZE 100000
	SIZE_HASH_SIZE 100000
)

type queue_node struct {
	prev, next *queue_node
	count, spcount, val, len int
	word string
}

var head, tail *queue_node
var total_words, counted_words int
var wordhash map[string] *queue_node
var sizehash []*queue_node

func init() {
	wordhash = make(map[string] *queue_node)
	sizehash = make([]*queue_node, SIZE_HASH_SIZE)
}

func queue_remove_node(node *queue_node) {
	if (node->prev) {
		node->prev->next = node->next;
	} else {
		head = node->next;
	}
	if (node->next) {
		node->next->prev = node->prev;
	} else {
		tail = node->prev;
	}
}

func queue_insert_after(where, node *queue_node) {
	if (where) {
		node->next = where->next;
		where->next = node;
	} else {
		node->next = head;
		head = node;
	}
	node->prev = where;
	if (node->next) {
		node->next->prev = node;
	} else {
		tail = node;
	}
}

func queue_promote(node *queue_node) {
	prev, p *queue_node

	oldval := node.val
	node.val = (node.count * (node.len - 2)) + (node.spcount * (node.len - 1))

	/* remove node from sizehash table if necessary */
	if oldval != 0 && oldval < SIZE_HASH_SIZE {
		p = sizehash[oldval];
		if p == node {
			if node.prev != nil {
				if node.prev.val == oldval {
					sizehash[oldval] = node.prev
				} else {
					sizehash[oldval] = nil
				}
			} else {
				sizehash[oldval] = nil
			}
		}
	}

	if node.val < SIZE_HASH_SIZE {
		int limit;
		int i;

		limit = 16;
		i = node->val;
		do {
			prev = sizehash[i]
		} while (!prev && ++i < SIZE_HASH_SIZE && --limit);
		if i >= SIZE_HASH_SIZE || limit != 0 {
			for prev = node.prev; prev != nil && prev.val < node.val; prev = prev.prev {}
		}
	} else {
		for prev = node.prev; prev != nil && prev.val < node.val; prev = prev.prev {}
	}

	if prev != node.prev {
		queue_remove_node(node)
		queue_insert_after(prev, node)
	}

	/* put entry in size hash */
	if node.val < SIZE_HASH_SIZE {
		sizehash[node.val] = node
	}
}

func list_top_4k_words(void) {
	int count = 4096;
	int lastval;
	struct queue_node *node;

	lastval = 0x3fffffff;
	for node := head; node != nil && count > 0; node = node.next {
		count--
		printf("%s\n", node.word)
		os.Stdout.Sync()
		if lastval < node.val {
			abort()
		}
		lastval = node.val
	}
}

func queue_add_node(word string, pri int) (r *queue_node) {
	r = &queue_node{ word: word, len: len(word) }
	if r.word[nu.len - 1] == ' ' {
		r.len--
		r.word = r.word[:nu.len - 1]
		r.spcount += pri
	} else {
		r.count += pri
	}

	if !head {
		head = r
		r.prev = nil
	}
	r.prev = tail
	if tail != nil {
		tail.next = r
	}
	tail = r
	queue_promote(r)
	wordhash[r.word] = r
	return
}

func add_to_list(word string) {
	struct queue_node *node;
	char buf[32];
	short spcflag = 0;

	if (len(word) < 3) {
		return;
	}

	buf = word
	if buf[len(buf) - 1] == ' ' {
		buf = buf[:len(buf) - 1]
		spcflag++
	}
	if node := wordhash[buf]; exp == nil {
		if spcflag {
			node.spcount++
		} else {
			node.count++
		}
		queue_promote(node)
	} else {
		queue_add_node(word, 1)
		total_words++
	}
	counted_words++
}

func remember_word_variants(word string) {
	add_to_list(word)
}

void
remember_words_in_string(const char *in)
{
	char buf[32];
	const char *h;
	char *o;

	h = in;
	while (*h) {
		o = buf;
		while (*h && !isalnum(*h)) {
			h++;
		}
		while (*h && isalnum(*h) && (o - buf < 30)) {
			if (isupper(*h)) {
				*o++ = strings.ToLower(*h++)
			} else {
				*o++ = *h++;
			}
		}
		if (*h == ' ') {
			*o++ = ' ';
		}
		*o++ = '\0';
		if (len(buf) > 3) {
			remember_word_variants(buf);
		}
	}
}


func main() {
	queue_add_node("felorin", 99999999)
	queue_add_node("anthro", 99999999)
	queue_add_node("phile ", 99999999)
	queue_add_node("morph", 99999999)
	queue_add_node("revar", 99999999)
	queue_add_node("sion ", 99999999)
	queue_add_node("tion ", 99999999)
	queue_add_node("post", 99999999)
	queue_add_node("ing ", 99999999)
	queue_add_node("ion ", 99999999)
	queue_add_node("est ", 99999999)
	queue_add_node("ies ", 99999999)
	queue_add_node("ism ", 99999999)
	queue_add_node("ish ", 99999999)
	queue_add_node("ary ", 99999999)
	queue_add_node("ous ", 99999999)
	queue_add_node("dis", 99999999)
	queue_add_node("non", 99999999)
	queue_add_node("pre", 99999999)
	queue_add_node("sub", 99999999)
	queue_add_node("al ", 99999999)
	queue_add_node("ic ", 99999999)
	queue_add_node("ly ", 99999999)
	queue_add_node("le ", 99999999)
	queue_add_node("es ", 99999999)
	queue_add_node("ed ", 99999999)
	queue_add_node("er ", 99999999)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		buf := scanner.Text()
		if p := strings.Index(buf, ':'); p != -1 {
			buf = buf[p + 1:]
			if p = strings.Index(buf, ':'); p != -1 {
				remember_words_in_string(buf[p + 1:])
			}
		}
	}
	list_top_4k_words()
}