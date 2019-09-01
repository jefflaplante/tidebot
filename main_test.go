package main

import (
	"testing"
	"time"
)

func Test_ymd(t *testing.T) {
	// RFC3339
	// 2006-01-02T15:04:05Z07:00

	time, err := time.Parse(time.RFC3339, "2019-05-09T12:01:02Z")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(time)

	got := ymd(time)
	want := "20190509"
	if got != want {
		t.Fatalf("got %v and wanted %v", got, want)
	}
	t.Logf("got: %v", got)
}