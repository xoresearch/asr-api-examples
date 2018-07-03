package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ExampleVoice			= "short_voice.flac"
	UploadingEndpoint		= "https://asr.sapiensapi.com/v1/speech:longrunningrecognize"
	FetchingResultsEndpoint	= "https://asr.sapiensapi.com/v1/operations/"
)

//-------------------------------------------------------------------------------------------------
func main() {
	// Set the number of threads to execute goroutines
	numCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)

	// Load the voice data
	voiceData, err := ioutil.ReadFile(path.Join(path.Dir(os.Args[0]), ExampleVoice))

	if err != nil {
		fmt.Printf("Failed to load the voice data: %s.\n", err.Error())
		return
	}

	// Compose uploading request
	uploadingRequest := LongRunningRecognizeRequest{
		Signal:				voiceData,
		LanguageCode:		"en-US",
		ExecuteBeamSearch:	false,
	}

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

		// Run fetching loop if something is uploaded
		operationsCount := countOperations(&operationsMap)

		if operationsCount == 0 {
			// No one voice has been successfully uploaded
			fmt.Print("No one voice has been successfully uploaded.\n")

		} else {
			// Report the number of successfully uploaded voices
			fmt.Printf("Uploaded %d voices. Fetching results...\n", operationsCount)

			// Wait for a while expecting one second per one operation
			time.Sleep(time.Second * time.Duration(operationsCount))

			// Fetch the operations results
			fetchingResultsLoop(&operationsMap, operationsCount)
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

			// Execute uploading call
			response, err := http.Post(UploadingEndpoint, "application/json;charset=utf-8", bytes.NewReader(uploadingRequestBody))

			if err != nil {
				fmt.Printf("Failed to upload the voice data: %s\n", err.Error())
				return
			}

			if response.StatusCode != http.StatusOK {
				fmt.Printf("Failed to upload the voice data: %s\n", deserializeErrorResponse(response.Body))
				return
			}

			// Deserialize uploading response
			operationId, err := deserializeUploadingResponse(response.Body)

			if err != nil {
				fmt.Printf("Failed to deserialize uploading response: %s\n", err.Error())
				return
			}

			// Save the operation id to check its state later on
			operationsMap.Store(operationId, true)
		}()
	}

	// Wait until all the goroutines are done
	waitGroup.Wait()
}

//-------------------------------------------------------------------------------------------------
func countOperations(operationsMap *sync.Map) uint32 {
	var operationsCount uint32

	operationsMap.Range(func(operationId, operationValue interface{}) bool {
		operationsCount++
		return true
	})

	return operationsCount
}

//-------------------------------------------------------------------------------------------------
func fetchingResultsLoop(operationsMap *sync.Map, operationsCount uint32) {
	for {
		// Iterate over all the operations and fetch results for each of them
		operationsMap.Range(func(operationId, operationValue interface{}) bool {
			// Execute fetching call
			response, err := http.Get(FetchingResultsEndpoint + strconv.FormatUint(operationId.(uint64), 10))

			if err != nil {
				fmt.Printf("Failed to fetch an operation state: %s\n", err.Error())
				return true
			}

			if response.StatusCode != http.StatusOK {
				fmt.Printf("Failed to fetch an operation state: %s\n", deserializeErrorResponse(response.Body))
				return true
			}

			// Deserialize fetching response
			transcriptions, completed, err := deserializeFetchingResponse(response.Body)

			if err != nil {
				fmt.Printf("Failed to deserialize fetching response: %s\n", err.Error())
				operationsMap.Delete(operationId)
				operationsCount--
				return true
			}

			if completed {
				fmt.Printf("Voice ID: %d\n", operationId.(uint64))

				for _, transcription := range transcriptions {
					var text string

					if len(transcription.Alternatives) != 0 {
						text = transcription.Alternatives[0].Transcript
					}

					fmt.Printf("\t%f-%f\t%s\n", transcription.TimeStart, transcription.TimeEnd, text)
				}

				operationsMap.Delete(operationId)
				operationsCount--
				return true
			}

			return true
		})

		// Stop fetching if all the operations have been completed
		if operationsCount == 0 {
			break
		}

		// Wait for a while before next fetching stage
		time.Sleep(time.Second)
	}
}

//-------------------------------------------------------------------------------------------------
func readUploadingParams() (int32, int32) {
	var uploadingIterations, uploadingConcurrency int32

	for {
		fmt.Print("Uploading iterations: \n")
		fmt.Scan(&uploadingIterations)

		if uploadingIterations > 0 && uploadingIterations <= 1000 {
			break
		}

		fmt.Print("Uploading iterations must be in range between 1 and 1000.\n")
	}

	for {
		fmt.Print("Uploading concurrency: \n")
		fmt.Scan(&uploadingConcurrency)

		if uploadingConcurrency > 0 && uploadingConcurrency <= 64 {
			break
		}

		fmt.Print("Uploading concurrency must be in range between 1 and 64.\n")
	}

	return uploadingIterations, uploadingConcurrency
}

//-------------------------------------------------------------------------------------------------
func readStoppingCondition() bool {
	var stoppingCondition string

	for {
		fmt.Print("Another test? (y/n): \n")
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
func deserializeUploadingResponse(responseBody io.ReadCloser) (uint64, error) {
	var response LongRunningRecognizeResponse

	err := json.NewDecoder(responseBody).Decode(&response)

	if err != nil {
		return 0, err
	}

	return response.OperationId, nil
}

//-------------------------------------------------------------------------------------------------
func deserializeFetchingResponse(responseBody io.ReadCloser) ([]*Transcription, bool, error) {
	var response FetchOperationResponse

	err := json.NewDecoder(responseBody).Decode(&response)

	if err != nil {
		return nil, false, err
	}

	if response.ProcessingStatus != "PROCESSING_COMPLETED" {
		return nil, false, nil
	}

	return response.Transcriptions, true, nil
}

//-------------------------------------------------------------------------------------------------
func deserializeErrorResponse(responseBody io.ReadCloser) string {
	readBuffer, _ := ioutil.ReadAll(responseBody)

	return string(readBuffer)
}
