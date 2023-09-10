package entity

type Post struct {
	ChId          int `json:"ch_id"`
	PostId        int `json:"post_id"`
	DonorChPostId int `json:"donor_ch_post_id"`
}
