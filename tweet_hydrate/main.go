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

// Hydrates a set of Tweets, specified by ID, to a file
package main

import (
	"bufio"
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
	InputFile  string
	OutputFile string
}

func parseArgs() *Args {
	a := &Args{}
	flag.StringVar(&a.ScreenName, "in", "tweet_ids.tsv", "Input file")
	flag.StringVar(&a.OutputFile, "out", "hydrated.tsv", "Output file")
	flag.Parse()
	return a
}

func getIds(scanner *bufio.Scanner, count int) (out string, err error) {
	buff := make([]string, 0, count)
	for scanner.Scan() {
		buff = append(buff, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return
	}
	out = strings.Join(buff, "\n")
	return
}

func main() {
	var (
		err     error
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		args    *Args
		ids     string
		out     *os.File
		in      *os.File
		query   url.Values
		results *twittergo.Timeline
		text    []byte
		scanner *bufio.Scanner
	)
	args = parseArgs()
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	if in, err = os.Open(args.InputFile); err != nil {
		fmt.Printf("Could not read input file %v: %v\n", args.InputFile, err)
		os.Exit(1)
	}
	defer in.Close()
	scanner = bufio.NewScanner(in)
	if out, err = os.Create(args.OutputFile); err != nil {
		fmt.Printf("Could not create output file %v: %v\n", args.OutputFile, err)
		os.Exit(1)
	}
	defer out.Close()
	const (
		count   int = 100
		urltmpl     = "/1.1/statuses/lookup.json?%v"
		minwait     = time.Duration(10) * time.Second
	)
	query = url.Values{}
	query.Set("map", "true")
	query.Set("trim_user", "true")
	total := 0
	for {
		if ids, err = getIds(scanner, count); err != nil {
			fmt.Printf("Problem reading IDs: %v\n", err)
			os.Exit(1)
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
