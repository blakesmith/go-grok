#include "grok.h"

int g_grok_global_initialized = 0;
pcre *g_pattern_re = NULL;
int g_pattern_num_captures = 0;
int g_cap_name = 0;
int g_cap_pattern = 0;
int g_cap_subname = 0;
int g_cap_predicate = 0;
int g_cap_definition = 0;

grok_t *grok_new() {
  grok_t *grok;
  grok = malloc(sizeof(grok_t));
  grok_init(grok);
  return grok;
}

void grok_init(grok_t *grok) {
  //int ret;
  grok->re = NULL;
  grok->pattern = NULL;
  grok->full_pattern = NULL;
  grok->pcre_capture_vector = NULL;
  grok->pcre_num_captures = 0;
  grok->max_capture_num = 0;
  grok->pcre_errptr = NULL;
  grok->pcre_erroffset = 0;
  grok->logmask = 0;
  grok->logdepth = 0;

#ifndef GROK_TEST_NO_PATTERNS
  grok->patterns = tctreenew();
#endif /* GROK_TEST_NO_PATTERNS */

#ifndef GROK_TEST_NO_CAPTURE
  grok->captures_by_id = tctreenew();
  grok->captures_by_name = tctreenew();
  grok->captures_by_subname = tctreenew();
  grok->captures_by_capture_number = tctreenew();
#endif /* GROK_TEST_NO_CAPTURE */

  if (g_grok_global_initialized == 0) {
    /* do first initalization */
    g_grok_global_initialized = 1;

    /* VALGRIND NOTE: Valgrind complains here, but this is a global variable.
     * Ignore valgrind here. */
    g_pattern_re = pcre_compile(PATTERN_REGEX, 0,
                                &grok->pcre_errptr,
                                &grok->pcre_erroffset,
                                NULL);
    if (g_pattern_re == NULL) {
      fprintf(stderr, "Internal compiler error: %s\n", grok->pcre_errptr);
      fprintf(stderr, "Regexp: %s\n", PATTERN_REGEX);
      fprintf(stderr, "Position: %d\n", grok->pcre_erroffset);
    }

    pcre_fullinfo(g_pattern_re, NULL, PCRE_INFO_CAPTURECOUNT,
                  &g_pattern_num_captures);
    g_pattern_num_captures++; /* include the 0th group */
    g_cap_name = pcre_get_stringnumber(g_pattern_re, "name");
    g_cap_pattern = pcre_get_stringnumber(g_pattern_re, "pattern");
    g_cap_subname = pcre_get_stringnumber(g_pattern_re, "subname");
    g_cap_predicate = pcre_get_stringnumber(g_pattern_re, "predicate");
    g_cap_definition = pcre_get_stringnumber(g_pattern_re, "definition");
  }
}

void grok_clone(grok_t *dst, const grok_t *src) {
  grok_init(dst);
  dst->patterns = src->patterns;
  dst->logmask = src->logmask;
  dst->logdepth = src->logdepth + 1;
}
