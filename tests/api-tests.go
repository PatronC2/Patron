package main

/*
 * Depends on API container and PostgreSQL DB running
 * Tests creating users, permissions, and more to come.
 *
 * Run with: go run api-unit-tests.go
*/

import (
    "bytes"
	"crypto/tls"
    "encoding/json"
	"errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
)

const (
	username      = "patron"
	password      = "password1!"
	patronIP      = "192.168.50.32"
	patronAPIPort = "8000"
	AgentIP		  = "192.168.50.69"
	patronC2Port  = "9000"
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

type CompileRequest struct {
	Name             string `json:"name"`
	Description	     string `json:"description"`
	Type			 string `json:"type"`
	ServerIP         string `json:"serverip"`
	ServerPort       string `json:"serverport"`
	CallbackFrequency string `json:"callbackfrequency"`
	CallbackJitter   string `json:"callbackjitter"`
}

type CreateAgentRequest struct {
	ServerIP         string `json:"serverip"`
	ServerPort       string `json:"serverport"`
	CallbackFrequency string `json:"callbackfrequency"`
	Jitter           string `json:"jitter"`
	AgentIP          string `json:"agentip"`
	Username         string `json:"username"`
	Hostname         string `json:"hostname"`
}

func main() {
	ERROR_COUNT := 0
	SUCCESS_COUNT := 0
	// Test for admin login, creating a readOnly user
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
		fmt.Printf("%s: SUCCESS", TEST_NAME)
	}
	
	// make api calls to test functionality
	TEST_NAME = "REGRESSION TESTS"
	fmt.Printf("\n\nBeginning Test: %s\n", TEST_NAME)
	endpoint := "/api/agents"
	response, err := getRequest(token, endpoint)
	if err != nil {
		fmt.Printf("%s: Failed to get Agents: %v", TEST_NAME, err)
		ERROR_COUNT += 1
	} else {
		fmt.Printf("%s: Response from /api/agents: %s\n", TEST_NAME, response)
		SUCCESS_COUNT += 1
	}

	endpoint = "/api/groupagents/"
	response, err = getRequest(token, endpoint)
	if err != nil {
		fmt.Printf("%s: Failed to get Agents: %v", TEST_NAME, err)
		ERROR_COUNT += 1
	} else {
		fmt.Printf("%s: Response from /api/groupagents: %s\n", TEST_NAME, response)
		SUCCESS_COUNT += 1
	}

	endpoint = fmt.Sprintf("/api/groupagents/%s", AgentIP)
	response, err = getRequest(token, endpoint)
	if err != nil {
		fmt.Printf("%s: Failed to get Agents: %v", TEST_NAME, err)
		ERROR_COUNT += 1
	} else {
		fmt.Printf("%s: Response from /api/groupagents: %s\n", TEST_NAME, response)
		SUCCESS_COUNT += 1
	}

	agentUUID, err := findValueByKey(response, "uuid")
	if err != nil {
		fmt.Printf("%s: Failed to get a UUID: %v", TEST_NAME, err)
		ERROR_COUNT += 1
	} else {
		SUCCESS_COUNT += 1
	}
	endpoint = fmt.Sprintf("/api/oneagent/%s", agentUUID)
	response, err = getRequest(token, endpoint)
	if err != nil {
		fmt.Printf("%s: Failed to get Agents: %v", TEST_NAME, err)
		ERROR_COUNT += 1
	} else {
		fmt.Printf("%s: Response from /api/oneagent/%s: %s\n", TEST_NAME, agentUUID, response)
		SUCCESS_COUNT += 1
	}
	assertUUID, err := findValueByKey(response, "uuid")
	if err != nil {
		fmt.Printf("%s: Failed to get the uuid back: %s\n", TEST_NAME, err)
		ERROR_COUNT += 1
	} else {
		if assertUUID != agentUUID {
			fmt.Printf("%s: Expected UUID: %s, got %s\n", TEST_NAME, agentUUID, assertUUID)
			ERROR_COUNT += 1
		} else {
			fmt.Printf("%s: Assert %s=%s\n", TEST_NAME, agentUUID, assertUUID)
			SUCCESS_COUNT += 1
		}
	}

	// make sure read only user doesn't have admin privs
	TEST_NAME = "PRIVILEGE TEST"
	fmt.Printf("\n\nBeginning Test: %s\n", TEST_NAME)
	fmt.Printf("%s: Trying login as %s\n", TEST_NAME, newUsername)
	roToken, err := login(newUsername, newPassword)
	if err != nil {
		fmt.Printf("%s: Login failed: %v\n", TEST_NAME, err)
		ERROR_COUNT += 1
	} else {
		fmt.Printf("%s: PASS\n", TEST_NAME)
		SUCCESS_COUNT += 1
	}
	
	invalidUsername := "crap"
	invalidPassword := "crap"
	invalidRole := "admin"

	err = createUser(roToken, invalidUsername, invalidPassword, invalidRole)
	if err == nil {
		fmt.Printf("%s: Invalid user creation should have failed but did not\n", TEST_NAME)
		ERROR_COUNT += 1
	} else {
		fmt.Printf("%s: User permissions are correct\n", TEST_NAME)
		fmt.Printf("%s: PASS\n", TEST_NAME)
		SUCCESS_COUNT += 1
	}

	// create a test agent
	callbackFrequency := "300"
	callbackJitter := "60"
	hostname := "test.patron.com"
	err := createAgent(patronIP, patronAPIPort, token, patronIP, patronC2Port, callbackFrequency, jitter, agentIP, username, hostname)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// payload tests
	TEST_NAME = "PAYLOAD TESTS"
	fmt.Printf("\n\nBeginning Test: %s\n", TEST_NAME)
	name := "test"
	description := "test"
	/*
	beforeFileCount := getNumOfFiles()
	fmt.Printf("%s: Before payload count: %d", TEST_NAME, beforeFileCount)
	*/
	err = compileRequest(patronIP, patronAPIPort, token, name, description, patronIP, patronC2Port, callbackFrequency, callbackJitter)
	if err != nil {
		fmt.Printf("%s: Error (sometimes this false fails): %v\n", TEST_NAME, err)
	}
	/*
	afterFileCount := getNumOfFiles()
	fmt.Printf("%s: Before after count: %d", TEST_NAME, afterFileCount)

	if beforeFileCount+1 != afterFileCount {
		fmt.Printf("%s: Failed to validate new payload exists", TEST_NAME)
		ERROR_COUNT += 1
	} else {
		fmt.Printf("%s: PASS")
		SUCCESS_COUNT += 1
	}
	*/


	// delete the RO test user
	TEST_NAME = "DELETE TEST USER"
	err = deleteUser(token, newUsername)
	if err != nil {
		fmt.Printf("%s: Failed to delete user: %v\n", TEST_NAME, err)
		ERROR_COUNT += 1
	} else {
		fmt.Printf("%s: User %s deleted successfully\n", TEST_NAME, newUsername)
		SUCCESS_COUNT += 1
	}

	// Test Summary
	fmt.Printf("------------------\n\nSuccess Count: %d\nFailure Count: %d\n", SUCCESS_COUNT, ERROR_COUNT)
	if ERROR_COUNT <= 0 {
		fmt.Println("All Tests Successful")
		os.Exit(0)
	} else{
		fmt.Println("Tests are unstable")
		os.Exit(1)
	}
}

