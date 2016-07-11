package main

func dummy(i int32) {
	var fp func(zorb *int32, glarch *int64) (bool, interface{})
	fpc := fp
	t := fpc
	if i > 0 {
		fpc = t
		t = fp
	}
}
