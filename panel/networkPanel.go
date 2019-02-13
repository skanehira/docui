package panel

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

// NetworkList network list panel.
type NetworkList struct {
	*Gui
	name string
	Position
	Networks []*Network
	Data     map[string]interface{}
	filter   string
	stop     chan int
}

// Network network info.
type Network struct {
	ID         string `tag:"ID" len:"min:0.1 max:0.2"`
	Name       string `tag:"NAME" len:"min:0.1 max:0.3"`
	Driver     string `tag:"DRIVER" len:"min:0.1 max:0.1"`
	Scope      string `tag:"SCOPE" len:"min:0.1 max:0.1"`
	Containers string `tag:"CONTAINERS" len:"min:0.1 max:0.3"`
}

// NewNetworkList create new network list panel.
func NewNetworkList(gui *Gui, name string, x, y, w, h int) *NetworkList {
	n := &NetworkList{
		Gui:      gui,
		name:     name,
		Position: Position{x, y, w, h},
		Data:     make(map[string]interface{}),
		stop:     make(chan int, 1),
	}

	return n
}

// Name return panel name.
func (n *NetworkList) Name() string {
	return n.name
}

// Edit filtering network list
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

	n.filter = ReadViewBuffer(v)

	if v, err := n.View(n.name); err == nil {
		n.GetNetworkList(v)
	}
}

// SetView set up network list panel.
func (n *NetworkList) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := common.SetViewWithValidPanelSize(g, NetworkListHeaderPanel, n.x, n.y, n.w, n.h); err != nil {
		if err != gocui.ErrUnknownView {
			common.Logger.Error(err)
			return err
		}

		v.Wrap = true
		v.Frame = true
		v.Title = v.Name()
		v.FgColor = gocui.AttrBold | gocui.ColorWhite
		common.OutputFormattedHeader(v, &Network{})
	}

	// set scroll panel
	v, err := common.SetViewWithValidPanelSize(g, n.name, n.x, n.y+1, n.w, n.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			common.Logger.Error(err)
			return err
		}
		v.Frame = false
		v.Wrap = true
		v.FgColor = gocui.ColorYellow
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
		v.SetOrigin(0, 0)
		v.SetCursor(0, 0)

		n.GetNetworkList(v)
	}

	n.SetKeyBinding()

	// monitoring network status.
	go n.Monitoring(n.stop, n.Gui.Gui, v)
	return nil
}

// Monitoring monitoring image list.
func (n *NetworkList) Monitoring(stop chan int, g *gocui.Gui, v *gocui.View) {
	common.Logger.Info("monitoring network list start")
	ticker := time.NewTicker(5 * time.Second)

LOOP:
	for {
		select {
		case <-ticker.C:
			n.Update(func(g *gocui.Gui) error {
				return n.Refresh(g, v)
			})
		case <-stop:
			ticker.Stop()
			break LOOP
		}
	}
	common.Logger.Info("monitoring network list stop")
}

// CloseView close panel
func (n *NetworkList) CloseView() {
	// stop monitoring
	n.stop <- 0
	close(n.stop)
}

// Refresh update network info
func (n *NetworkList) Refresh(g *gocui.Gui, v *gocui.View) error {
	n.Update(func(g *gocui.Gui) error {
		v, err := n.View(n.name)
		if err != nil {
			common.Logger.Error(err)
			return nil
		}
		n.GetNetworkList(v)
		return nil
	})

	return nil
}

// SetKeyBinding set key bind to this panel.
func (n *NetworkList) SetKeyBinding() {
	n.SetKeyBindingToPanel(n.name)

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

// selected return selected network info
func (n *NetworkList) selected() (*Network, error) {
	v, _ := n.View(n.name)
	_, cy := v.Cursor()
	_, oy := v.Origin()

	index := oy + cy
	length := len(n.Networks)

	if index >= length {
		return nil, common.ErrNoNetwork
	}

	return n.Networks[index], nil
}

// Filter filtering network list
func (n *NetworkList) Filter(g *gocui.Gui, nv *gocui.View) error {
	isReset := false
	closePanel := func(g *gocui.Gui, v *gocui.View) error {
		if isReset {
			n.filter = ""
		} else {
			nv.SetCursor(0, 0)
			n.filter = ReadViewBuffer(v)
		}
		if v, err := n.View(n.name); err == nil {
			n.GetNetworkList(v)
		}

		if err := g.DeleteView(v.Name()); err != nil {
			common.Logger.Error(err)
			return nil
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
		common.Logger.Error(err)
		return nil
	}

	return nil
}

// GetNetworkList return network info
func (n *NetworkList) GetNetworkList(v *gocui.View) {
	v.Clear()
	n.Networks = make([]*Network, 0)

	networks, err := n.Docker.Networks(types.NetworkListOptions{})
	if err != nil {
		common.Logger.Error(err)
		return
	}

	keys := make([]string, 0, len(networks))
	tmpMap := make(map[string]*Network)

	for _, network := range networks {
		if n.filter != "" {
			if strings.Index(strings.ToLower(network.Name), strings.ToLower(n.filter)) == -1 {
				continue
			}
		}

		var containers string

		net, err := n.Docker.InspectNetwork(network.ID)
		if err != nil {
			common.Logger.Error(err)
			continue
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
		common.OutputFormattedLine(v, net)
		n.Networks = append(n.Networks, net)
	}
}

// Detail display detail the specified network
func (n *NetworkList) Detail(g *gocui.Gui, v *gocui.View) error {
	common.Logger.Info("inspect network start")
	defer common.Logger.Info("inspect network end")

	selected, err := n.selected()
	if err != nil {
		n.ErrMessage(err.Error(), n.name)
		common.Logger.Error(err)
		return nil
	}

	net, err := n.Docker.InspectNetwork(selected.ID)
	if err != nil {
		n.ErrMessage(err.Error(), n.name)
		common.Logger.Error(err)
		return nil
	}

	n.PopupDetailPanel(g, v)

	v, err = g.View(DetailPanel)
	if err != nil {
		common.Logger.Error(err)
		return nil
	}

	v.Clear()
	v.SetOrigin(0, 0)
	v.SetCursor(0, 0)

	fmt.Fprint(v, common.StructToJSON(net))
	return nil
}

// RemoveNetwork remove the specified network
func (n *NetworkList) RemoveNetwork(g *gocui.Gui, v *gocui.View) error {
	selected, err := n.selected()
	if err != nil {
		n.ErrMessage(err.Error(), n.name)
		common.Logger.Error(err)
		return nil
	}

	n.ConfirmMessage("Are you sure you want to remove this network?", n.name, func() error {
		n.AddTask(fmt.Sprintf("Remove network %s", selected.Name), func() error {
			common.Logger.Info("remove network start")
			defer common.Logger.Info("remove network end")

			if err := n.Docker.RemoveNetwork(selected.ID); err != nil {
				n.ErrMessage(err.Error(), n.name)
				common.Logger.Error(err)
				return err
			}
			return n.Refresh(g, v)
		})
		return nil
	})

	return nil
}
