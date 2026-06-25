package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"replive/config"
	"replive/dal"
	"replive/rep_api"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"gorm.io/gorm"
)

func HandleGetChatRooms(ctx context.Context, c *app.RequestContext) {
	resp := &Resp{}
	rooms, err := dal.GetChatRooms()
	if err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}
	resp.Data = rooms
	c.JSON(consts.StatusOK, resp)
}

type ChatMessageDTO struct {
	Id            int64  `json:"id"`
	UserId        string `json:"user_id"`
	DisplayName   string `json:"display_name"`
	ChatRoomId    string `json:"chat_room_id"`
	ChatMessageId string `json:"chat_message_id"`
	MsgType       int32  `json:"msg_type"`
	Content       string `json:"content"`
	ImageUrl      string `json:"image_url"`
	VideoUrl      string `json:"video_url"`
	TimeStr       string `json:"time_str"`
	SendTime      int64  `json:"send_time"`
}

type GetChatMessagesResp struct {
	Messages     []*ChatMessageDTO `json:"messages"`
	NextCursorId int64             `json:"next_cursor_id"`
	PrevCursorId int64             `json:"prev_cursor_id"`
	HasMore      bool              `json:"has_more"`
	HasOlder     bool              `json:"has_older"`
	HasNewer     bool              `json:"has_newer"`
	AnchorId     int64             `json:"anchor_id"`
}

func HandleGetChatMessages(ctx context.Context, c *app.RequestContext) {
	displayName := strings.TrimSpace(string(c.Query("display_name")))
	direction := strings.TrimSpace(string(c.Query("direction")))
	date := strings.TrimSpace(string(c.Query("date")))

	msgType, err := parseInt32Query(c, "msg_type", 0)
	if err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}
	pageSize, err := parseInt32Query(c, "page_size", 20)
	if err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 1000 {
		pageSize = 1000
	}
	cursorID, err := parseInt64Query(c, "cursor_id", 0)
	if err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}
	anchorID, err := parseInt64Query(c, "anchor_id", 0)
	if err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}

	if displayName == "" {
		c.JSON(consts.StatusOK, BadResp("display_name 或 (user_id + chat_room_id) 不能为空"))
		return
	}
	room, err := findChatRoom(displayName)
	if err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}
	if room.UserId == "" || room.ChatRoomId == "" {
		// 没找到，按空结果返回
		c.JSON(consts.StatusOK, &Resp{Data: &GetChatMessagesResp{Messages: []*ChatMessageDTO{}, NextCursorId: 0}})
		return
	}
	userID := room.UserId
	chatRoomID := room.ChatRoomId

	if date != "" && anchorID == 0 {
		anchorID, err = findFirstMessageIDByDate(userID, chatRoomID, msgType, date)
		if err != nil {
			c.JSON(consts.StatusOK, BadResp(err.Error()))
			return
		}
		if anchorID == 0 {
			c.JSON(consts.StatusOK, &Resp{Data: &GetChatMessagesResp{Messages: []*ChatMessageDTO{}, NextCursorId: 0}})
			return
		}
	}
	if anchorID > 0 && direction == "" {
		direction = "around"
	}

	msgs, hasOlder, hasNewer, err := queryChatMessages(userID, chatRoomID, msgType, cursorID, anchorID, pageSize, direction)
	if err != nil {
		hlog.Errorf("query chat_messages failed, uid=%s room=%s err=%v", userID, chatRoomID, err)
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}

	respMsgs, prevCursor, nextCursor := buildChatMessageResp(msgs)
	c.JSON(consts.StatusOK, &Resp{Data: &GetChatMessagesResp{
		Messages:     respMsgs,
		NextCursorId: nextCursor,
		PrevCursorId: prevCursor,
		HasMore:      hasOlder,
		HasOlder:     hasOlder,
		HasNewer:     hasNewer,
		AnchorId:     anchorID,
	}})
}

