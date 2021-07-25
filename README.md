# check-lastlog

check recently unlogged in users

## usage

```
$ ./check-lastlog -h 
Usage:
  check-lastlog [OPTIONS]

Application Options:
      --before=           [Deprecated] Check for users whose login is older than DAYS
  -w, --warning=          warning if users whose login is older than DAYS (default: 60)
  -c, --critical=         critical if users whose login is older than DAYS (default: 85)
      --min-uid=          min uid to check lastlog (default: 500)
      --max-uid=          max uid to check lastlog (default: 60000)
      --white-user-names= comma separeted user names that white
  -v, --version           Show version

Help Options:
  -h, --help              Show this help message
```

## sample

```
$ ./check-lastlog --white-user-names boofy,pages
CRITICAL: Found users who have not logged in recently: testuser(129 days), sampleuser(106 days)
$ echo $?
2
```

## Install

```
$ mkr plugin install kazeburo/check-lastlog
```