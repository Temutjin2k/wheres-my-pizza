package kitchen

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Logic() error {
	return s.repo.Get()
}
