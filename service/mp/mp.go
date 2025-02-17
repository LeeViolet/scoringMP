package mp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"scoringMP/config"
)

// 定义微信返回的数据结构
type WxSessionData struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func Code2Session(code string) (string, error) {
	// 微信接口地址
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", config.Config.AppId, config.Config.AppSecret, code)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析 JSON 数据
	var wxData WxSessionData
	err = json.Unmarshal(body, &wxData)
	if err != nil {
		return "", errors.New("parse JSON failed")
	}
	if wxData.ErrCode != 0 {
		return "", errors.New(wxData.ErrMsg)
	}
	return wxData.OpenID, nil
}
