package ms

import (
	"encoding/json"
	"fmt"
)

func Example() {
	var mm MultiModel
	var m Model
	mm.Models = append(mm.Models, m)

	var p Part
	p.Base = 0
	mm.Parts = append(mm.Parts, p)

	b, err := json.MarshalIndent(mm, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(b))
	// Output:
	// {
	//   "Models": [
	//     {
	//       "Name": "",
	//       "Coordinates": null,
	//       "Elements": null
	//     }
	//   ],
	//   "Parts": [
	//     {
	//       "Name": "",
	//       "Elements": null,
	//       "Base": 0
	//     }
	//   ]
	// }
}
