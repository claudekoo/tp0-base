package common

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

var protocolLog = logging.MustGetLogger("protocol")

func recvAll(conn net.Conn, n int) ([]byte, error) {
	data := make([]byte, n)
	totalRead := 0
	for totalRead < n {
		bytesRead, err := conn.Read(data[totalRead:])
		if err != nil {
			return nil, err
		}
		if bytesRead == 0 {
			return nil, io.EOF
		}
		totalRead += bytesRead
	}
	return data, nil
}

func sendAll(conn net.Conn, data []byte) error {
	totalSent := 0
	for totalSent < len(data) {
		bytesSent, err := conn.Write(data[totalSent:])
		if err != nil {
			return err
		}
		if bytesSent == 0 {
			return fmt.Errorf("socket connection broken")
		}
		totalSent += bytesSent
	}
	return nil
}

func writeUint32BE(conn net.Conn, value uint32) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, value)
	return sendAll(conn, data)
}

func readUint32BE(conn net.Conn) (uint32, error) {
	data, err := recvAll(conn, 4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(data), nil
}

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
	// Client ID (4 bytes)
	clientID, err := strconv.ParseUint(batch.ClientID, 10, 32)
	if err != nil {
		protocolLog.Errorf("action: send_bet_batch | result: fail | field: client_id | error: %v", err)
		return fmt.Errorf("invalid client ID: %v", err)
	}

	messageSize := calculateBatchMessageSize(batch)
	batchSize := uint32(len(batch.Bets))
	
	var message []byte
	
	// Message type (4 bytes)
	messageTypeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(messageTypeBytes, uint32(MessageTypeBatch))
	message = append(message, messageTypeBytes...)
	
	// Message length (4 bytes)
	messageLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(messageLengthBytes, uint32(messageSize))
	message = append(message, messageLengthBytes...)
	
	// Client ID (4 bytes)
	clientIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(clientIDBytes, uint32(clientID))
	message = append(message, clientIDBytes...)
	
	// Batch size (4 bytes)
	batchSizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(batchSizeBytes, batchSize)
	message = append(message, batchSizeBytes...)
	
	for i, bet := range batch.Bets {
		betBytes, err := serializeBet(bet)
		if err != nil {
			protocolLog.Errorf("action: send_bet_batch | result: fail | bet_index: %d | error: %v", i, err)
			return fmt.Errorf("failed to serialize bet %d: %v", i, err)
		}
		message = append(message, betBytes...)
	}

	err = sendAll(conn, message)
	if err != nil {
		protocolLog.Errorf("action: send_bet_batch | result: fail | error: %v", err)
		return err
	}

	protocolLog.Debugf("action: send_bet_batch | result: success | client_id: %s | batch_size: %d | message_size: %d", batch.ClientID, batchSize, messageSize)
	return nil
}

func serializeBet(bet BetData) ([]byte, error) {
	var betBytes []byte
	
	// Nombre length (4 bytes) and Nombre
	nameBytes := []byte(bet.Nombre)
	nameLen := uint32(len(nameBytes))
	nameLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(nameLenBytes, nameLen)
	betBytes = append(betBytes, nameLenBytes...)
	betBytes = append(betBytes, nameBytes...)

	// Apellido length (4 bytes) and Apellido  
	surnameBytes := []byte(bet.Apellido)
	surnameLen := uint32(len(surnameBytes))
	surnameLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(surnameLenBytes, surnameLen)
	betBytes = append(betBytes, surnameLenBytes...)
	betBytes = append(betBytes, surnameBytes...)

	// Documento (4 bytes)
	documento, err := strconv.ParseUint(bet.Documento, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid documento: %v", err)
	}
	documentoBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(documentoBytes, uint32(documento))
	betBytes = append(betBytes, documentoBytes...)

	// Nacimiento (4 bytes: YYYYMMDD)
	// Parse date string YYYY-MM-DD to integer YYYYMMDD
	dateParts := strings.Split(bet.Nacimiento, "-")
	if len(dateParts) != 3 {
		return nil, fmt.Errorf("invalid date format: %s", bet.Nacimiento)
	}
	year, err := strconv.Atoi(dateParts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid year: %v", err)
	}
	month, err := strconv.Atoi(dateParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid month: %v", err)
	}
	day, err := strconv.Atoi(dateParts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid day: %v", err)
	}
	dateInt := uint32(year*10000 + month*100 + day)
	dateBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(dateBytes, dateInt)
	betBytes = append(betBytes, dateBytes...)

	// Numero (4 bytes)
	numero, err := strconv.ParseUint(bet.Numero, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid numero: %v", err)
	}
	numeroBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(numeroBytes, uint32(numero))
	betBytes = append(betBytes, numeroBytes...)

	return betBytes, nil
}