func HandleSearchChatMessages(ctx context.Context, c *app.RequestContext) {
	displayName := strings.TrimSpace(string(c.Query("display_name")))
	keyword := strings.TrimSpace(string(c.Query("keyword")))
	if displayName == "" {
		c.JSON(consts.StatusOK, BadResp("display_name 不能为空"))
		return
	}
	if keyword == "" {
		c.JSON(consts.StatusOK, BadResp("keyword 不能为空"))
		return
	}

	pageSize, err := parseInt32Query(c, "page_size", 20)
	if err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	cursorID, err := parseInt64Query(c, "cursor_id", 0)
	if err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}

	room, err := findChatRoom(displayName)
	if err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}
	if room.UserId == "" || room.ChatRoomId == "" {
		c.JSON(consts.StatusOK, &Resp{Data: &GetChatMessagesResp{Messages: []*ChatMessageDTO{}, NextCursorId: 0}})
		return
	}

	query := dal.ReadDB().Table(dal.ChatMessage{}.TableName()).
		Where("user_id = ? AND chat_room_id = ? AND content LIKE ?", room.UserId, room.ChatRoomId, "%"+keyword+"%")
	if cursorID > 0 {
		query = query.Where("id < ?", cursorID)
	}

	msgs := make([]dal.ChatMessage, 0, int(pageSize)+1)
	if err := query.Order("id desc").Limit(int(pageSize) + 1).Find(&msgs).Error; err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}
	hasMore := len(msgs) > int(pageSize)
	if hasMore {
		msgs = msgs[:pageSize]
	}

	respMsgs, _, nextCursor := buildChatMessageResp(msgs)
	c.JSON(consts.StatusOK, &Resp{Data: &GetChatMessagesResp{
		Messages:     respMsgs,
		NextCursorId: nextCursor,
		HasMore:      hasMore,
		HasOlder:     hasMore,
	}})
}

func findChatRoom(displayName string) (dal.ChatRoom, error) {
	var room dal.ChatRoom
	err := dal.ReadDB().Table(dal.ChatRoom{}.TableName()).
		Where("display_name = ?", displayName).
		Limit(1).
		Find(&room).Error
	if err != nil {
		hlog.Errorf("query chat_room failed, display_name=%s, err=%v", displayName, err)
		return room, err
	}
	return room, nil
}

func findFirstMessageIDByDate(userID string, chatRoomID string, msgType int32, date string) (int64, error) {
	start, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return 0, fmt.Errorf("date 参数非法，应为 yyyy-MM-dd: %v", err)
	}
	end := start.Add(24 * time.Hour)

	query := baseMessageQuery(userID, chatRoomID, msgType).
		Where("send_time >= ? AND send_time < ?", start.Unix(), end.Unix())
	var msg dal.ChatMessage
	if err := query.Order("send_time asc, id asc").Limit(1).Find(&msg).Error; err != nil {
		return 0, err
	}
	return msg.Id, nil
}

func queryChatMessages(userID string, chatRoomID string, msgType int32, cursorID int64, anchorID int64, pageSize int32, direction string) ([]dal.ChatMessage, bool, bool, error) {
	limit := int(pageSize) + 1
	query := baseMessageQuery(userID, chatRoomID, msgType)

	switch direction {
	case "newer":
		if cursorID <= 0 {
			return []dal.ChatMessage{}, false, false, nil
		}
		msgs := make([]dal.ChatMessage, 0, limit)
		if err := query.Where("id > ?", cursorID).Order("id asc").Limit(limit).Find(&msgs).Error; err != nil {
			return nil, false, false, err
		}
		hasNewer := len(msgs) > int(pageSize)
		if hasNewer {
			msgs = msgs[:pageSize]
		}
		hasOlder, err := hasMessageBefore(userID, chatRoomID, msgType, minMessageID(msgs))
		return msgs, hasOlder, hasNewer, err
	case "around":
		if anchorID <= 0 {
			return []dal.ChatMessage{}, false, false, nil
		}
		msgs := make([]dal.ChatMessage, 0, limit)
		if err := query.Where("id >= ?", anchorID).Order("id asc").Limit(limit).Find(&msgs).Error; err != nil {
			return nil, false, false, err
		}
		hasNewer := len(msgs) > int(pageSize)
		if hasNewer {
			msgs = msgs[:pageSize]
		}
		hasOlder, err := hasMessageBefore(userID, chatRoomID, msgType, minMessageID(msgs))
		return msgs, hasOlder, hasNewer, err
	default:
		if cursorID > 0 {
			query = query.Where("id < ?", cursorID)
		}
		msgs := make([]dal.ChatMessage, 0, limit)
		if err := query.Order("id desc").Limit(limit).Find(&msgs).Error; err != nil {
			return nil, false, false, err
		}
		hasOlder := len(msgs) > int(pageSize)
		if hasOlder {
			msgs = msgs[:pageSize]
		}
		hasNewer, err := hasMessageAfter(userID, chatRoomID, msgType, maxMessageID(msgs))
		return msgs, hasOlder, hasNewer, err
	}
}

func baseMessageQuery(userID string, chatRoomID string, msgType int32) *gorm.DB {
	query := dal.ReadDB().Table(dal.ChatMessage{}.TableName()).
		Where("user_id = ? AND chat_room_id = ?", userID, chatRoomID)
	if msgType != 0 {
		query = query.Where("msg_type = ?", msgType)
	}
	return query
}

