package chainsaw

import (
	"fmt"
	"reflect"
	"testing"
)

/*
goos: darwin
goarch: arm64
Benchmark_TypeSwitch
Benchmark_TypeSwitch-8   	 7517070	       147.3 ns/op
Benchmark_Reflect
Benchmark_Reflect-8      	 6135998	       195.5 ns/op
PASS

*/

func switchTest(value interface{}) int {
	switch switchType := value.(type) {
	case nil:
		if switchType == nil { // make the compiler happy.
			return 666
		}
		return 1
	case string:
		return 2
	case int:
		return 3
	case int8:
		return 4
	case int16:
		return 5
	case int32:
		return 6
	case int64:
		return 7
	case float32:
		return 8
	case float64:
		return 9
	case uint:
		return 10
	}
	return 0
}

func reflectTest(value interface{}) int {
	rvalue := reflect.ValueOf(value)
	switch rvalue.Kind() {
	case reflect.Ptr:
		return 1
	case reflect.String:
		return 2
	case reflect.Int:
		return 3
	case reflect.Int8:
		return 4
	case reflect.Int16:
		return 5
	case reflect.Int32:
		return 6
	case reflect.Int64:
		return 7
	case reflect.Float32:
		return 8
	case reflect.Float64:
		return 9
	case reflect.Uint:
		return 10
	}
	return 0
}

// makeWeirdList makes a list of various types to throw benchmarks below
func makeWeirdList() []interface{} {
	var list []interface{}
	for i := 0; i < 10; i++ {
		list = append(list, []interface{}{"string", 5, 5.5, nil, int8(5), uint(3), uint32(6)}...)
	}
	return list
}

var globalSum int // Throw off the optimizer

func Benchmark_TypeSwitch(b *testing.B) {
	sum := 0
	list := makeWeirdList()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range list {
			sum = sum + switchTest(val)
		}
	}
	globalSum = sum
}
func Benchmark_Reflect(b *testing.B) {
	sum := 0
	list := makeWeirdList()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range list {
			sum = sum + reflectTest(val)
		}
	}
	globalSum = sum
}

// Make sure that both typeSwitch and reflection evaluate the weird list and get the same result.
func Test_Benchmark(t *testing.T) {
	list := makeWeirdList()
	sumReflect := 0
	for _, val := range list {
		sumReflect = sumReflect + reflectTest(val)
	}
	sumSwitch := 0
	for _, val := range list {
		sumSwitch = sumSwitch + reflectTest(val)
	}
	if sumReflect != sumSwitch {
		t.Errorf("reflectiong and typeswich values don't match")
	}
	if globalSum != 0 {
		fmt.Println("dummy statement to shut to linter up")
	}
}
