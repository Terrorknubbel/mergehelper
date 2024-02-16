package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Terrorknubbel/mergehelper/internal/gitrunner"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	// "github.com/charmbracelet/log"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	// helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle = lipgloss.NewStyle().Height(0).Margin(1, 0, 1, 0)
	redStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list       list.Model
	choice     string
	quitting   bool
	output     string
	err        error
	commandIdx int
}

func executeChoice(branchName string, commandIdx int) tea.Cmd {
	return func() tea.Msg {
		hasMore, output, err := gitrunner.CheckBranchCondition(commandIdx)

		if err != nil {
			return errMsg{err}
		}

		return outputMsg{output, hasMore, ExecuteChoiceSource}
	}
}

type functionSource int

const (
   CheckPrerequisitesSource functionSource = iota
   ExecuteChoiceSource
)

type outputMsg struct {
  msg     string
  hasMore bool
  source  functionSource
}

func (m model) checkPrerequisites() tea.Msg {
	hasMore, output, err := gitrunner.CheckPrerequisites(m.commandIdx)

	if err != nil {
		return errMsg{err}
	}

	return outputMsg{output, hasMore, CheckPrerequisitesSource}
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() } // TODO: Wirklich nötig?

func (m model) Init() tea.Cmd {
	return m.checkPrerequisites
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
	case outputMsg:
		m.commandIdx++
		m.output += msg.msg
		if msg.hasMore {
			if msg.source == CheckPrerequisitesSource {
				return m, m.checkPrerequisites
			} else if msg.source == ExecuteChoiceSource {
				return m, executeChoice(m.choice, m.commandIdx)
			}
		} else {
			m.commandIdx = 0
		}
	case errMsg:
		m.err = msg
		return m, tea.Quit
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, executeChoice(m.choice, m.commandIdx)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	s := ""

	s += titleStyle.Render("Mergehelper\n")
	divider := strings.Repeat(lipgloss.NewStyle().SetString("-").String(), 80)
	s += divider + "\n" // TODO: divider Positionierung
	s += strings.TrimSpace(quitTextStyle.Render(m.output)) // Warum ist das trimming nötig?

	if m.err != nil {
		return s + quitTextStyle.Render(redStyle.Render(m.err.Error()))
	}
	if m.output != "" {
		return s
	}

	return "\n" + s + m.list.View()
}

func main() {
	items := []list.Item{
		item("Staging"),
		item("Master"),
	}

	const defaultWidth = 30

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "" // TODO
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	// l.Styles.HelpStyle = helpStyle

	m := model{list: l}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
