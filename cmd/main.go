package main

import (
	"fmt"
	"os/exec"

	"github.com/jroimartin/gocui"
	"github.com/neoguojing/openai"
)

type HistoryStack struct {
	stack []string
	top   int
}

func NewHistoryStack() *HistoryStack {
	return &HistoryStack{
		stack: make([]string, 0),
		top:   -1,
	}
}

func (s *HistoryStack) Push(str string) {
	s.stack = append(s.stack, str)
	s.top++
}

func (s *HistoryStack) Pop() string {
	if s.top == -1 {
		return ""
	}
	str := s.stack[s.top]
	s.stack = s.stack[:s.top]
	s.top--
	return str
}

func (s *HistoryStack) Peek() string {
	if s.top == -1 {
		return ""
	}
	return s.stack[s.top]
}

func (s *HistoryStack) IsEmpty() bool {
	return s.top == -1
}

var historyStack *HistoryStack

func init() {
	historyStack = NewHistoryStack()
}

func main() {

	// 设置终端字符集为 UTF-8
	cmd := exec.Command("stty", "-F", "/dev/tty", "encoding", "utf-8")
	if err := cmd.Run(); err != nil {
		fmt.Println("set tty utf8 failed", err)
		return
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		fmt.Println(err)
		return
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		fmt.Println(err)
	}
}

func showMessage(g *gocui.Gui, message string, align string) error {
	outputView, err := g.View("output")
	if err != nil {
		return err
	}

	vX, _ := outputView.Size()
	if align == "left" {
		fmt.Fprintln(outputView, message)
	} else if align == "right" {
		fmt.Fprintf(outputView, "%*s\n", vX, message)
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
		showMessage(g, message, "right")
		out, err := openAiChat(message)
		if err != nil {
			showMessage(g, err.Error(), "left")
		}

		for _, rmessage := range out {
			showMessage(g, rmessage, "left")
		}

		historyStack.Push(message)
	}

	inputView.SetCursor(0, 0)
	inputView.SetOrigin(0, 0)

	return nil
}

func openAiChat(input string) ([]string, error) {
	chat := openai.NewOpenAI("")
	resp, err := chat.Chat().Completions(input)
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
	if v, err := g.SetView("output", 0, 0, maxX-1, maxY-3); err != nil {
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
	if v, err := g.SetView("input", 0, maxY-3, maxX-1, maxY-1); err != nil {
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

		if err := g.SetKeybinding("input", gocui.KeyArrowUp, gocui.ModNone, history); err != nil {
			return err
		}
	}
	return nil
}

func history(g *gocui.Gui, v *gocui.View) error {
	historyView, err := g.View("input")
	if err != nil {
		return err
	}
	historyView.Clear()
	historyView.SetOrigin(0, 0)
	historyView.SetCursor(0, 0)

	message := historyStack.Pop()
	fmt.Fprintln(historyView, message)

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
