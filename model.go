package main

import (
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/charm/kv"
	"github.com/charmbracelet/lipgloss"
)

type (
	sessionState int
	QNAMsg       string
	inputState   int
)

const (
	answerState sessionState = iota
	questionState
	questionInput inputState = iota
	answerInput
)

// mock data
var qna = map[string]string{"what's 1 + 1": "2", "is red a warm colour?": "yes", "what is the best snack": "popcorn"}

// styles
var (
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	boxstyle  = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Border(lipgloss.DoubleBorder(), true).
			BorderForeground(highlight).
			Padding(0, 1).
			Width(16)
	width     = 96
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
)

type model struct {
	qna            map[string]string
	state          sessionState
	question       string
	answer         string
	kv             *kv.KV
	viewport       viewport.Model
	input          textinput.Model
	inputState     inputState
	addingQuestion string
}

func New() model {
	kv, err := kv.OpenWithDefaults("my-cute-db")
	if err != nil {
		log.Fatal(err)
	}
	m := model{state: questionState, kv: kv}
	//	msg := m.createKVList()
	//	m.qna = msg.(KVListMsg).kvs
	m.inputState = questionInput
	// init questions
	m.qna = qna
	q := getRandomQuestion(qna)
	log.Println(q)
	m.question = q
	m.answer = qna[q]
	// init nested models
	m.viewport = viewport.New(8, 8)
	m.viewport.SetContent(m.question)
	input := textinput.New()
	input.Prompt = "Question: "
	input.Placeholder = "your question here..."
	input.CharLimit = 250
	input.Width = 50
	m.input = input

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case KVListMsg:
		m.qna = msg.kvs
	case PlsRefreshMsg:
		cmd = m.createKVListCmd
	case GetAnswerMsg:
		m.input.SetValue("")
		m.input.Prompt = "Answer: "
		m.input.Placeholder = "your answer here..."
		m.input.Focus()
		m.addingQuestion = string(msg)
	case tea.KeyMsg:
		if m.input.Focused() {
			if m.inputState == questionInput {
				if msg.String() == "enter" {
					// TODO: add Cmd for setting question, then msg prompts for answer
					m.inputState = answerInput
					cmds = append(cmds, m.addQuestionCmd)
				}
			} else {
				if msg.String() == "enter" {
					cmds = append(cmds, m.addAnswerCmd(m.addingQuestion))
				}
			}
			// else add answer, refresh data from cloud KV
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			if msg.String() == "enter" {
				if m.state == questionState {
					m.state = answerState
					m.viewport.SetContent(m.answer)
				} else {
					m.state = questionState
					m.viewport.SetContent(m.question)
				}
			}
			if msg.String() == "n" {
				// refresh new question and answer
				cmd = m.getQNACmd
			}
			if msg.String() == "c" {
				m.input.Focus()
			}
		}
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	case QNAMsg:
		m.question = string(msg)
		m.answer = qna[string(msg)]
		m.viewport.SetContent(m.question)
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func showHelpMenu() string {
	return helpStyle("\n ↑/↓: navigate  • esc: back • c: add question • n: next • q: quit\n")
}

func (m model) View() string {
	if m.input.Focused() {
		return lipgloss.JoinVertical(lipgloss.Center, lipgloss.Place(width, 9,
			lipgloss.Center, lipgloss.Center,
			boxstyle.Render(m.viewport.View()),
			lipgloss.WithWhitespaceChars("猫咪"),
			lipgloss.WithWhitespaceForeground(subtle)),
			lipgloss.JoinVertical(lipgloss.Center, m.input.View(), showHelpMenu()))
	} else {
		return lipgloss.JoinVertical(lipgloss.Center, lipgloss.Place(width, 9,
			lipgloss.Center, lipgloss.Center,
			boxstyle.Render(m.viewport.View()),
			lipgloss.WithWhitespaceChars("猫咪"),
			lipgloss.WithWhitespaceForeground(subtle)), showHelpMenu())
	}
}
