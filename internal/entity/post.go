package entity

type Post struct {
	ChId          int
	PostId        int
	DonorChPostId int
}

func NewPost(chId, postId, donorChPostId int) Post {
	r := Post{
		ChId:          chId,
		PostId:        postId,
		DonorChPostId: donorChPostId,
	}
	return r
}
