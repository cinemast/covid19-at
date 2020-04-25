package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiOverall(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleApiTotal))
	defer ts.Close()
	_, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
}

func TestApiBezirk(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleApiBezirk))
	defer ts.Close()
	_, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
}

func TestApiBundesland(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleApiBundesland))
	defer ts.Close()
	_, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
}
