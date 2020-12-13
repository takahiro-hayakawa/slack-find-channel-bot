package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type channel struct {
	ID          string
	Name        string
	CreatedAt   int64
	Topic       string
	Description string
	MemberNum   int64
}

type responseJSON struct {
	Channels []channelsJSON `json:"channels"`
	MetaData metaDataJSON   `json:"response_metadata"`
}

type channelsJSON struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	CreatedAt   int64           `json:"created"`
	Topic       topicJSON       `json:"topic"`
	Description descriptionJSON `json:"purpose"`
	MemberNum   int64           `json:"num_members"`
}

type topicJSON struct {
	Value string `json:"value"`
}

type descriptionJSON struct {
	Value string `json:"value"`
}

type metaDataJSON struct {
	NextCursor string `json:"next_cursor"`
}

const (
	slackGetChannelAPIURL  = "https://slack.com/api/conversations.list"
	slackPostMessageAPIURL = "https://slack.com/api/chat.postMessage"
)

var slackWorkSpaceURL string
var postChannel string
var token string

func init() {
	workSpace := os.Getenv("WORK_SPACE")
	slackWorkSpaceURL = fmt.Sprintf("https://%s/archives/", workSpace)
	postChannel = os.Getenv("POST_CHANNEL")
	token = os.Getenv("TOKEN")
}

func main() {
	flag.Parse()
	targetDateStr := flag.Arg(0)

	// 実行時オプションの指定がなければデフォルトで前日日付を設定
	if flag.Arg(0) == "" {
		targetDateStr = time.Now().AddDate(0, 0, -1).Format("20060102")
	}
	targetDateTime, err := time.Parse("20060102", targetDateStr)
	if err != nil {
		fmt.Println(err)
	}

	client := http.Client{}
	channels := findChannelAfterTargetDate(&client, targetDateTime)
	message := makeSlackSendMessage(targetDateTime, channels)
	sendMessage(&client, message)
}

func findChannelAfterTargetDate(client *http.Client, targetDateTime time.Time) []channel {
	allChannels := findAllChannel(client)
	targetDateUnixTime := targetDateTime.Unix()

	var channels []channel
	for _, v := range allChannels {
		if targetDateUnixTime > v.CreatedAt {
			continue
		}
		channels = append(channels, v)
	}

	return channels
}

func findAllChannel(client *http.Client) []channel {
	var resJSON []responseJSON
	var cursor string

	for {
		req, err := http.NewRequest("GET", slackGetChannelAPIURL, nil)
		if err != nil {
			fmt.Println(err)
		}
		params := req.URL.Query()
		params.Add("limit", "1000")
		params.Add("token", token)

		if cursor != "" {
			params.Add("cursor", cursor)
		}
		req.URL.RawQuery = params.Encode()

		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}

		body, err := ioutil.ReadAll(res.Body)

		var responseJSON responseJSON
		err = json.Unmarshal(body, &responseJSON)
		if err != nil {
			fmt.Println(err)
		}
		res.Body.Close()
		resJSON = append(resJSON, responseJSON)

		if responseJSON.MetaData.NextCursor == "" {
			break
		}
		cursor = responseJSON.MetaData.NextCursor
	}

	var channels []channel
	for _, vRes := range resJSON {
		for _, v := range vRes.Channels {
			channel := channel{ID: v.ID, Name: v.Name, CreatedAt: v.CreatedAt, Topic: v.Topic.Value, Description: v.Description.Value, MemberNum: v.MemberNum}
			channels = append(channels, channel)
		}
	}

	return channels
}

func makeSlackSendMessage(targetDateTime time.Time, channels []channel) string {
	if len(channels) == 0 {
		return fmt.Sprintf("%s%s\n", targetDateTime.Format("2006/01/02"), "以降に作成されたチャンネルはありません")
	}

	sendMessage := []string{fmt.Sprintf("%s%s\n", targetDateTime.Format("2006/01/02"), "以降に作成されたチャンネル一覧")}
	for _, v := range channels {

		sendMessage = append(sendMessage, "====================================\n")
		sendMessage = append(sendMessage, fmt.Sprintf("%s<%s/%s|#%s>\n", "チャンネル名:", slackWorkSpaceURL, v.ID, v.Name))
		sendMessage = append(sendMessage, fmt.Sprintf("%s%d\n", "参加人数:", v.MemberNum))

		t := time.Unix(v.CreatedAt, 0)
		sendMessage = append(sendMessage, fmt.Sprintf("%s%s\n", "作成日:", t.Format("2006/01/02")))

		if v.Topic != "" {
			sendMessage = append(sendMessage, fmt.Sprintf("%s%s\n", "トピック:", v.Topic))
		}

		if v.Description != "" {
			sendMessage = append(sendMessage, fmt.Sprintf("%s%s\n", "説明:", v.Description))
		}
	}

	sendMessage = append(sendMessage, "====================================\n")
	return strings.Join(sendMessage, "\n")
}

func sendMessage(client *http.Client, message string) {
	values := url.Values{}
	values.Set("token", token)
	values.Add("channel", postChannel)
	values.Add("text", message)
	request, err := http.NewRequest("POST", slackPostMessageAPIURL, strings.NewReader(values.Encode()))
	if err != nil {
		fmt.Println(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}

	defer response.Body.Close()
}
