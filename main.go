package main

import (
	"encoding/json"
	"github.com/chuong777/bitrise-step-kobiton-execute-test/model"
	"github.com/chuong777/bitrise-step-kobiton-execute-test/utils"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const MAX_MS_WAIT_FOR_EXECUTION = 1 * 3600 * 1000 // 1 hour in miliseconds

var jobId = ""
var reportUrl = ""

func main() {
	stepConfig := new(model.StepConfig)
	stepConfig.Init()

	var headers = getRequestHeader(stepConfig)

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
		var jobResponse model.JobResponse
		var waitingBeginAt = time.Now().UnixNano() / int64(time.Millisecond)

		for ; true; <-ticker.C {
			var response = utils.SendRequest(client, "GET", getJobInfoUrl, headers, nil)
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
			var logResponse = utils.SendRequest(client, "GET", getJobLogUrl, headers, nil)

			log.Println("==============================================================================")
			log.Println(string(logResponse))

			var reportResponse = utils.SendRequest(client, "GET", getReportUrl, headers, nil)
			reportUrl = string(reportResponse)

			if stepConfig.GetScriptlessAutomation() {
				runScriptless(stepConfig)
			}
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

func runScriptless(stepConfig *model.StepConfig) {
	log.Println("INFO: Scriptless automation has begun...")

	var isTimeout = false
	var scriptlessResponse *model.ScriptlessStatusResponse
	var waitingBeginAt = time.Now().UnixNano() / int64(time.Millisecond)
	var statusUrl = stepConfig.GetExecutorUrl() + "/jobs/" + jobId + "/scriptless/status"
	scriptlessTicker := time.NewTicker(30 * time.Second)
	client := utils.HttpClient()
	var headers = getRequestHeader(stepConfig)

	for ; true; <-scriptlessTicker.C {
		var response = utils.SendRequest(client, "GET", statusUrl, headers, nil)
		json.Unmarshal(response, &scriptlessResponse)

		if len(scriptlessResponse.Messages) > 0 {
			for _, message := range scriptlessResponse.Messages {
				log.Println(message)
			}
		}

		log.Println("INFO: Scriptless automation is currently: ", scriptlessResponse.Status)

		if scriptlessResponse.Status == "COMPLETED" {
			break
		} else {
			var currentTime = time.Now().UnixNano() / int64(time.Millisecond)

			if currentTime-waitingBeginAt >= stepConfig.GetScriptlessTimeout()*1000 {
				isTimeout = true
				break
			}
		}
	}

	defer scriptlessTicker.Stop()

	utils.ExposeEnv("SCRIPTLESS_AUTOMATION_PASSED", strconv.FormatBool(
		!isTimeout && scriptlessResponse != nil && scriptlessResponse.Error == "" && scriptlessResponse.ExecutionsPassed))
	if isTimeout {
		log.Println("ERROR: Scriptless automation has reached a timeout.")
	} else {
		if scriptlessResponse == nil {
			log.Println("ERROR: Unable to retrieve scriptless automation status.")
			return
		}

		var errorMessage = scriptlessResponse.Error
		if errorMessage == "" {
			var resultMessage string
			if scriptlessResponse.ExecutionsPassed {
				resultMessage = "Passed"
			} else {
				resultMessage = "Failed"
			}

			log.Println("INFO: Scriptless testing has completed, result: " + resultMessage)
			fileUrl := stepConfig.GetExecutorUrl() + "/" + jobId + "/test-report.html"
			println("INFO: Start downloading test report from URL: " + fileUrl)
			var reportFilePath = os.Getenv("BITRISE_DEPLOY_DIR") + "/test-report.html"
			downloadError := downloadFile(reportFilePath, fileUrl)
			if downloadError == nil {
				log.Println("INFO: Download test report successfully." + reportFilePath)
			} else {
				log.Println("ERROR: Test report download failed with error:")
				log.Println(downloadError)
			}
		} else {
			log.Println("ERROR: Scriptless automation has failed with error: " + errorMessage)
		}
	}
}

func getRequestHeader(stepConfig *model.StepConfig) map[string]string {
	var executorBasicAuth = strings.Join([]string{stepConfig.GetExecutorUsername(), stepConfig.GetExecutorPassword()}, ":")
	var executorBasicAuthEncoded = utils.Base64Encode(executorBasicAuth)

	var headers = map[string]string{}
	headers["authorization"] = "Basic " + executorBasicAuthEncoded
	headers["content-type"] = "application/json"
	headers["accept"] = "application/json"

	return headers
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
