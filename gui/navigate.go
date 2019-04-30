package gui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type navigate struct {
	*tview.TextView
	keybindings map[string]string
}

func newNavigate() *navigate {
	return &navigate{
		TextView: tview.NewTextView().SetTextColor(tcell.ColorYellow),
		keybindings: map[string]string{
			"images":     " p: pull image, i: import image, s: save image, Ctrl+l: load image, f: search image, /: filter d: remove image,\n c: create container, Enter: inspect image, Ctrl+r: refresh images list",
			"containers": " e: export container, c: commit container, /: filter, Ctrl+c: exec container cmd\n u: start container, s: stop container, d: remove container, Enter: inspect container, Ctrl+r: refresh container list, Ctrl+l: show container logs",
			"networks":   " d: remove network, Enter: inspect network, /: filter",
			"volumes":    " c: create volume, d: remove volume\n /: filter, Enter: inspect volume, Ctrl+r: refresh volume list",
		},
	}
}

func (n *navigate) update(panel string) {
	n.SetText(n.keybindings[panel])
}
