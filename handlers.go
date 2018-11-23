package main

import (
	"fmt"
	"net/url"
)

var MessageHandlers = map[string]func(msg Message){
	`^/start$`: sayHello,
	`(?i)Братишка.*подскажи замены`:              handleSubstitutionsRequest,
	`(?i)Молодец.*братишка`:                      sayThankYou,
	`(?i)Спасибо.*братишка`:                      sayPlease,
	`(?i)Привет.*братишка`:                       sayHello,
	`(?i)Братишка.*привет`:                       sayHello,
	`(?i)Братишка.*спишь?`:                       sayNoSleep,
	`(?i)Братишка.*сообщи когда появятся замены`: watchUpdates,
	`(?i)Братишка.*ID`:                           sayChatID,
	`(?i)Плохой.*братишка`:                       sayWasOffensively,
	`(?i)Что вы\?`:                               sayThinking,
	`(?i)Спокойной ночи`:                         sayGoodNight,
	`(?i)Спите\?`:                                saySleeping,
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

func sayThankYou(msg Message) {
	text := fmt.Sprintf("Спасибо %v, ты тоже ничего 🤗", msg.From.FirstName)
	SendMessage(text, msg.Chat.ID)
}

func sayPlease(msg Message) {
	SendMessage("Пожалуйста 😉", msg.Chat.ID)
}

func sayNoSleep(msg Message) {
	SendMessage("Я не сплю, я думаю 🤔", msg.Chat.ID)
}

func watchUpdates(msg Message) {
	SendMessage("Хорошо 😊", msg.Chat.ID)
}

func sayChatID(msg Message) {
	chatID := msg.Chat.ID
	messageText := fmt.Sprintf("Вот, держи %v 🙃", chatID)
	SendMessage(messageText, chatID)
}

func sayWasOffensively(msg Message) {
	SendMessage(fmt.Sprintf("А вот сейчас обидно было 😥"), msg.Chat.ID)
}

func sayThinking(msg Message) {
	SendMessage("Думаем 🤔", msg.Chat.ID)
}

func sayGoodNight(msg Message) {
	SendMessage("Спокойной ночи 😚", msg.Chat.ID)
}

func saySleeping(msg Message) {
	SendMessage("Спим 😴", msg.Chat.ID)
}
