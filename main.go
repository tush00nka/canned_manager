package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type userState uint

const (
	OVERVIEW userState = iota
	NEW_TASK
	EDIT_TASK
)

func connectToDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("cannedManager.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	return db
}

type User struct {
	gorm.Model
	ID    uint
	Tasks []Task `gorm:foreignKey:UserID`
}

type Task struct {
	gorm.Model
	UserID      uint
	Description string
	DueTo       time.Time
}

func main() {
	db := connectToDB()
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Task{})

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_API_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	set_reminder(bot, db)

	var states map[uint]userState = make(map[uint]userState)
	var newDescriptions map[uint]string = make(map[uint]string)

	for update := range updates {
		if update.Message != nil {
			userID := uint(update.Message.From.ID)
			var user User
			db.Where(&User{ID: userID}).FirstOrCreate(&user)
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			var msg tgbotapi.MessageConfig

			switch states[userID] {
			case OVERVIEW:
				msg = handleOverview(update.Message, &states, db)
			case NEW_TASK:
				msg = handleNewTask(update.Message, &states, &newDescriptions, db)
			default:
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown state")
			}

			bot.Send(msg)

		} else if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				log.Fatal(err)
			}
			callback_data := strings.Split(update.CallbackQuery.Data, "_")
			var msg tgbotapi.MessageConfig
			switch callback_data[0] {
			case "delete":
				taskID, err := strconv.Atoi(callback_data[1])
				if err != nil {
					log.Fatal(err)
				}
				msg = delete_task(update.CallbackQuery.Message, uint(taskID), db)
			default:
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown Callback")
			}

			bot.Send(msg)
		}
	}
}

func handleOverview(message *tgbotapi.Message, states *map[uint]userState, db *gorm.DB) (msg tgbotapi.MessageConfig) {
	switch message.Command() {
	// case "start":
	// 	msg = start()

	case "new_task":
		msg = add_task(message)
		(*states)[uint(message.From.ID)] = NEW_TASK

	case "list":
		msg = list_tasks(message, db)

	case "delete":
		msg = select_task_to_delete(message, db)

	case "help":
		msg = tgbotapi.NewMessage(message.Chat.ID, "I can help you with the following commands:\n/start - Start the bot\n/help - Display this help message")
	default:
		msg = tgbotapi.NewMessage(message.Chat.ID, "I don't know that command")
	}
	return
}

func handleNewTask(
	message *tgbotapi.Message,
	states *map[uint]userState,
	descriptions *map[uint]string,
	db *gorm.DB) (msg tgbotapi.MessageConfig) {
	userID := uint(message.From.ID)

	if (*descriptions)[userID] == "" {
		(*descriptions)[userID] = message.Text
		msg = tgbotapi.NewMessage(message.Chat.ID, "Введите предельную дату выполнения задачи\n(ДД.ММ.ГГ или ДД.ММ)")
	} else {
		parsed_date := strings.Split(message.Text, ".")

		var year, month, day int

		if len(parsed_date) >= 3 {
			year, _ = strconv.Atoi(parsed_date[2])
		} else {
			year = message.Time().Year()
			if len(strings.Split(fmt.Sprint(year), "")) <= 3 {
				year += 2000
			}
		}

		month, _ = strconv.Atoi(parsed_date[1])
		day, _ = strconv.Atoi(parsed_date[0])

		newDueTo := time.Date(
			year,
			time.Month(month),
			day,
			0, 0, 0, 0, time.UTC)

		if year < time.Now().Year() ||
			month < int(time.Now().Month()) ||
			day < int(time.Now().Day()) {
			msg = tgbotapi.NewMessage(message.Chat.ID,
				"Нельзя создать задачу в прошлом!\n\nВведите предельную дату выполнения задачи\n(ДД.ММ.ГГ или ДД.ММ)")
			return
		}

		var user User
		db.First(&user, message.From.ID)

		task := Task{Description: (*descriptions)[userID], DueTo: newDueTo}
		user.Tasks = append(user.Tasks, task)
		db.Save(&user)

		(*states)[userID] = OVERVIEW
		(*descriptions)[userID] = ""
		msg = tgbotapi.NewMessage(message.Chat.ID, "Задача создана!")
	}

	return
}
