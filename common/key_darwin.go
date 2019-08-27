package common

import (
	"time"

	"github.com/micmonay/keybd_event"
)

// TODO remove this when tcell issue #194 is fixed
// https://github.com/rivo/tview/issues/165
// When use tivew suspend function, lost one keystroke
func SendExtraEventFix() error {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		return err
	}

	time.Sleep(80 * time.Millisecond)

	kb.SetKeys(keybd_event.VK_ENTER)
	return kb.Launching()
}
