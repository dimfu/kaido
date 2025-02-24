# kaido

## Overview

Send discord webhook whenever someone validated new record for a map on
Touge Union's KBT Asetto Corsa servers.

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
kaido run -leaderboard="gunma, kanagawa" -c
```

If you wished to execute the script automatically at given date and time,
you can schedule this script to be executed periodically by using cron job.

On Linux or UNIX system we could do simple setup that runs every hour like this:

```bash
# Open cron jobs entries
crontab -e
```

Append the following entry:

```bash
# Make sure to change `path_to_bin` to the kaido binary location
# and log to your desired location
0 * * * * path_to_bin/kaido run -c >> path_to_log/kaido.log 2>&1
```

Save and close the file.

To do something similar on Windows, you can follow this [guide](https://phoenixnap.com/kb/cron-job-windows).

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE)
file for details.
