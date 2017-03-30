func ts_newobject(thing *object) {
	now := time(nil)
	thing.ts.created = now
	thing.ts.modified = now
	thing.ts.lastused = now
	thing.ts.usecount = 0
}

func ts_useobject(thing dbref) {
	if thing != NOTHING {
		db.Fetch(thing).ts.lastused = time(nil)
		db.Fetch(thing).ts.usecount++
		db.Fetch(thing).flags |= OBJECT_CHANGED
		if Typeof(thing) == TYPE_ROOM {
			ts_useobject(db.Fetch(thing).location)
		}
	}
}

func ts_lastuseobject(thing dbref) {
	if thing != NOTHING {
		db.Fetch(thing).ts.lastused = time(nil)
		db.Fetch(thing).flags |= OBJECT_CHANGED
		if Typeof(thing) == TYPE_ROOM {
			ts_lastuseobject(db.Fetch(thing).location)
		}
	}
}

func ts_modifyobject(thing dbref) {
	db.Fetch(thing).ts.modified = time(nil)
	db.Fetch(thing).flags |= OBJECT_CHANGED
}