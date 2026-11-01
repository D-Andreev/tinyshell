package state

import (
	"fmt"
)

type State struct {
	state map[string]string
}

func NewState(currentPath string) *State {
	s := map[string]string{
		"currentDir": currentPath,
	}
	return &State{
		state: s,
	}
}

func (s *State) Get(key string) (string, error) {
	v, ok := s.state[key]
	if !ok {
		return "", fmt.Errorf("Could not read key: %s", key)
	}
	return v, nil
}

func (s *State) Set(key, value string) {
	s.state[key] = value
}
