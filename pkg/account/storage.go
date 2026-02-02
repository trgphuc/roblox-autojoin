package account

import (
    "encoding/json"
    "errors"
    "os"
    "path/filepath"
    "strings"
)

type Storage struct {
    filePath string
}

func NewStorage() (*Storage, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }

    appDir := filepath.Join(homeDir, ".roblox-autojoin")
    if err := os.MkdirAll(appDir, 0755); err != nil {
        return nil, err
    }

    return &Storage{
        filePath: filepath.Join(appDir, "accounts.json"),
    }, nil
}

func (s *Storage) GetAllAccounts() ([]*Account, error) {
    if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
        return []*Account{}, nil
    }

    data, err := os.ReadFile(s.filePath)
    if err != nil {
        return nil, err
    }

    var accounts []*Account
    if err := json.Unmarshal(data, &accounts); err != nil {
        return nil, err
    }

    return accounts, nil
}

func (s *Storage) AddAccount(acc *Account) error {
    accounts, err := s.GetAllAccounts()
    if err != nil {
        return err
    }

    // Check if account already exists
    for _, a := range accounts {
        if strings.EqualFold(a.Name, acc.Name) {
            return errors.New("account already exists")
        }
    }

    accounts = append(accounts, acc)
    return s.saveAccounts(accounts)
}

func (s *Storage) RemoveAccount(name string) error {
    accounts, err := s.GetAllAccounts()
    if err != nil {
        return err
    }

    found := false
    newAccounts := make([]*Account, 0)
    for _, acc := range accounts {
        if !strings.EqualFold(acc.Name, name) {
            newAccounts = append(newAccounts, acc)
        } else {
            found = true
        }
    }

    if !found {
        return errors.New("account not found")
    }

    return s.saveAccounts(newAccounts)
}

func (s *Storage) saveAccounts(accounts []*Account) error {
    data, err := json.MarshalIndent(accounts, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(s.filePath, data, 0644)
}