package rep_api

import (
	"context"
	"encoding/json"
	"fmt"
	"replive/model"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
)

// akh
// userId: 481074e0-5f19-4cdb-90d2-21ff2e9544ac
// chatroomId: 395b8889-204d-460a-8c91-60bee4c9ba2d

// eri
//     UserId:      "950c24bb-6cf2-49a4-b728-7b14fd623caa",
//     ChatRoomId:  "d646d12f-2221-4267-ad45-f4640d665206",

// GetListChatMessage
// 1. 只传 backward = false, 不传 cursorChatMessageId，会拉取最新一条。next_page_cursor_msg_id = 就是这条id
// 2. 传 cursor_id, backward = false, 则以这条为起点，一直往后直到拉取到最新
// 3. 传 cursor_id, backward = true, 则以这条为起点，一直往前直到拉取到首条打招呼的，就结束
// 4. 不传 cursor_id，backward = true, 则以第二条为起点（好怪？）
// 5. page 最大是100，超了会报错。
func GetListChatMessage() error {
	uri := "user.v1.ChatService/ListChatMessages"
	// hnk
	//msgReq := &model.ListChatMessagesRequest{ // hnk
	//	UserId:      "6c1e09da-6c91-42f7-ab9e-e8aaba49eda7",
	//	ChatRoomId:  "3e5b1ab6-0d22-4912-8764-19f2173372a8",
	//	MaxPageSize: 60,
	//	//CursorChatMessageId: "db735b66-47a8-4d8f-89d2-c075cf21a4b8",
	//	Backward: false,
	//}
	// eri:
	//msgReq := &model.ListChatMessagesRequest{ // eri
	//	UserId:              "950c24bb-6cf2-49a4-b728-7b14fd623caa",
	//	ChatRoomId:          "d646d12f-2221-4267-ad45-f4640d665206",
	//	MaxPageSize:         100,
	//	CursorChatMessageId: "3095dbc0-9349-4372-a342-5e63638fe288",
	//	Backward:            false,
	//}
	// akh:
	msgReq := &model.ListChatMessagesRequest{ // akh
		UserId:      "481074e0-5f19-4cdb-90d2-21ff2e9544ac",
		ChatRoomId:  "395b8889-204d-460a-8c91-60bee4c9ba2d",
		MaxPageSize: 22,
		//CursorChatMessageId: "84fe08a4-c767-422d-aaf6-58dc19c74902",
		Backward: false,
	}
	//msgReq.CursorChatMessageId = "84fe08a4-c767-422d-aaf6-58dc19c74902"
	resp, err := GetReplive(uri, msgReq)
	if err != nil {
		return fmt.Errorf("failed to get list chat message: %v", err)
	}
	//PrintBuf(resp)
	msgResp := new(model.ListChatMessagesResponse)
	if err := proto.Unmarshal(resp, msgResp); err != nil {
		return fmt.Errorf("failed to unmarshal ListChatMessagesResponse: %v", err)
	}
	for _, msg := range msgResp.Messages {
		timeVal := time.Unix(msg.Timestamp.Seconds+3600, msg.Timestamp.Nanos)
		msg.TimeStr = timeVal.Format("2006-01-02 15:04:05.000")
	}
	for i, msg := range msgResp.Messages {
		if i < 2 {
			continue
		}
		msg.UserId = ""
		msg.ChatRoomId = ""
		msg.UserProfile = nil
		if msg.ImageUrl != "" {
			fmt.Println(msg.ImageUrl)
		}
		if msg.VideoUrl != "" {
			fmt.Println(msg.VideoUrl)
		}
	}
	fmt.Println(ToDebugString(msgResp))
	return nil
}

var (
	chatApiLimiter = rate.NewLimiter(10, 10)
)

