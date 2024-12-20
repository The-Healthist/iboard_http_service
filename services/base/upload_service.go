package base_services

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"gorm.io/gorm"
)

type IUploadService interface {
	GetUploadParams(uploadDir string, callbackUrl string) (map[string]interface{}, error)
	UploadCallback() error
}

type UploadService struct {
	db         *gorm.DB
	expireTime int64
}

type ConfigStruct struct {
	Expiration string     `json:"expiration"`
	Conditions [][]string `json:"conditions"`
}

type CallbackParam struct {
	CallbackUrl      string `json:"callbackUrl"`
	CallbackBody     string `json:"callbackBody"`
	CallbackBodyType string `json:"callbackBodyType"`
}

type PolicyToken struct {
	AccessKeyId string `json:"accessid"`
	Host        string `json:"host"`
	Expire      int64  `json:"expire"`
	Signature   string `json:"signature"`
	Policy      string `json:"policy"`
	Directory   string `json:"dir"`
	Callback    string `json:"callback"`
}

func NewUploadService(db *gorm.DB) IUploadService {
	return &UploadService{
		db:         db,
		expireTime: 30, // 30 seconds
	}
}

func (s *UploadService) GetUploadParams(uploadDir string, callbackUrl string) (map[string]interface{}, error) {
	now := time.Now().Unix()
	expireEnd := now + s.expireTime
	tokenExpire := time.Unix(expireEnd, 0).UTC().Format("2006-01-02T15:04:05Z")

	// Create policy
	configStruct := ConfigStruct{
		Expiration: tokenExpire,
		Conditions: [][]string{
			{"starts-with", "$key", uploadDir},
		},
	}

	// Generate policy
	policyJson, err := json.Marshal(configStruct)
	if err != nil {
		return nil, fmt.Errorf("policy JSON serialization failed: %v", err)
	}
	policyBase64 := base64.StdEncoding.EncodeToString(policyJson)

	// Generate signature
	h := hmac.New(sha1.New, []byte(os.Getenv("ACCESS_KEY_SECRET")))
	io.WriteString(h, policyBase64)
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Create callback
	callbackParam := CallbackParam{
		CallbackUrl:      callbackUrl,
		CallbackBody:     "filename=${object}&size=${size}&mimeType=${mimeType}&height=${imageInfo.height}&width=${imageInfo.width}",
		CallbackBodyType: "application/x-www-form-urlencoded",
	}
	callbackJson, err := json.Marshal(callbackParam)
	if err != nil {
		return nil, fmt.Errorf("callback JSON serialization failed: %v", err)
	}
	callbackBase64 := base64.StdEncoding.EncodeToString(callbackJson)

	// Create policy token
	policyToken := PolicyToken{
		AccessKeyId: os.Getenv("ACCESS_KEY_ID"),
		Host:        os.Getenv("HOST"),
		Expire:      expireEnd,
		Signature:   signature,
		Directory:   uploadDir,
		Policy:      policyBase64,
		Callback:    callbackBase64,
	}

	// Convert to map
	tokenJson, err := json.Marshal(policyToken)
	if err != nil {
		return nil, fmt.Errorf("policy token JSON serialization failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(tokenJson, &result); err != nil {
		return nil, fmt.Errorf("failed to convert policy token to map: %v", err)
	}

	return result, nil
}

func (s *UploadService) UploadCallback() error {
	return nil
}
