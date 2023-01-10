package service

type AutoFreeSVC struct {
	envsWithAutoFree map[string]struct{}
}

func New(envsWithAutoFreeArr []string) *AutoFreeSVC {
	m := make(map[string]struct{})
	for _, env := range envsWithAutoFreeArr {
		m[env] = struct{}{}
	}
	return &AutoFreeSVC{
		envsWithAutoFree: m,
	}
}

func (s *AutoFreeSVC) IsAutoFree(env string) bool {
	_, ok := s.envsWithAutoFree[env]
	return ok
}
