package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	dsn := "postgresql://ltx:iutaol123@127.0.0.1:5432/jobView_db?sslmode=disable"
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		log.Fatalf("Failed to query database: %v", err)
	}

	fmt.Printf("Database connection successful! Result: %d\n", result)
	
	// 测试查询数据库列表
	rows, err := db.Query("SELECT datname FROM pg_database")
	if err != nil {
		log.Fatalf("Failed to query databases: %v", err)
	}
	defer rows.Close()
	
	fmt.Println("Available databases:")
	for rows.Next() {
		var dbname string
		if err := rows.Scan(&dbname); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("- %s\n", dbname)
	}
}