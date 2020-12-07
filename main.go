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

type responseJSON struct {
	Channels []channelsJSON `json:"channels"`
}

type channelsJSON struct {
	Name        string          `json:"name"`
	CreatedAt   int64           `json:"created"`
	Topic       topicJSON       `json:"topic"`
	Description descriptionJSON `json:"purpose"`
}

type topicJSON struct {
	Value string `json:"value"`
}

type descriptionJSON struct {
	Value string `json:"value"`
}

const (
	slackGetChannelAPIURL  = "https://slack.com/api/conversations.list?token="
	slackPostMessageAPIURL = "https://slack.com/api/chat.postMessage"
)

func main() {
	flag.Parse()
	targetDateStr := flag.Arg(0)

	// 実行時オプションの指定がなければデフォルトで前日日付を設定
	if flag.Arg(0) == "" {
		targetDateStr = time.Now().AddDate(0, 0, -1).Format("20060102")
	}
	targetDate, err := time.Parse("20060102", targetDateStr)
	if err != nil {
		fmt.Println(err)
	}

	targetDateUnixTime := targetDate.Unix()

	postChannel := os.Getenv("POST_CHANNEL")

	token := os.Getenv("TOKEN")
	reqURL := fmt.Sprintf("%s%s", slackGetChannelAPIURL, token)

	client := http.Client{}

	req, err := http.NewRequest("GET", reqURL, nil)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	var responseJSON responseJSON
	err = json.Unmarshal(body, &responseJSON)
	if err != nil {
		fmt.Println(err)
	}

	sendText := []string{fmt.Sprintf("%s%s%s", targetDate.Format("2006/01/02"), "以降に作成されたチャンネル一覧", "\n")}
	for _, v := range responseJSON.Channels {
		if targetDateUnixTime > v.CreatedAt {
			continue
		}

		sendText = append(sendText, fmt.Sprintf("%s%s%s", "チャンネル名:", v.Name, "\n"))

		t := time.Unix(v.CreatedAt, 0)
		sendText = append(sendText, fmt.Sprintf("%s%s%s", "作成日:", t.Format("2006/01/02"), "\n"))

		if v.Topic.Value != "" {
			sendText = append(sendText, fmt.Sprintf("%s%s%s", "トピック:", v.Topic.Value, "\n"))
		}

		if v.Description.Value != "" {
			sendText = append(sendText, fmt.Sprintf("%s%s%s", "説明:", v.Description.Value, "\n"))
		}
		sendText = append(sendText, "\n")

	}

	values := url.Values{}
	values.Set("token", token)
	values.Add("channel", postChannel)
	values.Add("text", strings.Join(sendText, "\n"))
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
	body, err = ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
}
