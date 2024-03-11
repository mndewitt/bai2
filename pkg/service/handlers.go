// Copyright 2022 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/moov-io/bai2/pkg/lib"
)

func outputError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func outputSuccess(w http.ResponseWriter, output string) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"groups": output,
	})
}

func parseInputFromRequest(r *http.Request) (*lib.Bai2, error) {
	inputFile, _, err := r.FormFile("input")
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	var input bytes.Buffer
	if _, err = io.Copy(&input, inputFile); err != nil {
		return nil, err
	}

	// convert byte slice to io.Reader
	scan := lib.NewBai2Scanner(bytes.NewReader(input.Bytes()))
	f := lib.NewBai2()
	err = f.Read(&scan)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func outputBufferToWriter(w http.ResponseWriter, f *lib.Bai2) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(f.String()))
}

// parse - parse bai2 report
func parse(w http.ResponseWriter, r *http.Request) {
	f, err := parseInputFromRequest(r)
	if err != nil {
		outputError(w, http.StatusBadRequest, err)
		return
	}

	fmt.Println(f)

	err = f.Validate()
	if err != nil {
		outputError(w, http.StatusNotImplemented, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// All the data objects from the file are in the 'f' struct
	// See pkg/lib/file.go - Bai2 struct to see how this is composed

	// So, at the top level there are file level data points..
	// log.Println(f.Receiver)
	// log.Println(f.Sender)

	// Then, the Groups array, which in turn contains an Accounts array with details of account balance,
	// transactions, etc...
	json.NewEncoder(w).Encode(map[string]interface{}{
		"groups": f.Groups,
		"sender": f.Sender,
		"receiver": f.Receiver,
	})
}

// print - print bai2 report after parse
func print(w http.ResponseWriter, r *http.Request) {
	f, err := parseInputFromRequest(r)
	if err != nil {
		outputError(w, http.StatusBadRequest, err)
		return
	}

	outputBufferToWriter(w, f)
}

// health - health check
func health(w http.ResponseWriter, r *http.Request) {
	outputSuccess(w, "alive")
}

// configure handlers
func ConfigureHandlers(r *mux.Router) error {

	r.HandleFunc("/health", health).Methods("GET")
	r.HandleFunc("/print", print).Methods("POST")
	r.HandleFunc("/parse", parse).Methods("POST")

	return nil
}
