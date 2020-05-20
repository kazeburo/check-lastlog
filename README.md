# check-lastlog

check recently unlogged in users

## usage

```
$ ./check-lastlog -h 
Usage:
  check-lastlog [OPTIONS]

Application Options:
      --before=           Check for users whose login is older than DAYS (default: 85)
      --min-uid=          min uid to check lastlog (default: 500)
      --max-uid=           max uid to check lastlog (default: 60000)
      --white-user-names= comma separeted user names that white

Help Options:
  -h, --help              Show this help message
```

## sample

```
$ ./check-lastlog --white-user-names boofy 
2020/05/20 15:02:06 Found users who have not logged in recently: testuser(129 days), sampleuser(106 days)
```

