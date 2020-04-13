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
	FATAL = iota
	SYSTEM
	ERROR
	WARNING
	INFO
	DEBUG
	defaultLevel int = WARNING
)

var logLevels [6]string = [6]string{"FTL", "SYS", "ERR", "WNG", "INF", "DBG"}

var logger *Logger

func GetLogger() *Logger {
	return logger
}

type Logger struct {
	sync.RWMutex
	level int
}

func (l *Logger) FATAL(source string, messages ...string) {
	if len(messages) > 1 {
		l.write(FATAL, source, messages[:len(messages)-1])
	}
	log.Fatalf("[%s] [%s] %s\n", logLevels[FATAL], source, messages[len(messages)-1])
}

func (l *Logger) SYSTEM(source string, messages ...string) {
	l.write(SYSTEM, source, messages)
}

func (l *Logger) ERROR(source string, messages ...string) {
	l.write(ERROR, source, messages)
}

func (l *Logger) WARNING(source string, messages ...string) {
	l.write(WARNING, source, messages)
}

func (l *Logger) INFO(source string, messages ...string) {
	l.write(INFO, source, messages)
}

func (l *Logger) DEBUG(source string, messages ...string) {
	l.write(DEBUG, source, messages)
}

func (l *Logger) SetLevel(level int) {
	if level < 1 || level >= len(logLevels) {
		l.SYSTEM("LOG", fmt.Sprintf("cannot set log level to %d", level))
	} else {
		l.Lock()
		l.level = level
		l.Unlock()
		l.SYSTEM("LOG", fmt.Sprintf("setting log level to %d (%s)", level, logLevels[level]))
	}
}

func (l *Logger) GetLevel() int {
	l.RLock()
	defer l.RUnlock()
	return l.level
}

func (l *Logger) write(level int, source string, messages []string) {
	if level <= SYSTEM || l.GetLevel() >= level {
		for _, message := range messages {
			log.Printf("[%s] [%s] %s\n", logLevels[level], source, message)
		}
	}
}

func (l *Logger) signalHandler() {
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

func init() {
	logger = new(Logger)
	logger.level = defaultLevel
	go logger.signalHandler()
}
