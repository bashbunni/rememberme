package main

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int
type QNAMsg string

const (
	answerState sessionState = iota
	questionState
)

var qna = map[string]string{"what's 1 + 1": "2", "is red a warm colour?": "yes", "what is the best snack": "popcorn"}
var (
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	boxstyle  = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Border(lipgloss.DoubleBorder(), true).
			BorderForeground(highlight).
			Padding(0, 1)
	width  = 96
	subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
)

type model struct {
	state    sessionState
	viewport viewport.Model
	question string
	answer   string
}

func getRandomQuestion(m map[string]string) string {
	rand.Seed(time.Now().UnixNano())
	return reflect.ValueOf(m).MapKeys()[rand.Intn(len(m))].String()
}

func (m model) getQNACmd() tea.Msg {
	return QNAMsg(getRandomQuestion(qna))
}

func New() model {
	q := getRandomQuestion(qna)
	m := model{state: questionState, question: q, answer: qna[q]}
	m.viewport = viewport.New(8, 8)
	m.viewport.SetContent(m.question)
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			if m.state == questionState {
				m.state = answerState
				m.viewport.SetContent(m.answer)
			} else {
				m.state = questionState
				m.viewport.SetContent(m.question)
			}
		}
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		if msg.String() == "n" {
			// refresh new question and answer
			return m, m.getQNACmd
		}
	case QNAMsg:
		m.question = string(msg)
		m.answer = qna[string(msg)]
		m.viewport.SetContent(m.question)
	}
	return m, nil
}

func (m model) View() string {
	return lipgloss.Place(width, 9,
		lipgloss.Center, lipgloss.Center,
		boxstyle.Render(m.viewport.View()),
		lipgloss.WithWhitespaceChars("猫咪"),
		lipgloss.WithWhitespaceForeground(subtle),
	)
}

func main() {
	p := tea.NewProgram(
		New(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if err := p.Start(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
