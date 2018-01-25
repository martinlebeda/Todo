#!/usr/bin/env bats

setup() {
  touch 'item1 tst.txt'
  touch '(A) item2 tst.txt'
  touch '(B) item3 tst.txt'
  touch 'x item4 tst.txt'
}

teardown() {
  rm *tst.txt
}

@test "set done to nonpriorotized file" {
  ./tdDone 'item1 tst.txt'
  [ -e 'x item1 tst.txt' ]
}

@test "set done for another prioritized file" {
  ./tdDone A '(B) item3 tst.txt'   
  [ -e 'x item3 tst.txt' ]
} 

@test "unset done" {
  ./tdDone 'x item4 tst.txt'   
  [ -e 'item4 tst.txt' ]
} 

@test "multi done" {
  ./tdDone 'item1 tst.txt' '(A) item2 tst.txt' '(B) item3 tst.txt'
  [ -e 'x item1 tst.txt' ]
  [ -e 'x item2 tst.txt' ]
  [ -e 'x item3 tst.txt' ]
}
