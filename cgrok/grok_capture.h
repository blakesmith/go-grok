#ifndef _GROK_CAPTURE_INTERNAL_H_
#define _GROK_CAPTURE_INTERNAL_H_

#include "grok.h"

struct grok_capture {
	int name_len;
	char *name;
	int subname_len;
	char *subname;
	int pattern_len;
	char *pattern;
	int id;
	int pcre_capture_number;
	int predicate_lib_len;
	char *predicate_lib;
	int predicate_func_name_len;
	char *predicate_func_name;
	struct {
		uint32_t extra_len;
		char *extra_val;
	} extra;
};
typedef struct grok_capture grok_capture;

void grok_capture_init(grok_t *grok, grok_capture *gct);
void grok_capture_free(grok_capture *gct);

void grok_capture_add(grok_t *grok, const grok_capture *gct, int only_renamed);
const grok_capture *grok_capture_get_by_id(const grok_t *grok, int id);
const grok_capture *grok_capture_get_by_name(const grok_t *grok, const char *name);
const grok_capture *grok_capture_get_by_subname(const grok_t *grok,
                                                const char *subname);
const grok_capture *grok_capture_get_by_capture_number(grok_t *grok,
                                                       int capture_number);

TCTREE_ITER *grok_capture_walk_init(const grok_t *grok);
const grok_capture *grok_capture_walk_next(const TCTREE_ITER *iter, const grok_t *grok);

int grok_capture_set_extra(grok_t *grok, grok_capture *gct, void *extra);
void _grok_capture_encode(grok_capture *gct, char **data_ret, int *size_ret);
void _grok_capture_decode(grok_capture *gct, char *data, int size);


#endif /* _GROK_CAPTURE_INTERNAL_H_ */