func ReceiveResponse(conn net.Conn) (bool, error) {
	messageLength, err := readUint32BE(conn)
	if err != nil {
		protocolLog.Errorf("action: receive_response | result: fail | field: message_length | error: %v", err)
		return false, err
	}
	
	response, err := recvAll(conn, int(messageLength))
	if err != nil {
		protocolLog.Errorf("action: receive_response | result: fail | error: %v", err)
		return false, err
	}
	
	success := response[0] == ResponseOK
	protocolLog.Debugf("action: receive_response | result: success | response_code: %d | success: %v", response[0], success)
	return success, nil
}

func SendFinishedNotification(conn net.Conn, clientID string) error {
	clientIDInt, err := strconv.ParseUint(clientID, 10, 32)
	if err != nil {
		protocolLog.Errorf("action: send_finished_notification | result: fail | field: client_id | error: %v", err)
		return fmt.Errorf("invalid client ID: %v", err)
	}

	// Build entire message in memory
	var message []byte
	
	// Message type (4 bytes)
	messageTypeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(messageTypeBytes, uint32(MessageTypeFinishedSending))
	message = append(message, messageTypeBytes...)
	
	// Message length (4 bytes)
	messageLength := uint32(4)
	messageLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(messageLengthBytes, messageLength)
	message = append(message, messageLengthBytes...)
	
	// Client ID (4 bytes)
	clientIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(clientIDBytes, uint32(clientIDInt))
	message = append(message, clientIDBytes...)

	err = sendAll(conn, message)
	if err != nil {
		protocolLog.Errorf("action: send_finished_notification | result: fail | error: %v", err)
		return err
	}

	protocolLog.Debugf("action: send_finished_notification | result: success | client_id: %s", clientID)
	return nil
}

func SendQueryWinners(conn net.Conn, clientID string) error {
	// Client ID (4 bytes)
	clientIDInt, err := strconv.ParseUint(clientID, 10, 32)
	if err != nil {
		protocolLog.Errorf("action: send_query_winners | result: fail | field: client_id | error: %v", err)
		return fmt.Errorf("invalid client ID: %v", err)
	}
	var message []byte
	
	// Message type (4 bytes)
	messageTypeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(messageTypeBytes, uint32(MessageTypeQueryWinners))
	message = append(message, messageTypeBytes...)
	
	// Message length (4 bytes)
	messageLength := uint32(4)
	messageLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(messageLengthBytes, messageLength)
	message = append(message, messageLengthBytes...)
	
	// Client ID (4 bytes)
	clientIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(clientIDBytes, uint32(clientIDInt))
	message = append(message, clientIDBytes...)

	err = sendAll(conn, message)
	if err != nil {
		protocolLog.Errorf("action: send_query_winners | result: fail | error: %v", err)
		return err
	}

	protocolLog.Debugf("action: send_query_winners | result: success | client_id: %s", clientID)
	return nil
}

func ReceiveWinners(conn net.Conn) ([]string, error) {
	// Message length (4 bytes)
	messageLength, err := readUint32BE(conn)
	if err != nil {
		protocolLog.Errorf("action: receive_winners | result: fail | field: message_length | error: %v", err)
		return nil, err
	}
	
	messageData, err := recvAll(conn, int(messageLength))
	if err != nil {
		protocolLog.Errorf("action: receive_winners | result: fail | field: message_data | error: %v", err)
		return nil, err
	}
	
	if len(messageData) < 1 {
		return nil, fmt.Errorf("message too short: %d bytes", len(messageData))
	}
	
	if messageData[0] != ResponseOK {
		return nil, fmt.Errorf("server returned error response: %d", messageData[0])
	}
	
	if len(messageData) < 5 { // 1 byte response + 4 bytes count
		return nil, fmt.Errorf("message too short for winners count: %d bytes", len(messageData))
	}

	winnersCount := binary.BigEndian.Uint32(messageData[1:5])
		
	winners := make([]string, winnersCount)
	offset := 5 // after response (1) + count (4)
	
	for i := uint32(0); i < winnersCount; i++ {
		documento := binary.BigEndian.Uint32(messageData[offset : offset+4])
		winners[i] = fmt.Sprintf("%08d", documento)
		offset += 4
	}

	protocolLog.Debugf("action: receive_winners | result: success | winners_count: %d", winnersCount)
	return winners, nil
}
