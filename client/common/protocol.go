package common

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

var protocolLog = logging.MustGetLogger("protocol")

const (
	ResponseOK = 0
	ResponseError = 1
)

const (
	MessageTypeBatch = 1
	MessageTypeFinishedSending = 2
	MessageTypeQueryWinners = 3
)

type BetData struct {
	ClientID   string
	Nombre     string
	Apellido   string
	Documento  string
	Nacimiento string
	Numero     string
}

type BatchData struct {
	ClientID string
	Bets     []BetData
}

func calculateBatchMessageSize(batch BatchData) int {
	// client_id (4 bytes) + batch_size (4 bytes)
	size := 4 + 4
	
	// For each bet: nombre_len(4) + nombre + apellido_len(4) + apellido + documento(4) + nacimiento(4) + numero(4)
	for _, bet := range batch.Bets {
		size += 4 + len(bet.Nombre) // nombre_len + nombre
		size += 4 + len(bet.Apellido) // apellido_len + apellido
		size += 4 // documento
		size += 4 // nacimiento
		size += 4 // numero
	}
	
	return size
}

// Protocol: total_message_length(4), client_id(4), batch_size(4), [bet1, bet2, ...]
// Where each bet: nombre_len(4), nombre, apellido_len(4), apellido, documento(4), nacimiento(4), numero(4)
func SendBetBatch(conn net.Conn, batch BatchData) error {
	// Message type (4 bytes)
	err := binary.Write(conn, binary.BigEndian, uint32(MessageTypeBatch))
	if err != nil {
		protocolLog.Errorf("action: send_bet_batch | result: fail | field: message_type | error: %v", err)
		return err
	}

	messageSize := calculateBatchMessageSize(batch)
	
	err = binary.Write(conn, binary.BigEndian, uint32(messageSize))
	if err != nil {
		protocolLog.Errorf("action: send_bet_batch | result: fail | field: message_length | error: %v", err)
		return err
	}

	// Client ID (4 bytes)
	clientID, err := strconv.ParseUint(batch.ClientID, 10, 32)
	if err != nil {
		protocolLog.Errorf("action: send_bet_batch | result: fail | field: client_id | error: %v", err)
		return fmt.Errorf("invalid client ID: %v", err)
	}
	err = binary.Write(conn, binary.BigEndian, uint32(clientID))
	if err != nil {
		protocolLog.Errorf("action: send_bet_batch | result: fail | field: client_id | error: %v", err)
		return err
	}

	// Batch size (4 bytes)
	batchSize := uint32(len(batch.Bets))
	err = binary.Write(conn, binary.BigEndian, batchSize)
	if err != nil {
		protocolLog.Errorf("action: send_bet_batch | result: fail | field: batch_size | error: %v", err)
		return err
	}

	for i, bet := range batch.Bets {
		err := sendSingleBet(conn, bet)
		if err != nil {
			protocolLog.Errorf("action: send_bet_batch | result: fail | bet_index: %d | error: %v", i, err)
			return fmt.Errorf("failed to send bet %d: %v", i, err)
		}
	}

	protocolLog.Debugf("action: send_bet_batch | result: success | client_id: %s | batch_size: %d | message_size: %d", batch.ClientID, batchSize, messageSize)
	return nil
}

