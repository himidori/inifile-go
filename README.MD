# inifile-go

# usage


## creating inifile object
'''
ini, err := inifile.NewIniFile("xxx")
'''

## adding new section
'''
err = ini.AddSection("xxx")
'''

## adding new key. section will be created automatically if it doesn't exist
'''
err = ini.WriteKey("section", "keyname", "keyvalue")
'''

## reading key
'''
val, err := ini.ReadKey("section", "keyname")
'''

## deleting key
'''
err = ini.DeleteKey("section", "keyname")
'''
