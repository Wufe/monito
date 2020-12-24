package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	"github.com/wufe/monito/worker/cli"
	"github.com/wufe/monito/worker/database"
	"github.com/wufe/monito/worker/services"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	parsedCliArguments := cli.ParseCliArguments(os.Args)

	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		log.Fatal("Error loading .env file")
	}

	db := database.ConnectToDatabase()
	defer db.Close()

	var waitGroup sync.WaitGroup

	waitGroup.Add(1)

	queues := services.RegisterWorker(db, parsedCliArguments)
	fmt.Println("Worker registered itself.")

	services.NewQueueOrchestrator(queues, db).Init()

	waitGroup.Wait()
}
