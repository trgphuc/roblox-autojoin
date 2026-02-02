package roblox

import (
    "crypto/rand"
    "fmt"
    "math/big"
    "net/url"
    "os/exec"
    "runtime"
    "time"
)

func LaunchRoblox(authTicket string, placeID int64, accessCode, linkCode string) error {
    launchTime := time.Now().Unix() * 1000
    browserTrackerID := generateBrowserTrackerID()

    placeLauncherURL := fmt.Sprintf(
        "https://assetgame.roblox.com/game/PlaceLauncher.ashx?request=RequestPrivateGame&placeId=%d&accessCode=%s&linkCode=%s",
        placeID, accessCode, linkCode,
    )

    launchURL := fmt.Sprintf(
        "roblox-player:1+launchmode:play+gameinfo:%s+launchtime:%d+placelauncherurl:%s+browsertrackerid:%s+robloxLocale:en_us+gameLocale:en_us+channel:",
        authTicket, launchTime, url.QueryEscape(placeLauncherURL), browserTrackerID,
    )

    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "windows":
        cmd = exec.Command("cmd", "/c", "start", launchURL)
    case "darwin":
        cmd = exec.Command("open", launchURL)
    case "linux":
        cmd = exec.Command("xdg-open", launchURL)
    default:
        return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
    }

    return cmd.Start()
}

func generateBrowserTrackerID() string {
    n1, _ := rand.Int(rand.Reader, big.NewInt(75000))
    n2, _ := rand.Int(rand.Reader, big.NewInt(800000))
    return fmt.Sprintf("%d%d", n1.Int64()+100000, n2.Int64()+100000)
}