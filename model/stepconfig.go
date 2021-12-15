package model

import (
	"os"
	"strconv"
)

type StepConfig struct {
	kobiUsername          string
	kobiApiKey            string
	executorUrl           string
	executorUsername      string
	executorPassword      string
	gitRepoUrl            string
	gitRepoBranch         string
	gitSSHKey             string
	kobiAppId             string
	useCustomDevice       bool
	deviceName            string
	devicePlatformVersion string
	devicePlatformName    string
	rootDirectory         string
	commands              string
	waitForExecution      bool
	logType               string
}

func (stepConfig *StepConfig) Init() {

	stepConfig.kobiUsername = os.Getenv("kobi_username_input")
	stepConfig.kobiApiKey = os.Getenv("kobi_apikey_input")
	stepConfig.executorUrl = os.Getenv("executor_url_input")
	stepConfig.executorUsername = os.Getenv("executor_username_input")
	stepConfig.executorPassword = os.Getenv("executor_password_input")
	stepConfig.gitRepoUrl = os.Getenv("git_repo_url_input")
	stepConfig.gitRepoBranch = os.Getenv("git_repo_branch_input")
	stepConfig.gitSSHKey = os.Getenv("git_repo_ssh_key_input")
	stepConfig.kobiAppId = os.Getenv("app_id_input")
	stepConfig.useCustomDevice, _ = strconv.ParseBool(os.Getenv("use_custom_device_input"))
	stepConfig.deviceName = os.Getenv("device_name_input")
	stepConfig.devicePlatformVersion = os.Getenv("device_platform_version_input")
	stepConfig.devicePlatformName = os.Getenv("device_platform_input")
	stepConfig.rootDirectory = os.Getenv("root_directory_input")
	stepConfig.commands = os.Getenv("command_input")
	stepConfig.waitForExecution, _ = strconv.ParseBool(os.Getenv("wait_for_execution_input"))

	switch os.Getenv("log_type_input") {
	case "output":
		stepConfig.logType = "out"
	case "error":
		stepConfig.logType = "error"
	default:
		stepConfig.logType = "all"
	}
}

func (stepConfig *StepConfig) GetKobiUsername() string {
	return stepConfig.kobiUsername
}

func (stepConfig *StepConfig) GetKobiPassword() string {
	return stepConfig.kobiApiKey
}

func (stepConfig *StepConfig) GetExecutorUrl() string {
	return stepConfig.executorUrl
}

func (stepConfig *StepConfig) GetExecutorUsername() string {
	return stepConfig.executorUsername
}

func (stepConfig *StepConfig) GetExecutorPassword() string {
	return stepConfig.executorPassword
}

func (stepConfig *StepConfig) GetGitRepoUrl() string {
	return stepConfig.gitRepoUrl
}

func (stepConfig *StepConfig) GetGitRepoBranch() string {
	return stepConfig.gitRepoBranch
}

func (stepConfig *StepConfig) GetGitSSHKey() string {
	return stepConfig.gitSSHKey
}

func (stepConfig *StepConfig) GetKobiAppId() string {
	return stepConfig.kobiAppId
}

func (stepConfig *StepConfig) IsUseCustomDevices() bool {
	return stepConfig.useCustomDevice
}

func (stepConfig *StepConfig) GetDeviceName() string {
	return stepConfig.deviceName
}

func (stepConfig *StepConfig) GetDevicePlatformVersion() string {
	return stepConfig.devicePlatformVersion
}

func (stepConfig *StepConfig) GetDevicePlatformname() string {
	return stepConfig.devicePlatformName
}

func (stepConfig *StepConfig) GetRootDirectory() string {
	return stepConfig.rootDirectory
}

func (stepConfig *StepConfig) GetCommands() string {
	return stepConfig.commands
}

func (stepConfig *StepConfig) IsWaitForExecution() bool {
	return stepConfig.waitForExecution
}

func (stepConfig *StepConfig) GetLogType() string {
	return stepConfig.logType
}
