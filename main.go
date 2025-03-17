package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

type Request struct {
	Data string `json:"data"`
}

type Response struct {
	Encoded string `json:"encoded"`
}

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	var req Request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("/shared-bin/sidecar", req.Data)
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, "Error executing sidecar", http.StatusInternalServerError)
		return
	}

	res := Response{Encoded: string(output)}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func main() {
	http.HandleFunc("/encode", encodeHandler)

	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
