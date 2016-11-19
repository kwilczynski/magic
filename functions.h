#if !defined(_FUNCTIONS_H)
#define _FUNCTIONS_H 1

#if defined(__cplusplus)
extern "C" {
#endif

#include "common.h"

#if defined(HAVE_BROKEN_MAGIC)
# define OVERRIDE_LOCALE(f, r, ...)	      \
    do {				      \
	save_t __l_##f;			      \
	override_current_locale(&(__l_##f));  \
	r = f(__VA_ARGS__);		      \
	restore_current_locale(&(__l_##f));   \
    } while(0)
#else
# define OVERRIDE_LOCALE(f, r, ...) \
    do {			    \
	r = f(__VA_ARGS__);	    \
    } while(0)
#endif

#define SUPPRESS_EVERYTHING(f, r, ...)	    \
    do {				    \
	save_t __e_##f;			    \
	suppress_error_output(&(__e_##f));  \
	OVERRIDE_LOCALE(f, r, __VA_ARGS__); \
	restore_error_output(&(__e_##f));   \
    } while(0)

#define MAGIC_FUNCTION(f, r, x, ...)		    \
     do {					    \
	if ((x) & MAGIC_ERROR) {		    \
	    OVERRIDE_LOCALE(f, r, __VA_ARGS__);     \
	}					    \
	else {					    \
	    SUPPRESS_EVERYTHING(f, r, __VA_ARGS__); \
	 }					    \
     } while(0)

struct file_data {
    int old_fd;
    int new_fd;
    fpos_t position;
};

typedef struct file_data file_data_t;

struct locale_data {
    locale_t old_locale;
    locale_t new_locale;
};

typedef struct locale_data locale_data_t;

struct save {
    int status;
    union {
	file_data_t file;
	locale_data_t locale;
    } data;
};

typedef struct save save_t;

extern const char* magic_getpath_wrapper(void);

extern int magic_setflags_wrapper(struct magic_set *ms, int flags);
extern int magic_load_wrapper(struct magic_set *ms, const char *magicfile, int flags);
extern int magic_compile_wrapper(struct magic_set *ms, const char *magicfile, int flags);
extern int magic_check_wrapper(struct magic_set *ms, const char *magicfile, int flags);

extern const char* magic_file_wrapper(struct magic_set *ms, const char *filename, int flags);
extern const char* magic_buffer_wrapper(struct magic_set *ms, const void *buffer, size_t size, int flags);
extern const char* magic_descriptor_wrapper(struct magic_set *ms, int fd, int flags);

extern int magic_version_wrapper(void);

#if defined(__cplusplus)
}
#endif

#endif /* _FUNCTIONS_H */

/* vim: set ts=8 sw=4 sts=4 noet : */
