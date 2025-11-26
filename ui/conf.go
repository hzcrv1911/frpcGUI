package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/lxn/walk"
	"github.com/samber/lo"

	"github.com/hzcrv1911/frpcgui/i18n"
	"github.com/hzcrv1911/frpcgui/pkg/config"
	"github.com/hzcrv1911/frpcgui/pkg/consts"
	"github.com/hzcrv1911/frpcgui/pkg/util"
	"github.com/hzcrv1911/frpcgui/services"
)

// The flag controls the running state of service.
type runFlag int

const (
	runFlagAuto runFlag = iota
	runFlagForceStart
	runFlagReload
)

// Conf contains all data of a config
type Conf struct {
	// Path of the config file
	Path string
	// State of service
	State consts.ConfigState
	Data  *config.ClientConfig
}

// PathOfConfInProfile returns the config path in profile directory (R_<IP>_<port>/)
func PathOfConfInProfile(data *config.ClientConfig, filename string) string {
	if data == nil || data.ServerAddress == "" {
		// If no server info, generate a temporary path (should be updated later)
		return filepath.Join("profiles", "temp", filename)
	}
	// Generate directory name: R_<ip>_<port>
	// Replace dots and colons with underscores
	serverAddr := strings.ReplaceAll(data.ServerAddress, ".", "_")
	serverAddr = strings.ReplaceAll(serverAddr, ":", "_")
	dirName := fmt.Sprintf("R_%s_%d", serverAddr, data.ServerPort)
	return filepath.Join("profiles", dirName, filename)
}

func NewConf(path string, data *config.ClientConfig) *Conf {
	if path == "" {
		// Use config name as filename, fallback to random token if no name
		filename := data.Name()
		if filename == "" {
			token, err := util.RandToken(16)
			if err != nil {
				panic(err)
			}
			filename = token
		}
		// Sanitize filename (remove invalid characters)
		filename = strings.Map(func(r rune) rune {
			if strings.ContainsRune(`<>:"/\|?*`, r) {
				return '_'
			}
			return r
		}, filename)
		// Use profile directory structure for new configs
		path = PathOfConfInProfile(data, filename+".conf")
	}
	return &Conf{
		Path:  path,
		State: consts.ConfigStateNotInstalled, // Default to not installed, tracker will update actual state
		Data:  data,
	}
}

func (conf *Conf) Name() string {
	return conf.Data.Name()
}

// Delete config will remove service, logs, config file in disk
func (conf *Conf) Delete() error {
	// Delete service
	running := conf.State == consts.ConfigStateStarted
	if err := services.UninstallService(conf.Path, true); err != nil && running {
		return err
	}
	// Delete logs
	if logs, _, err := util.FindLogFiles(conf.Data.LogFile); err == nil {
		util.DeleteFiles(logs)
	}
	// Delete config file
	if err := os.Remove(conf.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	// Try to remove the profile directory if it's empty
	// This only applies if the config is in a R_* subdirectory
	configDir := filepath.Dir(conf.Path)
	if filepath.Base(filepath.Dir(configDir)) == "profiles" && len(filepath.Base(configDir)) > 2 && filepath.Base(configDir)[:2] == "R_" {
		// This is a profile subdirectory, try to remove it (will fail if not empty)
		os.Remove(configDir)
	}

	return nil
}

// Save config to the disk. The config will be completed before saving
func (conf *Conf) Save() error {
	// Use config name as filename
	filename := conf.Data.Name()
	if filename == "" {
		// Keep existing filename if no name is set
		filename = util.FileNameWithoutExt(conf.Path)
	}
	// Sanitize filename
	filename = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`<>:"/\|?*`, r) {
			return '_'
		}
		return r
	}, filename)

	expectedPath := PathOfConfInProfile(conf.Data, filename+".conf")

	// If path needs to be updated (e.g., from temp or server info changed)
	if conf.Path != expectedPath {
		oldPath := conf.Path
		conf.Path = expectedPath

		// Remove old file if it exists in a different location
		if oldPath != "" && oldPath != expectedPath {
			if _, err := os.Stat(oldPath); err == nil {
				os.Remove(oldPath)
				// Try to remove old directory if empty
				oldDir := filepath.Dir(oldPath)
				os.Remove(oldDir)
			}
		}
	}

	// Ensure the directory exists before saving
	configDir := filepath.Dir(conf.Path)
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return err
	}

	logPath, err := filepath.Abs(filepath.Join("logs", util.FileNameWithoutExt(conf.Path)+".log"))
	if err != nil {
		return err
	}
	conf.Data.Complete(false)
	conf.Data.LogFile = filepath.ToSlash(logPath)
	return conf.Data.Save(conf.Path)
}

