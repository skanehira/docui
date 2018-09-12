package panel

func NewCreateContainerItems() Items {
	x0 := 2
	w0 := 17
	x1 := 16
	w1 := 66

	return Items{
		Item{
			Label: map[string]Position{"Name": {x0, 2, w0, 4}},
			Input: map[string]Position{"NameInput": {x1, 2, w1, 4}},
		},
		Item{
			Label: map[string]Position{"HostPort": {x0, 6, w0, 8}},
			Input: map[string]Position{"HostPortInput": {x1, 6, w1, 8}},
		},
		Item{
			Label: map[string]Position{"Port": {x0, 10, w0, 12}},
			Input: map[string]Position{"PortInput": {x1, 10, w1, 12}},
		},
		Item{
			Label: map[string]Position{"HostVolume": {x0, 14, w0, 16}},
			Input: map[string]Position{"HostVolumeInput": {x1, 14, w1, 16}},
		},
		Item{
			Label: map[string]Position{"Volume": {x0, 18, w0, 20}},
			Input: map[string]Position{"VolumeInput": {x1, 18, w1, 20}},
		},
		Item{
			Label: map[string]Position{"Image": {x0, 22, w0, 24}},
			Input: map[string]Position{"ImageInput": {x1, 22, w1, 24}},
		},
		Item{
			Label: map[string]Position{"Env": {x0, 26, w0, 28}},
			Input: map[string]Position{"EnvInput": {x1, 26, w1, 28}},
		},
	}
}
