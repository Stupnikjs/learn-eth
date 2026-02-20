package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	TARGETIP  net.IP = net.ParseIP("18.138.108.67") // Bootnode Fondation
	visited          = make(map[string]bool)
	visitedMu sync.Mutex
	jobs      = make(chan net.IP, 1000) // File d'attente des IPs à scanner
)

func job() {
	// Une node de boot
	jobs <- TARGETIP

	// On lance un petit nombre de workers pour ne pas se faire bannir
	for i := range 4 {
		_ = i
		time.Sleep(time.Second)
		go worker()
	}

	// Bloque le main
	select {}
}

func worker() {
	for ip := range jobs {
		// 1. Vérifier si déjà visitée
		visitedMu.Lock()
		if visited[ip.String()] {
			visitedMu.Unlock()
			continue
		}
		visited[ip.String()] = true
		visitedMu.Unlock()

		fmt.Printf("🔍 Exploration de : %s\n", ip)

		// 2. Lancer le scan (ton ancienne fonction V4 modifiée)
		// Elle doit maintenant s'arrêter après un timeout
		foundIPs := V4(ip)

		// 3. Ajouter les nouveaux voisins à la file
		for _, newIP := range foundIPs {
			jobs <- newIP
		}
	}
}
