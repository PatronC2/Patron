package main

/*
 * Depends on API container and PostgreSQL DB running
 * Tests creating users, permissions, and more to come.
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	username      = "patron"
	password      = "password1!"
	patronIP      = "192.168.50.240"
	patronAPIPort = "8080"
)

type LoginResponse struct {
	Token string `json:"token"`
	Error string `json:"error"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func main() {
	ERROR_COUNT := 0
	SUCCESS_COUNT := 0
	TEST_NAME := "LOGIN TEST"
	fmt.Printf("Beginning Test: %s\n", TEST_NAME)
	token, err := login(username, password)
	if err != nil {
		fmt.Printf("%s: Login failed: %v", TEST_NAME, err)
		ERROR_COUNT += 1
	} else {
		SUCCESS_COUNT += 1
		fmt.Printf("%s: Login successful.\n%s: Token: %s\n", TEST_NAME, TEST_NAME, token)
	}

	response, err := getData(token)
	if err != nil {
		fmt.Printf("%s: Failed to get data: %v", TEST_NAME, err)
		ERROR_COUNT += 1
	} else {
		fmt.Printf("%s: Response from /api/data: %s\n\n", TEST_NAME, response)
		SUCCESS_COUNT += 1
	}

	newUsername := "testuser"
	newPassword := "password1!"
	newRole := "readOnly"

	err = createUser(token, newUsername, newPassword, newRole)
	if err != nil {
		fmt.Printf("%s: Failed to create new user: %v\n", TEST_NAME, err)
		ERROR_COUNT += 1
	} else {
		fmt.Printf("%s: Created new user: %v\n", TEST_NAME, newUsername)
		SUCCESS_COUNT += 1
	}

	fmt.Printf("%s: Trying login as %s\n", TEST_NAME, newUsername)
	roToken, err := login(newUsername, newPassword)
	if err != nil {
		fmt.Printf("%s: Login failed: %v\n", TEST_NAME, err)
		ERROR_COUNT += 1
	} else {
		fmt.Printf("%s: PASS\n", TEST_NAME)
		SUCCESS_COUNT += 1
	}

	TEST_NAME = "PRIVILEGE TEST"
	fmt.Printf("Beginning Test: %s\n", TEST_NAME)
	invalidUsername := "crap"
	invalidPassword := "crap"
	invalidRole := "admin"

	err = createUser(roToken, invalidUsername, invalidPassword, invalidRole)
	if err == nil {
		fmt.Printf("%s: Invalid user creation should have failed but did not\n", TEST_NAME)
		ERROR_COUNT += 1
	} else {
		fmt.Printf("%s: PASS\n", TEST_NAME)
		SUCCESS_COUNT += 1
	}
	fmt.Printf("------------------\n\nSuccess Count: %d\nFailure Count: %d\n", SUCCESS_COUNT, ERROR_COUNT)
	if ERROR_COUNT <= 0 {
		fmt.Println("All Tests Successful")
	} else{
		fmt.Println("Tests are unstable")
	}
}

func login(username, password string) (string, error) {
	url := fmt.Sprintf("http://%s:%s/login", patronIP, patronAPIPort)
	reqBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to make login request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return "", fmt.Errorf("failed to parse login response: %w", err)
	}

	if loginResp.Token == "" {
		return "", fmt.Errorf("login failed: %s", loginResp.Error)
	}

	return loginResp.Token, nil
}

func getData(token string) (string, error) {
	url := fmt.Sprintf("http://%s:%s/api/data", patronIP, patronAPIPort)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("%s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make get data request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}

func createUser(token, username, password, role string) error {
	url := fmt.Sprintf("http://%s:%s/users", patronIP, patronAPIPort)
	user := CreateUserRequest{
		Username: username,
		Password: password,
		Role:     role,
	}
	reqBody, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("%s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make create user request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to create user: %s", string(body))
	}

	return nil
}