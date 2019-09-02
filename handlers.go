package main

import (
	"fmt"
	"net/url"
	"log"
	"regexp"
)

type Reactions map[int]string

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

var messageReactions = map[int]Reactions{}

var reactionEmoji = map[string]string{
	"reaction1": "😍",
	"reaction2": "🤔",
	"reaction3": "💩",
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

	reply_markup := `{"inline_keyboard": [[
		{"text": "😍", "callback_data": "reaction1"},
		{"text": "🤔", "callback_data": "reaction2"},
		{"text": "💩", "callback_data": "reaction3"}
	]]}`

	p := url.Values{
		"chat_id":    {fmt.Sprintf("%v", chatID)},
		"text":       {message},
		"parse_mode": {"Markdown"},
		"reply_markup": {reply_markup},
	}
	MakeTgapiRequest("sendMessage", p)
}

func HandleCallbackQuery(cq CallbackQuery) {
	// log.Printf("Handling query %v", cq.Data)
	matched, err := regexp.Match("reaction", []byte(cq.Data))
	if err != nil {
		return
	}
	if matched {
		updateMessageReaction(cq.From, cq.Message, cq.Data)
	}
}

func updateMessageReaction(from User, msg Message, reaction string) {
	userID, msgID, chatID := from.ID, msg.ID, msg.Chat.ID
	reactions, ok := messageReactions[msgID]
	if ok {
		reactions[userID] = reaction
		messageReactions[msgID] = reactions
	} else {
		messageReactions[msgID] = Reactions{ userID: reaction, }
	}
	reactions = messageReactions[msgID]
	counter := map[string]int{
		"reaction1": 0,
		"reaction2": 0,
		"reaction3": 0,
	}
	for _, reaction := range reactions {
		counter[reaction] = counter[reaction] + 1
	}
	log.Println("Updating reaction")
	buttons := map[string]string{}
	for r, c := range counter {
		if c > 0 {
			buttons[r] = fmt.Sprintf(`{"text": "%v %v", "callback_data": "%v"}`, reactionEmoji[r], c, r)
		} else {
			buttons[r] = fmt.Sprintf(`{"text": "%v", "callback_data": "%v"}`, reactionEmoji[r], r)
		}
	}
	reply_markup := fmt.Sprintf(`{"inline_keyboard": [[
		%v,
		%v,
		%v
	]]}`, buttons["reaction1"], buttons["reaction2"], buttons["reaction3"])
	p := url.Values{
		"chat_id":    {fmt.Sprintf("%v", chatID)},
		"message_id":    {fmt.Sprintf("%v", msgID)},
		"reply_markup": {reply_markup},
	}
	MakeTgapiRequest("editMessageReplyMarkup", p)
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
