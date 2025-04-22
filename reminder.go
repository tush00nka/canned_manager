package main

import (
	"time"

	"slices"

	gocron "github.com/go-co-op/gocron"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func set_reminder(bot *tg.BotAPI, db *gorm.DB) {

	scheduler := gocron.NewScheduler(time.UTC)

	scheduler.Every(1).Day().At("6:00").Do(func() {
		var users []User
		db.Find(&users)

		for _, user := range users {
			var tasks []Task
			db.Find(&tasks, Task{UserID: user.ID})

			// var day_left_message string = "Сегодня последний день по задачам:\n"
			var expired_message string = "Срок выполнения задач истёк, задачи удалены:\n\n"

			for i, task := range tasks {
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

				if task.DueTo.Day() > time.Now().Day() &&
					task.DueTo.Month() == time.Now().Month() &&
					task.DueTo.Year() == time.Now().Year() {
					expired_message += task.Description + "\n"
					user.Tasks = slices.Delete(tasks, i, i+1)
					i--
				}
			}
			db.Save(&user)
			if len(tasks) != len(user.Tasks) {
				msg := tg.NewMessage(int64(user.ID), expired_message)
				bot.Send(msg)
				continue
			}
		}
	})

	scheduler.StartAsync()
}
