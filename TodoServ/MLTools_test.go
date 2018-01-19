package main

import "testing"

func TestNormalizeString(t *testing.T) {
    if NormalizeString("ahoj") != "ahoj" {
        t.Error("Expectest unchanged input.")
    }

    if NormalizeString(" ahoj ") != "ahoj" {
        t.Error("Normalize spaces")
    }

    if NormalizeString(" AHoj ") != "ahoj" {
        t.Error("Normalize CamelCase")
    }

    if NormalizeString(" Příliš žluťoučký kůň ") != "prilis zlutoucky kun" {
        t.Error("Normalize accents")
    }

    //if NormalizeString(" ahoj ") != "ahoj" {
    //    t.Error("Normalize spaces")
    //}
}
