package main

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func setupLog(logFile string) zerolog.Logger {
	err := os.MkdirAll(filepath.Dir(logFile), 0700)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed creating 'log' directory!")
		os.Exit(1)
	}

	// log to file
	fileLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,    //days
		Compress:   false, // disabled by default
	}
	var logWriter io.Writer
	logWriter = fileLogger
	log := zerolog.New(logWriter).With().Timestamp().Logger()
	zerolog.TimestampFieldName = "T"
	zerolog.LevelFieldName = "L"
	zerolog.MessageFieldName = "M"
	zerolog.ErrorFieldName = "ERR"

	// log to console
	// log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	return log
}

func loadConfig(configFile string) *AppConfig {
	if _, err := os.Stat(configFile); err != nil {
		log.Fatal().Err(err).Msgf("Configuration file '%s' not found!?", configFile)
		os.Exit(1)
	}
	cfg, err := new(AppConfig).Parse(configFile)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed parsing configuration file!")
		os.Exit(1)
	}
	log.Info().Msg("Configuration loaded.")
	return cfg
}

func main() {
	const configFile string = "gukguk.yml"
	const appVersion string = "1.0.0"
	const logFile = "log/gukguk.log"

	logger := setupLog(logFile)
	logger.Info().Msgf("***** gukguk %s *****", appVersion)

	cfg := loadConfig(configFile)
	if len(cfg.Watch) == 0 {
		logger.Fatal().Msg("Entity not found for monitoring! Please check your 'watch' config!")
		os.Exit(1)
	}

	// configure logging
	switch cfg.LogLevel {
	case "off":
		zerolog.SetGlobalLevel(zerolog.NoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	}

	// channel for notification
	notifChannel := make(chan string)

	// manage checkers
	manager := NewManager().SetConfig(cfg).SetLog(&logger).SetChannel(notifChannel)

	// process & send notification
	alerter := new(Alerter).SetConfig(cfg).SetLog(&logger)
	go alerter.Process(notifChannel)

	// loop
	ticker := time.NewTicker(time.Duration(cfg.CheckInterval) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				manager.Check()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	logger.Info().Msg("Running...")

	//block forever
	select {}
}
