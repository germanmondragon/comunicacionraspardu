// server
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"

	//	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

//"syscall"
)

const (
	CONFIG  = "/Users/gm/Documents/proyectofondef/sh-berrys/configuracion.config"
	USER = "mac"
	PASS = "raspiscan"
	MSG_POW = "pow"
	FORMAT  = "%v-%v-%v"
	COMMAND = "raspistill -o /tmp/img/"
	EXT     = ".jpg"
	ARCH_TAR= "/tmp/img.tar.gz"
	CLEAN   = "rm -f /tmp/img/*.jpg & rm -f "+ARCH_TAR
	TAR     = "tar -czf"+ ARCH_TAR+" /tmp/img/"
	MKDIR   = "mkdir -p /tmp/img"
	ARCH_LOCAL ="/Users/gm/ftp/"
	
)

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func enviaMsg(ip string, msg string) {

	conn, err := net.Dial("tcp", ip+":1201")
	checkError(err)
	defer conn.Close()
	conn.Write([]byte(msg))
	reply := make([]byte, 1024)
	conn.Read(reply)
	fmt.Println(string(reply))
}

func remotoSSH(wg *sync.WaitGroup, comando chan string) {
	defer wg.Done()
	runner := <-comando
	client := getCliente(runner)	
	session, err := client.NewSession()
	checkError(err)
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	fmt.Println("el hilo", parametro)
	if err := session.Run(parametro); err != nil {
		fmt.Println("el hilo", parametro, runner)

		panic("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())

}
func getCliente(runner string)*ssh.Client{
	config := &ssh.ClientConfig{
		User: USER,
		Auth: []ssh.AuthMethod{
			ssh.Password(PASS),
		},
	}
	client, err := ssh.Dial("tcp", runner+":22", config)
	checkError(err)
	return client
	
}

func sftpRemoto(wg *sync.WaitGroup, comando chan string) {
	defer wg.Done()
	runner := <-comando
	fmt.Println("SCP desde", runner)
    client := getCliente(runner)
	
	clientFTP, err := sftp.NewClient(client)
	checkError(err)
	remoteFile, err := clientFTP.Open(ARCH_TAR)
	checkError(err)
	defer remoteFile.Close()
	localFile, err := os.Create(ARCH_LOCAL + runner + ".tar.gz")
	checkError(err)
	defer localFile.Close()
	t1 := time.Now()
	n, err := io.Copy(localFile, remoteFile)
	checkError(err)

	fmt.Println("read  bytes in %s", n, time.Since(t1))

	//_, err = io.Copy( remoteFile, localFile)
	//return err

}

func ejecutaSCP(lines []string) {
	message := make(chan string) // no buffer
	var wg sync.WaitGroup

	for i := 0; i < len(lines); i++ {
		wg.Add(1)
		go sftpRemoto(&wg, message)
		message <- string(lines[i])
	}
	wg.Wait()

}

func ejecutaHilos(lines []string, accion string) {
	parametro = accion
	message := make(chan string) // no buffer
	var wg sync.WaitGroup

	for i := 0; i < len(lines); i++ {
		wg.Add(1)
		go remotoSSH(&wg, message)
		message <- string(lines[i])
	}
	wg.Wait()

}
func geCommandTime() string {
	t1 := time.Now()
	timenow := fmt.Sprintf(FORMAT, t1.Day(), t1.Minute(), t1.Second())
	return COMMAND + timenow + EXT

}

var parametro string

func main() {

	contents, err := ioutil.ReadFile(CONFIG)
	checkError(err)
	lines := strings.Split(string(contents), "\n")
	fmt.Println(lines)

	ejecutaHilos(lines, MKDIR)
	ejecutaHilos(lines, CLEAN)

	ejecutaHilos(lines, geCommandTime())

	ejecutaHilos(lines, geCommandTime())
	ejecutaHilos(lines, TAR)
	ejecutaHilos(lines, SCP)

	ejecutaSCP(lines)

}
