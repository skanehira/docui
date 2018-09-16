package panel

func NewCommitContainerPanel(ix, iy, iw, ih int) Items {
	names := []string{
		"Container",
		"Repository",
		"Tag",
	}

	var items Items

	x := iw / 8                                          // label start position
	w := x + 12                                          // label length
	bh := 2                                              // input box height
	th := ((ih - iy) - len(names)*bh) / (len(names) + 1) // to next input height
	y := th
	h := 0

	for i, name := range names {
		if i != 0 {
			y = items[i-1].Label[names[i-1]].h + th
		}
		h = y + bh

		x1 := w + 1
		w1 := iw - (x + ix)

		item := Item{
			Label: map[string]Position{name: {x, y, w, h}},
			Input: map[string]Position{name + "Input": {x1, y, w1, h}},
		}

		items = append(items, item)
	}

	return items
}
