package handler

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	ip := "192.168.1.1"

	for i := 0; i < 5; i++ {
		if !rl.Allow(ip) {
			t.Fatalf("tentative %d : devrait être autorisée (limite = 5)", i+1)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	ip := "192.168.1.2"

	for i := 0; i < 3; i++ {
		rl.Allow(ip)
	}

	if rl.Allow(ip) {
		t.Fatal("la 4e tentative devrait être bloquée (limite = 3)")
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)
	ip := "192.168.1.3"

	rl.Allow(ip)
	rl.Allow(ip)

	if rl.Allow(ip) {
		t.Fatal("la 3e tentative dans la même fenêtre devrait être bloquée")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow(ip) {
		t.Fatal("après expiration de la fenêtre, la requête devrait être autorisée")
	}
}

func TestRateLimiter_IsolatesIPs(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	ip1 := "10.0.0.1"
	ip2 := "10.0.0.2"

	rl.Allow(ip1)
	rl.Allow(ip1)
	rl.Allow(ip1) // bloquée pour ip1

	if rl.Allow(ip2) == false {
		t.Fatal("ip2 devrait avoir son propre compteur indépendant de ip1")
	}
}

func TestRateLimiter_FirstRequestAlwaysAllowed(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	if !rl.Allow("1.2.3.4") {
		t.Fatal("la première requête doit toujours être autorisée")
	}
}
