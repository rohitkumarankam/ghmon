package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Repo represents a GitHub repository
type Repo struct {
    Name      string    `json:"name"`
    HTMLURL   string    `json:"html_url"`
    UpdatedAt time.Time `json:"updated_at"`
}

// User represents a GitHub user
type User struct {
    Login string `json:"login"`
}

// ChatMessage represents the payload we send to Google Chat
type ChatMessage struct {
    Text string `json:"text"`
}

// RateLimitInfo tracks GitHub API rate-limiting
type RateLimitInfo struct {
    Limit     string
    Remaining string
    Reset     string
}

// run is the main logic invoked by the Cobra command in root.go
func run(org, webhook, pat string) error {
    client := &http.Client{}
    var rateLimit RateLimitInfo

    // Fetch repositories for the organization
    repos, err := fetchRepos(client, org, pat, &rateLimit)
    if err != nil {
        return fmt.Errorf("fetching org repos: %w", err)
    }

    // Fetch users in the organization
    users, err := fetchUsers(client, org, pat, &rateLimit)
    if err != nil {
        return fmt.Errorf("fetching org users: %w", err)
    }

    // We only want to look for updates within the last hour
    oneHourAgo := time.Now().Add(-1 * time.Hour)
    var updatedRepos []Repo

    // Check org repos
    for _, repo := range repos {
        if repo.UpdatedAt.After(oneHourAgo) {
            updatedRepos = append(updatedRepos, repo)
            fmt.Printf("Org Repo - Name: %s, URL: %s\n", repo.Name, repo.HTMLURL)
        }
    }

    // Check each user's repos
    for _, user := range users {
        userRepos, err := fetchUserRepos(client, user.Login, pat, &rateLimit)
        if err != nil {
            // We skip errors here but log them
            log.Printf("Error fetching repos for user %s: %v", user.Login, err)
            continue
        }
        for _, repo := range userRepos {
            if repo.UpdatedAt.After(oneHourAgo) {
                updatedRepos = append(updatedRepos, repo)
                fmt.Printf("User %s Repo - Name: %s, URL: %s\n", user.Login, repo.Name, repo.HTMLURL)
            }
        }
    }

    // If any repos were updated in the last hour
    if len(updatedRepos) > 0 {
        if webhook == "" {
            // No webhook provided - just print to console
            fmt.Println("Repositories updated in the last hour:")
            for _, repo := range updatedRepos {
                fmt.Printf("Name: %s, URL: %s\n", repo.Name, repo.HTMLURL)
            }
        } else {
            // Webhook provided - send to Google Chat
            if err := notifyChat(webhook, updatedRepos); err != nil {
                return fmt.Errorf("sending chat notification: %w", err)
            }
        }
    } else {
        fmt.Println("No repos updated in the last 1 hour")
    }

    // Print rate limit info before exiting
    printRateLimitInfo(rateLimit)

    return nil
}

// fetchRepos retrieves all public repositories for a given org
func fetchRepos(client *http.Client, org, pat string, rateLimit *RateLimitInfo) ([]Repo, error) {
    url := fmt.Sprintf("https://api.github.com/orgs/%s/repos?type=public", org)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("creating request for org repos: %w", err)
    }
    req.Header.Set("Authorization", fmt.Sprintf("token %s", pat))

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetching org repos: %w", err)
    }
    defer resp.Body.Close()

    updateRateLimit(resp, rateLimit)

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("fetch org repos: received status %d", resp.StatusCode)
    }

    var repos []Repo
    if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
        return nil, fmt.Errorf("decoding org repos JSON: %w", err)
    }

    return repos, nil
}

// fetchUsers retrieves all members of the organization
func fetchUsers(client *http.Client, org, pat string, rateLimit *RateLimitInfo) ([]User, error) {
    url := fmt.Sprintf("https://api.github.com/orgs/%s/members", org)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("creating request for org members: %w", err)
    }
    req.Header.Set("Authorization", fmt.Sprintf("token %s", pat))

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetching org users: %w", err)
    }
    defer resp.Body.Close()

    updateRateLimit(resp, rateLimit)

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("fetch org users: received status %d", resp.StatusCode)
    }

    var users []User
    if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
        return nil, fmt.Errorf("decoding users JSON: %w", err)
    }

    return users, nil
}

// fetchUserRepos retrieves repos for a particular user (within the org)
func fetchUserRepos(client *http.Client, user, pat string, rateLimit *RateLimitInfo) ([]Repo, error) {
    url := fmt.Sprintf("https://api.github.com/users/%s/repos?type=owner", user)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("creating request for user repos: %w", err)
    }
    req.Header.Set("Authorization", fmt.Sprintf("token %s", pat))

    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetching user repos: %w", err)
    }
    defer resp.Body.Close()

    updateRateLimit(resp, rateLimit)

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("fetch user repos: received status %d", resp.StatusCode)
    }

    var repos []Repo
    if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
        return nil, fmt.Errorf("decoding user repos JSON: %w", err)
    }

    return repos, nil
}

// notifyChat sends a message to the given Google Chat webhook
func notifyChat(webhook string, updatedRepos []Repo) error {
    var msgBody string
    for _, repo := range updatedRepos {
        msgBody += fmt.Sprintf("Name: %s, URL: %s\n", repo.Name, repo.HTMLURL)
    }

    message := ChatMessage{
        Text: fmt.Sprintf("Repositories updated in the last hour:\n%s", msgBody),
    }

    msgBytes, _ := json.Marshal(message)
    resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(msgBytes))
    if err != nil {
        return fmt.Errorf("posting to chat webhook: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode > 299 {
        return fmt.Errorf("chat webhook returned status %d", resp.StatusCode)
    }

    return nil
}

// updateRateLimit extracts GitHub rate-limit info from response headers
func updateRateLimit(resp *http.Response, rateLimit *RateLimitInfo) {
    rateLimit.Limit = resp.Header.Get("X-RateLimit-Limit")
    rateLimit.Remaining = resp.Header.Get("X-RateLimit-Remaining")
    rateLimit.Reset = resp.Header.Get("X-RateLimit-Reset")
}

// printRateLimitInfo prints rate-limiting data if present
func printRateLimitInfo(rateLimit RateLimitInfo) {
    if rateLimit.Limit == "" && rateLimit.Remaining == "" && rateLimit.Reset == "" {
        return
    }

    resetInt, err := strconv.ParseInt(rateLimit.Reset, 10, 64)
    if err == nil {
        resetTime := time.Unix(resetInt, 0)
        fmt.Printf("Rate Limit: %s, Remaining: %s, Reset At: %s\n",
            rateLimit.Limit, rateLimit.Remaining, resetTime)
    } else {
        fmt.Printf("Rate Limit: %s, Remaining: %s, Reset At: %s\n",
            rateLimit.Limit, rateLimit.Remaining, rateLimit.Reset)
    }
}