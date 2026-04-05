package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	server := flag.String("server", "", "server to use: NL, PL, BG, LV (default: auto)")
	tunnelOnly := flag.Bool("n", false, "tunnel only, skip update")
	flag.Parse()

	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ltime | log.Lshortfile)

	// кэшируем sudo на всю сессию
	if err := exec.Command("sudo", "-v").Run(); err != nil {
		log.Fatalf("sudo required: %v", err)
	}

	// 1. Start session
	sessionID, vlessLinks, err := StartSession()
	if err != nil {
		log.Fatalf("failed to start session: %v", err)
	}
	log.Printf("session started: %s", sessionID)

	fatal := func(msg string, err error) {
		log.Printf("%s: %v", msg, err)
		if cerr := CompleteSession(sessionID, "", "", false); cerr != nil {
			log.Printf("complete session (failure): %v", cerr)
		}
		os.Exit(1)
	}

	// 2. Start sing-box
	proc, err := StartSingBox(vlessLinks, *server)
	if err != nil {
		fatal("failed to start sing-box", err)
	}
	defer StopSingBox(proc)
	defer os.Remove(SingBoxConfigPath)

	os.Setenv("http_proxy", "http://127.0.0.1:7890")
	os.Setenv("https_proxy", "http://127.0.0.1:7890")

	// 3. Wait for tunnel to come up
	time.Sleep(8 * time.Second)

	// 4. Check tunnel
	country, err := CheckTunnel()
	if err != nil {
		fatal("tunnel check failed", err)
	}
	log.Printf("tunnel active, country: %s", country)

	// 5. Heartbeat goroutine
	stopHeartbeat := make(chan struct{})
	heartbeatDone := make(chan struct{})
	go func() {
		defer close(heartbeatDone)
		ticker := time.NewTicker(HeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := SendHeartbeat(sessionID); err != nil {
					log.Printf("heartbeat error: %v", err)
				}
			case <-stopHeartbeat:
				return
			}
		}
	}()

	// 6. Run update (or hold tunnel if -n)
	var success bool
	if *tunnelOnly {
		log.Println("tunnel only mode, skipping update")
		// Hold tunnel until Ctrl+C / SIGTERM
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		close(stopHeartbeat)
		<-heartbeatDone
		success = true
	} else {
		log.Println("running steamos-update")
		updateErr := RunUpdate()
		close(stopHeartbeat)
		<-heartbeatDone

		success = updateErr == nil
		if updateErr != nil {
			log.Printf("update failed: %v", updateErr)
		} else {
			log.Println("update completed successfully")
		}
	}

	// 7. Complete session
	if err := CompleteSession(sessionID, country, country, success); err != nil {
		log.Printf("complete session: %v", err)
	}

	if !success {
		fmt.Fprintln(os.Stderr, "update failed, exiting with error")
		os.Exit(1)
	}
}
