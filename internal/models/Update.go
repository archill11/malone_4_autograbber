package models

type GetChatResp struct {
	Result Chat `json:"result"`
	BotErrResp
}

type BotErrResp struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

type ApiBotResp struct {
	Result User `json:"result"`
	BotErrResp
}

type APIResponse struct {
	Result      []Update           `json:"result"`
	Parameters  ResponseParameters `json:"parameters"`
	BotErrResp
}

type ResponseParameters struct {
	MigrateToChatID int `json:"migrate_to_chat_id"`
	RetryAfter      int `json:"retry_after"`
}

type Update struct {
	UpdateId           int                 `json:"update_id"`
	Message            *Message            `json:"message"`
	ChannelPost        *Message            `json:"channel_post"`
	EditedChannelPost  *Message            `json:"edited_channel_post"`
	CallbackQuery      *CallbackQuery      `json:"callback_query"`
	InlineQuery        *InlineQuery        `json:"inline_query"`
	ChosenInlineResult *ChosenInlineResult `json:"chosen_inline_result"`
	MyChatMember       *ChatMemberUpdated  `json:"my_chat_member"`
	ChatMember         *ChatMemberUpdated  `json:"chat_member"`
	ChatJoinRequest    *ChatJoinRequest    `json:"chat_join_request"`
}

type Message struct {
	MessageId            int             `json:"message_id"`
	MessageThreadId      *int            `json:"message_thread_id"`
	From                 User            `json:"from"`
	Date                 int             `json:"date"`
	Chat                 *Chat           `json:"chat"`
	ForwardFrom          *User           `json:"forward_from"`
	ForwardFromChat      *Chat           `json:"forward_from_chat"`
	ForwardFromMessageId *int            `json:"forward_from_message_id"`
	Text                 string          `json:"text"`
	AuthorSignature      *string         `json:"author_signature"`
	SenderChat           *Chat           `json:"sender_chat"`
	Entities             []MessageEntity `json:"entities"`
	Animation            *Animation      `json:"animation"`
	Voice                *Voice          `json:"voice"`
	ReplyToMessage       *ReplyToMessage `json:"reply_to_message"`
	LeftChatMember       *User           `json:"left_chat_member"`
	Caption              *string         `json:"caption"`
	CaptionEntities      []MessageEntity `json:"caption_entities"`
	NewChatMembers       []User          `json:"new_chat_members"`
	MediaGroupId         *string         `json:"media_group_id"`
	Photo                []PhotoSize     `json:"photo"`
	Sticker              *Sticker        `json:"sticker"`
	Video                *Video          `json:"video"`
	VideoNote            *VideoNote      `json:"video_note"`
	IsTopicMessage       *bool           `json:"is_topic_message"`
	ReplyMarkup          *InlineKeyboardMarkup `json:"reply_markup"`
	HasMediaSpoiler      bool           `json:"has_media_spoiler"`
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

type Voice struct {
	FileId       string    `json:"file_id"`
	FileUniqueId string    `json:"file_unique_id"`
	Duration     int       `json:"duration"`
	MimeType     string    `json:"mime_type"`
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
	OldChatMember ChatMember     `json:"old_chat_member"`
	NewChatMember ChatMember     `json:"new_chat_member"`
	InviteLink    ChatInviteLink `json:"invite_link"`
}

type ChatMember struct {
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
	Description       string `json:"description"`
	AllAdministrators bool   `json:"all_members_are_administrators"`
	InviteLink        string `json:"invite_link"`
	LinkedChatId      int    `json:"linked_chat_id"`
	IsForum           bool   `json:"is_forum"`
}

type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	Url    string `json:"url,omitempty"`
}

type ChatInviteLink struct {
	InviteLink              string `json:"invite_link"`
	Name                    string `json:"name"`
	Creator                 User   `json:"creator"`
	CreatesJoinRequest      bool   `json:"creates_join_request"`
	PendingJoinRequestCount int    `json:"pending_join_request_count"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string  `json:"text"`
	Url          *string `json:"url"`
	CallbackData *string `json:"callback_data"`
}

type GetFileResp struct {
	Result struct {
		File_id        string `json:"file_id"`
		File_unique_id string `json:"file_unique_id"`
		File_path      string `json:"file_path"`
	} `json:"result"`
	BotErrResp
}

type SendMediaGroupResp struct {
	Result []SendMediaRespResult `json:"result"`
	BotErrResp
}

type SendMediaResp struct {
	Result SendMediaRespResult `json:"result"`
	BotErrResp
}

type SendMediaRespResult struct {
	MessageId int         `json:"message_id"`
	Caption   string      `json:"caption"`
	Chat      Chat        `json:"chat"`
	Video     Video       `json:"video"`
	Photo     []PhotoSize `json:"photo"`
}
