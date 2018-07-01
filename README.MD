# inifile-go

# usage


## creating inifile object
```go
ini, err := inifile.NewIniFile("xxx")
```

## adding new section
```go
err = ini.AddSection("xxx")
```

## adding new key. section will be created automatically if it doesn't exist
```go
err = ini.WriteKey("section", "keyname", "keyvalue")
```

## reading key
```go
val, err := ini.ReadKey("section", "keyname")
```

## deleting key
```go
err = ini.DeleteKey("section", "keyname")
```
