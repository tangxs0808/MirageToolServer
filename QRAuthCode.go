package main

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type fetchQRRes struct {
	Status string `json:"status"`
	Code   string `json:"code"`
}

// 由蜃境服务器拉取小程序码或者授权状态
func (tool *MirageTool) fetchQR(
	w http.ResponseWriter,
	r *http.Request,
) {
	reqData := make(map[string]string)
	json.NewDecoder(r.Body).Decode(&reqData)
	if user, ok := tool.QRCache.Get(reqData["state"]); ok {
		if user != nil { // 该state码已被扫码授权过
			authCode := tool.GenAuthCode()
			tool.AuthCodeCache.Set(authCode, user, 2*time.Minute)
			res := fetchQRRes{
				Status: "OK",
				Code:   authCode,
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(&res)
			if err != nil {
				log.Error().
					Caller().
					Err(err).
					Msg("Failed to write response")
			}
			return
		}
		// 该state码已拉取过二维码但未被扫码授权过
		res := fetchQRRes{
			Status: "Wait",
			Code:   "",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(&res)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}

	// 该state码还未获取过二维码
	miniQR := tool.genMiniProgramQR(reqData["state"])
	if miniQR != "" {
		// 需注意，cache里以state为索引记录的是user信息，而不进行二维码缓存
		tool.QRCache.Set(reqData["state"], nil, 2*time.Minute)

		res := fetchQRRes{
			Status: "New",
			Code:   miniQR,
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(&res)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	res := fetchQRRes{
		Status: "Error",
		Code:   "",
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(&res)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
	return
}

type authQRRes struct {
	Status string
}

// 小程序上用户完成授权确认
func (tool *MirageTool) authQR(
	w http.ResponseWriter,
	r *http.Request,
) {
	reqData := make(map[string]string)
	json.NewDecoder(r.Body).Decode(&reqData)
	openID := tool.exchangeCodeToIDs(reqData["logincode"])
	user := tool.GetUserByID(openID)
	if user == nil {
		res := authQRRes{
			Status: "NoUser",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(&res)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	stateCode := reqData["state"]
	userInterface, expiration, ok := tool.QRCache.GetWithExpiration(stateCode)
	if !ok || userInterface != nil {
		res := authQRRes{
			Status: "StateExpire",
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(&res)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	tool.QRCache.Set(stateCode, *user, expiration.Sub(time.Now()))
	res := authQRRes{
		Status: "OK",
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(&res)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
	return
}

func (t *MirageTool) GenAuthCode() string {
	const letterBytes = "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 64)

	index := make([]byte, 64)
	rand.Read(index)
	for i := 0; i < 64; i++ {
		b[i] = letterBytes[index[i]&63]
	}
	return string(b)
}
