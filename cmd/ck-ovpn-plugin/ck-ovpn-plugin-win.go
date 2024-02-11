// +build windows

//
// Copyright (c) 2022, Mykola Baibuz <mykola.baibuz@gmail.com> for Amnezia-VPN project (https://amnezia.org/)
//  All rights reserved.
//


package main

import (
	"C"
	"encoding/base64"
	"net"
	"reflect"
	"unsafe"

	"github.com/cbeuw/Cloak/internal/common"
	"github.com/cbeuw/connutil"

	"github.com/cbeuw/Cloak/internal/client"
	"github.com/cbeuw/Cloak/internal/multiplex"
	log "github.com/sirupsen/logrus"
)
import (
	"encoding/binary"
	"encoding/json"
	"syscall"
)

var fd int

func socket_get_fd(network string, address string, c syscall.RawConn) error {
	fn := func(s uintptr) {
		fd = int(s)
	}

	if err := c.Control(fn); err != nil {
		return err
	}

	return nil
}

//export Cloak_native_handle
func Cloak_native_handle() int {
 	return fd
}

func generateClientConfigs(rawConfig client.RawConfig, state common.WorldState) (client.LocalConnConfig, client.RemoteConnConfig, client.AuthInfo) {
	lcl, rmt, auth, err := rawConfig.ProcessRawConfig(state)
	if err != nil {
		log.Fatal(err)
	}
	return lcl, rmt, auth
}

var singleplexTCPConfig client.RawConfig
var CloakConnections = map[int]net.Conn{}
var nextID = 0

//export Initialize_cloak_c_client
func Initialize_cloak_c_client(base64Config *C.char) (clientKey int) {

	gobase64Config := C.GoString(base64Config)
	CkJsonConfig, err := base64.StdEncoding.DecodeString(gobase64Config)

	if err != nil {
		log.Error(err)
		return -1
	}

	var rawConfig *client.RawConfig

	rawConfig = new(client.RawConfig)
	err = json.Unmarshal(CkJsonConfig, &rawConfig)

	if rawConfig.LocalHost == "" {
		rawConfig.LocalHost = "127.0.0.1"
	}
	if rawConfig.LocalPort == "" {
		rawConfig.LocalPort = "9999"
	}

	if err != nil {
		log.Error(err)
		return -1
	}

	singleplexTCPConfig = *rawConfig

	return 0
}

//export Cloak_listen
func Cloak_listen(address_string *C.char) {
}

//export Cloak_dial
func Cloak_dial() (clientKey int) {

	var seshMaker func() *multiplex.Session

	worldState := common.RealWorldState
	localConfig, remoteConfig, authInfo := generateClientConfigs(singleplexTCPConfig, worldState)

	var localDialer common.Dialer
	var localListener *connutil.PipeListener

	d := &net.Dialer{Control: socket_get_fd, KeepAlive: remoteConfig.KeepAlive}

	seshMaker = func() *multiplex.Session {
		authInfo := authInfo // copy the struct because we are overwriting SessionId

		randByte := make([]byte, 1)
		common.RandRead(authInfo.WorldState.Rand, randByte)
		authInfo.MockDomain = localConfig.MockDomainList[int(randByte[0])%len(localConfig.MockDomainList)]

		// sessionID is usergenerated. There shouldn't be a security concern because the scope of
		// sessionID is limited to its UID.
		quad := make([]byte, 4)
		common.RandRead(authInfo.WorldState.Rand, quad)
		authInfo.SessionId = binary.BigEndian.Uint32(quad)
		return client.MakeSession(remoteConfig, authInfo, d)
	}

	localDialer, localListener = connutil.DialerListener(10 * 1024)

	go client.RouteTCP(localListener, localConfig.Timeout, remoteConfig.Singleplex, seshMaker)

	clientConn, err := localDialer.Dial("", "")

	if err != nil {
		log.Error(err)
		return -1
	}

	CloakConnections[nextID] = clientConn

	// This is the return value
	clientKey = nextID

	nextID += 1
	return clientKey
}

//export Cloak_write
func Cloak_write(client_id int, buffer unsafe.Pointer, buffer_length C.int) int {
	var connection = CloakConnections[client_id]
	var bytesBuffer = C.GoBytes(buffer, buffer_length)
	numberOfBytesWritten, error := connection.Write(bytesBuffer)

	if error != nil {
		return -1
	} else {
		return numberOfBytesWritten
	}
}

//export Cloak_read
func Cloak_read(client_id int, buffer unsafe.Pointer, buffer_length int) int {

	var connection = CloakConnections[client_id]
	if connection == nil {
		return -1
	}
	header := reflect.SliceHeader{uintptr(buffer), buffer_length, buffer_length}
	bytesBuffer := *(*[]byte)(unsafe.Pointer(&header))

	numberOfBytesRead, error := connection.Read(bytesBuffer)

	if error != nil {
		return -1
	} else {
		return numberOfBytesRead
	}
}

//export Cloak_close_connection
func Cloak_close_connection(client_id int) {
	var connection = CloakConnections[client_id]
	connection.Close()
	syscall.Close(syscall.Handle(fd))
	delete(CloakConnections, client_id)
}

func main() {}
