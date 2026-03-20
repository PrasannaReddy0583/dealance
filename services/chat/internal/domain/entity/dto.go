package entity

type CreateConversationRequest struct {
	ConvType      string   `json:"conv_type" validate:"required,oneof=DM GROUP DEAL_ROOM"`
	Title         string   `json:"title,omitempty" validate:"omitempty,max=200"`
	ParticipantIDs []string `json:"participant_ids" validate:"required,min=1"`
	DealID        string   `json:"deal_id,omitempty" validate:"omitempty,uuid"`
}

type SendMessageRequest struct {
	ConversationID string `json:"conversation_id" validate:"required,uuid"`
	MessageType    string `json:"message_type,omitempty" validate:"omitempty,oneof=TEXT IMAGE FILE AUDIO VIDEO"`
	Body           string `json:"body" validate:"required,max=4000"`
	MediaURL       string `json:"media_url,omitempty"`
	ReplyToID      string `json:"reply_to_id,omitempty" validate:"omitempty,uuid"`
}

type EditMessageRequest struct {
	Body string `json:"body" validate:"required,max=4000"`
}

type ConversationResponse struct {
	ID            string `json:"id"`
	ConvType      string `json:"conv_type"`
	Title         string `json:"title,omitempty"`
	LastMessageAt string `json:"last_message_at"`
	UnreadCount   int    `json:"unread_count"`
}

type MessageResponse struct {
	ID          string `json:"id"`
	SenderID    string `json:"sender_id"`
	MessageType string `json:"message_type"`
	Body        string `json:"body"`
	MediaURL    string `json:"media_url,omitempty"`
	ReplyToID   string `json:"reply_to_id,omitempty"`
	IsEdited    bool   `json:"is_edited"`
	CreatedAt   string `json:"created_at"`
}

// WebSocket message envelope
type WSMessage struct {
	Type    string      `json:"type"`    // MESSAGE, TYPING, READ, REACTION, PRESENCE
	Payload interface{} `json:"payload"`
}
