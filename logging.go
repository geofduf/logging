package logging

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	Fatal = iota
	System
	Error
	Warning
	Info
	Debug
	defaultLevel int = Warning
)

var logLevels [6]string = [6]string{"FTL", "SYS", "ERR", "WNG", "INF", "DBG"}

type logger struct {
	mu     sync.RWMutex
	level  int
	listen bool
}

func New() *logger {
	return &logger{level: defaultLevel}
}

func (l *logger) GetLevel() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

func (l *logger) SetLevel(level int) {
	if level < 1 || level >= len(logLevels) {
		l.System("LOG", fmt.Sprintf("cannot set log level to %d", level))
	} else {
		l.mu.Lock()
		l.level = level
		l.mu.Unlock()
		l.System("LOG", fmt.Sprintf("setting log level to %d (%s)", level, logLevels[level]))
	}
}

func (l *logger) Fatal(source string, messages ...string) {
	if len(messages) > 1 {
		l.write(Fatal, source, messages[:len(messages)-1])
	}
	log.Fatalf("[%s] [%s] %s\n", logLevels[Fatal], source, messages[len(messages)-1])
}

func (l *logger) System(source string, messages ...string) {
	l.write(System, source, messages)
}

func (l *logger) Error(source string, messages ...string) {
	l.write(Error, source, messages)
}

func (l *logger) Warning(source string, messages ...string) {
	l.write(Warning, source, messages)
}

func (l *logger) Info(source string, messages ...string) {
	l.write(Info, source, messages)
}

func (l *logger) Debug(source string, messages ...string) {
	l.write(Debug, source, messages)
}

func (l *logger) write(level int, source string, messages []string) {
	if level <= System || l.GetLevel() >= level {
		for _, message := range messages {
			log.Printf("[%s] [%s] %s\n", logLevels[level], source, message)
		}
	}
}

func (l *logger) ListenForSignal() {
	var message string
	l.mu.Lock()
	if l.listen {
		message = "cannot register signal handler more than once"
	} else {
		go l.signalHandler()
		l.listen = true
		message = "registering signal handler"
	}
	l.mu.Unlock()
	l.System("LOG", message)
}

func (l *logger) signalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		s := <-c
		currentLevel := l.GetLevel()
		switch s {
		case syscall.SIGUSR1:
			l.SetLevel(currentLevel + 1)
		case syscall.SIGUSR2:
			l.SetLevel(currentLevel - 1)
		}
	}
}
