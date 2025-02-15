package main

import (
	"encoding/json"
	"os"
	"path"

	"github.com/dimfu/kaido/config"
)

func createWorkDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	wPath := path.Join(dir, ".kaido")
	_, err = os.Stat(wPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(wPath, os.ModePerm); err != nil {
			return "", err
		}
	}
	return wPath, nil
}

func createCfgFile(workDir string) error {
	cfg := config.GetConfig()
	cfgPath := path.Join(workDir, "config.json")
	_, err := os.Stat(cfgPath)

	if os.IsNotExist(err) {
		file, err := os.Create(cfgPath)
		if err != nil {
			return err
		}
		defer file.Close()

		cfg.KBTBaseUrl = "http://5.161.130.32:8000"
		cfg.WorkspacePath = workDir

		bytes, err := json.MarshalIndent(cfg, "", "\t")
		if err != nil {
			return err
		}

		_, err = file.Write(bytes)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	bytes, err := os.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, cfg)
	if err != nil {
		return err
	}

	return nil
}

func setup() error {
	workDir, err := createWorkDir()
	if err != nil {
		return err
	}
	if err := createCfgFile(workDir); err != nil {
		return err
	}

	return nil
}
