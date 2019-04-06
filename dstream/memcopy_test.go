package dstream

import "testing"

func TestMemCopy1(t *testing.T) {

	x1 := []interface{}{
		[]float64{0, 1, 1},
		[]float64{0, 0, 1, 0},
	}
	x2 := []interface{}{
		[]float64{1, 1, 1},
		[]float64{1, 1, 1, 1},
	}
	x3 := []interface{}{
		[]float64{4, 1, -1},
		[]float64{3, 5, -5, 3},
	}
	x4 := []interface{}{
		[]float64{1, 1, 1},
		[]float64{2, 2, 1, 1},
	}
	x5 := []interface{}{
		[]float64{1, 1, 1},
		[]float64{2, 2, 2, 3},
	}
	dat := [][]interface{}{x1, x2, x3, x4, x5}
	na := []string{"x1", "x2", "x3", "x4", "x5"}
	da := NewFromArrays(dat, na)

	db := MemCopy(da)

	if !EqualReport(da, db, true) {
		t.Fail()
	}
}
