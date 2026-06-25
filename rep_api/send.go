package rep_api

import (
	"crypto/rand"
	"fmt"
	"replive/model"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
)

func SendChatMessage(userID, chatRoomID, content string) (*model.SendChatMessageResponse, error) {
	userID = strings.TrimSpace(userID)
	chatRoomID = strings.TrimSpace(chatRoomID)
	content = strings.TrimSpace(content)
	if userID == "" || chatRoomID == "" || content == "" {
		return nil, fmt.Errorf("user_id, chat_room_id and content are required")
	}
	req := &model.SendChatMessageRequest{
		UserId:                              userID,
		ChatRoomId:                          chatRoomID,
		Content:                             content,
		ChatMessageId:                       newChatMessageID(),
		ConfirmContainsForbiddenWordsWarning: true,
	}
	resp, err := Post("user.v1.ChatService/SendChatMessage", req)
	if err != nil {
		return nil, fmt.Errorf("send chat message failed: %v", err)
	}
	out := new(model.SendChatMessageResponse)
	if err := proto.Unmarshal(resp, out); err != nil {
		return nil, fmt.Errorf("unmarshal SendChatMessageResponse failed: %v", err)
	}
	return out, nil
}

func newChatMessageID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func CreateCard(userID, liveID, content string, coinAmount int64) (*model.CreateCardResponse, error) {
	userID = strings.TrimSpace(userID)
	liveID = strings.TrimSpace(liveID)
	content = strings.TrimSpace(content)
	if userID == "" || content == "" || coinAmount <= 0 {
		return nil, fmt.Errorf("user_id, content and positive coin_amount are required")
	}
	req := &model.CreateCardRequest{
		UserId:     userID,
		Content:    content,
		CoinAmount: coinAmount,
		LiveId:     liveID,
	}
	resp, err := Post("user.v1.LiveService/CreateCard", req)
	if err != nil {
		return nil, fmt.Errorf("create card failed: %v", err)
	}
	out := new(model.CreateCardResponse)
	if err := proto.Unmarshal(resp, out); err != nil {
		return nil, fmt.Errorf("unmarshal CreateCardResponse failed: %v", err)
	}
	return out, nil
}
