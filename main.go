package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var App *firebase.App
var Ctx context.Context

type Log struct {
	Timestamp string `json:"timestamp"`
	Fields    struct {
		Client        string `json:"client"`
		RemoteUser    string `json:"remote_user"`
		XForwardedFor string `json:"x_forwarded_for"`
		HitMiss       string `json:"hit_miss"`
		Bytes         int    `json:"bytes"`
		DurationUsec  int    `json:"duration_usec"`
		Status        int    `json:"status"`
		Request       string `json:"request"`
		Virtualhost   string `json:"virtualhost"`
		Method        string `json:"method"`
		TimeFirstByte string `json:"time_first_byte"`
		Handling      string `json:"handling"`
		Referrer      string `json:"referrer"`
		UserAgent     string `json:"user_agent"`
	} `json:"fields"`
}

func (l *Log) Parse(s string) error {
	bytes := []byte(s)

	return json.Unmarshal(bytes, &l)
}

func varnishStat() {

	cmd := exec.Command("/bin/sh", "./ncsa.sh")

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		return
	}

	scanner := bufio.NewScanner(cmdReader)
	client := connFirestore()
	defer client.Close()
	go func() {
		for scanner.Scan() {
			log.Print("Error caught")
			log.Print(scanner.Text())
			sendLog500(client, scanner.Text())
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

func sendLog500(client *firestore.Client, log string) {

	logStruct := Log{}
	if err := logStruct.Parse(log); err != nil {
		panic(err)
	}

	ctx := context.Background()
	_, _, err := client.Collection("errors-500").Add(ctx, logStruct)
	if err != nil {
		panic(err)
	}
}

func connFirebase() {
	log.Print("Connecting Firabase")
	Ctx := context.Background()
	opt := option.WithCredentialsFile("monitor-key.json")
	app, err := firebase.NewApp(Ctx, nil, opt)
	if err != nil {
		panic(err)
	}
	App = app
	log.Print("Firebase Connected")
}

func connFirestore() *firestore.Client {
	log.Print("Connecting Firestore and creating client")
	ctx := context.Background()
	client, err := App.Firestore(ctx)
	if err != nil {
		panic(err)
	}
	log.Print("Firestone Connected")
	return client
}

func main() {
	connFirebase()
	varnishStat()
}
