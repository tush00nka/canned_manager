package main

import (
	"time"

	"slices"

	gocron "github.com/go-co-op/gocron"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func set_schedule(bot *tg.BotAPI, db *gorm.DB, at string,
	function func(bot *tg.BotAPI, db *gorm.DB, user *User, tasks *[]Task)) {
	scheduler := gocron.NewScheduler(time.Local)
	scheduler.Every(1).Day().At(at).Do(func() {
		var users []User
		db.Find(&users)

		for _, user := range users {
			var tasks []Task
			db.Find(&tasks, Task{UserID: user.ID})
			function(bot, db, &user, &tasks)
		}
	})
	scheduler.StartAsync()
}

// неиспользуемый указатель на БД-шку для совместимости с типом функции,
// которую передаём в set_schedule
func remind(bot *tg.BotAPI, _ *gorm.DB, user *User, tasks *[]Task) { // todo: переписать, чтобы возвращало текст со всеми задачами, как одно сообщение
	for _, task := range *tasks {
		if task.DueTo.Day() == time.Now().Day() &&
			task.DueTo.Month() == time.Now().Month() &&
			task.DueTo.Year() == time.Now().Year() {
			msg := tg.NewMessage(int64(user.ID), "Сегодня последний день по задаче:\n\n"+task.Description)
			bot.Send(msg)
			continue
		}

		if task.DueTo.Day() == time.Now().Day()+1 &&
			task.DueTo.Month() == time.Now().Month() &&
			task.DueTo.Year() == time.Now().Year() {
			msg := tg.NewMessage(int64(user.ID), "Завтра последний день по задаче:\n\n"+task.Description)
			bot.Send(msg)
			continue
		}
	}
}

func expire(bot *tg.BotAPI, db *gorm.DB, user *User, tasks *[]Task) {
	var expired_message string = "Срок выполнения задач истёк, задачи удалены:\n\n"

	for i, task := range *tasks {
		if task.DueTo.Day() > time.Now().Day() &&
			task.DueTo.Month() == time.Now().Month() &&
			task.DueTo.Year() == time.Now().Year() {
			expired_message += task.Description + "\n"
			user.Tasks = slices.Delete(*tasks, i, i+1)
			user.ExpiredTasksNumber++
			i--
		}
	}
	db.Save(&user)
	if len(*tasks) != len(user.Tasks) {
		msg := tg.NewMessage(int64(user.ID), expired_message)
		bot.Send(msg)
	}
}
