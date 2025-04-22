package main

import (
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

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
		msg = select_task(message, db, "delete")

	case "complete":
		msg = select_task(message, db, "complete")

	case "stats":
		var user User
		db.First(&user, message.From.ID)
		msg = tg.NewMessage(message.Chat.ID,
			"Статистика:\n\n"+
				fmt.Sprintf("Всего выполнено задач: %d\n", user.CompletedTasksNumber)+
				fmt.Sprintf("Всего просрочено задач: %d", user.ExpiredTasksNumber))

	case "help":
		msg = tg.NewMessage(message.Chat.ID,
			"Вот что я умею:\n"+
				"/start - Запуск бота\n"+
				"/new_task - Добавить задачу\n"+
				"/list - Отобразить список задач\n"+
				"/delete - Удалить задачу\n"+
				"/help - Отобразить эту справку\n")
	default:
		msg = tg.NewMessage(message.Chat.ID, "Я Вас не понимаю(")
	}
	return
}
