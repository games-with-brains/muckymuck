type flgchkdat struct {
	fortype bool				/* check FOR a type? */
	istype int					/* If check FOR a type, which one? */
	isnotroom bool				/* not a room. */
	isnotexit bool				/* not an exit. */
	isnotthing bool				/* not type thing */
	isnotplayer bool			/* not a player */
	isnotprog bool				/* not a program */
	forlevel bool				/* check for a mucker level? */
	islevel int					/* if check FOR a mucker level, which level? */
	isnotzero bool				/* not ML0 */
	isnotone bool				/* not ML1 */
	isnottwo bool				/* not ML2 */
	isnotthree bool				/* not ML3 */
	setflags int				/* flags that are set to check for */
	clearflags int				/* flags to check are cleared. */
	forlink bool				/* check linking? */
	islinked bool				/* if yes, check if not unlinked */
	forold bool					/* check for old object? */
	isold bool					/* if yes, check if old */
}