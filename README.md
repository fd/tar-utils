# `tar-utils`

## Usage

```
usage: x-tar [<flags>] <command> [<args> ...]

Tar utilities

Flags:
  --help     Show context-sensitive help (also try --help-long and --help-man).
  --version  Show application version.

Commands:
  help [<command>...]
    Show help.


  build [<flags>] [<context-dir>]
    Make a new tar file

    -t, --tarfile=FILE  Tarfile location
    -o, --output=FILE   Path to output Tar archive
```

## Tarfile format

```
COPY <src> <dst>
MKDIR <src> <dst>
CHMOD [-R] <mode> <targets>
CHOWN [-R] (<user> | <user>:<group> | :<group>) <targets>
```

