package roblox

import (
    "errors"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "regexp"
    "strconv"
)

type VipLinkData struct {
    PlaceID  int64
    LinkCode string
}

func ParseVipLink(link string) (*VipLinkData, error) {
    u, err := url.Parse(link)
    if err != nil {
        return nil, err
    }

    // Extract PlaceID
    re := regexp.MustCompile(`/games/(\d+)`)
    matches := re.FindStringSubmatch(u.Path)
    if len(matches) < 2 {
        return nil, errors.New("invalid game link")
    }

    placeID, err := strconv.ParseInt(matches[1], 10, 64)
    if err != nil {
        return nil, err
    }

    // Extract LinkCode
    query := u.Query()
    linkCode := query.Get("privateServerLinkCode")
    if linkCode == "" {
        return nil, errors.New("no private server link code found")
    }

    return &VipLinkData{
        PlaceID:  placeID,
        LinkCode: linkCode,
    }, nil
}

func GetCSRFToken(cookie string) (string, error) {
    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }

    req, err := http.NewRequest("POST", "https://auth.roblox.com/v1/authentication-ticket/", nil)
    if err != nil {
        return "", err
    }

    req.Header.Set("Cookie", ".ROBLOSECURITY="+cookie)
    req.Header.Set("Referer", "https://www.roblox.com/")

    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    csrfToken := resp.Header.Get("x-csrf-token")
    if csrfToken == "" {
        return "", errors.New("failed to get CSRF token")
    }

    return csrfToken, nil
}

func GetAuthTicket(cookie, csrfToken string) (string, error) {
    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }

    req, err := http.NewRequest("POST", "https://auth.roblox.com/v1/authentication-ticket/", nil)
    if err != nil {
        return "", err
    }

    req.Header.Set("Cookie", ".ROBLOSECURITY="+cookie)
    req.Header.Set("X-CSRF-TOKEN", csrfToken)
    req.Header.Set("Referer", "https://www.roblox.com/")

    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    authTicket := resp.Header.Get("rbx-authentication-ticket")
    if authTicket == "" {
        return "", errors.New("failed to get auth ticket")
    }

    return authTicket, nil
}

func GetAccessCode(cookie string, placeID int64, linkCode string) (string, error) {
    client := &http.Client{}

    url := fmt.Sprintf("https://www.roblox.com/games/%d?privateServerLinkCode=%s", placeID, linkCode)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return "", err
    }

    req.Header.Set("Cookie", ".ROBLOSECURITY="+cookie)

    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    // Extract access code from JavaScript
    re := regexp.MustCompile(`Roblox\.GameLauncher\.joinPrivateGame\(\d+\s*,\s*'([\w-]+)'`)
    matches := re.FindStringSubmatch(string(body))
    if len(matches) < 2 {
        return "", errors.New("failed to extract access code")
    }

    return matches[1], nil
}