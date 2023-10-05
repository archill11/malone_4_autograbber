package models

type ErrResp struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type UB_add_channel_resp struct {
	Channel struct {
		Id     int    `json:"id"`
		Title  string `json:"title"`
		IsScam bool   `json:"is_scam"`
	} `json:"channel"`
	ErrResp
}