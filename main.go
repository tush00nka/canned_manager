package main

import (
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
