# kaido

## Overview

Send notifications (such as discord message with webhook) whenever
someone validated new record for a map on Touge Union's KBT Asetto Corsa servers.

Previously was [KBT-Leaderboard](https://github.com/dimfu/KBT-leaderboard)
but it was over engineered and I decided to make it more less confusing to use
it personally.

## Installation

To use kaido, you need to have Golang installed on your machine.
Once Golang is installed, you can install kaido with the following command:

```bash
go install github.com/dimfu/kaido@latest
```

## Usage

After installation, you can run kaido using the following command:

```bash
kaido [command] [flags]
```

### Commands

```bash
run, r        collect all or some map records
leaderboards  See all available leaderboards
help, h       Shows a list of commands or help for one command
```

### Examples

To get all leaderboard records:

```bash
kaido run
```

To get specific leaderboard regions for current month:

```bash
# Add the -c flag to scope the records to the current month.
kaido run -leaderboard="gunma, kanagawa -c"
```

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE)
file for details.
