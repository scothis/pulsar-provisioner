/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	gateway := os.Getenv("GATEWAY")
	if gateway == "" {
		log.Fatal("Environment variable GATEWAY should contain the host and port of a liiklus gRPC endpoint")
	}
	broker := os.Getenv("BROKER")
	if broker == "" {
		log.Fatal("Environment variable BROKER should contain the service URL of a Pulsar cluster")
	}
	tenant := os.Getenv("TENANT")
	if tenant == "" {
		log.Fatal("Environment variable TENANT should contain a tenant of the Pulsar cluster")
	}
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		log.Fatal("Environment variable NAMESPACE should contain a namespace within a tenant of the Pulsar cluster")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handleProvisionRequest(broker, gateway, tenant, namespace, w, r)
	})
	_ = http.ListenAndServe(":8080", nil)
}

func handleProvisionRequest(broker, gateway, tenant, namespace string, writer http.ResponseWriter, request *http.Request) {

	parts := strings.Split(request.URL.Path[1:], "/")
	if len(parts) != 2 {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(writer, "URLs should be of the form /<namespace>/<stream-name>\n")
		return
	}
	// TODO define a better scheme to define topic names
	topicName := fmt.Sprintf("persistent://%s/%s/%s_%s", tenant, namespace, parts[0], parts[1])

	writer.WriteHeader(http.StatusOK)

	if err := encodeResponse(writer, gateway, topicName); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to write json response: %v", err)
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, "Reported successful topic %q\n", topicName)

}

func encodeResponse(w http.ResponseWriter, gateway string, topicName string) error {
	w.Header().Set("Content-Type", "application/json")
	res := result{
		Gateway: gateway,
		Topic:   topicName,
	}
	return json.NewEncoder(w).Encode(res)
}

type result struct {
	Gateway string `json:"gateway"`
	Topic   string `json:"topic"`
}
