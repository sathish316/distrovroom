package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	sshAgentSocketPattern = regexp.MustCompile(`SSH_AUTH_SOCK=([^;]+);`)
	sshAgentPIDPattern    = regexp.MustCompile(`SSH_AGENT_PID=([0-9]+);`)
)

func newSetupCommand() *cobra.Command {
	setup := &cobra.Command{
		Use:   "setup",
		Short: "Set up tools and credentials for your environment",
	}
	setup.AddCommand(newGitHubSSHKeysCommand())
	return setup
}

func newGitHubSSHKeysCommand() *cobra.Command {
	var email string
	var keyPath string

	command := &cobra.Command{
		Use:   "github-ssh-keys",
		Short: "Generate an SSH key and add it to ssh-agent for GitHub",
		Long: "Generate an Ed25519 SSH key when one is not already present, add it to ssh-agent, " +
			"and print the public key so you can add it to your GitHub account. Existing keys are reused and never overwritten.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if email == "" {
				email = configuredGitEmail()
			}
			return setupGitHubSSHKeys(cmd, email, keyPath)
		},
	}
	command.Flags().StringVar(&email, "email", "", "email to use as the SSH key comment (defaults to git config user.email)")
	command.Flags().StringVar(&keyPath, "key-path", "", "SSH private key path (defaults to ~/.ssh/id_ed25519)")
	return command
}

func setupGitHubSSHKeys(cmd *cobra.Command, email, requestedKeyPath string) error {
	keygenPath, err := exec.LookPath("ssh-keygen")
	if err != nil {
		return fmt.Errorf("ssh-keygen was not found; install OpenSSH and try again")
	}
	sshAddPath, err := exec.LookPath("ssh-add")
	if err != nil {
		return fmt.Errorf("ssh-add was not found; install OpenSSH and try again")
	}
	if runtime.GOOS == "windows" {
		if err := checkWindowsSSHAgent(sshAddPath); err != nil {
			return err
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home directory: %w", err)
	}
	keyPath, err := resolveSSHKeyPath(requestedKeyPath, home)
	if err != nil {
		return err
	}
	publicKeyPath := keyPath + ".pub"
	privateExists, err := pathExists(keyPath)
	if err != nil {
		return fmt.Errorf("check SSH key path: %w", err)
	}
	publicExists, err := pathExists(publicKeyPath)
	if err != nil {
		return fmt.Errorf("check public key path: %w", err)
	}
	if privateExists != publicExists {
		return fmt.Errorf("only one of %q and %q exists; choose another --key-path or repair the key pair", keyPath, publicKeyPath)
	}

	if !privateExists {
		if runtime.GOOS != "windows" {
			if _, err := exec.LookPath("ssh-agent"); err != nil {
				return fmt.Errorf("ssh-agent was not found; install OpenSSH and try again")
			}
		}
		if err := os.MkdirAll(filepath.Dir(keyPath), 0o700); err != nil {
			return fmt.Errorf("create SSH directory: %w", err)
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Generating an Ed25519 SSH key at %s. Enter a passphrase when prompted.\n", keyPath); err != nil {
			return err
		}
		args := []string{"-t", "ed25519", "-f", keyPath}
		if email != "" {
			args = append(args, "-C", email)
		}
		keygen := exec.Command(keygenPath, args...)
		keygen.Stdin = cmd.InOrStdin()
		keygen.Stdout = cmd.OutOrStdout()
		keygen.Stderr = cmd.ErrOrStderr()
		if err := keygen.Run(); err != nil {
			return fmt.Errorf("generate SSH key: %w", err)
		}
	} else if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Using existing SSH key at %s.\n", keyPath); err != nil {
		return err
	}

	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("read public key %q: %w", publicKeyPath, err)
	}
	if strings.TrimSpace(string(publicKey)) == "" {
		return fmt.Errorf("public key %q is empty", publicKeyPath)
	}

	agentVars, startedAgent, err := sshAgentEnvironment(cmd)
	if err != nil {
		return err
	}
	sshAdd := exec.Command(sshAddPath, keyPath)
	sshAdd.Env = environmentWith(agentVars)
	sshAdd.Stdin = cmd.InOrStdin()
	sshAdd.Stdout = cmd.OutOrStdout()
	sshAdd.Stderr = cmd.ErrOrStderr()
	if err := sshAdd.Run(); err != nil {
		if startedAgent {
			stopSSHAgent(agentVars)
		}
		return fmt.Errorf("add SSH key to ssh-agent: %w", err)
	}

	if startedAgent {
		if _, err := fmt.Fprintln(cmd.OutOrStdout(), "\nThe SSH agent was started for this command. To reuse it in this shell, run:"); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "export SSH_AUTH_SOCK=%s\nexport SSH_AGENT_PID=%s\n",
			shellQuote(agentVars["SSH_AUTH_SOCK"]), shellQuote(agentVars["SSH_AGENT_PID"])); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(cmd.OutOrStdout(), "\nAdd this public key to your GitHub account at https://github.com/settings/ssh/new:\n%s\n", strings.TrimSpace(string(publicKey))); err != nil {
		return err
	}
	return nil
}

