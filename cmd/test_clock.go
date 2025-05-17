package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type CheckInRequest struct {
	ID        string `json:"access_id"`
	UserID    string `json:"user_id"`
	Time      string `json:"access_time"`
	Direction string `json:"direction"`
	Gate_type string `json:"gate_type"`
	CheckinAt string `json:"gate_name"`
}

func test_clock() {
	total := 4000
	url := "http://localhost:8080/api/clock"

	for i := 0; i < total; i++ {
		req := CheckInRequest{
			ID:        uuid.New().String(),
			UserID:    fmt.Sprintf("user-%d", i%100), // 模擬 100 個 user
			Time:      time.Now().Format(time.RFC3339),
			Direction: "IN", // 或 "OUT"
			Gate_type: "NFC",
			CheckinAt: fmt.Sprintf("Gate-%d", i%5), // 模擬 5 個閘門
		}

		payload, _ := json.Marshal(req)
		reqBody := bytes.NewBuffer(payload)

		resp, err := http.Post(url, "application/json", reqBody)
		if err != nil {
			fmt.Printf("Failed at %d: %v\n", i, err)
			continue
		}
		resp.Body.Close()

		if i%500 == 0 {
			fmt.Printf("Sent %d requests\n", i)
		}
	}

	fmt.Println("Done.")
}
