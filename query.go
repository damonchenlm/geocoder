package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const routineCountTotal = 5

var client *http.Client

type GLimit struct {
	Num int
	C   chan struct{}
}

func NewG(num int) *GLimit {
	return &GLimit{
		Num: num,
		C:   make(chan struct{}, num),
	}
}

func (g *GLimit) Run(f func()) {
	g.C <- struct{}{}
	go func() {
		f()
		<-g.C
	}()
}

type Point struct {
	Coordinates []float64
	Type string
}

func queryRequest(keyword string) (bool,[]float64, error) {
	// 生成请求 webUrl
	keyword = strings.ReplaceAll(keyword, " ", "%20")
	key := "Ah-r2Ekqur182_bwwc8tJvrXKzM-llqvr0N88KAwosvNEfrkZjV2O71wurPWmLHP"

	webUrl := fmt.Sprintf("http://dev.virtualearth.net/REST/v1/Locations/%s?include=queryParse&maxRes=1&key=%s", keyword, key)
	resp, err := client.Get(webUrl)
	if err != nil {
		return false, []float64{}, err
	}
	if resp.StatusCode != http.StatusOK {
		data, _ := ioutil.ReadAll(resp.Body)
		return false,[]float64{}, fmt.Errorf("response status code is not OK, response code is %d, body:%s", resp.StatusCode, string(data))
	}
	if resp != nil && resp.Body != nil {
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, []float64{}, err
	}

	// 解析 json
	var tmp map[string]interface{}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		return false,[]float64{}, err
	}
	str, err := json.Marshal(tmp["resourceSets"])
	if err != nil {
		return false,[]float64{}, err
	}
	var resourceSets []map[string]interface{}
	_ = json.Unmarshal(str, &resourceSets)
	var estimatedTotal float64
	estimatedTotal = resourceSets[0]["estimatedTotal"].(float64)
	if estimatedTotal > 0 {
		str, err := json.Marshal(resourceSets[0]["resources"])
		if err != nil {
			return false,[]float64{}, err
		}
		var resources []map[string]interface{}
		_ = json.Unmarshal(str, &resources)
		str, err = json.Marshal(resources[0]["point"])
		if err != nil {
			return false,[]float64{}, err
		}
		var point Point
		_ = json.Unmarshal(str, &point)
		return true, point.Coordinates, nil
	}

	return false, []float64{},nil
}
