package panel

import (
	"strconv"

	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

// SearchImage search image panel.
type SearchImage struct {
	*Gui
	Position
	name        string
	resultPanel *SearchImageResult
}

// NewSearchImage create new search image panel.
func NewSearchImage(g *Gui, name string, p Position) *SearchImage {
	_, y := g.Size()
	resultPanel := NewSearchImageResult(
		g,
		SearchImageResultPanel,
		Position{p.x, p.y + 3, p.w, y - p.y},
	)

	s := &SearchImage{
		Gui:         g,
		name:        name,
		Position:    p,
		resultPanel: resultPanel,
	}

	if err := s.SetView(g.Gui); err != nil {
		common.Logger.Error(err)
		panic(err)
	}

	g.SwitchPanel(SearchImagePanel)

	return s
}

// Name return panel name.
func (s *SearchImage) Name() string {
	return s.name
}

// SetView set up search image panel.
func (s *SearchImage) SetView(g *gocui.Gui) error {
	v, err := common.SetViewWithValidPanelSize(g, s.name, s.x, s.y, s.w, s.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			common.Logger.Error(err)
			return err
		}

		v.Title = v.Name()
		v.Autoscroll = true
		v.Editable = true
		v.Wrap = true
	}

	s.SwitchPanel(v.Name())
	s.SetKeyBinding()
	return nil
}

// Refresh do nothing
func (s *SearchImage) Refresh(g *gocui.Gui, v *gocui.View) error {
	return nil
}

// SetKeyBinding set key bind to this panel.
func (s *SearchImage) SetKeyBinding() {
	if err := s.SetKeybinding(s.name, gocui.KeyCtrlW, gocui.ModNone, s.ClosePanel); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, gocui.KeyEsc, gocui.ModNone, s.ClosePanel); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, gocui.KeyEnter, gocui.ModNone, s.SearchImage); err != nil {
		panic(err)
	}
	if err := s.SetKeybinding(s.name, gocui.KeyTab, gocui.ModNone, s.SwitchToResult); err != nil {
		panic(err)
	}
}

// SwitchToResult stitch result panel.
func (s *SearchImage) SwitchToResult(g *gocui.Gui, v *gocui.View) error {
	if !s.IsSetView(SearchImageResultPanel) {
		return nil
	}

	s.SwitchPanel(SearchImageResultPanel)
	return nil
}

// SearchImage search image
func (s *SearchImage) SearchImage(g *gocui.Gui, v *gocui.View) error {
	name := ReadViewBuffer(v)

	if name != "" {
		g.Update(func(g *gocui.Gui) error {
			s.StateMessage("image searching...")

			g.Update(func(g *gocui.Gui) error {
				common.Logger.Info("search image start")
				defer common.Logger.Info("search image end")

				s.CloseStateMessage()

				// clear result
				s.resultPanel.images = make([]*SearchResult, 0)

				images, err := s.Docker.SearchImage(name)

				if err != nil {
					s.ErrMessage(err.Error(), s.name)
					common.Logger.Error(err)
					return nil
				}

				if len(images) == 0 {
					if s.IsSetView(SearchImageResultPanel) {
						s.resultPanel.ClosePanel(g, v)
					}

					s.ErrMessage("not found image", s.name)
					return nil
				}

				names := make([]string, 0, len(images))
				tmpMap := make(map[string]*SearchResult)

				for _, image := range images {
					var official string
					if image.IsOfficial {
						official = "[OK]"
					}

					result := &SearchResult{
						Name:        image.Name,
						Stars:       strconv.Itoa(image.StarCount),
						Official:    official,
						Description: common.CutNewline(image.Description),
					}

					names = append(names, image.Name)
					tmpMap[image.Name] = result
				}

				for _, name := range common.SortKeys(names) {
					s.resultPanel.images = append(s.resultPanel.images, tmpMap[name])
				}

				if s.IsSetView(SearchImageResultPanel) {
					s.resultPanel.ClosePanel(g, v)
				}

				if err := s.resultPanel.SetView(g); err != nil {
					common.Logger.Error(err)
					return err
				}

				s.SwitchPanel(SearchImageResultPanel)

				return nil
			})

			return nil
		})
	}

	return nil
}

// ClosePanel close search panel.
func (s *SearchImage) ClosePanel(g *gocui.Gui, v *gocui.View) error {
	if err := s.resultPanel.ClosePanel(g, v); err != nil {
		if err != gocui.ErrUnknownView {
			common.Logger.Error(err)
			return err
		}
	}

	s.DeleteKeybindings(s.name)
	if err := s.DeleteView(s.name); err != nil {
		common.Logger.Error(err)
		return err
	}

	s.NextPanel = ImageListPanel
	s.SwitchPanel(s.NextPanel)

	return nil
}
