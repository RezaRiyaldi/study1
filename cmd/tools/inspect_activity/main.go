package main

import (
	"fmt"
	"log"

	"study1/internal/core/config"
	"study1/internal/core/database"
	"study1/internal/modules/activity"
)

func main() {
	cfg := config.LoadConfig()
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	var rows []activity.ActivityLog
	if err := db.Order("created_at desc").Limit(10).Find(&rows).Error; err != nil {
		log.Fatalf("query failed: %v", err)
	}

	if len(rows) == 0 {
		fmt.Println("no activity logs found")
		return
	}

	for _, r := range rows {
		fmt.Printf("%s %s -> %d (%dms) ip=%s user=%v\n", r.Method, r.Path, r.Status, r.LatencyMs, r.IP, r.UserID)
	}
}