func sendSingleBet(conn net.Conn, bet BetData) error {
	// Nombre length (4 bytes) and Nombre
	nameBytes := []byte(bet.Nombre)
	nameLen := uint32(len(nameBytes))
	err := binary.Write(conn, binary.BigEndian, nameLen)
	if err != nil {
		protocolLog.Errorf("action: send_single_bet | result: fail | field: nombre_length | error: %v", err)
		return err
	}
	_, err = conn.Write(nameBytes)
	if err != nil {
		protocolLog.Errorf("action: send_single_bet | result: fail | field: nombre | error: %v", err)
		return err
	}

	// Apellido length (4 bytes) and Apellido  
	surnameBytes := []byte(bet.Apellido)
	surnameLen := uint32(len(surnameBytes))
	err = binary.Write(conn, binary.BigEndian, surnameLen)
	if err != nil {
		protocolLog.Errorf("action: send_single_bet | result: fail | field: apellido_length | error: %v", err)
		return err
	}
	_, err = conn.Write(surnameBytes)
	if err != nil {
		protocolLog.Errorf("action: send_single_bet | result: fail | field: apellido | error: %v", err)
		return err
	}

	// Documento (4 bytes)
	documento, err := strconv.ParseUint(bet.Documento, 10, 32)
	if err != nil {
		protocolLog.Errorf("action: send_single_bet | result: fail | field: documento | error: %v", err)
		return fmt.Errorf("invalid documento: %v", err)
	}
	err = binary.Write(conn, binary.BigEndian, uint32(documento))
	if err != nil {
		protocolLog.Errorf("action: send_single_bet | result: fail | field: documento | error: %v", err)
		return err
	}

	// Nacimiento (4 bytes: YYYYMMDD)
	// Parse date string YYYY-MM-DD to integer YYYYMMDD
	dateParts := strings.Split(bet.Nacimiento, "-")
	if len(dateParts) != 3 {
		err := fmt.Errorf("invalid date format: %s", bet.Nacimiento)
		protocolLog.Errorf("action: send_single_bet | result: fail | field: nacimiento | error: %v", err)
		return err
	}
	year, err := strconv.Atoi(dateParts[0])
	if err != nil {
		protocolLog.Errorf("action: send_single_bet | result: fail | field: nacimiento_year | error: %v", err)
		return fmt.Errorf("invalid year: %v", err)
	}
	month, err := strconv.Atoi(dateParts[1])
	if err != nil {
		protocolLog.Errorf("action: send_single_bet | result: fail | field: nacimiento_month | error: %v", err)
		return fmt.Errorf("invalid month: %v", err)
	}
	day, err := strconv.Atoi(dateParts[2])
	if err != nil {
		protocolLog.Errorf("action: send_single_bet | result: fail | field: nacimiento_day | error: %v", err)
		return fmt.Errorf("invalid day: %v", err)
	}
	dateInt := uint32(year*10000 + month*100 + day)
	err = binary.Write(conn, binary.BigEndian, dateInt)
	if err != nil {
		protocolLog.Errorf("action: send_single_bet | result: fail | field: nacimiento | error: %v", err)
		return err
	}

	// Numero (4 bytes)
	numero, err := strconv.ParseUint(bet.Numero, 10, 32)
	if err != nil {
		protocolLog.Errorf("action: send_single_bet | result: fail | field: numero | error: %v", err)
		return fmt.Errorf("invalid numero: %v", err)
	}
	err = binary.Write(conn, binary.BigEndian, uint32(numero))
	if err != nil {
		protocolLog.Errorf("action: send_single_bet | result: fail | field: numero | error: %v", err)
		return err
	}

	return nil
}

func ReceiveResponse(conn net.Conn) (bool, error) {
	response := make([]byte, 1)
	n, err := conn.Read(response)
	if err != nil {
		protocolLog.Errorf("action: receive_response | result: fail | error: %v", err)
		return false, err
	}
	
	if n != 1 {
		protocolLog.Errorf("action: receive_response | result: fail | expected_bytes: 1 | received_bytes: %d", n)
		return false, fmt.Errorf("invalid response length: %d", n)
	}
	
	success := response[0] == ResponseOK
	protocolLog.Debugf("action: receive_response | result: success | response_code: %d | success: %v", response[0], success)
	return success, nil
}

