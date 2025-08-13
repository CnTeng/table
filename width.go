package table

import "slices"

type widths []int

type widthDiff struct {
	idx  int
	diff int
}

func (ws widths) sum() int {
	var sum int
	for _, w := range ws {
		sum += w
	}
	return sum
}

func (ws widths) expand(maxWidths []int, extra int) {
	wd := []*widthDiff{}
	for i, w := range ws {
		wd = append(wd, &widthDiff{idx: i, diff: maxWidths[i] - w})
	}

	slices.SortFunc(wd, func(a, b *widthDiff) int {
		return a.diff - b.diff
	})

	for _, w := range wd {
		expended := min(w.diff, extra)
		ws[w.idx] += expended
		extra -= expended
		if extra == 0 {
			return
		}
	}
}

func (ws widths) shrink(minWidths []int, extra int) {
	wd := []*widthDiff{}
	for i, w := range ws {
		wd = append(wd, &widthDiff{idx: i, diff: w - minWidths[i]})
	}

	slices.SortFunc(wd, func(a, b *widthDiff) int {
		return b.diff - a.diff
	})

	for _, w := range wd {
		shrunk := min(w.diff, extra)
		ws[w.idx] -= shrunk
		extra -= shrunk
		if extra == 0 {
			return
		}
	}
}
