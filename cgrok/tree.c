#include "tree.h"
#include <stdio.h>
#include <stdint.h>
#include <string.h>

// Toyko cabinet accepts any arbitrary-sized data as a tree key.
// We support this by using the first 4 bytes of the key to store the length.
// So we have to wrap the libdict comparators with implementations that skip the first four bytes
// of the key
int tccmpint32(const void* k1, const void* k2) {
  uint32_t a = *(uint32_t*)(k1+4);
  uint32_t b = *(uint32_t*)(k2+4);
  return (a > b) - (a < b);
}

int dict_var_str_cmp(const void* k1, const void* k2) {
  return dict_str_cmp(k1+4, k2+4);
}

void tcfree(void *key, void *value) {
  free(key);
  free(value);
}

TCTREE *tctreenew(void) {
  return tctreenew2(dict_var_str_cmp, NULL);
}

// Make a new tree with a user-defined comparator (must be tccmpint32 or dict_var_str_cmp),
// because of how we pack the keys
TCTREE *tctreenew2(dict_compare_func cp, void *cmpop) {
  TCTREE *tree = malloc(sizeof(TCTREE));
  if (tree == NULL) {
    fprintf(stderr, "Failed to malloc new tree\n");
    exit(0);
  }
  tree->dict = hb_dict_new(cp, tcfree);
  tree->iter = NULL;
  return tree; 
}

// Create an iterator to walk keys in ascending order
// TokyoCabinet only allows one iterator per tree,
// so we do the same.
void tctreeiterinit(TCTREE *tree) {
  if (tree->iter != NULL) {
    dict_itor_free(tree->iter);
  }
  tree->iter = dict_itor_new(tree->dict);
  dict_itor_first(tree->iter);
}

// Get the next key from the iterator
const void *tctreeiternext(TCTREE *tree, int *sp) {
  void *key = dict_itor_key(tree->iter);
  dict_itor_next(tree->iter);
  *sp = 0;
  if (key) {
    *sp = *(uint32_t*)key;
    return key+4;
  }
  return NULL;
}

// Pack a key or value: uint32_t + body + NULL
// Callers might use integers, structs, etc. for 
// keys and values, so we can't depend on null-termination
// to find the lenght of a key.
void *tcpack(const void *value, uint32_t size) {
  void *buf = malloc(size + 5);
  if (buf == NULL) {
    return NULL;
  }
  memset(buf, 0, size+5);
  memcpy(buf, &size, 4);
  memcpy(buf+4, value, size);
  return buf;
}

// Insert a key-value pair. If the key already exists the value will be overwritten
void tctreeput(TCTREE *tree, const void *kbuf, int ksiz, const void *vbuf, int vsiz) {
  void *key = tcpack(kbuf, ksiz);
  if (key == NULL) {
    fprintf(stderr, "Failed to malloc tree key for (tctreeput)\n");
    exit(0);
  }
  void *val = tcpack(vbuf, vsiz);
  if (val == NULL) {
    fprintf(stderr, "Failed to malloc tree value (tctreeput)\n");
    exit(0);
  }
  bool inserted; 
  void ** valPtr = dict_insert(tree->dict, key, &inserted);
  if (!inserted) {
    free(key);
    free(*valPtr);
  }
  *valPtr = val;
}

// Insert a key-value pair. If the key already exists return false and keep the original value
bool tctreeputkeep(TCTREE *tree, const void *kbuf, int ksiz, const void *vbuf, int vsiz) {
  void *key = tcpack(kbuf, ksiz);
  if (key == NULL) {
    fprintf(stderr, "Failed to malloc tree key (tctreeputkeep)\n");
    exit(0);
  } 
  void *val = tcpack(vbuf, vsiz); 
  if (val == NULL) {
    fprintf(stderr, "Failed to malloc tree value (tctreeputkeep)\n");
    exit(0);
  }
  bool inserted;
  void **valPtr = dict_insert(tree->dict, key, &inserted);
  if (inserted) {
    *valPtr = val;
  } else {
    free(key);
    free(val); 
  }
  return inserted;
}

