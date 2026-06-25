package live

import (
	"fmt"
	"math/rand"
	"net/url"
	"nsy_chat_live/api"
	"nsy_chat_live/model"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

func StartLiveWatcher() {
	ticker := time.NewTicker(2 * time.Second)
	liveRecordChannel = make(chan *NsyLiveInfo, 10)
	go func() {
		fmt.Println("live_watcher start")
		for {
			select {
			case <-ticker.C:
				// 检查直播状态
				if err := checkLive(); err != nil {
					fmt.Println("live_watcher check error:", err)
				}
				time.Sleep(time.Duration(rand.Intn(1900)+100) * time.Millisecond)
			}
		}
	}()
	startFfmpegWatcher()
}

type NsyLiveInfo struct {
	*model.LiveStream
	Name    string
	RtmpUrl string
}

var (
	knownLives        sync.Map
	liveRecordChannel chan *NsyLiveInfo
)

func checkLive() error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("checkLive panic err: ", err)
		}
	}()
	uri := "user.v1.LiveService/CheckStreamingLive"
	req := &model.CheckStreamLiveRequest{}
	resp, err := api.GetReplive(uri, req)
	if err != nil {
		return fmt.Errorf("failed to get chat streaming live: %v", err)
	}
	msgResp := new(model.CheckStreamLiveResponse)
	if err := proto.Unmarshal(resp, msgResp); err != nil {
		return fmt.Errorf("failed to unmarshal CheckStreamLiveResponse: %v", err)
	}
	for i, live := range msgResp.LiveInfo {
		if _, exist := knownLives.Load(live.LiveId); exist {
			continue
		}
		nsyInfo := msgResp.UserProfile[i]
		rtmpUrl, err := parseRtmpUrl(live.WebrtcUrl)
		if err != nil {
			return fmt.Errorf("failed to parse url: %v", err)
		}
		nsyLiveInfo := &NsyLiveInfo{
			LiveStream: live,
			Name:       nsyInfo.Info.DisplayName,
			RtmpUrl:    rtmpUrl,
		}
		fmt.Println(nsyInfo.Info.DisplayName + " start live, title: " + live.Title)
		fmt.Println("rtmp url: " + rtmpUrl)
		liveRecordChannel <- nsyLiveInfo
		knownLives.Store(nsyLiveInfo.LiveId, nsyLiveInfo)
	}
	return nil
}

func parseRtmpUrl(webrtcUrl string) (string, error) {
	u, err := url.Parse(webrtcUrl)
	if err != nil {
		return "", err
	}
	u.Scheme = "rtmp"
	queryParams := u.Query()
	keepParams := map[string]bool{
		"txSecret": true,
		"txTime":   true,
	}
	newQueryParams := url.Values{}
	for key, values := range queryParams {
		if keepParams[key] {
			for _, value := range values {
				newQueryParams.Add(key, value)
			}
		}
	}
	u.RawQuery = newQueryParams.Encode()
	return u.String(), nil
}
