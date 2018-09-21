package panel

import (
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
)

type SearchImage struct {
	*Gui
	Position
	name        string
	input       string
	resultPanel *SearchImageResult
}

func NewSearchImage(g *Gui, name string, p Position) *SearchImage {
	_, y := g.Size()
	resultPanel := NewSearchImageResult(
		g,
		SearchImageResultPanel,
		Position{p.x, p.y + 3, p.w, y - p.y},
	)

	return &SearchImage{
		Gui:         g,
		name:        name,
		Position:    p,
		resultPanel: resultPanel,
	}
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
		v.Editor = s
	}

	if _, err := SetCurrentPanel(g, v.Name()); err != nil {
		return err
	}

	s.SetKeyBinding()
	return nil
}

func (s *SearchImage) Refresh() error {
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

	v = s.SwitchPanel("", SearchImageResultPanel)
	v.SetCursor(0, 1)
	return nil
}

func (s *SearchImage) SearchImage(g *gocui.Gui, v *gocui.View) error {
	name := ReadLine(v, nil)

	if name != "" {
		g.Update(func(g *gocui.Gui) error {
			s.StateMessage("image searching...")

			g.Update(func(g *gocui.Gui) error {
				s.CloseStateMessage()
				SetCurrentPanel(g, SearchImagePanel)

				// clear result
				s.resultPanel.images = make(map[string]docker.APIImageSearch)

				images, err := s.Docker.SearchImageWithName(name)
				if err != nil {
					s.ErrMessage(err.Error(), s.name)
					return nil
				}
				if len(images) == 0 {
					s.ErrMessage("not found page", s.name)
					return nil
				}

				for _, image := range images {
					s.resultPanel.cachedImages[image.Name] = image
					s.resultPanel.images[image.Name] = image
				}

				if _, err := s.View(SearchImageResultPanel); err != nil {
					if err == gocui.ErrUnknownView {
						if err := s.resultPanel.SetView(g); err != nil {
							panic(err)
						}
						return nil
					}
				}

				s.resultPanel.Refresh()

				return nil
			})

			return nil
		})
	}

	return nil
}

func (s *SearchImage) SearchFromCache(v *gocui.View, name string) {
	if name != "" {
		// clear result
		s.resultPanel.images = make(map[string]docker.APIImageSearch)

		for cname, image := range s.resultPanel.cachedImages {
			// 前方一致したイメージ名を検索結果に表示
			if strings.Index(cname, name) != -1 {
				s.resultPanel.images[name] = image
			}
		}

		if _, err := s.View(SearchImageResultPanel); err != nil {
			if err == gocui.ErrUnknownView {
				if err := s.resultPanel.SetView(s.Gui.Gui); err != nil {
					panic(err)
				}
			}
		}
		s.resultPanel.Refresh()

	}
}

func (s *SearchImage) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {

	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	}

	// TODO search from cache
	// name := ReadLine(v, nil)
	//	if len(s.resultPanel.cachedImages) > 0 {
	//		s.SearchFromCache(v, name)
	//	}
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

	if s.NextPanel == s.name {
		s.NextPanel = ImageListPanel
	}

	SetCurrentPanel(g, s.NextPanel)

	return nil
}
