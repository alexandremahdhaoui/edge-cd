package cloudinit

import (
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/yaml"
)

// userData := fmt.Sprintf(`
// #cloud-config
// hostname: %s
// users:
//   - name: %s
//     sudo: ALL=(ALL) NOPASSWD:ALL
//     shell: /bin/bash
//     ssh_authorized_keys:
//       - %q
// ssh_keys:
//   rsa_private: |
//     %s
//   rsa_public: %q
// ssh_deletekeys: false
// `, vmName, targetUser, vmSSHPublicKey, readFileAndIndent(vmSSHKeyPath))

type User struct {
	Name              string   `json:"name"`
	Sudo              string   `json:"sudo"`
	Shell             string   `json:"shell"`
	SSHAuthorizedKeys []string `json:"ssh_authorized_keys"`
}

func NewUser(name string, authorizedKeyPathList ...string) (User, error) {
	authorizedKeys := make([]string, 0, len(authorizedKeyPathList))
	for _, path := range authorizedKeyPathList {
		b, err := os.ReadFile(path)
		if err != nil {
			return User{}, fmt.Errorf("ERROR: Failed to read file: %v", err)
		}
		authorizedKeys = append(authorizedKeys, string(b))
	}
	return User{
		Name:              name,
		Sudo:              "ALL=(ALL) NOPASSWD:ALL",
		Shell:             "/bin/bash",
		SSHAuthorizedKeys: authorizedKeys,
	}, nil
}

type SSHKeys struct {
	RSAPrivate string `json:"rsa_private"`
	RSAPublic  string `json:"rsa_public"`
}

type UserData struct {
	Hostname      string  `json:"hostname"`
	Users         []User  `json:"users"`
	SSHKeys       SSHKeys `json:"ssh_keys"`
	SSHDeleteKeys bool    `json:"ssh_deletekeys"`
}

func (ud UserData) Render() (string, error) {
	if ud.SSHKeys.RSAPublic != "" {
		ud.SSHDeleteKeys = true
	}

	b, err := yaml.Marshal(ud)
	if err != nil {
		return "", fmt.Errorf("Cannot render cloud-config from UserData: %v", err)
	}
	return fmt.Sprintf("#cloud-config\n%s", string(b)), nil
}

func NewRSAKeyFromPrivateKeyFile(privateKeyPath string) (SSHKeys, error) {
	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return SSHKeys{}, fmt.Errorf("Cannot read SSH private key at %s", privateKeyPath)
	}

	// bit hacky
	publicKeyPath := privateKeyPath + ".pub"
	if _, err := os.Stat(publicKeyPath); os.IsNotExist(err) {
		return SSHKeys{}, fmt.Errorf("SSH public key not found at %s", publicKeyPath)
	}

	publicKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return SSHKeys{}, fmt.Errorf("failed to read SSH public key: %w", err)
	}

	return SSHKeys{
		RSAPrivate: string(privateKey),
		RSAPublic:  strings.TrimSpace(string(publicKey)),
	}, nil
}
