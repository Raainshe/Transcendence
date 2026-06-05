package ws

import "encoding/json"

const (
	TypeLobbyUpdated = "lobby.updated"
	TypeLobbyClosed  = "lobby.closed"
	TypeMatchStart   = "match.start"
	TypeError        = "error"
)

type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func NewEnvelope(typ string, payload any) (Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{Type: typ, Payload: raw}, nil
}

func LobbyRoomID(lobbyID string) string {
	return "lobby:" + lobbyID
}