// Get the value for the given key, or NULL if the key is not in the tree
const void *tctreeget(TCTREE *tree, const void *kbuf, int ksiz, int *sp) {
  const void *key = tcpack(kbuf, ksiz);
  if (key == NULL) {
    fprintf(stderr, "Failed to malloc tree key (tctreeget)\n");
    exit(0);
  }
  void *value = dict_search(tree->dict, key);
  *sp = 0;
  free(key);
  if (value) {
    *sp = *(uint32_t*)value;
    return value+4;
  }
  return NULL;
}

// Remove all elements from the tree
void tctreeclear(TCTREE *tree) {
  dict_clear(tree->dict); 
}

// Free the tree and associated iterator
void tctreedel(TCTREE *tree) {
  if (tree == NULL) {
    return;
  }
  dict_free(tree->dict);
  if (tree->iter != NULL) {
    dict_itor_free(tree->iter);
  }
  free(tree);
}

// A simple, singly linked list
TCLIST *tclistnew(void) {
  TCLIST *list = malloc(sizeof(TCLIST));
  if (list == NULL) {
    fprintf(stderr, "Failed to malloc new list\n");
    exit(0);
  }
  list->len = 0;
  TCLISTNODE *newElem = malloc(sizeof(TCLISTNODE));
  if (newElem == NULL) {
    fprintf(stderr, "Failed to malloc new list element\n");
    exit(0);
  }
  newElem->val = NULL;
  list->head = newElem;
  return list;
}

// Get the number of elements in the list
int tclistnum(const TCLIST *list) {
  return list->len;
}

// Null-terminate the value to be inserted, as TokyoCabinet does
void *tclistpack(void *buf, uint32_t size) {
  void *val = malloc(size+1);
  if (val == NULL) {
    return NULL;
  }
  memset(val, 0 , size+1);
  memcpy(val, buf, size);
  return val;
}

// Append a new element at the end of the list
void tclistpush(TCLIST *list, const void *ptr, int size) {
  TCLISTNODE *cur = list->head;
  for (int i=0; i < list->len; i++) {
    cur = cur->next;
  }
  TCLISTNODE *newElem = malloc(sizeof(TCLISTNODE));
  if (newElem == NULL) {
    fprintf(stderr, "Failed to malloc list node (tclistpush)\n");
    exit(0);   
  }

  newElem->val = tclistpack(ptr, size);
  if (newElem->val == NULL) {
    free(newElem);
    fprintf(stderr, "Failed to malloc list node contents\n");
    exit(0);   
  }

  newElem->size = size;
  cur->next = newElem;
  list->len += 1;
}

// Remove the element at index, and return a pointer to it.
// The caller is responsible for freeing the returned element.
// Returns NULL if index is out of bounds.
void *tclistremove(TCLIST *list, int index, int *sp) {
  if (index < 0 || index >= list->len) {
    return NULL;
  }
  TCLISTNODE *cur = list->head;
  for (int i=0; i < index; i++) {
    cur = cur->next;
  }
  
  TCLISTNODE *removed = cur->next;
  cur->next = removed->next;
  void* val = removed->val;
  *sp = removed->size;
  free(removed);
  list->len -= 1;
  return val;
}

// Overwrite the existing value at the given index. Noop if the index is out of bounds
void tclistover(TCLIST *list, int index, const void *ptr, int size) {
  if (index < 0 || index >= list->len) {
    // Out of bounds, do nothing
    return;
  }
  TCLISTNODE *cur = list->head;
  for (int i=0; i <= index; i++) {
    cur = cur->next;
  }

  if (cur->val != NULL) {
    free(cur->val);
  }
 
  cur->val = tclistpack(ptr, size);
  if (cur->val == NULL) {
    fprintf(stderr, "Failed to malloc list node (tclistover)\n");
    exit(0);
  }
  cur->size = size; 
}

// Get the value at index, or NULL if the index is out of bounds
const void *tclistval(const TCLIST *list, int index, int *sp) {
  if (index >= list->len) {
    // Out of bounds, do nothing
    return NULL;
  }
  TCLISTNODE *cur = list->head;
  for (int i=0; i <= index; i++) {
    cur = cur->next;
  }
  *sp = cur->size;
  return cur->val;
}

// Delete the entire list, freeing all elements
void tclistdel(TCLIST *list) {
  TCLISTNODE *cur = list->head;
  for (int i=0; i <= list->len; i++) {
    if (cur->val) {
      free(cur->val);
    }
    TCLISTNODE *removed = cur;
    cur = cur->next;
    free(removed);
  }
  free(list);
}
