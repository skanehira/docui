package panel

import (
	"github.com/jroimartin/gocui"
	component "github.com/skanehira/gocui-component"
)

type FormItem struct {
	label, text    string
	labelw, fieldw int
	validator      *Validator
}

type ButtonHandler struct {
	label   string
	handler func(*gocui.Gui, *gocui.View) error
}

type Handler struct {
	key     interface{}
	handler func(*gocui.Gui, *gocui.View) error
}

type Form struct {
	*component.Form
}

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

func NewItem(label, text string, lw, fw int, v *Validator) FormItem {
	return FormItem{
		label:     label,
		text:      text,
		labelw:    lw,
		fieldw:    fw,
		validator: v,
	}
}

func (f *Form) AddFormItems(items []FormItem) *Form {
	for _, item := range items {
		i := f.AddInputField(item.label, item.labelw, item.fieldw)

		if item.text != "" {
			i.SetText(item.text)
		}

		if item.validator != nil {
			i.AddValidate(item.validator.Message, item.validator.Validate)
		}
	}

	return f
}

func (f *Form) AddInput(label string, labelw, fieldw int) *component.InputField {
	return f.AddInputField(label, labelw, fieldw).AddHandler(gocui.KeyEsc, f.Close)
}

func (f *Form) AddSelectOption(label string, labelw, listw int) *component.Select {
	selectOption := f.AddSelect(label, labelw, listw)
	selectOption.AddHandler(gocui.KeyEsc, f.Close)
	return selectOption
}

func (f *Form) AddButton(label string, handler func(g *gocui.Gui, v *gocui.View) error) *component.Button {
	return f.Form.AddButton(label, handler).AddHandler(gocui.KeyEsc, f.Close)
}

func (f *Form) AddButtonFuncs(buttonHandlers []ButtonHandler) *Form {
	for _, b := range buttonHandlers {
		f.Form.AddButton(b.label, b.handler)
	}
	return f
}

func (f *Form) AddGlobalFuncs(globalHandler []Handler) *Form {
	for _, h := range globalHandler {
		for _, c := range f.GetItems() {
			c.AddHandlerOnly(h.key, h.handler)
		}
	}
	return f
}

func (f *Form) AddGlobalFunc(h Handler) *Form {
	for _, c := range f.GetItems() {
		c.AddHandlerOnly(h.key, h.handler)
	}
	return f
}
