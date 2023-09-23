package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:25565") // Default Minecraft port
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Server started on port 25565")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read the packet length (VarInt)
	packetLength, _ := readVarInt(conn)

	// Read the packet ID (VarInt)
	packetID, _ := readVarInt(conn)

	// For now, we'll just handle the handshake packet (ID 0x00)
	if packetID == 0x00 {
		handleHandshake(conn, packetLength)
	}
}

func handleHandshake(conn net.Conn, packetLength int) {
	// Read protocol version (VarInt)
	_, _ = readVarInt(conn)

	// Read server address (String)
	serverAddress, _ := readString(conn)

	// Read server port (Unsigned Short)
	port := make([]byte, 2)
	_, _ = conn.Read(port)
	serverPort := binary.BigEndian.Uint16(port)

	// Read next state (VarInt)
	_, _ = readVarInt(conn)

	fmt.Printf("Received handshake from %s for server %s:%d\n", conn.RemoteAddr(), serverAddress, serverPort)
}

func readVarInt(conn net.Conn) (int, int) {
	var result int32
	var bytesRead int
	var read byte
	var err error

	for shift := uint(0); ; shift += 7 {
		if shift >= 32 {
			// VarInt is too large
			return 0, bytesRead
		}

		read, err = readByte(conn)
		if err != nil {
			return 0, bytesRead
		}
		bytesRead++

		result |= int32(read&0x7F) << shift

		if (read & 0x80) == 0 {
			break
		}
	}

	return int(result), bytesRead
}

func readString(conn net.Conn) (string, int) {
	length, bytesRead := readVarInt(conn)
	bytes := make([]byte, length)
	_, _ = conn.Read(bytes)
	return string(bytes), bytesRead + length
}

func readByte(conn net.Conn) (byte, error) {
	b := make([]byte, 1)
	_, err := conn.Read(b)
	return b[0], err
}
