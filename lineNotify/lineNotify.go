package lineNotify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
)

type Notify struct {
	AccessToken string
}

type OauthTokenResponseStruct struct {
	Access_token string
}

// 用來建構 Notify 的假建構子
func NewLineNotify() (notify *Notify) {
	fileExists := CheckFileExist(os.Getenv("LINE_NOTIFY_TOKEN_FILE"))
	accessToken := "none"
	if fileExists {
		f, _ := os.Open(os.Getenv("LINE_NOTIFY_TOKEN_FILE"))
		defer f.Close()
		fd, _ := ioutil.ReadAll(f)

		accessToken = string(fd)
	}

	notify = &Notify{AccessToken: accessToken}

	// 這裡會回傳一個型態是 *Notify 建構體的 notify 變數
	return notify
}

func Auth(state string) string {
	clientID := os.Getenv("LINE_NOTIFY_CLIENT_ID")
	// clientSecret := os.Getenv("LINE_NOTIFY_CLIENT_SECRET")
	callbackURL := os.Getenv("LINE_NOTIFY_CALLBACK_URL")

	return "https://notify-bot.line.me/oauth/authorize?response_type=code&scope=notify&response_mode=form_post&client_id=" + clientID + "&redirect_uri=" + callbackURL + "&state=" + state
}

func OauthToken(code string) string {
	// get access_token
	postURL := "https://notify-bot.line.me/oauth/token"
	postParams := url.Values{}
	postParams.Add("grant_type", "authorization_code")
	postParams.Add("code", code)
	postParams.Add("redirect_uri", os.Getenv("LINE_NOTIFY_CALLBACK_URL"))
	postParams.Add("client_id", os.Getenv("LINE_NOTIFY_CLIENT_ID"))
	postParams.Add("client_secret", os.Getenv("LINE_NOTIFY_CLIENT_SECRET"))

	resp, err := http.PostForm(postURL, postParams)
	accessToken := "none"

	if err != nil {
		// handle error
	} else {
		var otResponse OauthTokenResponseStruct
		json.NewDecoder(resp.Body).Decode(&otResponse)
		accessToken = otResponse.Access_token
	}

	defer resp.Body.Close()

	return accessToken
}

type RevokeResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// 撤銷 Access Token
func (notify *Notify) Revoke() {
	postURL := "https://notify-api.line.me/api/revoke"
	postParams := url.Values{}
	requestBody := strings.NewReader(postParams.Encode())

	client := &http.Client{}
	req, err := http.NewRequest("POST", postURL, requestBody)
	if err != nil {
		// handle error
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+notify.AccessToken)

	resp, _ := client.Do(req)

	defer resp.Body.Close()

	var revokeResponse RevokeResponse
	json.NewDecoder(resp.Body).Decode(&revokeResponse)

	fmt.Println(revokeResponse)
}

type NotifyResponse struct {
	Status  int
	Message string
}

// 傳送訊息
func (notify *Notify) Notify(message string) bool {
	notifyStatus := false
	postURL := "https://notify-api.line.me/api/notify"
	postParams := url.Values{
		"message": {message},
	}
	requestBody := strings.NewReader(postParams.Encode())

	client := &http.Client{}
	req, err := http.NewRequest("POST", postURL, requestBody)
	if err != nil {
		// handle error
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+notify.AccessToken)

	resp, _ := client.Do(req)

	defer resp.Body.Close()

	var notifyResponse NotifyResponse
	json.NewDecoder(resp.Body).Decode(&notifyResponse)

	// fmt.Println("notify at: ", notify.AccessToken)
	fmt.Println("notify: ", notifyResponse)
	fmt.Println("notify message: ", message)

	if notifyResponse.Status == 200 {
		notifyStatus = true
	}

	return notifyStatus
}

/**
* 檢查檔案是否存在
 */
func CheckFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
