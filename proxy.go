package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

type User struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

const usersEndpointURL = "https://jsonplaceholder.typicode.com/users"

var errUserNotFound = errors.New("user not found")

func main() {
	// GET /users/:id
	http.HandleFunc("/users/", userHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start HTTP server.
	log.Println("Listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
	case http.MethodOptions:
		addCORSHeaders(w)
		return
	default:
		errResponse(w, errors.New(http.StatusText(http.StatusMethodNotAllowed)), http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from URL path.
	id, err := strconv.Atoi(path.Base(r.URL.String()))
	if err != nil {
		// We'll treat inability to parse the user ID as not having found the user, and return a 404 Not Found.
		errResponse(w, errors.New(http.StatusText(http.StatusNotFound)), http.StatusNotFound)
		return
	}

	// Get user from external service.
	user, err := getUser(id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == errUserNotFound {
			status = http.StatusNotFound
		}
		errResponse(w, err, status)
		return
	}

	// Create JSON response.
	addCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		panic(err)
	}
}

func errResponse(w http.ResponseWriter, e error, statusCode int) {
	body := struct {
		Error      string `json:"error"`
		StatusCode int    `json:"status_code"`
	}{
		Error:      e.Error(),
		StatusCode: statusCode,
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		panic(err)
	}
}

func getUser(id int) (*User, error) {
	url := fmt.Sprintf("%s/%d", usersEndpointURL, id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errUserNotFound
	}

	var u User
	err = json.NewDecoder(resp.Body).Decode(&u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func addCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{
		"Accept",
		"Accept-Encoding",
		"Content-Type",
		"Content-Length",
	}, ", "))
}
