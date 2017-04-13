package fbmuck

struct Object {
	*TimeStamps
	name string
	Location dbref				/* pointer to container */
	Owner dbref
	Contents dbref
	Exits dbref
	next dbref					/* pointer to next in contents/exits chain */
	properties *Plist
	flags object_flag_type
}
