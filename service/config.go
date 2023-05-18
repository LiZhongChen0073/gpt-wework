package service

import (
	"github.com/patrickmn/go-cache"
	"os"
	"time"
)

var token, encodingAesKey, corpId, corpSecret, openAiKey string
var tokenCache *cache.Cache

func InitConfig() {

	token = os.Getenv("WEWORK_TOKEN")

	encodingAesKey = os.Getenv("WEWORK_ENCODING_AEK_KEY")

	// 企业微信企业id
	corpId = os.Getenv("WEWORK_CORP_ID")

	// 企业微信secret
	corpSecret = os.Getenv("WEWORK_CROP_SECRET")

	// openai key
	openAiKey = os.Getenv("OPENAI_KEY")

	// 企业微信 token 缓存，请求频次过高可能有一些额外的问题
	tokenCache = cache.New(5*time.Minute, 5*time.Minute)
}
