package envconfig

import (
	"fmt"
	"os"
	"testing"
)

// Float64SliceEmpty keeps linter happy.
func Float64SliceEmpty() {
}

// Float64Slice keeps linter happy.
func Float64Slice() {
}

func ExampleFloat64SliceEmpty() {

	os.Setenv("SLICE", "")

	env := NewSimple("ExampleFloat64SliceEmpty")

	fmt.Println(env.Float64Slice("SLICE", []float64{1, 2}))
	// Output: [1 2]
}

func ExampleFloat64Slice() {

	os.Setenv("SLICE", " -4.4 , 1.1 , 2.2 , 3.3 ")

	env := NewSimple("ExampleFloat64Slice")

	fmt.Println(env.Float64Slice("SLICE", []float64{1, 2}))
	// Output: [-4.4 1.1 2.2 3.3]
}

func TestFloat64(t *testing.T) {
	t.Setenv("VALUE", "13")

	env := NewSimple("TestFloat64")

	v := env.Float64("VALUE", 1)

	if v != 13 {
		t.Errorf("expected=13 got=%f", v)
	}
}

func TestInt(t *testing.T) {
	t.Setenv("VALUE", "13")

	env := NewSimple("TestFloat64")

	v := env.Int("VALUE", 1)

	if v != 13 {
		t.Errorf("expected=13 got=%d", v)
	}
}

func TestInt64(t *testing.T) {
	t.Setenv("VALUE", "13")

	env := NewSimple("TestFloat64")

	v := env.Int64("VALUE", 1)

	if v != 13 {
		t.Errorf("expected=13 got=%d", v)
	}
}

func TestUint64(t *testing.T) {
	t.Setenv("VALUE", "13")

	env := NewSimple("TestFloat64")

	v := env.Uint64("VALUE", 1)

	if v != 13 {
		t.Errorf("expected=13 got=%d", v)
	}
}
