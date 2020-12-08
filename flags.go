package main

import (
	"flag"

	config "github.com/harrisonhesslink/pythia/configuration"
)

//parseFlags This evaluates the flags used when the program was run
//and assigns the values of those flags according to sane defaults.
func flags(c *config.Config) {

	apiport := flag.Int("apiport", 4200, "Port to run Karai Coordinator API on.")
	wantclean := flag.Bool("clean", false, "Clear all peer certs")
	dir := flag.String("dir", "./config", "Change the dir of all duh fyles")
	lport := flag.Int("l", 4201, "wait for incoming connections")
	name := flag.String("database-name", "transactions", "set database-name for psql")
	flag.Parse()

	c.KaraiAPIPort = *apiport
	c.WantsClean = *wantclean
	c.ConfigDir = *dir
	c.Lport = *lport
	c.TableName = *name

}
