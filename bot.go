package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"github.com/uber/gonduit"
	"github.com/uber/gonduit/core"
	"github.com/uber/gonduit/requests"
)

var phabricatorClient *gonduit.Conn
var telegramClient *tgbotapi.BotAPI

var emojiForTypes = map[string]string{
	"TASK": "📝",
	"USER": "👤",
	"WIKI": "📖",
	"CEVT": "📅",
	"PROJ": "📁",
}

var emojiForActions = map[string]string{

	"created":  "\U0001F4A1",
	"added":    "\U0001F4AC",
	"lowered":  "\U0001F53B",
	"raised":   "\U0001F53A",
	"awarded":  "\U0001F3C6",
	"triaged":  "\U0001F4AD",
	"updated":  "\U0001F449",
	"changed":  "\U0000270F\U0000FE0F ",
	"claimed":  "\U0001F44C",
	"set":      "\U0000270F\U0000FE0F ",
	"reopened": "\U0001F504",
	"closed":   "\U0001F510",
	"renamed":  "\U0001F449",
	"edited":   "\U0001F4DD",
}

// fetchFeed calls feed.query and then uses PHIDLookup to get the actual data for each feed item.
func fetchFeed(after string) ([]FeedItem, error) {
	var feed map[string]FeedQueryResponseItem
	req := &FeedQueryRequest{
		After: after,
		View:  "text",
	}
	err := phabricatorClient.Call("feed.query", req, &feed)
	if err != nil {
		return nil, fmt.Errorf("error fetching feed, %s", err)
	}
	// transpose to a list and sort by epoch
	// transpose to a list and sort by epoch
	feedList := make([]FeedQueryResponseItem, len(feed))
	i := 0
	for _, v := range feed {
		feedList[i] = v
		i++
	}

	sort.Slice(feedList, func(i, j int) bool {
		return feedList[i].Epoch < feedList[j].Epoch
	})
	phids := make([]string, 0, len(feedList)*2)
	for _, v := range feedList {
		phids = append(phids, v.AuthorPHID, v.ObjectPHID)
	}
	phids = removeDuplicates(phids)
	lookedUpPhids, err := phabricatorClient.PHIDLookup(requests.PHIDLookupRequest{
		Names: phids,
	})
	if err != nil {
		return nil, fmt.Errorf("error looking up phids, %s", err)
	}

	var feedItems []FeedItem
	for _, v := range feedList {
		feedItems = append(feedItems, FeedItem{
			URL:              lookedUpPhids[v.ObjectPHID].URI,
			Title:            lookedUpPhids[v.ObjectPHID].FullName,
			Time:             time.Unix(int64(v.Epoch), 0).Format(time.RFC1123),
			Author:           lookedUpPhids[v.AuthorPHID].FullName,
			Type:             lookedUpPhids[v.ObjectPHID].Type,
			TypeName:         lookedUpPhids[v.ObjectPHID].TypeName,
			TimeData:         time.Unix(int64(v.Epoch), 0),
			Text:             v.Text,
			ChronologicalKey: v.ChronologicalKey,
		})
	}

	return feedItems, nil

}

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/run/secrets/")
	viper.AddConfigPath("/etc/phabricator-tele-notifier/")
	viper.AddConfigPath("$HOME/.config/phabricator-tele-notifier")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	phabricatorClient, err = gonduit.Dial(viper.GetString("phabricator.url"), &core.ClientOptions{
		APIToken: viper.GetString("phabricator.token"),
		Timeout:  time.Second * 20,
		// Client:   client,
	})
	if err != nil {
		log.Fatalf("Error connecting to phabricator, %s", err)
	}

	runTaskServer()

	telegramClient, err = tgbotapi.NewBotAPI(viper.GetString("telegram.token"))
	if err != nil {
		log.Fatalf("Error connecting to telegram, %s", err)
	}
	self, err := telegramClient.GetMe()
	if err != nil {
		log.Fatalf("Error getting telegram bot info, %s", err)
	}
	log.Printf("Authorized on account %s", self.UserName)
	chat, err := telegramClient.GetChat(tgbotapi.ChatInfoConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: viper.GetInt64("telegram.chat_id"),
		},
	})
	if err != nil {
		log.Fatalf("Error getting chat with id %v, %s", viper.GetInt64("telegram.chat_id"), err)
	}
	log.Printf("Will send messages to %#v", chat.Title)

	notifyTypes := viper.GetStringSlice("telegram.notify_types")
	if len(notifyTypes) == 0 {
		notifyTypes = []string{"TASK", "USER", "WIKI", "CEVT", "PROJ"}
		log.Printf("No notify types specified, defaulting to %v", notifyTypes)
	}
	notifyTypesMap := make(map[string]bool)
	for _, v := range notifyTypes {
		notifyTypesMap[v] = true
	}

	var lastMsgTime = time.Now()
	for {
		feedItems, err := fetchFeed("")
		if err != nil {
			log.Fatalf("Error fetching feed, %s", err)
		}

		log.Printf("Fetched feed, got %v items", len(feedItems))

		var limit = 0
		for _, v := range feedItems {
			if !notifyTypesMap[v.Type] || v.TimeData.Before(lastMsgTime) || v.TimeData == lastMsgTime {
				continue
			}

			actionEmoji := ""
			for k, e := range emojiForActions {
				if strings.Contains(v.Text, k) {
					actionEmoji = e
					break
				}
			}

			text := fmt.Sprintf("%s <b>%s</b>\n%s %s\nAutor: %s\nCzas: %s", emojiForTypes[v.Type], v.Title, actionEmoji, v.Text, v.Author, v.Time)

			msg := tgbotapi.NewMessage(viper.GetInt64("telegram.chat_id"), text)
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonURL(v.URL, v.URL),
				),
			)

			telegramClient.Send(msg)
			lastMsgTime = v.TimeData
			limit++
			if limit > 10 {
				log.Printf("Limit reached, stopping for 5 seconds")
				time.Sleep(time.Second * 5)
				limit = 0
			}
		}
		time.Sleep(viper.GetDuration("phabricator.poll_interval"))
	}
}
