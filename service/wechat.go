package service

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CorpWxXmlReceiveMsg struct {
	ToUserName   CDATA `xml:"ToUserName"`
	FromUserName CDATA `xml:"FromUserName"`
	CreateTime   int64 `xml:"CreateTime"`
	MsgType      CDATA `xml:"MsgType"`
	Content      CDATA `xml:"Content"`
	MsgId        int64 `xml:"MsgId"`
	AgentID      int64 `xml:"AgentID"`
}

type AccessToken struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type MsgRet struct {
	Errcode    int    `json:"errcode"`
	Errmsg     string `json:"errmsg"`
	NextCursor string `json:"next_cursor"`
	MsgList    []Msg  `json:"msg_list"`
}
type Msg struct {
	Msgid    string `json:"msgid"`
	SendTime int64  `json:"send_time"`
	Origin   int    `json:"origin"`
	Msgtype  string `json:"msgtype"`
	Event    struct {
		EventType      string `json:"event_type"`
		Scene          string `json:"scene"`
		OpenKfid       string `json:"open_kfid"`
		ExternalUserid string `json:"external_userid"`
		WelcomeCode    string `json:"welcome_code"`
	} `json:"event"`
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
	OpenKfid       string `json:"open_kfid"`
	ExternalUserid string `json:"external_userid"`
}

type ReplyMsg struct {
	Touser   string `json:"touser,omitempty"`
	OpenKfid string `json:"open_kfid,omitempty"`
	Msgid    string `json:"msgid,omitempty"`
	Msgtype  string `json:"msgtype,omitempty"`
	Text     struct {
		Content string `json:"content,omitempty"`
	} `json:"text,omitempty"`
}

func TalkWeiXin(c *gin.Context) {
	receiverId := corpId
	verifyMsgSign := c.Query("msg_signature")
	verifyTimestamp := c.Query("timestamp")
	verifyNonce := c.Query("nonce")
	bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
	crypt := NewWXBizMsgCrypt(token, encodingAesKey, receiverId, XmlType)
	data, _ := crypt.DecryptMsg(verifyMsgSign, verifyTimestamp, verifyNonce, bodyBytes)
	var receiveMsg CorpWxXmlReceiveMsg
	err := xml.Unmarshal(data, &receiveMsg)
	if err != nil {
		fmt.Println("err:  " + err.Error())
	}
	fmt.Println("receiveMsg.Content: ", receiveMsg.Content.Value)
	if receiveMsg.MsgType.Value == "text" {

		go orderMeeting(receiveMsg.FromUserName.Value, receiveMsg.Content.Value)
	}
	c.JSON(200, "success")
}

type UserMeeting struct {
	Num       int      `json:"number"`
	Duration  int64    `json:"duration"`
	StartTime string   `json:"start_time"`
	Attendees []string `json:"people"`
}

func orderMeeting(userid, message string) {
	// 1. 发送给gpt要求获取到格式化的消息
	complete, err := OpenAiComplete(message)
	if err != nil {
		fmt.Println("GPT err: ", err.Error())
	}
	fmt.Println("complete: ", complete)
	//complete := `{"num": 2, "duration": 2, "start_time": "2023-05-12 15:00:00", "people": ["sunmingjian", "tanchanghao"]}`
	// 用户对会议室的要求
	meeting := UserMeeting{}
	if err := json.Unmarshal([]byte(complete), &meeting); err != nil {
		fmt.Println("Error:", err)
		return
	}
	// 2. 查询当前的会议室列表
	meetingRoomList, err := ListMeetingRoom()
	if err != nil {
		fmt.Println("ListMeetingRoom err: ", err.Error())
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", meeting.StartTime, loc)
	attendEmails := make([]string, 0)
	for _, attendee := range meeting.Attendees {
		attendEmails = append(attendEmails, attendee+"@pidan999.onexmail.com")
	}
	marshal, _ := json.Marshal(meetingRoomList)
	fmt.Println("meetingRoomList: ", string(marshal))
	personNumOkList := make([]int, 0)
	for _, r := range meetingRoomList {
		if r.Capacity >= meeting.Num {
			personNumOkList = append(personNumOkList, r.MeetingRoomId)
		}
	}
	// 在指定时间内没有被预定过的会议室ID
	timeOkNum := getUnbookedMeetingRoom(t.Unix(), t.Add(time.Duration(meeting.Duration)*time.Hour).Unix())
	fmt.Println("timeOkNum: ", timeOkNum)
	doubleOk := intersection(personNumOkList, timeOkNum)
	fmt.Println("personNumOkList: ", personNumOkList)
	fmt.Println("doubleOk: ", doubleOk)
	success := false
	for _, meetingRoomId := range doubleOk {
		if err = tryOrderMeetingRoom(meetingRoomId, t.Unix(), t.Add(time.Duration(meeting.Duration)*time.Hour).Unix(), attendEmails, userid); err == nil {
			success = true
			fmt.Println("预定成功")
			break
		}
		continue
	}
	if !success {
		fmt.Println("预定失败")
	}
	// todo 发送通知给用户
}

func accessToken() string {
	var tokenCacheKey = "tokenCache"
	data, found := tokenCache.Get(tokenCacheKey)
	if found {
		return fmt.Sprintf("%v", data)
	}
	urlBase := "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s"
	url := fmt.Sprintf(urlBase, corpId, corpSecret)
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	s := string(body)
	var accessToken AccessToken
	json.Unmarshal([]byte(s), &accessToken)
	t := accessToken.AccessToken
	tokenCache.Set(tokenCacheKey, t, 5*time.Minute)
	return t
}

// CheckWeiXinSign https://developer.work.weixin.qq.com/document/path/90238
func CheckWeiXinSign(c *gin.Context) {
	wxCrypt := NewWXBizMsgCrypt(token, encodingAesKey, corpId, 1)
	verifyMsgSign := c.Query("msg_signature")
	verifyTimestamp := c.Query("timestamp")
	verifyNonce := c.Query("nonce")
	verifyEchoStr := c.Query("echostr")
	echoStr, cryptErr := wxCrypt.VerifyURL(verifyMsgSign, verifyTimestamp, verifyNonce, verifyEchoStr)
	if nil != cryptErr {
		panic(111)
	}
	c.Data(200, "text/plain;charset=utf-8", echoStr)
}

// 计算两个数组的交集
func intersection(arr1, arr2 []int) []int {
	// 使用map记录第一个数组中出现的元素
	map1 := make(map[int]bool)
	for _, v := range arr1 {
		map1[v] = true
	}
	// 遍历第二个数组，如果元素在map中出现过，则为交集
	var res []int
	for _, v := range arr2 {
		if map1[v] {
			res = append(res, v)
		}
	}
	return res
}