func SendFinishedNotification(conn net.Conn, clientID string) error {
	// Message type (4 bytes)
	err := binary.Write(conn, binary.BigEndian, uint32(MessageTypeFinishedSending))
	if err != nil {
		protocolLog.Errorf("action: send_finished_notification | result: fail | field: message_type | error: %v", err)
		return err
	}

	// Message length (4 bytes)
	messageLength := uint32(4)
	err = binary.Write(conn, binary.BigEndian, messageLength)
	if err != nil {
		protocolLog.Errorf("action: send_finished_notification | result: fail | field: message_length | error: %v", err)
		return err
	}

	// Client ID (4 bytes)
	clientIDInt, err := strconv.ParseUint(clientID, 10, 32)
	if err != nil {
		protocolLog.Errorf("action: send_finished_notification | result: fail | field: client_id | error: %v", err)
		return fmt.Errorf("invalid client ID: %v", err)
	}
	err = binary.Write(conn, binary.BigEndian, uint32(clientIDInt))
	if err != nil {
		protocolLog.Errorf("action: send_finished_notification | result: fail | field: client_id | error: %v", err)
		return err
	}

	protocolLog.Debugf("action: send_finished_notification | result: success | client_id: %s", clientID)
	return nil
}

func SendQueryWinners(conn net.Conn, clientID string) error {
	// Message type (4 bytes)
	err := binary.Write(conn, binary.BigEndian, uint32(MessageTypeQueryWinners))
	if err != nil {
		protocolLog.Errorf("action: send_query_winners | result: fail | field: message_type | error: %v", err)
		return err
	}

	// Message length (4 bytes)
	messageLength := uint32(4)
	err = binary.Write(conn, binary.BigEndian, messageLength)
	if err != nil {
		protocolLog.Errorf("action: send_query_winners | result: fail | field: message_length | error: %v", err)
		return err
	}

	// Client ID (4 bytes)
	clientIDInt, err := strconv.ParseUint(clientID, 10, 32)
	if err != nil {
		protocolLog.Errorf("action: send_query_winners | result: fail | field: client_id | error: %v", err)
		return fmt.Errorf("invalid client ID: %v", err)
	}
	err = binary.Write(conn, binary.BigEndian, uint32(clientIDInt))
	if err != nil {
		protocolLog.Errorf("action: send_query_winners | result: fail | field: client_id | error: %v", err)
		return err
	}

	protocolLog.Debugf("action: send_query_winners | result: success | client_id: %s", clientID)
	return nil
}

func ReceiveWinners(conn net.Conn) ([]string, error) {
	// Response status (1 byte)
	response := make([]byte, 1)
	n, err := conn.Read(response)
	if err != nil {
		return nil, err
	}
	
	if response[0] != ResponseOK {
		return nil, fmt.Errorf("server returned error response: %d", response[0])
	}

	// Number of winners (4 bytes)
	winnersCountBytes := make([]byte, 4)
	n, err = conn.Read(winnersCountBytes)
	if err != nil {
		protocolLog.Errorf("action: receive_winners | result: fail | field: winners_count | error: %v", err)
		return nil, err
	}
	
	if n != 4 {
		protocolLog.Errorf("action: receive_winners | result: fail | expected_bytes: 4 | received_bytes: %d", n)
		return nil, fmt.Errorf("invalid winners count length: %d", n)
	}
	
	winnersCount := binary.BigEndian.Uint32(winnersCountBytes)
	
	winners := make([]string, winnersCount)
	for i := uint32(0); i < winnersCount; i++ {
		documentoBytes := make([]byte, 4)
		n, err = conn.Read(documentoBytes)
		if err != nil {
			protocolLog.Errorf("action: receive_winners | result: fail | field: documento | winner_index: %d | error: %v", i, err)
			return nil, err
		}
		
		if n != 4 {
			protocolLog.Errorf("action: receive_winners | result: fail | expected_bytes: 4 | received_bytes: %d | winner_index: %d", n, i)
			return nil, fmt.Errorf("invalid documento size for winner %d: %d", i, n)
		}
		
		documento := binary.BigEndian.Uint32(documentoBytes)
		winners[i] = fmt.Sprintf("%08d", documento)
	}

	protocolLog.Debugf("action: receive_winners | result: success | winners_count: %d", winnersCount)
	return winners, nil
}
