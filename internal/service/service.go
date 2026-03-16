package service

import (
	"log"

	"github.com/University-Of-Sri-Jayewardenepura/Notify/internal/github"
)

type Dispatcher interface {
	DispatchGitHubEvent(eventType string, payload map[string]any) error
}

type loggingDispatcher struct{}

func (d loggingDispatcher) DispatchGitHubEvent(eventType string, payload map[string]any) error {
	log.Printf("accepted github event: %s", eventType)
	return nil
}

type Service struct {
	organization string
	dispatcher   Dispatcher
}

func New(organization string, dispatcher Dispatcher) *Service {
	if dispatcher == nil {
		dispatcher = NewLoggingDispatcher()
	}

	return &Service{
		organization: organization,
		dispatcher:   dispatcher,
	}
}

func NewLoggingDispatcher() Dispatcher {
	return loggingDispatcher{}
}

func (s *Service) HandleGitHubEvent(eventType string, payload map[string]any) error {
	if eventType != "ping" && !github.IsFromOrganization(payload, s.organization) {
		return nil
	}

	return s.dispatcher.DispatchGitHubEvent(eventType, payload)
}
