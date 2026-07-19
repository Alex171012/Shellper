package safety

type Tier int

const (
	TierSafe Tier = iota
	TierRisky
	TierDangerous
)

func (t Tier) String() string {
	switch t {
	case TierSafe:
		return "safe"
	case TierRisky:
		return "risky"
	case TierDangerous:
		return "dangerous"
	default:
		return "unknown"
	}
}

type Pattern struct {
	Pattern string
	Tier    Tier
	Label   string
}

var Patterns = []Pattern{
	{Pattern: `rm\s+-rf\s+/`, Tier: TierDangerous, Label: "rm -rf /"},
	{Pattern: `rm\s+-rf\s+--no-preserve-root`, Tier: TierDangerous, Label: "rm --no-preserve-root"},
	{Pattern: `rm\s+-rf\s+\.`, Tier: TierDangerous, Label: "rm -rf ."},
	{Pattern: `rm\s+-rf\s+\*`, Tier: TierDangerous, Label: "rm -rf *"},
	{Pattern: `sudo\s+`, Tier: TierDangerous, Label: "sudo command"},
	{Pattern: `:\(\)\s*\{`, Tier: TierDangerous, Label: "fork bomb"},
	{Pattern: `\|\s*sh\b`, Tier: TierDangerous, Label: "pipe to shell"},
	{Pattern: `\|\s*bash\b`, Tier: TierDangerous, Label: "pipe to shell"},
	{Pattern: `\|\s*zsh\b`, Tier: TierDangerous, Label: "pipe to shell"},
	{Pattern: `> /dev/sd`, Tier: TierDangerous, Label: "write to block device"},
	{Pattern: `dd\s+if=/dev/zero`, Tier: TierDangerous, Label: "dd zero write"},
	{Pattern: `mkfs\.`, Tier: TierDangerous, Label: "filesystem creation"},
	{Pattern: `chmod\s+777\s+/`, Tier: TierDangerous, Label: "chmod 777 root"},
	{Pattern: `chmod\s+-R\s+777`, Tier: TierDangerous, Label: "recursive chmod 777"},
	{Pattern: `chown\s+-R\s+`, Tier: TierDangerous, Label: "recursive chown"},
	{Pattern: `mv\s+/\s+`, Tier: TierDangerous, Label: "move root"},
	{Pattern: `shutdown\b`, Tier: TierDangerous, Label: "shutdown"},
	{Pattern: `halt\b`, Tier: TierDangerous, Label: "halt"},
	{Pattern: `reboot\b`, Tier: TierDangerous, Label: "reboot"},
	{Pattern: `poweroff\b`, Tier: TierDangerous, Label: "poweroff"},
	{Pattern: `init\s+0`, Tier: TierDangerous, Label: "init 0"},
	{Pattern: `init\s+6`, Tier: TierDangerous, Label: "init 6"},
	{Pattern: `> /dev/sd`, Tier: TierDangerous, Label: "overwrite block device"},
	{Pattern: `wget\s+.*\||curl\s+.*\|`, Tier: TierDangerous, Label: "download and pipe"},
	{Pattern: `\|\s*sudo\s+`, Tier: TierDangerous, Label: "pipe to sudo"},

	{Pattern: `rm\s+`, Tier: TierRisky, Label: "rm"},
	{Pattern: `mv\s+`, Tier: TierRisky, Label: "mv"},
	{Pattern: `cp\s+`, Tier: TierRisky, Label: "cp"},
	{Pattern: `chmod\s+`, Tier: TierRisky, Label: "chmod"},
	{Pattern: `chown\s+`, Tier: TierRisky, Label: "chown"},
	{Pattern: `mkfs\s+`, Tier: TierRisky, Label: "mkfs"},
	{Pattern: `dd\s+`, Tier: TierRisky, Label: "dd"},
	{Pattern: `kill\s+`, Tier: TierRisky, Label: "kill"},
	{Pattern: `systemctl\s+`, Tier: TierRisky, Label: "systemctl"},
	{Pattern: `service\s+`, Tier: TierRisky, Label: "service"},
	{Pattern: `apt\s+install\s+`, Tier: TierRisky, Label: "apt install"},
	{Pattern: `pacman\s+-S\s+`, Tier: TierRisky, Label: "pacman install"},
	{Pattern: `dnf\s+install\s+`, Tier: TierRisky, Label: "dnf install"},
	{Pattern: `yum\s+install\s+`, Tier: TierRisky, Label: "yum install"},
	{Pattern: `pip\s+install\s+`, Tier: TierRisky, Label: "pip install"},
	{Pattern: `npm\s+install\s+-g\s+`, Tier: TierRisky, Label: "npm global install"},
	{Pattern: `cargo\s+install\s+`, Tier: TierRisky, Label: "cargo install"},
	{Pattern: `go\s+install\s+`, Tier: TierRisky, Label: "go install"},
	{Pattern: `ln\s+-s\s+`, Tier: TierRisky, Label: "symlink"},
	{Pattern: `mount\s+`, Tier: TierRisky, Label: "mount"},
	{Pattern: `umount\s+`, Tier: TierRisky, Label: "umount"},
	{Pattern: `passwd\s+`, Tier: TierRisky, Label: "passwd"},
	{Pattern: `useradd\s+`, Tier: TierRisky, Label: "useradd"},
	{Pattern: `userdel\s+`, Tier: TierRisky, Label: "userdel"},
	{Pattern: `groupadd\s+`, Tier: TierRisky, Label: "groupadd"},
	{Pattern: `groupdel\s+`, Tier: TierRisky, Label: "groupdel"},
	{Pattern: `ufw\s+`, Tier: TierRisky, Label: "ufw"},
	{Pattern: `iptables\s+`, Tier: TierRisky, Label: "iptables"},
	{Pattern: `docker\s+`, Tier: TierRisky, Label: "docker"},
	{Pattern: `journalctl\s+--rotate`, Tier: TierRisky, Label: "journalctl rotate"},

	{Pattern: `\bls\b`, Tier: TierSafe, Label: "ls"},
	{Pattern: `\bcat\b`, Tier: TierSafe, Label: "cat"},
	{Pattern: `\becho\b`, Tier: TierSafe, Label: "echo"},
	{Pattern: `\btouch\b`, Tier: TierSafe, Label: "touch"},
	{Pattern: `\bpwd\b`, Tier: TierSafe, Label: "pwd"},
	{Pattern: `\bgrep\b`, Tier: TierSafe, Label: "grep"},
	{Pattern: `\bfind\b`, Tier: TierSafe, Label: "find"},
	{Pattern: `\bhead\b`, Tier: TierSafe, Label: "head"},
	{Pattern: `\btail\b`, Tier: TierSafe, Label: "tail"},
	{Pattern: `\bwc\b`, Tier: TierSafe, Label: "wc"},
	{Pattern: `\bsort\b`, Tier: TierSafe, Label: "sort"},
	{Pattern: `\buniq\b`, Tier: TierSafe, Label: "uniq"},
	{Pattern: `\bwhich\b`, Tier: TierSafe, Label: "which"},
	{Pattern: `\bwhoami\b`, Tier: TierSafe, Label: "whoami"},
	{Pattern: `\bdate\b`, Tier: TierSafe, Label: "date"},
	{Pattern: `\bcal\b`, Tier: TierSafe, Label: "cal"},
	{Pattern: `\bdf\b`, Tier: TierSafe, Label: "df"},
	{Pattern: `\bdu\b`, Tier: TierSafe, Label: "du"},
	{Pattern: `\bfree\b`, Tier: TierSafe, Label: "free"},
	{Pattern: `\buptime\b`, Tier: TierSafe, Label: "uptime"},
	{Pattern: `\benv\b`, Tier: TierSafe, Label: "env"},
	{Pattern: `\bprintenv\b`, Tier: TierSafe, Label: "printenv"},
	{Pattern: `\bhistory\b`, Tier: TierSafe, Label: "history"},
	{Pattern: `\blocale\b`, Tier: TierSafe, Label: "locale"},
	{Pattern: `\buname\b`, Tier: TierSafe, Label: "uname"},
	{Pattern: `\bhostname\b`, Tier: TierSafe, Label: "hostname"},
	{Pattern: `\bid\b`, Tier: TierSafe, Label: "id"},
	{Pattern: `\bwho\b`, Tier: TierSafe, Label: "who"},
	{Pattern: `\bw\b`, Tier: TierSafe, Label: "w"},
	{Pattern: `\busers\b`, Tier: TierSafe, Label: "users"},
	{Pattern: `\bgroups\b`, Tier: TierSafe, Label: "groups"},
	{Pattern: `\blogs\b`, Tier: TierSafe, Label: "logs"},
	{Pattern: `\btype\b`, Tier: TierSafe, Label: "type"},
	{Pattern: `\bcommand\b`, Tier: TierSafe, Label: "command"},
	{Pattern: `\bhash\b`, Tier: TierSafe, Label: "hash"},
	{Pattern: `\bbind\b`, Tier: TierSafe, Label: "bind"},
	{Pattern: `\benable\b`, Tier: TierSafe, Label: "enable"},
	{Pattern: `\bhelp\b`, Tier: TierSafe, Label: "help"},
	{Pattern: `\bbasename\b`, Tier: TierSafe, Label: "basename"},
	{Pattern: `\bdirname\b`, Tier: TierSafe, Label: "dirname"},
	{Pattern: `\brealpath\b`, Tier: TierSafe, Label: "realpath"},
	{Pattern: `\bseq\b`, Tier: TierSafe, Label: "seq"},
	{Pattern: `\bshuf\b`, Tier: TierSafe, Label: "shuf"},
	{Pattern: `\bcut\b`, Tier: TierSafe, Label: "cut"},
	{Pattern: `\btr\b`, Tier: TierSafe, Label: "tr"},
	{Pattern: `\bfold\b`, Tier: TierSafe, Label: "fold"},
	{Pattern: `\bpaste\b`, Tier: TierSafe, Label: "paste"},
	{Pattern: `\bjoin\b`, Tier: TierSafe, Label: "join"},
	{Pattern: `\bexpand\b`, Tier: TierSafe, Label: "expand"},
	{Pattern: `\bunexpand\b`, Tier: TierSafe, Label: "unexpand"},
	{Pattern: `\btee\b`, Tier: TierSafe, Label: "tee"},
	{Pattern: `\bxargs\b`, Tier: TierSafe, Label: "xargs"},
	{Pattern: `\bprintf\b`, Tier: TierSafe, Label: "printf"},
	{Pattern: `\byes\b`, Tier: TierSafe, Label: "yes"},
	{Pattern: `\bfactor\b`, Tier: TierSafe, Label: "factor"},
	{Pattern: `\bnproc\b`, Tier: TierSafe, Label: "nproc"},
	{Pattern: `\btrue\b`, Tier: TierSafe, Label: "true"},
	{Pattern: `\bfalse\b`, Tier: TierSafe, Label: "false"},
	{Pattern: `\bnohup\b`, Tier: TierSafe, Label: "nohup"},
	{Pattern: `\btimeout\b`, Tier: TierSafe, Label: "timeout"},
	{Pattern: `\bwatch\b`, Tier: TierSafe, Label: "watch"},
	{Pattern: `\btee\b`, Tier: TierSafe, Label: "tee"},
	{Pattern: `\bstdbuf\b`, Tier: TierSafe, Label: "stdbuf"},
	{Pattern: `\btput\b`, Tier: TierSafe, Label: "tput"},
	{Pattern: `\btty\b`, Tier: TierSafe, Label: "tty"},
	{Pattern: `\bmkdir\s+-p\b`, Tier: TierSafe, Label: "mkdir -p"},
}
