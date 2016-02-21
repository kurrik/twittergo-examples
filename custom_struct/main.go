// Copyright 2016 Arne Roomann-Kurrik
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

// This example demonstrates using custom struct models with `encoding/json`
// annotations.  You can use your own models when integrating with twittergo
// and still get the benefits of auth and standard response handling.

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

type CustomTweet struct {
	CustomID   string `json:"id_str"`
	CustomText string `json:"text"`
}

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
	user := oauth1a.NewAuthorizedConfig(lines[2], lines[3])
	client = twittergo.NewClient(config, user)
	return
}

func main() {
	var (
		err         error
		client      *twittergo.Client
		req         *http.Request
		resp        *twittergo.APIResponse
		customTweet *CustomTweet
	)
	client, err = LoadCredentials()
	if err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}

	query := url.Values{}
	query.Set("id", "641727937201311744") // https://twitter.com/kurrik/status/641727937201311744
	url := fmt.Sprintf("%v?%v", "/1.1/statuses/show.json", query.Encode())
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Could not parse request: %v\n", err)
		os.Exit(1)
	}
	resp, err = client.SendRequest(req)
	if err != nil {
		fmt.Printf("Could not send request: %v\n", err)
		os.Exit(1)
	}
	customTweet = &CustomTweet{}
	err = resp.Parse(customTweet)
	if err != nil {
		if rle, ok := err.(twittergo.RateLimitError); ok {
			fmt.Printf("Rate limited, reset at %v\n", rle.Reset)
		} else if errs, ok := err.(twittergo.Errors); ok {
			for i, val := range errs.Errors() {
				fmt.Printf("Error #%v - ", i+1)
				fmt.Printf("Code: %v ", val.Code())
				fmt.Printf("Msg: %v\n", val.Message())
			}
		} else {
			fmt.Printf("Problem parsing response: %v\n", err)
		}
		os.Exit(1)
	}
	fmt.Printf("ID:                   %v\n", customTweet.CustomID)
	fmt.Printf("Tweet:                %v\n", customTweet.CustomText)
	if resp.HasRateLimit() {
		fmt.Printf("Rate limit:           %v\n", resp.RateLimit())
		fmt.Printf("Rate limit remaining: %v\n", resp.RateLimitRemaining())
		fmt.Printf("Rate limit reset:     %v\n", resp.RateLimitReset())
	} else {
		fmt.Printf("Could not parse rate limit from response.\n")
	}
}
