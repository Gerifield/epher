package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type config struct {
	Routes []route `json:"routes"`
}
type route struct {
	WSAddress       string `json:"ws_address"`
	HTTPURL         string `json:"http_url"`
	ForwardResponse bool   `json:"forward_response"`
}

func main() {
	configFile := flag.String("configFile", "config.json", "Config file for routing")
	flag.Parse()

	f, err := os.Open(*configFile)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	var config config
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		log.Fatalln(err)
	}

	var wg sync.WaitGroup
	log.Printf("connecting to %d endpoints\n", len(config.Routes))
	log.Println("Routes:", printRoutes(config.Routes))

	wg.Add(len(config.Routes))
	for _, r := range config.Routes {
		go func(r route) {
			defer wg.Done()
			if err := handleWSConnection(r); err != nil {
				log.Printf("routing error in %s, err: %s", r.WSAddress, err.Error())
			}
		}(r)
	}
	wg.Wait()
	log.Println("Stopped")
}

func printRoutes(routes []route) string {
	sb := strings.Builder{}
	sb.WriteString("\n") // Meh
	for _, r := range routes {
		sb.WriteString(r.WSAddress)
		sb.WriteString(" -> ")
		sb.WriteString(r.WSAddress)
		sb.WriteString(" (forwarding: ")
		if r.ForwardResponse {
			sb.WriteString("enabled")
		} else {
			sb.WriteString("disabled")
		}
		sb.WriteString(")")
		sb.WriteString("\n")
	}
	return sb.String()
}

func handleWSConnection(r route) error {
	conn, _, err := websocket.DefaultDialer.Dial(r.WSAddress, nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		_, b, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return nil
			}
			return err
		}

		// For logging
		//log.Printf("send message to %s, payload %s\n", r.HTTPURL, string(b))
		resp, err := sendHTTPPayload(r.HTTPURL, b)
		if err != nil {
			log.Printf("http send failed, err: %s\n", err.Error())
			continue
		}

		if r.ForwardResponse {
			// TODO: determine message type based on the response
			err = conn.WriteMessage(websocket.TextMessage, resp)
			if err != nil {
				log.Printf("websocket next writer failed, err: %s\n", err.Error())
				continue
			}
		}
	}
}

func sendHTTPPayload(url string, data []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
