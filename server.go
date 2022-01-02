/*----------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See LICENSE in the project root for license information.
 *---------------------------------------------------------------------------------------*/

package main

import (
	"fmt"
	"net/http"
)

type home struct{}

func (h *home) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("home page ServeHTTP")
	w.Write([]byte("This is my home page"))
}

func main() {
	portNumber := "9000"
	var plugin Plugin

	mux := http.NewServeMux()

	mux.Handle("/", plugin.Middleware(&home{}))

	fmt.Println("Server listening on port ", portNumber)
	http.ListenAndServe("localhost:"+portNumber, mux)
}
