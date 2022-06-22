package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/hibiken/asynq"
)

const SYNC_TASK_QUEUE = "sync-ops"
const SYNC_TASK_TYPE = "SYNC:OPS"

var redisConnOpt = asynq.RedisClientOpt{Addr: "localhost:63790"}

func schedulerRun() {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	scheduler := asynq.NewScheduler(
		redisConnOpt,
		&asynq.SchedulerOpts{
			Location: loc,
		},
	)

	task := asynq.NewTask(SYNC_TASK_TYPE, nil)

	entryID, err := scheduler.Register("@every 5s", task, asynq.Queue(SYNC_TASK_QUEUE))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("registered an entry: %q\n", entryID)

	if err := scheduler.Run(); err != nil {
		log.Fatal(err)
	}
}

func handler(ctx context.Context, t *asynq.Task) error {
	println("in ->")
	switch t.Type() {
	case SYNC_TASK_TYPE:
		println("---- --- --- --- ------------------------", time.Now().String())
	default:
		println(":-:-:-:-:-:-:-:-:-")
	}
	return nil
}
func main() {
	args := os.Args
	if args[1] == "0" {
		srv := asynq.NewServer(
			redisConnOpt,
			asynq.Config{Concurrency: 10, Queues: map[string]int{
				SYNC_TASK_QUEUE: 6,
			}},
		)
		mux := asynq.NewServeMux()
		mux.HandleFunc(SYNC_TASK_TYPE, handler)
		if err := srv.Run(mux); err != nil {
			log.Fatalf("could not run server: %v", err)
		}
	} else if args[1] == "1" {
		schedulerRun()
	}
}
