package entity

type Bot struct {
	Id          int    `json:"id"`
	Token       string `json:"token"`
	Username    string `json:"username"`
	Firstname   string `json:"first_name"`
	IsDonor     int    `json:"is_donor"`
	ChId        int    `json:"ch_id"`
	ChLink      string `json:"ch_link"`
	GroupLinkId int    `json:"group_link_id"`
	Lichka      string `json:"lichka"`
	ChIsSkam    int    `json:"ch_is_skam"`
}
