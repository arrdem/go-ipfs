package main

import (
	"encoding/base64"
	"errors"
	"os"

	"github.com/gonuts/flag"
	"github.com/jbenet/commander"
	config "github.com/jbenet/go-ipfs/config"
	"github.com/jbenet/go-ipfs/identify"
	u "github.com/jbenet/go-ipfs/util"
)

var cmdIpfsInit = &commander.Command{
	UsageLine: "init",
	Short:     "Initialize ipfs local configuration",
	Long: `ipfs init

	Initializes ipfs configuration files and generates a
	new keypair.
`,
	Run:  initCmd,
	Flag: *flag.NewFlagSet("ipfs-init", flag.ExitOnError),
}

func init() {
	cmdIpfsInit.Flag.Int("b", 4096, "number of bits for keypair")
	cmdIpfsInit.Flag.String("p", "", "passphrase for encrypting keys")
	cmdIpfsInit.Flag.Bool("f", false, "force overwrite of existing config")
}

func initCmd(c *commander.Command, inp []string) error {
	_, err := os.Lstat(config.DefaultConfigFilePath)
	force := c.Flag.Lookup("f").Value.Get().(bool)
	if err != nil && !force {
		return errors.New("ipfs configuration file already exists!\nReinitializing would overwrite your keys.\n(use -f to force overwrite)")
	}
	cfg := new(config.Config)

	cfg.Datastore = new(config.Datastore)
	dspath, err := u.TildeExpansion("~/.go-ipfs/datastore")
	if err != nil {
		return err
	}
	cfg.Datastore.Path = dspath
	cfg.Datastore.Type = "leveldb"

	cfg.Identity = new(config.Identity)
	// This needs thought
	// cfg.Identity.Address = ""

	nbits := c.Flag.Lookup("b").Value.Get().(int)
	if nbits < 1024 {
		return errors.New("Bitsize less than 1024 is considered unsafe.")
	}
	kp, err := identify.GenKeypair(nbits)
	if err != nil {
		return err
	}

	// pretend to encrypt key, then store it unencrypted
	enckey := base64.StdEncoding.EncodeToString(kp.PrivBytes())
	cfg.Identity.PrivKey = enckey

	id, err := kp.ID()
	if err != nil {
		return err
	}
	cfg.Identity.PeerID = id.Pretty()

	path, err := u.TildeExpansion(config.DefaultConfigFilePath)
	if err != nil {
		return err
	}
	err = config.WriteConfigFile(path, cfg)
	if err != nil {
		return err
	}
	return nil
}