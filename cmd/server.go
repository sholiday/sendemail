package main

import (
	"fmt"
	"log"

	"github.com/sholiday/sendemail"
	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	viper.SetConfigName("sendemail")
	viper.AddConfigPath("$HOME/.config/sendemail/")
	viper.AddConfigPath("/etc/sendemail/")
	viper.AddConfigPath("/config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Failed to read config file: %s", err)
	}
	var config sendemail.Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %s", err)
	}
	fmt.Printf("%+v\n", config)

	s := sendemail.New(config)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*.tmpl")
	r.GET("/", s.Main)
	r.POST("/", s.Main)
	r.Run(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port))
}
