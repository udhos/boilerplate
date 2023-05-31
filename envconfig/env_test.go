package envconfig

import (
	"fmt"
	"os"
)

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
