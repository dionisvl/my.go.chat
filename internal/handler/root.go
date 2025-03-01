package handler

import (
	"fmt"
	"net/http"
	"os"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("web/index.html")
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, string(data))
}
