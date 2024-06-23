package entity

type GroupLink struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Link  string `json:"link"`
	UserCreator int `json:"user_creator"`
}