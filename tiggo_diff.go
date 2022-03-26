package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rivo/tview"
)

func GetCommitFromHash(selectCommitHash plumbing.Hash) (*object.Commit, error) {
	return gitRepos.CommitObject(selectCommitHash)
}

func CommitGetPatch(selectCommit *object.Commit) (*object.Patch, error) {
	selectTree, _ := selectCommit.Tree()
	next := &object.Tree{}
	c_next, err := selectCommit.Parents().Next()
	if err == nil {
		t_next, _ := c_next.Tree()
		next = t_next
	}
	return next.Patch(selectTree)
}

func ViewDiffStatusbar(selectCommit gitcommit, status_bar *tview.TextView, table *tview.Table) {
	row, _ := table.GetSelection()
	status_bar_text := fmt.Sprintf("[%s]", "diff")
	if selectCommit.commit != nil {
		status_bar_text += fmt.Sprintf(" %s", selectCommit.commit.String())
	}
	status_bar_text += fmt.Sprintf(" - line %d of %d",
		row+1, table.GetRowCount())
	status_bar.SetText(status_bar_text)
}

func ViewDiff(selectCommit gitcommit, parent *tview.Grid) tview.Primitive {
	commit := selectCommit.commit
	worktree := selectCommit.worktree
	var stats_output string
	var patch_output string
	if worktree != nil {
		wtstatus, _ := worktree.Status()
		stats_output += wtstatus.String()
	}
	if commit != nil {
		patch, _ := CommitGetPatch(commit)
		stats_output += patch.Stats().String()
		patch_output += patch.String()
	}

	table := tview.NewTable()

	idx := 0
	if commit != nil {
		table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
			"[green]commit:       %s[white]",
			tview.Escape(commit.Hash.String()))))
		idx++
		table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
			"[blue]Author:       %s[white]",
			tview.Escape(commit.Author.String()))))
		idx++
		table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
			"[yellow]AuthorDate:   %s[white]",
			tview.Escape(commit.Author.When.String()))))
		idx++
		table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
			"[purple]Commiter:     %s[white]",
			tview.Escape(commit.Committer.String()))))
		idx++
		table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
			"[yellow]CommiterDate: %s[white]",
			tview.Escape(commit.Committer.When.String()))))
		idx++
		table.SetCell(idx, 0, tview.NewTableCell(""))
		idx++
		for _, v := range strings.Split(commit.Message, "\n") {
			table.SetCell(idx, 0, tview.NewTableCell(
				fmt.Sprintf("    %s", tview.Escape(v))))
			idx++
		}
		table.SetCell(idx, 0, tview.NewTableCell("---"))
		idx++
	}

	for _, v := range strings.Split(stats_output, "\n") {
		table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
			"%s", tview.Escape(v))).SetExpansion(1))
		idx++
	}

	for _, v := range strings.Split(patch_output, "\n") {
		table.SetCell(idx, 0, tview.NewTableCell(fmt.Sprintf(
			"%s", tview.Escape(v))).SetExpansion(1))
		idx++
	}

	table.SetSelectable(true, false).
		SetBorder(false)

	status_bar := tview.NewTextView().
		SetTextAlign(tview.AlignLeft)
	status_bar.SetBackgroundColor(tcell.ColorBlueViolet)
	ViewDiffStatusbar(selectCommit, status_bar, table)

	grid := tview.NewGrid().
		SetRows(0, 1).
		SetColumns(0).
		SetBorders(false).
		AddItem(table, 0, 0, 1, 1, 1, 1, false).
		AddItem(status_bar, 1, 0, 1, 1, 1, 1, false)

	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				offset, _ := table.GetSelection()
				if offset < table.GetRowCount()-1 {
					offset++
				}
				ViewDiffStatusbar(selectCommit, status_bar, table)
				table.Select(offset, 0)
			case 'k':
				offset, _ := table.GetSelection()
				if offset > 0 {
					offset--
				}
				ViewDiffStatusbar(selectCommit, status_bar, table)
				table.Select(offset, 0)
			case 'q':
				parent.RemoveItem(grid)
				tiggoApp.SetFocus(parent)
				return nil
			}
		}
		return nil
	})
	tiggoApp.SetFocus(grid)
	return grid
}
