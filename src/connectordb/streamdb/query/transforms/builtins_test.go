package transforms

import (
	"connectordb/streamdb/datastream"
	"fmt"
)

func ExampleSumTransform() {
	summer, _ := ParseTransform("sum()")
	result, err := summer(&datastream.Datapoint{Data: 1})

	// total is one so far
	fmt.Printf("%v, %v\n", result.Data, err)

	// total will be five
	result, err = summer(&datastream.Datapoint{Data: 4})
	fmt.Printf("%v, %v\n", result.Data, err) // 1 + 4

	// total will be seven
	result, err = summer(&datastream.Datapoint{Data: 2})
	fmt.Printf("%v, %v\n", result.Data, err) // 1 + 4 + 2

	// total will be 9
	result, err = summer(&datastream.Datapoint{Data: 2})
	fmt.Printf("%v, %v\n", result.Data, err) // 1 + 4 + 2 + 2

	// total will be 11
	result, err = summer(&datastream.Datapoint{Data: 2})
	fmt.Printf("%v, %v\n", result.Data, err) // 1 + 4 + 2 + 2 + 2

	// Output:
	// 1, <nil>
	// 5, <nil>
	// 7, <nil>
	// 9, <nil>
	// 11, <nil>
}
