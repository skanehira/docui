package panel

import (
	"fmt"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

type NetworkList struct {
	*Gui
	name string
	Position
	Networks []*Network
	Data     map[string]interface{}
	filter   string
}

type Network struct {
	ID         string `tag:"ID" len:"min:0.1 max:0.2"`
	Name       string `tag:"NAME" len:"min:0.1 max:0.3"`
	Driver     string `tag:"DRIVER" len:"min:0.1 max:0.1"`
	Scope      string `tag:"SCOPE" len:"min:0.1 max:0.1"`
	Containers string `tag:"CONTAINERS" len:"min:0.1 max:0.3"`
}

func NewNetworkList(gui *Gui, name string, x, y, w, h int) *NetworkList {
	n := &NetworkList{
		Gui:      gui,
		name:     name,
		Position: Position{x, y, w, h},
		Data:     make(map[string]interface{}),
	}

	return n
}

func (n *NetworkList) Name() string {
	return n.name
}

func (n *NetworkList) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
		return
	case key == gocui.KeyArrowRight:
		v.MoveCursor(+1, 0, false)
		return
	}

	n.filter = ReadLine(v, nil)

	if v, err := n.View(n.name); err == nil {
		n.GetNetworkList(v)
	}
}

func (n *NetworkList) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := g.SetView(NetworkListHeaderPanel, n.x, n.y, n.w, n.h); err != nil {
		if err != gocui.ErrUnknownView {
			panic(err)
		}

		v.Wrap = true
		v.Frame = true
		v.Title = v.Name()
		v.FgColor = gocui.AttrBold | gocui.ColorWhite
		common.OutputFormatedHeader(v, &Network{})
	}

	// set scroll panel
	v, err := g.SetView(n.name, n.x, n.y+1, n.w, n.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Wrap = true
		v.FgColor = gocui.ColorYellow
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
		v.SetOrigin(0, 0)
		v.SetCursor(0, 0)
	}

	n.SetKeyBinding()

	//  monitoring container status interval 5s
	go func() {
		for {
			n.Update(func(g *gocui.Gui) error {
				n.Refresh(g, v)
				return nil
			})
			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}

func (n *NetworkList) Refresh(g *gocui.Gui, v *gocui.View) error {
	n.Update(func(g *gocui.Gui) error {
		v, err := n.View(n.name)
		if err != nil {
			panic(err)
		}
		n.GetNetworkList(v)
		return nil
	})

	return nil
}

func (n *NetworkList) SetKeyBinding() {
	n.SetKeyBindingToPanel(n.name)

	if err := n.SetKeybinding(n.name, 'j', gocui.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := n.SetKeybinding(n.name, 'k', gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := n.SetKeybinding(n.name, gocui.KeyCtrlR, gocui.ModNone, n.Refresh); err != nil {
		panic(err)
	}
	if err := n.SetKeybinding(n.name, 'f', gocui.ModNone, n.Filter); err != nil {
		panic(err)
	}
	if err := n.SetKeybinding(n.name, 'o', gocui.ModNone, n.Detail); err != nil {
		panic(err)
	}
	if err := n.SetKeybinding(n.name, gocui.KeyEnter, gocui.ModNone, n.Detail); err != nil {
		panic(err)
	}
	if err := n.SetKeybinding(n.name, 'd', gocui.ModNone, n.RemoveNetwork); err != nil {
		panic(err)
	}
}

func (n *NetworkList) selected() (*Network, error) {
	v, _ := n.View(n.name)
	_, cy := v.Cursor()
	_, oy := v.Origin()

	index := oy + cy
	length := len(n.Networks)

	if index >= length {
		return nil, common.NoNetwork
	}

	return n.Networks[index], nil
}

func (n *NetworkList) Filter(g *gocui.Gui, nv *gocui.View) error {
	isReset := false
	closePanel := func(g *gocui.Gui, v *gocui.View) error {
		if isReset {
			n.filter = ""
		} else {
			nv.SetCursor(0, 0)
			n.filter = ReadLine(v, nil)
		}
		if v, err := n.View(n.name); err == nil {
			n.GetNetworkList(v)
		}

		if err := g.DeleteView(v.Name()); err != nil {
			panic(err)
		}

		g.DeleteKeybindings(v.Name())
		n.SwitchPanel(n.name)
		return nil
	}

	reset := func(g *gocui.Gui, v *gocui.View) error {
		isReset = true
		return closePanel(g, v)
	}

	if err := n.NewFilterPanel(n, reset, closePanel); err != nil {
		panic(err)
	}

	return nil
}

func (n *NetworkList) GetNetworkList(v *gocui.View) {
	v.Clear()
	n.Networks = make([]*Network, 0)

	var keys []string
	tmpMap := make(map[string]*Network)

	for _, network := range n.Docker.Networks() {
		if n.filter != "" {
			if strings.Index(strings.ToLower(network.Name), strings.ToLower(n.filter)) == -1 {
				continue
			}
		}

		var containers string
		net, err := n.Docker.NetworkInfo(network.ID)
		if err != nil {
			n.ErrMessage(err.Error(), n.name)
			return
		}

		for _, endpoint := range net.Containers {
			containers += fmt.Sprintf("%s ", endpoint.Name)
		}

		tmpMap[network.ID[:12]] = &Network{
			ID:         network.ID,
			Name:       network.Name,
			Driver:     network.Driver,
			Scope:      network.Scope,
			Containers: containers,
		}

		keys = append(keys, network.ID[:12])

	}

	for _, key := range common.SortKeys(keys) {
		net := tmpMap[key]
		common.OutputFormatedLine(v, net)
		n.Networks = append(n.Networks, net)
	}
}

func (n *NetworkList) Detail(g *gocui.Gui, v *gocui.View) error {
	selected, err := n.selected()
	if err != nil {
		n.ErrMessage(err.Error(), n.name)
		return nil
	}

	net, err := n.Docker.NetworkInfo(selected.ID)
	if err != nil {
		n.ErrMessage(err.Error(), n.name)
		return nil
	}

	n.PopupDetailPanel(g, v)

	v, err = g.View(DetailPanel)
	if err != nil {
		panic(err)
	}

	v.Clear()
	v.SetOrigin(0, 0)
	v.SetCursor(0, 0)

	fmt.Fprint(v, common.StructToJson(net))
	return nil
}

func (n *NetworkList) RemoveNetwork(g *gocui.Gui, v *gocui.View) error {
	selected, err := n.selected()
	if err != nil {
		n.ErrMessage(err.Error(), n.name)
		return nil
	}

	n.ConfirmMessage("Are you sure you want to remove this network?", n.name, func() error {
		defer n.Refresh(g, v)
		if err := n.Docker.RemoveNetwork(selected.ID); err != nil {
			n.ErrMessage(err.Error(), n.name)
			return nil
		}

		return nil
	})

	return nil
}