func hasMessageBefore(userID string, chatRoomID string, msgType int32, id int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}
	var msg dal.ChatMessage
	err := baseMessageQuery(userID, chatRoomID, msgType).
		Where("id < ?", id).
		Order("id desc").
		Limit(1).
		Find(&msg).Error
	return msg.Id > 0, err
}

func hasMessageAfter(userID string, chatRoomID string, msgType int32, id int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}
	var msg dal.ChatMessage
	err := baseMessageQuery(userID, chatRoomID, msgType).
		Where("id > ?", id).
		Order("id asc").
		Limit(1).
		Find(&msg).Error
	return msg.Id > 0, err
}

func buildChatMessageResp(msgs []dal.ChatMessage) ([]*ChatMessageDTO, int64, int64) {
	respMsgs := make([]*ChatMessageDTO, 0, len(msgs))
	for i := range msgs {
		respMsgs = append(respMsgs, chatMessageDTO(msgs[i]))
	}
	return respMsgs, minMessageID(msgs), maxMessageID(msgs)
}

func chatMessageDTO(m dal.ChatMessage) *ChatMessageDTO {
	return &ChatMessageDTO{
		Id:            m.Id,
		UserId:        m.UserId,
		DisplayName:   m.DisplayName,
		ChatRoomId:    m.ChatRoomId,
		ChatMessageId: m.ChatMessageId,
		MsgType:       m.MsgType,
		Content:       m.Content,
		ImageUrl:      m.ImageUrl,
		VideoUrl:      m.VideoUrl,
		TimeStr:       m.TimeStr,
		SendTime:      m.SendTime,
	}
}

func minMessageID(msgs []dal.ChatMessage) int64 {
	minID := int64(0)
	for _, msg := range msgs {
		if minID == 0 || msg.Id < minID {
			minID = msg.Id
		}
	}
	return minID
}

func maxMessageID(msgs []dal.ChatMessage) int64 {
	maxID := int64(0)
	for _, msg := range msgs {
		if msg.Id > maxID {
			maxID = msg.Id
		}
	}
	return maxID
}

// HandleGetChatMedia 根据 chat_message_id 查本地文件并返回。
// 路由建议：GET /media/:file，其中 :file 形如 {chat_message_id}.mp4 / {chat_message_id}.jpeg
func HandleGetChatMedia(ctx context.Context, c *app.RequestContext) {
	file := strings.TrimSpace(c.Param("file"))
	if file == "" {
		c.SetStatusCode(consts.StatusBadRequest)
		_, _ = c.Write([]byte("missing file"))
		return
	}
	msgID, ext := splitMessageFile(file)
	if msgID == "" {
		c.SetStatusCode(consts.StatusBadRequest)
		_, _ = c.Write([]byte("invalid file"))
		return
	}
	msg, err := getChatMessageByID(msgID)
	if err != nil {
		c.SetStatusCode(consts.StatusInternalServerError)
		_, _ = c.Write([]byte(err.Error()))
		return
	}
	if msg.ChatMessageId == "" {
		c.SetStatusCode(consts.StatusNotFound)
		_, _ = c.Write([]byte("not found"))
		return
	}

	path := pickMediaPath(msg, ext)
	if path == "" {
		c.SetStatusCode(consts.StatusNotFound)
		_, _ = c.Write([]byte("media not found"))
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		c.SetStatusCode(consts.StatusNotFound)
		_, _ = c.Write([]byte("read media failed"))
		return
	}

	ct := mime.TypeByExtension(filepath.Ext(path))
	if ct == "" {
		ct = http.DetectContentType(data)
	}
	c.SetContentType(ct)
	c.SetStatusCode(consts.StatusOK)
	_, _ = c.Write(data)
}

// HandleDownloadVideo 兼容历史路由：GET /api/video/download?message_id=xxx
func HandleDownloadVideo(ctx context.Context, c *app.RequestContext) {
	msgID := strings.TrimSpace(string(c.Query("message_id")))
	if msgID == "" {
		c.JSON(consts.StatusOK, BadResp("message_id 不能为空"))
		return
	}
	msg, err := getChatMessageByID(msgID)
	if err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}
	if msg.ChatMessageId == "" {
		c.JSON(consts.StatusOK, BadResp("message_id 不存在"))
		return
	}
	if msg.VideoPath == "" {
		c.JSON(consts.StatusOK, BadResp("video_path 为空"))
		return
	}
	data, err := os.ReadFile(msg.VideoPath)
	if err != nil {
		c.JSON(consts.StatusOK, BadResp("读取视频失败"))
		return
	}
	c.SetContentType("video/mp4")
	c.SetStatusCode(consts.StatusOK)
	_, _ = c.Write(data)
}

