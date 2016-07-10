
package demangler

type S2 struct {
	b bool
	bsl []bool
	f32 float32
	f64 float64
	u32 uint32
	u64 uint64
	i32 int32
	i64 int64
	pi32 *int32
	fptr func()
	c128 complex128
	ba [32]byte
}

func Dummy(p *S2) S2 {
	var aa [8]int32
	var x struct {
		b bool
		bsl []bool
		f32 float32
		f64 float64
		u32 uint32
		u64 uint64
		i32 int32
		i64 int64
		pi32 *int32
		fptr func()
		c128 complex128
		ba [32]byte
	}
	x.f32 = p.f32
	p.f64 = x.f64
	p.ba[3] = byte(p.i32)
	aa[p.i32] = int32(p.ba[3])
	p.ba[4] = byte(aa[3])
	p.b = true

	return *p
}
