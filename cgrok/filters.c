/* ANSI-C code produced by gperf version 3.0.3 */
/* Command-line: gperf  */
/* Computed positions: -k'' */


#define _GPERF_
#include "grok.h"
#include "grok_logging.h"
#include "stringhelper.h"

int filter_jsonencode(grok_match_t *gm, char **value, int *value_len,
                      int *value_size);
int filter_shellescape(grok_match_t *gm, char **value, int *value_len,
                      int *value_size);
int filter_shelldqescape(grok_match_t *gm, char **value, int *value_len,
                      int *value_size);

struct filter {
  const char *name;
  int (*func)(grok_match_t *gm, char **value, int *value_len,
              int *value_size); 
};

#define TOTAL_KEYWORDS 3
#define MIN_WORD_LENGTH 10
#define MAX_WORD_LENGTH 13
#define MIN_HASH_VALUE 10
#define MAX_HASH_VALUE 13
/* maximum key range = 4, duplicates = 0 */

#ifdef __GNUC__
__inline
#else
#ifdef __cplusplus
inline
#endif
#endif
/*ARGSUSED*/
static unsigned int
_string_filter_hash (register const char *str, register unsigned int len)
{
  return len;
}

#ifdef __GNUC__
__inline
#ifdef __GNUC_STDC_INLINE__
__attribute__ ((__gnu_inline__))
#endif
#endif
const struct filter *
string_filter_lookup (register const char *str, register unsigned int len)
{
  static const struct filter wordlist[] =
    {
      {"jsonencode",filter_jsonencode},
      {"shellescape",filter_shellescape},
      {"shelldqescape",filter_shelldqescape}
    };

  if (len <= MAX_WORD_LENGTH && len >= MIN_WORD_LENGTH)
    {
      register int key = _string_filter_hash (str, len);

      if (key <= MAX_HASH_VALUE && key >= MIN_HASH_VALUE)
        {
          register const struct filter *resword;

          switch (key - 10)
            {
              case 0:
                resword = &wordlist[0];
                goto compare;
              case 1:
                resword = &wordlist[1];
                goto compare;
              case 3:
                resword = &wordlist[2];
                goto compare;
            }
          return 0;
        compare:
          {
            register const char *s = resword->name;

            if (*str == *s && !strncmp (str + 1, s + 1, len - 1) && s[len] == '\0')
              return resword;
          }
        }
    }
  return 0;
}


int filter_jsonencode(grok_match_t *gm, char **value, int *value_len,
                      int *value_size) {
  grok_log(gm->grok, LOG_REACTION, "filter executing");

  /* json.org says " \ and / should be escaped, in addition to 
   * contol characters (newline, etc).
   *
   * Some validators will pass non-escaped forward slashes (solidus) but
   * we'll escape it anyway. */
  string_escape(value, value_len, value_size, "\\\"/", 3, ESCAPE_LIKE_C);
  string_escape(value, value_len, value_size, "", 0,
                ESCAPE_NONPRINTABLE | ESCAPE_UNICODE);
  return 0;
}

int filter_shellescape(grok_match_t *gm, char **value, int *value_len,
                       int *value_size) {
  grok_log(gm->grok, LOG_REACTION, "filter executing");
  string_escape(value, value_len, value_size, "`^()&{}[]$*?!|;'\"\\", -1, ESCAPE_LIKE_C);
  return 0;
}

int filter_shelldqescape(grok_match_t *gm, char **value, int *value_len,
                       int *value_size) {
  grok_log(gm->grok, LOG_REACTION, "filter executing");
  string_escape(value, value_len, value_size, "\\`$\"", -1, ESCAPE_LIKE_C);
  return 0;
}

