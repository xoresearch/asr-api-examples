package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TestAudioFile			= "./SampleVoices-KimBorge.flac"
	UploadingEndpoint		= "http://192.168.80.102:50120/longrunningrecognize"
	PollingEndpoint			= "http://192.168.80.102:50120/operations/"
)

//-------------------------------------------------------------------------------------------------
func main() {
	// Set the number of threads to execute goroutines
	numCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)

	// Load the audio data
	audioData, err := ioutil.ReadFile(TestAudioFile)

	if err != nil {
		fmt.Printf("Failed to load the audio data: %s.\n", err.Error())
		return
	}

	// Compose uploading request
	var uploadingRequest struct {
		Signal					[]byte					`json:"signal"`
		LanguageCode			string					`json:"language_code"`
		ExecuteBeamSearch		bool					`json:"execute_beam_search"`
	}

	uploadingRequest.Signal = audioData
	uploadingRequest.LanguageCode = "en-US"
	uploadingRequest.ExecuteBeamSearch = false

	// Serialize uploading request
	uploadingRequestBody, err := json.Marshal(&uploadingRequest)

	if err != nil {
		fmt.Printf("Failed to serialize uploading request: %s.\n", err.Error())
		return
	}

	// Main loop
	for {
		// Ask user for uploading params
		uploadingIterations, uploadingConcurrency := readUploadingParams()

		// Execute uploading
		operationsMap := sync.Map{}

		uploadingLoop(uploadingRequestBody, uploadingIterations, uploadingConcurrency, &operationsMap)

		// Check the uploading results
		operationsIds := extractOperationsIds(&operationsMap)

		if len(operationsIds) == 0 {
			// No one uploading iteration has been successfully completed
			println("No one uploading iteration has been successfully completed.")

		} else {
			// Wait for a while expecting one second per one operation
			time.Sleep(time.Second * time.Duration(len(operationsIds)))

			// Execute polling for the operations results
			pollingLoop(operationsIds)
		}

		// Ask user if this is enough
		if readStoppingCondition() {
			break
		}
	}
}

//-------------------------------------------------------------------------------------------------
func uploadingLoop(uploadingRequestBody []byte, uploadingIterations int32, uploadingConcurrency int32, operationsMap *sync.Map) {
	concurrencyLevel := int32(0)
	waitGroup := sync.WaitGroup{}

	for iterationIndex := 0; iterationIndex != int(uploadingIterations); iterationIndex++ {
		// Make sure the running goroutines number does not overpass the specified threshold
		for atomic.LoadInt32(&concurrencyLevel) >= uploadingConcurrency {
			time.Sleep(time.Millisecond * 5)
		}

		// Increment the goroutines counter
		waitGroup.Add(1)
		atomic.AddInt32(&concurrencyLevel, 1)

		// Spawn a new goroutine
		go func () {
			defer waitGroup.Done()
			defer atomic.AddInt32(&concurrencyLevel, -1)

			// Execute the uploading call
			response, err := http.Post(UploadingEndpoint, "application/json;charset=utf-8", bytes.NewReader(uploadingRequestBody))

			if err != nil {
				fmt.Printf(fmt.Sprintf("Failed to upload the audio data: %s\n", err.Error()))
				return
			}

			if response.StatusCode != http.StatusOK {
				fmt.Printf(fmt.Sprintf("Failed to upload the audio data: %s\n", deserializeFaultyUploadingResponse(response.Body)))
				return
			}

			// Deserialize uploading response
			operationId, err := deserializeSuccessfulUploadingResponse(response.Body)

			if err != nil {
				fmt.Printf(fmt.Sprintf("Failed to deserialize uploading response: %s\n", err.Error()))
				return
			}

			// Save the operation id to check its state later on
			operationsMap.Store(iterationIndex, operationId)
		}()
	}

	// Wait until all the goroutines are done
	waitGroup.Wait()
}

//-------------------------------------------------------------------------------------------------
func extractOperationsIds(operationsMap *sync.Map) []uint32 {
	operationsIds := make([]uint32, 0, 1)

	operationsMap.Range(func(iterationIndex, operationId interface{}) bool {
		operationsIds = append(operationsIds, operationId.(uint32))
		return true
	})

	return operationsIds
}

//-------------------------------------------------------------------------------------------------
func pollingLoop(operationsIds []uint32) {
}

//-------------------------------------------------------------------------------------------------
func readUploadingParams() (int32, int32) {
	var uploadingIterations, uploadingConcurrency int32

	for {
		println("Uploading iterations: ")
		fmt.Scan(&uploadingIterations)

		if uploadingIterations > 0 && uploadingIterations <= 1000 {
			break
		}

		println("Uploading iterations must be in range between 1 and 1000.")
	}

	for {
		println("Uploading concurrency: ")
		fmt.Scan(&uploadingConcurrency)

		if uploadingConcurrency > 0 && uploadingConcurrency <= 64 {
			break
		}

		println("Uploading concurrency must be in range between 1 and 64.")
	}

	return uploadingIterations, uploadingConcurrency
}

//-------------------------------------------------------------------------------------------------
func readStoppingCondition() bool {
	var stoppingCondition string

	for {
		println("Another test? (y/n): ")
		fmt.Scan(&stoppingCondition)

		stoppingCondition = strings.ToLower(stoppingCondition)

		if stoppingCondition == "y" {
			return false
		}

		if stoppingCondition == "n" {
			return true
		}
	}
}

//-------------------------------------------------------------------------------------------------
func deserializeSuccessfulUploadingResponse(responseBody io.ReadCloser) (uint32, error) {
	var uploadingResponse struct {
		OperationId				uint32					`json:"operation_id"`
	}

	err := json.NewDecoder(responseBody).Decode(&uploadingResponse)

	if err != nil {
		return 0, err
	}

	return uploadingResponse.OperationId, nil
}

//-------------------------------------------------------------------------------------------------
func deserializeFaultyUploadingResponse(responseBody io.ReadCloser) string {
	readBuffer, _ := ioutil.ReadAll(responseBody)

	return string(readBuffer)
}
