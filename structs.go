package main

type Updates struct {
	Ok     bool
	Result []Update
}

type Update struct {
	UpdateID int `json:"update_id"`
	Message  Message
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
}

type Chat struct {
	ID int
}
