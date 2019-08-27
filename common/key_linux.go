package common

// TODO remove this when tcell issue #194 is fixed
// https://github.com/rivo/tview/issues/165
// When use tivew suspend function, lost one keystroke
func SendExtraEventFix() error {
	return nil
}
