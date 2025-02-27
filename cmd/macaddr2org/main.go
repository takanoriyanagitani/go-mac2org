package main

import (
	"context"
	"fmt"
	"iter"
	"log"
	"os"
	"strconv"

	mo "github.com/takanoriyanagitani/go-mac2org"
	. "github.com/takanoriyanagitani/go-mac2org/util"
)

var envValByKey func(key string) IO[string] = Lift(
	func(key string) (string, error) {
		val, found := os.LookupEnv(key)
		switch found {
		case true:
			return val, nil
		default:
			return "", fmt.Errorf("env var %s missing", key)
		}
	},
)

var listNames IO[bool] = Bind(
	envValByKey("ENV_LIST_NAMES"),
	Lift(strconv.ParseBool),
).Or(Of(false))

var getNames func() iter.Seq[string] = mo.GetNames

func printNames(names iter.Seq[string]) IO[Void] {
	return func(ctx context.Context) (Void, error) {
		for name := range names {
			select {
			case <-ctx.Done():
				return Empty, ctx.Err()
			default:
			}

			fmt.Println(name)
		}

		return Empty, nil
	}
}

var names2stdout IO[Void] = Bind(
	Of(getNames()),
	printNames,
)

var macPrefix2org mo.MacPrefixToOrg = mo.MacPrefixToOrgDefault

var mac2org IO[Void] = func(ctx context.Context) (Void, error) {
	return Empty, macPrefix2org.StdinToMacStringsToStdout(ctx)
}

var convertOrList IO[Void] = Bind(
	listNames,
	func(list bool) IO[Void] {
		switch list {
		case true:
			return names2stdout
		default:
			return mac2org
		}
	},
)

var sub IO[Void] = func(ctx context.Context) (Void, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return convertOrList(ctx)
}

func main() {
	_, e := sub(context.Background())
	if nil != e {
		log.Printf("%v\n", e)
	}
}
