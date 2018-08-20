#include <stdio.h>
#include <string.h>
#include <stdlib.h>

#include "zle.h"

void
dumpslice(char *s, size_t l) {
	printf("[]byte{ ");
	for (size_t i = 0; i < l; i++) {
		printf("%d, ", s[i]);
	}
	printf("}\n");
}

int
main(
	int argc __attribute__ ((unused)),
	char *argv[] __attribute__ ((unused))) {

	char src[] = {0, 0, 0, 0, 12, 0, 0, 0, 0, 0, 255, 0, 0, 2, 8, 10, 0, 0, 0, 128, 0, 0, 0, 0, 8};

	size_t	srclen = sizeof(src);
	void		*cmp = malloc(srclen),
					*dcm = malloc(srclen);

	size_t	cmplen =   zle_compress(src, cmp, srclen, srclen, 2),
					dcmlen = zle_decompress(cmp, dcm, cmplen, srclen, 2);

	printf("// srclen = %zu, dcmlen = %zu\n", srclen, dcmlen);
	dumpslice(src, srclen);
	dumpslice(cmp, cmplen);
	dumpslice(dcm, srclen);
}
