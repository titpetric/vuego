package tests

import (
	"log"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	p, _ := os.Getwd()
	if strings.HasSuffix(p, "/tests") {
		os.Chdir("..")
	}
	log.Println("Tests working dir:", p)
	os.Exit(m.Run())
}
