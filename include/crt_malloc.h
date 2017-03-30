/* CrT's own silly little malloc wrappers for debugging purposes: */

#ifndef _CRT_MALLOC_H
#define _CRT_MALLOC_H

#define malloc(x)            CrT_malloc(           x,    __FILE__, __LINE__)
#define calloc(x,y)          CrT_calloc(           x, y, __FILE__, __LINE__)
#define realloc(x,y)         CrT_realloc(          x, y, __FILE__, __LINE__)
#define free(x)              CrT_free(             x,    __FILE__, __LINE__)

#endif /* _CRT_MALLOC_H */