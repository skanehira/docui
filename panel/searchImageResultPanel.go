package panel

import (
	"fmt"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

type SearchImageResult struct {
	*Gui
	Position
	name   string
	images []*SearchResult
}

type SearchResult struct {
	Name        string `tag:"NAME" len:"min:0.1 max:0.4"`
	Stars       string `tag:"STARS" len:"min:0.1 max:0.1"`
	Official    string `tag:"OFFICIAL" len:"min:0.1 max:0.1"`
	Description string `tag:"DESCRIPTION" len:"min:0.1 max:0.4"`
}

func NewSearchImageResult(g *Gui, name string, p Position) *SearchImageResult {
	return &SearchImageResult{
		Gui:      g,
		name:     name,
		Position: p,
	}
}

func (s *SearchImageResult) Name() string {
	return s.name
}

func (s *SearchImageResult) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := g.SetView(SearchImageResultHeaderPanel, s.x, s.y, s.w, s.h); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Wrap = true
		v.Title = v.Name()
		v.FgColor = gocui.AttrBold | gocui.ColorWhite
		common.OutputFormatedHeader(v, &SearchResult{})
	}

	// set scroll panel
	if v, err := g.SetView(s.name, s.x, s.y+1, s.w, s.h); err != nil {
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
		s.DisplayResult(v)
	}

	s.SetKeyBinding()

	return nil
}

func (s *SearchImageResult) Refresh(g *gocui.Gui, v *gocui.View) error {
	s.Update(func(g *gocui.Gui) error {
		s.DisplayResult(v)
		return nil
	})

	return nil
}

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

func (s *SearchImageResult) SwitchToSearch(g *gocui.Gui, v *gocui.View) error {
	s.SwitchPanel(SearchImagePanel)
	return nil
}

func (s *SearchImageResult) ClosePanel(g *gocui.Gui, v *gocui.View) error {
	s.DeleteKeybindings(s.name)

	if err := s.DeleteView(s.name); err != nil {
		return err
	}

	s.CloseResultHeaderPanel()
	s.SwitchToSearch(g, v)

	return nil
}

func (s *SearchImageResult) getImageName() string {
	return s.selected().Name
}

func (s *SearchImageResult) selected() *SearchResult {
	v, _ := s.View(s.name)
	_, cy := v.Cursor()
	_, oy := v.Origin()
	return s.images[cy+oy]
}

func (s *SearchImageResult) PullImage(g *gocui.Gui, v *gocui.View) error {
	name := s.getImageName()

	s.ClosePanel(g, v)
	s.CloseSearchPanel()
	s.SwitchPanel(ImageListPanel)

	s.AddTask(fmt.Sprintf("Pull image %s", name), func() error {
		options := docker.PullImageOptions{
			Repository: name,
			Tag:        "latest",
		}
		if err := s.Docker.PullImageWithOptions(options); err != nil {
			s.ErrMessage(err.Error(), s.name)
			return nil
		}

		return s.Panels[ImageListPanel].Refresh(g, v)
	})
	return nil
}

func (s *SearchImageResult) CloseSearchPanel() {
	panel := SearchImagePanel
	s.DeleteKeybindings(panel)
	s.DeleteView(panel)
}

func (s *SearchImageResult) CloseResultHeaderPanel() {
	panel := SearchImageResultHeaderPanel
	s.DeleteKeybindings(panel)
	s.DeleteView(panel)
}

func (s *SearchImageResult) DisplayResult(v *gocui.View) {
	v.Clear()
	for _, image := range s.images {
		common.OutputFormatedLine(v, image)
	}
}
