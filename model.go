package main

import (
	"log"

	"github.com/pkg/errors"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/charm/kv"
	"github.com/charmbracelet/lipgloss"
)

type (
	sessionState    int
	newFlashcardMsg string
	inputState      int
)

const (
	answerState sessionState = iota
	questionState
	questionInput inputState = iota
	answerInput
)

// styles
var (
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	boxstyle  = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Border(lipgloss.DoubleBorder(), true).
			BorderForeground(highlight).
			Padding(0, 1).
			Width(25)
	width     = 96
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
)

type model struct {
	keys           [][]byte
	state          sessionState
	question       []byte
	answer         []byte
	kv             *kv.KV
	viewport       viewport.Model
	input          textinput.Model
	inputState     inputState
	addingQuestion string
	error          error
}

func New() model {
	kv, err := kv.OpenWithDefaults("flashcards")
	if err != nil {
		log.Fatal(err)
	}
	m := model{state: questionState, inputState: questionInput, kv: kv}
	if m.keys, err = kv.Keys(); err != nil {
		m.error = errors.Wrap(err, "unable to get keys from charm-kv")
	}
	q := m.getRandomQuestion()
	log.Println(string(q))
	m.question = q
	m.answer, err = m.kv.Get(q)
	if err != nil {
		m.error = errors.Wrap(err, "unable to get answer from charm-kv")
	}
	// init nested models
	m.viewport = viewport.New(8, 8)
	m.viewport.SetContent(string(m.question))
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
	var err error
	switch msg := msg.(type) {
	case syncKeysMsg:
		m.keys, err = m.kv.Keys()
		if err != nil {
			m.error = errors.Wrap(err, "unable to get keys from charm-kv")
		}
		cmds = append(cmds, m.getAnswerCmd)
	case inputAnswerMsg:
		m.input.SetValue("")
		m.input.Prompt = "Answer: "
		m.input.Placeholder = "your answer here..."
		m.input.Focus()
		m.addingQuestion = string(msg)
	case answerMsg:
		m.question = []byte(m.addingQuestion)
		m.answer = []byte(msg)
		m.inputState = questionInput
	case tea.KeyMsg:
		if m.input.Focused() {
			if m.inputState == questionInput {
				if msg.String() == "enter" {
					m.inputState = answerInput
					cmds = append(cmds, m.addQuestionCmd)
				}
			} else {
				if msg.String() == "enter" {
					m.input.SetValue("")
					m.input.Blur()
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
					m.viewport.SetContent(string(m.answer))
				} else {
					m.state = questionState
					m.viewport.SetContent(string(m.question))
				}
			}
			if msg.String() == "n" {
				// refresh new question and answer
				cmd = m.getFlashcardCmd
			}
			if msg.String() == "c" {
				m.input.Focus()
			}
		}
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	case newFlashcardMsg:
		m.question = []byte(msg)
		m.answer, err = m.kv.Get(m.question)
		if err != nil {
			m.error = errors.Wrap(err, "unable to get new answer")
		}
		m.viewport.SetContent(string(m.question))
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func showHelpMenu() string {
	return helpStyle("↑/↓: navigate  • c: add question • n: next • q: quit\n")
}

func (m model) View() string {
	// TODO: make styling not garbo
	// TODO: add errors to View
	if m.input.Focused() {
		return lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.PlaceVertical(9,
				lipgloss.Center,
				boxstyle.Render(m.viewport.View()),
				lipgloss.WithWhitespaceChars("猫咪"),
				lipgloss.WithWhitespaceForeground(subtle)),
			m.input.View(), showHelpMenu())
	} else {
		return lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.Place(width, 9,
				lipgloss.Center, lipgloss.Center,
				boxstyle.Render(m.viewport.View()),
				lipgloss.WithWhitespaceChars("猫咪"),
				lipgloss.WithWhitespaceForeground(subtle)),
			showHelpMenu())
	}
}
