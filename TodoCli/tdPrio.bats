#!/usr/bin/env bats

setup() {
  touch 'item1 tst.txt'
  touch '(A) item2 tst.txt'
  touch '(B) item3 tst.txt'
}

teardown() {
  rm *tst.txt
}

@test "set priority to nonpriorotized file" {
  ./tdPrio A 'item1 tst.txt'
  [ -e '(A) item1 tst.txt' ]
}

@test "set priority for another prioritized file" {
  ./tdPrio A '(B) item3 tst.txt'   
  [ -e '(A) item3 tst.txt' ]
} 

@test "unset priority" {
  ./tdPrio _ '(B) item3 tst.txt'   
  [ -e 'item3 tst.txt' ]
} 

@test "multi rename" {
  ./tdPrio C 'item1 tst.txt' '(A) item2 tst.txt' '(B) item3 tst.txt'
  [ -e '(C) item1 tst.txt' ]
  [ -e '(C) item2 tst.txt' ]
  [ -e '(C) item3 tst.txt' ]
}
