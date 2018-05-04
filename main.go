package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func varnishStat() {

	cmd := exec.Command("/bin/sh", "./ncsa.sh")

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		return
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			sendLog(scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		return
	}
}

func sendLog(log string) {
	fmt.Println(log)
}

func main() {
	ctx := context.Background()
	opt := option.WithCredentialsFile("monitor-key.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		panic(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	if err != nil {
		log.Fatalf("Failed adding alovelace: %v", err)
	}

	varnishStat()
}
