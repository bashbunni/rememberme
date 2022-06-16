package main

import (
	"fmt"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	badger "github.com/dgraph-io/badger/v3"
)

type KVListMsg struct {
	kvs map[string]string
}

type (
	GetAnswerMsg  string // hold question
	PlsRefreshMsg struct{}
	ErrMsg        struct {
		err error
	}
)

func (m model) getQNACmd() tea.Msg {
	question := getRandomQuestion(m.qna)
	return QNAMsg(question)
}

func (m model) addQuestionCmd() tea.Msg {
	question := m.input.Value()
	err := m.kv.Set([]byte(question), []byte(""))
	if err != nil {
		// TODO: handle errors in tui
		return ErrMsg{err}
	}
	return GetAnswerMsg(question)
}

func (m model) addAnswerCmd(question string) tea.Cmd {
	// get answer -> from text input
	// update the key val in KV
	// update list of QNAs
	return func() tea.Msg {
		answer := m.input.Value()
		err := m.kv.Set([]byte(question), []byte(answer))
		if err != nil {
			return ErrMsg{err}
		}
		m.input.SetValue("")
		m.input.Blur()
		return PlsRefreshMsg{}
	}
}

func (m model) createKVListCmd() tea.Msg {
	kvs := make(map[string]string)
	err := m.kv.Sync()
	if err != nil {
		return ErrMsg{err}
	}
	err = m.kv.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				kvs[string(k)] = string(v)
				fmt.Printf("key=%s, value=%s\n", k, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ErrMsg{err}
	}
	return KVListMsg{kvs}
}

/* helpers */

func getRandomQuestion(questions map[string]string) string {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(len(questions))
	j := 0
	for question := range questions {
		if j == r {
			return question
		}
		j++
	}
	return "you don't have any questions!"
}
