package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/kobiton/bitrise-step-kobiton-execute-test/model"
	"github.com/kobiton/bitrise-step-kobiton-execute-test/utils"
)

const MAX_MS_WAIT_FOR_EXECUTION = 1 * 3600 * 1000 // 1 hour in miliseconds

var jobId = ""
var reportUrl = ""

func main() {

	stepConfig := new(model.StepConfig)
	stepConfig.Init()

	var executorBasicAuth = strings.Join([]string{stepConfig.GetExecutorUsername(), stepConfig.GetExecutorPassword()}, ":")
	var executorBasicAuthEncoded = utils.Base64Encode(executorBasicAuth)

	var headers = map[string]string{}
	headers["x-kobiton-credential-username"] = stepConfig.GetKobiUsername()
	headers["x-kobiton-credential-api-key"] = stepConfig.GetKobiPassword()
	headers["authorization"] = "Basic " + executorBasicAuthEncoded
	headers["content-type"] = "application/json"
	headers["accept"] = "application/json"

	executorPayload := new(model.ExecutorRequestPayload)
	model.BuildExecutorRequestPayload(executorPayload, stepConfig)
	executorJsonPayload, _ := json.MarshalIndent(executorPayload, "", "   ")
	client := utils.HttpClient()

	var executorUrl = stepConfig.GetExecutorUrl() + "/submit"

	var response = utils.SendRequest(client, "POST", executorUrl, headers, executorJsonPayload)

	jobId = string(response)

	if stepConfig.IsWaitForExecution() {

		log.Printf("Requesting to get logs for job %s", jobId)

		var getJobInfoUrl = stepConfig.GetExecutorUrl() + "/jobs/" + jobId
		var getJobLogUrl = getJobInfoUrl + "/logs?type=" + stepConfig.GetLogType()
		var getReportUrl = getJobInfoUrl + "/report"
		var isTimeout = false

		ticker := time.NewTicker(30 * time.Second)
		var authHeader = map[string]string{"authorization": "Basic " + executorBasicAuthEncoded}
		var jobResponse model.JobResponse
		var waitingBeginAt = time.Now().UnixNano() / int64(time.Millisecond)

		for range ticker.C {
			var response = utils.SendRequest(client, "GET", getJobInfoUrl, authHeader, nil)
			json.Unmarshal(response, &jobResponse)
			log.Println("Job Status: ", jobResponse.Status)

			if jobResponse.Status == "COMPLETED" || jobResponse.Status == "FAILED" {
				log.Printf("Job ID %s is finish with status: %s", jobId, jobResponse.Status)
				break
			} else {
				var currentTime = time.Now().UnixNano() / int64(time.Millisecond)

				if currentTime-waitingBeginAt >= MAX_MS_WAIT_FOR_EXECUTION {
					isTimeout = true
					break
				}
			}
		}
		defer ticker.Stop()

		if isTimeout {
			log.Println("==============================================================================")
			log.Println("Execution has reached maximum waiting time")
		} else {
			var logResponse = utils.SendRequest(client, "GET", getJobLogUrl, authHeader, nil)

			log.Println("==============================================================================")
			log.Println(string(logResponse))

			var reportResponse = utils.SendRequest(client, "GET", getReportUrl, authHeader, nil)
			reportUrl = string(reportResponse)
		}
	}

	log.Println("==============================================================================")

	if jobId != "" {
		log.Println("Job ID: ", jobId)
	}

	if reportUrl != "" {
		log.Println("Report URL: ", reportUrl)
	}
	//
	// --- Step Outputs: Export Environment Variables for other Steps:
	// You can export Environment Variables for other Steps with
	//  envman, which is automatically installed by `bitrise setup`.
	// A very simple example:
	utils.ExposeEnv("JOB_ID", jobId)
	utils.ExposeEnv("REPORT_URL", reportUrl)
	// You can find more usage examples on envman's GitHub page
	//  at: https://github.com/bitrise-io/envman

	//
	// --- Exit codes:
	// The exit code of your Step is very important. If you return
	//  with a 0 exit code `bitrise` will register your Step as "successful".
	// Any non zero exit code will be registered as "failed" by `bitrise`.
	os.Exit(0)
}
