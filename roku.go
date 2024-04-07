package main

import (
	"github.com/koron/go-ssdp"
	"log"
	"net/http"
)

func FindRoku() []ssdp.Service {
	rokus := []ssdp.Service{}
	list, err := ssdp.Search(ssdp.All, 15, "239.255.255.250:1900")
	if err != nil {
		log.Fatal(err)
	}
	for _, srv := range list {
		//fmt.Printf("%d: %#v\n", i, srv)
		if srv.Type == "roku:ecp" {
			//can't handle multiple roku devices on the network. Which is sort of important.
			rokus = append(rokus, srv)
		}
	}
	return rokus
}

func SendCommand(roku ssdp.Service, command string) {
	if roku.Location == "" {
		return
	}
	r, err := http.Post(roku.Location+"/"+command, "text/plain", nil)
	defer r.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
}

//keypress key values:
//"keypress/{}"
//Home
//Rev
//Fwd
//Play
//Select
//Left
//Right
//Down
//Up
//Back
//InstantReplay
//Info
//Backspace
//Search
//Enter
