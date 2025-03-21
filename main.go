package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"time-format-alfred/dateparse"
	"time-format-alfred/model"
)

var (
	paramTime     string
	formatPattern string
	paramLoc      = "Local"
	help          bool
	icon          = model.Icon{
		Path: "./icon.png",
	}
	resultItems = model.Items{
		Items: make([]model.Item, 0, 3),
	}
	regx = regexp.MustCompile("gmt|utc|UTC")
)

func init() {
	flag.StringVar(&paramTime, "time", "", "时间信息，支持多种格式")
	flag.StringVar(&formatPattern, "format", "", "时间格式")
	flag.BoolVar(&help, "h", false, "this help")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	dotIndex := strings.LastIndex(paramTime, ",")
	if dotIndex > 0 {
		// 时区真是dt
		paramLoc = regx.ReplaceAllString(paramTime[dotIndex+1:], "GMT")
		if strings.HasPrefix(paramLoc, "GMT") {
			if strings.LastIndex(paramLoc, "+") > 0 {
				paramLoc = strings.Replace(paramLoc, "+", "-", -1)
			} else {
				paramLoc = strings.Replace(paramLoc, "-", "+", -1)
			}
			paramLoc = "Etc/" + paramLoc
		}
		paramTime = paramTime[:dotIndex]
	}

	// 支持now
	if strings.HasPrefix(paramTime, "now") {
		paramTime = strconv.FormatInt(time.Now().Unix(), 10)
	}

	defaultLoc, _ := time.LoadLocation(paramLoc)
	result, e := dateparse.ParseIn(paramTime, defaultLoc)
	if e != nil {
		formatError(e)
		return
	}

	formatTimestampPattern(result.UnixNano(), formatPattern)

	zoneArgs := []string{time.Local.String()}
	zoneArgs = append(zoneArgs, flag.Args()...)
	formatTimestamp(result.UnixNano(), zoneArgs)
}

func formatTimestampPattern(timeNano int64, formatPattern string) {
	defer func() {
		if p := recover(); p != nil {
			formatError(fmt.Errorf("defer %+v", p))
		}
	}()
	patterns := strings.Split(formatPattern, ",")
	unix := time.Unix(convertSecond(timeNano), timeNano%1000000)
	for _, v := range patterns {
		dt := dateparse.FormatDate(unix, dateparse.DateStyle(v))
		if dt != "" {
			result := dt
			item := model.Item{
				Uid:      "1",
				Title:    v,
				Subtitle: result,
				Arg:      result,
			}
			resultItems.Items = append(resultItems.Items, item)
		}
	}
}

// 错误信息输出
func formatError(e error) {
	item := model.Item{
		Uid:      "1",
		Title:    "无法解析该格式",
		Subtitle: e.Error(),
		Icon:     icon,
	}
	resultItems.Items = append(resultItems.Items, item)
	bytes, _ := json.Marshal(resultItems)
	fmt.Println(string(bytes))
}

// 按照指定时区输出
func formatTimestamp(timeNano int64, timeZones []string) {
	unix := time.Unix(convertSecond(timeNano), timeNano%1000000)
	addTimeStampItem(timeNano)
	for _, zone := range timeZones {
		loc, _ := time.LoadLocation(zone)
		result := unix.In(loc).Format("2006-01-02T15:04:05 -07:00 MST")
		item := model.Item{
			Uid:      "1",
			Title:    loc.String(),
			Subtitle: result,
			Arg:      result,
			Icon:     icon,
		}
		resultItems.Items = append(resultItems.Items, item)
	}
	bytes, _ := json.Marshal(resultItems)
	fmt.Println(string(bytes))
}

func addTimeStampItem(timeNano int64) {
	timeStamp := strconv.FormatInt(convertMillisecond(timeNano), 10)
	item := model.Item{
		Uid:      "1",
		Title:    "TimeStamp",
		Subtitle: timeStamp + " ms",
		Arg:      timeStamp,
		Icon:     icon,
	}
	resultItems.Items = append(resultItems.Items, item)
}

func convertSecond(timeNano int64) int64 {
	return timeNano / int64(time.Second)
}

func convertMillisecond(timeNano int64) int64 {
	return timeNano / int64(time.Millisecond)
}
