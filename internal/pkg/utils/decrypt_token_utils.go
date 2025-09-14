package utils

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
)

type DecryptTokenData struct {
	Token  string   `json:"token"`
	Orders []string `json:"orders"`
}

type DecryptTokenResponse struct {
	Success bool             `json:"success"`
	Msg     string           `json:"msg"`
	Data    DecryptTokenData `json:"data"`
}

// GetDecryptToken 获取解密token
func GetDecryptToken() (string, error) {
	// 构造请求 URL
	url := fmt.Sprintf("http://101.227.171.238:5000/api/login?username=%s&password=%s", "zkzp", "askewrp23k2j")

	// 发起 GET 请求
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result DecryptTokenResponse
	if err := sonic.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if !result.Success {
		return "", nil
	}
	return result.Data.Token, nil
}
