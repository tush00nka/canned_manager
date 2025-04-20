package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

func handleNewTask(
	message *tgbotapi.Message,
	states *map[uint]userState,
	descriptions *map[uint]string,
	db *gorm.DB) (msg tgbotapi.MessageConfig) {

	userID := uint(message.From.ID)

	if (*descriptions)[userID] == "" {
		(*descriptions)[userID] = message.Text
		msg = tgbotapi.NewMessage(message.Chat.ID, "Введите предельную дату выполнения задачи\n(ДД.ММ.ГГ или ДД.ММ)")
	} else {
		msg = getDate(message, descriptions, db)
		(*states)[userID] = OVERVIEW
		(*descriptions)[userID] = ""
	}

	return
}

func newDueTo(parsed_date *[]string) (dueTo time.Time, ok bool) {
	ok = true

	if len(*parsed_date) < 2 {
		ok = false
		return
	}

	var year int

	if len(*parsed_date) >= 3 {
		parsed_year, err := strconv.Atoi((*parsed_date)[2])
		if err != nil {
			ok = false
			return
		}
		year = parsed_year
	} else {
		year = time.Now().Year()
		if len(strings.Split(fmt.Sprint(year), "")) <= 3 {
			year += 2000
		}
	}

	month, err := strconv.Atoi((*parsed_date)[1])
	if err != nil || month <= 0 || month > 12 {
		ok = false
		return
	}
	day, err := strconv.Atoi((*parsed_date)[0])
	if err != nil || day <= 0 || day > 31 {
		ok = false
		return
	}

	dueTo = time.Date(year, time.Month(month), day,
		0, 0, 0, 0, time.UTC)

	return
}

func getDate(message *tgbotapi.Message, descriptions *map[uint]string, db *gorm.DB) (msg tgbotapi.MessageConfig) {
	userID := uint(message.From.ID)
	parsed_date := strings.Split(message.Text, ".")

	msg = tgbotapi.NewMessage(message.Chat.ID,
		"Неверный формат даты!\n\nВведите предельную дату выполнения задачи\n(ДД.ММ.ГГ или ДД.ММ)")

	dueTo, ok := newDueTo(&parsed_date)

	if !ok {
		return
	}

	msg = tgbotapi.NewMessage(message.Chat.ID,
		"Нельзя создать задачу в прошлом!\n\nВведите предельную дату выполнения задачи\n(ДД.ММ.ГГ или ДД.ММ)")

	if dueTo.Year() < message.Time().Year() {
		return
	} else if dueTo.Month() < message.Time().Month() {
		return
	} else if dueTo.Day() < message.Time().Day() && dueTo.Month() == message.Time().Month() {
		return
	}

	var user User
	db.First(&user, message.From.ID)

	task := Task{Description: (*descriptions)[userID], DueTo: dueTo}
	user.Tasks = append(user.Tasks, task)
	db.Save(&user)

	msg = tgbotapi.NewMessage(message.Chat.ID, "Задача создана!")
	return
}
