package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rivo/tview"
)

type gitcommit struct {
	worktree *git.Worktree
	commit   *object.Commit
	ref      []*plumbing.Reference
}

func view_main_table(commitlist []gitcommit) *tview.Table {
	table := tview.NewTable()

	var commit_when *tview.TableCell
	var commit_authorname *tview.TableCell
	var commit_message *tview.TableCell

	for idx, commit_e := range commitlist {
		if commit_e.worktree != nil {
			wtstat, _ := commit_e.worktree.Status()
			if wtstat.IsClean() == true {
				continue
			}
			commit_when = tview.NewTableCell(time.Now().Format("2006-01-02 15:04:05 07:00")).
				SetAlign(tview.AlignLeft)
			commit_authorname = tview.NewTableCell("Unknown").
				SetAlign(tview.AlignLeft)
			commit_message = tview.NewTableCell("Unstaged changes").
				SetAlign(tview.AlignLeft)
		} else {
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
			commit_when = tview.NewTableCell(c.Author.When.Format("2006-01-02 15:04:05 07:00")).
				SetAlign(tview.AlignLeft)
			commit_authorname = tview.NewTableCell(c.Author.Name).
				SetAlign(tview.AlignLeft)
			commit_message = tview.NewTableCell(commit_message_text).
				SetAlign(tview.AlignLeft)
		}
		commit_when.SetTextColor(tcell.ColorBlue)
		commit_authorname.SetTextColor(tcell.ColorGreen)
		table.SetCell(idx, 0, commit_when)
		table.SetCell(idx, 1, commit_authorname)
		table.SetCell(idx, 2, commit_message)
	}
	table.SetSelectable(true, false)
	return table
}

func view_main_statusbar(selectCommit gitcommit, table *tview.Table, status_bar *tview.TextView) {
	row, _ := table.GetSelection()
	var status_bar_text string
	status_bar_text = "(main)"
	if selectCommit.commit != nil {
		status_bar_text += fmt.Sprintf(" %s - commit %d of %d",
			selectCommit.commit.Hash.String(), row+1, table.GetRowCount())
	} else {
		status_bar_text += " Unstaged changes"
	}
	status_bar.SetText(tview.Escape(status_bar_text))
}

func load_logs() []gitcommit {
	var commitlist []gitcommit

	var branches []*plumbing.Reference
	var tags []*plumbing.Reference
	ref, err := gitRepos.Head()
	if err != nil {
		panic(err)
	}

	wt, _ := gitRepos.Worktree()
	var wtcommit gitcommit
	wtcommit.worktree = wt
	commitlist = append(commitlist, wtcommit)

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
	return commitlist
}

func view_main() tview.Primitive {
	commitlist := load_logs()

	status_bar := tview.NewTextView().
		SetTextAlign(tview.AlignLeft)
	status_bar.SetBackgroundColor(tcell.ColorBlueViolet)

	table := view_main_table(commitlist)

	view_main_statusbar(commitlist[0], table, status_bar)

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
			v_diff := view_diff(commitlist[selected], grid)
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
				view_main_statusbar(commitlist[selected], table, status_bar)
				return nil
			case 'k':
				selected, _ := table.GetSelection()
				if selected > 0 {
					selected--
				}
				table.Select(selected, 0)
				view_main_statusbar(commitlist[selected], table, status_bar)
				return nil
			}
		}
		return event
	})

	return grid
}
