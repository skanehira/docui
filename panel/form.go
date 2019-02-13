package panel

import (
	"github.com/jroimartin/gocui"
	component "github.com/skanehira/gocui-component"
)

// FormItem form item.
type FormItem struct {
	label, text    string
	labelw, fieldw int
	validator      *Validator
}

// ButtonHandler button handler.
type ButtonHandler struct {
	label   string
	handler func(*gocui.Gui, *gocui.View) error
}

// Handler handler.
type Handler struct {
	key     interface{}
	handler func(*gocui.Gui, *gocui.View) error
}

// Form form info.
type Form struct {
	*component.Form
}

// NewForm create new form.
func NewForm(gui *gocui.Gui, name string, x, y, w, h int) *Form {
	form := &Form{
		Form: component.NewForm(gui, name, x, y, w, h),
	}

	form.AddCloseFunc(func() error {
		gui.Cursor = false
		return nil
	})

	return form
}

// AddInput add input filed.
func (f *Form) AddInput(label string, labelw, fieldw int) *component.InputField {
	return f.AddInputField(label, labelw, fieldw).AddHandler(gocui.KeyEsc, f.Close)
}

// AddSelectOption add option filed.
func (f *Form) AddSelectOption(label string, labelw, listw int) *component.Select {
	selectOption := f.AddSelect(label, labelw, listw)
	selectOption.AddHandler(gocui.KeyEsc, f.Close)
	return selectOption
}

// AddButton add button.
func (f *Form) AddButton(label string, handler func(g *gocui.Gui, v *gocui.View) error) *component.Button {
	return f.Form.AddButton(label, handler).AddHandler(gocui.KeyEsc, f.Close)
}

// AddButtonFuncs add button functions.
func (f *Form) AddButtonFuncs(buttonHandlers []ButtonHandler) *Form {
	for _, b := range buttonHandlers {
		f.Form.AddButton(b.label, b.handler)
	}
	return f
}

// AddGlobalFuncs add the specified functions to all form item.
func (f *Form) AddGlobalFuncs(globalHandler []Handler) *Form {
	for _, h := range globalHandler {
		for _, c := range f.GetItems() {
			c.AddHandlerOnly(h.key, h.handler)
		}
	}
	return f
}

// AddGlobalFunc add the specified function to all form item.
func (f *Form) AddGlobalFunc(h Handler) *Form {
	for _, c := range f.GetItems() {
		c.AddHandlerOnly(h.key, h.handler)
	}
	return f
}
