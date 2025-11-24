package templates

import (
	"fmt"
	"html"
)

const (
	// maxMessagePreviewLength defines the maximum character length for message previews in emails.
	// Set to 150 to ensure previews display well across email clients while providing enough context for the user.
	maxMessagePreviewLength = 150
)

// ChatEmailData contains data for chat email templates
type ChatEmailData struct {
	SenderName     string
	ReceiverName   string
	ConversationID uint
	LastMessage    string
	UnreadCount    int
	ServiceTitle   string
	IsForBuyer     bool
}

// UnreadMessagesEmailData contains data for unread messages notification emails
type UnreadMessagesEmailData struct {
	SenderName     string
	ReceiverName   string
	ConversationID uint
	UnreadCount    int
	ServiceTitle   string
	IsForBuyer     bool
}

// ConversationDigestItem represents a single conversation with unread messages
type ConversationDigestItem struct {
	ConversationID uint
	SenderName     string
	ServiceTitle   string
	UnreadCount    int
}

// UnreadMessagesDigestData contains data for consolidated unread messages digest emails
type UnreadMessagesDigestData struct {
	ReceiverName  string
	Conversations []ConversationDigestItem
	TotalUnread   int
}

// NewChatMessage creates email notification for new chat message
func NewChatMessage(data ChatEmailData) (subject, htmlBody, textBody string) {
	subject = "💬 New Message in Your Conversation"

	unreadText := ""
	if data.UnreadCount > 1 {
		unreadText = fmt.Sprintf(`
			<div style="margin: 16px 0; padding: 12px; background-color: #fafafa; border-left: 3px solid #000000; border-radius: 4px;">
				<strong style="color: #09090b;">Unread Messages:</strong> <span style="color: #52525b;">%d new message%s</span>
			</div>
		`, data.UnreadCount, func() string {
			if data.UnreadCount > 1 {
				return "s"
			}
			return ""
		}())
	}

	// Safely truncate message preview using runes to handle UTF-8 correctly
	messagePreview := data.LastMessage
	runes := []rune(messagePreview)
	if len(runes) > maxMessagePreviewLength {
		messagePreview = string(runes[:maxMessagePreviewLength]) + "..."
	}

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">You have a new message from <strong>%s</strong>.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Message Preview:</strong>
	<div style="margin: 0; padding: 16px; background-color: #f4f4f5; border-radius: 6px; border-left: 3px solid #000000;">
		<p style="margin: 0; color: #52525b; font-style: italic;">"%s"</p>
	</div>
</div>
%s
<p style="margin: 20px 0 0 0;">Click below to view and reply to this message in your conversation.</p>`,
		html.EscapeString(data.SenderName),
		html.EscapeString(messagePreview),
		unreadText,
	)

	viewLink := fmt.Sprintf("https://app.hellobutter.io/marketplace/chats/%d", data.ConversationID)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Message",
		viewLink,
	)

	textBody = fmt.Sprintf(`You have a new message from %s.

Message Preview:
"%s"

%s

View message: %s`,
		data.SenderName,
		messagePreview,
		func() string {
			if data.UnreadCount > 1 {
				return fmt.Sprintf("You have %d unread messages.", data.UnreadCount)
			}
			return ""
		}(),
		viewLink,
	)

	return
}

// NewUnreadMessagesNotification creates email notification for unread messages
func NewUnreadMessagesNotification(data UnreadMessagesEmailData) (subject, htmlBody, textBody string) {
	messageWord := "message"
	if data.UnreadCount > 1 {
		messageWord = "messages"
	}

	subject = fmt.Sprintf("💬 You have %d unread %s", data.UnreadCount, messageWord)

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">You have <strong>%d unread %s</strong> from <strong>%s</strong> in your <strong>%s</strong> conversation.</p>

<div style="margin: 20px 0; padding: 16px; background-color: #f4f4f5; border-radius: 6px; border-left: 3px solid #000000;">
<p style="margin: 0; color: #09090b; font-weight: bold;">%d Unread %s</p>
</div>

<p style="margin: 20px 0 0 0;">Click below to view and reply to these messages.</p>`,
		data.UnreadCount,
		messageWord,
		html.EscapeString(data.SenderName),
		html.EscapeString(data.ServiceTitle),
		data.UnreadCount,
		func() string {
			if data.UnreadCount > 1 {
				return "Messages"
			}
			return "Message"
		}(),
	)

	viewLink := fmt.Sprintf("https://app.hellobutter.io/marketplace/chats/%d", data.ConversationID)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View Messages",
		viewLink,
	)

	textBody = fmt.Sprintf(`You have %d unread %s from %s in your %s conversation.

View messages: %s`,
		data.UnreadCount,
		messageWord,
		data.SenderName,
		data.ServiceTitle,
		viewLink,
	)

	return
}

