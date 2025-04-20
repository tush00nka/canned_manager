package main

import (
	"fmt"
	"sort"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func start() (msg string) {
	msg = "Привет! Я буду ебать тебя с дедлайнами)"
	return
}

func list_tasks(message *tgbotapi.Message, db *gorm.DB) (msg tgbotapi.MessageConfig) {
	userID := uint(message.From.ID)

	var tasks []Task
	db.Find(&tasks, Task{UserID: userID})

	if len(tasks) <= 0 {
		msg = tgbotapi.NewMessage(message.Chat.ID, "У вас нет задач!")
		return
	}

	output := "Ваши задачи:\n\n"

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].DueTo.Before(tasks[j].DueTo)
	})

	for i, task := range tasks {
		output += fmt.Sprintf("%d. %s (%d.%d.%d)\n", i+1, task.Description, task.DueTo.Day(), task.DueTo.Month(), task.DueTo.Year())
	}

	msg = tgbotapi.NewMessage(message.Chat.ID, output)

	return
}

func add_task(message *tgbotapi.Message) (msg tgbotapi.MessageConfig) {
	msg = tgbotapi.NewMessage(message.Chat.ID, "Опишите задачу:")

	return
}

func select_task_to_delete(message *tgbotapi.Message, db *gorm.DB) (msg tgbotapi.MessageConfig) {
	userID := uint(message.From.ID)

	var tasks []Task
	db.Find(&tasks, Task{UserID: userID})

	if len(tasks) <= 0 {
		msg = tgbotapi.NewMessage(message.Chat.ID, "У вас нет задач!")
		return
	}

	msg = tgbotapi.NewMessage(message.Chat.ID, "Выберите задачу для удаления:")

	keyboard := tgbotapi.NewInlineKeyboardMarkup()

	for _, task := range tasks {
		keyboard.InlineKeyboard =
			append(keyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(task.Description, "delete_"+fmt.Sprint(task.ID))))
	}

	msg.ReplyMarkup = keyboard

	return
}

func delete_task(message *tgbotapi.Message, taskID uint, db *gorm.DB) (msg tgbotapi.MessageConfig) {
	msg = tgbotapi.NewMessage(message.Chat.ID, "Задача удалена!")

	var task Task
	db.First(&task, taskID)
	db.Delete(&task)

	return
}
