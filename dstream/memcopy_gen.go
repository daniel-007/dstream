// GENERATED CODE, DO NOT EDIT

package dstream

import (
	"time"
)

// MemCopy returns a Dstream that copies the provided Dstream into
// in-memory storage.
func MemCopy(data Dstream) Dstream {

	nvar := data.NumVar()
	bdata := make([][]interface{}, nvar)

	data.Reset()
	for data.Next() {
		for j := 0; j < nvar; j++ {
			var y interface{}
			v := data.GetPos(j)
			switch v := v.(type) {
			case []string:
				z := make([]string, len(v))
				copy(z, v)
				y = z
			case []time.Time:
				z := make([]time.Time, len(v))
				copy(z, v)
				y = z
			case []uint8:
				z := make([]uint8, len(v))
				copy(z, v)
				y = z
			case []uint16:
				z := make([]uint16, len(v))
				copy(z, v)
				y = z
			case []uint32:
				z := make([]uint32, len(v))
				copy(z, v)
				y = z
			case []uint64:
				z := make([]uint64, len(v))
				copy(z, v)
				y = z
			case []int8:
				z := make([]int8, len(v))
				copy(z, v)
				y = z
			case []int16:
				z := make([]int16, len(v))
				copy(z, v)
				y = z
			case []int32:
				z := make([]int32, len(v))
				copy(z, v)
				y = z
			case []int64:
				z := make([]int64, len(v))
				copy(z, v)
				y = z
			case []float32:
				z := make([]float32, len(v))
				copy(z, v)
				y = z
			case []float64:
				z := make([]float64, len(v))
				copy(z, v)
				y = z
			}
			bdata[j] = append(bdata[j], y)
		}
	}

	da := &DataFrame{
		data: bdata,
		xform: xform{
			names: data.Names(),
		},
	}

	da.init()

	return da
}
