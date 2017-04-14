package fbmuck

func (o ObjectID) IsValid() bool {
	return o > -1 && o < db_top
}

/*
	These helper functions operate in two modes:

		if one or more functions are passed in then each of these is executed in sequence for a valid database reference, or NOTHING is returned
		if no functions are passed in then either the valid database reference is returned or a panic occurs
*/

func (o ObjectID) ValidObject(f ...func(ObjectID)) (r ObjectID) {
	switch {
	case o.IsValid():
		for _, f := range f {
			f(o)
		}
		r = o
	case f == nil:
		panic("Not a valid object reference")
	default:
		r = NOTHING
	}
	return
}

func (o ObjectID) ValidRemoteObject(player ObjectID, mlev int f ...func(ObjectID)) (r ObjectID) {
	r = o.ValidObject(f...)
	check_remote(r)
	return
}

func (o ObjectID) ValidPlayer(f ...func(ObjectID)) (r ObjectID) {
	switch {
	case o.IsValid() && IsPlayer(o):
		for _, f := range f {
			f(o)
		}
		r = o
	case f == nil:
		panic("Not a valid object reference")
	default:
		r = NOTHING
	return
}

func (o ObjectID) ValidRemoteObject(player ObjectID, mlev int f ...func(ObjectID)) (r ObjectID) {
	r = o.ValidPlayer(f...)
	check_remote(r)
	return
}

func valid_object_or_home(oper interface{}, f ...func(ObjectID)) (obj ObjectID) {
	if obj = oper.(ObjectID); obj == HOME {
		for _, f := range f {
			f(HOME)
		}
	} else {
		obj = obj.ValidObject(f...)
	}
	return
}
