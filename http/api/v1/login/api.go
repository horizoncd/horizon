package login

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/horizon/http/api/v1/response"
	"github.com/satori/go.uuid"
)

var stateCache *sync.Map

func init() {
	stateCache = &sync.Map{}
}

type OpenIDUser struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Sub      string `json:"sub"`
	Nickname string `json:"nickname"`
}

type State struct {
	fromHost    string
	redirectURL string
	user        *OpenIDUser
}

// Login - user login
func Login(c *gin.Context) {
	const openIDLoginURL = "https://oidc.mockserver.org/connect/authorize?" +
		"response_type=code" +
		"&scope=openid%20fullname%20nickname%20email" +
		"&client_id=" +
		"&state={state}"  +
		"&redirect_uri={redirectURL}"

	fromHostParam := c.Query("fromHost")
	redirectURLParam := c.Query("redirectUrl")
	state := &State{
		fromHost:    fromHostParam,
		redirectURL: redirectURLParam,
	}
	stateID := uuid.NewV4().String()
	stateCache.Store(stateID, state)

	// 写客户端cookie， openIDState记录当前用户登录的stateID
	c.SetCookie("openIDState", stateID, 60*60*30, "", "", false, false)

	redirectURL := strings.ReplaceAll(openIDLoginURL, "{state}", stateID)
	redirectURL = strings.ReplaceAll(redirectURL,
		"{redirectURL}", url.QueryEscape(fmt.Sprintf("http://%s/api/v1/login/callback", c.Request.Host)))


	response.SuccessWithData(c, redirectURL)
}

// Callback - user login callback
func Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	urlAddr := "https://oidc.mockserver.org/connect/token"
	redirect := url.QueryEscape(fmt.Sprintf("http://%s/api/v1/login/callback", c.Request.Host))

	resp, err := http.PostForm(urlAddr, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirect},
		"client_id":     {""},
		"client_secret": {""},
	})
	if err != nil {
		log.Printf("err2: %v", err.Error())
		response.AbortWithInternalError(c)
		return
	}
	if resp.StatusCode != 200 {
		log.Printf("err3: not 200")
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("body not 200 : %v", string(body))
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("err"))
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("err4: %v", err.Error())
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("err"))
		return
	}
	log.Printf("body1: %v", string(body))
	tokenStruct := new(struct {
		AccessToken string `json:"access_token"`
	})
	if err := json.Unmarshal(body, &tokenStruct); err != nil {
		log.Printf("err5: %v", err.Error())
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	urlAddr = "https://oidc.mockserver.org/connect/userinfo?access_token=" + tokenStruct.AccessToken
	resp, err = http.Get(urlAddr)
	if err != nil {
		log.Printf("err6: %v", err.Error())
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("body not 200 - 2 : %v", string(body))
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("err"))
		return
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("err7: %v", err.Error())
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	log.Printf("body2: %v", string(body))
	var user *OpenIDUser
	err = json.Unmarshal(body, &user)
	if err != nil {
		log.Printf("err8: %v", err.Error())
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	stateInterface, ok := stateCache.Load(state)
	if !ok {
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("no state cache"))
		return
	}
	stateStruct := stateInterface.(*State)
	stateStruct.user = user
	stateCache.Store("openIDState", stateStruct)

	redirectTo := "http://" + stateStruct.fromHost + stateStruct.redirectURL
	log.Printf("redirectTo: %v", redirectTo)
	c.Redirect(http.StatusMovedPermanently, redirectTo)
}

// Logout - user logout
func Logout(c *gin.Context) {
	openIDState, _ := c.Cookie("openIDState")
	stateCache.Delete(openIDState)
	c.SetCookie("openIDState", "", -1, "", "", false, false)
	response.Success(c)
}

// UserStatus -
func UserStatus(c *gin.Context) {
	openIDState, _ := c.Cookie("openIDState")
	stateInterface, ok := stateCache.Load(openIDState)
	if !ok  {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}
	stateStruct := stateInterface.(*State)
	if stateStruct.user == nil {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}
	u := stateStruct.user
	response.SuccessWithData(c, User{
		HzNumber:    "XXXX",
		ID:          func() int { i, _ := strconv.Atoi(openIDState); return i }(),
		Login:       true,
		MailAddress: u.Email,
		Name:        u.FullName,
		NickName:    u.Nickname,
		SuperAdmin:  false,
	})
}
