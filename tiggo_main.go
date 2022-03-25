package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rivo/tview"
)

func view_main_table(table *tview.Table) {
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

func view_main_statusbar(selectOffset int, status_bar *tview.TextView) {
	status_bar_text := fmt.Sprintf("(%s) %s - commit %d of %d",
		"main",
		commitlist[selectOffset].commit.Hash.String(), selectOffset+1, len(commitlist))
	status_bar.SetText(tview.Escape(status_bar_text))
}

func load_logs() {
	var branches []*plumbing.Reference
	var tags []*plumbing.Reference
	ref, err := gitRepos.Head()
	if err != nil {
		panic(err)
	}

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

func view_main() tview.Primitive {
	load_logs()

	status_bar := tview.NewTextView().
		SetTextAlign(tview.AlignLeft)
	status_bar.SetBackgroundColor(tcell.ColorBlueViolet)
	view_main_statusbar(0, status_bar)

	table := tview.NewTable()
	view_main_table(table)

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
			v_diff := view_diff(commitlist[selected].commit.Hash, grid)
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
				view_main_statusbar(selected, status_bar)
				return nil
			case 'k':
				selected, _ := table.GetSelection()
				if selected > 0 {
					selected--
				}
				table.Select(selected, 0)
				view_main_statusbar(selected, status_bar)
				return nil
			}
		}
		return event
	})

	return grid
}
