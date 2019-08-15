package maverik

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"text/template"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

var apiAuth = Auth{}
var baseUrl = "https://gateway.maverik.com"
var cookieJar, _ = cookiejar.New(nil)

var currentUserInfo = UserInfo{}

type Auth struct {
	Token string `json:"access_token"`
}

type UserInfo struct {
	AccountId  int    `json:"accountId"`
	CardNumber string `json:"primaryCardNumber"`
}

type AccountInfo struct {
	Fields struct {
		EnrollDate string `json:"enrollDate"`
	} `json:"fields"`
	Redeemables []Item `json:"myStuff"`
	PunchCards  []Item `json:"punchIts"`
	Points      Item   `json:"trailPoints"`
}

type Item struct {
	Amount      float32      `json:"balance,float64"`
	Name        string       `json:"name"`
	Expirations []Expiration `json:"expirations,omitempty"`
}

type Expiration struct {
	Quantity       int    `json:"amount"`
	ExpirationDate string `json:"expirationDate"`
}

func (exp Expiration) DaysToExpire() int32 {
	dateLayout := "2006-01-02"
	expTime, err := time.Parse(dateLayout, exp.ExpirationDate)
	if err != nil {
		panic(err)
	}
	return int32(expTime.Sub(time.Now()).Hours()/24) + 1
}

func sendRequest(req *http.Request) ([]byte, error) {
	if apiAuth.Token != "" {
		req.Header.Add("AUTH-TOKEN", apiAuth.Token)
	}
	req.Header.Add("APP-ID", "PAYX")
	req.Header.Add("content-type", "application/json")

	client := &http.Client{
		Timeout: time.Second * 10,
		Jar:     cookieJar,
	}

	res, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	return ioutil.ReadAll(res.Body)
}

func Login(username string, password string) {
	url := fmt.Sprintf("%s%s", baseUrl, "/api/oauth/requestToken")

	payload := strings.NewReader(fmt.Sprintf("{\"username\":\"%s\",\"password\":\"%s\"}", username, password))

	req, _ := http.NewRequest("POST", url, payload)

	resData, err := sendRequest(req)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(resData, &apiAuth)
	if err != nil {
		panic(err)
	}
}

func UserInformation() UserInfo {
	if (UserInfo{}) != currentUserInfo {
		return currentUserInfo
	}
	url := fmt.Sprintf("%s%s", baseUrl, "/ac-acct/userInfo")

	req, _ := http.NewRequest("GET", url, nil)

	resData, _ := sendRequest(req)

	err := json.Unmarshal(resData, &currentUserInfo)
	if err != nil {
		panic(err)
	}
	return currentUserInfo
}

func GetAccountInfo() AccountInfo {
	if (UserInfo{}) == currentUserInfo {
		UserInformation()
	}
	path := fmt.Sprintf("%s/%d", "/ac-acct/account/refresh", currentUserInfo.AccountId)
	url := fmt.Sprintf("%s%s", baseUrl, path)
	req, _ := http.NewRequest("GET", url, nil)

	resData, _ := sendRequest(req)
	accountInfo := AccountInfo{}
	err := json.Unmarshal(resData, &accountInfo)
	if err != nil {
		panic(err)
	}
	return accountInfo
}

func PrintSummary() { // TODO: pass maverik config
	if (UserInfo{}) == currentUserInfo {
		UserInformation()
	}

	accountInfo := GetAccountInfo()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	templates := template.Must(template.ParseGlob(path.Dir(filename) + "/../../templates/*"))

	err := templates.ExecuteTemplate(os.Stdout, "account-summary.tmpl", accountInfo)

	if err != nil {
		panic(err)
	}
}