func configuredGitEmail() string {
	output, err := exec.Command("git", "config", "user.email").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func resolveSSHKeyPath(requestedPath, home string) (string, error) {
	if requestedPath == "" {
		requestedPath = filepath.Join(home, ".ssh", "id_ed25519")
	} else if requestedPath == "~" {
		requestedPath = home
	} else if strings.HasPrefix(requestedPath, "~"+string(filepath.Separator)) {
		requestedPath = filepath.Join(home, requestedPath[2:])
	}
	absolutePath, err := filepath.Abs(requestedPath)
	if err != nil {
		return "", fmt.Errorf("resolve SSH key path: %w", err)
	}
	return absolutePath, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func sshAgentEnvironment(cmd *cobra.Command) (map[string]string, bool, error) {
	if runtime.GOOS == "windows" {
		return nil, false, nil
	}
	if socket := os.Getenv("SSH_AUTH_SOCK"); socket != "" {
		vars := map[string]string{"SSH_AUTH_SOCK": socket}
		if pid := os.Getenv("SSH_AGENT_PID"); pid != "" {
			vars["SSH_AGENT_PID"] = pid
		}
		return vars, false, nil
	}

	var output strings.Builder
	agent := exec.Command("ssh-agent", "-s")
	agent.Stdout = &output
	agent.Stderr = cmd.ErrOrStderr()
	if err := agent.Run(); err != nil {
		return nil, false, fmt.Errorf("start ssh-agent: %w", err)
	}
	contents := output.String()
	socketMatch := sshAgentSocketPattern.FindStringSubmatch(contents)
	pidMatch := sshAgentPIDPattern.FindStringSubmatch(contents)
	if len(socketMatch) != 2 || len(pidMatch) != 2 {
		return nil, false, fmt.Errorf("could not read SSH_AUTH_SOCK and SSH_AGENT_PID from ssh-agent output")
	}
	return map[string]string{
		"SSH_AUTH_SOCK": socketMatch[1],
		"SSH_AGENT_PID": pidMatch[1],
	}, true, nil
}

func checkWindowsSSHAgent(sshAddPath string) error {
	command := exec.Command(sshAddPath, "-l")
	output, err := command.CombinedOutput()
	if err == nil || strings.Contains(strings.ToLower(string(output)), "no identities") {
		return nil
	}
	return fmt.Errorf("the Windows OpenSSH agent is not responding; start it in an elevated PowerShell window with `Get-Service -Name ssh-agent | Set-Service -StartupType Manual` and `Start-Service ssh-agent`, then rerun this command")
}

func environmentWith(overrides map[string]string) []string {
	updated := make([]string, 0, len(os.Environ())+len(overrides))
	seen := make(map[string]bool, len(overrides))
	for _, entry := range os.Environ() {
		key, _, ok := strings.Cut(entry, "=")
		if value, replace := overrides[key]; ok && replace {
			if !seen[key] {
				updated = append(updated, key+"="+value)
				seen[key] = true
			}
			continue
		}
		updated = append(updated, entry)
	}
	keys := make([]string, 0, len(overrides))
	for key := range overrides {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if !seen[key] {
			updated = append(updated, key+"="+overrides[key])
		}
	}
	return updated
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func stopSSHAgent(agentVars map[string]string) {
	stop := exec.Command("ssh-agent", "-k")
	stop.Env = environmentWith(agentVars)
	stop.Stdout = io.Discard
	stop.Stderr = io.Discard
	_ = stop.Run()
}
