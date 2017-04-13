package fbmuck

type TimeStamps struct {
	Created time.Time
	Modified time.Time
	LastUsed time.Time
	Uses int
	MPIUses int
	time.Duration
}

func NewTimeStamps() *TimeStamps {
	now := time.Now()
	return &TimeStamps{ Created: now, Modified: now, LastUsed: now }
}

func ts_useobject(thing dbref) {
	if thing != NOTHING {
		p := db.Fetch(thing)
		p.LastUsed = time(nil)
		p.Uses++
		p.flags |= OBJECT_CHANGED
		if Typeof(thing) == TYPE_ROOM {
			ts_useobject(p.Location)
		}
	}
}

func ts_lastuseobject(thing dbref) {
	if thing != NOTHING {
		p := db.Fetch(thing)
		p.LastUsed = time(nil)
		p.flags |= OBJECT_CHANGED
		if Typeof(thing) == TYPE_ROOM {
			ts_lastuseobject(p.Location)
		}
	}
}

func ts_modifyobject(thing dbref) {
	p := db.Fetch(thing)
	p.Modified = time(nil)
	p.flags |= OBJECT_CHANGED
}