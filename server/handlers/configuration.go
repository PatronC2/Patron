package handlers

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

type StartupHandler struct{}
type ConfigurationHandler struct{}

func configValue(stored string, reported string) string {
	if stored == "" || stored == "Unknown" {
		return reported
	}
	return stored
}

func calculateSleepSeconds(callbackFrequency string, callbackJitter string) int64 {
	frequency, err := strconv.Atoi(callbackFrequency)
	if err != nil || frequency < 0 {
		frequency = 0
	}

	jitter, err := strconv.Atoi(callbackJitter)
	if err != nil || jitter < 0 {
		jitter = 0
	}

	baseTime := float64(frequency)
	jitterPercent := float64(jitter) * 0.01
	variance := baseTime * jitterPercent * rand.Float64()
	finalInterval := baseTime - (jitterPercent * baseTime) + 2*variance
	if finalInterval < 0 {
		return 0
	}

	return int64(finalInterval)
}

func (h *StartupHandler) Handle(request *patronobuf.Request, stream types.CommonStream) *patronobuf.Response {
	payload := request.GetStartup()
	if payload == nil {
		logger.Logf(logger.Debug, "Startup payload is nil")
		return &patronobuf.Response{
			Type: patronobuf.ResponseType_STARTUP_RESPONSE,
			Payload: &patronobuf.Response_StartupResponse{
				StartupResponse: &patronobuf.StartupResponse{},
			},
		}
	}

	logger.Logf(logger.Debug, "Startup request: filepath=%v, agent_ip=%v, username=%v, hostname=%v, OS=%s %s %s, CPUs=%s, Memory=%s, Capabilities=%v",
		payload.GetFilepath(), payload.GetAgentip(), payload.GetUsername(), payload.GetHostname(),
		payload.GetOstype(), payload.GetOsbuild(), payload.GetArch(), payload.GetCpus(), payload.GetMemory(),
		payload.GetCapabilities())

	agentUUID, err := data.FindOrCreateAgentFromStartup(payload)
	if err != nil {
		logger.Logf(logger.Error, "Failed to resolve startup UUID: %v", err)
		return &patronobuf.Response{
			Type: patronobuf.ResponseType_STARTUP_RESPONSE,
			Payload: &patronobuf.Response_StartupResponse{
				StartupResponse: &patronobuf.StartupResponse{},
			},
		}
	}

	return &patronobuf.Response{
		Type: patronobuf.ResponseType_STARTUP_RESPONSE,
		Payload: &patronobuf.Response_StartupResponse{
			StartupResponse: &patronobuf.StartupResponse{Uuid: agentUUID},
		},
	}
}

func validateAgentConfiguration(c *patronobuf.ConfigurationRequest) (*patronobuf.ConfigurationResponse, bool) {
	fetch, err := data.FetchOneAgent(c.GetUuid())
	if err != nil {
		logger.Logf(logger.Warning, "Couldn't fetch agent: %v\n", err)
		return &patronobuf.ConfigurationResponse{}, false
	}
	if fetch == nil || fetch.GetUuid() == "" {
		logger.Logf(logger.Warning, "Configuration request for unknown agent: %v", c.GetUuid())
		return &patronobuf.ConfigurationResponse{}, false
	}

	logger.Logf(logger.Debug, "Beacon ID: %v, Callback IP: %v, Callback Port: %v, Callback Freq: %v, Callback Jitter: %v, Transport Protocol: %s",
		c.GetUuid(), c.GetServerip(), c.GetServerport(), c.GetCallbackfrequency(), c.GetCallbackjitter(),
		c.GetTransportprotocol())

	serverIP := configValue(fetch.GetServerip(), c.GetServerip())
	serverPort := configValue(fetch.GetServerport(), c.GetServerport())
	callbackFrequency := configValue(fetch.GetCallbackfrequency(), c.GetCallbackfrequency())
	callbackJitter := configValue(fetch.GetCallbackjitter(), c.GetCallbackjitter())
	transportProtocol := configValue(fetch.GetTransportprotocol(), c.GetTransportprotocol())

	data.UpdateAgentConfigIfMissing(
		c.GetUuid(),
		c.GetServerip(),
		c.GetServerport(),
		c.GetCallbackfrequency(),
		c.GetCallbackjitter(),
		c.GetTransportprotocol(),
	)

	sleepSeconds := calculateSleepSeconds(callbackFrequency, callbackJitter)
	nextCallback := time.Now().UTC().Add(time.Duration(sleepSeconds) * time.Second)

	data.UpdateAgentNextCallback(c.GetUuid(), nextCallback)

	fetch, err = data.FetchOneAgent(c.GetUuid())
	if err != nil {
		logger.Logf(logger.Warning, "Couldn't fetch agent after config update: %v\n", err)
		return &patronobuf.ConfigurationResponse{}, false
	}
	if fetch == nil || fetch.GetUuid() == "" {
		logger.Logf(logger.Warning, "Configuration update lost agent: %v", c.GetUuid())
		return &patronobuf.ConfigurationResponse{}, false
	}

	resp := &patronobuf.ConfigurationResponse{
		Serverip:          serverIP,
		Serverport:        serverPort,
		Transportprotocol: transportProtocol,
		SleepSeconds:      sleepSeconds,
	}

	return resp, fetch.GetUuid() == c.GetUuid()
}

func (h *ConfigurationHandler) Handle(request *patronobuf.Request, stream types.CommonStream) *patronobuf.Response {
	payload := request.GetConfiguration()

	if payload == nil {
		logger.Logf(logger.Debug, "Payload is nil")
		return &patronobuf.Response{
			Type: patronobuf.ResponseType_CONFIGURATION_RESPONSE,
			Payload: &patronobuf.Response_ConfigurationResponse{
				ConfigurationResponse: &patronobuf.ConfigurationResponse{},
			},
		}
	}

	respData, ok := validateAgentConfiguration(payload)
	if !ok {
		logger.Logf(logger.Debug, "Failed to create agent in DB")
		return &patronobuf.Response{
			Type: patronobuf.ResponseType_CONFIGURATION_RESPONSE,
			Payload: &patronobuf.Response_ConfigurationResponse{
				ConfigurationResponse: &patronobuf.ConfigurationResponse{},
			},
		}
	}

	_ = data.UpdateAgentCheckIn(payload)

	logger.Logf(logger.Debug, "Sending configuration response: %+v", respData)

	return &patronobuf.Response{
		Type: patronobuf.ResponseType_CONFIGURATION_RESPONSE,
		Payload: &patronobuf.Response_ConfigurationResponse{
			ConfigurationResponse: respData,
		},
	}
}
