package main

import (
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rivo/tview"
)

type gitcommit struct {
	commit *object.Commit
	ref    []*plumbing.Reference
}

var tiggo_app *tview.Application

var gitRepos *git.Repository
var gitReposHead *plumbing.Reference
var commitlist []gitcommit

func create_tiggo_app() {
	newPrimitive := func(text string) tview.Primitive {
		v := tview.NewTextView().
			SetTextAlign(tview.AlignLeft)
		v.SetBackgroundColor(tcell.ColorBlueViolet)
		v.SetText(text)
		return v
	}

	main := view_main()

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

	tiggo_app = app
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

	create_tiggo_app()

	if err := tiggo_app.Run(); err != nil {
		panic(err)
	}
}
