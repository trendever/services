package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"time"
	"utils/rpc"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/net/context"
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

	err = yaml.Unmarshal(bytes, argument.Interface())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not decode:", err.Error())
		os.Exit(1)
	}

	fmt.Printf("%#v\n", argument.Interface())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	fmt.Println("Do calling")
	then := time.Now()

	result := call.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		argument,
	})

	fmt.Printf("Request took %v\n", time.Since(then).String())
	fmt.Printf("Error is: %v\n", result[1].Interface())
	spew.Dump(result[0].Interface())
}
