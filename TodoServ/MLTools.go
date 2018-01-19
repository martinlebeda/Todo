package main

import (
    "strings"
    "golang.org/x/text/unicode/norm"
    "golang.org/x/text/transform"
    "unicode"
    "os"
    "io"
    "golang.org/x/text/runes"
)

// Remove diacritics and make lowercase.
// http://stackoverflow.com/questions/26722450/remove-diacritics-using-go
func NormalizeString(s string) string {
    l := strings.ToLower(strings.TrimSpace(s))
    t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
    n, _, _ := transform.String(t, l)
    return n
}

func IsEmpty(name string) (bool, error) {
    f, err := os.Open(name)
    if err != nil {
        return false, err
    }
    defer f.Close()

    _, err = f.Readdirnames(1) // Or f.Readdir(1)
    if err == io.EOF {
        return true, nil
    }
    return false, err // Either not empty or error, suits both cases
}