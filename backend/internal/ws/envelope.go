package ws

import "encoding/json"

const (
	TypeLobbyUpdated = "lobby.updated"
	TypeLobbyClosed  = "lobby.closed"
	TypeMatchStart         = "match.start"
	TypeMatchEnded         = "match.ended"
	TypePlayerState        = "player.state"
	TypePlayerEliminated   = "player.eliminated"
	TypePlayerDisconnected = "player.disconnected"
	TypePlayerReconnected  = "player.reconnected"
	TypeError              = "error"

	MatrixWidth        = 10
	MatrixTotalHeight  = 40
	MatrixCellCount    = MatrixWidth * MatrixTotalHeight
	MaxPlayerStateScore = 9_999_999
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

func MatchRoomID(gameID string) string {
	return "match:" + gameID
}
