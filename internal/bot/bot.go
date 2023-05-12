package bot

import (
	"fmt"
	"github.com/and3rson/telemux/v2"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"markoslav/internal/bot/handler"
	"markoslav/internal/config"
	"time"
)

type Bot struct {
	API *tgbotapi.BotAPI
	Mux *telemux.Mux
}

func New(conf config.Bot) *Bot {
	api, err := tgbotapi.NewBotAPI(conf.Token)
	if err != nil {
		log.Fatal(err)
	}

	if conf.Debug {
		api.Debug = true
	}

	return &Bot{
		API: api,
		Mux: telemux.NewMux(),
	}
}

func (bot *Bot) Handle(captionHandler *handler.CaptionHandler) *Bot {
	captionHandler.Register(bot.Mux)

	return bot
}

func (bot *Bot) Run() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updatesChannel := bot.API.GetUpdatesChan(updateConfig)

	fmt.Println("clear updates")

	bot.clearUpdatesChannel(updatesChannel)

	fmt.Println("bot started")

	for update := range updatesChannel {
		bot.Mux.Dispatch(bot.API, update)
	}
}

func (bot *Bot) clearUpdatesChannel(updatesChannel tgbotapi.UpdatesChannel) {
	timer := time.NewTimer(5 * time.Second)

	for {
		select {
		case <-updatesChannel:
			timer.Reset(5 * time.Second)
		case <-timer.C:
			return
		}
	}
}
