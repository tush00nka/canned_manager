package main

import (
	"fmt"
	"sort"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func start() (text string) {
	text = "Привет! Я - бот-менеджер. Я помогаю организовать задачи и напоминаю о дедлайнах.\n\nДля получения справки используйте команду /help"
	return
}

func list_tasks(message *tg.Message, db *gorm.DB) (text string) {
	userID := uint(message.From.ID)

	var tasks []Task
	db.Find(&tasks, Task{UserID: userID})

	if len(tasks) <= 0 {
		text = "У вас нет задач!"
		return
	}

	text = "Ваши задачи:\n\n"

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].DueTo.Before(tasks[j].DueTo)
	})

	for i, task := range tasks {
		text += fmt.Sprintf("%d. %s (%d.%d.%d)\n", i+1, task.Description, task.DueTo.Day(), task.DueTo.Month(), task.DueTo.Year())
	}

	return
}

func add_task() (text string) {
	text = "Опишите задачу"
	return
}

func select_task(message *tg.Message, db *gorm.DB, callback_name string) (text string, keyboard tg.InlineKeyboardMarkup) {
	userID := uint(message.From.ID)

	var tasks []Task
	db.Find(&tasks, Task{UserID: userID})

	if len(tasks) <= 0 {
		text = "У вас нет задач!"
		return
	}

	var goal string

	switch callback_name {
	case "delete":
		goal = "удаления"
	case "complete":
		goal = "выполнения"
	default:
		goal = "действия"
	}

	text = fmt.Sprintf("Выберите задачу для %s", goal)

	keyboard = tg.NewInlineKeyboardMarkup()

	for _, task := range tasks {
		keyboard.InlineKeyboard =
			append(keyboard.InlineKeyboard,
				tg.NewInlineKeyboardRow(tg.NewInlineKeyboardButtonData(task.Description, fmt.Sprintf("%s_%d", callback_name, task.ID))))
	}

	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard,
		tg.NewInlineKeyboardRow(tg.NewInlineKeyboardButtonData("Отмена", "cancel")))

	return
}

func delete_task(taskID uint, db *gorm.DB) (text string) {
	text = "Задача удалена!"

	var task Task
	db.First(&task, taskID)
	db.Delete(&task)

	return
}

func complete_task(taskID uint, db *gorm.DB) (text string) {

	var task Task
	var user User
	db.First(&task, taskID)
	db.First(&user, task.UserID)
	user.CompletedTasksNumber++
	db.Save(&user)
	db.Delete(&task)

	text = fmt.Sprintf("Задача Выполнена!\nВсего выполнено задач: %d", user.CompletedTasksNumber)

	return
}
