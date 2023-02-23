package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type MirageTool struct {
	accessToken   *cache.Cache
	QRCache       *cache.Cache
	AuthCodeCache *cache.Cache

	db  *gorm.DB
	cfg *Config
}

type addUserRes struct {
	Status string
}

// 接收小程序用户注册信息
func (tool *MirageTool) addUser(
	w http.ResponseWriter,
	r *http.Request,
) {
	reqData := make(map[string]string)
	json.NewDecoder(r.Body).Decode(&reqData)
	openID := tool.exchangeCodeToIDs(reqData["logincode"])
	user := tool.UpdateOrCreateUser(openID, reqData["nickname"], reqData["avatarbase64"])
	if user != nil {
		res := addUserRes{
			Status: "success",
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
	res := addUserRes{
		Status: "fail",
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

type verifyRes struct {
	Status      string `json:"status"`
	UserName    string `json:"user_name"`
	DisplayName string `json:"display_name"`
}

func (tool *MirageTool) authVerify(
	w http.ResponseWriter,
	r *http.Request,
) {
	reqData := make(map[string]string)
	json.NewDecoder(r.Body).Decode(&reqData)
	if userInterface, ok := tool.AuthCodeCache.Get(reqData["code"]); ok {
		user := userInterface.(User)
		res := verifyRes{
			Status:      "OK",
			UserName:    "wx-" + strings.ToLower(user.ID)[6:14],
			DisplayName: user.Name,
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
		tool.AuthCodeCache.Delete(reqData["code"])
		return
	}
	res := verifyRes{
		Status: "FAIL",
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

func (tool *MirageTool) createRouter() *mux.Router {
	router := mux.NewRouter()

	//注册
	router.HandleFunc("/addUser", tool.addUser).Methods(http.MethodPost)

	//二维码
	router.HandleFunc("/fetchQR", tool.fetchQR).Methods(http.MethodPost)
	router.HandleFunc("/authQR", tool.authQR).Methods(http.MethodPost)

	//验证授权
	router.HandleFunc("/verify", tool.authVerify).Methods(http.MethodPost)

	return router
}

func (tool *MirageTool) Serve() error {
	errorGroup := new(errgroup.Group)

	router := tool.createRouter()
	httpServer := &http.Server{
		Addr:        "0.0.0.0:5566",
		Handler:     router,
		ReadTimeout: 30 * time.Second,
		// Go does not handle timeouts in HTTP very well, and there is
		// no good way to handle streaming timeouts, therefore we need to
		// keep this at unlimited and be careful to clean up connections
		// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/#aboutstreaming
		WriteTimeout: 0,
	}

	var httpListener net.Listener
	httpListener, err := net.Listen("tcp", "0.0.0.0:5566")
	if err != nil {
		return fmt.Errorf("failed to bind to TCP address: %w", err)
	}

	errorGroup.Go(func() error { return httpServer.Serve(httpListener) })

	return errorGroup.Wait()
}

func main() {
	accessToken := cache.New(0, 0)
	qrCache := cache.New(0, 0)
	authCache := cache.New(0, 0)
	cfg, _ := GetConfig()

	tool := MirageTool{
		accessToken:   accessToken,
		QRCache:       qrCache,
		AuthCodeCache: authCache,
		cfg:           cfg,
	}
	err := tool.initDB()
	if err != nil {
		log.Fatal().Caller().Err(err).Msg("Error starting database")
	}

	err = tool.Serve()
	if err != nil {
		log.Fatal().Caller().Err(err).Msg("Error starting server")
	}
}
