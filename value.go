package cache

//value type of cache, it can only be []byte
type Value struct {
	b []byte
}

//return len(v.b)
func (v *Value) Len() int {
	return len(v.b)
}

//return as string
func (v *Value) String() string {
	return string(v.b)
}

//return a copy of v.b
func (v *Value) ByteSlice() []byte {
	return copyByteSlice(v.b)
}

//copy byte slice
func copyByteSlice(b []byte) []byte {
	res := make([]byte, len(b))
	copy(res, b)
	return res
}
