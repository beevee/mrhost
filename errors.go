package mrhost

type NoMoreQuestionsError struct{}

func (err NoMoreQuestionsError) Error() string {
	return "there are no more questions in the database for this chat"
}
