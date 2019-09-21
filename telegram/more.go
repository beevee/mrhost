package telegram

import (
	"bytes"
	"fmt"

	"mrhost"

	"gopkg.in/tucnak/telebot.v2"
)

const (
	NoMoreQuestions = "У господина ведущего закончились вопросы. Музыкальная пауза!"
	CantAskForMore  = "Сначала господин %s должен дать ответ на текущий вопрос."
)

func (b *ChatBot) handleMore(message *telebot.Message) {
	b.Logger.Log("msg", "handling /more command", "chatid", message.Chat.ID)

	chatMeta, err := b.Repository.GetChatMeta(message.Chat.ID)
	if err != nil {
		b.Logger.Log("msg", "cannot get chat meta", "chatid", message.Chat.ID, "error", err)
		return
	}

	if chatMeta.State != mrhost.ChatIdleState {
		b.Logger.Log("msg", "chat is not idle, question not allowed", "chatid", message.Chat.ID)
		if _, err := b.telebot.Send(message.Chat, fmt.Sprintf(CantAskForMore, chatMeta.CurrentSender.LastName)); err != nil {
			b.Logger.Log("msg", "cannot send message", "chatid", message.Chat.ID, "error", err)
		}
		return
	}

	question, questionID, err := b.Repository.GetNextQuestion(chatMeta.CurrentQuestionID)
	if err != nil {
		b.Logger.Log("msg", "cannot get next question", "chatid", message.Chat.ID, "error", err)
		if _, ok := err.(mrhost.NoMoreQuestionsError); ok {
			if _, err := b.telebot.Send(message.Chat, NoMoreQuestions); err != nil {
				b.Logger.Log("msg", "cannot send message", "chatid", message.Chat.ID, "error", err)
			}
		}
		return
	}

	var sentMessage *telebot.Message
	if question.Image == nil {
		sentMessage, err = b.telebot.Send(message.Chat, question.Text)
	} else {
		photo := telebot.Photo{
			File:    telebot.FromReader(bytes.NewReader(question.Image)),
			Caption: question.Text,
		}
		_, err = b.telebot.Send(message.Chat, photo)
	}
	if err != nil {
		b.Logger.Log("msg", "cannot send question", "chatid", message.Chat.ID, "error", err)
		return
	}
	if err := b.telebot.Pin(sentMessage); err != nil {
		b.Logger.Log("msg", "cannot pin question", "chatid", message.Chat.ID, "error", err)
	}

	chatMeta.CurrentQuestionID = questionID
	chatMeta.CurrentSender.ID = message.Sender.ID
	chatMeta.CurrentSender.LastName = message.Sender.LastName
	chatMeta.State = mrhost.ChatInQuestionState

	if err := b.Repository.SetChatMeta(message.Chat.ID, chatMeta); err != nil {
		b.Logger.Log("msg", "cannot set in question chat state", "chatid", message.Chat.ID, "error", err)
	}
}
