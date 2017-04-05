#include "tree.h"

/* Do everything once so we can valgrind and ensure there are no memory leaks */

void main () {

  uint32_t key = 123;
  const char *val = "abcdefkrnglrg";

  // Make a new tree
  TCTREE *tree = tctreenew();

  // Put an integer key
  tctreeput(tree, &key, sizeof(key), val, strlen(val));
  
  // Put a different key
  key = 122;
  tctreeput(tree, &key, sizeof(key), val, strlen(val));
  
  // Put the same key twice
  tctreeput(tree, &key, sizeof(key), val, strlen(val));

  // Put the same key but keep the old value
  tctreeputkeep(tree, &key, sizeof(key), val, strlen(val));

  // Get back a value
  int size;
  void *newVal = tctreeget(tree, &key, sizeof(key), &size);
  printf("Got value %s\n", newVal);

  // Create an iterator
  tctreeiterinit(tree);

  // Walk the tree
  tctreeiternext(tree, &size);
  tctreeiternext(tree, &size);
  tctreeiternext(tree, &size);

  // Clear the tree
  tctreeclear(tree);

  // Put one value back in the tree to make sure it's freed on delete 
  tctreeput(tree, &key, sizeof(key), val, strlen(val));

  // Delete the tree
  tctreedel(tree);

  // Make a list
  TCLIST *list = tclistnew();

  // Push a few times
  tclistpush(list, &key, sizeof(key));
  key += 1;
  tclistpush(list, &key, sizeof(key));
  key += 1;
  tclistpush(list, &key, sizeof(key));
 
  // Overwrite an existing element
  tclistover(list, 1, &key, sizeof(key));

  // Get a value
  tclistval(list, 2, &size);

  // Remove some values
  newVal = tclistremove(list, 2, &size);
  free(newVal);
  newVal = tclistremove(list, 0, &size);
  free(newVal);

  // Free the whole list
  tclistdel(list);
}
