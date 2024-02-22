package main

import "github.com/astriaorg/messenger-rollup/messenger"

func main() {
	println("hello, world! i'm a messenger rollup!")

	// load env vars

	// executionPort, sequencerPort, restApiPort
	app := messenger.NewApp(":50051", "http://localhost:26657", ":8080")
	app.Run()
}
