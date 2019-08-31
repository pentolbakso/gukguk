package main

import (
	"github.com/rs/zerolog"
)

type Manager struct {
	logger   *zerolog.Logger
	cfg      *AppConfig
	channel  chan string
	checkers map[int]*Checker
}

func NewManager() *Manager {
	return &Manager{
		checkers: make(map[int]*Checker),
	}
}

func (m *Manager) SetConfig(c *AppConfig) *Manager   { m.cfg = c; return m }
func (m *Manager) SetLog(l *zerolog.Logger) *Manager { m.logger = l; return m }
func (m *Manager) SetChannel(c chan string) *Manager { m.channel = c; return m }

func (m *Manager) Check() {
	for _, task := range m.cfg.Watch {
		checker, found := m.checkers[task.ID]
		if !found {
			checker = NewChecker(task).SetLog(m.logger).SetNotifChannel(m.channel)
			m.checkers[task.ID] = checker
		}
		go checker.Start()
	}
}
