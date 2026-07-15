package main

import "github.com/michaelquigley/push/build"

func init() {
	// vane is pre-release; advertise the dev base as v0.1.x for unstamped
	// developer builds.
	build.DevVersion = "v0.1.x"
}
