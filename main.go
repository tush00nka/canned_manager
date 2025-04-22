package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	ID                   uint
	CompletedTasksNumber uint
	Tasks                []Task `gorm:"foreignKey:UserID"`
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

	bot, err := tg.NewBotAPI(os.Getenv("TELEGRAM_BOT_API_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	updateConfig := tg.NewUpdate(0)
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

			var msg tg.MessageConfig

			switch states[userID] {
			case OVERVIEW:
				msg = handleOverview(update.Message, &states, db)
			case NEW_TASK:
				msg = handleNewTask(update.Message, &states, &newDescriptions, db)
			default:
				msg = tg.NewMessage(update.Message.Chat.ID, "Unknown state")
			}

			bot.Send(msg)

		} else if update.CallbackQuery != nil {
			callback := tg.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				log.Fatal(err)
			}
			callback_data := strings.Split(update.CallbackQuery.Data, "_")
			var msg tg.MessageConfig
			switch callback_data[0] {
			case "delete":
				taskID, err := strconv.Atoi(callback_data[1])
				if err != nil {
					log.Fatal(err)
				}
				msg = delete_task(update.CallbackQuery.Message, uint(taskID), db)
			case "complete":
				taskID, err := strconv.Atoi(callback_data[1])
				if err != nil {
					log.Fatal(err)
				}
				msg = complete_task(update.CallbackQuery.Message, uint(taskID), db)
			case "cancel":
				msg = tg.NewMessage(update.CallbackQuery.Message.Chat.ID, "Действие отменено")
			default:
				msg = tg.NewMessage(update.CallbackQuery.Message.Chat.ID, "Unknown Callback")
			}

			edit := tg.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID,
				tg.InlineKeyboardMarkup{
					InlineKeyboard: make([][]tg.InlineKeyboardButton, 0),
				})

			bot.Send(edit)
			bot.Send(msg)
		}
	}
}

func handleOverview(message *tg.Message, states *map[uint]userState, db *gorm.DB) (msg tg.MessageConfig) {
	switch message.Command() {
	case "start":
		msg = start(message)

	case "new_task":
		msg = add_task(message)
		(*states)[uint(message.From.ID)] = NEW_TASK

	case "list":
		msg = list_tasks(message, db)

	case "delete":
		msg = select_task_to_delete(message, db)

	case "complete":
		msg = select_task_to_complete(message, db)

	case "stats":
		var user User
		db.First(&user, message.From.ID)
		msg = tg.NewMessage(message.Chat.ID,
			"Статистика:\n\n"+
				"Всего выполнено задач: "+fmt.Sprint(user.CompletedTasksNumber))

	case "help":
		msg = tg.NewMessage(message.Chat.ID,
			"Вот что я умею:\n"+
				"/start - Запуск бота\n"+
				"/new_task - Добавить задачу\n"+
				"/list - Отобразить список задач\n"+
				"/delete - Удалить задачу\n"+
				"/help - Отобразить эту справку\n")
	default:
		msg = tg.NewMessage(message.Chat.ID, "I don't know that command")
	}
	return
}
