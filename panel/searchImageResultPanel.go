package panel

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
)

type SearchImageResult struct {
	*Gui
	Position
	name   string
	images map[string]docker.APIImageSearch
}

func NewSearchImageResult(g *Gui, name string, p Position) *SearchImageResult {
	return &SearchImageResult{
		Gui:      g,
		name:     name,
		Position: p,
		images:   make(map[string]docker.APIImageSearch),
	}
}

func (s *SearchImageResult) Name() string {
	return s.name
}

func (s *SearchImageResult) SetView(g *gocui.Gui) error {
	v, err := g.SetView(s.name, s.x, s.y, s.w, s.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Wrap = true
		v.Title = v.Name()
	}

	s.SetKeyBinding()
	s.DisplayImages()

	return nil
}

func (s *SearchImageResult) Refresh() error {
	s.Update(func(g *gocui.Gui) error {
		s.DisplayImages()
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

			s.Panels[ImageListPanel].Refresh()
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
	s.DeleteKeybindings(SearchImagePanel)
	s.DeleteView(SearchImagePanel)
}

func (s *SearchImageResult) DisplayImages() {
	v, err := s.View(s.name)
	if err != nil {
		panic(err)
	}
	v.Clear()
	v.SetCursor(0, 1)
	v.SetOrigin(0, 0)

	format := "%-45s %-10s %-10s %-60s\n"
	fmt.Fprintf(v, format, "NAME", "STARS", "OFFICIAL", "DESCRIPTION")

	// sort
	var sortedName []string

	for _, image := range s.images {
		sortedName = append(sortedName, image.Name)
	}

	sort.Strings(sortedName)

	for _, name := range sortedName {
		image := s.images[name]
		var official string
		if image.IsOfficial {
			official = "[OK]"
		}

		if strings.Index("\n", image.Description) == -1 {
			image.Description = strings.Replace(image.Description, "\n", " ", -1)
		}

		fmt.Fprintf(v, format, image.Name, strconv.Itoa(image.StarCount), official, image.Description)
	}
}