func parseInt32Query(c *app.RequestContext, key string, def int32) (int32, error) {
	v := strings.TrimSpace(string(c.Query(key)))
	if v == "" {
		return def, nil
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%s 参数非法: %v", key, err)
	}
	return int32(n), nil
}

func parseInt64Query(c *app.RequestContext, key string, def int64) (int64, error) {
	v := strings.TrimSpace(string(c.Query(key)))
	if v == "" {
		return def, nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s 参数非法: %v", key, err)
	}
	return n, nil
}

func buildMediaURL(c *app.RequestContext, msgID, ext string) string {
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	scheme := "http"
	if v := c.Request.Header.Peek("X-Forwarded-Proto"); len(v) > 0 {
		scheme = string(v)
	} else if s := c.URI().Scheme(); len(s) > 0 {
		scheme = string(s)
	}
	host := string(c.Host())
	return fmt.Sprintf("%s://%s/media/%s%s", scheme, host, msgID, ext)
}

func splitMessageFile(file string) (msgID string, ext string) {
	file = strings.TrimPrefix(file, "/")
	if file == "" {
		return "", ""
	}
	// 只允许一段，避免路径穿越
	if strings.Contains(file, "/") || strings.Contains(file, "\\") {
		return "", ""
	}
	msgID = file
	if i := strings.LastIndex(file, "."); i > 0 {
		msgID = file[:i]
		ext = file[i:]
	}
	return msgID, ext
}

func getChatMessageByID(chatMessageID string) (*dal.ChatMessage, error) {
	var msg dal.ChatMessage
	err := dal.ReadDB().Table(dal.ChatMessage{}.TableName()).
		Where("chat_message_id = ?", chatMessageID).
		Limit(1).
		Find(&msg).Error
	if err != nil {
		hlog.Errorf("query chat_message failed, id=%s err=%v", chatMessageID, err)
		return nil, err
	}
	return &msg, nil
}

func pickMediaPath(msg *dal.ChatMessage, ext string) string {
	ext = strings.ToLower(ext)
	// 有 ext 时按 ext 选，没 ext 时优先视频再图片
	if ext != "" {
		if bytes.HasPrefix([]byte(ext), []byte(".mp4")) {
			return msg.VideoPath
		}
		if bytes.HasPrefix([]byte(ext), []byte(".jpg")) || bytes.HasPrefix([]byte(ext), []byte(".jpeg")) || bytes.HasPrefix([]byte(ext), []byte(".png")) || bytes.HasPrefix([]byte(ext), []byte(".webp")) {
			return msg.ImagePath
		}
	}
	if msg.VideoPath != "" {
		return msg.VideoPath
	}
	return msg.ImagePath
}

type SendMessageRequest struct {
	ChatRoomId string `json:"chat_room_id"`
	Content    string `json:"content"`
}

// HandleSendChatMessage 发送聊天消息
func HandleSendChatMessage(ctx context.Context, c *app.RequestContext) {
	if !config.Conf.SendChatEnabled {
		c.JSON(consts.StatusOK, BadResp("发送消息功能已关闭，请在 config.yaml 中设置 send_chat: true 开启"))
		return
	}
	var req SendMessageRequest
	if err := json.Unmarshal(c.Request.Body(), &req); err != nil {
		c.JSON(consts.StatusOK, BadResp("请求格式错误: "+err.Error()))
		return
	}
	req.ChatRoomId = strings.TrimSpace(req.ChatRoomId)
	req.Content = strings.TrimSpace(req.Content)
	if req.ChatRoomId == "" || req.Content == "" {
		c.JSON(consts.StatusOK, BadResp("chat_room_id 和 content 不能为空"))
		return
	}

	// 通过 chat_room_id 查找主播（Talent）的 ID 作为发送目标
	room, err := dal.GetChatRoomByChatRoomId(req.ChatRoomId)
	if err != nil {
		hlog.Errorf("HandleSendChatMessage: GetChatRoomByChatRoomId failed: %v", err)
		c.JSON(consts.StatusOK, BadResp("查找聊天室失败: "+err.Error()))
		return
	}
	hlog.Infof("HandleSendChatMessage: talent_user_id=%s chat_room_id=%s", room.UserId, req.ChatRoomId)
	if room == nil || room.UserId == "" {
		c.JSON(consts.StatusOK, BadResp("聊天室不存在或 user_id 为空"))
		return
	}

	resp, err := rep_api.SendChatMessage(room.UserId, req.ChatRoomId, req.Content)
	if err != nil {
		hlog.Errorf("HandleSendChatMessage: SendChatMessage failed: %v", err)
		c.JSON(consts.StatusOK, BadResp("发送消息失败: "+err.Error()))
		return
	}
	hlog.Infof("HandleSendChatMessage: success")
	c.JSON(consts.StatusOK, &Resp{Data: resp})
}
