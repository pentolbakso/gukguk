package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type Checker struct {
	Entity       Entity
	FailCounter  int
	IsRunning    bool
	EventTime    time.Time
	logger       *zerolog.Logger
	notifChannel chan string
}

func NewChecker(entity Entity) *Checker {
	return &Checker{
		Entity:      entity,
		FailCounter: 0,
		IsRunning:   true,
		EventTime:   time.Now(),
	}
}

func (m *Checker) SetLog(l *zerolog.Logger) *Checker      { m.logger = l; return m }
func (m *Checker) SetNotifChannel(c chan string) *Checker { m.notifChannel = c; return m }

func (m *Checker) Start() {
	var success bool
	//m.logger.Debug().Msgf("Start checking %d - %s", m.Entity.ID, m.Entity.Name)

	if m.Entity.Http.Url != "" {
		success, _ = m.checkHttp(m.Entity.Http.Url)
	} else if m.Entity.Process.Path != "" {
		success, _ = m.checkProcess(m.Entity.Process.Path)
	}

	if success {
		if !m.IsRunning {
			//is back running
			elapsed := time.Since(m.EventTime)
			m.EventTime = time.Now()
			m.IsRunning = true
			m.FailCounter = 0

			m.notifChannel <- fmt.Sprintf("Entity '%s' is UP! Previous downtime: %s", m.Entity.Name, elapsed)
		}

	} else {
		if m.IsRunning {
			//is down
			elapsed := time.Since(m.EventTime)
			m.EventTime = time.Now()
			m.IsRunning = false
			m.FailCounter++

			m.notifChannel <- fmt.Sprintf("Entity '%s' is DOWN! Previous uptime: %s", m.Entity.Name, elapsed)
		}
	}
}

func (m *Checker) checkHttp(url string) (bool, error) {
	resp, err := http.Get(url)
	if err != nil {
		m.logger.Debug().Err(err).Msgf("GET http failed: %s", url)
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		m.logger.Debug().Msgf("GET success - %s", url)
		return false, nil
	}
	m.logger.Debug().Msgf("GET failed - %d - %s", resp.StatusCode, url)
	return false, fmt.Errorf("Http status: %d ", resp.StatusCode)
}

func (m *Checker) checkProcess(path string) (bool, error) {
	return true, nil
}
