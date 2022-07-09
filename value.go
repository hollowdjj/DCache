package DCache

//缓存k-v中的值类型，只能是byte数组或者字符串。
//若b != nil，那么使用的是b，否则为s
type Value struct {
	b []byte
	s string
}

//返回数据的大小
func (v *Value) Len() int {
	if v.b != nil {
		return len(v.b)
	}
	return len(v.s)
}

//以字符串形式返回数据
func (v *Value) String() string {
	if v.b != nil {
		return string(v.b)
	}
	return v.s
}

//以byte切片的形式返回数据
func (v *Value) ByteSlice() []byte {
	if v.b != nil {
		return copyByteSlice(v.b)
	}
	return []byte(v.s)
}

//拷贝byte切片
func copyByteSlice(b []byte) []byte {
	res := make([]byte, len(b))
	copy(res, b)
	return res
}
