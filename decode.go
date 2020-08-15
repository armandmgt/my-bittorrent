package main

import (
	"fmt"
	"regexp"
	"strconv"
)

type bEncodeValue struct {
	dict    *bEncodeDict
	list    *bEncodeList
	integer *int
	str     *string
}

type bEncodeKeyValue struct {
	key   *string
	value *bEncodeValue
}

type bEncodeDict struct {
	items []bEncodeKeyValue
}

type bEncodeList []bEncodeValue

func tokenize(contents string, tokens chan string) {
	singleTokens := []string{"i", "d", "l", "e"}
	r := regexp.MustCompile(`([idle])|(\d+):|(-?\d+)`)

	for i := 0; i < len(contents); {
		tok := r.FindStringSubmatch(contents[i:])
		switch {
		case Contains(singleTokens, tok[1]):
			tokens <- tok[0]
			i++
		case tok[2] == "":
			tokens <- tok[3]
			i += len(tok[3])
		default:
			tokens <- "s"
			i += len(tok[0])
			length, _ := strconv.Atoi(tok[2])
			tokens <- contents[i : i+length]
			i += length
		}
	}
	tokens <- ""
}

func data(value *bEncodeValue, token string, tokens chan string, infoTokens chan string, inInfo bool) {
	if inInfo {
		infoTokens <- token
	}
	switch token {
	case "i":
		token = <-tokens
		integer, _ := strconv.Atoi(token)
		value.integer = &integer
		token = <-tokens
	case "s":
		token = <-tokens
		value.str = &token
	case "l":
		value.list = &bEncodeList{}
		token = <-tokens
		for ; token != "e"; token = <-tokens {
			val := bEncodeValue{}
			data(&val, token, tokens, infoTokens, inInfo)
			newList := append(*value.list, val)
			value.list = &newList
		}
	case "d":
		startInfo := false
		value.dict = &bEncodeDict{}
		token = <-tokens
		for ; token != "e"; token = <-tokens {
			key := bEncodeValue{}
			data(&key, token, tokens, infoTokens, inInfo)
			if *key.str == "info" {
				startInfo = true
			}
			token = <-tokens
			val := bEncodeValue{}
			data(&val, token, tokens, infoTokens, inInfo || startInfo)
			value.dict.items = append(value.dict.items, bEncodeKeyValue{key.str, &val})
		}
		if startInfo {
			infoTokens <- ""
		}
	}
}

func getDictValue(dict *bEncodeDict, key string) (*bEncodeValue, error) {
	for i := 0; i < len(dict.items); i++ {
		if *dict.items[i].key == key {
			return dict.items[i].value, nil
		}
	}
	err := fmt.Errorf("invalid torrent file (no '%s')", key)
	return nil, err
}
