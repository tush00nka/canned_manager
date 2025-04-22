package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

const TASK_INSTRUCTION string = "Опишите задачу"
const DATE_INSTRUCTION string = "Введите предельную дату выполнения задачи\n(ДД.ММ.ГГ или ДД.ММ)"

func handleNewTask(
	message *tg.Message,
	states *map[uint]userState,
	descriptions *map[uint]string,
	db *gorm.DB) (msg tg.MessageConfig) {

	userID := uint(message.From.ID)

	if (*descriptions)[userID] == "" {
		if len(message.Photo) > 0 {
			msg = tg.NewMessage(message.Chat.ID,
				fmt.Sprintf("Классная картинка, но описание задачи должно содержать текст!\n\n%s", TASK_INSTRUCTION))
			return
		}

		if message.Text == "" {
			msg = tg.NewMessage(message.Chat.ID,
				fmt.Sprintf("Описание задачи должно содежрать текст!\n\n%s", TASK_INSTRUCTION))
			return
		}

		(*descriptions)[userID] = message.Text
		msg = tg.NewMessage(message.Chat.ID, DATE_INSTRUCTION)
	} else {
		var ok bool
		msg, ok = getDate(message, descriptions, db)

		if !ok {
			return
		}

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
		if len(strings.Split(fmt.Sprint(year), "")) <= 3 {
			year += 2000
		}
	} else {
		year = time.Now().Year()

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

func getDate(message *tg.Message, descriptions *map[uint]string, db *gorm.DB) (msg tg.MessageConfig, ok bool) {
	ok = true
	userID := uint(message.From.ID)
	parsed_date := strings.Split(message.Text, ".")

	msg = tg.NewMessage(message.Chat.ID,
		fmt.Sprintf("Неверный формат даты!\n\n%s", DATE_INSTRUCTION))

	dueTo, ok := newDueTo(&parsed_date)

	if !ok {
		return
	}

	msg = tg.NewMessage(message.Chat.ID,
		fmt.Sprintf("Нельзя создать задачу в прошлом!\n\n%s", DATE_INSTRUCTION))

	if dueTo.Year() < message.Time().Year() {
		ok = false
		return
	} else if dueTo.Month() < message.Time().Month() {
		ok = false
		return
	} else if dueTo.Day() < message.Time().Day() && dueTo.Month() == message.Time().Month() {
		ok = false
		return
	}

	var user User
	db.First(&user, message.From.ID)

	task := Task{Description: (*descriptions)[userID], DueTo: dueTo}
	user.Tasks = append(user.Tasks, task)
	db.Save(&user)

	msg = tg.NewMessage(message.Chat.ID,
		fmt.Sprintf("Задача \"%s\" создана!", task.Description))
	return
}