// pluralizeMessage returns "message" or "messages" based on count
func pluralizeMessage(count int) string {
	if count > 1 {
		return "messages"
	}
	return "message"
}

// NewUnreadMessagesDigest creates a consolidated email notification for unread messages across multiple conversations
func NewUnreadMessagesDigest(data UnreadMessagesDigestData) (subject, htmlBody, textBody string) {
	messageWord := pluralizeMessage(data.TotalUnread)
	conversationWord := "conversation"
	if len(data.Conversations) > 1 {
		conversationWord = "conversations"
	}

	subject = fmt.Sprintf("💬 You have %d unread %s", data.TotalUnread, messageWord)

	// Build conversation list HTML
	conversationListHTML := ""
	for _, conv := range data.Conversations {
		conversationListHTML += fmt.Sprintf(`
		<div style="margin: 12px 0; padding: 16px; background-color: #fafafa; border-radius: 6px; border-left: 3px solid #000000;">
			<div style="margin-bottom: 8px;">
				<strong style="color: #09090b;">%s</strong> • <span style="color: #52525b;">%s</span>
			</div>
			<div style="color: #71717a; font-size: 14px;">
				%d unread %s
			</div>
			<div style="margin-top: 12px;">
				<a href="https://app.hellobutter.io/marketplace/chats/%d" style="color: #000000; text-decoration: underline;">View conversation →</a>
			</div>
		</div>`,
			html.EscapeString(conv.SenderName),
			html.EscapeString(conv.ServiceTitle),
			conv.UnreadCount,
			pluralizeMessage(conv.UnreadCount),
			conv.ConversationID,
		)
	}

	content := fmt.Sprintf(`<p style="margin: 0 0 20px 0;">You have <strong>%d unread %s</strong> across <strong>%d %s</strong>.</p>

<div style="margin: 20px 0;">
	<strong style="color: #09090b; display: block; margin-bottom: 12px;">Your Conversations:</strong>
	%s
</div>

<p style="margin: 20px 0 0 0;">Click any conversation above to view and reply to your messages.</p>`,
		data.TotalUnread,
		messageWord,
		len(data.Conversations),
		conversationWord,
		conversationListHTML,
	)

	htmlBody = BaseEmailTemplate(
		subject,
		content,
		"View All Conversations",
		"https://app.hellobutter.io/marketplace/chats",
	)

	// Build conversation list for text body
	conversationListText := ""
	for _, conv := range data.Conversations {
		conversationListText += fmt.Sprintf("\n- %s (%s): %d unread %s\n  View: https://app.hellobutter.io/marketplace/chats/%d\n",
			conv.SenderName,
			conv.ServiceTitle,
			conv.UnreadCount,
			pluralizeMessage(conv.UnreadCount),
			conv.ConversationID,
		)
	}

	textBody = fmt.Sprintf(`You have %d unread %s across %d %s.

Your Conversations:%s

View all conversations: https://app.hellobutter.io/marketplace/chats`,
		data.TotalUnread,
		messageWord,
		len(data.Conversations),
		conversationWord,
		conversationListText,
	)

	return
}
