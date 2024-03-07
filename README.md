# nfson
recursive JSON-parser for nested structs using [valyala/fastjson](https://github.com/valyala/fastjson)

Uses the struct-tag `nfson`.

A JSON-path is specified in the `nfson`-struct tag as such:

`<json-element>.<json-sub-element>.<json-sub-sub-element>` etc.

```
Map(data []byte, obj interface{}, timeLoc *time.Location, subTagName string, recurseSubTag bool, baseTags ...string)
```

- `data` is the JSON as a byte array.
- `obj` is a reference to the struct the JSOn should be mapped to
- `timeLoc` is the location times in the json should be parse for.
- `subTagName` is for having multiple different JSON structures that should map to the same struct. It is an addition to the base `nfson` struct tag, for example: to use struct tag `nfson_subTag` use `Map()` with `_subTag` as the `subTagName` value.
- `recurseSubTag` indicates whether the `subTagName` should be kept when mapping recursively, if set it will look for the struct tag `nfson<subTagName>` in sub-structs as well, otherwise it will use the base `nfson`-struct tag for sub-structs.
- `baseTags` is for "skipping" steps in the JSON, can be used if your struct is only for mapping part of the JSON. Is also used for recursive mapping so that the sub-structs don't have to specify the entire JSON-element path and is instead able to start from the JSON-path of the parent struct.

## Supported types

Values and references to values of the following types:

- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool`
- `time.Time`
- `struct`

## Supported time formats

- `MM/dd/yyy HH:mm:ss`
- `yyyy-MM-dd HH:mm:ss`
- `MM/dd/yyyy`
- `yyyy-MM-dd`
- `yyyy-MM`


## Example
```json
{
    "data": {
        "teststring":   "test",
        "testbool":     false,
        "testpointer":  null,
        "nest": {
            "key_1":    null,
            "key_2":    null,
            "key_3":    "not null"
        },
        "type":     "user",
        "user":     {
            "name": "username"
        },
        "owner": {},
        "date": "1999-12-31 13:33:37"
    }
}
```

```go
type test struct {
    TestString      String      `nfson:"data.teststring"`
    TestBool        bool        `nfson:"data.testbool"`
    TestPointer     *string     `nfson:"data.testpointer"`
    Type            string      `nfson:"data.type"`
    Date            *time.Time  `nfson:"data.date"`
    NestTest        *nest       `nfson:"data.nest"`
    PersonTest      *person     `nfson_user:"data.user" nfson_owner:"data.owner"`
}

type nest struct { 
    Key1    *string `nfson:"key_1"`
    Key2    *string `nfson:"key_2"`
    Key3    *string `nfson:"key_3"`
}

type person {
    Name    string  `nfson:"name"`
}

func main() {
    // Pretend this gets the example JSON above
    bytes := whatever.GetSomeJSONBytes()

    location, _ := time.LoadLocation("Europe/Stockholm")

    example := test{}

    nfson.Map(bytes, &example, location, "", false)

    if example.Type == "user" {
        nfson.Map(bytes, &example, location, "_user", false)
    } else if example.Type == "owner" {
        nfson.Map(bytes, &example, location, "_owner", false)
    }
}
```

Checking for which subtag to use can be done without a `data.type`-field in the JSON, but it is a bit more involved and requires some manual fiddling with the fastjson parser and tags.