#!/usr/bin/env bash

function red {
    echo -e "\033[0;31;40m$1\033[0m"
}

function green {
    echo -e "\033[0;32;40m$1\033[0m"
}

TMP_DIR=_temp

mkdir $TMP_DIR

function clean_up {
    rm -rf $TMP_DIR
}

echo "Compiling program"

go build -o $TMP_DIR/main src/main.go || {
    echo "Error during compilation. Skipping tests."
    clean_up
    exit 1
}

echo "Running tests..."

declare -i TOTAL; TOTAL+=0
declare -i PASSED; PASSED+=0
declare -i FAILED; FAILED+=0

function run_test_match {
    TOTAL+=1
    echo ""
    echo "'$1' should match '$2'"
    echo "$1" | $TMP_DIR/main --silent -E "$2" && {
        green "    Passed"
        PASSED+=1
    } || {
        red "    Failed"
        echo "    Debug: echo '$1' | go run src/main.go -- --debug -E '$2'"
        FAILED+=1
    }
}

function run_test_no_match {
    TOTAL+=1
    echo ""
    echo "'$1' should not match '$2'"
    echo "$1" | $TMP_DIR/main --silent -E "$2" && {
        red "    Failed"
        echo "    Debug: echo '$1' | go run src/main.go -- --debug -E '$2'"
        FAILED+=1
    } || {
        green "    Passed"
        PASSED+=1
    }
}

run_test_match "aaabbbccc" "bc"
run_test_no_match "aaabbbccc" "bcccc"

run_test_match "apple" "[^xyz]"
run_test_no_match "banana" "[^anb]"

run_test_match "a" "[abcd]"
run_test_no_match "efgh" "[abcd]"

run_test_match "word" "\w"
run_test_no_match '$!?' "\w"

run_test_match "123" "\d"
run_test_no_match "apple" "\d"

run_test_match "dog" "d"
run_test_match "1 apple" "\d apple"

run_test_match "100 apple" "\d\d\d apple"
run_test_no_match "1 apples" "\d\d\d apple"
run_test_match "3 dogs" "\d \w\w\ws"
run_test_no_match "1 dog" "\d \w\w\ws"

run_test_match "log" "^log"
run_test_no_match "xlog" "^log"

run_test_match "dog" 'dog$'
run_test_no_match "dogs" 'dog$'

run_test_match "banana" 'a+'
run_test_match "SaaS" 'a+'
run_test_match "caaaats" 'ca+t'
run_test_no_match "dog" 'a+'
run_test_match "grep!" 'g\w+p'
run_test_match "grep!" 'g\w+[pd]'

run_test_match "dogs" 'dogs?'
run_test_match "dog" 'dogs?'
run_test_match "dogsx" 'dogs?x'
run_test_match "dogx" 'dogs?x'
run_test_no_match "dogssx" 'dogs?x'

run_test_match "dog" 'd.g'
run_test_no_match "doc" 'd.g'

run_test_match "cat" '(cat|dog)'
run_test_no_match "apple" '(cat|dog)'
run_test_match "batman" 'bat(girl|man)'
run_test_no_match "batmobile" 'bat(girl|man)'

run_test_match "batman and catman" 'bat(man|dog) and cat\1'
run_test_no_match "batman and catdog" 'bat(man|dog) and cat\1'

run_test_match "batman and catman and coolcat" 'bat(man|dog) and (cat)\1 and cool\2'
run_test_no_match "batman and catdog and coolmat" 'bat(man|dog) and (cat)\1 and cool\2'

echo ""
echo "Passed: $PASSED/$TOTAL"
echo "Failed: $FAILED/$TOTAL"
echo ""

echo "Cleaning up"
clean_up

echo "Finished"
