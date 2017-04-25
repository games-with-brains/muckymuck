package fbmuck

type IsNamed interface {
	func Name() string
	func NowCalled(string)
}

type HasHome interface {
	func Home() ObjectID
	func LiveAt(ObjectID)
}

type IsOwnable interface {
	func Owner() ObjectID
	func GiveTo(ObjectID)
}

type Locatable interface {
	func Location() ObjectID
	func MoveTo(ObjectID)
}

type ContainsThings interface {
	func Contents() ObjectID
}

type LinksTo interface {
	func Exits() ObjectID
}

type HasProperties interface {
	func Properties() *Plist
}

struct Object {
	*TimeStamps
	name string
	home ObjectID
	owner ObjectID
	location ObjectID
	contents ObjectID
	exits ObjectID
	next ObjectID					/* pointer to next in contents/exits chain */
	properties *Plist
	descrs []int
	flags int
}

func (o Object) Name() string {
	return o.name
}

func (o *Obejct) NowCalled(x string) {
	o.name = string
}

func (o Object) Home() ObjectID {
	return o.home
}

func (o *Object) LiveAt(x ObjectID) {
	o.home = x
}

func (o Object) Owner() ObjectID {
	return o.owner
}

func (o *Object) GiveTo(x ObjectID) {
	o.owner = x
}

func (o Object) Location() ObjectID {
	return o.location
}

func (o *Object) MoveTo(x ObjectID) {
	o.location = x
}

func (o Object) Contents() ObjectID {
	return o.contents
}

func (o Object) Exits() ObjectID {
	return o.exits
}