package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

var DB *gorm.DB

type GeocodeRes struct {
	ID    uint
	Text  string
	QueryText string
	Lat   string
	Lon   string
	Valid bool
}

type QueryTask struct {
	ID   uint
	Text string
}

// clearText 对输入文本进行清洗，用于组成查询URL
func clearText(input string) string {
	//var queryArray []string
	// 对大小地点进行正则匹配
	// 处理特殊符号
	input = strings.ReplaceAll(input, ".", "")
	input = strings.ReplaceAll(input, "&", "")
	input = strings.ReplaceAll(input, ":", "")
	input = strings.ReplaceAll(input, ";", "")
	input = strings.ReplaceAll(input, "<", "")

	largeLocationMatch := regexp.MustCompile(`%[\s\S]*?%`)
	smallLocationMatch := regexp.MustCompile(`\$[\s\S]*?\$`)
	largeLocationMatchRes := largeLocationMatch.FindAllString(input, -1)
	smallLocationMatchRes := smallLocationMatch.FindAllString(input, -1)
	if len(largeLocationMatchRes)+len(smallLocationMatchRes) == 0 {
		// 如果没有匹配到任何结果
		return strings.ReplaceAll(input, " ", "%20")
	}
	if len(smallLocationMatchRes) > 0 {
		// 如果存在小地点
		for i := range smallLocationMatchRes {
			smallLocationMatchRes[i] = strings.Trim(strings.ReplaceAll(smallLocationMatchRes[i], " ", "%20"), "$")
		}
		return strings.Join(smallLocationMatchRes, "%20")
	} else {
		// 如果存在大地点
		for i := range largeLocationMatchRes {
			largeLocationMatchRes[i] = strings.Trim(strings.ReplaceAll(largeLocationMatchRes[i], " ", "%20"), "%")
		}
		return strings.Join(largeLocationMatchRes, "%20")
	}

}

func main() {
	dsn := "damon:Cyl851106@tcp(rm-bp1z6653s2t65e4774o.mysql.rds.aliyuncs.com:3306)/geocoder?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	_ = db.AutoMigrate(&GeocodeRes{})

	// 访问数据库获取所有 location 记录
	var locations []GeocodeRes
	db.Model(&GeocodeRes{}).Where("valid",false).Order("id ASC").Find(&locations)

	// 遍历 locations 获取请求任务列表
	var queryTasks []QueryTask
	for i := 0; i < len(locations); i++ {
		queryTasks = append(queryTasks, QueryTask{
			ID:   locations[i].ID,
			Text: locations[i].Text,
		})
	}
	client = &http.Client{
		Timeout: time.Second * 5,
	}
	g := NewG(routineCountTotal)
	wg := &sync.WaitGroup{}
	total := len(queryTasks)
	solved := 0
	beg := time.Now()
	for i := 0; i < total; i++ {
		wg.Add(1)
		task := queryTasks[i]
		g.Run(func() {
			var geocodeRes GeocodeRes
			geocodeRes.ID = task.ID
			db.Model(&GeocodeRes{}).Find(&geocodeRes)
			if geocodeRes.Lat != "" {
				return
			}
			queryText := clearText(task.Text)
			valid, res, err := queryRequest(queryText)
			if err != nil {
				fmt.Println(err)
				go db.Model(&GeocodeRes{}).Where("id = ?", task.ID).Updates(map[string]interface{}{"query_text":queryText,"valid":false})
			} else if valid == false {
				go db.Model(&GeocodeRes{}).Where("id = ?", task.ID).Updates(map[string]interface{}{"query_text":queryText,"valid":false})
			} else {
				go db.Model(&GeocodeRes{}).Where("id = ?", task.ID).Updates(map[string]interface{}{"valid": 1, "lat": res[0], "lon": res[1],"query_text":queryText})
			}
			solved++
			fmt.Println(solved, "/", total)
			wg.Done()
		})
	}
	wg.Wait()
	fmt.Printf("time consumed %fs", time.Now().Sub(beg).Seconds())

}
