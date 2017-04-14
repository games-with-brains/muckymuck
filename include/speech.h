#ifndef _SPEECH_H_
#define _SPEECH_H_

void
do_say(ObjectID player, const char *message)
;

void
do_whisper(int descr, ObjectID player, const char *arg1, const char *arg2)
;

void
do_pose(ObjectID player, const char *message)
;

void
do_wall(ObjectID player, const char *message)
;

void
do_gripe(ObjectID player, const char *message)
;

void
do_page(ObjectID player, const char *arg1, const char *arg2)
;

void
notify_listeners(ObjectID who, ObjectID xprog, ObjectID obj, ObjectID room, const char *msg, int isprivate)
;

void
notify_except(ObjectID first, ObjectID exception, const char *msg, ObjectID who)
;

void
parse_oprop(int descr, ObjectID player, ObjectID dest, ObjectID exit, const char *propname,
			   const char *prefix, const char *whatcalled)
;

void
parse_omessage(int descr, ObjectID player, ObjectID dest, ObjectID exit, const char *msg,
			   const char *prefix, const char *whatcalled, int mpiflags)
;


int
blank(const char *s)
;

#endif /* !defined _SPEECH_H_ */

#ifdef DEFINE_HEADER_VERSIONS

#ifndef speechh_version
#define speechh_version
const char *speech_h_version = "$RCSfile: speech.h,v $ $Revision: 1.4 $";
#endif
#else
extern const char *speech_h_version;
#endif

