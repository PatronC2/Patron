package handlers

import (
	"net"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/lib/logger"
)

// All agents use this handler on their first request

func validateOrCreateAgent(c types.ConfigurationRequest) (types.ConfigurationResponse, bool) {
	fetch, err := data.FetchOneAgent(c.AgentID)
	if err != nil {
		logger.Logf(logger.Warning, "Couldn't fetch agent: %v\n", err)
		return types.ConfigurationResponse{}, false
	}

	logger.Logf(logger.Debug, "Beacon ID: %v, Callback IP: %v, Callback Port: %v, Callback Freq: %v, Callback Jitter: %v, Agent IP: %v, Username: %v, Hostname: %v", 
		c.AgentID, c.ServerIP, c.ServerPort, c.CallbackFrequency, c.CallbackJitter, c.AgentIP, c.Username, c.Hostname)

	if fetch.AgentID == "" && c.MasterKey == "MASTERKEY" {
		logger.Logf(logger.Info, "Registering new agent: %v", c.AgentID)
		data.CreateAgent(c.AgentID, c.ServerIP, c.ServerPort, c.CallbackFrequency, c.CallbackJitter, c.AgentIP, c.Username, c.Hostname)
		data.CreateKeys(c.AgentID)
		fetch, err = data.FetchOneAgent(c.AgentID)
		if err != nil {
			logger.Logf(logger.Warning, "Couldn't fetch agent after creation: %v\n", err)
			return types.ConfigurationResponse{}, false
		}
	}

	response := types.ConfigurationResponse{
		ServerIP:         fetch.ServerIP,
		ServerPort:       fetch.ServerPort,
		CallbackFrequency: fetch.CallbackFrequency,
		CallbackJitter:   fetch.CallbackJitter,
	}

	return response, fetch.AgentID == c.AgentID
}

type ConfigurationHandler struct{}

func (h *ConfigurationHandler) Handle(request types.Request, conn net.Conn) types.Response {
    configReq, ok := request.Payload.(types.ConfigurationRequest)
    if !ok {
        return types.Response{
            Type:    types.ConfigurationResponseType,
            Payload: types.ConfigurationResponse{},
        }
    }

    configResponse, success := validateOrCreateAgent(configReq)
    if !success {
        return types.Response{
            Type:    types.ConfigurationResponseType,
            Payload: types.ConfigurationResponse{},
        }
    }

	err := data.UpdateAgentCheckIn(configReq.AgentID)
	if err != nil {
		logger.Logf(logger.Error, "Could not update last callback for %v, %v", configReq.AgentID, err)
	}

    return types.Response{
        Type:    types.ConfigurationResponseType,
        Payload: configResponse,
    }
}
