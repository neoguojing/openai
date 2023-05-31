package main

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/neoguojing/openai"
	"github.com/neoguojing/openai/config"
	"github.com/neoguojing/openai/models"
	"github.com/neoguojing/openai/role"
)

type historyStack struct {
	stack []string
	size  int
	index int
	iter  int
}

func newHistoryStack(size int) *historyStack {
	return &historyStack{
		stack: make([]string, size),
		size:  size,
		index: -1,
		iter:  -1,
	}
}

func (h *historyStack) Push(message string) {
	h.index = (h.index + 1) % h.size
	h.stack[h.index] = message
	h.iter = h.index
}

func (h *historyStack) Pop() string {
	if h.index == -1 {
		return ""
	}
	message := h.stack[h.index]
	h.index = (h.index - 1 + h.size) % h.size
	return message
}

func (h *historyStack) Top() string {
	if h.index == -1 {
		return ""
	}
	return h.stack[h.index]
}

func (h *historyStack) Up() string {
	if h.iter == -1 {
		return ""
	}
	message := h.stack[h.iter]
	if message != "" {
		h.iter = (h.iter - 1 + h.size) % h.size
	}

	return message
}

func (h *historyStack) Down() string {
	if h.iter == -1 {
		return ""
	}

	tmp := (h.iter + 1) % h.size
	message := h.stack[tmp]
	if message != "" {
		h.iter = tmp
	}

	return message
}

// top func for history stack

var history = newHistoryStack(100)

func main() {
	role.LoadRoles2DB()
	g, err := gocui.NewGui(gocui.Output256, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer g.Close()
	g.ASCII = false

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		fmt.Println(err)
		return
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		fmt.Println(err)
	}
}

func showMessageInOutput(g *gocui.Gui, message string, align string) error {
	outputView, err := g.View("output")
	if err != nil {
		return err
	}

	vX, _ := outputView.Size()
	if align == "left" {
		fmt.Fprintln(outputView, message)
	} else if align == "right" {
		fmt.Fprintf(outputView, "\033[1;32m%*s\033[0m\n", vX-len(message), message)
	}

	return nil
}

func send(g *gocui.Gui, v *gocui.View) error {
	inputView, err := g.View("input")
	if err != nil {
		return err
	}
	message := inputView.Buffer()
	inputView.Clear()
	if message != "" {
		showMessageInOutput(g, message, "right")
		out, err := openAiChat(message)
		if err != nil {
			showMessageInOutput(g, err.Error(), "left")
		}

		for _, rmessage := range out {
			showMessageInOutput(g, rmessage, "left")
		}

		history.Push(message)
	}

	inputView.SetCursor(0, 0)
	inputView.SetOrigin(0, 0)

	return nil
}

func openAiChat(input string) ([]string, error) {
	config := config.GetConfig()

	if config.OpenAI.ApiKey == "" {
		panic("pls put a api key in config.yml")
	}

	chat := openai.NewOpenAI(config.OpenAI.ApiKey)
	resp, err := chat.Chat(openai.WithPlatform(models.Chatbot)).Complete(input)
	if err != nil {
		return nil, err
	}

	outPut := make([]string, 0)
	for _, choice := range resp.Choices {
		outPut = append(outPut, choice.Message.Content)
	}
	return outPut, nil

}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("output", 0, 0, maxX-1, maxY-3, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Output"
		v.Autoscroll = true
		v.Wrap = true
		v.Frame = true
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		v.FgColor = gocui.ColorWhite
		v.BgColor = gocui.ColorBlack
	}
	if v, err := g.SetView("input", 0, maxY-3, maxX-1, maxY-1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Input"
		v.Editable = true
		v.Wrap = true
		v.Frame = true
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		v.FgColor = gocui.ColorWhite
		v.BgColor = gocui.ColorBlack
		if _, err := g.SetCurrentView("input"); err != nil {
			return err
		}
		if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, send); err != nil {
			return err
		}

		if err := g.SetKeybinding("input", gocui.KeyArrowUp, gocui.ModNone, historyUp); err != nil {
			return err
		}
		if err := g.SetKeybinding("input", gocui.KeyArrowDown, gocui.ModNone, historyDown); err != nil {
			return err
		}

		if err := g.SetKeybinding("input", gocui.KeyBackspace, gocui.ModNone, deleteChar); err != nil {
			return err
		}
	}
	return nil
}

func historyUp(g *gocui.Gui, v *gocui.View) error {
	message := history.Up()

	return showMessageInInput(g, message)
}

func historyDown(g *gocui.Gui, v *gocui.View) error {
	message := history.Down()

	return showMessageInInput(g, message)
}

func deleteChar(g *gocui.Gui, v *gocui.View) error {
	cx, _ := v.Cursor()
	if cx > 0 {
		ox, oy := v.Origin()
		v.EditDelete(true)
		v.SetCursor(cx-1, oy)
		v.SetOrigin(ox, oy)
	}
	return nil
}

func showMessageInInput(g *gocui.Gui, message string) error {
	if message == "" {
		return nil
	}

	inputView, err := g.View("input")
	if err != nil {
		return err
	}
	inputView.Clear()
	fmt.Fprintln(inputView, message)
	inputView.SetCursor(len(message), 0)
	inputView.SetOrigin(len(message), 0)

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
