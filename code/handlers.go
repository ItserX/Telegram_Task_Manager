package main

import (
	"sort"
	"strconv"
	"strings"

	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const ErrSendMsg = "send message error"

func (api *Handler) DefineCommand(command string) string {
	for _, val := range api.commands {
		if strings.Contains(command, val) {
			return val
		}
	}
	return ""
}

func (api *Handler) SendTasks(bot *tgbotapi.BotAPI, userID int64) {

	n := 0
	msg := ""
	tmp := make([]int, 0)
	for i := range api.allTasks {
		tmp = append(tmp, i)
	}
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i] < tmp[j]
	})

	var task *Task
	for _, taskID := range tmp {
		task = api.allTasks[taskID]
		if task.status {
			n += 1
			if n > 1 {
				msg += "\n"
			}
			msg += strconv.Itoa(taskID) + ". " + task.Name + " by @" + task.byUser.Name
			switch {
			case task.desUser.ID == 0:
				msg += "\n/assign_" + strconv.Itoa(taskID) + "\n"
			case task.desUser.ID != 0 && task.desUser.ID == userID:
				msg += "\nassignee: я" + "\n/unassign_" + strconv.Itoa(taskID) + " /resolve_" + strconv.Itoa(taskID) + "\n"
			case task.desUser.ID != 0 && task.desUser.ID != userID:
				msg += "\nassignee: @" + task.desUser.Name + "\n"
			}
		}
	}

	if n == 0 {
		msg := tgbotapi.NewMessage(userID, "Нет задач")
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf(ErrSendMsg)
		}
		return
	}

	msg = msg[:len(msg)-1]
	_, err := bot.Send(tgbotapi.NewMessage(userID, msg))
	if err != nil {
		log.Printf(ErrSendMsg)
	}
}

func (api *Handler) NewTask(bot *tgbotapi.BotAPI, userID int64, problem string) {

	api.mu.Lock()
	api.taskCount += 1
	api.allTasks[api.taskCount] = &Task{byUser: *api.Users[userID], Name: problem, status: true}
	api.Users[userID].OwnerTaskIDs = append(api.Users[userID].OwnerTaskIDs, api.taskCount)
	api.mu.Unlock()

	msg := tgbotapi.NewMessage(userID, "Задача "+"\""+problem+"\""+" создана, "+"id="+strconv.Itoa(api.taskCount))
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf(ErrSendMsg)
	}

}

func (api *Handler) Assign(bot *tgbotapi.BotAPI, userID int64, taskID int) {

	task := api.allTasks[taskID]
	if task.desUser.ID == 0 {
		api.mu.Lock()
		task.desUser = *api.Users[userID]
		api.Users[userID].MyTaskIDs = append(api.Users[userID].MyTaskIDs, taskID)
		currentMsg := tgbotapi.NewMessage(userID, "Задача "+"\""+task.Name+"\""+" назначена на вас")
		_, err := bot.Send(currentMsg)
		if err != nil {
			log.Printf(ErrSendMsg)
		}
		if task.byUser.ID != task.desUser.ID {
			pastMsg := tgbotapi.NewMessage(task.byUser.ID, "Задача "+"\""+task.Name+"\""+" назначена на @"+task.desUser.Name)
			_, err = bot.Send(pastMsg)
			if err != nil {
				log.Printf(ErrSendMsg)
			}
		}
		api.mu.Unlock()
		return
	}
	pastUserMsg := tgbotapi.NewMessage(task.desUser.ID, "Задача "+"\""+task.Name+"\""+" назначена на @"+api.Users[userID].Name)
	_, err := bot.Send(pastUserMsg)
	if err != nil {
		log.Printf(ErrSendMsg)
	}
	currentUserMsg := tgbotapi.NewMessage(userID, "Задача "+"\""+task.Name+"\""+" назначена на вас")
	api.mu.Lock()
	task.desUser = *api.Users[userID]
	api.mu.Unlock()

	_, err = bot.Send(currentUserMsg)
	if err != nil {
		log.Printf(ErrSendMsg)
	}
}

