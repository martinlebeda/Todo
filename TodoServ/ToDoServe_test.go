package main

import (
    "testing"
)

func TestRemoveAllPrio(t *testing.T) {
    if RemoveAllPrio("ahoj") != "ahoj" {
        t.Error("Expectest unchanged input.")
    }
    if RemoveAllPrio("(A) ahoj") != "ahoj" {
        t.Error("Expectest changed input A.")
    }
}


func TestCheckFilterItem(t *testing.T) {
    if !CheckFilterItem("", "(A) item") { t.Error("empty search") }
    if !CheckFilterItem("(A)", "(A) item") { t.Error("simple prefix") }
    if !CheckFilterItem("~([A])", "(A) item") { t.Error("regex prefix") }
    if !CheckFilterItem("~([AB])", "(B) item") { t.Error("regex B prefix") }
    if !CheckFilterItem("~([AB]).*", "(A) item") { t.Error("regex AB prefix") }
    if CheckFilterItem("~([AB]).*", "90 otestovat a vygenerovat dokumentaci.txt") { t.Error("regex AB nonprefix") }

    //matched, err := regexp.MatchString("^\\([AB]\\) ", "90 otestovat a vygenerovat dokumentaci.txt")
    //fmt.Println(matched)
    //fmt.Println(err)
                                                                          //
}

func TestIsRoot(t *testing.T) {
    if !isRoot("/home/martin/Todo/tst", "/home/martin/Todo") { t.Error("path is in root") }
    if isRoot("/home/martin/Todo/tst/aaa", "/home/martin/Todo") { t.Error("path is not in root") }
}
