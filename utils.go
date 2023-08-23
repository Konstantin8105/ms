package ms

import (
	"sort"
	"strconv"
	"strings"
)

func angle360(a float64) float64 {
	if 360.0 < a {
		return a - 360.0
	}
	if a < -360.0 {
		return a + 360.0
	}
	return a
}

func isOne(f func() []uint) (value uint, ok bool) {
	b := f()
	if len(b) != 1 {
		return
	}
	return b[0], true
}

func clearValue(str *string) {
	*str = strings.ReplaceAll(*str, "[", " ")
	*str = strings.ReplaceAll(*str, "]", " ")
}

func convertUint(str string) (ids []uint) {
	clearValue(&str)
	if str == "NONE" {
		return
	}
	fs := strings.Fields(str)
	for i := range fs {
		u, err := strconv.ParseUint(fs[i], 10, 64)
		if err != nil {
			AddInfo("convertUint error: %v", err)
			continue
		}
		ids = append(ids, uint(u))
	}
	return
}

// compress
//
//	Example:
//	from: 1 2 3 4 5 6 7
//	to  : 1 TO 7
//
//	from: 1 2 3 4 6 7
//	to  : 1 TO 4 6 7
//
//	from: 1 3 5 7
//	to  : 1 3 5 7
// func compress(ids []uint) (res string) {
// 	ids := uniqUint(ids)
// 	for i, id := range ids {
// 		res += fmt.Sprintf(" %d ", id)
// 	}
// 	return
// }

func uniqUint(ids []uint) (res []uint) {
	sort.SliceStable(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	res = make([]uint, 0, len(ids))
	for i, id := range ids {
		if i == 0 {
			res = append(res, id)
			continue
		}
		if res[len(res)-1] == ids[i] {
			continue
		}
		res = append(res, id)
	}
	return
}
