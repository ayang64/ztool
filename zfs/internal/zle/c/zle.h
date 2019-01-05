/* zle.c */
size_t zle_compress(void *s_start, void *d_start, size_t s_len, size_t d_len, int n);
int zle_decompress(void *s_start, void *d_start, size_t s_len, size_t d_len, int n);
