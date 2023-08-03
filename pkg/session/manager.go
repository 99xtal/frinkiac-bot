package session

type SessionManager struct {
	sessions	map[string]*FrinkiacSession
}

func (m *SessionManager) Get(interactionId string) (*FrinkiacSession, error) {
	existingSession := m.sessions[interactionId]
	return existingSession, nil
}

func (m *SessionManager) Set(interactionId string, sessionUpdate *FrinkiacSession) error {
	m.sessions[interactionId] = sessionUpdate
	return nil
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*FrinkiacSession),
	}
}