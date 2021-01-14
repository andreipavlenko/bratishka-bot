package main

import (
	"fmt"
	"net/url"
	"regexp"
)

var MessageHandlers = map[string]func(msg Message){
	`^/start$`:    sayHello,
	`(?i)!замены`: handleSubstitutionsRequest,
	`(?i)!пары`:   handleLessonsSheduleRequest,
	`(?i)!чат`:    sendChatInfo,
}

func sayHello(msg Message) {
	text := fmt.Sprintf("Привет, %v 🙂", msg.From.FirstName)
	SendMessage(text, msg.Chat.ID)
}

func handleSubstitutionsRequest(msg Message) {
	SendSubstitutions(msg.Chat.ID)
}

func SendSubstitutions(chatID int) {
	message, err := GetSubstitutions()
	if err != nil {
		SendMessage("Ой, что-то пошло не так 😐", chatID)
		return
	}

	p := url.Values{
		"chat_id":    {fmt.Sprintf("%v", chatID)},
		"text":       {message},
		"parse_mode": {"Markdown"},
	}
	MakeTgapiRequest("sendMessage", p)
}

func HandleCallbackQuery(cq CallbackQuery) {
	isRequestSheduleCb, err := regexp.Match("group", []byte(cq.Data))

	if err != nil {
		return
	}

	if isRequestSheduleCb {
		RespondLessonsSheduleCallbackQuery(cq)
	}
}

func handleLessonsSheduleRequest(msg Message) {
	message := "Выбери своего бойца 🧐"

	replyMarkup := `{"inline_keyboard": [[
		{"text": "ЕІ-81", "callback_data": "group_ei81"},
		{"text": "П-81", "callback_data": "group_p81"}
	]]}`

	p := url.Values{
		"chat_id":      {fmt.Sprintf("%v", msg.Chat.ID)},
		"text":         {message},
		"parse_mode":   {"Markdown"},
		"reply_markup": {replyMarkup},
	}
	MakeTgapiRequest("sendMessage", p)
}

func sendChatInfo(msg Message) {
	chatID := msg.Chat.ID
	m := fmt.Sprintf("🤖 ID чата: %v", chatID)
	SendMessage(m, chatID)
}
