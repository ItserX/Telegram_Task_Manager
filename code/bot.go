package main

import (
	"context"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	BotToken = ""
)

func startTaskBot(ctx context.Context) error {

	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		return err

	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)

	updates := bot.GetUpdatesChan(u)
	api := Handler{
		Users:     make(map[int64]*User),
		allTasks:  make(map[int]*Task),
		taskCount: 0,
		commands: []string{
			"/tasks",
			"/new",
			"/assign",
			"/unassign",
			"/resolve",
			"/my",
			"/owner",
		},
		mu: &sync.Mutex{},
	}
	for update := range updates {
		if update.Message != nil {

			userID := update.Message.Chat.ID
			messageText := update.Message.Text
			userName := update.Message.Chat.UserName

			if _, ok := api.Users[userID]; !ok {
				api.mu.Lock()
				api.Users[userID] = &User{Name: userName, ID: userID}
				api.Users[userID].MyTaskIDs = make([]int, 0)
				api.Users[userID].OwnerTaskIDs = make([]int, 0)
				api.mu.Unlock()
			}
			switch api.DefineCommand(messageText) {

			case "/tasks":
				go api.SendTasks(bot, userID)

			case "/new":
				go api.NewTask(bot, userID, messageText[5:])

			case "/assign":
				go api.RunCommand(bot, messageText, userID, api.Assign)

			case "/unassign":
				go api.RunCommand(bot, messageText, userID, api.UnAssign)

			case "/resolve":
				go api.RunCommand(bot, messageText, userID, api.Resolve)

			case "/my":
				go api.MyTasks(bot, userID)

			case "/owner":
				go api.OwnerTasks(bot, userID)
			}
		}
	}
	<-ctx.Done()

	return nil
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}

}
