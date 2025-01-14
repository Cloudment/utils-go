# Cloudment Utilities Go

[![Go Report Card](https://goreportcard.com/badge/github.com/cloudment/utils-go)](https://goreportcard.com/report/github.com/cloudment/utils-go)

Useful utilities for reducing dependencies and improving performance within Cloudment Go projects.

## Install

```bash
go get github.com/cloudment/utils-go
```

## Usage

### Env - Main features

Before we required both `joho/godotenv` and `caarlos0/env` to parse environment variables and store them in a struct.

However, this package eliminates the need for both, in a faster and simpler way.

In order to read a `.env`, parse, and store it in a struct, we would need to use both libraries.

#### Benchmarks taken on an AMD Ryzen 9 7950X3D

| Library<br/>Function                                   | Benchmark time  | Change / Percentage Difference       |
|--------------------------------------------------------|-----------------|--------------------------------------|
| `joho/godotenv` <br/> `Load()`                         | 22,207 ns/op    | N/A                                  |
| `cloudment/utils-go` <br/> `ParseFromFileIntoStruct()` | **9,013 ns/op** | **13,194 ns/op quicker <br/>   84%** |
| `caarlos0/env`  <br/>  `Parse()`                       | 4,750 ns/op     | N/A                                  |
| `cloudment/utils-go` <br/> `Parse()`                   | **2,971 ns/op** | **1,779 ns/op quicker <br/>   46%**  |

Note: `joho/godotenv` tests were done including `caarlos0/env` as it is required to parse the `.env` file and store it in a struct.

#### Example
```go
package main

import (
    "fmt"
    "github.com/cloudment/utils-go/env"
)

type Config struct {
    Port int `env:"PORT" envDefault:"8080"`
}

func main() { 
    // with just os.Environ() 
    var cfg Config
    env.Parse(&cfg)
    fmt.Println(cfg.Port)
	
    // with both os.Environ() and a .env file
    var cfg2 Config
    _ = env.ParseFromFileIntoStruct(&cfg2, ".env") // uses os.Environ() and .env file
    fmt.Println(cfg.Port)
}
```

### Utils - Main features

#### `bind_request.go` - Binds a request to a struct.

This allows you to bind a request to a struct, which is useful for APIs.

```go
package main

import (
    "net/http"
    "github.com/cloudment/utils-go/utils"
)

type Request struct {
    Field1     string  `query:"field1" form:"field1" json:"field1" required:"true"`
    Field2     string  `query:"field2" form:"field2" json:"field2"`
}

func handler(w http.ResponseWriter, r *http.Request) {
    var req Request
    if err := BindRequest(r, &req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
}
```

#### `gorm_search_query.go` - Generates a search query for GORM.

This would allow you to specify multiple query parameters and generate a query for GORM.

If both `ID` and `Array` are provided, the query would be `(id = ? AND ? = ANY(array))`. If not, it only uses 1.

This could be used to allow someone to search for a `name`, `id`, `rank`, etc. in a database or all 3 at once.

```go
package main

import (
    "github.com/cloudment/utils-go/utils"
)

type OptionalQueryParams struct {
     ID    string `query:"id = ?"`
     Array string `query:"? = ANY(array)"`
}

func main() {
	// Ignore this, it's just an example
	db, _ := pretendARealDatabaseConnection().ToARealDatabase().ToARealTable()
	
	
	params := OptionalQueryParams{ID: "123", Array: "type1"}
	query, args := GormSearchQuery(params)

	// query = "(id = ? AND ? = ANY(array))"
	// args = ["123", "type1"]
	
	// Now you can use this query and args in your GORM query, 
	// in this example it would require the ID to be 123 and the Array to contain "type1"
	db = db.Where(query, args...).Find(&results)
}
```