package panel

import (
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

type SearchImageResult struct {
	*Gui
	Position
	name   string
	images map[string]*SearchResult
}

type SearchResult struct {
	Name        string `tag:"NAME" len:"min:20 max:0.3"`
	Stars       string `tag:"STARS" len:"min:10 max:0.1"`
	Official    string `tag:"OFFICIAL" len:"min:10 max:0.2"`
	Description string `tag:"DESCRIPTION" len:"min:30 max:0.4"`
}

func NewSearchImageResult(g *Gui, name string, p Position) *SearchImageResult {
	return &SearchImageResult{
		Gui:      g,
		name:     name,
		Position: p,
		images:   make(map[string]*SearchResult),
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
	if err := s.SetKeybinding(s.name, gocui.KeyEsc, gocui.ModNone, s.ClosePanel); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, gocui.KeyTab, gocui.ModNone, s.SwitchToSearch); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, 'j', gocui.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, 'k', gocui.ModNone, CursorUp); err != nil {
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

func (s *SearchImageResult) getImageName(v *gocui.View) string {

	line := ReadLine(v, nil)
	if line == "" {
		return ""
	}

	return strings.Split(line, " ")[0]
}

func (s *SearchImageResult) PullImage(g *gocui.Gui, v *gocui.View) error {
	g.Update(func(g *gocui.Gui) error {
		s.StateMessage("image pulling...")

		g.Update(func(g *gocui.Gui) error {
			defer s.CloseStateMessage()

			options := docker.PullImageOptions{
				Repository: s.getImageName(v),
				Tag:        "latest",
			}
			if err := s.Docker.PullImageWithOptions(options); err != nil {
				s.ErrMessage(err.Error(), s.name)
				return nil
			}

			s.ClosePanel(g, v)
			s.CloseSearchPanel()

			s.SwitchPanel(ImageListPanel)

			return nil
		})

		return nil
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

	var names []string

	for _, image := range s.images {
		names = append(names, image.Name)
	}

	for _, name := range common.SortKeys(names) {
		common.OutputFormatedLine(v, s.images[name])
	}
}
