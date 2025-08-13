package table

import (
	"reflect"
	"testing"
)

func TestWidthsSum(t *testing.T) {
	ws := widths{1, 2, 3, 4}
	want := 10
	got := ws.sum()
	if got != want {
		t.Errorf("sum() = %d, want %d", got, want)
	}
}

func TestWidthsExpand(t *testing.T) {
	ws := widths{3, 2, 4}
	maxWidths := []int{5, 5, 6}
	extra := 4
	want := widths{5, 2, 6}
	ws.expand(maxWidths, extra)
	if !reflect.DeepEqual(ws, want) {
		t.Errorf("expand() = %v, want %v", ws, want)
	}
}

func TestWidthsExpand_ExtraExceedsDiff(t *testing.T) {
	ws := widths{3, 2, 4}
	maxWidths := []int{5, 5, 6}
	extra := 10
	want := widths{5, 5, 6}
	ws.expand(maxWidths, extra)
	if !reflect.DeepEqual(ws, want) {
		t.Errorf("expand() = %v, want %v", ws, want)
	}
}

func TestWidthsShrink(t *testing.T) {
	ws := widths{6, 5, 4}
	minWidths := []int{3, 3, 2}
	extra := 4
	want := widths{3, 4, 4}
	ws.shrink(minWidths, extra)
	if !reflect.DeepEqual(ws, want) {
		t.Errorf("shrink() = %v, want %v", ws, want)
	}
}

func TestWidthsShrink_ExtraExceedsDiff(t *testing.T) {
	ws := widths{5, 5}
	minWidths := []int{2, 2}
	extra := 10
	want := widths{2, 2}
	ws.shrink(minWidths, extra)
	if !reflect.DeepEqual(ws, want) {
		t.Errorf("shrink() = %v, want %v", ws, want)
	}
}
