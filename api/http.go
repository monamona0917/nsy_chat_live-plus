package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"nsy_chat_live/config"
	"nsy_chat_live/model"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	RepLiveHost = "https://api.replive.com/"
)

var (
	client       *http.Client
	accessToken  *model.RefreshAccessTokenResponse
	mutex        sync.Mutex
	refreshToken string
)

func InitHttp() error {
	refreshToken = config.Conf.RefreshToken
	proxyHost := config.Conf.Proxy.Host + ":" + strconv.Itoa(config.Conf.Proxy.Port)
	client = &http.Client{
		Timeout: 10 * time.Second,
		// 设置代理
		Transport: &http.Transport{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http",
				Host:   proxyHost,
			}),
		},
	}
	accessToken = &model.RefreshAccessTokenResponse{
		AccessToken: "",
		AccessTokenExpireTime: &model.Timestamp{
			Seconds: 0,
		},
	}

	if _, err := getToken(); err != nil {
		return fmt.Errorf("get token err: %v", err)
	}
	return nil
}

func getToken() (string, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if accessToken != nil && accessToken.AccessTokenExpireTime != nil && accessToken.AccessTokenExpireTime.GetSeconds()-180 > time.Now().Unix() {
		return accessToken.GetAccessToken(), nil
	}
	req := &model.RefreshAccessTokenRequest{
		RefreshToken: refreshToken,
	}
	resp, err := Post("user.v1.UserService/RefreshAccessToken", req)
	if err != nil {
		return "", fmt.Errorf("get token err: %v", err)
	}
	tokenResp := new(model.RefreshAccessTokenResponse)
	if err := proto.Unmarshal(resp, tokenResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}
	accessToken = tokenResp
	fmt.Printf("refresh access token, expire time: %s\n", time.Unix(accessToken.AccessTokenExpireTime.GetSeconds(), 0).Format("2006-01-02 15:04:05"))
	return accessToken.GetAccessToken(), nil
}

func setHeaders(req *http.Request) error {
	req.Header.Set("Host", RepLiveHost)
	req.Header.Set("Content-Type", "application/proto")
	req.Header.Set("accept-encoding", "gzip")
	req.Header.Set("accept-charset", "UTF-8")
	req.Header.Set("accept", "application/json")
	req.Header.Set("user-agent", "v3.1.1 23116PN5BC Android 12")
	if !strings.Contains(req.URL.Path, "user.v1.UserService/RefreshAccessToken") {
		token, err := getToken()
		if err != nil {
			return fmt.Errorf("failed to get token: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		//req.Header.Set("x-replive-guest-token", token)
	}
	return nil
}

func GetReplive(uri string, params protoreflect.ProtoMessage) ([]byte, error) {
	buf, err := proto.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal proto: %v", err)
	}
	base64Str := base64.StdEncoding.EncodeToString(buf)
	url := fmt.Sprintf(RepLiveHost+uri+"?encoding=proto&base64=1&message=%v", base64Str)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	if err := setHeaders(req); err != nil {
		return nil, fmt.Errorf("failed to set headers: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to do request, status code: %v, message: %v", resp.StatusCode, resp.Status)
	}
	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	respBuf, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	return respBuf, nil
}

func Post(url string, body protoreflect.ProtoMessage) ([]byte, error) {
	buf, err := proto.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %v", err)
	}
	reader := bytes.NewReader(buf)
	req, err := http.NewRequest("POST", RepLiveHost+url, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	err = setHeaders(req)
	if err != nil {
		return nil, fmt.Errorf("failed to set headers: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to send request, status code: %v, message: %v", resp.StatusCode, resp.Status)
	}
	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	respBuf, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	return respBuf, nil
}
