package plugin

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	plugin "github.com/hashicorp/go-plugin"
	"github.com/mitchellh/go-homedir"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/host2plugin"
	"github.com/tofuutils/tofulint/tofulint"
)

// Discovery searches and launches plugins according the passed configuration.
// If the plugin is not enabled, skip without starting.
// The Terraform Language plugin is treated specially. Plugins for which no version
// is specified will launch the bundled plugin instead of returning an error.
func Discovery(config *tofulint.Config) (*Plugin, error) {
	clients := map[string]*plugin.Client{}
	rulesets := map[string]*host2plugin.Client{}

	for _, pluginCfg := range config.Plugins {
		installCfg := NewInstallConfig(config, pluginCfg)
		pluginPath, err := FindPluginPath(installCfg)
		var cmd *exec.Cmd
		if os.IsNotExist(err) {
			if pluginCfg.Name == "terraform" && installCfg.ManuallyInstalled() {
				log.Print(`[INFO] Plugin "terraform" is not installed, but the bundled plugin is available.`)
				self, err := os.Executable()
				if err != nil {
					return nil, err
				}
				cmd = exec.Command(self, "--act-as-bundled-plugin")
			} else {
				if installCfg.ManuallyInstalled() {
					pluginDir, err := getPluginDir(config)
					if err != nil {
						return nil, err
					}
					return nil, fmt.Errorf(`Plugin "%s" not found in %s`, pluginCfg.Name, pluginDir)
				}
				return nil, fmt.Errorf(`Plugin "%s" not found. Did you run "tofulint --init"?`, pluginCfg.Name)
			}
		} else {
			cmd = exec.Command(pluginPath)
		}

		if pluginCfg.Enabled {
			log.Printf(`[INFO] Plugin "%s" found`, pluginCfg.Name)

			client := host2plugin.NewClient(&host2plugin.ClientOpts{
				Cmd: cmd,
			})
			rpcClient, err := client.Client()
			if err != nil {
				return nil, pluginClientError(err, pluginCfg)
			}
			raw, err := rpcClient.Dispense("ruleset")
			if err != nil {
				return nil, err
			}
			ruleset := raw.(*host2plugin.Client)

			clients[pluginCfg.Name] = client
			rulesets[pluginCfg.Name] = ruleset
		} else {
			log.Printf(`[INFO] Plugin "%s" found, but the plugin is disabled`, pluginCfg.Name)
		}
	}

	return &Plugin{RuleSets: rulesets, clients: clients}, nil
}

// FindPluginPath returns the plugin binary path.
func FindPluginPath(config *InstallConfig) (string, error) {
	dir, err := getPluginDir(config.globalConfig)
	if err != nil {
		return "", err
	}

	path, err := findPluginPath(filepath.Join(dir, config.InstallPath()))
	if err != nil {
		return "", err
	}
	log.Printf("[DEBUG] Find plugin path: %s", path)
	return path, err
}

// getPluginDir returns the base plugin directory.
// Adopted with the following priorities:
//
//  1. `plugin_dir` in a global config
//  2. `TOFULINT_PLUGIN_DIR` environment variable
//  3. Current directory (./.tofulint.d/plugins)
//  4. Home directory (~/.tofulint.d/plugins)
//
// If the environment variable is set, other directories will not be considered,
// but if the current directory does not exist, it will fallback to the home directory.
func getPluginDir(cfg *tofulint.Config) (string, error) {
	if cfg.PluginDir != "" {
		return homedir.Expand(cfg.PluginDir)
	}

	if dir := os.Getenv("TOFULINT_PLUGIN_DIR"); dir != "" {
		return dir, nil
	}

	_, err := os.Stat(localPluginRoot)
	if os.IsNotExist(err) {
		return homedir.Expand(PluginRoot)
	}

	return localPluginRoot, err
}

// findPluginPath returns the path of the existing plugin.
// Only in the case of Windows, the pattern with the `.exe` is also considered,
// and if it exists, the extension is added to the argument.
func findPluginPath(path string) (string, error) {
	if runtime.GOOS != "windows" {
		return checkPluginExistance(path)
	}

	returnPath, err := checkPluginExistance(path)
	if os.IsNotExist(err) {
		return checkPluginExistance(path + ".exe")
	}

	return returnPath, err
}

func checkPluginExistance(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	// directory is invalid as a plugin path
	if info.IsDir() {
		return "", os.ErrNotExist
	}

	return path, nil
}

func pluginClientError(err error, config *tofulint.PluginConfig) error {
	if err == nil {
		return nil
	}

	message := err.Error()
	search := "Incompatible API version with plugin."

	if strings.Contains(message, search) {
		message = strings.Replace(
			message,
			search,
			fmt.Sprintf(`TofuLint is not compatible with this version of the %q plugin. A newer TofuLint or plugin version may be required.`, config.Name),
			-1,
		)

		return errors.New(message)
	}

	return err
}
