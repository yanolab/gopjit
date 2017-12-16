package gopjit

import "testing"

func TestJIT(t *testing.T) {
	jit := NewJIT()

	addFuncSrc := `package main
	func F0(v0, v1 int) int {
		return v0 + v1
	}`
	sym, err := jit.BuildSrc(addFuncSrc)
	if err != nil {
		t.Fatal("build add func")
	}

	f := sym.(func(int, int) int)
	if v := f(1, 2); v != 3 {
		t.Fatalf("unexpected value: %d", v)
	}
}
