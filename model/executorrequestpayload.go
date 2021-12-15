package model

import "strings"

type DesiredCaps struct {
	DeviceName      string `json:"deviceName,omitempty"`
	PlatformVersion string `json:"platformVersion,omitempty"`
	PlatformName    string `json:"platformName,omitempty"`
	AppId           string `json:"appId,omitempty"`
}

type TestConfig struct {
	Git           string   `json:"git"`
	Ssh           string   `json:"ssh"`
	Branch        string   `json:"branch"`
	RootDirectory string   `json:"rootDirectory,omitempty`
	Commands      []string `json:"commands"`
}

type BitriseConfig struct {
	ReleaseId string `json:"releaseId"`
}

type ExecutorRequestPayload struct {
	DesiredCaps DesiredCaps   `json:"desiredCaps,omitempty"`
	TestConfig  TestConfig    `json:"testConfig"`
	AzureConfig BitriseConfig `json:"azureConfig"`
}

type JobResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func BuildExecutorRequestPayload(e *ExecutorRequestPayload, s *StepConfig) {
	// TestConfig
	e.TestConfig.Git = s.gitRepoUrl
	e.TestConfig.Branch = s.gitRepoBranch
	e.TestConfig.Ssh = s.gitSSHKey
	e.TestConfig.Commands = strings.Split(s.commands, "\n")
	e.TestConfig.RootDirectory = s.rootDirectory

	// DesiredCaps
	if s.useCustomDevice {
		e.DesiredCaps.DeviceName = s.deviceName
		e.DesiredCaps.PlatformName = s.devicePlatformName
		e.DesiredCaps.PlatformVersion = s.devicePlatformVersion
		e.DesiredCaps.AppId = s.kobiAppId
	}

	// BitriseConfig
	e.AzureConfig.ReleaseId = "123"
}
