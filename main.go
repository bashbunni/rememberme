package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type sessionState int

const (
	answerState sessionState = iota
	questionState
)

var qna = map[string]string{"what's 1 + 1": "2", "is red a warm colour?": "yes", "what is the best snack": "popcorn"}

type model struct {
	state    sessionState
	question string
	answer   string
}

type QNAMsg string

func (m model) getQNACmd() tea.Msg {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return QNAMsg("hello")
}

func New() model {
	return model{state: questionState, question: "", answer: ""}
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
			} else {
				m.state = questionState
			}
		}
		if msg.String() == "n" {
			// refresh new question and answer
		}
	}
	return m, nil
}

func View() {}

func main() {
	fmt.Println("do things")
}
