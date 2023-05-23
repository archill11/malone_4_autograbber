package entity

type GroupLink struct {
	Id    int
	Title string
	Link  string
}

func NewGroupLink(title, link string) GroupLink {
	b := GroupLink{
		Title: title,
		Link:  link,
	}
	return b
}
