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

func main() {
	var (
		err    error
		client *twittergo.Client
		req    *http.Request
		resp   *twittergo.APIResponse
		tweet  *twittergo.Tweet
	)
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	query := url.Values{
		"id": []string{"451453919017697280"},
	}
	url := fmt.Sprintf("/1.1/statuses/show.json?%v", query.Encode())
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		fmt.Printf("Could not parse request: %v\n", err)
		os.Exit(1)
	}
	if resp, err = client.SendRequest(req); err != nil {
		fmt.Printf("Could not send request: %v\n", err)
		os.Exit(1)
	}
	tweet = &twittergo.Tweet{}
	if err = resp.Parse(tweet); err != nil {
		fmt.Printf("There was an error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(tweet.Text())
}
