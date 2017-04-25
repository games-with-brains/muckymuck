package fbmuck

//	FIXME: Changed will be replaced by a new layered DB design in which records are stored in a list of maps which act as snapshots of currently loaded records

type TimeStamps struct {
	Created time.Time
	Modified time.Time
	LastUsed time.Time
	Uses int
	MPIUses int
	time.Duration
	Changed
}

func (t *TimeStamps) Touch() {
	t.Modified = time.Now()
	t.Changed = true
}

func NewTimeStamps() *TimeStamps {
	now := time.Now()
	return &TimeStamps{ Created: now, Modified: now, LastUsed: now }
}

func ts_useobject(thing ObjectID) {
	if thing != NOTHING {
		p := DB.Fetch(thing)
		p.LastUsed = time(nil)
		p.Uses++
		p.Touch()
		if Typeof(thing) == TYPE_ROOM {
			ts_useobject(p.Location)
		}
	}
}

func ts_lastuseobject(thing ObjectID) {
	if thing != NOTHING {
		p := DB.Fetch(thing)
		p.LastUsed = time(nil)
		p.Touch()
		if Typeof(thing) == TYPE_ROOM {
			ts_lastuseobject(p.Location)
		}
	}
}

func ts_modifyobject(thing ObjectID) {
	DB.Fetch(thing).Touch()
}