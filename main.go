package main

import (
	"fmt"
	"os"
	"strings"

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

func view_logs_table(table *tview.Table) {
	table.Clear()
	for idx, commit_e := range commitlist {
		c := commit_e.commit
		commit_message_text := ""
		for _, v := range commit_e.ref {
			if v.Name().IsTag() {
				commit_message_text += fmt.Sprintf("[purple]<%s>[white] ", v.Name().Short())
			} else if v.Name().IsBranch() {
				commit_message_text += fmt.Sprintf("[blue](%s)[white] ", v.Name().Short())
			} else {
				commit_message_text += fmt.Sprintf("[yellow]{%s}[white] ", v.Name())
			}
		}
		commit_message_text += tview.Escape(c.Message)
		commit_when := tview.NewTableCell(c.Author.When.Format("2006-01-02 15:04:05 07:00")).
			SetAlign(tview.AlignLeft)
		commit_authorname := tview.NewTableCell(c.Author.Name).
			SetAlign(tview.AlignLeft)
		commit_message := tview.NewTableCell(commit_message_text).
			SetAlign(tview.AlignLeft)
		commit_when.SetTextColor(tcell.ColorBlue)
		commit_authorname.SetTextColor(tcell.ColorGreen)
		table.SetCell(idx, 0, commit_when)
		table.SetCell(idx, 1, commit_authorname)
		table.SetCell(idx, 2, commit_message)
	}
	table.SetSelectable(true, false)
}

func view_logs_statusbar(selectOffset int, status_bar *tview.TextView) {
	status_bar_text := fmt.Sprintf("(%s) %s - commit %d of %d",
		"main",
		commitlist[selectOffset].commit.Hash.String(), selectOffset+1, len(commitlist))
	status_bar.SetText(tview.Escape(status_bar_text))
}

func view_logs_reload(table *tview.Table, status_bar *tview.TextView) {
	view_logs_table(table)
	view_logs_statusbar(0, status_bar)
}

func load_logs() {
	var branches []*plumbing.Reference
	var tags []*plumbing.Reference
	ref, err := gitRepos.Head()
	if err != nil {
		panic(err)
	}

	gitReposHead = ref

	bIter, _ := gitRepos.Branches()
	bIter.ForEach(func(r *plumbing.Reference) error {
		branches = append(branches, r)
		return nil
	})
	tIter, _ := gitRepos.Tags()
	tIter.ForEach(func(r *plumbing.Reference) error {
		tags = append(tags, r)
		return nil
	})
	cIter, err := gitRepos.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		panic(err)
	}

	cIter.ForEach(func(c *object.Commit) error {
		var commit gitcommit
		commit.commit = c
		for _, r := range branches {
			if c.Hash == r.Hash() {
				commit.ref = append(commit.ref, r)
			}
		}
		for _, r := range tags {
			if c.Hash == r.Hash() {
				commit.ref = append(commit.ref, r)
			}
		}
		commitlist = append(commitlist, commit)
		return nil
	})
}

func view_diff(selectOffset int, parent *tview.Grid) tview.Primitive {
	selectCommit := commitlist[selectOffset].commit
	patch, _ := selectCommit.Patch(nil)
	if selectOffset < len(commitlist)-1 {
		next, _ := selectCommit.Parents().Next()
		nextpatch, _ := next.Patch(selectCommit)
		patch = nextpatch
	}

	table := tview.NewTable()

	idx := 0

	table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
		"[green]commit       %s[white]",
		tview.Escape(selectCommit.Hash.String()))))
	idx++
	table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
		"[blue]Author       %s[white]",
		tview.Escape(selectCommit.Author.String()))))
	idx++
	table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
		"[yellow]AuthorDate   %s[white]",
		tview.Escape(selectCommit.Author.When.String()))))
	idx++
	table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
		"[purple]Commiter     %s[white]",
		tview.Escape(selectCommit.Committer.String()))))
	idx++
	table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
		"[yellow]CommiterDate %s[white]",
		tview.Escape(selectCommit.Committer.When.String()))))
	idx++
	table.SetCell(idx, 0, tview.NewTableCell(""))
	idx++

	stats_output := patch.Stats().String()
	for _, v := range strings.Split(stats_output, "\n") {
		table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
			tview.Escape(v))))
		idx++
	}

	patch_output := patch.String()
	for _, v := range strings.Split(patch_output, "\n") {
		table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
			tview.Escape(v))))
		idx++
	}

	view_diff_textbox := table
	view_diff_textbox.
		SetSelectable(true, false).
		SetBorder(false)

	view_diff_textbox.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				offset, _ := view_diff_textbox.GetSelection()
				if offset < view_diff_textbox.GetRowCount()-1 {
					offset++
				}
				view_diff_textbox.Select(offset, 0)
			case 'k':
				offset, _ := view_diff_textbox.GetSelection()
				if offset > 0 {
					offset--
				}
				view_diff_textbox.Select(offset, 0)
			case 'q':
				parent.RemoveItem(view_diff_textbox)
				tiggo_app.SetFocus(parent)
				return nil
			}
		}
		return nil
	})
	tiggo_app.SetFocus(view_diff_textbox)
	return view_diff_textbox
}

func view_logs() tview.Primitive {
	load_logs()

	status_bar := tview.NewTextView().
		SetTextAlign(tview.AlignLeft)
	status_bar.SetBackgroundColor(tcell.ColorBlueViolet)
	view_logs_statusbar(0, status_bar)

	table := tview.NewTable()
	view_logs_table(table)

	grid_log := tview.NewGrid().
		SetRows(0, 1).
		SetColumns(0).
		SetBorders(false).
		AddItem(table, 0, 0, 1, 1, 1, 1, false).
		AddItem(status_bar, 1, 0, 1, 1, 1, 1, false)

	grid := tview.NewGrid().
		SetRows(0).
		SetColumns(0).
		SetBorders(false).
		AddItem(grid_log, 0, 0, 1, 1, 1, 1, false)

	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if tiggo_app.GetFocus() != grid {
			return event
		}
		switch event.Key() {
		case tcell.KeyEnter:
			selected, _ := table.GetSelection()
			v_diff := view_diff(selected, grid)
			grid.AddItem(v_diff, 0, 1, 1, 1, 1, 1, false)
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				selected, _ := table.GetSelection()
				if selected < table.GetRowCount()-1 {
					selected++
				}
				table.Select(selected, 0)
				view_logs_statusbar(selected, status_bar)
				return nil
			case 'k':
				selected, _ := table.GetSelection()
				if selected > 0 {
					selected--
				}
				table.Select(selected, 0)
				view_logs_statusbar(selected, status_bar)
				return nil
			}
		}
		return event
	})

	return grid
}

func view_main() tview.Primitive {
	return view_logs()
}

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
