package handlers

import (
	"net"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
)

// All agents use this handler on their first request

type ConfigurationHandler struct{}

func validateOrCreateAgent(c *patronobuf.ConfigurationRequest) (*patronobuf.ConfigurationResponse, bool) {
	fetch, err := data.FetchOneAgent(c.GetUuid())
	if err != nil {
		logger.Logf(logger.Warning, "Couldn't fetch agent: %v\n", err)
		return &patronobuf.ConfigurationResponse{}, false
	}

	logger.Logf(logger.Debug, "Beacon ID: %v, Callback IP: %v, Callback Port: %v, Callback Freq: %v, Callback Jitter: %v, Agent IP: %v, Username: %v, Hostname: %v, OS: %s %s %s, CPUs: %s, Memory: %s",
		c.GetUuid(), c.GetServerip(), c.GetServerport(), c.GetCallbackfrequency(), c.GetCallbackjitter(),
		c.GetAgentip(), c.GetUsername(), c.GetHostname(), c.GetOstype(), c.GetOsbuild(), c.GetArch(),
		c.GetCpus(), c.GetMemory())

	// Register agent if new and master key matches
	if fetch == nil || fetch.GetUuid() == "" && c.GetMasterkey() == "MASTERKEY" {
		logger.Logf(logger.Info, "Registering new agent: %v", c.GetUuid())

		if err := data.CreateAgent(c); err != nil {
			logger.Logf(logger.Error, "Failed to create agent: %v", err)
			return &patronobuf.ConfigurationResponse{}, false
		}

		data.CreateKeys(c.GetUuid())

		fetch, err = data.FetchOneAgent(c.GetUuid())
		if err != nil {
			logger.Logf(logger.Warning, "Couldn't fetch agent after creation: %v\n", err)
			return &patronobuf.ConfigurationResponse{}, false
		}
	}

	// Build the response
	resp := &patronobuf.ConfigurationResponse{
		Uuid:              fetch.GetUuid(),
		Serverip:          fetch.GetServerip(),
		Serverport:        fetch.GetServerport(),
		Callbackfrequency: fetch.GetCallbackfrequency(),
		Callbackjitter:    fetch.GetCallbackjitter(),
	}

	return resp, fetch.GetUuid() == c.GetUuid()
}

func (h *ConfigurationHandler) Handle(request *patronobuf.Request, conn net.Conn) *patronobuf.Response {
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

	respData, ok := validateOrCreateAgent(payload)
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
