package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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
	slackAPIURL = "https://slack.com/api/conversations.list?token="
)

func main() {
	token := os.Getenv("TOKEN")
	reqURL := fmt.Sprintf("%s%s", slackAPIURL, token)

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

	for _, v := range responseJSON.Channels {
		fmt.Println("チャンネル名:", v.Name)
		t := time.Unix(v.CreatedAt, 0)
		fmt.Println("作成日:", t.Format("2006/01/02"))

		if v.Topic.Value != "" {
			fmt.Println("Topic:", v.Topic.Value)
		}

		if v.Description.Value != "" {
			fmt.Println("Description:", v.Description.Value)
		}
	}
}
