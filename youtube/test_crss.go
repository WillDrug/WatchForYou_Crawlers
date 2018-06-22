package main
import (
	"./crss"
	"fmt"
	"time"
)

func main() {
	dateString := "2018-06-01T11:45:26.371Z"
    time1, _ := time.Parse(time.RFC3339,dateString)
	var test int64
	test = time1.Unix()
	fmt.Println(crss.GetFeedUpdatesWithPOSIX("https://www.youtube.com/feeds/videos.xml?channel_id=UCBa659QWEk1AI4Tg--mrJ2A",test))
}