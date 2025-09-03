package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BatchMaxAmount int
	Nombre        string
	Apellido      string
	Documento     string
	Nacimiento    string
	Numero        string
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	shutdown bool
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:   config,
		shutdown: false,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send bet data as batches to server using CSV data
func (c *Client) StartClientLoop() {
	err := c.createClientSocket()
	if err != nil {
		log.Errorf("action: create_persistent_connection | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	defer c.closeConnection()

	csvFile := fmt.Sprintf("/data/agency-%s.csv", c.config.ID)
	if _, err := os.Stat(csvFile); err == nil {
		c.processBatchesFromCSV(csvFile)
	} else {
		log.Infof("action: csv_file_not_found | result: fail | client_id: %v | file: %s", c.config.ID, csvFile)
	}
}

func (c *Client) processBatchesFromCSV(csvFile string) {
	err := c.processBetsFromCSVAsBatchesAndSend(csvFile)
	if err != nil {
		log.Errorf("action: read_csv_streaming | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	
	c.sendFinishedNotification()
	
	c.queryWinners()
}

func (c *Client) processBetsFromCSVAsBatchesAndSend(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	maxBatchSize := c.config.BatchMaxAmount

	var currentBatch []BetData
	totalBets := 0
	rowNumber := 0

	for {
		if c.shutdown {
			return nil
		}

		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				if len(currentBatch) > 0 {
					c.sendBatch(currentBatch)
				}
				break
			}
			return fmt.Errorf("failed to read CSV row %d: %v", rowNumber, err)
		}

		rowNumber++

		if len(record) < 5 {
			log.Errorf("action: parse_csv_row | result: fail | row: %d | error: insufficient columns", rowNumber)
			continue
		}

		bet := BetData{
			ClientID:   c.config.ID,
			Nombre:     strings.TrimSpace(record[0]),
			Apellido:   strings.TrimSpace(record[1]),
			Documento:  strings.TrimSpace(record[2]),
			Nacimiento: strings.TrimSpace(record[3]),
			Numero:     strings.TrimSpace(record[4]),
		}

		currentBatch = append(currentBatch, bet)
		totalBets++

		if len(currentBatch) >= maxBatchSize {
			c.sendBatch(currentBatch)
			currentBatch = nil
		}
	}

	log.Infof("action: csv_processing | result: success | client_id: %v | total_bets: %d", c.config.ID, totalBets)
	return nil
}

func (c *Client) sendBatch(bets []BetData) {
	if c.conn == nil || c.shutdown {
		return
	}

	batch := BatchData{
		ClientID: c.config.ID,
		Bets:     bets,
	}

	err := SendBetBatch(c.conn, batch)
	if err != nil {
		log.Errorf("action: send_bet_batch | result: fail | client_id: %v | batch_size: %d | error: %v", 
			c.config.ID, len(bets), err)
		return
	}

	success, err := ReceiveResponse(c.conn)
	if err != nil {
		log.Errorf("action: receive_response | result: fail | client_id: %v | batch_size: %d | error: %v", 
			c.config.ID, len(bets), err)
		return
	}

	if success {
		log.Infof("action: batch_sent | result: success | client_id: %v | batch_size: %d", 
			c.config.ID, len(bets))
	} else {
		log.Errorf("action: batch_sent | result: fail | client_id: %v | batch_size: %d | response: ERROR", 
			c.config.ID, len(bets))
	}
}

func (c *Client) Shutdown() {
	log.Infof("action: shutdown_initiated | result: success | client_id: %v", c.config.ID)
	c.shutdown = true
	c.closeConnection()
}

func (c *Client) closeConnection() {
	if c.conn != nil {
		log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
		c.conn.Close()
		c.conn = nil
	}
}

func (c *Client) sendFinishedNotification() {
	if c.conn == nil || c.shutdown {
		return
	}

	err := SendFinishedNotification(c.conn, c.config.ID)
	if err != nil {
		log.Errorf("action: send_finished_notification | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	success, err := ReceiveResponse(c.conn)
	if err != nil {
		log.Errorf("action: receive_finished_response | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	if success {
		log.Infof("action: finished_notification_sent | result: success | client_id: %v", c.config.ID)
	} else {
		log.Errorf("action: finished_notification_sent | result: fail | client_id: %v | response: ERROR", c.config.ID)
	}
}

func (c *Client) queryWinners() {
	const retryDelay = 1 * time.Second
	
	attempt := 1
	for !c.shutdown {
		if c.conn == nil {
			return
		}

		err := SendQueryWinners(c.conn, c.config.ID)
		if err != nil {
			log.Errorf("action: send_query_winners | result: fail | client_id: %v | attempt: %d | error: %v", c.config.ID, attempt, err)
			return
		}

		winners, err := ReceiveWinners(c.conn)
		if err != nil {
			log.Debugf("action: receive_winners | result: fail | client_id: %v | attempt: %d | error: %v", c.config.ID, attempt, err)
			
			time.Sleep(retryDelay)
			attempt++
			continue
		}

		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d", len(winners))
		return
	}
}
