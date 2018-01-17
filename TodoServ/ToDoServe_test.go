package main

import "testing"

func TestRemoveAllPrio(t *testing.T) {
    if RemoveAllPrio("ahoj") != "ahoj" {
        t.Error("Expectest unchanged input.")
    }
    if RemoveAllPrio("(A) ahoj") != "ahoj" {
        t.Error("Expectest changed input A.")
    }
}
