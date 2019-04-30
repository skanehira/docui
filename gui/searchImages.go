package gui

import (
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker"
)

type searchImageResult struct {
	Name        string
	Stars       string
	Official    string
	Description string
}

type searchImageResults struct {
	keyword            string
	searchImageResults []*searchImageResult
	*tview.Table
}

func newSearchImageResults(g *Gui, keyword string) *searchImageResults {
	searchImageResults := &searchImageResults{
		keyword: keyword,
		Table:   tview.NewTable().SetSelectable(true, false).Select(0, 0).SetFixed(1, 1),
	}

	searchImageResults.SetTitle("search result").SetTitleAlign(tview.AlignLeft)
	searchImageResults.SetBorder(true)
	searchImageResults.setEntries(g)
	searchImageResults.setKeybinding(g)
	return searchImageResults
}

func newSearchInputField(g *Gui) {
	viewName := "searchImageInput"
	searchInput := tview.NewInputField().SetLabel("Image")
	searchInput.SetLabelWidth(6)
	searchInput.SetBorder(true)

	searchInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			g.pages.AddAndSwitchToPage("searchImageResults", g.modal(newSearchImageResults(g, searchInput.GetText()), 100, 50), true).ShowPage("main")
		}
	})

	closeSearchInput := func() {
		currentPanel := g.state.panels.panel[g.state.panels.currentPanel]
		g.closeAndSwitchPanel(viewName, currentPanel.name())
	}

	searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			closeSearchInput()
		}

		switch event.Rune() {
		case 'q':
			closeSearchInput()
		}

		return event
	})

	g.pages.AddAndSwitchToPage("searchImageInput", g.modal(searchInput, 80, 3), true).ShowPage("main")
}

func (s *searchImageResults) name() string {
	return "searchImageResults"
}

func (s *searchImageResults) pullImage(g *Gui) {
	currentPanel := g.state.panels.panel[g.state.panels.currentPanel]
	g.pullImage(s.selectedSearchImageResult().Name, s.name(), currentPanel.name())
}

func (s *searchImageResults) setKeybinding(g *Gui) {
	s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			s.closePanel(g)
		case tcell.KeyEnter:
			s.pullImage(g)
		}

		switch event.Rune() {
		case 'p':
			s.pullImage(g)
		case 'q':
			s.closePanel(g)
		}

		return event
	})
}

func (s *searchImageResults) entries(g *Gui) {
	images, err := docker.Client.SearchImage(s.keyword)

	if err != nil {
		// TODO display error message
		return
	}

	if len(images) == 0 {
		// TODO display message "not found message"
		return
	}

	s.searchImageResults = make([]*searchImageResult, 0)

	var official string
	for _, image := range images {
		if image.IsOfficial {
			official = "[OK]"
		}

		s.searchImageResults = append(s.searchImageResults, &searchImageResult{
			Name:        image.Name,
			Stars:       strconv.Itoa(image.StarCount),
			Official:    official,
			Description: common.CutNewline(image.Description),
		})
	}

}

func (s *searchImageResults) setEntries(g *Gui) {
	s.entries(g)
	table := s.Clear()

	headers := []string{
		"Name",
		"Star",
		"Official",
		"Description",
	}

	for i, header := range headers {
		table.SetCell(0, i, &tview.TableCell{
			Text:            header,
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorWhite,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrBold,
		})
	}

	for i, image := range s.searchImageResults {
		table.SetCell(i+1, 0, tview.NewTableCell(image.Name).
			SetTextColor(tcell.ColorLightYellow).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 1, tview.NewTableCell(image.Stars).
			SetTextColor(tcell.ColorLightYellow))

		table.SetCell(i+1, 2, tview.NewTableCell(image.Official).
			SetTextColor(tcell.ColorLightYellow))

		table.SetCell(i+1, 3, tview.NewTableCell(image.Description).
			SetTextColor(tcell.ColorLightYellow).
			SetMaxWidth(1).
			SetExpansion(1))

	}
}

func (s *searchImageResults) updateEntries(g *Gui) {
	g.app.QueueUpdateDraw(func() {
		s.setEntries(g)
	})
}

func (s *searchImageResults) focus(g *Gui) {
	s.SetSelectable(true, false)
	g.app.SetFocus(s)
}

func (s *searchImageResults) unfocus() {
	s.SetSelectable(false, false)
}

func (s *searchImageResults) closePanel(g *Gui) {
	currentPanel := g.state.panels.panel[g.state.panels.currentPanel]
	g.closeAndSwitchPanel(s.name(), currentPanel.name())
}

func (s *searchImageResults) selectedSearchImageResult() *searchImageResult {
	row, _ := s.GetSelection()

	if len(s.searchImageResults) == 0 || row-1 < 0 {
		return nil
	}

	return s.searchImageResults[row-1]
}
