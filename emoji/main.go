// Copyright 2014 Arne Roomann-Kurrik
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Attempts to fetch a Tweet containing Emoji
package main

import (
	"fmt"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func LoadCredentials() (client *twittergo.Client, err error) {
	credentials, err := ioutil.ReadFile("CREDENTIALS")
	if err != nil {
		return
	}
	lines := strings.Split(string(credentials), "\n")
	config := &oauth1a.ClientConfig{
		ConsumerKey:    lines[0],
		ConsumerSecret: lines[1],
	}
	client = twittergo.NewClient(config, nil)
	return
}

func GetTweet(client *twittergo.Client, id string) (tweet *twittergo.Tweet, err error) {
	var (
		query = url.Values{"id": []string{id}}
		url   = fmt.Sprintf("/1.1/statuses/show.json?%v", query.Encode())
		req   *http.Request
		resp  *twittergo.APIResponse
	)
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		err = fmt.Errorf("Could not parse request: %v\n", err)
		return
	}
	if resp, err = client.SendRequest(req); err != nil {
		err = fmt.Errorf("Could not send request: %v\n", err)
		return
	}
	tweet = &twittergo.Tweet{}
	if err = resp.Parse(tweet); err != nil {
		err = fmt.Errorf("There was an error: %v\n", err)
	}
	return
}

func main() {
	var (
		err    error
		client *twittergo.Client
		tweet  *twittergo.Tweet
	)
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	tweet_ids := []string{
		"451453919017697280",
		"451453883575848960",
		"451453847622262784",
		"451453811035357184",
		"451453775882899456",
		"451453738603909120",
		"451453703375953920",
		"451453667409817600",
		"451453613231964160",
		"451453567233040384",
		"451453517819949056",
		"451453478502539264",
		"451453436769218560",
	}

	for _, id := range tweet_ids {
		if tweet, err = GetTweet(client, id); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(tweet.Text())
	}
}
