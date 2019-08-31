package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rs/zerolog"
)

type Alerter struct {
	logger *zerolog.Logger
	cfg    *AppConfig
}

func (m *Alerter) SetConfig(c *AppConfig) *Alerter   { m.cfg = c; return m }
func (m *Alerter) SetLog(l *zerolog.Logger) *Alerter { m.logger = l; return m }

func (m *Alerter) Process(c chan string) {
	message := <-c
	m.logger.Info().Msgf("Send alert => %s", message)

	if m.cfg.Notify.Telegram.AccessToken != "" {
		m.sendTelegram(message)
	}
}

type TelegramApiAnswer struct {
	Ok     bool              `json:"ok"`
	Result TelegramApiResult `json:"result"`
}

type TelegramApiResult struct {
	MessageId int `json:"message_id"`
}

func (m *Alerter) sendTelegram(message string) {
	result := TelegramApiAnswer{
		Result: TelegramApiResult{},
	}

	encodedMsg := url.QueryEscape(message)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s",
		m.cfg.Notify.Telegram.AccessToken, m.cfg.Notify.Telegram.ChannelId, encodedMsg)

	resp, err := http.Get(url)
	if err != nil {
		m.logger.Error().Err(err).Msg("Send telegram message failed")
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		m.logger.Error().Err(err).Msg("Decode telegram's json response failed")
		return
	}
	if !result.Ok {
		m.logger.Error().Err(err).Msg("Telegram resp not OK => failed")
		return
	}
	m.logger.Debug().Msg("Send telegram message success")
}
