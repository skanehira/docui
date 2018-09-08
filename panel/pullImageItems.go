package panel

func NewPullImageItems() Items {
	x0 := 2
	w0 := 17
	x1 := 16
	w1 := 66

	return Items{
		Item{
			Label: map[string]Position{"Name": {x0, 2, w0, 4}},
			Input: map[string]Position{"NameInput": {x1, 2, w1, 4}},
		},
	}
}
