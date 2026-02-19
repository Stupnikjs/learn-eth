package main

import "sync"

type Crawler struct {
	VisitedIPs map[string]bool
	Mu         sync.Mutex // Pour éviter les problèmes si tu lances plusieurs goroutines
}

func NewCrawler() *Crawler {
	return &Crawler{
		VisitedIPs: make(map[string]bool),
	}
}
