package fbmuck

func do_rob(int descr, ObjectID player, const char *what) {
	md := NewMatch(descr, player, what, IsPlayer).MatchNeighbor().MatchMe()
	if Wizard(DB.Fetch(player).Owner) {
		md.MatchAbsolute().MatchPlayer()
	}
	switch thing := md.MatchResult(); {
	case thing == NOTHING:
		notify(player, "Rob whom?")
	case thing == AMBIGUOUS:
		notify(player, "I don't know who you mean!")
	case TYPEOF(thing) != TYPE_PLAYER:
		notify(player, "Sorry, you can only rob other players.")
	case get_property_value(thing, MESGPROP_VALUE) < 1:
		notify(player, fmt.Sprintf("%s has no %s.", DB.Fetch(thing).name, tp_pennies))
		notify(thing, fmt.Sprintf("%s tried to rob you, but you have no %s to take.", DB.Fetch(player).name, tp_pennies))
	case can_doit(descr, player, thing, "Your conscience tells you not to."):
		/* steal a penny */
		add_property(DB.Fetch(player).Owner, MESGPROP_VALUE, nil, get_property_value(DB.Fetch(player).Owner, MESGPROP_VALUE) + 1)
		DB.Fetch(player).flags |= OBJECT_CHANGED
		add_property(thing, MESGPROP_VALUE, nil, get_property_value(thing, MESGPROP_VALUE) - 1)
		DB.Fetch(thing).flags |= OBJECT_CHANGED
		notify_fmt(player, "You stole a %s.", tp_penny)
		notify(thing, fmt.Sprintf("%s stole one of your %s!", DB.Fetch(player).name, tp_pennies))
	}
}

func do_kill(descr int, player ObjectID, what string, cost int) {
	if cost < tp_kill_min_cost {
		cost = tp_kill_min_cost
	}
	md := NewMatch(descr, player, what, IsPlayer).MatchNeighbor().MatchMe()
	if Wizard(DB.Fetch(player).Owner) {
		md.MatchPlayer().MatchAbsolute()
	}
	switch victim := md.MatchResult(); {
	case victim == NOTHING:
		notify(player, "I don't see that player here.")
	case victim == AMBIGUOUS:
		notify(player, "I don't know who you mean!")
	case Typeof(victim) != TYPE_PLAYER:
		notify(player, "Sorry, you can only kill other players.")
	case DB.Fetch(DB.Fetch(player).Location).flags & HAVEN != 0:
		notify(player, "You can't kill anyone here!")
	case tp_restrict_kill && DB.Fetch(player).flags & KILL_OK == 0:
		notify(player, "You have to be set Kill_OK to kill someone.")
	case tp_restrict_kill && DB.Fetch(victim).flags & KILL_OK == 0:
		notify(player, "They don't want to be killed.")
	case !payfor(player, cost):
		notify_fmt(player, "You don't have enough %s.", tp_pennies)
	case RANDOM() % tp_kill_base_cost < cost && !Wizard(DB.Fetch(victim).Owner):
		/* you killed him */
		if get_property_class(victim, MESGPROP_DROP) {
			notify(player, get_property_class(victim, MESGPROP_DROP))
		} else {
			notify(player, fmt.Sprintf("You killed %s!", DB.Fetch(victim).name))
		}

		var buf string
		if get_property_class(victim, MESGPROP_ODROP) {
			buf = fmt.Sprintf("%s killed %s! ", DB.Fetch(player).name, DB.Fetch(victim).name)
			parse_oprop(descr, player, DB.Fetch(player).Location, victim, MESGPROP_ODROP, buf, "(@Odrop)")
		} else {
			buf = fmt.Sprintf("%s killed %s!", DB.Fetch(player).name, DB.Fetch(victim).name)
		}
		notify_except(DB.Fetch(DB.Fetch(player).Location).Contents, player, buf, player)

		/* maybe pay off the bonus */
		if get_property_value(victim, MESGPROP_VALUE) < tp_max_pennies {
			notify(victim, fmt.Sprintf("Your insurance policy pays %d %s.", tp_kill_bonus, tp_pennies))
			add_property(victim, MESGPROP_VALUE, nil, get_property_value(victim, MESGPROP_VALUE) + tp_kill_bonus)
			DB.Fetch(victim).flags |= OBJECT_CHANGED
		} else {
			notify(victim, "Your insurance policy has been revoked.")
		}
		send_home(descr, victim, 1)
	default:
		notify(player, "Your murder attempt failed.")
		notify(victim, fmt.Sprintf("%s tried to kill you!", DB.Fetch(player).name))
	}
}

func do_give(descr int, player ObjectID, recipient string, amount int) {
	switch {
	case amount < 0 && !Wizard(DB.Fetch(player).Owner):
		notify(player, "Try using the \"rob\" command.")
	case amount == 0:
		notify_fmt(player, "You must specify a positive number of %s.", tp_pennies)
	default:
		/* check recipient */
		md := NewMatch(descr, player, recipient, IsPlayer).MatchNeighbor().MatchMe()
		if Wizard(DB.Fetch(player).Owner) {
			md.MatchPlayer().MatchAbsolute()
		}
		switch who := md.MatchResult(); who {
		case NOTHING:
			notify(player, "Give to whom?")
		case AMBIGUOUS:
			notify(player, "I don't know who you mean!")
		default:
			switch wiz_owned := Wizard(DB.Fetch(player).Owner); {
			case !wiz_owned && Typeof(who) != TYPE_PLAYER:
				notify(player, "You can only give to other players.")
			case !wiz_owned && GTVALUE(who) + amount > tp_max_pennies:
				notify_fmt(player, "That player doesn't need that many %s!", tp_pennies)
			case !payfor(player, amount):
				notify_fmt(player, "You don't have that many %s to give!", tp_pennies)
			default:
				switch who.(type) {
				case TYPE_PLAYER:
					add_property(who, MESGPROP_VALUE, nil, get_property_value(who, MESGPROP_VALUE) + amount)
					switch amount {
					case -1, 1:
						notify(player, fmt.Sprintf("You take %d %s from %s.", -amount, tp_penny, DB.Fetch(who).name))
						notify(who, fmt.Sprintf("%s takes %d %s from you!", DB.Fetch(player).name, -amount, tp_penny))
					default:
						notify(player, fmt.Sprintf("You take %d %s from %s.", -amount, tp_pennies, DB.Fetch(who).name))
						notify(who, fmt.Sprintf("%s takes %d %s from you!", DB.Fetch(player).name, -amount, tp_pennies))
					}
				case TYPE_THING:
					add_property(who, MESGPROP_VALUE, nil, get_property_value(who, MESGPROP_VALUE) + amount)
					if v := get_property_value(who, MESGPROP_VALUE); v == 1 {
						notify(player, fmt.Sprintf("You change the value of %s to %d %s.", DB.Fetch(who).name, v, tp_penny))
					} else {
						notify(player, fmt.Sprintf("You change the value of %s to %d %s.", DB.Fetch(who).name, v, tp_pennies))
					}
				default:
					notify_fmt(player, "You can't give %s to that!", tp_pennies)
				}
				DB.Fetch(who).flags |= OBJECT_CHANGED
			}
		}
	}
}