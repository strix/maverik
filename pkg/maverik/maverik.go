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

var maverikAuthToken = "23630dddba52715dea5ec378394666de"
var baseUrl = "https://maverik.com/.api/v1"
var cookieJar, _ = cookiejar.New(nil)

const (
	Bonfire = 180
	Drinks  = 181
	Energy  = 182
)

var punchCards = map[string]int{
	"bonfire": 180,
	"drinks":  181,
	"energy":  182,
}

var PunchCardNames = map[int]string{
	Bonfire: "Bonfire Items",
	Drinks:  "Fountain Drinks",
	Energy:  "Energy Drinks",
}

var currentUserInfo = UserInfo{}

type Points struct {
	Earned    float64 `json:"earned"`
	Spent     float64 `json:"spent"`
	Available float64 `json:"available"`
}

type UserInfo struct {
	User struct {
		UserId          int `json:"user_seq_number"`
		AvailablePoints int `json:"available_points_summary"`
		Points          `json:"userPointsSummaryView"`
	} `json:"user"`
}

type PunchCard struct {
	Name         string
	PunchSummary struct {
		Total   int  `json:"total"`
		Reward  bool `json:"reward"`
		PunchId int  `json:"punch_id"`
	} `json:"PunchSummary"`
}

type Reward struct {
	PunchId  int    `json:"punchId"`
	Name     string `json:"promoName"`
	Redeemed bool   `json:"redeemed"`
	Expired  bool   `json:"expired"`
	Quantity int32  `json:"quantity"`
	Issued   int64  `json:"rewardDate"`
}

func (r Reward) DateIssued() time.Time {
	return time.Unix(0, r.Issued*int64(time.Millisecond))
}

func (r Reward) ExpirationDate() time.Time {
	// adding 14 days for punch card expiration and 1 day for js date correction.
	expMillis := r.Issued + 1296000000
	return time.Unix(0, expMillis*int64(time.Millisecond))
}

func (r Reward) DaysToExpire() int32 {
	// maybe floor instead of convert to int32?
	return int32(r.ExpirationDate().Sub(r.DateIssued()).Hours() / 24)
}

func sendRequest(req *http.Request) ([]byte, error) {
	req.Header.Add("authorization", maverikAuthToken)

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
	url := fmt.Sprintf("%s%s", baseUrl, "/login")

	payload := strings.NewReader(fmt.Sprintf("{\"username\":\"%s\",\"password\":\"%s\"}", username, password))

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", maverikAuthToken)

	_, err := sendRequest(req)
	if err != nil {
		panic(err)
	}
}

func UserInformation() UserInfo {
	if (UserInfo{}) != currentUserInfo {
		return currentUserInfo
	}
	url := fmt.Sprintf("%s%s", baseUrl, "/userinformation")

	req, _ := http.NewRequest("GET", url, nil)

	resData, _ := sendRequest(req)

	err := json.Unmarshal(resData, &currentUserInfo)
	if err != nil {
		panic(err)
	}
	return currentUserInfo
}

func GetPunchCard(punchCardType int) PunchCard {
	// TODO: validate punchCardType with punchCards map keys
	if (UserInfo{}) == currentUserInfo {
		UserInformation()
	}
	path := fmt.Sprintf("%s/%d/%d", "/punch/user-id", currentUserInfo.User.UserId, punchCardType)
	url := fmt.Sprintf("%s%s", baseUrl, path)

	req, _ := http.NewRequest("GET", url, nil)

	resData, _ := sendRequest(req)

	punchCard := PunchCard{}
	err := json.Unmarshal(resData, &punchCard)
	punchCard.Name = PunchCardNames[punchCardType]
	if err != nil {
		panic(err)
	}

	return punchCard
}

func GetPunchCards(punchCardTypes []int) []PunchCard {
	punchCards := []PunchCard{}
	for _, punchCardName := range punchCardTypes {
		// TODO: run in parallel
		punchCard := GetPunchCard(punchCardName)
		punchCards = append(punchCards, punchCard)
	}
	return punchCards
}

func GetPointInfo() Points {
	if (UserInfo{}) == currentUserInfo {
		UserInformation()
	}
	return currentUserInfo.User.Points
}

func GetRewardInfo() []Reward {
	if (UserInfo{}) == currentUserInfo {
		UserInformation()
	}
	path := fmt.Sprintf("%s/%d", "/virtualbuy/userrewards", currentUserInfo.User.UserId)
	url := fmt.Sprintf("%s%s", baseUrl, path)

	req, _ := http.NewRequest("GET", url, nil)

	resData, _ := sendRequest(req)

	rewards := []Reward{}
	err := json.Unmarshal(resData, &rewards)
	if err != nil {
		panic(err)
	}

	return rewards
}

func PrintSummary() { // TODO: pass maverik config
	if (UserInfo{}) == currentUserInfo {
		UserInformation()
	}
	cards := []int{Drinks, Bonfire, Energy}
	cardResults := GetPunchCards(cards)

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	templates := template.Must(template.ParseGlob(path.Dir(filename) + "/../../templates/*"))

	err := templates.ExecuteTemplate(os.Stdout, "user-summary.tmpl", currentUserInfo)

	if err != nil {
		panic(err)
	}

	err = templates.ExecuteTemplate(os.Stdout, "card-summary.tmpl", cardResults)
	if err != nil {
		panic(err)
	}
}
