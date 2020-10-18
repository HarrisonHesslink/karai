package main

import (
	"flag"
	config "github.com/karai/go-karai/configuration"
)

// parseFlags This evaluates the flags used when the program was run
// and assigns the values of those flags according to sane defaults.
func flags(c *config.Config) {
	// // flag.StringVar(&c.matrixToken, "matrixToken", "", "Matrix homeserver token")
	// // flag.StringVar(&c.matrixURL, "matrixURL", "", "Matrix homeserver URL")
	// // flag.StringVar(&c.matrixRoomID, "matrixRoomID", "", "Room ID for matrix publishd events")
	// c.SetkaraiAPIPort(flag.Int("apiport", 4200, "Port to run Karai Coordinator API on."))
	// c.SetwantsClean(flag.Bool("clean", false, "Clear all peer certs"))
	// // flag.BoolVar(&c.wantsMatrix, "matrix", false, "Enable Matrix functions. Requires -matrixToken, -matrixURL, and -matrixRoomID")
	// c.SetconfigDir(flag.String("dir", "./config", "Change the dir of all duh fyles"))
	// c.Setlport(flag.Int("l", 0, "wait for incoming connections"))

	//c.SettableName(flag.String("db-name", "transactions", "set db-name for psql"))

	apiport := flag.Int("apiport", 4200, "Port to run Karai Coordinator API on.")
	wantclean := flag.Bool("clean", false, "Clear all peer certs")
	dir := flag.String("dir", "./config", "Change the dir of all duh fyles")
	lport := flag.Int("l", 4201, "wait for incoming connections")
	name := flag.String("db-name", "transactions", "set db-name for psql")
	flag.Parse()
	
	c.KaraiAPIPort = *apiport
	c.WantsClean = *wantclean
	c.ConfigDir = *dir
	c.Lport = *lport
	c.TableName = *name


}
