#include <stdio.h>
#include <string.h>
#include <stdlib.h>

#include "zle.h"

int
main(int argc, char *argv[]) {
	char *src = "all good things come to an end.";
	size_t len = strlen(src)+1;
	void *dst = malloc(strlen(src)+1);

	int b = zle_compress(src, dst, len, len, 64);

	printf("len = %zu, b = %d\n", len, b);
}
