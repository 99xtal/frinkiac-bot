package session

import (
	"fmt"
	"log"
)

type SessionManager struct {
	sessions map[string]*FrinkiacSession
}

func (m *SessionManager) Get(interactionId string) (*FrinkiacSession, error) {
	existingSession := m.sessions[interactionId]
	return existingSession, nil
}

func (m *SessionManager) Set(interactionId string, sessionUpdate *FrinkiacSession) error {
	m.sessions[interactionId] = sessionUpdate
	return nil
}

func (m *SessionManager) Delete(interactionId string) {
	log.Println(fmt.Sprintf("Deleting session with interactionId: %s", interactionId))
	delete(m.sessions, interactionId)
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*FrinkiacSession),
	}
}
