package entity

type User struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
	IsAdmin   int    `json:"is_admin"`
}