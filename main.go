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
	setEnv()
	stepConfig := new(model.StepConfig)
	stepConfig.Init()
	log.Println("nhc step v2")

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

		for range ticker.C {
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
		log.Println("Execute session with create session and generate test run is successful")
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
	log.Println("Check scriptless status...")

	var isTimeout = false
	var scriptlessResponse *model.ScriptlessStatusResponse
	var waitingBeginAt = time.Now().UnixNano() / int64(time.Millisecond)
	var statusUrl = stepConfig.GetExecutorUrl() + "/jobs/" + jobId + "/scriptless/status"
	scriptlessTicker := time.NewTicker(30 * time.Second)
	client := utils.HttpClient()
	var headers = getRequestHeader(stepConfig)

	for range scriptlessTicker.C {
		var response = utils.SendRequest(client, "GET", statusUrl, headers, nil)
		json.Unmarshal(response, &scriptlessResponse)

		if len(scriptlessResponse.Messages) > 0 {
			for _, message := range scriptlessResponse.Messages {
				log.Println(message)
			}
		}

		log.Println("Scriptless Status: ", scriptlessResponse.Status)

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

	utils.ExposeEnv("SCRIPTLESS_PASSED", strconv.FormatBool(
		!isTimeout && scriptlessResponse != nil && scriptlessResponse.Error == ""))
	if isTimeout {
		log.Println("Scriptless testing is timeout")
	} else {
		if scriptlessResponse == nil {
			log.Println("Cannot get scriptless testing status")
			return
		}

		var errorMessage = scriptlessResponse.Error
		if errorMessage == "" {
			log.Println("Scriptless testing is passed")
			fileUrl := stepConfig.GetExecutorUrl() + "/" + jobId + "/scriptless-report.html"
			println("Start downloading scriptless report from URL: " + fileUrl)
			var reportFilePath = os.Getenv("BITRISE_DEPLOY_DIR") + "/scriptless-report.html"
			downloadError := downloadFile(reportFilePath, fileUrl)
			if downloadError == nil {
				log.Println("Scriptless report is available at: " + reportFilePath)
			} else {
				log.Println("Upload report failed with error:")
				log.Println(downloadError)
			}
		} else {
			log.Println("Scriptless testing is failed with error: " + errorMessage)
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

func setEnv() {
	os.Setenv("BITRISE_DEPLOY_DIR", "/Users/chuong.nguyen/pj/kobiton/bitrise-step-kobiton-execute-test/report")
	os.Setenv("kobi_username_input", "chuong777")
	os.Setenv("kobi_apikey_input", "884f4fc8-3fbe-42d0-90f3-3fa706d6554a")
	os.Setenv("executor_url_input", "http://localhost:4545")
	os.Setenv("executor_username_input", "chuong777")
	os.Setenv("executor_password_input", "884f4fc8-3fbe-42d0-90f3-3fa706d6554a")
	os.Setenv("git_repo_url_input", "https://github.com/chuong777/azure-devops-sample-java-prod.git")
	os.Setenv("git_repo_branch_input", "master")
	os.Setenv("git_repo_ssh_key_input", "-----BEGIN OPENSSH PRIVATE KEY-----\n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\n-----END OPENSSH PRIVATE KEY-----")
	os.Setenv("app_input", "kobiton-store:v117")
	os.Setenv("use_custom_device_input", "true")
	os.Setenv("device_name_input", "Nexus 6P")
	os.Setenv("device_platform_version_input", "8.0.0")
	os.Setenv("device_platform_input", "Android")
	os.Setenv("root_directory_input", "/")
	os.Setenv("command_input", "mvn test")
	os.Setenv("wait_for_execution_input", "true")
	os.Setenv("scriptless_automation", "true")
	os.Setenv("device_bundle", "20")
	os.Setenv("scriptless_timeout", "300")
}
