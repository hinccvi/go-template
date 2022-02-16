package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"hinccvi/go-template/config"
	"hinccvi/go-template/dao/gorm"
	"hinccvi/go-template/dao/redis"
	"hinccvi/go-template/log"
	"hinccvi/go-template/route"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//	1. Init system timezone
	time.Local = time.FixedZone("GMT", 7*3600)
	//	2. Init yaml config
	config.Init()
	//	3. Init logger
	log.Init(config.Conf.AppConfig.Env)
	//	4. Init gorm
	gorm.Init(config.Conf.AppConfig.Env)
	//	5. Init redis
	redis.Init()
	//	6. Init gin router
	router := route.Init(config.Conf.AppConfig.Env)
	//	7. Init & Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Conf.AppConfig.Port),
		Handler: router,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panic("Fail to listen on port", "port", config.Conf.AppConfig.Port, zap.Error(err))
		}
	}()
	log.Info("Listening on", "port", config.Conf.AppConfig.Port)
	//	8. Gracefully shutdown server with 5 sec delay
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Panic("Server failed to shutdown", zap.Error(err))
	}
	log.Info("Server shut down")
}