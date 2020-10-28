package ssm

import (
	"github.com/mariotoffia/ssm/internal/asm"
	"github.com/mariotoffia/ssm/internal/pms"
)

func (s *Serializer) getAndConfigurePms() (*pms.Serializer, error) {
	if s.hasconfig {
		return pms.NewFromConfig(s.config, s.service).
			SeDefaultTier(s.tier), nil
	}

	pmsRepository, err := pms.New(s.service)
	if err != nil {
		return nil, err
	}

	return pmsRepository.SeDefaultTier(s.tier), nil
}

func (s *Serializer) getAndConfigureAsm() (*asm.Serializer, error) {
	if s.hasconfig {
		return asm.NewFromConfig(s.config, s.service), nil
	}

	asmRepository, err := asm.New(s.service)
	if err != nil {
		return nil, err
	}

	return asmRepository, nil
}

func find(slice []Usage, val Usage) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
