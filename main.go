package main

import (
	"fmt"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"path/filepath"
)

type episode struct {
	id       string
	name     string
	soundId  string
	soundStr string
}

var headers = req.Header{
	"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36",
	"Referer":    "https://www.missevan.com",
}
var cookies string

func getDramaMessage(dramaId string) (map[string]gjson.Result, map[string]episode, error) {
	dramaMessage := make(map[string]gjson.Result)
	dramaEpisode := make(map[string]episode)
	param := req.Param{
		"drama_id": dramaId,
	}
	r, err := req.Get("https://www.missevan.com/dramaapi/getdrama", headers, param)
	if err != nil {
		return make(map[string]gjson.Result), make(map[string]episode), err
	}
	dramaMessage = gjson.Get(r.Dump(), "info.drama").Map()
	dramaEpisodeTemp := gjson.Get(r.Dump(), "info.episodes.episode").Array()
	for _, data := range dramaEpisodeTemp {
		episodeData := data.Map()
		id := episodeData["id"].String()
		dramaEpisode[id] = episode{id, episodeData["name"].String(), episodeData["sound_id"].String(), episodeData["soundstr"].String()}
	}
	return dramaMessage, dramaEpisode, nil
}

func getDramaSound(soundId string, fileLocation string) error {
	soundMessage, err := req.Get("https://www.missevan.com/sound/getsound?soundid=" + soundId)
	if err != nil {
		return err
	}
	soundUrl := gjson.Get(soundMessage.Dump(), "info.sound.soundurl").String()
	tempHeaders := headers
	tempHeaders["Cookie"] = cookies
	soundFile, err := req.Get(soundUrl, tempHeaders)
	if err != nil {
		return err
	}
	err = soundFile.ToFile(fileLocation)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var dramaId string
	bytes, err := ioutil.ReadFile("cookies.txt")
	if err != nil {
		println("Read cookie failed...")
		return
	}
	cookies = string(bytes)
	print("Enter the Drama ID: ")
	_, err = fmt.Scanln(&dramaId)
	if err != nil {
		println("Read ID failed...")
		return
	}
	dramaMessage, dramaEpisode, err := getDramaMessage(dramaId)
	println("广播剧：", dramaMessage["name"].String())
	err = os.Mkdir("广播剧", os.ModePerm)
	err = os.Mkdir(filepath.Join("广播剧", dramaMessage["name"].String()), os.ModePerm)
	if err != nil {
		println("Drama folder is exist! Please remove the folder and try again!")
		return
	}
	println("downloading...")
	count := 1
	for _, aEpisode := range dramaEpisode {
		fileLocation := filepath.Join("广播剧", dramaMessage["name"].String(), aEpisode.soundStr+".mp3")
		err = getDramaSound(aEpisode.soundId, fileLocation)
		if err != nil {
			println(err.Error())
			println("Download Failed! Please check the cookie or Drama ID!")
			break
		}
		fmt.Printf("%2d/%-3d%-10s%s\n", count, len(dramaEpisode), "SUCCESS:", aEpisode.soundStr)
		count += 1
	}
}
