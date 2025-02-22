// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package clab

import (
	"fmt"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/srl-labs/containerlab/types"
	"github.com/srl-labs/containerlab/utils"
)

func initCrpdNode(c *CLab, nodeCfg NodeConfig, node *types.Node, user string, envs map[string]string) error {
	var err error

	node.Config, err = c.configInit(&nodeCfg, node.Kind)
	if err != nil {
		return err
	}
	node.Image = c.imageInitialization(&nodeCfg, node.Kind)
	node.Group = c.groupInitialization(&nodeCfg, node.Kind)
	node.Position = c.positionInitialization(&nodeCfg, node.Kind)
	node.User = user

	// initialize license file
	lp, err := c.licenseInit(&nodeCfg, node)
	if err != nil {
		return err
	}
	node.License = lp

	// mount config and log dirs
	node.Binds = append(node.Binds, fmt.Sprint(path.Join(node.LabDir, "config"), ":/config"))
	node.Binds = append(node.Binds, fmt.Sprint(path.Join(node.LabDir, "log"), ":/var/log"))
	// mount sshd_config
	node.Binds = append(node.Binds, fmt.Sprint(path.Join(node.LabDir, "config/sshd_config"), ":/etc/ssh/sshd_config"))

	return err
}

func (c *CLab) createCRPDFiles(node *types.Node) error {
	// create config and logs directory that will be bind mounted to crpd
	utils.CreateDirectory(path.Join(node.LabDir, "config"), 0777)
	utils.CreateDirectory(path.Join(node.LabDir, "log"), 0777)

	// copy crpd config from default template or user-provided conf file
	cfg := path.Join(node.LabDir, "/config/juniper.conf")

	err := node.GenerateConfig(cfg, defaultConfigTemplates[node.Kind])
	if err != nil {
		log.Errorf("node=%s, failed to generate config: %v", node.ShortName, err)
	}

	// copy crpd sshd conf file to crpd node dir
	src := "/etc/containerlab/templates/crpd/sshd_config"
	dst := node.LabDir + "/config/sshd_config"
	err = copyFile(src, dst)
	if err != nil {
		return fmt.Errorf("file copy [src %s -> dst %s] failed %v", src, dst, err)
	}
	log.Debugf("CopyFile src %s -> dst %s succeeded\n", src, dst)

	if node.License != "" {
		// copy license file to node specific lab directory
		src = node.License
		dst = path.Join(node.LabDir, "/config/license.conf")
		if err = copyFile(src, dst); err != nil {
			return fmt.Errorf("file copy [src %s -> dst %s] failed %v", src, dst, err)
		}
		log.Debugf("CopyFile src %s -> dst %s succeeded", src, dst)
	}
	return err
}
