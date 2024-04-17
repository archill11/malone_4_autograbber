package entity

type User struct {
	Id           int    `json:"id"`
	Username     string `json:"username"`
	Firstname    string `json:"firstname"`
	IsAdmin      int    `json:"is_admin"`
	IsSuperAdmin int    `json:"is_super_admin"`
	IsUser       int    `json:"is_user"`
}