func GetChatMessages(ctx context.Context, uid, roomId string, cursorMsgId *string, backward bool, size int32) ([]*model.ListChatMessages, string, error) {
	uri := "user.v1.ChatService/ListChatMessages"
	req := &model.ListChatMessagesRequest{
		UserId:      uid,
		ChatRoomId:  roomId,
		MaxPageSize: size,
		Backward:    backward,
	}
	if cursorMsgId != nil {
		req.CursorChatMessageId = *cursorMsgId
	}
	for {
		if chatApiLimiter.Allow() {
			break
		}
		hlog.Warnf("GetChatMessages rate limit, query uid: %s, msgId: %v, backward: %v", uid, cursorMsgId, backward)
		time.Sleep(time.Millisecond * 200)
	}
	resp, err := GetReplive(uri, req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get list chat message: %v", err)
	}
	msgResp := new(model.ListChatMessagesResponse)
	if err := proto.Unmarshal(resp, msgResp); err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal ListChatMessagesResponse: %v", err)
	}
	for _, msg := range msgResp.Messages {
		timeVal := time.Unix(msg.Timestamp.Seconds+3600, msg.Timestamp.Nanos) // 转换回日本时间
		msg.TimeStr = timeVal.Format("2006-01-02 15:04:05.000")
	}
	return msgResp.Messages, msgResp.NextPageCursorMessageId, nil
}

func GetChatRooms() ([]*model.ChatRoom, error) {
	uri := "user.v1.ChatService/ListChatRooms"
	req := &model.ListChatRoomsRequest{
		MaxPageSize: 32,
	}
	resp, err := GetReplive(uri, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat rooms: %v", err)
	}
	msgResp := new(model.ListChatRoomsResponse)
	if err := proto.Unmarshal(resp, msgResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ListChatRoomsResponse: %v", err)
	}
	return msgResp.ChatRooms, nil
}

func ListMyOshis(maxPageSize int64, pageToken string) (*model.ListMyOshisResponse, error) {
	uri := "user.v1.UserService/ListMyOshis"
	req := &model.ListMyOshisRequest{
		MaxPageSize: maxPageSize,
		PageToken:   pageToken,
	}
	resp, err := GetReplive(uri, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list my oshis: %v", err)
	}
	msgResp := new(model.ListMyOshisResponse)
	if err := proto.Unmarshal(resp, msgResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ListMyOshisResponse: %v", err)
	}
	return msgResp, nil
}

func ListFollowings(maxPageSize int64, pageToken string) (*model.ListFollowingsResponse, error) {
	uri := "user.v1.UserService/ListFollowings"
	req := &model.ListFollowingsRequest{
		UserId:      "me",
		MaxPageSize: maxPageSize,
		PageToken:   pageToken,
		Type:        model.FollowTargetType_FOLLOW_TARGET_TYPE_OSHI,
	}
	resp, err := GetReplive(uri, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list followings: %v", err)
	}
	msgResp := new(model.ListFollowingsResponse)
	if err := proto.Unmarshal(resp, msgResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ListFollowingsResponse: %v", err)
	}
	return msgResp, nil
}

func GetUserPrivate() (*model.UserPrivate, error) {
	uri := "user.v1.UserService/GetUserPrivate"
	req := &model.GetUserPrivateRequest{}
	resp, err := GetReplive(uri, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user private: %v", err)
	}
	msgResp := new(model.GetUserPrivateResponse)
	if err := proto.Unmarshal(resp, msgResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GetUserPrivateResponse: %v", err)
	}
	return msgResp.GetUser(), nil
}

func GetStreamingLive() (*model.CheckStreamLiveResponse, error) {
	uri := "user.v1.LiveService/CheckStreamingLive"
	req := &model.CheckStreamLiveRequest{}
	resp, err := GetReplive(uri, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat streaming live: %v", err)
	}
	// rtmp://lvplay.rep_api.com/rep_api/4e20d62f-47da-4dca-8364-6e2cd3574f28?txSecret=e415ac573fd7d4e274d575584c0b52a842f6a09e44a9ccf2128eb1f97db29ffd&txTime=6A7735BD
	msgResp := new(model.CheckStreamLiveResponse)
	if err := proto.Unmarshal(resp, msgResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CheckStreamLiveResponse: %v", err)
	}
	return msgResp, nil
}

func ToDebugString(msg interface{}) string {
	buf, _ := json.Marshal(msg)
	return string(buf)
}
