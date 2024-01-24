package main

import "github.com/astriaorg/messenger-rollup/messenger"

func main() {
	println("hello, world!")

	// load env vars

	app := messenger.NewApp(":50051", ":26658", ":8080")
	app.Run()
}
