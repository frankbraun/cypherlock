// package main implements a github.com/JonathanLogan/cypherlock server.
package main

import (
	"github.com/JonathanLogan/cypherlock/clrpcserver"
	"github.com/JonathanLogan/cypherlock/ratchetserver"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
)

// Methods:
// - Create
//		- PersistencePath
//		- KeyPeriod
//		- PregenPeriod
// - Serve
//		- PersistencePath
//		- ListenAddr

var (
	flagCreate       bool
	flagServe        bool
	flagPath         string
	flagAddr         string
	flagKeyPeriod    int
	flagPregenPeriod int
)

func init() {
	flag.BoolVar(&flagCreate, "create", false, "create new github.com/JonathanLogan/cypherlock")
	flag.BoolVar(&flagServe, "serve", false, "serve github.com/JonathanLogan/cypherlock")
	flag.StringVar(&flagAddr, "addr", "127.0.0.1:11139", "path in which to store persistence data.")
	flag.StringVar(&flagPath, "path", "/tmp/github.com/JonathanLogan/cypherlockd", "path in which to store persistence data.")
	flag.IntVar(&flagKeyPeriod, "keyperiod", 3600, "time in seconds until ratchet private key proceeds.")
	flag.IntVar(&flagPregenPeriod, "genperiod", 24*3600, "time for which to pre-generate ratchet public keys.")
	flag.Parse()
}

func main() {
	fmt.Println("github.com/JonathanLogan/cypherlockd: Cypherlock minimal server. BSD version.")
	if !flagCreate && !flagServe {
		fmt.Println("ERR: -create or -serve required.")
		os.Exit(1)
	}
	if flagCreate && flagServe {
		fmt.Println("ERR: Either -create OR -serve.")
		os.Exit(1)
	}
	persistence := &ratchetserver.DummyFileStore{
		Path: flagPath,
	}

	if flagCreate {
		rs, err := ratchetserver.NewRatchetServer(persistence, rand.Reader, int64(flagKeyPeriod), int64(flagPregenPeriod))
		if err != nil {
			fmt.Printf("ERR: %s", err)
			os.Exit(1)
		}
		rs.GenerateKeys()
		err = rs.Persist()
		if err != nil {
			fmt.Printf("ERR: %s", err)
			os.Exit(1)
		}
		fmt.Println("Server created.")
		sigkeyB := rs.SignatureKey()
		fmt.Printf("SignatureKey: %s\n", hex.EncodeToString(sigkeyB[:]))
		os.Exit(0)
	}
	if flagServe {
		rs, err := ratchetserver.LoadRatchetServer(persistence, rand.Reader)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			os.Exit(1)
		}
		rs.StartService()
		c := make(chan struct{}, 1)
		fmt.Println("Serving...")
		sigkeyB := rs.SignatureKey()
		fmt.Printf("SignatureKey: %s\n", hex.EncodeToString(sigkeyB[:]))
		_, err = clrpcserver.NewRPCServer(rs, flagAddr)
		if err != nil {
			fmt.Printf("ERR: %s", err)
			os.Exit(1)
		}
		<-c
		// UnReachable.
		os.Exit(0)
	}
	os.Exit(0) // Unreachable.
}
