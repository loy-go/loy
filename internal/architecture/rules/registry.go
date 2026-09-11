package rules

import (
	"github.com/uloydev/loy/internal/architecture"
)

// DefaultRules returns instances of all 14 standard architecture rules.
func DefaultRules() []architecture.Rule {
	return []architecture.Rule{
		&RuleArch001{},
		&RuleArch002{},
		&RuleArch003{},
		&RuleArch004{},
		&RuleArch005{},
		&RuleArch006{},
		&RuleArch007{},
		&RuleArch008{},
		&RuleArch009{},
		&RuleArch010{},
		&RuleArch011{},
		&RuleArch012{},
		&RuleArch013{},
		&RuleArch014{},
	}
}
