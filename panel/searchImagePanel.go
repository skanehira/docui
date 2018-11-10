package panel

import (
	"strconv"

	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

type SearchImage struct {
	*Gui
	Position
	name        string
	resultPanel *SearchImageResult
}

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
		panic(err)
	}

	g.SwitchPanel(SearchImagePanel)

	return s
}

func (s *SearchImage) Name() string {
	return s.name
}

func (s *SearchImage) SetView(g *gocui.Gui) error {
	v, err := g.SetView(s.name, s.x, s.y, s.w, s.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
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

func (s *SearchImage) Refresh(g *gocui.Gui, v *gocui.View) error {
	return nil
}

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

func (s *SearchImage) SwitchToResult(g *gocui.Gui, v *gocui.View) error {
	if !s.IsSetView(SearchImageResultPanel) {
		return nil
	}

	s.SwitchPanel(SearchImageResultPanel)
	return nil
}

func (s *SearchImage) SearchImage(g *gocui.Gui, v *gocui.View) error {
	name := ReadLine(v, nil)

	if name != "" {
		g.Update(func(g *gocui.Gui) error {
			s.StateMessage("image searching...")

			g.Update(func(g *gocui.Gui) error {
				s.CloseStateMessage()

				// clear result
				s.resultPanel.images = make([]*SearchResult, 0)

				images, err := s.Docker.SearchImageWithName(name)

				if err != nil {
					s.ErrMessage(err.Error(), s.name)
					return nil
				}

				if len(images) == 0 {
					if s.IsSetView(SearchImageResultPanel) {
						s.resultPanel.ClosePanel(g, v)
					}

					s.ErrMessage("not found image", s.name)
					return nil
				}

				var names []string
				tmpMap := make(map[string]*SearchResult)

				for _, image := range images {
					var official string
					if image.IsOfficial {
						official = "[OK]"
					}

					stars := strconv.Itoa(image.StarCount)

					result := &SearchResult{
						Name:        image.Name,
						Stars:       stars,
						Official:    official,
						Description: image.Description,
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
					panic(err)
				}

				s.SwitchPanel(SearchImageResultPanel)

				return nil
			})

			return nil
		})
	}

	return nil
}

func (s *SearchImage) ClosePanel(g *gocui.Gui, v *gocui.View) error {
	if err := s.resultPanel.ClosePanel(g, v); err != nil {
		if err != gocui.ErrUnknownView {
			panic(err)
		}
	}

	s.DeleteKeybindings(s.name)
	if err := s.DeleteView(s.name); err != nil {
		panic(err)
	}

	s.NextPanel = ImageListPanel
	s.SwitchPanel(s.NextPanel)

	return nil
}
