package mrhost

type Repository interface {
	GetChatMeta(chatID int64) (ChatMeta, error)
	SetChatMeta(chatID int64, meta ChatMeta) error
	GetNextQuestion(currentQuestionID int) (Question, int, error)
	GetQuestionByID(questionID int) (Question, error)
	AddQuestion(question Question) error
}

type Logger interface {
	Log(...interface{}) error
}

type ChatMeta struct {
	State         ChatState
	CurrentSender struct {
		ID       int
		LastName string
	}
	CurrentQuestionID int
	CasinoScore       int
	PlayersScore      int
}

type ChatState int

const (
	ChatIdleState ChatState = iota
	ChatInQuestionState
)

type Question struct {
	Text          string
	Image         []byte
	AnswerMatch   [][]string
	AnswerText    string
	AnswerImage   []byte
	AnswerType    AnswerType
	AnswerComment string
	Author        string
}

type AnswerType int

const (
	AnswerStrictType = iota
	AnswerContainsType
)
