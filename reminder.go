package main

import (
	"time"

	gocron "github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func set_reminder(bot *tgbotapi.BotAPI, db *gorm.DB) {

	scheduler := gocron.NewScheduler(time.UTC)

	scheduler.Every(1).Day().At("6:00").Do(func() {
		var users []User
		db.Find(&users)

		for _, user := range users {
			var tasks []Task
			db.Find(&tasks, Task{UserID: user.ID})
			for _, task := range tasks {
				if task.DueTo.Day() == time.Now().Day() &&
					task.DueTo.Month() == time.Now().Month() &&
					task.DueTo.Year() == time.Now().Year() {
					msg := tgbotapi.NewMessage(int64(user.ID), "Сегодня последний день по задаче:\n"+task.Description)
					bot.Send(msg)
					continue
				}

				if task.DueTo.Day() == time.Now().Day()+1 &&
					task.DueTo.Month() == time.Now().Month() &&
					task.DueTo.Year() == time.Now().Year() {
					msg := tgbotapi.NewMessage(int64(user.ID), "Завтра последний день по задаче:\n"+task.Description)
					bot.Send(msg)
					continue
				}
			}
		}
	})

	scheduler.StartAsync()
}
