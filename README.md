# Dux - A Package Manager Manager

Basically just a command-line tool to declaratively control package managers.
You specify how your package manager works in your configuration file (located
at `~/.config/dux/config.yaml`), and `dux` will control your package manager to do
what you want it to.

The most important feature of Dux is that it's declarative. I personally find
this useful because I install things and then completely forget to get rid of
them after I'm finished with them. With `dux`, everything you install is
explicitly listed, so keeping track of all of your installed packages is
arguably a bit easier.

I'm the only person who uses this, so it's not documented at all. If you want to
use this and you're not me, add an issue I guess.

## Getting started
 - Copy a config file (example in `example-config/config.yaml`) to `~/.config/dux/config.yaml`
   - Look at the config file to see how it works
 - Run `dux generate` to generate an example config. The config language parser
   is in `utils/util.go`.
 - Clean up your package lists, separate into different files, and add comments!
 - You're set up! After this, intermittently run `dux` to apply the files in
   `~/.config/dux/pkgs/*` to your config. Update your package files whenever
   you've installed a package that you want to keep around.