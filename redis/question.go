package redis

import (
	"encoding/json"
	"fmt"

	"mrhost"

	"github.com/go-redis/redis"
)

func getQuestionKey(questionID int) string {
	return fmt.Sprintf("questions:%d", questionID)
}

func (r Repository) GetNextQuestion(currentQuestionID int) (mrhost.Question, int, error) {
	question, err := r.GetQuestionByID(currentQuestionID + 1)
	if err == redis.Nil {
		return question, 0, mrhost.NoMoreQuestionsError{}
	}
	if err != nil {
		return question, 0, err
	}
	return question, currentQuestionID + 1, err
}

func (r Repository) GetQuestionByID(questionID int) (mrhost.Question, error) {
	var question mrhost.Question
	result, err := r.client.Get(getQuestionKey(questionID)).Result()
	if err != nil {
		return mrhost.Question{}, err
	}
	err = json.Unmarshal([]byte(result), &question)
	return question, err
}

func (r Repository) AddQuestion(question mrhost.Question) error {
	nextQuestionID, err := r.client.Get("next-question-id").Int()
	if err != nil && err != redis.Nil {
		return err
	}
	if nextQuestionID == 0 {
		nextQuestionID = 1 // question ids start with 1
	}

	questionJSON, err := json.Marshal(question)
	if err != nil {
		return err
	}
	_, err = r.client.Set(getQuestionKey(nextQuestionID), questionJSON, 0).Result()
	if err != nil {
		return err
	}
	_, err = r.client.Set("next-question-id", nextQuestionID+1, 0).Result()
	if err != nil {
		return err
	}
	return nil
}
