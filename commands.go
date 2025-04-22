package main

import (
	"fmt"
	"sort"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func start(message *tg.Message) (msg tg.MessageConfig) {
	msg = tg.NewMessage(message.Chat.ID, "Привет! Я - бот-менеджер. Я помогаю организовать задачи и напоминаю о дедлайнах.\n\nДля получения справки используйте команду /help")
	return
}

func list_tasks(message *tg.Message, db *gorm.DB) (msg tg.MessageConfig) {
	userID := uint(message.From.ID)

	var tasks []Task
	db.Find(&tasks, Task{UserID: userID})

	if len(tasks) <= 0 {
		msg = tg.NewMessage(message.Chat.ID, "У вас нет задач!")
		return
	}

	output := "Ваши задачи:\n\n"

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].DueTo.Before(tasks[j].DueTo)
	})

	for i, task := range tasks {
		output += fmt.Sprintf("%d. %s (%d.%d.%d)\n", i+1, task.Description, task.DueTo.Day(), task.DueTo.Month(), task.DueTo.Year())
	}

	msg = tg.NewMessage(message.Chat.ID, output)

	return
}

func add_task(message *tg.Message) (msg tg.MessageConfig) {
	msg = tg.NewMessage(message.Chat.ID, "Опишите задачу:")

	return
}

func select_task_to_delete(message *tg.Message, db *gorm.DB) (msg tg.MessageConfig) {
	userID := uint(message.From.ID)

	var tasks []Task
	db.Find(&tasks, Task{UserID: userID})

	if len(tasks) <= 0 {
		msg = tg.NewMessage(message.Chat.ID, "У вас нет задач!")
		return
	}

	msg = tg.NewMessage(message.Chat.ID, "Выберите задачу для удаления")

	keyboard := tg.NewInlineKeyboardMarkup()

	for _, task := range tasks {
		keyboard.InlineKeyboard =
			append(keyboard.InlineKeyboard,
				tg.NewInlineKeyboardRow(tg.NewInlineKeyboardButtonData(task.Description, "delete_"+fmt.Sprint(task.ID))))
	}

	msg.ReplyMarkup = keyboard

	return
}

func delete_task(message *tg.Message, taskID uint, db *gorm.DB) (msg tg.MessageConfig) {
	msg = tg.NewMessage(message.Chat.ID, "Задача удалена!")

	var task Task
	db.First(&task, taskID)
	db.Delete(&task)

	return
}

func select_task_to_complete(message *tg.Message, db *gorm.DB) (msg tg.MessageConfig) {
	userID := uint(message.From.ID)

	var tasks []Task
	db.Find(&tasks, Task{UserID: userID})

	if len(tasks) <= 0 {
		msg = tg.NewMessage(message.Chat.ID, "У вас нет задач!")
		return
	}

	msg = tg.NewMessage(message.Chat.ID, "Выберите задачу для выполнения")

	keyboard := tg.NewInlineKeyboardMarkup()

	for _, task := range tasks {
		keyboard.InlineKeyboard =
			append(keyboard.InlineKeyboard,
				tg.NewInlineKeyboardRow(tg.NewInlineKeyboardButtonData(task.Description, "complete_"+fmt.Sprint(task.ID))))
	}

	msg.ReplyMarkup = keyboard

	return
}

func complete_task(message *tg.Message, taskID uint, db *gorm.DB) (msg tg.MessageConfig) {

	var task Task
	var user User
	db.First(&task, taskID)
	db.First(&user, task.UserID)
	user.CompletedTasksNumber++
	db.Save(&user)
	db.Delete(&task)

	msg = tg.NewMessage(message.Chat.ID, "Задача Выполнена!\n Всего выполнено задач: "+fmt.Sprint(user.CompletedTasksNumber))

	return
}
