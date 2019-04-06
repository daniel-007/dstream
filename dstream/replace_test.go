package dstream

import "testing"

func TestReplace1(t *testing.T) {

	x0 := []interface{}{
		[]float64{5, 8, 1},
		[]float64{2, 1, 3, 1},
	}
	x1 := []interface{}{
		[]float64{5, 8, 1},
		[]float64{2, 1, 3, 1},
	}
	x2 := []interface{}{
		[]float64{2, 3, 4},
		[]float64{1, 2, 1, 1},
	}
	da := NewFromArrays([][]interface{}{x0, x1, x2}, []string{"x0", "x1", "x2"})

	v := []float64{2, 3, 4, 5, 6, 7, 8}
	dx := Replace(da, "x2", v)

	x2 = []interface{}{
		[]float64{2, 3, 4},
		[]float64{5, 6, 7, 8},
	}
	db := NewFromArrays([][]interface{}{x0, x1, x2}, []string{"x0", "x1", "x2"})

	for j := 0; j < 2; j++ {
		if !EqualReport(dx, db, true) {
			t.Fail()
		}
		dx.Reset()
		dx = MemCopy(dx)
	}
}
