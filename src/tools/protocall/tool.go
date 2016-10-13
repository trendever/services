package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"utils/rpc"

	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

var conn *grpc.ClientConn

func init() {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: %v ADDR:PORT WhateverService MethodName\n", os.Args[0])
		os.Exit(1)
	}

	conn = rpc.Connect(os.Args[1])
	connect()

}

func main() {

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	serviceName, callName := os.Args[2], os.Args[3]

	client, found := services[serviceName]
	if !found {
		fmt.Fprintln(os.Stderr, "No such service found")
		os.Exit(1)
	}

	call := reflect.ValueOf(client).MethodByName(callName)

	if !call.IsValid() {
		fmt.Fprintln(os.Stderr, "No such call found")
		os.Exit(1)
	}

	argument := reflect.New(call.Type().In(1).Elem())

	fmt.Printf("%#v\n", argument.Interface())

	yaml.Unmarshal(bytes, argument.Interface())

	ctx, cancel := rpc.DefaultContext()
	defer cancel()

	fmt.Println("Do calling")

	result := call.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		argument,
	})

	fmt.Printf("Error is: %v", result[0].Interface())
	spew.Dump(result[1].Interface())
}
