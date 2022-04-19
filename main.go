package main

import (
	"fmt"
	"log"
	"os"

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

var qna = map[string]string{"what's 1 + 1": "2", "is red a warm colour?": "yes", "what is the best snack": "popcorn"}
var (
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	boxstyle  = lipgloss.NewStyle().
			Align(lipgloss.Center).
			Border(lipgloss.DoubleBorder(), true).
			BorderForeground(highlight).
			Padding(0, 1).
			Width(16)
	width  = 96
	subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
)

type model struct {
	qna        map[string]string
	state      sessionState
	question   string
	answer     string
	kv         *kv.KV
	viewport   viewport.Model
	input      textinput.Model
	inputState inputState
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
		cmd = m.addAnswerCmd(string(msg))
	case tea.KeyMsg:
		if m.input.Focused() {
			if m.inputState == questionInput {
				if msg.String() == "enter" {
					// TODO: add Cmd for setting question, then msg prompts for answer
					m.inputState = answerInput
					cmd = m.addQuestionCmd
				}
			}
			// else add answer, refresh data from cloud KV
			m.input, cmd = m.input.Update(msg)
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
		if msg.String() == "ctrl+c" {
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

func (m model) View() string {
	if m.input.Focused() {
		return lipgloss.Place(width, 9,
			lipgloss.Center, lipgloss.Center,
			boxstyle.Render(m.viewport.View()+"\n"+m.input.View()),
			lipgloss.WithWhitespaceChars("猫咪"),
			lipgloss.WithWhitespaceForeground(subtle))
	} else {
		return lipgloss.Place(width, 9,
			lipgloss.Center, lipgloss.Center,
			boxstyle.Render(m.viewport.View()),
			lipgloss.WithWhitespaceChars("猫咪"),
			lipgloss.WithWhitespaceForeground(subtle),
		)
	}
}

func main() {
	if os.Getenv("HELP_DEBUG") != "" {
		if f, err := tea.LogToFile("debug.log", "help"); err != nil {
			fmt.Println("Couldn't open a file for logging:", err)
			os.Exit(1)
		} else {
			defer func() {
				err = f.Close()
				if err != nil {
					log.Fatal(err)
				}
			}()
		}
	}
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
