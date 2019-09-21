package main

import (
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"mrhost/redis"
	"mrhost/telegram"

	"github.com/go-kit/kit/log"
	"github.com/jessevdk/go-flags"
)

func main() {
	var opts struct {
		TelegramToken string `short:"t" long:"telegram-token" description:"@MisterHostBot Telegram token" env:"MISTERHOST_BOT_TOKEN"`
		ProxyURL      string `short:"p" long:"proxy-url" description:"Telegram proxy URL (socks5:// supported)" env:"MISTERHOST_BOT_PROXY"`
		RedisAddr     string `short:"r" long:"redis-addr" description:"Host and port for Redis" env:"MISTERHOST_REDIS_ADDR"`
	}

	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(0)
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	var (
		proxyURL *url.URL
		err      error
	)
	if opts.ProxyURL != "" {
		proxyURL, err = url.Parse(opts.ProxyURL)
		if err != nil {
			logger.Log("msg", "failed to parse proxy URL", "error", err)
			os.Exit(1)
		}
	}
	bot := &telegram.ChatBot{
		Repository:    redis.NewRepository(opts.RedisAddr),
		TelegramToken: opts.TelegramToken,
		ProxyURL:      proxyURL,
		Logger:        log.With(logger, "component", "telegram"),
	}

	logger.Log("msg", "starting Telegram bot")
	if err := bot.Start(); err != nil {
		logger.Log("msg", "error starting Telegram bot", "error", err)
		os.Exit(1)
	}
	logger.Log("msg", "started Telegram bot")

	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	logger.Log("msg", "received signal", "signal", <-signalChannel)

	logger.Log("msg", "stopping Telegram bot")
	if err := bot.Stop(); err != nil {
		logger.Log("msg", "error stopping Telegram bot", "error", err)
	}
	logger.Log("msg", "stopped Telegram bot")
}
