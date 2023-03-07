# envsubst - the expanded version of os.Expand

## Features
* Supports `string`, `[]byte` and streams (`io.Reader`)
* The streams are buffered using `bufio`
* Data is assumed to be text so using `rune` instead of `byte`
* Use `SetPrefix` to choose from a set of different prefix-characters (default '`$`')
* Use `SetWrapper` to choose from a set of different wrapper-characters (default '`()`')
* Default mapping (`nil` argument) is `Getenv`

## Get started
```sh
go get github.com/ninlil/envsubst
```

## Example: read a config-file and replace from env-vars
```go
var buf bytes.Buffer

f,_ := os.Open("config.yaml")
defer f.Close()

err := envsubst.Convert(f, &buf, envsubst.LookupEnv)
if err != nil {
  panic(err)
}
bytes := buf.Bytes()
```

## What about the `os.Expand` ?

Yes, the stdlib have the `os.Expand*`-functions, but they only supports `string` and are not that suited for larger datasets.