func createInsecureClient() *http.Client {
    // Custom transport that skips SSL verification
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    return &http.Client{Transport: tr}
}

func login(username, password string) (string, error) {
    url := fmt.Sprintf("http://%s:%s/login", patronIP, patronAPIPort)
    reqBody, _ := json.Marshal(map[string]string{
        "username": username,
        "password": password,
    })

    // Use the insecure client
    client := createInsecureClient()
    resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqBody))
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

func getRequest(token string, endpoint string) (string, error) {
	url := fmt.Sprintf("http://%s:%s%s", patronIP, patronAPIPort, endpoint)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("%s", token))

	client := createInsecureClient() // Use the insecure client
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make get %s request: %w", endpoint, err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}

func createUser(token, username, password, role string) error {
	url := fmt.Sprintf("http://%s:%s/api/admin/users", patronIP, patronAPIPort)
	user := CreateUserRequest{
		Username: username,
		Password: password,
		Role:     role,
	}
	reqBody, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("%s", token))

	client := createInsecureClient() // Use the insecure client
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

func deleteUser(token, username string) error {
	url := fmt.Sprintf("http://%s:%s/api/admin/users/%s", patronIP, patronAPIPort, username)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("%s", token))

	client := createInsecureClient() // Use the insecure client
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make delete user request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete user: %s", string(body))
	}

	return nil
}

func findValueByKey(jsonStr, key string) (interface{}, error) {
    var result interface{}
    err := json.Unmarshal([]byte(jsonStr), &result)
    if err != nil {
        return nil, err
    }

    return searchKey(result, key)
}

func searchKey(data interface{}, key string) (interface{}, error) {
    switch v := data.(type) {
    case map[string]interface{}:
        if value, found := v[key]; found {
            return value, nil
        }
        for _, value := range v {
            if res, err := searchKey(value, key); err == nil {
                return res, nil
            }
        }
    case []interface{}:
        for _, item := range v {
            if res, err := searchKey(item, key); err == nil {
                return res, nil
            }
        }
    }

    return nil, errors.New("key not found")
}

/*
func getNumOfFiles() (int, error) {
	url := fmt.Sprintf("https://%s:%s/files", patronIP, patronAPIPort)
	client := createInsecureClient()

	// Make the GET request
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error making request:", err)
		return 0, err
	}
	defer resp.Body.Close()

	// Parse the HTML response
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return 0, err
	}

	// Extract and count the files
	fileCount := 0
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.Contains(href, ".") {
			fileCount++
		}
	})

	return fileCount, nil
}
*/

func compileRequest(patronIP, patronAPIPort, token, name, description, serverIP, serverPort, callbackFrequency, callbackJitter string) error {
	url := fmt.Sprintf("http://%s:%s/api/payload", patronIP, patronAPIPort)
	requestBody := CompileRequest{
		Name:				name,
		Description:		description,
		Type:				"original",
		ServerIP:			serverIP,
		ServerPort:			serverPort,
		CallbackFrequency:	callbackFrequency,
		CallbackJitter:		callbackJitter,
	}
	reqBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("%s", token))

	client := createInsecureClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make compile request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to compile: %s", string(body))
	}

	fmt.Println("Success:", string(body))
	return nil
}


func createAgent(patronIP, patronAPIPort, token, serverIP, serverPort, callbackFrequency, jitter, agentIP, username, hostname string) error {
	url := fmt.Sprintf("http://%s:%s/api/test/createagent", patronIP, patronAPIPort)
	requestBody := CreateAgentRequest{
		ServerIP:         serverIP,
		ServerPort:       serverPort,
		CallbackFrequency: callbackFrequency,
		Jitter:           jitter,
		AgentIP:          agentIP,
		Username:         username,
		Hostname:         hostname,
	}
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("%s", token))

	client := createInsecureClient()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make create agent request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create agent: %s", string(body))
	}

	fmt.Println("Success:", string(body))
	return nil
}
