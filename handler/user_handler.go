package handler

import (
	"context"
	"replive/config"
	"replive/rep_api"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type UserProfileDTO struct {
	UserId          string `json:"user_id"`
	UniqueId        string `json:"unique_id"`
	DisplayName     string `json:"display_name"`
	AvatarUrl       string `json:"avatar_url"`
	SendChatEnabled bool   `json:"send_chat"`
}

// HandleGetCurrentUser 返回当前登录用户的信息
func HandleGetCurrentUser(ctx context.Context, c *app.RequestContext) {
	user, err := rep_api.GetUserPrivate()
	if err != nil {
		hlog.Errorf("HandleGetCurrentUser: GetUserPrivate failed: %v", err)
		c.JSON(consts.StatusOK, BadResp("获取用户信息失败: "+err.Error()))
		return
	}
	hlog.Infof("HandleGetCurrentUser: user_id=%s display_name=%s", user.GetUserId(), user.GetDisplayName())
	c.JSON(consts.StatusOK, &Resp{Data: &UserProfileDTO{
		UserId:          user.GetUserId(),
		UniqueId:        user.GetUniqueId(),
		DisplayName:     user.GetDisplayName(),
		AvatarUrl:       user.GetProfileImageUrl(),
		SendChatEnabled: config.Conf.SendChatEnabled,
	}})
}
