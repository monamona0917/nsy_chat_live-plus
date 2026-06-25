package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"replive/config"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type TranslateResp struct {
	Source         string `json:"source"`
	Target         string `json:"target"`
	TranslatedText string `json:"translated_text"`
}

func HandleTranslate(ctx context.Context, c *app.RequestContext) {
	text := strings.TrimSpace(string(c.Query("text")))
	source := strings.TrimSpace(string(c.Query("source")))
	target := strings.TrimSpace(string(c.Query("target")))
	if source == "" {
		source = "ja"
	}
	if target == "" {
		target = "zh-CN"
	}
	if text == "" {
		c.JSON(consts.StatusOK, BadResp("text 不能为空"))
		return
	}

	translatedText, err := requestGoogleTranslate(ctx, text, source, target)
	if err != nil {
		c.JSON(consts.StatusOK, BadResp(err.Error()))
		return
	}

	c.JSON(consts.StatusOK, &Resp{
		Data: &TranslateResp{
			Source:         source,
			Target:         target,
			TranslatedText: translatedText,
		},
	})
}

func requestGoogleTranslate(ctx context.Context, text string, source string, target string) (string, error) {
	requestURL, err := url.Parse("https://translate.googleapis.com/translate_a/single")
	if err != nil {
		return "", err
	}
	query := requestURL.Query()
	query.Set("client", "gtx")
	query.Set("sl", source)
	query.Set("tl", target)
	query.Set("dt", "t")
	query.Set("q", text)
	requestURL.RawQuery = query.Encode()

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: config.ProxyFromEnvironmentOrConfig,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("翻译请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("翻译接口返回状态码: %d", resp.StatusCode)
	}

	var raw any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", fmt.Errorf("解析翻译响应失败: %w", err)
	}

	translatedText := parseGoogleTranslateText(raw)
	if translatedText == "" {
		return "", fmt.Errorf("翻译响应为空")
	}
	return translatedText, nil
}

func parseGoogleTranslateText(raw any) string {
	root, ok := raw.([]any)
	if !ok || len(root) == 0 {
		return ""
	}
	segments, ok := root[0].([]any)
	if !ok {
		return ""
	}

	var builder strings.Builder
	for _, item := range segments {
		segment, ok := item.([]any)
		if !ok || len(segment) == 0 {
			continue
		}
		text, ok := segment[0].(string)
		if !ok {
			continue
		}
		builder.WriteString(text)
	}
	return strings.TrimSpace(builder.String())
}
