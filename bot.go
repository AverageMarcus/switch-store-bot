package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type Subscription struct {
	User      int64
	Game      string
	GameTitle string
}

var subscriptions []Subscription

func InitBot(debug bool) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		panic(err)
	}

	bot.Debug = debug

	go checkForUpdates(bot)

	return bot
}

var messages []string = []string{
	"Hey! Great news '%s' is on sale for £%g. That's %g%% off! Wow!\n\nFind out more at https://www.nintendo.co.uk%s",
	"Have you heard? '%s' is on sale for £%g with %g%% off! How great is that?!\n\nFind out more at https://www.nintendo.co.uk%s",
	"Woo! '%s' is on sale. Now £%g at %g%% off!\n\nFind out more at https://www.nintendo.co.uk%s",
}

func PostMessage(bot *tgbotapi.BotAPI, game Game) {
	rand.Seed(time.Now().Unix())
	message := fmt.Sprintf(messages[rand.Intn(len(messages))], game.Title, game.Price, game.PriceDiscountPercentage, game.URL)
	for _, subscription := range subscriptions {
		if subscription.Game == game.ID || subscription.Game == "*" {
			msg := tgbotapi.NewMessage(subscription.User, message)
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
		}
	}
}

func filterUser(userID int64) func(Subscription) bool {
	return func(el Subscription) bool {
		return el.User != userID
	}
}

func filterUserAndGame(userID int64, gameId string) func(Subscription) bool {
	return func(el Subscription) bool {
		return el.User != userID || (el.User == userID && el.Game != gameId)
	}
}

func choose(ss []Subscription, test func(Subscription) bool) (ret []Subscription) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func checkForUpdates(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil || !update.Message.IsCommand() {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		switch update.Message.Command() {
		case "start":
			msg = buildStartMessage(msg)
		case "all":
			msg = buildAllMessage(msg, update.Message.Chat.ID)
		case "stop":
			msg = buildStopMessage(msg, update.Message.Chat.ID)
		case "watch":
			msg = buildWatchMessage(msg, update.Message.Chat.ID, update.Message.CommandArguments())
		case "unwatch":
			msg = buildUnwatchMessage(msg, update.Message.Chat.ID, update.Message.CommandArguments())
		case "list":
			msg = buildListMessage(msg, update.Message.Chat.ID)
		case "help":
			msg = buildHelpMessage(msg)
		default:
			msg.Text = "Sorry, I don't know that command"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func buildStartMessage(msg tgbotapi.MessageConfig) tgbotapi.MessageConfig {
	msg.Text = "Hello. I'm able to notify you when Nintendo Switch games go on sale. To find out what I can do use the /help command'"
	return msg
}

func buildAllMessage(msg tgbotapi.MessageConfig, userID int64) tgbotapi.MessageConfig {
	for _, sub := range subscriptions {
		if sub.User == userID && sub.Game == "*" {
			msg.Text = "Hey, it looks like you're already signed up to get notifications about all games"
			return msg
		}
	}
	subscriptions = append(subscriptions, Subscription{
		User: userID,
		Game: "*",
	})
	msg.Text = "Hi there. You'll now get notifications when any Nintendo Switch game goes on sale"
	return msg
}

func buildStopMessage(msg tgbotapi.MessageConfig, userID int64) tgbotapi.MessageConfig {
	subscriptions = choose(subscriptions, filterUser(userID))
	msg.Text = "Sure thing! We'll no longer send you notifications when games go on sale"
	return msg
}

func buildWatchMessage(msg tgbotapi.MessageConfig, userID int64, query string) tgbotapi.MessageConfig {
	games := FindGame(query)

	var keyboard = [][]tgbotapi.KeyboardButton{}
	for _, game := range games {
		if game.Title == query {
			for _, sub := range subscriptions {
				if sub.User == userID && sub.Game == game.Title {
					msg.Text = "Hey, it looks like you're already signed up to get notifications about this game"
					return msg
				}
			}

			subscriptions = append(subscriptions, Subscription{
				User:      userID,
				Game:      game.ID,
				GameTitle: game.Title,
			})
			msg.Text = "You got it! We'll notify you when '" + game.Title + "' goes on sale"
			return msg
		}

		keyboard = append(keyboard, []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("/watch " + game.Title)})
	}

	msg.Text = "Which game is it you want to watch?"
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        keyboard,
		OneTimeKeyboard: true,
	}
	return msg
}

func buildUnwatchMessage(msg tgbotapi.MessageConfig, userID int64, query string) tgbotapi.MessageConfig {
	games := FindGame(query)

	var keyboard = [][]tgbotapi.KeyboardButton{}
	for _, game := range games {
		if game.Title == query {
			subscriptions = choose(subscriptions, filterUserAndGame(userID, game.ID))
			msg.Text = "No problem! We'll no longer notify you when '" + game.Title + "' goes on sale"
			return msg
		}

		keyboard = append(keyboard, []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("/unwatch " + game.Title)})
	}

	msg.Text = "Which game is it you want to unwatch?"
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:        keyboard,
		OneTimeKeyboard: true,
	}
	return msg
}

func buildHelpMessage(msg tgbotapi.MessageConfig) tgbotapi.MessageConfig {
	msg.Text = `
/all - Get notifications for all Switch games that go on sale
/stop - Turn off notifications when Switch games go on sale
/watch [title] - Get notifications for a specific Switch game
/unwatch [title] - Turn off notifications for a specific Switch game
/list - Show a list of all games I will get notifications for
	`
	return msg
}

func buildListMessage(msg tgbotapi.MessageConfig, userID int64) tgbotapi.MessageConfig {
	var titles []string = []string{}
	for _, sub := range subscriptions {
		if sub.Game != "*" {
			titles = append(titles, sub.GameTitle)
		}
	}
	if len(titles) > 0 {
		msg.Text = "We will notify you when the following games go on sale:\n"
		for _, title := range titles {
			msg.Text += "- " + title + "\n"
		}
	} else {
		msg.Text = "You are not currently setup to be notified for any games"
	}
	return msg
}
