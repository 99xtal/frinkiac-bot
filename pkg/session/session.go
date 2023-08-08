package session

import (
	"errors"

	"github.com/99xtal/frinkiac-bot/pkg/api"
)

type FrinkiacSession struct {
	Cursor              int
	SearchResults       []*api.Frame
	CurrentFrameCaption *api.Caption
	CurrentImageLink    string
}

func (s *FrinkiacSession) NextResult() (*api.Frame, error) {
	if s.Cursor >= len(s.SearchResults) {
		return nil, errors.New("Attempted to access out of bounds")
	}

	s.Cursor += 1
	return s.GetCurrentFrame(), nil
}

func (s *FrinkiacSession) PreviousResult() (*api.Frame, error) {
	if s.Cursor < 0 { 
		return nil, errors.New("Attempted to access out of bounds")
	}

	s.Cursor -= 1
	return s.GetCurrentFrame(),  nil
}

func (s *FrinkiacSession) GetCurrentFrame() *api.Frame {
	return s.SearchResults[s.Cursor]
}

func NewFrinkiacSession() *FrinkiacSession {
	return &FrinkiacSession{
		Cursor:        0,
		SearchResults: []*api.Frame{},
	}
}
