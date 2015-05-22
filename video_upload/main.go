// Copyright 2015 Arne Roomann-Kurrik
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
	"bytes"
	"fmt"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	"io"
	"io/ioutil"
	"mime/multipart"
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

func SendApiRequest(client *twittergo.Client, reqUrl string, params map[string]string) (resp *twittergo.APIResponse, err error) {
	var (
		body io.Reader
		data url.Values = url.Values{}
		req  *http.Request
	)
	for key, value := range params {
		data.Set(key, value)
	}
	body = strings.NewReader(data.Encode())
	if req, err = http.NewRequest("POST", reqUrl, body); err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.SendRequest(req)
	return
}

func SendMediaRequest(client *twittergo.Client, reqUrl string, params map[string]string, media []byte) (mediaResp twittergo.MediaResponse, err error) {
	var (
		req         *http.Request
		resp        *twittergo.APIResponse
		body        io.ReadWriter = bytes.NewBufferString("")
		mp          *multipart.Writer
		writer      io.Writer
		contentType string
	)
	mp = multipart.NewWriter(body)
	for key, value := range params {
		mp.WriteField(key, value)
	}
	if media != nil {
		if writer, err = mp.CreateFormField("media"); err != nil {
			return
		}
		writer.Write(media)
	}
	contentType = fmt.Sprintf("multipart/form-data;boundary=%v", mp.Boundary())
	mp.Close()
	if req, err = http.NewRequest("POST", reqUrl, body); err != nil {
		return
	}
	req.Header.Set("Content-Type", contentType)
	if resp, err = client.SendRequest(req); err != nil {
		return
	}
	err = resp.Parse(&mediaResp)
	return
}

func main() {
	var (
		err        error
		client     *twittergo.Client
		apiResp    *twittergo.APIResponse
		mediaResp  twittergo.MediaResponse
		mediaId    string
		mediaBytes []byte
	)
	client, err = LoadCredentials()
	if err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	if mediaBytes, err = ioutil.ReadFile("video_upload/twitter_media_upload.mp4"); err != nil {
		fmt.Printf("Error reading media: %v\n", err)
		os.Exit(1)
	}
	if mediaResp, err = SendMediaRequest(
		client,
		"https://upload.twitter.com/1.1/media/upload.json",
		map[string]string{
			"command":     "INIT",
			"media_type":  "video/mp4",
			"total_bytes": fmt.Sprintf("%d", len(mediaBytes)),
		},
		nil,
	); err != nil {
		fmt.Printf("Problem sending INIT request: %v\n", err)
		os.Exit(1)
	}
	mediaId = fmt.Sprintf("%v", mediaResp.MediaId())
	if mediaResp, err = SendMediaRequest(
		client,
		"https://upload.twitter.com/1.1/media/upload.json",
		map[string]string{
			"command":       "APPEND",
			"media_id":      mediaId,
			"segment_index": "0",
		},
		mediaBytes,
	); err != nil {
		fmt.Printf("Problem sending APPEND request: %v\n", err)
		os.Exit(1)
	}
	if mediaResp, err = SendMediaRequest(
		client,
		"https://upload.twitter.com/1.1/media/upload.json",
		map[string]string{
			"command":  "FINALIZE",
			"media_id": mediaId,
		},
		nil,
	); err != nil {
		fmt.Printf("Problem sending FINALIZE request: %v\n", err)
		os.Exit(1)
	}
	if apiResp, err = SendApiRequest(
		client,
		"/1.1/statuses/update.json",
		map[string]string{
			"status":    fmt.Sprintf("Media! %v", time.Now()),
			"media_ids": mediaId,
		},
	); err != nil {
		fmt.Printf("Problem sending Tweet request: %v\n", err)
		os.Exit(1)
	}
	var (
		tweet *twittergo.Tweet = &twittergo.Tweet{}
	)
	if err = apiResp.Parse(tweet); err != nil {
		fmt.Printf("Problem parsing Tweet response: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("ID:                         %v\n", tweet.Id())
	fmt.Printf("Tweet:                      %v\n", tweet.Text())
	fmt.Printf("User:                       %v\n", tweet.User().Name())
	fmt.Printf("Media Id:                   %v\n", mediaResp.MediaId())
	if apiResp.HasRateLimit() {
		fmt.Printf("Rate limit:                 %v\n", apiResp.RateLimit())
		fmt.Printf("Rate limit remaining:       %v\n", apiResp.RateLimitRemaining())
		fmt.Printf("Rate limit reset:           %v\n", apiResp.RateLimitReset())
	} else {
		fmt.Printf("Could not parse rate limit from response.\n")
	}
	if apiResp.HasMediaRateLimit() {
		fmt.Printf("Media Rate limit:           %v\n", apiResp.MediaRateLimit())
		fmt.Printf("Media Rate limit remaining: %v\n", apiResp.MediaRateLimitRemaining())
		fmt.Printf("Media Rate limit reset:     %v\n", apiResp.MediaRateLimitReset())
	} else {
		fmt.Printf("Could not parse media rate limit from response.\n")
	}
}
