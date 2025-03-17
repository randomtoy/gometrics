package model

type App string

const (
	AgentApp  App = "agent"
	ServerApp App = "server"
)

type Config struct {
	Server ServerConfig
	Agent  AgentConfig
}
