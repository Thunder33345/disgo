package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	"github.com/DiscoOrg/disgo/api"
	"github.com/DiscoOrg/disgo/api/constants"
	"github.com/DiscoOrg/disgo/api/endpoints"
	"github.com/DiscoOrg/disgo/api/models"
)

// GatewayImpl is what is used to connect to discord
type GatewayImpl struct {
	DisgoClient           api.Disgo
	conn                  *websocket.Conn
	connectionStatus      constants.ConnectionStatus
	heartbeatInterval     int
	lastHeartbeatSent     time.Time
	lastHeartbeatReceived time.Time
	sessionID             string
	lastSequenceReceived  *int
	url                   *string
}

// Close cleans up the gateway internals
func (g GatewayImpl) Close() {
	log.Info("Implement closing smh...")
}

// Disgo returns the gateway's disgo client
func (g GatewayImpl) Disgo() api.Disgo {
	return g.DisgoClient
}

// Open initializes the client and connection to discord
func (g GatewayImpl) Open() error {
	g.connectionStatus = constants.Connecting
	log.Info("starting ws...")

	gatewayBase := "wss://gateway.discord.gg"
	g.url = &gatewayBase

	if g.url == nil {
		log.Println("GetGateway url is nil, fetching...")
		gatewayRs := models.GatewayRs{}
		if err := g.Disgo().RestClient().Request(endpoints.GetGateway, nil, &gatewayRs); err != nil {
			return err
		}
		g.url = &gatewayRs.URL
	}

	gatewayUrl := *g.url + "?v=" + endpoints.APIVersion + "&encoding=json"
	wsConn, _, err := websocket.DefaultDialer.Dial(gatewayUrl, nil)
	if err != nil {
		log.Errorf("error connecting to gateway. url: %s, error: %s", gatewayUrl, err.Error())
		return err
	}
	wsConn.SetCloseHandler(func(code int, error string) error {
		log.Errorf("connection to websocket closed with code: %d, error: %s", code, error)
		return nil
	})

	g.conn = wsConn
	g.connectionStatus = constants.WaitingForHello

	mt, data, err := g.conn.ReadMessage()
	if err != nil {
		return err
	}
	event, err := parseGatewayEvent(mt, data)
	if err != nil {
		return err
	}
	if event.Op != constants.OpHello {
		return fmt.Errorf("expected op: hello type: 10, received: %d", mt)
	}

	g.lastHeartbeatReceived = time.Now().UTC()

	var eventData models.HelloCommand
	if err = json.Unmarshal(event.D, &eventData); err != nil {
		return err
	}

	g.connectionStatus = constants.Identifying
	g.heartbeatInterval = eventData.HeartbeatInterval

	if err = wsConn.WriteJSON(models.IdentifyCommand{
		UnresolvedGatewayCommand: models.UnresolvedGatewayCommand{
			Op: constants.OpIdentify,
		},
		D: models.IdentifyCommandData{
			Token: g.Disgo().Token(),
			Properties: models.OpIdentifyDataProperties{
				OS:      api.GetOS(),
				Browser: "disgo",
				Device:  "disgo",
			},
			Compress:       false,
			LargeThreshold: 50,
			Intents:        g.Disgo().Intents(),
		},
	}); err != nil {
		return err
	}

	g.connectionStatus = constants.WaitingForReady

	go g.heartbeat()
	go g.listen()

	return nil
}

func (g GatewayImpl) heartbeat() {
	defer func() {
		log.Info("Shutting down heartbeat...")
	}()

	for {
		time.Sleep(time.Duration(g.heartbeatInterval) * time.Millisecond)
		g.sendHeartbeat()
	}
}

func (g GatewayImpl) sendHeartbeat() {
	log.Info("sending heartbeat...")

	err := g.conn.WriteJSON(models.HeartbeatCommand{
		UnresolvedGatewayCommand: models.UnresolvedGatewayCommand{
			Op: constants.OpHeartbeat,
		},
		D: g.lastSequenceReceived,
	})
	if err != nil {
		log.Errorf("failed to send heartbeat with error: %s", err)
		_ = g.conn.Close()
		// Todo: reconnect
	}
	g.lastHeartbeatSent = time.Now().UTC()
}

func (g GatewayImpl) listen() {
	defer func() {
		log.Info("Shutting down listen...")
	}()
	for {
		mt, data, err := g.conn.ReadMessage()
		if err != nil {
			log.Errorf("error while reading from ws. error: %s", err)
		}

		event, err := parseGatewayEvent(mt, data)
		if err != nil {
			log.Errorf("error while unpacking gateway event. error: %s", err)
		}

		switch op := event.Op; op {

		case constants.OpDispatch:
			//log.Infof("received: OpDispatch")
			if event.S != nil {
				g.lastSequenceReceived = event.S
			}

			if event.T != nil && *event.T == "READY" {
				var readyEvent models.ReadyEventData
				if err := parseEventToStruct(event, &readyEvent); err != nil {
					return
				}
				g.sessionID = readyEvent.SessionID
				g.Disgo().SetSelfUser(readyEvent.User)
				log.Info("Client Ready")
			}

			if event.T == nil {
				log.Errorf("received event without T. playload: %s", string(data))
				continue
			}
			g.Disgo().EventManager().Handle(*event.T, event.D)

		case constants.OpHeartbeat:
			log.Infof("received: OpHeartbeat")
			g.sendHeartbeat()

		case constants.OpReconnect:
			log.Infof("received: OpReconnect")

		case constants.OpInvalidSession:
			log.Infof("received: OpInvalidSession")

		case constants.OpHeartbeatACK:
			log.Infof("received: OpHeartbeatACK")
			g.lastHeartbeatReceived = time.Now().UTC()
		}
	}
}

func parseEventToStruct(event *models.GatewayCommand, v interface{}) error {
	if err := json.Unmarshal(event.D, v); err != nil {
		log.Errorf("error while unmarshaling event. error: %s", err)
		return err
	}
	return nil
}

func parseGatewayEvent(mt int, data []byte) (*models.GatewayCommand, error) {

	var reader io.Reader = bytes.NewBuffer(data)

	if mt == websocket.BinaryMessage {
		return nil, errors.New("we don't handle compressed yet")
	}
	if mt != websocket.TextMessage {
		return nil, fmt.Errorf("recieved unexpected message type: %d", mt)
	}
	var event models.GatewayCommand

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&event); err != nil {
		log.Errorf("error decoding websocket message, %s", err)
		return nil, err
	}
	return &event, nil
}
