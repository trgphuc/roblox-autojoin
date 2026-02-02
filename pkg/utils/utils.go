package utils

import (
    "strings"
)

func TrimCookie(cookie string) string {
    return strings.TrimSpace(cookie)
}

func ValidateCookie(cookie string) bool {
    return strings.HasPrefix(cookie, "_|WARNING:")
}