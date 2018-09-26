// Copyright © 2018 Alexander Pinnecke <alexander.pinnecke@googlmain.com>
//

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Scalify/puppet-master-client-go"
	"github.com/Scalify/website-content-watcher/pkg/config"
	"github.com/Scalify/website-content-watcher/pkg/mail"
	"github.com/Scalify/website-content-watcher/pkg/notifier"
	"github.com/Scalify/website-content-watcher/pkg/storage"
	"github.com/Scalify/website-content-watcher/pkg/watcher"
	"github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/kelseyhightower/envconfig"
	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"gopkg.in/gomail.v2"
)

type env struct {
	RedisDb              int    `required:"true" split_words:"true"`
	RedisPort            int    `required:"true" split_words:"true"`
	RedisHost            string `required:"true" split_words:"true"`
	PuppetMasterEndpoint string `required:"true" split_words:"true"`
	PuppetMasterAPIToken string `required:"true" split_words:"true" envconfig:"PUPPET_MASTER_API_TOKEN"`
	MailNotifierEnabled  bool   `default:"false" split_words:"true"`
	Verbose              bool   `default:"false" split_words:"true"`
}

type mailEnv struct {
	SMTPHost          string `required:"true" split_words:"true" envconfig:"SMTP_HOST"`
	SMTPPort          int    `required:"true" split_words:"true" envconfig:"SMTP_PORT"`
	SMTPUser          string `required:"false" split_words:"true" envconfig:"SMTP_USER"`
	SMTPPass          string `required:"false" split_words:"true" envconfig:"SMTP_PASS"`
	MailSenderAddress string `required:"false" split_words:"true"`
}

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use: "watch <config-file>",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()
		ctx := newExitHandlerContext(logger)

		if len(args) < 1 {
			if err := cmd.Usage(); err != nil {
				logger.Fatal(err)
			}
			os.Exit(1)
		}

		var cfg env
		if err := envconfig.Process("", &cfg); err != nil {
			logger.Fatal(err)
		}

		setupLogger(logger, cfg.Verbose)
		c := cron.New()

		redisClient := storage.NewRedis(connectRedis(logger, cfg))
		pmClient, err := puppetmaster.NewClient(cfg.PuppetMasterEndpoint, cfg.PuppetMasterAPIToken)
		if err != nil {
			logger.Fatalf("failed to connect puppet master: %v", err)
		}

		configFile, err := filepath.Abs(args[0])
		if err != nil {
			logger.Fatalf("failed to resolve config file path: %v", err)
		}

		conf, err := config.Load(configFile)
		if err != nil {
			logger.Fatalf("failed to load config from %q: %v", configFile, err)
		}

		w := watcher.New(logger.WithFields(logrus.Fields{}), redisClient, pmClient, configFile, conf)

		addNotifiers(logger, w, cfg)

		if err := w.CheckConfig(); err != nil {
			logger.Fatal(err)
		}

		if err := w.RegisterCronJobs(c); err != nil {
			logger.Fatal(err)
		}
		c.Start()

		logger.Info("Started cron job.")

		<-ctx.Done()
		logger.Info("Stopping ...")
		c.Stop()
	},
}

func addNotifiers(logger *logrus.Logger, w *watcher.Watcher, cfg env) {
	if cfg.MailNotifierEnabled {
		var mailCfg mailEnv
		if err := envconfig.Process("", &mailCfg); err != nil {
			logger.Fatal(err)
		}

		mailClient := mail.New(gomail.NewDialer(mailCfg.SMTPHost, mailCfg.SMTPPort, mailCfg.SMTPUser, mailCfg.SMTPPass))
		mailNotifier := notifier.NewMail(mailCfg.MailSenderAddress, mailClient)
		if err := w.AddNotifier(mailNotifier); err != nil {
			logger.Fatal(err)
		}
	}
}

func connectRedis(logger *logrus.Logger, cfg env) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		DB:   cfg.RedisDb,
	})
	if pong, err := redisClient.Ping().Result(); err != nil || pong != "PONG" {
		logger.Fatalf("Error pinging redis: %v --> %v", pong, err)
	}

	return redisClient
}

func init() {
	RootCmd.AddCommand(watchCmd)
}
