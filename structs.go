package main

type Updates struct {
	Ok     bool
	Result []Update
}

type Update struct {
	UpdateID int `json:"update_id"`
	Message  Message
	CallbackQuery CallbackQuery `json:"callback_query"`
}

type Message struct {
	ID   int `json:"message_id"`
	From User
	Date int
	Chat Chat
	Text string
}

type User struct {
	ID        int
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string
	IsBot bool `json:"is_bot"`
}

type Chat struct {
	ID int
}

type CallbackQuery struct {
	ID string
	Data string
	Message Message
	From User
}