func (api *Handler) UnAssign(bot *tgbotapi.BotAPI, userID int64, taskID int) {

	task := api.allTasks[taskID]

	if task.desUser.ID != userID {
		_, err := bot.Send(tgbotapi.NewMessage(userID, "Задача не на вас"))
		if err != nil {
			log.Printf(ErrSendMsg)
		}
		return
	}
	_, err := bot.Send(tgbotapi.NewMessage(userID, "Принято"))
	if err != nil {
		log.Printf(ErrSendMsg)
	}
	if task.byUser.ID != task.desUser.ID {
		_, err := bot.Send(tgbotapi.NewMessage(task.byUser.ID, "Задача "+"\""+task.Name+"\""+" осталась без исполнителя"))
		if err != nil {
			log.Printf(ErrSendMsg)
		}
	}

	api.mu.Lock()
	task.desUser = User{Name: "", ID: 0}
	tmpSl := make([]int, 0)
	for _, val := range api.Users[userID].MyTaskIDs {
		if val != taskID {
			tmpSl = append(tmpSl, val)
		}
	}

	api.Users[userID].MyTaskIDs = make([]int, len(tmpSl))
	api.Users[userID].MyTaskIDs = tmpSl
	api.mu.Unlock()
}

func (api *Handler) Resolve(bot *tgbotapi.BotAPI, userID int64, taskID int) {

	task := api.allTasks[taskID]

	if task.desUser.ID != userID {
		_, err := bot.Send(tgbotapi.NewMessage(userID, "Задача не на вас"))
		if err != nil {
			log.Printf(ErrSendMsg)
		}
		return
	}
	_, err := bot.Send(tgbotapi.NewMessage(userID, "Задача "+"\""+task.Name+"\""+" выполнена"))
	if err != nil {
		log.Printf(ErrSendMsg)
	}

	if task.byUser.ID != task.desUser.ID {
		_, err = bot.Send(tgbotapi.NewMessage(task.byUser.ID, "Задача "+"\""+task.Name+"\""+" выполнена @"+task.desUser.Name))
		if err != nil {
			log.Printf(ErrSendMsg)
		}
	}
	api.mu.Lock()
	api.allTasks[taskID].status = false
	api.mu.Unlock()
}

func (api *Handler) MyTasks(bot *tgbotapi.BotAPI, userID int64) {

	msg := ""

	for _, taskID := range api.Users[userID].MyTaskIDs {
		if api.allTasks[taskID].status {
			api.mu.Lock()
			msg += strconv.Itoa(taskID) + ". " + api.allTasks[taskID].Name + " by @" + api.allTasks[taskID].byUser.Name
			msg += "\n/unassign_" + strconv.Itoa(taskID) + " /resolve_" + strconv.Itoa(taskID) + "\n"
			api.mu.Unlock()
		}
	}

	if len(msg) != 0 {
		api.mu.Lock()
		msg = msg[:len(msg)-1]
		api.mu.Unlock()

	}
	_, err := bot.Send(tgbotapi.NewMessage(userID, msg))
	if err != nil {
		log.Printf(ErrSendMsg)
	}
}

func (api *Handler) OwnerTasks(bot *tgbotapi.BotAPI, userID int64) {

	msg := ""

	for _, taskID := range api.Users[userID].OwnerTaskIDs {
		if api.allTasks[taskID].status {
			api.mu.Lock()
			msg += strconv.Itoa(taskID) + ". " + api.allTasks[taskID].Name + " by @" + api.allTasks[taskID].byUser.Name
			msg += "\n" + "/assign_" + strconv.Itoa(taskID) + "\n"
			api.mu.Unlock()
		}
	}

	if len(msg) != 0 {
		api.mu.Lock()
		msg = msg[:len(msg)-1]
		api.mu.Unlock()

	}
	_, err := bot.Send(tgbotapi.NewMessage(userID, msg))
	if err != nil {
		log.Printf(ErrSendMsg)
	}
}

func (api *Handler) RunCommand(bot *tgbotapi.BotAPI, msg string, userID int64, cmd func(bot *tgbotapi.BotAPI, userID int64, taskID int)) {
	ID := strings.Split(msg, "_")
	if len(ID) < 2 {
		return
	}
	taskID, err := strconv.Atoi(ID[1])
	if err != nil {
		_, errSend := bot.Send(tgbotapi.NewMessage(userID, "Неверное ID"))
		if errSend != nil {
			log.Printf(ErrSendMsg)
		}
		return
	}
	if taskID > 0 && taskID <= api.taskCount {
		cmd(bot, userID, taskID)
	}
}
