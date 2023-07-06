package models

type GetChatResult struct {
	Result Chat `json:"result"`
}

type APIRBotresp struct {
	Ok     bool `json:"ok"`
	Result User `json:"result,omitempty"`
	ErrorCode int `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}

type APIResponse struct {
	Ok          bool               `json:"ok"`
	Result      []Update           `json:"result,omitempty"`
	ErrorCode   int                `json:"error_code,omitempty"`
	Description string             `json:"description,omitempty"`
	Parameters  ResponseParameters `json:"parameters,omitempty"`
}

type ResponseParameters struct {
	MigrateToChatID int `json:"migrate_to_chat_id,omitempty"`
	RetryAfter      int `json:"retry_after,omitempty"`
}

type Update struct {
	UpdateId           int                `json:"update_id"`
	Message            *Message            `json:"message"`
	ChannelPost        *Message            `json:"channel_post,omitempty"`
	CallbackQuery      *CallbackQuery      `json:"callback_query,omitempty"`
	InlineQuery        *InlineQuery        `json:"inline_query,omitempty"`
	ChosenInlineResult *ChosenInlineResult `json:"chosen_inline_result,omitempty"`
	MyChatMember       *ChatMemberUpdated  `json:"my_chat_member,omitempty"`
	ChatMember         *ChatMemberUpdated  `json:"chat_member,omitempty"`
	ChatJoinRequest    *ChatJoinRequest    `json:"chat_join_request,omitempty"`
}

type Message struct {
	MessageId            int             `json:"message_id"`
	MessageThreadId      *int             `json:"message_thread_id"`
	From                 User            `json:"from"`
	Date                 int             `json:"date"`
	Chat                 *Chat            `json:"chat"`
	ForwardFrom          *User            `json:"forward_from"`
	ForwardFromChat      *Chat            `json:"forward_from_chat"`
	ForwardFromMessageId *int             `json:"forward_from_message_id"`
	Text                 string          `json:"text"`
	AuthorSignature      *string          `json:"author_signature"`
	SenderChat           *Chat            `json:"sender_chat"`
	Entities             []MessageEntity `json:"entities"`
	Animation            *Animation       `json:"animation"`
	ReplyToMessage       *ReplyToMessage  `json:"reply_to_message"`
	LeftChatMember       *User            `json:"left_chat_member"`
	Caption              *string          `json:"caption"`
	CaptionEntities      []MessageEntity `json:"caption_entities"`
	NewChatMembers       []User          `json:"new_chat_members"`
	MediaGroupId         *string          `json:"media_group_id"`
	Photo                []PhotoSize     `json:"photo"`
	Sticker              *Sticker         `json:"sticker"`
	Video                *Video           `json:"video"`
	VideoNote            *VideoNote       `json:"video_note"`
	IsTopicMessage       *bool            `json:"is_topic_message"`
	// ReplyMarkup InlineKeyboardMarkup `json:"reply_markup"`
}

type ReplyToMessage struct {
	Chat              Chat              `json:"chat"`
	From              User              `json:"from"`
	ForumTopicCreated ForumTopicCreated `json:"forum_topic_created"`
	Date              int               `json:"date"`
	UpdateId          int               `json:"update_id"`
	MessageId         int               `json:"message_id"`
	Text              string            `json:"text"`
	IsTopicMessage    bool              `json:"is_topic_message"`
}

type ForumTopicCreated struct {
	Name string `json:"name"`
}

type Sticker struct {
	FileId       string `json:"file_id"`
	FileUniqueId string `json:"file_unique_id"`
	Type         string `json:"type"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

type Video struct {
	FileId       string    `json:"file_id"`
	FileUniqueId string    `json:"file_unique_id"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Duration     int       `json:"duration"`
	Thumbnail    PhotoSize `json:"thumbnail"`
	FileSize     int       `json:"file_size"`
}

type PhotoSize struct {
	FileId       string `json:"file_id"`
	FileUniqueId string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int    `json:"file_size"`
}

type InputMedia struct {
	Type            string          `json:"type"`
	Media           string          `json:"media"`
	Caption         string          `json:"caption"`
	CaptionEntities []MessageEntity `json:"caption_entities"`
}

type Animation struct {
	FileId       string    `json:"file_id"`
	FileUniqueId string    `json:"file_unique_id"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Duration     int       `json:"duration"`
	Thumbnail    PhotoSize `json:"thumbnail"`
	FileSize     int       `json:"file_size"`
}

type VideoNote struct {
	FileId       string    `json:"file_id"`
	FileUniqueId string    `json:"file_unique_id"`
	Thumbnail    PhotoSize `json:"thumbnail"`
}

type ChatMemberUpdated struct {
	Chat          Chat           `json:"chat"`
	From          User           `json:"from"`
	Date          int            `json:"date"`
	OldChatMember ChatMember1    `json:"old_chat_member"`
	NewChatMember ChatMember1    `json:"new_chat_member"`
	InviteLink    ChatInviteLink `json:"invite_link"`
}

type ChatMember1 struct {
	Status string `json:"status"`
	User   User   `json:"user"`
}

type ChatJoinRequest struct {
	Chat     Chat `json:"chat"`
	From     User `json:"from"`
	Date     int  `json:"date"`
	UpdateId int  `json:"update_id"`
}

type CallbackQuery struct {
	Data    string  `json:"data"`
	From    User    `json:"from"`
	Message Message `json:"message"`
}

type InlineQuery struct {
	Query string `json:"query"`
	From  User   `json:"from"`
}

type ChosenInlineResult struct {
	From            User   `json:"from"`
	InlineMessageId User   `json:"inline_message_id"`
	Query           string `json:"query"`
}

type SendMessage struct {
	Ok     bool   `json:"ok"`
	Result Result `json:"result"`
}

type Result struct {
	MessageId int    `json:"message_id"`
	Date      int    `json:"date"`
	Text      string `json:"text"`
	From      User   `json:"from"`
	Chat      Chat   `json:"chat"`
}

type User struct {
	Id           int    `json:"id"`
	FirstName    string `json:"first_name"`
	UserName     string `json:"username"`
	LanguageCode string `json:"language_code"`
	IsBot        bool   `json:"is_bot"`
	InviteLink   string `json:"invite_link"`
}

type Chat struct {
	Id                int    `json:"id"`
	FirstName         string `json:"first_name"`
	UserName          string `json:"username"`
	Type              string `json:"type"`
	Title             string `json:"title"`
	AllAdministrators bool   `json:"all_members_are_administrators"`
	InviteLink        string `json:"invite_link"`
	LinkedChatId      int    `json:"linked_chat_id"`
	IsForum           bool   `json:"is_forum"`
}

type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	Url    string `json:"url"`
}

type ChatInviteLink struct {
	InviteLink              string `json:"invite_link"`
	Name                    string `json:"name"`
	Creator                 User   `json:"creator"`
	CreatesJoinRequest      bool   `json:"creates_join_request"`
	PendingJoinRequestCount int    `json:"pending_join_request_count"`
}

// type InlineKeyboardMarkup struct {
// 	Inlinekeyboard []InlineKeyboardButton `json:"inline_keyboard"`
// }

// type InlineKeyboardButton struct {
// 	Text string `json:"text"`
// 	Url string `json:"url"`
// }
