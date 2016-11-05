// Copyright 2012 Arne Roomann-Kurrik
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

// Downloads a user's favorites timeline and writes it to a file.
package main

// Reads as much of a user's public Likes (formerly Favorites) as it can
// and prints each Tweet to a file.
//
// This example respects rate limiting and will wait until the rate limit
// reset time to finish pulling a timeline.
//
// Example non-rate-limited call:
//   $ go run examples/user_timeline/main.go -screen_name=kurrik
//   Got 100 Tweets, 39 calls available.
//   Got 100 Tweets, 38 calls available.
//   Got 100 Tweets, 37 calls available.
//   Got 100 Tweets, 36 calls available.
//   Got 100 Tweets, 35 calls available.
//   Got 100 Tweets, 34 calls available.
//   Got 100 Tweets, 33 calls available.
//   Got 99 Tweets, 32 calls available.
//   Got 100 Tweets, 31 calls available.
//   Got 100 Tweets, 30 calls available.
//   Got 100 Tweets, 29 calls available.
//   Got 100 Tweets, 28 calls available.
//   Got 100 Tweets, 27 calls available.
//   Got 99 Tweets, 26 calls available.
//   Got 100 Tweets.
//   Got 100 Tweets, 24 calls available.
//   Got 99 Tweets, 23 calls available.
//   Got 99 Tweets, 22 calls available.
//   Got 100 Tweets, 21 calls available.
//   Got 100 Tweets, 20 calls available.
//   Got 98 Tweets, 19 calls available.
//   Got 100 Tweets, 18 calls available.
//   Got 100 Tweets, 17 calls available.
//   Got 100 Tweets, 16 calls available.
//   Got 100 Tweets, 15 calls available.
//   Got 100 Tweets, 14 calls available.
//   Got 100 Tweets, 13 calls available.
//   Got 100 Tweets, 12 calls available.
//   Got 100 Tweets, 11 calls available.
//   Got 100 Tweets, 10 calls available.
//   Got 42 Tweets, 9 calls available.
//   No more results, end of timeline.
//   --------------------------------------------------------
//   Wrote 3036 Tweets to user_timeline.json
//

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
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
	client = twittergo.NewClient(config, user)
	return
}

type Args struct {
	ScreenName string
	OutputFile string
}

func parseArgs() *Args {
	a := &Args{}
	flag.StringVar(&a.ScreenName, "screen_name", "twitterapi", "Screen name")
	flag.StringVar(&a.OutputFile, "out", "favorites.json", "Output file")
	flag.Parse()
	return a
}

func main() {
	var (
		err     error
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		args    *Args
		max_id  uint64
		out     *os.File
		query   url.Values
		results *twittergo.Timeline
		text    []byte
	)
	args = parseArgs()
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	if out, err = os.Create(args.OutputFile); err != nil {
		fmt.Printf("Could not create output file: %v\n", args.OutputFile)
		os.Exit(1)
	}
	defer out.Close()
	const (
		count   int = 200
		urltmpl     = "/1.1/favorites/list.json?%v"
		minwait     = time.Duration(10) * time.Second
	)
	query = url.Values{}
	query.Set("count", fmt.Sprintf("%v", count))
	query.Set("screen_name", args.ScreenName)
	total := 0
	for {
		if max_id != 0 {
			query.Set("max_id", fmt.Sprintf("%v", max_id))
		}
		endpoint := fmt.Sprintf(urltmpl, query.Encode())
		if req, err = http.NewRequest("GET", endpoint, nil); err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			os.Exit(1)
		}
		if resp, err = client.SendRequest(req); err != nil {
			fmt.Printf("Could not send request: %v\n", err)
			os.Exit(1)
		}
		results = &twittergo.Timeline{}
		if err = resp.Parse(results); err != nil {
			if rle, ok := err.(twittergo.RateLimitError); ok {
				dur := rle.Reset.Sub(time.Now()) + time.Second
				if dur < minwait {
					// Don't wait less than minwait.
					dur = minwait
				}
				msg := "Rate limited. Reset at %v. Waiting for %v\n"
				fmt.Printf(msg, rle.Reset, dur)
				time.Sleep(dur)
				continue // Retry request.
			} else {
				fmt.Printf("Problem parsing response: %v\n", err)
			}
		}
		batch := len(*results)
		if batch == 0 {
			fmt.Printf("No more results, end of timeline.\n")
			break
		}
		for _, tweet := range *results {
			if text, err = json.Marshal(tweet); err != nil {
				fmt.Printf("Could not encode Tweet: %v\n", err)
				os.Exit(1)
			}
			out.Write(text)
			out.Write([]byte("\n"))
			max_id = tweet.Id() - 1
			total += 1
		}
		fmt.Printf("Got %v Tweets", batch)
		if resp.HasRateLimit() {
			fmt.Printf(", %v calls available", resp.RateLimitRemaining())
		}
		fmt.Printf(".\n")
	}
	fmt.Printf("--------------------------------------------------------\n")
	fmt.Printf("Wrote %v Tweets to %v\n", total, args.OutputFile)
}
