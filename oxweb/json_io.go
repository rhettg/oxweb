package oxweb

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
)

type JSONConn struct {
	bufConn *bufio.ReadWriter
}

func NewJSONConn(conn io.ReadWriter) *JSONConn {
	reader := bufio.NewReaderSize(conn, 1024*8)
	writer := bufio.NewWriter(conn)

	bufConn := bufio.NewReadWriter(reader, writer)

	return &JSONConn{bufConn}
}

func (jsonConn *JSONConn) ReadJSON() (data JSONData, err error) {
	// Get our query from the client
	input, _, err := jsonConn.bufConn.ReadLine()
	if err != nil {
		log.Fatal("Failed to read from client", err)
		return nil, err
	}

	log.Println("Found: ", string(input))

	// Parse the query
	var parsedInput JSONData
	err = json.Unmarshal(input, &parsedInput)
	if err != nil {
		log.Printf("Failure to decode: %s %s", input, err)
		return nil, err
	}

	return parsedInput, nil
}

func (jsonConn *JSONConn) WriteJSON(data JSONData) (err error) {
	outputBytes, err := json.Marshal(data)
	if err != nil {
		log.Println("Failed to marshall", err)
		return err
	}

	_, err = jsonConn.bufConn.WriteString(string(outputBytes) + "\n")
	if err != nil {
		return
	}

	// _, err := jsonConn.bufConn.WriteString("\n");
	// if err != nil {
	// 	return
	// }

	err = jsonConn.bufConn.Flush()
	return
}