var (
	appConf = config.App{
		CheckUpdate: true,
		Defaults: config.DefaultValue{
			LogLevel:   consts.LogLevelInfo,
			LogMaxDays: consts.DefaultLogMaxDays,
			TCPMux:     true,
			TLSEnable:  true,
		},
	}
	confDB *walk.DataBinder
)

func loadAllConfs() ([]*Conf, error) {
	// Load and migrate application configuration.
	if lang, _ := config.UnmarshalAppConf(config.DefaultAppFile, &appConf); lang != nil {
		if _, ok := i18n.IDToName[*lang]; ok {
			appConf.Lang = *lang
			if saveAppConfig() == nil {
				os.Remove(config.LangFile)
			}
		} else {
			os.Remove(config.LangFile)
		}
	}
	// Find all config files in `profiles` directory subdirectories only.
	var files []string
	// Scan R_* subdirectories
	profilesDir := "profiles"
	if entries, err := os.ReadDir(profilesDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && len(entry.Name()) > 2 && entry.Name()[:2] == "R_" {
				subdirFiles, err := filepath.Glob(filepath.Join(profilesDir, entry.Name(), "*.conf"))
				if err == nil {
					files = append(files, subdirFiles...)
				}
			}
		}
	}
	cfgList := make([]*Conf, 0)
	for _, f := range files {
		if conf, err := config.UnmarshalClientConf(f); err == nil {
			c := NewConf(f, conf)
			if c.Name() == "" {
				conf.ClientCommon.Name = util.FileNameWithoutExt(f)
			}
			cfgList = append(cfgList, c)
		}
	}
	slices.SortStableFunc(cfgList, func(a, b *Conf) int {
		i := slices.Index(appConf.Sort, util.FileNameWithoutExt(a.Path))
		j := slices.Index(appConf.Sort, util.FileNameWithoutExt(b.Path))
		if i < 0 && j >= 0 {
			return 1
		} else if j < 0 && i >= 0 {
			return -1
		}
		return i - j
	})
	return cfgList, nil
}

// ConfBinder is the view model of configs
type ConfBinder struct {
	// Current selected config
	Current *Conf
	// List of configs
	List func() []*Conf
	// Set Config state
	SetState func(conf *Conf, state consts.ConfigState) bool
	// Commit will save the given config and try to reload service
	Commit func(conf *Conf, flag runFlag)
}

// getCurrentConf returns the current selected config
func getCurrentConf() *Conf {
	if confDB != nil {
		if ds, ok := confDB.DataSource().(*ConfBinder); ok {
			return ds.Current
		}
	}
	return nil
}

// setCurrentConf set the current selected config, the views will get notified
func setCurrentConf(conf *Conf) {
	if confDB != nil {
		if ds, ok := confDB.DataSource().(*ConfBinder); ok {
			ds.Current = conf
			confDB.Reset()
		}
	}
}

// commitConf will save the given config and try to reload service
func commitConf(conf *Conf, flag runFlag) {
	if confDB != nil {
		if ds, ok := confDB.DataSource().(*ConfBinder); ok {
			ds.Commit(conf, flag)
		}
	}
}

// getConfList returns a list of all configs.
func getConfList() []*Conf {
	if confDB != nil {
		if ds, ok := confDB.DataSource().(*ConfBinder); ok {
			return ds.List()
		}
	}
	return nil
}

func setConfState(conf *Conf, state consts.ConfigState) bool {
	if confDB != nil {
		if ds, ok := confDB.DataSource().(*ConfBinder); ok {
			return ds.SetState(conf, state)
		}
	}
	return false
}

func newDefaultClientConfig() *config.ClientConfig {
	return &config.ClientConfig{
		ClientCommon: appConf.Defaults.AsClientConfig(),
	}
}

func saveAppConfig() error {
	return appConf.Save(config.DefaultAppFile)
}

func setConfOrder(cfgList []*Conf) {
	appConf.Sort = lo.Map(cfgList, func(item *Conf, index int) string {
		return util.FileNameWithoutExt(item.Path)
	})
	saveAppConfig()
}
