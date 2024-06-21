package main

import (
	"context"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"project/internal/config"
	"project/internal/item"
	db2 "project/internal/item/db"
	"project/internal/user"
	"project/internal/user/db"
	"project/pkg/client/mongodb"
	"time"
)

func main() {
	log.Println("CreateRouter")
	router := httprouter.New()
	cfg := config.GetConfig()
	cfgMongo := cfg.MongoDB
	MongoDBClient, err := mongodb.NewClient(context.Background(), cfgMongo.Host, cfgMongo.Port, cfgMongo.Username,
		cfgMongo.Password, cfgMongo.Database, cfgMongo.AuthDB)
	if err != nil {
		panic(err)
	}

	itemStorage := db2.NewItemStorage(MongoDBClient, "items")
	create, err := itemStorage.Create(context.Background(), item.Item{
		Name:   "Valenennok",
		Price:  15000,
		UserID: "",
	})
	if err != nil {
		panic(err)
	}
	one, err := itemStorage.FindOne(context.Background(), create)
	if err != nil {
		panic(err)
	}
	fmt.Println(one)

	handler := item.NewItemHandler(itemStorage)
	handler.Register(router)

	storage := db.NewStorage(MongoDBClient, cfg.MongoDB.Collection)
	//create, err := storage.Create(context.Background(), user.User{
	//	Username:     "admin",
	//	Email:        "admin",
	//	IsAdmin:      true,
	//})
	//findOne, err := storage.FindOne(context.Background(), create)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(findOne)

	log.Println("Register user handler")

	handler = user.NewHandler(storage)
	handler.Register(router)

	Start(router, cfg)
}
func Start(router *httprouter.Router, cfg *config.Config) {
	log.Println("Start Application")
	var listener net.Listener
	var listenerErr error
	if cfg.Listen.Type == "sock" {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			panic(err)
		}
		log.Println("create socket")
		socketPath := path.Join(appDir, "app.sock")
		log.Printf("socket path %s", socketPath)
		log.Println("listen unix socket ")
		listener, listenerErr = net.Listen("unix", socketPath)
		log.Printf("server is listening unix socket: %s", socketPath)
	} else {
		log.Println("listen tcp")
		listener, listenerErr = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIp, cfg.Listen.PortIp))
		log.Printf("Server is listening port %s:%s", cfg.Listen.BindIp, cfg.Listen.PortIp)
	}
	if listenerErr != nil {
		log.Fatalln(listenerErr)
	}

	server := &http.Server{
		Handler:      router,
		WriteTimeout: time.Second * 10,
		ReadTimeout:  time.Second * 10,
	}

	log.Fatalln(server.Serve(listener))
}
