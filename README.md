# nfson
Recursive JSON-parser for nested structs using [valyala/fastjson](https://github.com/valyala/fastjson).

Allows more dynamic JSON-mapping using subtags to conditionally map different parts of the JSON to different structs.

The built in nesting support in the tags also allows you to have an arbitrary struct structure and still map the entire JSON. 

```
Map(data []byte, obj interface{}, timeLoc *time.Location, subTagName string, recurseSubTag bool, baseTags ...string)
```

- `data` is the JSON as a byte array.
- `obj` is a reference to the struct the JSOn should be mapped to
- `timeLoc` is the location times in the json should be parse for.
- `subTagName` is for having multiple different JSON structures that should map to the same struct. It is an addition to the base `nfson` struct tag, for example: to use struct tag `nfson_subTag` use `Map()` with `_subTag` as the `subTagName` value.
- `recurseSubTag` indicates whether the `subTagName` should be kept when mapping recursively, if set it will look for the struct tag `nfson<subTagName>` in sub-structs as well, otherwise it will use the base `nfson`-struct tag for sub-structs.
- `baseTags` is for "skipping" steps in the JSON, can be used if your struct is only for mapping part of the JSON. Is also used for recursive mapping so that the sub-structs don't have to specify the entire JSON-element path and is instead able to start from the JSON-path of the parent struct.

## Tags

Uses the struct-tag `nfson`.

A JSON-path is specified in the `nfson`-struct using JSON-element names seperated by periods (`.`) to specify which element to map:

`<json-element>.<json-sub-element>.<json-sub-sub-element>` etc.

As an example the following will map `key_3` to `testStruct.TestField`:
```json
{
    "data": {
        "subelement":{
            "key_1": true
        }
    }
}
```
```go
type testStruct struct {
    TestField    *bool   `nfson:"data.subelement.key_1"`
}
```
In the case of nested structs the nested struct will automatically continue from the parent struct tag:
```go
type testStruct struct {
    SubStruct   *testSubStruct  `nfson:"data.subelement"`
}

type testSubStruct struct {
    TestField   *bool   `nfson:"key_1"`
}
```

## SubTags

Sub tags allows having multiple different JSON-paths for the same struct field using an arbitrary suffix for the `nfson`-base tag.

This can be useful if you wish to save the structs to a relational database. In the example below you can receive a JSON-object with detailed information about a car and its user/owner, map the car and any combination of person/company user/owner, save them to the database, and set the values required for the relationships.

```json
{
    "data": {
        "reg": "abc123",
        "user": {
            "data": {
                "type": "person",
                "ssn": "0123456789",
                "name": "Jon Doe"
            }
        },
        "owner": {
            "data": {
                "type": "company",
                "orgno": "9876543210",
                "name": "umbrella corp"
            }
        }
    }
}
```

```go
const (
    userSubTag = "_user"
    ownerSubTag = "_owner"

    personTypeName = "person"
    companyTypeName = "company"

    userJSONpath = "data.user.data.type"
    ownerJSONpath = "data.owner.data.type"
)

type Company struct {
    OrgNo   *int64  `nfson_user:"data.user.data.orgno" nfson_owner:"data.owner.data.orgno"`
    Name    string  `nfson_user:"data.user.data.name" nfson_owner:"data.owner.data.name`

    // other relevant data
}

type Person struct {
    SSN     *int64  `nfson_user:"data.user.data.ssn" nfson_owner:"data.owner.data.ssn"`
    Name    string  `nfson_user:"data.user.data.name" nfson_owner:"data.owner.data.name"`

    // other relevant data
}

type Car struct {
    Reg             string  `nfson:"data.reg"`

    UserSSN         *int64
    UserOrgNo       *int64

    OwnerSSN        *int64
    OwnerOrgNo      *int64

    // other relevant data
}

type (c *Car) Map(json []byte, location *time.Location) {
    nfson.Map(bytes, c, location, "", false)

    parser, err := fastjson.ParseBytes(bytes)
	if err != nil {
		return err
	}

    // set user
	userType := string(parser.GetStringBytes(nfson.SplitTag(userJSONpath)...))
    if userType == personTypeName {
        person := Person{}
        nfson.Map(bytes, &person, location, userSubTag, false)
        person.SaveToDatebase() // pretend this saves the person to a database of your choice
        c.UserSSN = person.SSN
    } else if userType == companyTypeName {
        company := Company{}
        nfson.Map(bytes, &company, location, userSubTag, false)
        company.SaveToDatabase() // pretend this saves the company to a database of your choice
        c.UserOrgNo = company.OrgNo
    }

    // set owner
	ownerType := string(parser.GetStringBytes(nfson.SplitTag(ownerJSONpath)...))
    if ownerype == personTypeName {
        person := Person{}
        nfson.Map(bytes, &person, location, ownerSubTag, false)
        person.SaveToDatebase() // pretend this saves the person to a database of your choice
        c.OwnerSSN = person.SSN
    } else if ownerype == companyTypeName {
        company := Company{}
        nfson.Map(bytes, &company, location, ownerSubTag, false)
        company.SaveToDatabase() // pretend this saves the company to a database of your choice
        c.OwnerOrgNo = company.OrgNo
    }
}

func main() {
    bytes := whatever.GetSomeJSONBytes() // Pretend this gets the example JSON above

    location, _ := time.LoadLocation("Europe/Stockholm")

    car := Car{}
    car.Map(bytes, location)
    car.SaveToDatabase() // pretend this saves the car to a database of your choice
}
```

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