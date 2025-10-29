package main

import "time"

// config holds runtime options and comparison settings.
type config struct {
  Source         string
  All            bool
  AllAll		 bool
  Comment        bool
  Composer       bool
  DiscCountZero  bool
  TrackCountZero bool
  Picture        bool
  MinPicSize     int
  JSON           bool
  CSV            bool
  Workers        int
  Quiet          bool
}

// Now returns current time.
// (Tiny helper so code reads nicely; same as time.Now()).
func Now() time.Time {
  return time.Now()
}
