package telegram

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"mrhost"

	"gopkg.in/tucnak/telebot.v2"
)

const (
	CantGiveAnswer = "Подождите, господин %s. Я еще вопрос не задал."
	CorrectAnswer  = "Ответ: %s\n\nКомментарий: %s\n\nАвтор вопроса: %s"
	ChatTitle      = "Бесконечный ЧГК %d:%d"
)

func (b *ChatBot) handleAnswer(message *telebot.Message) {
	b.Logger.Log("msg", "handling /answer command", "chatid", message.Chat.ID)

	chatMeta, err := b.Repository.GetChatMeta(message.Chat.ID)
	if err != nil {
		b.Logger.Log("msg", "cannot get chat meta", "chatid", message.Chat.ID, "error", err)
		return
	}

	if chatMeta.State != mrhost.ChatInQuestionState {
		b.Logger.Log("msg", "chat is not in question, answer not allowed", "chatid", message.Chat.ID)
		if _, err := b.telebot.Send(message.Chat, fmt.Sprintf(CantGiveAnswer, message.Sender.LastName)); err != nil {
			b.Logger.Log("msg", "cannot send message", "chatid", message.Chat.ID, "error", err)
		}
		return
	}

	question, err := b.Repository.GetQuestionByID(chatMeta.CurrentQuestionID)
	if err != nil {
		b.Logger.Log("msg", "cannot get question", "questionid", chatMeta.CurrentQuestionID, "chatid", message.Chat.ID, "error", err)
		return
	}

	if question.AnswerImage == nil {
		_, err = b.telebot.Send(message.Chat, fmt.Sprintf(CorrectAnswer, question.AnswerText, question.AnswerComment, question.Author))
	} else {
		photo := &telebot.Photo{
			File:    telebot.FromReader(bytes.NewReader(question.AnswerImage)),
			Caption: fmt.Sprintf(CorrectAnswer, question.AnswerText, question.AnswerComment, question.Author),
		}
		_, err = b.telebot.Send(message.Chat, photo)
	}
	if err != nil {
		b.Logger.Log("msg", "cannot send answer", "chatid", message.Chat.ID, "error", err)
		return
	}

	if answerIsValid(question.AnswerType, question.AnswerMatch, strings.Fields(message.Text)[1:]) {
		chatMeta.PlayersScore++
	} else {
		chatMeta.CasinoScore++
	}

	chatMeta.State = mrhost.ChatIdleState
	if err := b.Repository.SetChatMeta(message.Chat.ID, chatMeta); err != nil {
		b.Logger.Log("msg", "cannot set in question chat state", "chatid", message.Chat.ID, "error", err)
	}

	if err := b.telebot.SetGroupTitle(message.Chat, fmt.Sprintf(ChatTitle, chatMeta.PlayersScore, chatMeta.CasinoScore)); err != nil {
		b.Logger.Log("msg", "cannot set chat title", "chatid", message.Chat.ID, "error", err)
	}

	if err := b.telebot.Unpin(message.Chat); err != nil {
		b.Logger.Log("msg", "cannot unpin question", "chatid", message.Chat.ID, "error", err)
	}
}

func answerIsValid(answerType mrhost.AnswerType, answerMatch [][]string, playerAnswer []string) (matched bool) {
	for i := range playerAnswer {
		playerAnswer[i] = normalize(playerAnswer[i])
	}

	switch answerType {
	case mrhost.AnswerStrictType:
		for _, matchCandidate := range answerMatch {
			if slicesAreEqual(matchCandidate, playerAnswer) {
				matched = true
			}
		}
	case mrhost.AnswerContainsType:
		for _, matchCandidate := range answerMatch {
			if sliceContainsAllSubstrings(matchCandidate, playerAnswer) {
				matched = true
			}
		}
	}
	return
}

func sliceContainsAllSubstrings(substrings, sliceToCheck []string) bool {
OUTER:
	for _, substring := range substrings {
		for _, fullString := range sliceToCheck {
			if strings.Contains(fullString, substring) {
				continue OUTER
			}
		}
		return false
	}
	return true
}

func slicesAreEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func normalize(s string) string {
	result := strings.Builder{}
	for _, r := range s {
		if unicode.IsDigit(r) || unicode.IsLetter(r) {
			result.WriteRune(unicode.ToLower(r))
		}
	}
	return result.String()
}
