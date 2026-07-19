package safety

import (
	"regexp"
	"strings"
)

type Result struct {
	Tier   Tier
	Label  string
	Blocks []string
}

func Check(script string, mode string) *Result {
	res := &Result{Tier: TierSafe}
	lines := strings.Split(script, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		for _, p := range Patterns {
			if p.Tier == TierSafe {
				continue
			}

			re := regexp.MustCompile(p.Pattern)
			if re.MatchString(line) {
				if p.Tier == TierDangerous {
					res.Tier = TierDangerous
					res.Label = "Blocked"
					res.Blocks = append(res.Blocks, line)
					break
				}
				if p.Tier == TierRisky && res.Tier != TierDangerous {
					res.Tier = TierRisky
					res.Label = "Needs confirmation"
					res.Blocks = append(res.Blocks, line)
				}
			}
		}
	}

	return res
}

func (r *Result) IsDangerous() bool {
	return r.Tier == TierDangerous
}

func (r *Result) IsRisky() bool {
	return r.Tier == TierRisky
}

func (r *Result) IsSafe() bool {
	return r.Tier == TierSafe
}
