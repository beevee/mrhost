package telegram

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"mrhost"

	"golang.org/x/sync/errgroup"
	"gopkg.in/tucnak/telebot.v2"
)

type ChatBot struct {
	TelegramToken string
	Repository    mrhost.Repository
	ProxyURL      *url.URL
	Logger        mrhost.Logger
	telebot       *telebot.Bot
	ctx           context.Context
	ctxCancel     context.CancelFunc
	ctxGroup      *errgroup.Group
}

func (b *ChatBot) Start() error {
	transport := &http.Transport{}
	if b.ProxyURL != nil {
		transport.Proxy = http.ProxyURL(b.ProxyURL)
		b.Logger.Log("msg", "working via proxy", "proxy_host", b.ProxyURL.Host,
			"proxy_user", b.ProxyURL.User.Username(), "proxy_proto", b.ProxyURL.Scheme)
	}
	client := http.DefaultClient
	client.Transport = transport

	var err error
	b.telebot, err = telebot.NewBot(telebot.Settings{
		Token:  b.TelegramToken,
		Poller: &telebot.LongPoller{Timeout: 1 * time.Second},
		Client: client,
		Reporter: func(err error) {
			b.Logger.Log("msg", "telebot error", "error", err)
		},
	})
	if err != nil {
		return err
	}

	b.telebot.Handle("/more", b.handleMore)
	b.telebot.Handle("/answer", b.handleAnswer)

	b.ctx, b.ctxCancel = context.WithCancel(context.Background())
	b.ctxGroup, b.ctx = errgroup.WithContext(b.ctx)

	go b.telebot.Start()

	return nil
}

func (b *ChatBot) Stop() error {
	b.ctxCancel()
	err := b.ctxGroup.Wait()

	b.telebot.Stop()

	return err
}
