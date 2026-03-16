package service

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) HandleGitHubEvent(eventType string, payload map[string]any) error {
	return nil
}
