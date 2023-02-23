package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func (t *MirageTool) genMiniProgramQR(stateCode string) string {
	accessToken := t.fetchAccessToken()

	url := fmt.Sprintf("%s/wxa/getwxacodeunlimit?access_token=%s", t.cfg.WX.URL, accessToken)

	message := map[string]interface{}{
		"page":        "pages/scanAuth/scanAuth",
		"scene":       stateCode,
		"env_version": "trial",
		"check_path":  false,
	}

	// 将 message 转换为 JSON 格式
	requestBody, err := json.Marshal(message)
	if err != nil {
		log.Error().Caller().Msgf("创建微信小程序码拉取请求结构体出错")
	}

	// 创建一个新的请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Error().Caller().Msgf("创建微信小程序码拉取请求出错")
	}
	// 设置请求的 Content-Type
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Caller().Msgf("发送微信小程序码拉取请求出错")
	}
	defer resp.Body.Close()
	/*
		// 读取响应体中的数据
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			return ""
		}
	*/
	// 根据响应的 MIME 类型处理响应
	switch resp.Header.Get("Content-Type") {
	case "application/json; charset=UTF-8":
		// 响应是 JSON
		var respData map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&respData)
		// 处理 JSON 数据
		fmt.Printf("Received JSON response: %v\n", respData)
		return ""
	case "image/jpeg", "image/png":
		// 响应是图像二进制数据
		/*
			img, _, err := image.Decode(resp.Body)
			if err != nil {
				fmt.Printf("Error decoding image: %v\n", err)
				return ""
			}
			// 处理图像数据
			fmt.Printf("Received image with size %dx%d\n", img.Bounds().Dx(), img.Bounds().Dy())
		*/
		jpegData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error receiving image: %v\n", err)
			return ""
		}
		imgBase64Str := base64.StdEncoding.EncodeToString(jpegData)
		return imgBase64Str
	}
	fmt.Println("Unsupported response type")
	return ""
}

func (t *MirageTool) fetchAccessToken() string {
	accessToken, ok := t.accessToken.Get("WXAccessToken")
	if !ok {
		url := fmt.Sprintf(
			"%s/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
			t.cfg.WX.URL,
			t.cfg.WX.AppId,
			t.cfg.WX.AppSecret,
		)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Error().Caller().Msgf("创建微信小程序授权码请求出错")
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Error().Caller().Msgf("发送微信小程序授权码请求出错")
		}
		defer resp.Body.Close()

		resData := make(map[string]string)
		json.NewDecoder(resp.Body).Decode(&resData)
		t.accessToken.Set("WXAccessToken", resData["access_token"], 100*time.Minute)
		accessToken = resData["access_token"]
	}
	return accessToken.(string)
}

func (t *MirageTool) exchangeCodeToIDs(code string) string {
	url := fmt.Sprintf("%s/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		t.cfg.WX.URL, t.cfg.WX.AppId, t.cfg.WX.AppSecret, code)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error().Caller().Msgf("创建微信小程序Code交换ID请求出错")
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Caller().Msgf("发送微信小程序Code交换ID请求出错")
	}

	defer resp.Body.Close()

	resData := make(map[string]string)
	json.NewDecoder(resp.Body).Decode(&resData)
	fmt.Println(resData)

	return resData["openid"]
}
