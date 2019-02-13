package panel

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

// SearchImageResult search result panel.
type SearchImageResult struct {
	*Gui
	Position
	name   string
	images []*SearchResult
}

// SearchResult result info.
type SearchResult struct {
	Name        string `tag:"NAME" len:"min:0.1 max:0.3"`
	Stars       string `tag:"STARS" len:"min:0.1 max:0.1"`
	Official    string `tag:"OFFICIAL" len:"min:0.1 max:0.1"`
	Description string `tag:"DESCRIPTION" len:"min:0.1 max:0.5"`
}

// NewSearchImageResult create new result panel.
func NewSearchImageResult(g *Gui, name string, p Position) *SearchImageResult {
	return &SearchImageResult{
		Gui:      g,
		name:     name,
		Position: p,
	}
}

// Name return panel name.
func (s *SearchImageResult) Name() string {
	return s.name
}

// SetView set up result panel.
func (s *SearchImageResult) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := common.SetViewWithValidPanelSize(g, SearchImageResultHeaderPanel, s.x, s.y, s.w, s.h); err != nil {
		if err != gocui.ErrUnknownView {
			common.Logger.Error(err)
			return err
		}

		v.Wrap = true
		v.Title = v.Name()
		v.FgColor = gocui.AttrBold | gocui.ColorWhite
		common.OutputFormattedHeader(v, &SearchResult{})
	}

	// set scroll panel
	if v, err := common.SetViewWithValidPanelSize(g, s.name, s.x, s.y+1, s.w, s.h); err != nil {
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
		s.DisplayResult(v)
	}

	s.SetKeyBinding()

	return nil
}

// Refresh update result info.
func (s *SearchImageResult) Refresh(g *gocui.Gui, v *gocui.View) error {
	s.Update(func(g *gocui.Gui) error {
		s.DisplayResult(v)
		return nil
	})

	return nil
}

// SetKeyBinding set key bind to this panel.
func (s *SearchImageResult) SetKeyBinding() {
	if err := s.SetKeybinding(s.name, gocui.KeyEnter, gocui.ModNone, s.PullImage); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, 'p', gocui.ModNone, s.PullImage); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, gocui.KeyCtrlW, gocui.ModNone, s.ClosePanel); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, 'q', gocui.ModNone, s.ClosePanel); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, gocui.KeyEsc, gocui.ModNone, s.ClosePanel); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, gocui.KeyTab, gocui.ModNone, s.SwitchToSearch); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, 'j', gocui.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, gocui.KeyArrowDown, gocui.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, 'k', gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, gocui.KeyArrowUp, gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
}

// SwitchToSearch switch to search panel.
func (s *SearchImageResult) SwitchToSearch(g *gocui.Gui, v *gocui.View) error {
	s.SwitchPanel(SearchImagePanel)
	return nil
}

// ClosePanel close result panel.
func (s *SearchImageResult) ClosePanel(g *gocui.Gui, v *gocui.View) error {
	s.DeleteKeybindings(s.name)

	if err := s.DeleteView(s.name); err != nil {
		common.Logger.Error(err)
		return err
	}

	s.CloseResultHeaderPanel()
	s.SwitchToSearch(g, v)

	return nil
}

// getImageName return selected image name
func (s *SearchImageResult) getImageName() string {
	return s.selected().Name
}

// selected return selected image
func (s *SearchImageResult) selected() *SearchResult {
	v, _ := s.View(s.name)
	_, cy := v.Cursor()
	_, oy := v.Origin()
	return s.images[cy+oy]
}

// PullImage pull the specified image.
func (s *SearchImageResult) PullImage(g *gocui.Gui, v *gocui.View) error {
	name := s.getImageName()

	s.ClosePanel(g, v)
	s.CloseSearchPanel()
	s.SwitchPanel(ImageListPanel)

	s.AddTask(fmt.Sprintf("Pull image %s", name), func() error {
		common.Logger.Info("pull image start")
		defer common.Logger.Info("pull image end")

		if err := s.Docker.PullImage(name); err != nil {
			s.ErrMessage(err.Error(), s.name)
			common.Logger.Error(err)
			return nil
		}

		return s.Panels[ImageListPanel].Refresh(g, v)
	})
	return nil
}

// CloseSearchPanel close search panel.
func (s *SearchImageResult) CloseSearchPanel() {
	panel := SearchImagePanel
	s.DeleteKeybindings(panel)
	s.DeleteView(panel)
}

// CloseResultHeaderPanel close result header panel.
func (s *SearchImageResult) CloseResultHeaderPanel() {
	panel := SearchImageResultHeaderPanel
	s.DeleteKeybindings(panel)
	s.DeleteView(panel)
}

// DisplayResult display result info
func (s *SearchImageResult) DisplayResult(v *gocui.View) {
	v.Clear()
	for _, image := range s.images {
		common.OutputFormattedLine(v, image)
	}
}
