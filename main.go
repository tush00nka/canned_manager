package main

import (
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
	ExpiredTasksNumber   uint
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
			handleMessages(bot, update.Message, &states, &newDescriptions, db)
		} else if update.CallbackQuery != nil {
			handleCallbacks(bot, update.CallbackQuery, db)
		}
	}
}

func handleMessages(
	bot *tg.BotAPI,
	message *tg.Message,
	states *map[uint]userState,
	newDescriptions *map[uint]string,
	db *gorm.DB) {

	userID := uint(message.From.ID)
	var user User
	db.Where(&User{ID: userID}).FirstOrCreate(&user)
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	var msg tg.MessageConfig

	switch (*states)[userID] {
	case OVERVIEW:
		msg = handleOverview(message, states, db)
	case NEW_TASK:
		msg = handleNewTask(message, states, newDescriptions, db)
	default:
		msg = tg.NewMessage(message.Chat.ID, "Unknown state")
	}

	bot.Send(msg)
}

func handleCallbacks(bot *tg.BotAPI, callback_query *tg.CallbackQuery, db *gorm.DB) {
	callback := tg.NewCallback(callback_query.ID, callback_query.Data)
	if _, err := bot.Request(callback); err != nil {
		log.Fatal(err)
	}
	callback_data := strings.Split(callback_query.Data, "_")
	var reply tg.MessageConfig
	var message = callback_query.Message
	switch callback_data[0] {
	case "delete":
		taskID, err := strconv.Atoi(callback_data[1])
		if err != nil {
			log.Fatal(err)
		}
		reply = delete_task(message, uint(taskID), db)
	case "complete":
		taskID, err := strconv.Atoi(callback_data[1])
		if err != nil {
			log.Fatal(err)
		}
		reply = complete_task(message, uint(taskID), db)
	case "cancel":
		reply = tg.NewMessage(message.Chat.ID, "Действие отменено")
	default:
		reply = tg.NewMessage(message.Chat.ID, "Unknown Callback")
	}

	edit := tg.NewEditMessageReplyMarkup(message.Chat.ID, message.MessageID,
		tg.InlineKeyboardMarkup{
			InlineKeyboard: make([][]tg.InlineKeyboardButton, 0),
		})

	bot.Send(edit)
	bot.Send(reply)
}
