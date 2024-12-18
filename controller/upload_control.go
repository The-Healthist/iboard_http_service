package http_controller

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	databases "github.com/The-Healthist/iboard_http_service/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Set the expiration time for the policy token (seconds)
var expire_time int64 = 30

// FileService struct contains a pointer to gorm.DB
type FileService struct {
	db *gorm.DB
}

// NewFileService creates a new instance of FileService
func NewFileService(db *gorm.DB) *FileService {
	return &FileService{
		db: db,
	}
}

// ConfigStruct is used to generate the upload policy
type ConfigStruct struct {
	Expiration string     `json:"expiration"` // Expiration time of the policy
	Conditions [][]string `json:"conditions"` // Upload conditions
}

// CallbackParam struct for callback parameters after upload completion
type CallbackParam struct {
	CallbackUrl      string `json:"callbackUrl"`      // Callback URL
	CallbackBody     string `json:"callbackBody"`     // Content of the callback request body
	CallbackBodyType string `json:"callbackBodyType"` // Type of the callback request body
}

// PolicyToken struct returned to the frontend
type PolicyToken struct {
	AccessKeyId string `json:"accessid"`  // Access Key ID
	Host        string `json:"host"`      // Host address
	Expire      int64  `json:"expire"`    // Policy expiration time
	Signature   string `json:"signature"` // Signature
	Policy      string `json:"policy"`    // Policy content
	Directory   string `json:"dir"`       // Upload directory
	Callback    string `json:"callback"`  // Callback parameters
}

// getGMTISO8501 converts a Unix timestamp to a GMT ISO 8501 formatted string
func getGMTISO8501(expire_end int64) string {
	var tokenExpire = time.Unix(expire_end, 0).UTC().Format("2006-01-02T15:04:05Z")
	return tokenExpire
}

// GetPolicyToken generates an upload policy token
func GetPolicyToken(upload_dir string, callbackUrl string) (*map[string]interface{}, error) {
	now := time.Now().Unix()
	// Calculate the expiration time of the policy
	expire_end := now + expire_time
	var tokenExpire = getGMTISO8501(expire_end)

	// Create the JSON structure for the upload policy
	var configStruct ConfigStruct
	configStruct.Expiration = tokenExpire
	var condition []string
	condition = append(condition, "starts-with")
	condition = append(condition, "$key")
	condition = append(condition, upload_dir)
	configStruct.Conditions = append(configStruct.Conditions, condition)

	// Calculate the signature
	result, err := json.Marshal(configStruct)
	if err != nil {
		return nil, fmt.Errorf("policy JSON serialization failed: %v", err)
	}
	debyte := base64.StdEncoding.EncodeToString(result)
	// Create HMAC-SHA1 hash
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(os.Getenv("ACCESS_KEY_SECRET")))
	_, err = io.WriteString(h, debyte)
	if err != nil {
		return nil, fmt.Errorf("failed to write to HMAC hash: %v", err)
	}
	signedStr := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Set callback parameters
	var callbackParam CallbackParam
	callbackParam.CallbackUrl = callbackUrl
	callbackParam.CallbackBody = "filename=${object}&size=${size}&mimeType=${mimeType}&height=${imageInfo.height}&width=${imageInfo.width}"
	callbackParam.CallbackBodyType = "application/x-www-form-urlencoded"
	callback_str, err := json.Marshal(callbackParam)
	if err != nil {
		log.Println("Callback parameters JSON serialization error:", err)
	}
	// Base64 encode the callback parameters
	callbackBase64 := base64.StdEncoding.EncodeToString(callback_str)

	// Construct the policy token
	var policyToken PolicyToken
	policyToken.AccessKeyId = os.Getenv("ACCESS_KEY_ID")
	policyToken.Host = os.Getenv("HOST")
	policyToken.Expire = expire_end
	policyToken.Signature = string(signedStr)
	policyToken.Directory = upload_dir
	policyToken.Policy = string(debyte)
	policyToken.Callback = string(callbackBase64)

	// Add log output
	log.Printf("PolicyToken: %+v", policyToken)

	// Serialize the policy token to JSON
	response, err := json.Marshal(policyToken)
	if err != nil {
		log.Println("Policy token JSON serialization error:", err)
	}

	// Deserialize the JSON into a map
	var data map[string]interface{}
	err = json.Unmarshal(response, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetUploadParams retrieves upload parameters, including the policy token
func (s *FileService) GetUploadParams(uploadDir string, callbackUrl string) (*map[string]interface{}, error) {
	policy, err := GetPolicyToken(uploadDir, callbackUrl)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// UploadCallback handles the callback after upload; specific logic not implemented here
func (s *FileService) UploadCallback() error {
	return nil
}

// GetUploadParams handles the HTTP request for upload parameters (supports JSON and form formats)
func GetUploadParams(c *gin.Context) {
	var req struct {
		UploadDir   string `json:"upload_dir" binding:"required"`
		CallbackURL string `json:"callback_url" binding:"required"`
	}

	// Try to parse the JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		// If JSON parsing fails, try to parse form data
		req.UploadDir = c.PostForm("upload_dir")
		req.CallbackURL = c.PostForm("callback_url")
		if req.UploadDir == "" || req.CallbackURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Request parameter format is incorrect or missing required fields",
			})
			log.Printf("Failed to parse request parameters: %v", err)
			return
		}
	}

	log.Printf("Received upload_dir: %s, callback_url: %s", req.UploadDir, req.CallbackURL)

	// Create a FileService instance
	fileService := NewFileService(databases.DB_CONN)

	// Call FileService's GetUploadParams method
	policy, err := fileService.GetUploadParams(req.UploadDir, req.CallbackURL)
	if err != nil {
		// If there's an error, return the error message with a 500 status code
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		log.Printf("Failed to get policy token: %v", err)
		return
	}

	// On success, return the policy token with a 200 status code
	c.JSON(http.StatusOK, policy)
	log.Printf("Successfully returned policy token: %+v", policy)
}
