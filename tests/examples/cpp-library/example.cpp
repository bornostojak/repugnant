// ?rPg: C++ bounds check
// ?~ Keep indexing safe before reading the vector.
int at(const int* values, int size, int index) {
  if (index < 0 || index >= size) return -1;
  return values[index];
}
// !rPg
