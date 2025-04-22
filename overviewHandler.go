package main

import (
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func handleOverview(message *tg.Message, states *map[uint]userState, db *gorm.DB) (msg tg.MessageConfig) {
	msg.ChatID = message.Chat.ID

	switch message.Command() {
	case "start":
		msg.Text = start()

	case "new_task":
		msg.Text = add_task()
		(*states)[uint(message.From.ID)] = NEW_TASK

	case "list":
		msg.Text = list_tasks(message, db)

	case "delete":
		msg.Text, msg.ReplyMarkup = select_task(message, db, "delete")

	case "complete":
		msg.Text, msg.ReplyMarkup = select_task(message, db, "complete")

	case "stats":
		var user User
		db.First(&user, message.From.ID)
		msg.Text = "Статистика:\n\n" +
			fmt.Sprintf("Всего выполнено задач: %d\n", user.CompletedTasksNumber) +
			fmt.Sprintf("Всего просрочено задач: %d", user.ExpiredTasksNumber)

	case "help":
		msg.Text = "Вот что я умею:\n" +
			"/start - Запуск бота\n" +
			"/new_task - Добавить задачу\n" +
			"/list - Отобразить список задач\n" +
			"/delete - Удалить задачу\n" +
			"/help - Отобразить эту справку\n"
	default:
		msg.Text = "Я Вас не понимаю("
	}
	return
}
