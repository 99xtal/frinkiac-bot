package session

import (
	"github.com/99xtal/frinkiac-bot/pkg/api"
)

type FrinkiacSession struct {
	Cursor int
	memeMode bool
	SearchResults []*api.Frame
}

func (s *FrinkiacSession) NextPage() {
	if (s.Cursor < len(s.SearchResults)) {
		s.Cursor += 1
	}
}

func (s *FrinkiacSession) PrevPage() {
	if (s.Cursor > 0) {
		s.Cursor -= 1
	}
}

func (s *FrinkiacSession) GetCurrentFrame() *api.Frame {
	return s.SearchResults[s.Cursor]
}

func (s *FrinkiacSession) toggleMemeMode() {
	s.memeMode = !s.memeMode;
}

func NewFrinkiacSession() *FrinkiacSession {
	return &FrinkiacSession{
		Cursor: 0,
		SearchResults: []*api.Frame{},
		memeMode: false,
	}
}