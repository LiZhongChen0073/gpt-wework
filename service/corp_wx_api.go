package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/errgo.v2/errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Coordinate struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

type MeetingRoom struct {
	MeetingRoomId int        `json:"meetingroom_id"`
	Name          string     `json:"name"`
	Capacity      int        `json:"capacity"`
	City          string     `json:"city"`
	Building      string     `json:"building"`
	Floor         string     `json:"floor"`
	Equipment     []int      `json:"equipment"`
	Coordinate    Coordinate `json:"coordinate"`
	NeedApproval  int        `json:"need_approval"`
}

type MeetingRoomListResponse struct {
	ErrCode         int           `json:"errcode"`
	ErrMsg          string        `json:"errmsg"`
	MeetingRoomList []MeetingRoom `json:"meetingroom_list"`
}

// ListMeetingRoom 查询会议室列表
func ListMeetingRoom() ([]MeetingRoom, error) {
	url := "https://qyapi.weixin.qq.com/cgi-bin/oa/meetingroom/list?access_token=%s"
	req, err := http.NewRequest("POST", fmt.Sprintf(url, accessToken()), nil)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var response MeetingRoomListResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrap(err)
	}
	return response.MeetingRoomList, nil
}

type Request struct {
	MeetingRoomID int    `json:"meetingroom_id"`
	StartTime     int64  `json:"start_time"`
	EndTime       int64  `json:"end_time"`
	City          string `json:"city"`
	Building      string `json:"building"`
	Floor         string `json:"floor"`
}

type Schedule struct {
	BookingID  string `json:"booking_id"`
	ScheduleID string `json:"schedule_id"`
	StartTime  int64  `json:"start_time"`
	EndTime    int64  `json:"end_time"`
	Booker     string `json:"booker"`
	Status     int    `json:"status"`
}

type Booking struct {
	MeetingRoomID int        `json:"meetingroom_id"`
	Schedule      []Schedule `json:"schedule"`
}

type Response struct {
	ErrCode     int       `json:"errcode"`
	ErrMsg      string    `json:"errmsg"`
	BookingList []Booking `json:"booking_list"`
}

// GetUnbookedMeetingRoom 获取在指定时间内未预定的会议室
func getUnbookedMeetingRoom(startTime, endTime int64) []int {
	// 请求参数
	requestData := Request{
		StartTime: startTime,
		EndTime:   endTime,
	}
	// 转换请求参数为 JSON
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		fmt.Println("JSON marshal error:", err)
		return nil
	}
	fmt.Println("meetingroom/get_booking_info", string(requestBody))
	url := "https://qyapi.weixin.qq.com/cgi-bin/oa/meetingroom/get_booking_info?access_token=%s"
	// 创建 HTTP 请求
	s := accessToken()
	fmt.Println("accessToken: ", s)
	req, err := http.NewRequest("POST", fmt.Sprintf(url, s), bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("NewRequest error:", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Do error:", err)
		return nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	// 读取 HTTP 响应内容
	var responseData Response
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		fmt.Println("JSON decode error:", err)
		return nil
	}
	var unbookedMeetingRoom []int
	for _, booking := range responseData.BookingList {
		if len(booking.Schedule) == 0 {
			unbookedMeetingRoom = append(unbookedMeetingRoom, booking.MeetingRoomID)
		}
	}
	return unbookedMeetingRoom
}

// 定日程

type ScheduleReq struct {
	StartTime int64               `json:"start_time"`
	EndTime   int64               `json:"end_time"`
	Attendees []map[string]string `json:"attendees"`
	Summary   string              `json:"summary"`
}

type ScheduleResponse struct {
	Errcode    int    `json:"errcode"`
	Errmsg     string `json:"errmsg"`
	ScheduleId string `json:"schedule_id"`
}

func tryOrderMeetingRoom(meetingRoomId int, startTime int64, endTime int64, emails []string, userId string) error {
	userIdList := append(getUserId(emails), userId)
	// 初始化请求体
	data := map[string]interface{}{
		"meetingroom_id": meetingRoomId,
		"subject":        "会议",
		"start_time":     startTime,
		"end_time":       endTime,
		"booker":         userId,
		"attendees":      userIdList,
	}
	requestBody, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err)
	}
	fmt.Println("request meetingroom/book: ", string(requestBody))

	url := "https://qyapi.weixin.qq.com/cgi-bin/oa/meetingroom/book?access_token=%s"
	req, err := http.NewRequest("POST", fmt.Sprintf(url, accessToken()), bytes.NewBuffer(requestBody))
	if err != nil {
		return errors.Wrap(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	type Response struct {
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
		BookingID  string `json:"booking_id"`
		ScheduleID string `json:"schedule_id"`
	}

	r := Response{}
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("tryOrderMeetingRoom resp: ", string(bodyBytes))
	if err = json.Unmarshal(bodyBytes, &r); err != nil {
		log.Fatal(err)
	}
	if r.ErrCode == 0 {
		return nil
	}
	return errors.New(r.ErrMsg)
}

func getUserId(emails []string) []string {
	userIds := make([]string, 0)
	for _, email := range emails {
		// 初始化请求体
		data := map[string]interface{}{
			"email":      email,
			"email_type": 1,
		}
		requestBody, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}
		url := "https://qyapi.weixin.qq.com/cgi-bin/user/get_userid_by_email?access_token=%s"
		// 发送HTTP请求
		req, err := http.NewRequest("POST", fmt.Sprintf(url, accessToken()), bytes.NewBuffer(requestBody))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		// 处理响应
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		if s, ok := result["userid"].(string); ok {
			userIds = append(userIds, s)
		}
	}
	return userIds
}
