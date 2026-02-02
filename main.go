package main

import (
    "fmt"
    "os"
    "strings"
    "time"

    "github.com/yourusername/roblox-autojoin/pkg/account"
    "github.com/yourusername/roblox-autojoin/pkg/roblox"
    "github.com/fatih/color"
    "github.com/spf13/cobra"
)

var (
    storage *account.Storage
    delay   int
)

func main() {
    var err error
    storage, err = account.NewStorage()
    if err != nil {
        color.Red("Error initializing storage: %v", err)
        os.Exit(1)
    }

    rootCmd := &cobra.Command{
        Use:   "roblox-join",
        Short: "Roblox Auto Join - Join VIP servers with multiple accounts",
        Long:  `A CLI tool to automatically join Roblox VIP servers with one or multiple accounts.`,
    }

    rootCmd.AddCommand(addCmd())
    rootCmd.AddCommand(listCmd())
    rootCmd.AddCommand(removeCmd())
    rootCmd.AddCommand(joinCmd())

    if err := rootCmd.Execute(); err != nil {
        color.Red("Error: %v", err)
        os.Exit(1)
    }
}

func addCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "add [name] [cookie]",
        Short: "Add a new account",
        Args:  cobra.ExactArgs(2),
        Run: func(cmd *cobra.Command, args []string) {
            name := args[0]
            cookie := args[1]

            if !strings.HasPrefix(cookie, "_|WARNING:") {
                color.Red("Error: Invalid cookie format")
                return
            }

            acc := &account.Account{
                Name:   name,
                Cookie: cookie,
            }

            if err := storage.AddAccount(acc); err != nil {
                color.Red("Error: %v", err)
                return
            }

            color.Green("✓ Added account: %s", name)
        },
    }
}

func listCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "list",
        Short: "List all accounts",
        Run: func(cmd *cobra.Command, args []string) {
            accounts, err := storage.GetAllAccounts()
            if err != nil {
                color.Red("Error: %v", err)
                return
            }

            if len(accounts) == 0 {
                color.Yellow("No accounts found. Use 'add' command to add accounts.")
                return
            }

            color.Cyan("\nFound %d account(s):\n", len(accounts))
            for i, acc := range accounts {
                fmt.Printf("%d. %s\n", i+1, acc.Name)
            }
        },
    }
}

func removeCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "remove [name]",
        Short: "Remove an account",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            name := args[0]

            if err := storage.RemoveAccount(name); err != nil {
                color.Red("Error: %v", err)
                return
            }

            color.Green("✓ Removed account: %s", name)
        },
    }
}

func joinCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "join [account|all|acc1,acc2,...] [vip-link]",
        Short: "Join VIP server with account(s)",
        Long: `Join a Roblox VIP server with one or multiple accounts.
        
Examples:
  join MainAcc https://www.roblox.com/games/123/game?privateServerLinkCode=abc
  join all https://www.roblox.com/games/123/game?privateServerLinkCode=abc
  join MainAcc,Alt1,Alt2 https://www.roblox.com/games/123/game?privateServerLinkCode=abc`,
        Args: cobra.ExactArgs(2),
        Run: func(cmd *cobra.Command, args []string) {
            accountSelector := args[0]
            vipLink := args[1]

            handleJoin(accountSelector, vipLink, delay)
        },
    }

    cmd.Flags().IntVarP(&delay, "delay", "d", 5, "Delay in seconds between each join")
    return cmd
}

func handleJoin(accountSelector, vipLink string, delaySeconds int) {
    accounts, err := storage.GetAllAccounts()
    if err != nil {
        color.Red("Error loading accounts: %v", err)
        return
    }

    var toJoin []*account.Account

    if strings.ToLower(accountSelector) == "all" {
        toJoin = accounts
    } else if strings.Contains(accountSelector, ",") {
        names := strings.Split(accountSelector, ",")
        for _, name := range names {
            name = strings.TrimSpace(name)
            if acc := findAccount(accounts, name); acc != nil {
                toJoin = append(toJoin, acc)
            } else {
                color.Yellow("Warning: Account '%s' not found", name)
            }
        }
    } else {
        if acc := findAccount(accounts, accountSelector); acc != nil {
            toJoin = append(toJoin, acc)
        } else {
            color.Red("Error: Account '%s' not found", accountSelector)
            return
        }
    }

    if len(toJoin) == 0 {
        color.Yellow("No accounts to join")
        return
    }

    color.Cyan("\nJoining %d account(s) with %ds delay...\n", len(toJoin), delaySeconds)

    for i, acc := range toJoin {
        color.White("[%d/%d] Joining with %s...", i+1, len(toJoin), acc.Name)

        if err := joinServer(acc, vipLink); err != nil {
            color.Red("✗ Error: %v", err)
        } else {
            color.Green("✓ Launched Roblox for %s", acc.Name)
        }

        if i < len(toJoin)-1 {
            color.Yellow("Waiting %d seconds...", delaySeconds)
            time.Sleep(time.Duration(delaySeconds) * time.Second)
        }
    }

    color.Green("\n✓ Done!")
}

func findAccount(accounts []*account.Account, name string) *account.Account {
    for _, acc := range accounts {
        if strings.EqualFold(acc.Name, name) {
            return acc
        }
    }
    return nil
}

func joinServer(acc *account.Account, vipLink string) error {
    // Parse VIP link
    linkData, err := roblox.ParseVipLink(vipLink)
    if err != nil {
        return fmt.Errorf("invalid VIP link: %v", err)
    }

    // Get CSRF Token
    csrfToken, err := roblox.GetCSRFToken(acc.Cookie)
    if err != nil {
        return fmt.Errorf("failed to get CSRF token: %v", err)
    }

    // Get Auth Ticket
    authTicket, err := roblox.GetAuthTicket(acc.Cookie, csrfToken)
    if err != nil {
        return fmt.Errorf("failed to get auth ticket: %v", err)
    }

    // Get Access Code
    accessCode, err := roblox.GetAccessCode(acc.Cookie, linkData.PlaceID, linkData.LinkCode)
    if err != nil {
        return fmt.Errorf("failed to get access code: %v", err)
    }

    // Launch Roblox
    if err := roblox.LaunchRoblox(authTicket, linkData.PlaceID, accessCode, linkData.LinkCode); err != nil {
        return fmt.Errorf("failed to launch Roblox: %v", err)
    }

    return nil
}