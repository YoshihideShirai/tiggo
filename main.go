package main

import (
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/go-git/go-git/v5"
	"github.com/rivo/tview"
)

var tiggoApp *tview.Application
var gitRepos *git.Repository

func CreateTiggoApp() {
	newPrimitive := func(text string) tview.Primitive {
		v := tview.NewTextView().
			SetTextAlign(tview.AlignLeft)
		v.SetBackgroundColor(tcell.ColorBlueViolet)
		v.SetText(text)
		return v
	}

	main := ViewMain()

	grid := tview.NewGrid().
		SetRows(1, 0).
		SetBorders(false).
		AddItem(newPrimitive("Tiggo"), 0, 0, 1, 3, 0, 0, false).
		AddItem(main, 1, 0, 1, 3, 0, 0, false)

	app := tview.NewApplication().SetRoot(grid, true).SetFocus(main)

	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				if app.GetFocus() == main {
					app.Stop()
					return nil
				}
				return event
			}
		}
		return event
	})

	tiggoApp = app
}

func main() {
	reposDir := "."
	if len(os.Args) >= 2 {
		reposDir = os.Args[1]
	}
	r, err := git.PlainOpen(reposDir)
	if err != nil {
		return
	}
	gitRepos = r

	CreateTiggoApp()

	if err := tiggoApp.Run(); err != nil {
		panic(err)
	}
}
