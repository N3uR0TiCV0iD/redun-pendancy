package gui

import (
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type HoverableCheckbox struct {
	//inherit
	widget.Check

	mouseIn  func()
	mouseOut func()
}

func NewHoverableCheckbox(label string, changed func(bool)) *HoverableCheckbox {
	checkbox := &HoverableCheckbox{}
	checkbox.ExtendBaseWidget(checkbox)
	checkbox.Text = label
	checkbox.OnChanged = changed
	return checkbox
}

func (checkbox *HoverableCheckbox) MouseIn(me *desktop.MouseEvent) {
	if checkbox.mouseIn != nil {
		checkbox.mouseIn()
	}
}

func (checkbox *HoverableCheckbox) MouseOut() {
	if checkbox.mouseOut != nil {
		checkbox.mouseOut()
	}
}

func (checkbox *HoverableCheckbox) MouseMoved(me *desktop.MouseEvent) {
}

func (checkbox *HoverableCheckbox) SetMouseInCallback(mouseIn func()) {
	checkbox.mouseIn = mouseIn
}

func (checkbox *HoverableCheckbox) SetMouseOutCallback(mouseOut func()) {
	checkbox.mouseOut = mouseOut
}
