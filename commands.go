package main

import (
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	newQuestionMsg []byte
	syncKeysMsg    struct{}
	inputAnswerMsg string // hold question
	errMsg         struct {
		err error
	}
	answerMsg []byte
)

func (m model) getAnswerCmd() tea.Msg {
	answerMsg, err := m.kv.Get(m.question)
	if err != nil {
		return errMsg{err}
	}
	return answerMsg
}

func (m model) getFlashcardCmd() tea.Msg {
	question := m.getRandomQuestion()
	return newFlashcardMsg(question)
}

func (m model) addQuestionCmd() tea.Msg {
	question := m.input.Value()
	err := m.kv.Set([]byte(question), []byte(""))
	if err != nil {
		return errMsg{err}
	}
	return inputAnswerMsg(question)
}

func (m model) addAnswerCmd(question string) tea.Cmd {
	return func() tea.Msg {
		answer := m.input.Value()
		err := m.kv.Set([]byte(question), []byte(answer))
		if err != nil {
			return errMsg{err}
		}
		return syncKeysMsg{}
	}
}

/* helpers */

func (m *model) getRandomQuestion() []byte {
	if len(m.keys) > 0 {
		rand.Seed(time.Now().UnixNano())
		r := rand.Intn(len(m.keys))
		return m.keys[r]
	}
	return []byte("you don't have any questions!")
}
