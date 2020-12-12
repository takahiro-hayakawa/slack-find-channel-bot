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
	MetaData metaDataJSON   `json:"response_metadata"`
}

type channelsJSON struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	CreatedAt   int64           `json:"created"`
	Topic       topicJSON       `json:"topic"`
	Description descriptionJSON `json:"purpose"`
	MembersNum  int64           `json:"num_members"`
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

	workSpace := os.Getenv("WORK_SPACE")
	slackWorkSpaceURL := fmt.Sprintf("https://%s/archives/", workSpace)
	postChannel := os.Getenv("POST_CHANNEL")
	token := os.Getenv("TOKEN")

	client := http.Client{}

	var cursor string
	sendText := []string{fmt.Sprintf("%s%s\n", targetDate.Format("2006/01/02"), "以降に作成されたチャンネル一覧")}

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

		for _, v := range responseJSON.Channels {
			if targetDateUnixTime > v.CreatedAt {
				continue
			}

			sendText = append(sendText, "==================================\n")
			sendText = append(sendText, fmt.Sprintf("%s<%s/%s|#%s>\n", "チャンネル名:", slackWorkSpaceURL, v.ID, v.Name))
			sendText = append(sendText, fmt.Sprintf("%s%d\n", "参加人数:", v.MembersNum))

			t := time.Unix(v.CreatedAt, 0)
			sendText = append(sendText, fmt.Sprintf("%s%s\n", "作成日:", t.Format("2006/01/02")))

			if v.Topic.Value != "" {
				sendText = append(sendText, fmt.Sprintf("%s%s\n", "トピック:", v.Topic.Value))
			}

			if v.Description.Value != "" {
				sendText = append(sendText, fmt.Sprintf("%s%s\n", "説明:", v.Description.Value))
			}
		}

		res.Body.Close()

		if responseJSON.MetaData.NextCursor == "" {
			break
		}
		cursor = responseJSON.MetaData.NextCursor
	}

	sendText = append(sendText, "==================================\n")

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
}
