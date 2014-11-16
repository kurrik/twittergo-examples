// Copyright 2011 Arne Roomann-Kurrik
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

package main

import (
	"fmt"
	"github.com/kurrik/oauth1a"
	"github.com/codingneo/twittergo"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"flag"

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
	user := oauth1a.NewAuthorizedConfig(lines[2], lines[3])
	client = twittergo.NewClient(config, user, "api.twitter.com")
	return
}

type Args struct {
	Id string
}

func parseArgs() *Args {
	a := &Args{}
	flag.StringVar(&a.Id, "id", "", "Id to look up")
	flag.Parse()
	return a
}

func main() {
	var (
		err    error
		client *twittergo.Client
		req    *http.Request
		resp   *twittergo.APIResponse
		tweet  *twittergo.Tweet
		args   *Args
	)

	args = parseArgs()

	client, err = LoadCredentials()
	if err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	query := url.Values{}
	query.Set("id", args.Id)
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
	tweet = &twittergo.Tweet{}
	err = resp.Parse(tweet)
	if err != nil {
		if rle, ok := err.(twittergo.RateLimitError); ok {
			fmt.Printf("Rate limited, reset at %v\n", rle.Reset)
		} else if errs, ok := err.(twittergo.Errors); ok {
			for i, val := range errs.Errors() {
				fmt.Printf("Error #%v - ", i + 1)
				fmt.Printf("Code: %v ", val.Code())
				fmt.Printf("Msg: %v\n", val.Message())
			}
		} else {
			fmt.Printf("Problem parsing response: %v\n", err)
		}
		os.Exit(1)
	}
	fmt.Printf("ID:                   %v\n", tweet.Id())
	fmt.Printf("Tweet:                %v\n", tweet.Text())
	fmt.Printf("User:                 %v\n", tweet.User().Name())
	if resp.HasRateLimit() {
		fmt.Printf("Rate limit:           %v\n", resp.RateLimit())
		fmt.Printf("Rate limit remaining: %v\n", resp.RateLimitRemaining())
		fmt.Printf("Rate limit reset:     %v\n", resp.RateLimitReset())
	} else {
		fmt.Printf("Could not parse rate limit from response.\n")
	}
}
