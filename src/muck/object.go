package fbmuck

struct Object {
	*TimeStamps
	name string
	Location ObjectID				/* pointer to container */
	Owner ObjectID
	Contents ObjectID
	Exits ObjectID
	next ObjectID					/* pointer to next in contents/exits chain */
	properties *Plist
	flags object_flag_type
}
