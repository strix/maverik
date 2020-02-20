package maverik

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"text/tabwriter"
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

type Transactions struct {
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Description string `json:"description"`
	Type        string `json:"tranType"`
	Date        string `json:"tranDate"`
	Points      int    `json:"points"`
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
	req.Header.Add("origin", "https://loyalty.maverik.com")
	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36")
	req.Header.Add("sec-fetch-site", "same-site")
	req.Header.Add("sec-fetch-mode", "cors")

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

func GetTransactions() []Transaction {
	if (UserInfo{}) == currentUserInfo {
		UserInformation()
	}

	url := fmt.Sprintf("%s%s", baseUrl, "/ac-acct/trans")

	currentTime := time.Now()
	formattedEndDate := currentTime.Format("2006-01-02")
	// TODO: make this customizable with flags
	// Transactions from 60 days ago
	startDate := currentTime.AddDate(0, 0, -60)
	formattedStartDate := startDate.Format("2006-01-02")
	req, _ := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("start", formattedStartDate)
	q.Add("end", formattedEndDate)
	req.URL.RawQuery = q.Encode()

	resData, _ := sendRequest(req)

	transactions := Transactions{}

	err := json.Unmarshal(resData, &transactions)
	if err != nil {
		panic(err)
	}
	return transactions.Transactions
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

func PrintTransactions() {
	if (UserInfo{}) == currentUserInfo {
		UserInformation()
	}

	transactions := GetTransactions()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	templates := template.Must(template.ParseGlob(path.Dir(filename) + "/../../templates/*"))

	w := tabwriter.NewWriter(os.Stdout, 4, 4, 4, ' ', 0)
	err := templates.ExecuteTemplate(w, "transactions.tmpl", transactions)
	w.Flush()

	if err != nil {
		panic(err)
	}
}
